/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"context"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/client/cluster/helm"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// HelmReleaseReconciler reconciles a HelmRelease object
type HelmReleaseReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

// +kubebuilder:rbac:groups=getupcloud.com,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=getupcloud.com,resources=helmreleases/status,verbs=get;update;patch

func (r *HelmReleaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("helmrelease", req.NamespacedName)
	var hr undistrov1.HelmRelease
	if err := r.Get(ctx, req.NamespacedName, &hr); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't get object", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	undistroClient, err := uclient.New("")
	if err != nil {
		return ctrl.Result{}, err
	}
	wc, err := undistroClient.GetWorkloadCluster(uclient.Kubeconfig{
		RestConfig: r.RestConfig,
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	nm := hr.GetClusterNamespacedName()
	h, err := wc.GetHelm(nm.Name, nm.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !hr.DeletionTimestamp.IsZero() {
		log.Info("running uninstall")
		err = h.Uninstall(hr.GetReleaseName(), helm.UninstallOptions{
			Namespace:   hr.GetTargetNamespace(),
			KeepHistory: false,
			Timeout:     hr.GetTimeout(),
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	ch, err := helm.PrepareChart(h, &hr)
	if err != nil {
		log.Error(err, "couldn't prepare chart", "chartPath", ch.ChartPath, "revision", ch.Revision)
		hr.Status.Phase = undistrov1.HelmReleasePhaseChartFetchFailed
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		return ctrl.Result{}, err
	}
	if ch.Changed {
		hr.Status.Phase = undistrov1.HelmReleasePhaseChartFetched
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		return ctrl.Result{Requeue: true}, nil
	}
	values, err := helm.ComposeValues(ctx, r.Client, &hr, ch.ChartPath)
	if err != nil {
		log.Error(err, "failed to compose values for release", "name", hr.Name)
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		return ctrl.Result{}, err
	}
	curRel, err := h.Get(hr.GetReleaseName(), helm.GetOptions{Namespace: hr.GetTargetNamespace()})
	if err != nil {
		log.Error(err, "failed to get release", "name", hr.Name)
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		return ctrl.Result{}, err
	}
	if curRel == nil {
		log.Info("running instalation")
		_, err = h.UpgradeFromPath(ch.ChartPath, hr.GetReleaseName(), values, helm.UpgradeOptions{
			Namespace:         hr.GetTargetNamespace(),
			Timeout:           hr.GetTimeout(),
			Install:           true,
			Force:             hr.Spec.ForceUpgrade,
			SkipCRDs:          hr.Spec.SkipCRDs,
			MaxHistory:        hr.GetMaxHistory(),
			Wait:              hr.GetWait(),
			DisableValidation: false,
		})
		if err != nil {
			hr.Status.Phase = undistrov1.HelmReleasePhaseDeployFailed
			hr.Status.Revision = ch.Revision
			serr := r.Status().Update(ctx, &hr)
			if serr != nil {
				log.Error(serr, "couldn't update status", "name", req.NamespacedName)
				return ctrl.Result{}, serr
			}
			return ctrl.Result{}, err
		}
		hr.Status.Phase = undistrov1.HelmReleasePhaseDeployed
		hr.Status.Revision = ch.Revision
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
	}
	return ctrl.Result{}, nil

}

func (r *HelmReleaseReconciler) SetupWithManager(mgr ctrl.Manager, opts controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(opts).
		For(&undistrov1.HelmRelease{}).
		Complete(r)
}
