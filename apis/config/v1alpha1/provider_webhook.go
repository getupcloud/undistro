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

	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/version"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var providerlog = logf.Log.WithName("provider-resource")

const defaultRepo = "https://charts.undistro.io"

func (r *Provider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-config-undistro-io-v1alpha1-provider,mutating=true,failurePolicy=fail,groups=config.undistro.io,resources=providers,verbs=create;update,versions=v1alpha1,name=mprovider.undistro.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &Provider{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Provider) Default() {
	providerlog.Info("default", "name", r.Name)
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[meta.LabelUndistro] = ""
	r.Labels[meta.LabelUndistroClusterName] = ""
	r.Labels[meta.LabelUndistroClusterType] = "management"
	if r.Spec.Repository.URL == "" {
		r.Spec.Repository.URL = defaultRepo
	}
	if r.Spec.ProviderName == "undistro" && r.Spec.Repository.URL == defaultRepo {
		r.Labels[meta.LabelProviderType] = string(CoreProviderType)
	}
	if r.Spec.ProviderType == "" {
		r.Spec.ProviderType = string(InfraProviderType)
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-config-undistro-io-v1alpha1-provider,mutating=false,failurePolicy=fail,groups=config.undistro.io,resources=providers,versions=v1alpha1,name=vprovider.undistro.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Validator = &Provider{}

func (r *Provider) validate(old *Provider) error {
	var allErrs field.ErrorList
	if r.Spec.ProviderName == "" {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec", "providerName"),
			"spec.providerName to be populated",
		))
	}
	if r.Spec.Repository.URL == "" {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec", "repository", "url"),
			"spec.repository.url to be populated",
		))
	}
	if old != nil && old.Spec.ProviderName != r.Spec.ProviderName {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "providerName"),
			r.Spec.ProviderName,
			"field is immutable",
		))
	}
	if old != nil && old.Spec.Repository.URL != r.Spec.Repository.URL {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "repository", "url"),
			r.Spec.Repository.URL,
			"field is immutable",
		))
	}
	_, err := version.ParseVersion(r.Spec.ProviderVersion)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "providerVersion"),
			r.Spec.ProviderVersion,
			err.Error(),
		))
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("HelmRelease").GroupKind(), r.Name, allErrs)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Provider) ValidateCreate() error {
	providerlog.Info("validate create", "name", r.Name)
	return r.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Provider) ValidateUpdate(old runtime.Object) error {
	providerlog.Info("validate update", "name", r.Name)
	oldP, ok := old.(*Provider)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Provider but got a %T", old))
	}
	return r.validate(oldP)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Provider) ValidateDelete() error {
	providerlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
