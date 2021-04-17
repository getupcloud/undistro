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
	"fmt"
	"time"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/fs"
	"github.com/getupio-undistro/undistro/pkg/kube"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/template"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// DefaultPoliciesReconciler reconciles a DefaultPolicies object
type DefaultPoliciesReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *DefaultPoliciesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	p := appv1alpha1.DefaultPolicies{}
	if err := r.Get(ctx, req.NamespacedName, &p); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log := r.Log.WithValues("defalutpolicies", req.NamespacedName)
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
	cl := &appv1alpha1.Cluster{}
	if p.Spec.ClusterName != "" {
		key := client.ObjectKey{
			Name:      p.Spec.ClusterName,
			Namespace: p.GetNamespace(),
		}
		err = r.Get(ctx, key, cl)
		if err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}
	if !p.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, &p, cl)
	}
	p, result, err := r.reconcile(ctx, log, p, cl)
	return result, err
}

func (r *DefaultPoliciesReconciler) reconcileDelete(ctx context.Context, p *appv1alpha1.DefaultPolicies, cl *appv1alpha1.Cluster) (ctrl.Result, error) {
	if p.Spec.ClusterName != "" && cl.Name != "" {
		return ctrl.Result{Requeue: true}, nil
	}
	controllerutil.RemoveFinalizer(p, meta.Finalizer)
	return ctrl.Result{}, nil
}

func (r *DefaultPoliciesReconciler) reconcile(ctx context.Context, log logr.Logger, p appv1alpha1.DefaultPolicies, cl *appv1alpha1.Cluster) (appv1alpha1.DefaultPolicies, ctrl.Result, error) {
	if p.Generation < p.Status.ObservedGeneration {
		log.V(2).Info("skipping this old version of reconciled object")
		return p, ctrl.Result{}, nil
	}
	var err error
	clusterClient := r.Client
	if p.Spec.ClusterName != "" {
		if !meta.InReadyCondition(cl.Status.Conditions) {
			return appv1alpha1.DefaultPoliciesNotReady(p, meta.WaitProvisionReason, "wait cluster to be provisioned"), ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		clusterClient, err = kube.NewClusterClient(ctx, r.Client, p.Spec.ClusterName, p.Namespace)
		if err != nil {
			return appv1alpha1.DefaultPoliciesNotReady(p, meta.GetClusterFailed, err.Error()), ctrl.Result{}, err
		}
	}
	p, err = r.installKyverno(ctx, clusterClient, p, cl)
	if err != nil {
		return appv1alpha1.DefaultPoliciesNotReady(p, meta.ObjectsApliedFailedReason, err.Error()), ctrl.Result{}, err
	}
	hr := appv1alpha1.HelmRelease{}
	key := client.ObjectKey{
		Name:      fmt.Sprintf("kyverno-%s", cl.Name),
		Namespace: p.GetNamespace(),
	}
	err = r.Get(ctx, key, &hr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return p, ctrl.Result{Requeue: true}, nil
		}
		return p, ctrl.Result{}, err
	}
	if !meta.InReadyCondition(hr.Status.Conditions) {
		return appv1alpha1.DefaultPoliciesNotReady(p, meta.WaitProvisionReason, "wait Kyverno to be installed"), ctrl.Result{Requeue: true}, nil
	}
	p, err = r.applyPolicies(ctx, clusterClient, p)
	if err != nil {
		appv1alpha1.DefaultPoliciesNotReady(p, meta.ArtifactFailedReason, err.Error())
	}
	return appv1alpha1.DefaultPoliciesReady(p), ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *DefaultPoliciesReconciler) applyPolicies(ctx context.Context, clusterClient client.Client, p appv1alpha1.DefaultPolicies) (appv1alpha1.DefaultPolicies, error) {
	dir, err := fs.PoliciesFS.ReadDir("apps/policies")
	if err != nil {
		return p, err
	}
	for _, f := range dir {
		if f.IsDir() {
			continue
		}
		byt, err := fs.PoliciesFS.ReadFile(f.Name())
		if err != nil {
			return p, err
		}
		objs, err := util.ToUnstructured(byt)
		if err != nil {
			return p, err
		}
		for _, o := range objs {
			_, err = util.CreateOrUpdate(ctx, clusterClient, &o)
			if err != nil {
				return p, err
			}
			p.Status.AppliedPolicies = append(p.Status.AppliedPolicies, o.GetName())
			set := sets.NewString(p.Status.AppliedPolicies...)
			p.Status.AppliedPolicies = set.UnsortedList()
		}
	}
	return p, nil
}

func (r *DefaultPoliciesReconciler) installKyverno(ctx context.Context, clusterClient client.Client, p appv1alpha1.DefaultPolicies, cl *appv1alpha1.Cluster) (appv1alpha1.DefaultPolicies, error) {
	vars := map[string]interface{}{
		"Cluster": cl,
	}
	objs, err := template.GetObjs(fs.AppsFS, "apps", "kyverno", vars)
	if err != nil {
		return p, err
	}
	for _, o := range objs {
		_, err = util.CreateOrUpdate(ctx, clusterClient, &o)
		if err != nil {
			return p, err
		}
	}
	meta.SetResourceCondition(&p, meta.ObjectsAppliedCondition, metav1.ConditionTrue, meta.ObjectsAppliedSuccessReason, "Kyverno installed")
	return p, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DefaultPoliciesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.DefaultPolicies{}).
		Complete(r)
}
