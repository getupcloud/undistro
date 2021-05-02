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
	"github.com/getupio-undistro/undistro/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var defaultpolicieslog = logf.Log.WithName("defaultpolicies-resource")

func (r *DefaultPolicies) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-app-undistro-io-v1alpha1-defaultpolicies,mutating=true,failurePolicy=fail,groups=app.undistro.io,resources=defaultpolicies,verbs=create;update,versions=v1alpha1,name=mdefaultpolicies.kb.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &DefaultPolicies{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *DefaultPolicies) Default() {
	defaultpolicieslog.Info("default", "name", r.Name)
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
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:verbs=create;update,path=/validate-app-undistro-io-v1alpha1-defaultpolicies,mutating=false,failurePolicy=fail,groups=app.undistro.io,resources=defaultpolicies,versions=v1alpha1,name=vdefaultpolicies.kb.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Validator = &DefaultPolicies{}

func (r *DefaultPolicies) validate(old *DefaultPolicies) error {
	var allErrs field.ErrorList
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
	return apierrors.NewInvalid(GroupVersion.WithKind("DefaultPolicies").GroupKind(), r.Name, allErrs)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DefaultPolicies) ValidateCreate() error {
	defaultpolicieslog.Info("validate create", "name", r.Name)
	return r.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DefaultPolicies) ValidateUpdate(old runtime.Object) error {
	defaultpolicieslog.Info("validate update", "name", r.Name)
	oldDp, ok := old.(*DefaultPolicies)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a DefaultPolicies but got a %T", old))
	}
	return r.validate(oldDp)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DefaultPolicies) ValidateDelete() error {
	defaultpolicieslog.Info("validate delete", "name", r.Name)
	return nil
}
