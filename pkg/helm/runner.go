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
package helm

import (
	"errors"
	"fmt"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Runner represents a Helm action runner capable of performing Helm
// operations for a v2beta1.HelmRelease.
type Runner struct {
	config *action.Configuration
}

// NewRunner constructs a new Runner configured to run Helm actions with the
// given genericclioptions.RESTClientGetter, and the release and storage
// namespace configured to the provided values.
func NewRunner(getter genericclioptions.RESTClientGetter, storageNamespace string, logger logr.Logger) (*Runner, error) {
	cfg := new(action.Configuration)
	if err := cfg.Init(getter, storageNamespace, "secret", debugLogger(logger)); err != nil {
		return nil, err
	}
	return &Runner{config: cfg}, nil
}

// Install runs an Helm install action for the given v2beta1.HelmRelease.
func (r *Runner) Install(hr appv1alpha1.HelmRelease, chart *chart.Chart, values chartutil.Values) (*release.Release, error) {
	install := action.NewInstall(r.config)
	install.ReleaseName = hr.Spec.ReleaseName
	install.Namespace = hr.Spec.TargetNamespace
	install.Timeout = hr.Spec.Timeout.Duration
	install.Wait = *hr.Spec.Wait
	install.SkipCRDs = hr.Spec.SkipCRDs
	install.DependencyUpdate = true
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>", hr.Spec.TargetNamespace)
	return install.Run(chart, values.AsMap())
}

// Upgrade runs an Helm upgrade action for the given v2beta1.HelmRelease.
func (r *Runner) Upgrade(hr appv1alpha1.HelmRelease, chart *chart.Chart, values chartutil.Values) (*release.Release, error) {
	upgrade := action.NewUpgrade(r.config)
	upgrade.Namespace = hr.Spec.TargetNamespace
	upgrade.ResetValues = *hr.Spec.ResetValues
	upgrade.ReuseValues = !*hr.Spec.ResetValues
	upgrade.MaxHistory = *hr.Spec.MaxHistory
	upgrade.Timeout = hr.Spec.Timeout.Duration
	upgrade.Wait = *hr.Spec.Wait
	upgrade.Force = hr.Spec.ForceUpgrade
	upgrade.CleanupOnFail = true
	return upgrade.Run(hr.Spec.ReleaseName, chart, values.AsMap())
}

// Test runs an Helm test action for the given v2beta1.HelmRelease.
func (r *Runner) Test(hr appv1alpha1.HelmRelease) (*release.Release, error) {
	test := action.NewReleaseTesting(r.config)
	test.Namespace = hr.Spec.TargetNamespace
	test.Timeout = hr.Spec.Test.Timeout.Duration

	return test.Run(hr.Spec.ReleaseName)
}

// Rollback runs an Helm rollback action for the given v2beta1.HelmRelease.
func (r *Runner) Rollback(hr appv1alpha1.HelmRelease) error {
	rollback := action.NewRollback(r.config)
	rollback.Timeout = hr.Spec.Rollback.Timeout.Duration
	rollback.Wait = hr.Spec.Rollback.Wait
	rollback.DisableHooks = hr.Spec.Rollback.DisableHooks
	rollback.Force = hr.Spec.Rollback.Force
	rollback.Recreate = hr.Spec.Rollback.Recreate
	rollback.CleanupOnFail = true

	return rollback.Run(hr.Spec.ReleaseName)
}

// Uninstall runs an Helm uninstall action for the given v2beta1.HelmRelease.
func (r *Runner) Uninstall(hr appv1alpha1.HelmRelease) error {
	uninstall := action.NewUninstall(r.config)
	uninstall.Timeout = hr.Spec.Timeout.Duration
	uninstall.DisableHooks = false
	uninstall.KeepHistory = *hr.Spec.MaxHistory > 0

	_, err := uninstall.Run(hr.Spec.ReleaseName)
	return err
}

func (r *Runner) Status(hr appv1alpha1.HelmRelease) (*release.Release, error) {
	status := action.NewStatus(r.config)
	rel, err := status.Run(hr.Spec.ReleaseName)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

func (r *Runner) List() ([]*release.Release, error) {
	list := action.NewList(r.config)
	list.AllNamespaces = true
	list.All = true
	return list.Run()
}

// ObserveLastRelease observes the last revision, if there is one,
// for the actual Helm release associated with the given v2beta1.HelmRelease.
func (r *Runner) ObserveLastRelease(hr appv1alpha1.HelmRelease) (*release.Release, error) {
	rel, err := r.config.Releases.Last(hr.Spec.ReleaseName)
	if err != nil && errors.Is(err, driver.ErrReleaseNotFound) {
		err = nil
	}
	return rel, err
}

func debugLogger(logger logr.Logger) func(format string, v ...interface{}) {
	return func(format string, v ...interface{}) {
		logger.V(1).Info(fmt.Sprintf(format, v...))
	}
}
