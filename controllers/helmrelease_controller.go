/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"context"
	"os"
	"strings"
	"time"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	uclient "github.com/getupio-undistro/undistro/client"
	"github.com/getupio-undistro/undistro/client/cluster"
	"github.com/getupio-undistro/undistro/client/cluster/helm"
	"github.com/getupio-undistro/undistro/internal/patch"
	"github.com/getupio-undistro/undistro/internal/record"
	"github.com/getupio-undistro/undistro/internal/util"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/cluster-api/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

func (r *HelmReleaseReconciler) clusterClient(ctx context.Context, wc cluster.WorkloadCluster, nm types.NamespacedName) (client.Client, error) {
	workloadCfg, err := wc.GetRestConfig(nm.Name, nm.Namespace)
	if err != nil {
		return nil, err
	}
	return client.New(workloadCfg, client.Options{Scheme: r.Scheme})
}

func (r *HelmReleaseReconciler) execObjs(ctx context.Context, c client.Client, objs []apiextensionsv1.JSON) error {
	if len(objs) > 0 {
		for _, obj := range objs {
			st, err := yaml.ToUnstructured(obj.Raw)
			if err != nil {
				return err
			}
			for _, o := range st {
				err = util.CreateOrUpdate(ctx, c, o)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *HelmReleaseReconciler) hasNonDeployedDeps(ctx context.Context, hr *undistrov1.HelmRelease) bool {
	log := r.Log
	if len(hr.Spec.Dependencies) > 0 {
		for _, d := range hr.Spec.Dependencies {
			namespace := d.Namespace
			if namespace == "" {
				namespace = "default"
			}
			nm := types.NamespacedName{
				Name:      d.Name,
				Namespace: namespace,
			}
			o := unstructured.Unstructured{}
			o.SetGroupVersionKind(d.GroupVersionKind())
			err := r.Get(ctx, nm, &o)
			if err != nil {
				log.Error(err, "error when check dependency", "name", d.Name)
				return true
			}
			if o.GroupVersionKind() == hr.GroupVersionKind() {
				dep := undistrov1.HelmRelease{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, &dep)
				if err != nil {
					log.Error(err, "error when convert helmrelease")
					return true
				}
				if dep.Status.Phase != undistrov1.HelmReleasePhaseDeployed {
					log.Info("wait dependency to be deployed", "name", d.Name)
					return true
				}
			}
		}
	}
	return false
}

func (r *HelmReleaseReconciler) delete(ctx context.Context, hr *undistrov1.HelmRelease, h helm.Client) error {
	s, err := h.Status(hr.GetReleaseName(), helm.StatusOptions{
		Namespace: hr.GetTargetNamespace(),
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			hr.Status.Phase = undistrov1.HelmReleasePhaseUninstalled
			return nil
		}
		return err
	}
	if s == helm.StatusUninstalling {
		hr.Status.Phase = undistrov1.HelmReleasePhaseUninstalling
		return nil
	}
	err = h.Uninstall(hr.GetReleaseName(), helm.UninstallOptions{
		Namespace:   hr.GetTargetNamespace(),
		KeepHistory: false,
		Timeout:     hr.GetTimeout(),
	})
	if err != nil {
		return err
	}
	hr.Status.Phase = undistrov1.HelmReleasePhaseUninstalling
	return nil
}

// +kubebuilder:rbac:groups=undistro.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=undistro.io,resources=*,verbs=get;update;patch

func (r *HelmReleaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("helmrelease", req.NamespacedName)
	var hr undistrov1.HelmRelease
	if err := r.Get(ctx, req.NamespacedName, &hr); err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "couldn't get object", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(&hr, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer func() {
		err = patch.ControllerObject(ctx, patchHelper, &hr, err)
	}()
	if !controllerutil.ContainsFinalizer(&hr, undistrov1.Finalizer) {
		controllerutil.AddFinalizer(&hr, undistrov1.Finalizer)
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
		if !hr.DeletionTimestamp.IsZero() {
			controllerutil.RemoveFinalizer(&hr, undistrov1.Finalizer)
			return ctrl.Result{}, nil
		}
		log.Info("cluster is not ready yet", "namespaced name", nm.String())
		hr.Status.Phase = undistrov1.HelmReleasePhaseWaitClusterReady
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	if r.hasNonDeployedDeps(ctx, &hr) {
		log.Info("wating all dependencies to be deployed", "namespaced name", nm.String())
		hr.Status.Phase = undistrov1.HelmReleasePhaseWaitDependency
		return ctrl.Result{Requeue: true}, nil
	}
	wClient, err := r.clusterClient(ctx, wc, nm)
	if err != nil {
		return ctrl.Result{}, err
	}
	h, err := wc.GetHelm(nm.Name, nm.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}
	ch, err := helm.PrepareChart(h, &hr)
	if err != nil {
		log.Error(err, "couldn't prepare chart", "chartPath", ch.ChartPath, "revision", ch.Revision)
		record.Warn(&hr, "FailledDownloadChart", "chart template failed to downloaded")
		hr.Status.Phase = undistrov1.HelmReleasePhaseChartFetchFailed
		return ctrl.Result{}, err
	}
	record.Event(&hr, "DownloadChart", "chart template downloaded successfully")
	if !hr.DeletionTimestamp.IsZero() {
		log.Info("running uninstall")
		err = r.delete(ctx, &hr, h)
		if err != nil {
			return ctrl.Result{}, err
		}
		if hr.Status.Phase == undistrov1.HelmReleasePhaseUninstalled {
			controllerutil.RemoveFinalizer(&hr, undistrov1.Finalizer)
			os.RemoveAll(ch.ChartPath)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	}
	if ch.Changed {
		hr.Status.Phase = undistrov1.HelmReleasePhaseChartFetched
		hr.Status.LastAttemptedRevision = ch.Revision
		hr.Status.Revision = ch.Revision
		return ctrl.Result{Requeue: true}, nil
	}
	values, err := helm.ComposeValues(ctx, r.Client, &hr, ch.ChartPath)
	if err != nil {
		log.Error(err, "failed to compose values for release", "name", hr.Name)
		record.Warn(&hr, "FailledSetValues", "failed to set chart values")
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		return ctrl.Result{}, err
	}
	record.Event(&hr, "SetValues", "values setted successfully")
	curRel, err := h.Get(hr.GetReleaseName(), helm.GetOptions{Namespace: hr.GetTargetNamespace()})
	if err != nil {
		log.Error(err, "failed to get release", "name", hr.Name)
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		hr.Status.LastAttemptedRevision = ""
		return ctrl.Result{}, err
	}
	err = r.execObjs(ctx, wClient, hr.Spec.BeforeApplyObjects)
	if err != nil {
		record.Warn(&hr, "FailedExecBefore", "failed to apply objects before instalation")
		log.Error(err, "failed to exec before", "name", hr.Name)
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		hr.Status.LastAttemptedRevision = ""
		return ctrl.Result{}, err
	}
	status, _ := h.Status(hr.GetReleaseName(), helm.StatusOptions{
		Namespace: hr.GetTargetNamespace(),
	})
	rollback := false
	switch status {
	case helm.StatusPendingUpgrade, helm.StatusPendingRollback, helm.StatusPendingInstall:
		return ctrl.Result{Requeue: true}, nil
	default:
		if curRel == nil {
			record.Event(&hr, "InstallChart", "installing chart")
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
		} else {
			record.Event(&hr, "UpgradeChart", "upgrading chart")
			log.Info("running upgrade")
			_, err = h.UpgradeFromPath(ch.ChartPath, hr.GetReleaseName(), values, helm.UpgradeOptions{
				Namespace:         hr.GetTargetNamespace(),
				Timeout:           hr.GetTimeout(),
				Install:           false,
				Force:             hr.Spec.ForceUpgrade,
				SkipCRDs:          hr.Spec.SkipCRDs,
				MaxHistory:        hr.GetMaxHistory(),
				Wait:              hr.GetWait(),
				DisableValidation: false,
			})
		}
		if err != nil {
			if curRel == nil {
				log.Error(err, "fail to install")
				hr.Status.Phase = undistrov1.HelmReleasePhaseDeployFailed
				hr.Status.Revision = ""
				hr.Status.LastAttemptedRevision = ""
			}
			record.Event(&hr, "RollbackChart", "rollingback chart")
			rollback = true
			_, err = h.Rollback(hr.GetReleaseName(), helm.RollbackOptions{
				Namespace:    hr.GetTargetNamespace(),
				Timeout:      hr.Spec.Rollback.GetTimeout(),
				Wait:         hr.Spec.Rollback.Wait,
				DisableHooks: hr.Spec.Rollback.DisableHooks,
				Recreate:     hr.Spec.Rollback.Recreate,
				Force:        hr.Spec.Rollback.Force,
			})
			if err != nil {
				hr.Status.Phase = undistrov1.HelmReleasePhaseRollbackFailed
				hr.Status.Revision = ""
				hr.Status.LastAttemptedRevision = ""
				return ctrl.Result{}, err
			}

		}
	}
	status, err = h.Status(hr.GetReleaseName(), helm.StatusOptions{
		Namespace: hr.GetTargetNamespace(),
	})
	if err != nil {
		log.Error(err, "fail to get status")
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		hr.Status.LastAttemptedRevision = ""
		return ctrl.Result{}, err
	}
	err = r.execObjs(ctx, wClient, hr.Spec.AfterApplyObjects)
	if err != nil {
		record.Warn(&hr, "FailedExecAfter", "failed to apply objects after instalation")
		log.Error(err, "failed to exec after", "name", hr.Name)
		hr.Status.Phase = undistrov1.HelmReleasePhaseFailed
		return ctrl.Result{}, err
	}
	if rollback {
		record.Event(&hr, "ChartRolledBack", "chart is rolledback")
		hr.Status.Phase = undistrov1.HelmReleasePhaseRolledBack
	} else {
		record.Event(&hr, "ChartDeployed", "chart is deployed")
		hr.Status.Phase = undistrov1.HelmReleasePhaseDeployed
	}
	hr.Status.ReleaseName = hr.GetReleaseName()
	hr.Status.ReleaseStatus = status.String()

	return ctrl.Result{}, nil

}

func (r *HelmReleaseReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, opts controller.Options) error {
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
	if !controllerutil.ContainsFinalizer(oldHr, undistrov1.Finalizer) && controllerutil.ContainsFinalizer(newHr, undistrov1.Finalizer) {
		return true
	}
	switch helm.Status(newHr.Status.ReleaseStatus) {
	case helm.StatusPendingUpgrade, helm.StatusPendingRollback, helm.StatusPendingInstall:
		return true
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
