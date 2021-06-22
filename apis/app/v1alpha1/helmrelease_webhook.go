/*
Copyright 2020-2021 The UnDistro authors

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

package v1alpha1

import (
	"context"
	"fmt"
	"time"

	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/getupio-undistro/undistro/pkg/version"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var helmreleaselog = logf.Log.WithName("helmrelease-resource")

func (r *HelmRelease) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if k8sClient == nil {
		k8sClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-app-undistro-io-v1alpha1-helmrelease,mutating=true,failurePolicy=fail,sideEffects=None,groups=app.undistro.io,resources=helmreleases,verbs=create;update,versions=v1alpha1,name=mhelmrelease.undistro.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &HelmRelease{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HelmRelease) Default() {
	helmreleaselog.Info("default", "name", r.Name)
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[meta.LabelUndistro] = ""
	key := util.ObjectKeyFromString(r.Spec.ClusterName)
	r.Labels[meta.LabelUndistroClusterName] = key.Name
	r.Labels[capi.ClusterLabelName] = r.Name
	if r.Spec.ClusterName == "" {
		r.Labels[meta.LabelUndistroClusterType] = "management"
	} else {
		r.Labels[meta.LabelUndistroClusterType] = "workload"
	}
	defaultTimeout := &metav1.Duration{
		Duration: 300 * time.Second,
	}
	if r.Spec.TargetNamespace == "" {
		r.Spec.TargetNamespace = r.GetNamespace()
	}
	if r.Spec.ReleaseName == "" {
		r.Spec.ReleaseName = fmt.Sprintf("%s-%s", r.Spec.TargetNamespace, r.Name)
	}
	if r.Spec.Timeout == nil {
		r.Spec.Timeout = defaultTimeout
	}
	if r.Spec.Test.Timeout == nil {
		r.Spec.Test.Timeout = defaultTimeout
	}
	if r.Spec.Rollback.Timeout == nil {
		r.Spec.Rollback.Timeout = defaultTimeout
	}
	if r.Spec.Wait == nil {
		wait := true
		r.Spec.Wait = &wait
	}
	if r.Spec.ResetValues == nil {
		reset := false
		r.Spec.ResetValues = &reset
	}
	if r.Spec.MaxHistory == nil {
		h := 10
		r.Spec.MaxHistory = &h
	}
	if r.Spec.ForceUpgrade == nil {
		force := true
		r.Spec.ForceUpgrade = &force
	}
	for i := range r.Spec.ValuesFrom {
		if r.Spec.ValuesFrom[i].ValuesKey == "" {
			r.Spec.ValuesFrom[i].ValuesKey = "values.yaml"
		}
	}
}

//+kubebuilder:webhook:path=/validate-app-undistro-io-v1alpha1-helmrelease,mutating=false,failurePolicy=fail,sideEffects=None,groups=app.undistro.io,resources=helmreleases,verbs=create;update,versions=v1alpha1,name=vhelmrelease.undistro.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &HelmRelease{}

func (r *HelmRelease) validate(old *HelmRelease) error {
	var allErrs field.ErrorList
	if r.Spec.Chart.Name == "" {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec", "chart", "name"),
			"spec.chart.name to be populated",
		))
	}
	for _, d := range r.Spec.Dependencies {
		if d.Name == "" {
			allErrs = append(allErrs, field.Required(
				field.NewPath("spec", "dependencies[]", "name"),
				"spec.dependencies[].name to be populated",
			))
		}
	}
	if r.Spec.Chart.RepoURL == "" {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec", "chart", "repository"),
			"spec.chart.repository to be populated",
		))
	}
	if old != nil && old.Spec.Chart.Name != r.Spec.Chart.Name {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "chart", "name"),
			r.Spec.Chart.Name,
			"field is immutable",
		))
	}
	if old != nil && old.Spec.Chart.RepoURL != r.Spec.Chart.RepoURL {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "chart", "repository"),
			r.Spec.Chart.RepoURL,
			"field is immutable",
		))
	}
	if old != nil && old.Spec.ClusterName != r.Spec.ClusterName {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "clusterName"),
			r.Spec.ClusterName,
			"field is immutable",
		))
	}
	_, err := version.ParseVersion(r.Spec.Chart.Version)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "chart", "version"),
			r.Spec.Chart.Version,
			err.Error(),
		))
	}
	if r.Spec.ClusterName != "" {
		cl := Cluster{}
		key := util.ObjectKeyFromString(r.Spec.ClusterName)
		err := k8sClient.Get(context.TODO(), key, &cl)
		if err != nil {
			allErrs = append(allErrs, field.NotFound(
				field.NewPath("spec", "clusterName"),
				r.Spec.ClusterName,
			))
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("HelmRelease").GroupKind(), r.Name, allErrs)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HelmRelease) ValidateCreate() error {
	helmreleaselog.Info("validate create", "name", r.Name)
	return r.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HelmRelease) ValidateUpdate(old runtime.Object) error {
	helmreleaselog.Info("validate update", "name", r.Name)
	oldHr, ok := old.(*HelmRelease)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a HelmRelease but got a %T", old))
	}
	return r.validate(oldHr)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HelmRelease) ValidateDelete() error {
	helmreleaselog.Info("validate delete", "name", r.Name)
	return nil
}
