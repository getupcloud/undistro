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
	"encoding/json"
	"fmt"
	"net/url"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/helm"
	"github.com/getupio-undistro/undistro/pkg/kube"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/predicate"
	"github.com/getupio-undistro/undistro/pkg/record"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/strvals"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	getters = getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
	}
)

// HelmReleaseReconciler reconciles a HelmRelease object
type HelmReleaseReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	config *rest.Config
}

func (r *HelmReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	hr := appv1alpha1.HelmRelease{}
	if err := r.Get(ctx, req.NamespacedName, &hr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log := r.Log.WithValues("helmrelease", req.NamespacedName)
	// Add our finalizer if it does not exist
	if !controllerutil.ContainsFinalizer(&hr, meta.Finalizer) {
		controllerutil.AddFinalizer(&hr, meta.Finalizer)
		if err := r.Update(ctx, &hr); err != nil {
			log.Error(err, "unable to register finalizer")
			return ctrl.Result{}, err
		}
	}
	if hr.Spec.Paused {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	if !hr.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, hr)
	}
	hr, result, err := r.reconcile(ctx, log, hr)
	// Update status after reconciliation.
	if updateStatusErr := r.patchStatus(ctx, &hr); updateStatusErr != nil {
		log.Error(updateStatusErr, "unable to update status after reconciliation")
		return ctrl.Result{Requeue: true}, updateStatusErr
	}
	return result, err
}

func (r *HelmReleaseReconciler) reconcile(ctx context.Context, log logr.Logger, hr appv1alpha1.HelmRelease) (appv1alpha1.HelmRelease, ctrl.Result, error) {
	var clientOpts []getter.Option
	if hr.Status.ObservedGeneration != hr.Generation {
		hr.Status.ObservedGeneration = hr.Generation
		hr = appv1alpha1.HelmReleaseProgressing(hr)
		if updateStatusErr := r.patchStatus(ctx, &hr); updateStatusErr != nil {
			log.Error(updateStatusErr, "unable to update status after generation update")
			return hr, ctrl.Result{Requeue: true}, updateStatusErr
		}
	}
	if hr.Spec.Chart.SecretRef != nil {
		name := types.NamespacedName{
			Name:      hr.Spec.Chart.SecretRef.Name,
			Namespace: hr.GetNamespace(),
		}
		var secret corev1.Secret
		err := r.Client.Get(ctx, name, &secret)
		if err != nil {
			err = fmt.Errorf("auth secret error: %w", err)
			hr = appv1alpha1.HelmReleaseNotReady(hr, meta.AuthenticationFailedReason, err.Error())
			return hr, ctrl.Result{}, err
		}
		opts, cleanup, err := helm.ClientOptionsFromSecret(secret)
		if err != nil {
			err = fmt.Errorf("auth options error: %w", err)
			hr = appv1alpha1.HelmReleaseNotReady(hr, meta.AuthenticationFailedReason, err.Error())
			return hr, ctrl.Result{}, err
		}
		defer cleanup()
		clientOpts = opts
	}
	clientOpts = append(clientOpts, getter.WithTimeout(hr.Spec.Timeout.Duration))
	chartRepo, err := helm.NewChartRepository(hr.Spec.Chart.RepoURL, getters, clientOpts)
	if err != nil {
		switch err.(type) {
		default:
			hr = appv1alpha1.HelmReleaseNotReady(hr, meta.IndexationFailedReason, err.Error())
		case *url.Error:
			hr = appv1alpha1.HelmReleaseNotReady(hr, meta.URLInvalidReason, err.Error())
		}
		return hr, ctrl.Result{}, err
	}
	if err := chartRepo.DownloadIndex(); err != nil {
		err = fmt.Errorf("failed to download repository index: %w", err)
		hr = appv1alpha1.HelmReleaseNotReady(hr, meta.IndexationFailedReason, err.Error())
		return hr, ctrl.Result{}, err
	}
	ch, err := chartRepo.Get(hr.Spec.Chart.Name, hr.Spec.Chart.Version)
	if err != nil {
		hr = appv1alpha1.HelmReleaseNotReady(hr, meta.ChartPullFailedReason, err.Error())
		return hr, ctrl.Result{Requeue: true}, err
	}
	res, err := chartRepo.DownloadChart(ch)
	if err != nil {
		hr = appv1alpha1.HelmReleaseNotReady(hr, meta.ChartPullFailedReason, err.Error())
		return hr, ctrl.Result{Requeue: true}, err
	}
	// Compose values
	values, err := r.composeValues(ctx, hr)
	if err != nil {
		hr = appv1alpha1.HelmReleaseNotReady(hr, meta.InitFailedReason, err.Error())
		return hr, ctrl.Result{Requeue: true}, nil
	}
	hc, err := loader.LoadArchive(res)
	if err != nil {
		hr = appv1alpha1.HelmReleaseNotReady(hr, meta.StorageOperationFailedReason, err.Error())
		return hr, ctrl.Result{Requeue: true}, err
	}
	hr, err = r.reconcileRelease(ctx, log, *hr.DeepCopy(), hc, values)
	return hr, ctrl.Result{}, err
}

func (r *HelmReleaseReconciler) reconcileRelease(ctx context.Context, log logr.Logger,
	hr appv1alpha1.HelmRelease, chart *chart.Chart, values chartutil.Values) (appv1alpha1.HelmRelease, error) {
	getter, err := r.getRESTClientGetter(ctx, hr)
	if err != nil {
		return appv1alpha1.HelmReleaseNotReady(hr, meta.InitFailedReason, err.Error()), err
	}
	runner, err := helm.NewRunner(getter, hr.GetNamespace(), log)
	if err != nil {
		return appv1alpha1.HelmReleaseNotReady(hr, meta.InitFailedReason, "failed to initialize Helm action runner"), err
	}
	rel, err := runner.ObserveLastRelease(hr)
	if err != nil {
		return appv1alpha1.HelmReleaseNotReady(hr, meta.GetLastReleaseFailedReason, "failed to get last release revision"), err
	}
	revision := chart.Metadata.Version
	releaseRevision := util.ReleaseRevision(rel)
	valuesChecksum := util.ValuesChecksum(values)
	hr, hasNewState := appv1alpha1.HelmReleaseAttempted(hr, revision, releaseRevision, valuesChecksum)
	if hasNewState {
		hr = appv1alpha1.HelmReleaseProgressing(hr)
		if updateStatusErr := r.patchStatus(ctx, &hr); updateStatusErr != nil {
			log.Error(updateStatusErr, "unable to update status after state update")
			return hr, updateStatusErr
		}
	}
	// Check status of any previous release attempt.
	released := apimeta.FindStatusCondition(hr.Status.Conditions, meta.ReleasedCondition)
	if released != nil {
		if released.Status == metav1.ConditionTrue {
			return appv1alpha1.HelmReleaseReady(hr), nil
		}
	}
	if rel == nil {
		rel, err = runner.Install(hr, chart, values)
		err = r.handleHelmActionResult(&hr, revision, err, "install", meta.ReleasedCondition, meta.InstallSucceededReason, meta.InstallFailedReason)
	} else {
		rel, err = runner.Upgrade(hr, chart, values)
		err = r.handleHelmActionResult(&hr, revision, err, "upgrade", meta.ReleasedCondition, meta.UpgradeSucceededReason, meta.UpgradeFailedReason)

	}
	if util.ReleaseRevision(rel) > releaseRevision {
		if err == nil && hr.Spec.Test.Enable {
			_, err = runner.Test(hr)
			err = r.handleHelmActionResult(&hr, revision, err, "test", meta.TestSuccessCondition, meta.TestSucceededReason, meta.TestFailedReason)
			if err != nil && hr.Spec.Test.IgnoreFailures {
				err = nil
			}
		}
	}
	if err != nil {
		if util.ReleaseRevision(rel) <= releaseRevision {
			log.Info("skip, no new revision created")
		} else {
			err = runner.Rollback(hr)
			err = r.handleHelmActionResult(&hr, revision, err, "rollback", meta.RemediatedCondition, meta.RollbackSucceededReason, meta.RollbackSucceededReason)
		}
	}
	rel, observeLastReleaseErr := runner.ObserveLastRelease(hr)
	if observeLastReleaseErr != nil {
		err = observeLastReleaseErr
	}
	hr.Status.LastReleaseRevision = util.ReleaseRevision(rel)
	if err != nil {
		return appv1alpha1.HelmReleaseNotReady(hr, meta.ReconciliationFailedReason, err.Error()), err
	}
	return appv1alpha1.HelmReleaseReady(hr), nil
}

func (r *HelmReleaseReconciler) handleHelmActionResult(hr *appv1alpha1.HelmRelease, revision string, err error, action string, condition string, succeededReason string, failedReason string) error {
	if err != nil {
		msg := fmt.Sprintf("Helm %s failed: %s", action, err.Error())
		meta.SetResourceCondition(hr, condition, metav1.ConditionFalse, failedReason, msg)
		record.Warn(hr, failedReason, msg)
		return err
	} else {
		msg := fmt.Sprintf("Helm %s succeeded", action)
		meta.SetResourceCondition(hr, condition, metav1.ConditionTrue, succeededReason, msg)
		record.Event(hr, succeededReason, msg)
		return nil
	}
}

// composeValues attempts to resolve all \ValuesReference resources
// and merges them as defined. Referenced resources are only retrieved once
// to ensure a single version is taken into account during the merge.
func (r *HelmReleaseReconciler) composeValues(ctx context.Context, hr appv1alpha1.HelmRelease) (chartutil.Values, error) {
	result := chartutil.Values{}

	configMaps := make(map[string]*corev1.ConfigMap)
	secrets := make(map[string]*corev1.Secret)

	for _, v := range hr.Spec.ValuesFrom {
		namespacedName := types.NamespacedName{Namespace: hr.Namespace, Name: v.Name}
		var valuesData []byte

		switch v.Kind {
		case "ConfigMap":
			resource, ok := configMaps[namespacedName.String()]
			if !ok {
				// The resource may not exist, but we want to act on a single version
				// of the resource in case the values reference is marked as optional.
				configMaps[namespacedName.String()] = nil

				resource = &corev1.ConfigMap{}
				if err := r.Get(ctx, namespacedName, resource); err != nil {
					if apierrors.IsNotFound(err) {
						if v.Optional {
							r.Log.Info("could not find optional %s '%s'", v.Kind, namespacedName)
							continue
						}
						return nil, fmt.Errorf("could not find %s '%s'", v.Kind, namespacedName)
					}
					return nil, err
				}
				configMaps[namespacedName.String()] = resource
			}
			if resource == nil {
				if v.Optional {
					r.Log.Info("could not find optional %s '%s'", v.Kind, namespacedName)
					continue
				}
				return nil, fmt.Errorf("could not find %s '%s'", v.Kind, namespacedName)
			}
			if data, ok := resource.Data[v.ValuesKey]; !ok {
				return nil, fmt.Errorf("missing key '%s' in %s '%s'", v.ValuesKey, v.Kind, namespacedName)
			} else {
				valuesData = []byte(data)
			}
		case "Secret":
			resource, ok := secrets[namespacedName.String()]
			if !ok {
				// The resource may not exist, but we want to act on a single version
				// of the resource in case the values reference is marked as optional.
				secrets[namespacedName.String()] = nil

				resource = &corev1.Secret{}
				if err := r.Get(ctx, namespacedName, resource); err != nil {
					if apierrors.IsNotFound(err) {
						if v.Optional {
							r.Log.Info("could not find optional %s '%s'", v.Kind, namespacedName)
							continue
						}
						return nil, fmt.Errorf("could not find %s '%s'", v.Kind, namespacedName)
					}
					return nil, err
				}
				secrets[namespacedName.String()] = resource
			}
			if resource == nil {
				if v.Optional {
					r.Log.Info("could not find optional %s '%s'", v.Kind, namespacedName)
					continue
				}
				return nil, fmt.Errorf("could not find %s '%s'", v.Kind, namespacedName)
			}
			if data, ok := resource.Data[v.ValuesKey]; !ok {
				return nil, fmt.Errorf("missing key '%s' in %s '%s'", v.ValuesKey, v.Kind, namespacedName)
			} else {
				valuesData = data
			}
		default:
			return nil, fmt.Errorf("unsupported ValuesReference kind '%s'", v.Kind)
		}
		switch v.TargetPath {
		case "":
			values, err := chartutil.ReadValues(valuesData)
			if err != nil {
				return nil, fmt.Errorf("unable to read values from key '%s' in %s '%s': %w", v.ValuesKey, v.Kind, namespacedName, err)
			}
			result = util.MergeMaps(result, values)
		default:
			// TODO(hidde): this is a bit of hack, as it mimics the way the option string is passed
			// 	to Helm from a CLI perspective. Given the parser is however not publicly accessible
			// 	while it contains all logic around parsing the target path, it is a fair trade-off.
			singleValue := v.TargetPath + "=" + string(valuesData)
			if err := strvals.ParseInto(singleValue, result); err != nil {
				return nil, fmt.Errorf("unable to merge value from key '%s' in %s '%s' into target path '%s': %w", v.ValuesKey, v.Kind, namespacedName, v.TargetPath, err)
			}
		}
	}
	m := map[string]interface{}{}
	if hr.Spec.Values != nil {
		json.Unmarshal(hr.Spec.Values.Raw, &m)
	}
	return util.MergeMaps(result, m), nil
}

func (r *HelmReleaseReconciler) patchStatus(ctx context.Context, hr *appv1alpha1.HelmRelease) error {
	latest := &appv1alpha1.HelmRelease{}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(hr), latest); err != nil {
		return err
	}
	return r.Client.Status().Patch(ctx, hr, client.MergeFrom(latest))
}

func (r *HelmReleaseReconciler) getRESTClientGetter(ctx context.Context, hr appv1alpha1.HelmRelease) (genericclioptions.RESTClientGetter, error) {
	if hr.Spec.ClusterName == "" {
		return kube.NewInClusterRESTClientGetter(r.config, hr.GetNamespace()), nil
	}
	secretName := types.NamespacedName{
		Namespace: hr.GetNamespace(),
		Name:      fmt.Sprintf("%s-kubeconfig", hr.Spec.ClusterName),
	}
	var secret corev1.Secret
	if err := r.Get(ctx, secretName, &secret); err != nil {
		return nil, fmt.Errorf("could not find KubeConfig secret '%s': %w", secretName, err)
	}
	kubeConfig, ok := secret.Data["value"]
	if !ok {
		return nil, fmt.Errorf("KubeConfig secret '%s' does not contain a 'value' key", secretName)
	}
	return kube.NewMemoryRESTClientGetter(kubeConfig, hr.GetNamespace()), nil
}

func (r *HelmReleaseReconciler) reconcileDelete(ctx context.Context, logger logr.Logger, hr appv1alpha1.HelmRelease) (ctrl.Result, error) {
	restClient, err := r.getRESTClientGetter(ctx, hr)
	if err != nil {
		return ctrl.Result{}, err
	}
	runner, err := helm.NewRunner(restClient, hr.GetNamespace(), logger)
	if err != nil {
		return ctrl.Result{}, err
	}
	rel, err := runner.ObserveLastRelease(hr)
	if err != nil {
		return ctrl.Result{}, err
	}
	if rel == nil {
		return ctrl.Result{}, nil
	}
	err = runner.Uninstall(hr)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Remove our finalizer from the list and update it.
	controllerutil.RemoveFinalizer(&hr, meta.Finalizer)
	if err := r.Update(ctx, &hr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HelmReleaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.config = mgr.GetConfig()
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.HelmRelease{}, builder.WithPredicates(predicate.Changed{})).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}