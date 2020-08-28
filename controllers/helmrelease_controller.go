/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package controllers

import (
	"context"
	"strings"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/client/cluster/helm"
	"github.com/getupcloud/undistro/status"
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

type errCollection []error

func (err errCollection) Error() string {
	var errs []string
	for i := len(err) - 1; i >= 0; i-- {
		errs = append(errs, err[i].Error())
	}
	return strings.Join(errs, ", previous error:")
}

func (err errCollection) Empty() bool {
	return len(err) == 0
}

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
	defer status.SetObservedGeneration(ctx, r.Client, &hr, hr.Generation)
	ch, err := helm.PrepareChart(h, &hr)
	if err != nil {
		log.Error(err, "couldn't prepare chart", "chartPath", ch.ChartPath, "revision", ch.Revision)
		err = status.SetStatusPhase(ctx, r.Client, &hr, undistrov1.HelmReleasePhaseChartFetchFailed)
		if err != nil {
			log.Error(err, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}
	if ch.Changed {
		err = status.SetStatusPhase(ctx, r.Client, &hr, undistrov1.HelmReleasePhaseChartFetched)
		if err != nil {
			log.Error(err, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}
	values, err := helm.ComposeValues(ctx, r.Client, &hr, ch.ChartPath)
	if err != nil {
		log.Error(err, "failed to compose values for release", "name", hr.Name)
		err = status.SetStatusPhase(ctx, r.Client, &hr, undistrov1.HelmReleasePhaseFailed)
		if err != nil {
			log.Error(err, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}
	action, curRel, err := helm.ChooseAction(ctx, h, &hr, ch, status.HasSynced(&hr), status.HasRolledBack(&hr), status.ShouldRetryUpgrade(&hr))
	if err != nil {
		log.Error(err, "couldn't to choose action for release", "name", hr.Name)
		err = status.SetStatusPhase(ctx, r.Client, &hr, undistrov1.HelmReleasePhaseFailed)
		if err != nil {
			log.Error(err, "couldn't update status", "name", req.NamespacedName)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, r.run(ctx, h, action, &hr, curRel, ch, values)

}

func (r *HelmReleaseReconciler) run(
	ctx context.Context,
	h helm.Client,
	action undistrov1.HelmAction,
	hr *undistrov1.HelmRelease,
	curRel *helm.Release,
	chart helm.ChartState,
	values []byte) error {
	log := r.Log
	var newRel *helm.Release
	errs := errCollection{}
next:
	var err error
	switch action {
	case undistrov1.DryRunCompareAction:
		log.Info("running dry-run upgrade to compare with release", "version", curRel.Version, "action", action)
		var diff string
		newRel, err = h.UpgradeFromPath(chart.ChartPath, hr.GetReleaseName(), values, helm.UpgradeOptions{
			DryRun:      true,
			Namespace:   hr.GetTargetNamespace(),
			Force:       hr.Spec.ForceUpgrade,
			ResetValues: hr.GetReuseValues(),
			ReuseValues: !hr.GetReuseValues(),
		})
		if err != nil {
			log.Error(err, "couldn't execute dryRun for release", "name", hr.GetReleaseName())
			errs = append(errs, err)
			err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseFailed)
			if err != nil {
				log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
			}
			break
		}
		diff = helm.Diff(curRel, newRel)
		if diff != "" {
			action = undistrov1.UpgradeAction
			goto next
		}
		if !status.HasRolledBack(hr) {
			err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseSucceeded)
			if err != nil {
				log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
			}
		}
		log.Info("no changes", "action", action)
	case undistrov1.InstallAction:
		log.Info("running instalation", "action", action)
		err = status.SetStatusPhaseWithRevision(ctx, r.Client, hr, undistrov1.HelmReleasePhaseInstalling, chart.Revision)
		if err != nil {
			log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
		}
		_, err = h.UpgradeFromPath(chart.ChartPath, hr.GetReleaseName(), values, helm.UpgradeOptions{
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
			log.Error(err, "couldn't execute install for release", "name", hr.GetReleaseName())
			errs = append(errs, err)
			err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseDeployFailed)
			if err != nil {
				log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
			}
			action = undistrov1.UninstallAction
			goto next
		}
		err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseDeployed)
		if err != nil {
			log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
		}
		log.Info("installation succeeded", "revision", chart.Revision, "action", action)
		action = undistrov1.TestAction
		goto next
	case undistrov1.UpgradeAction:
		log.Info("running upgrade", "action", action)
		err = status.SetStatusPhaseWithRevision(ctx, r.Client, hr, undistrov1.HelmReleasePhaseUpgrading, chart.Revision)
		if err != nil {
			log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
		}
		_, err = h.UpgradeFromPath(chart.ChartPath, hr.GetReleaseName(), values, helm.UpgradeOptions{
			Namespace:         hr.GetTargetNamespace(),
			Timeout:           hr.GetTimeout(),
			Install:           false,
			Force:             hr.Spec.ForceUpgrade,
			ReuseValues:       hr.GetReuseValues(),
			ResetValues:       !hr.GetReuseValues(),
			SkipCRDs:          hr.Spec.SkipCRDs,
			MaxHistory:        hr.GetMaxHistory(),
			Wait:              hr.GetWait(),
			DisableValidation: false,
		})
		if err != nil {
			log.Error(err, "couldn't execute upgrade for release", "name", hr.GetReleaseName())
			errs = append(errs, err)
			err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseDeployFailed)
			if err != nil {
				log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
			}
			action = undistrov1.RollbackAction
			goto next
		}
		err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseDeployed)
		if err != nil {
			log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
		}
		log.Info("upgrade succeeded", "revision", chart.Revision, "action", action)
		action = undistrov1.TestAction
		goto next
	case undistrov1.TestAction:
		if hr.Spec.Test.Enable {
			log.Info("running test", "action", action)
			err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseTesting)
			if err != nil {
				log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
			}
			err = h.Test(hr.GetReleaseName(), helm.TestOptions{
				Namespace: hr.GetTargetNamespace(),
				Timeout:   hr.Spec.Test.GetTimeout(),
				Cleanup:   hr.Spec.Test.GetCleanup(),
			})
			if err != nil {
				log.Error(err, "couldn't execute test for release", "name", hr.GetReleaseName())
				errs = append(errs, err)
				err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseTestFailed)
				if err != nil {
					log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
				}
				if !hr.Spec.Test.GetIgnoreFailures() {
					if curRel == nil {
						action = undistrov1.UninstallAction
					} else {
						action = undistrov1.RollbackAction
					}
					goto next
				} else {
					log.Info("test failed - ignoring failures", "revision", chart.Revision)
				}
			}
		} else {
			log.Info("test succeeded", "revision", chart.Revision, "action", action)
		}
		err = status.SetStatusPhaseWithRevision(ctx, r.Client, hr, undistrov1.HelmReleasePhaseSucceeded, chart.Revision)
		if err != nil {
			log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
		}
	case undistrov1.RollbackAction:
		if hr.Spec.Rollback.Enable {
			latestRel, err := h.Get(hr.GetReleaseName(), helm.GetOptions{Namespace: hr.GetTargetNamespace(), Version: 0})
			if err != nil {
				errs = append(errs, err)
				break
			}
			if curRel.Version < latestRel.Version {
				log.Info("running rollback", "action", action)
				err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseRollingBack)
				if err != nil {
					log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
				}
				_, err = h.Rollback(hr.GetReleaseName(), helm.RollbackOptions{
					Namespace:    hr.GetTargetNamespace(),
					Timeout:      hr.Spec.Rollback.GetTimeout(),
					Wait:         hr.Spec.Rollback.Wait,
					DisableHooks: hr.Spec.Rollback.DisableHooks,
					Recreate:     hr.Spec.Rollback.Recreate,
					Force:        hr.Spec.Rollback.Force,
				})
				if err != nil {
					errs = append(errs, err)
					err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseRollbackFailed)
					if err != nil {
						log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
					}
					break
				}
				log.Info("rollback succeeded", "action", action)
				err = status.SetStatusPhase(ctx, r.Client, hr, undistrov1.HelmReleasePhaseRolledBack)
				if err != nil {
					log.Error(err, "couldn't update status", "name", hr.GetReleaseName())
				}
			}
		}
	case undistrov1.UninstallAction:
		log.Info("running uninstall", "action", action)
		err = h.Uninstall(hr.GetReleaseName(), helm.UninstallOptions{
			Namespace:   hr.GetTargetNamespace(),
			KeepHistory: false,
			Timeout:     hr.GetTimeout(),
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	if errs.Empty() {
		return nil
	}
	return errs
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
	n, ok := e.ObjectNew.(*undistrov1.HelmRelease)
	if !ok {
		return false
	}
	o, ok := e.ObjectOld.(*undistrov1.HelmRelease)
	if !ok {
		return false
	}
	diff := cmp.Diff(o.Spec, n.Spec)
	if sDiff := cmp.Diff(o.Status, n.Status); diff == "" && sDiff != "" {
		return false
	}
	return true
}
