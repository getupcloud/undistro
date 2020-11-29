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

package v1alpha1

import (
	"fmt"
	"time"

	"github.com/getupio-undistro/undistro/pkg/meta"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var helmreleaselog = logf.Log.WithName("helmrelease-resource")

func (r *HelmRelease) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-app-undistro-io-v1alpha1-helmrelease,mutating=true,failurePolicy=fail,groups=app.undistro.io,resources=helmreleases,verbs=create;update;delete,versions=v1alpha1,name=mhelmrelease.undistro.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &HelmRelease{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *HelmRelease) Default() {
	helmreleaselog.Info("default", "name", r.Name)
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[meta.LabelUndistroClusterName] = r.Spec.ClusterName
	if r.Spec.ClusterName == "" {
		r.Labels[meta.LabelUndistroClusterType] = "management"
	} else {
		r.Labels[meta.LabelUndistroClusterType] = "workload"
	}
	defaultTimeout := &metav1.Duration{
		Duration: 300 * time.Second,
	}
	fmt.Println(">>>>>>>>>>>>>>>>>>>>> default webhook", r.Namespace)
	if r.Namespace == "" {
		r.Namespace = "default"
	}
	if r.Spec.TargetNamespace == "" {
		r.Spec.TargetNamespace = r.Namespace
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
	for i := range r.Spec.ValuesFrom {
		if r.Spec.ValuesFrom[i].ValuesKey == "" {
			r.Spec.ValuesFrom[i].ValuesKey = "values.yaml"
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-app-undistro-io-v1alpha1-helmrelease,mutating=false,failurePolicy=fail,groups=app.undistro.io,resources=helmreleases,versions=v1alpha1,name=vhelmrelease.undistro.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Validator = &HelmRelease{}

func (r *HelmRelease) validate(old *HelmRelease) error {
	var allErrs field.ErrorList
	if r.Spec.Chart.Name == "" {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec", "chart", "name"),
			"spec.chart.name to be populated",
		))
	}
	if r.Spec.Chart.Version == "" {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec", "chart", "version"),
			"spec.chart.version to be populated",
		))
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