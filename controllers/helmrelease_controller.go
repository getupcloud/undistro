/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"context"
	"os"
	"strings"
	"time"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/client/cluster"
	"github.com/getupcloud/undistro/client/cluster/helm"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// HelmReleaseReconciler reconciles a HelmRelease object
type HelmReleaseReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

// +kubebuilder:rbac:groups=getupcloud.com,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=getupcloud.com,resources=*,verbs=get;update;patch

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
	if ok, _ := cluster.IsReady(ctx, r.Client, nm); !ok {
		log.Info("cluster is not ready yet", "namespaced name", nm.String())
		hr.Status.Phase = undistrov1.HelmReleasePhaseWaitClusterReady
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	h, err := wc.GetHelm(nm.Name, nm.Namespace)
	if err != nil {
		return ctrl.Result{}, err
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
		serr = r.Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
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
		os.RemoveAll(ch.ChartPath)
		return ctrl.Result{}, nil
	}
	if ch.Changed {
		hr.Status.Phase = undistrov1.HelmReleasePhaseChartFetched
		hr.Status.LastAttemptedRevision = ch.Revision
		hr.Status.Revision = ch.Revision
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		serr = r.Update(ctx, &hr)
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
		serr = r.Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		return ctrl.Result{}, err
	}
	curRel, err := h.Get(hr.GetReleaseName(), helm.GetOptions{Namespace: hr.GetTargetNamespace()})
	if err != nil {
		if strings.Contains(err.Error(), "no such host") {
			log.Info("cluster is unreachable.", "err", err.Error(), "name", hr.GetClusterNamespacedName().String())
			return ctrl.Result{Requeue: true}, nil
		}
		log.Error(err, "failed to get release", "name", hr.Name)
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		hr.Status.Revision = ""
		hr.Status.LastAttemptedRevision = ""
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		serr = r.Update(ctx, &hr)
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
			log.Error(err, "fail to install")
			hr.Status.Phase = undistrov1.HelmReleasePhaseDeployFailed
			hr.Status.Revision = ""
			hr.Status.LastAttemptedRevision = ""
			serr := r.Status().Update(ctx, &hr)
			if serr != nil {
				log.Error(serr, "couldn't update status", "name", req.NamespacedName)
				return ctrl.Result{}, serr
			}
			serr = r.Update(ctx, &hr)
			if serr != nil {
				log.Error(serr, "couldn't update status", "name", req.NamespacedName)
				return ctrl.Result{}, serr
			}
			return ctrl.Result{}, err
		}
		status, err := h.Status(hr.GetReleaseName(), helm.StatusOptions{
			Namespace: hr.GetTargetNamespace(),
		})
		if err != nil {
			log.Error(err, "fail to get status")
			hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
			hr.Status.LastAttemptedRevision = ""

			serr := r.Status().Update(ctx, &hr)
			if serr != nil {
				log.Error(serr, "couldn't update status", "name", req.NamespacedName)
				return ctrl.Result{}, serr
			}
			serr = r.Update(ctx, &hr)
			if serr != nil {
				log.Error(serr, "couldn't update status", "name", req.NamespacedName)
				return ctrl.Result{}, serr
			}
		}
		hr.Status.Phase = undistrov1.HelmReleasePhaseDeployed
		hr.Status.ReleaseName = hr.GetReleaseName()
		hr.Status.ReleaseStatus = status.String()
		serr := r.Status().Update(ctx, &hr)
		if serr != nil {
			log.Error(serr, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, serr
		}
		serr = r.Update(ctx, &hr)
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
		WithEventFilter(predicate.Funcs{
			UpdateFunc: r.updateFilter,
		}).
		Complete(r)
}

func (r *HelmReleaseReconciler) updateFilter(e event.UpdateEvent) bool {
	newHr, ok := e.ObjectNew.(*undistrov1.HelmRelease)
	if !ok {
		return false
	}
	oldHr, ok := e.ObjectOld.(*undistrov1.HelmRelease)
	if !ok {
		return false
	}
	if newHr.Status.Phase == undistrov1.HelmReleasePhaseChartFetched || newHr.Status.Phase == undistrov1.HelmReleasePhaseWaitClusterReady {
		return true
	}
	diff := cmp.Diff(oldHr.Spec, newHr.Spec)

	// Filter out any update notifications that are due to status
	// updates, as the dry-run that determines if we should upgrade
	// is expensive, but _without_ filtering out updates that are
	// from the periodic refresh, as we still want to detect (and
	// undo) mutations to Helm charts.
	if sDiff := cmp.Diff(oldHr.Status, newHr.Status); diff == "" && sDiff != "" {
		return false
	}
	return true
}
