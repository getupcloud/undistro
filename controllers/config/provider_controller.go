/*
Copyright 2020 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/cloud"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/retry"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var errNotReady = errors.New("chart isn't in ready condition")

// ProviderReconciler reconciles a Provider object
type ProviderReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	p := configv1alpha1.Provider{}
	if err := r.Get(ctx, req.NamespacedName, &p); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log := r.Log.WithValues("provider", req.NamespacedName)
	patchHelper, err := patch.NewHelper(&p, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		patchOpts := []patch.Option{}
		if err == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		patchErr := patchHelper.Patch(ctx, &p, patchOpts...)
		if patchErr != nil {
			err = kerrors.NewAggregate([]error{patchErr, err})
		}
	}()
	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&p, meta.Finalizer) {
		controllerutil.AddFinalizer(&p, meta.Finalizer)
		return ctrl.Result{}, nil
	}
	if p.Spec.Paused {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	if !p.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, p)
	}
	p, result, err := r.reconcile(ctx, log, p)
	return result, err
}

func (r *ProviderReconciler) reconcileDelete(ctx context.Context, logger logr.Logger, p configv1alpha1.Provider) (ctrl.Result, error) {
	key := client.ObjectKey{
		Name:      p.Status.HelmReleaseName,
		Namespace: p.GetNamespace(),
	}
	hr := appv1alpha1.HelmRelease{}
	err := r.Get(ctx, key, &hr)
	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{Requeue: true}, err
	}
	if apierrors.IsNotFound(err) {
		// Remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(&p, meta.Finalizer)
		_, err := util.CreateOrUpdate(ctx, r.Client, &p)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	err = r.Delete(ctx, &hr)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *ProviderReconciler) reconcile(ctx context.Context, log logr.Logger, p configv1alpha1.Provider) (configv1alpha1.Provider, ctrl.Result, error) {
	if p.Status.LastAttemptedVersion == "" {
		p, err := cloud.Init(ctx, r.Client, p)
		if err != nil {
			p = configv1alpha1.ProviderNotReady(p, meta.InitFailedReason, err.Error())
			return p, ctrl.Result{}, err
		}
	}
	p, err := r.reconcileChart(ctx, log, p)
	if err != nil {
		p = configv1alpha1.ProviderNotReady(p, meta.ChartAppliedFailedReason, err.Error())
		return p, ctrl.Result{}, err
	}
	p, err = r.checkState(ctx, log, p)
	if err != nil {
		p = configv1alpha1.ProviderNotReady(p, meta.WaitChartReason, err.Error())
		if err == errNotReady {
			err = nil
		}
		return p, ctrl.Result{}, err
	}
	return configv1alpha1.ProviderReady(p), ctrl.Result{}, nil
}

func (r *ProviderReconciler) reconcileChart(ctx context.Context, log logr.Logger, p configv1alpha1.Provider) (configv1alpha1.Provider, error) {
	err := retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
		hr := appv1alpha1.HelmRelease{
			TypeMeta: metav1.TypeMeta{
				APIVersion: appv1alpha1.GroupVersion.String(),
				Kind:       "HelmRelease",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.Name,
				Namespace: "undistro-system",
				Labels:    p.Labels,
			},
			Spec: appv1alpha1.HelmReleaseSpec{
				Paused:          p.Spec.Paused,
				AutoUpgrade:     p.Spec.AutoUpgrade,
				TargetNamespace: "undistro-system",
				ReleaseName:     p.Spec.ProviderName,
				ValuesFrom:      p.Spec.ConfigurationFrom,
				Chart: appv1alpha1.ChartSource{
					RepoChartSource: appv1alpha1.RepoChartSource{
						RepoURL: p.Spec.Repository.URL,
						Name:    p.Spec.ProviderName,
						Version: p.Spec.ProviderVersion,
					},
					SecretRef: p.Spec.Repository.SecretRef,
				},
			},
		}
		err := ctrl.SetControllerReference(&p, &hr, r.Scheme)
		if err != nil {
			return err
		}
		hasDiff, err := util.CreateOrUpdate(ctx, r.Client, &hr)
		if err != nil {
			return err
		}
		if hasDiff {
			p, err = cloud.Upgrade(ctx, r.Client, p)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return configv1alpha1.ProviderNotReady(p, meta.InitFailedReason, err.Error()), err
	}
	return configv1alpha1.ProviderAttempted(p, p.Name, p.Spec.ProviderVersion), nil
}

func (r *ProviderReconciler) checkState(ctx context.Context, log logr.Logger, p configv1alpha1.Provider) (configv1alpha1.Provider, error) {
	hr := appv1alpha1.HelmRelease{}
	key := client.ObjectKey{
		Name:      p.Status.HelmReleaseName,
		Namespace: p.GetNamespace(),
	}
	err := r.Get(ctx, key, &hr)
	if err != nil {
		return p, err
	}
	if !meta.InReadyCondition(hr.Status.Conditions) {
		return p, errNotReady
	}

	p.Status.LastAppliedVersion = p.Spec.ProviderVersion
	meta.SetResourceCondition(&p, meta.ChartAppliedCondition, metav1.ConditionTrue, meta.ChartAppliedSuccessReason, "chart applied")
	return p, nil
}

func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.Provider{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}
