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
	"reflect"

	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/version"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

func (r *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-app-undistro-io-v1alpha1-cluster,mutating=true,failurePolicy=fail,groups=app.undistro.io,resources=clusters,verbs=create;update,versions=v1alpha1,name=mcluster.undistro.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Defaulter = &Cluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cluster) Default() {
	clusterlog.Info("default", "name", r.Name)
	if r.Labels == nil {
		r.Labels = make(map[string]string)
	}
	r.Labels[meta.LabelUndistro] = ""
	r.Labels[meta.LabelUndistroClusterName] = r.Name
	r.Labels[meta.LabelUndistroClusterType] = "workload"
	if r.Spec.ControlPlane == nil {
		r.Spec.ControlPlane = &ControlPlaneNode{}
	}
	bastionEnabled := true
	if r.Spec.Bastion == nil {
		r.Spec.Bastion = &Bastion{
			Enabled:             &bastionEnabled,
			DisableIngressRules: true,
		}
	}
	if r.Spec.Bastion.Enabled == nil {
		r.Spec.Bastion.Enabled = &bastionEnabled
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-app-undistro-io-v1alpha1-cluster,mutating=false,failurePolicy=fail,groups=app.undistro.io,resources=clusters,versions=v1alpha1,name=vcluster.undistro.io,sideEffects=None,admissionReviewVersions=v1beta1

var _ webhook.Validator = &Cluster{}

func (r *Cluster) validate(old *Cluster) error {
	var allErrs field.ErrorList
	if r.Spec.ControlPlane == nil && !r.Spec.InfrastructureProvider.Managed {
		allErrs = append(allErrs, field.Required(
			field.NewPath("spec", "controlPlane"),
			"controlPlane must to be populated when is a self hosted cluster",
		))
	}
	_, err := version.ParseVersion(r.Spec.KubernetesVersion)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec", "kubernetesVersion"),
			r.Spec.KubernetesVersion,
			"kubernetesVersion must to be a semantic versioning",
		))
	}
	if old != nil && r.Spec.ControlPlane != nil && !r.Spec.InfrastructureProvider.Managed {
		if !reflect.DeepEqual(old.Spec.ControlPlane.Endpoint, capi.APIEndpoint{}) &&
			!reflect.DeepEqual(r.Spec.ControlPlane.Endpoint, old.Spec.ControlPlane.Endpoint) {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec", "controlPlane", "endpoint"),
				r.Spec.ControlPlane.Endpoint,
				"field is immutable",
			))
		}
	}
	if old != nil {
		if !reflect.DeepEqual(old.Spec.ControlPlane.Endpoint, capi.APIEndpoint{}) && !reflect.DeepEqual(old.Spec.Network.ClusterNetwork, capi.ClusterNetwork{}) {
			if !meta.InReadyCondition(r.Status.Conditions) {
				return apierrors.NewBadRequest("can't update cluster that isn't ready")
			}
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("Cluster").GroupKind(), r.Name, allErrs)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateCreate() error {
	clusterlog.Info("validate create", "name", r.Name)
	return r.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateUpdate(old runtime.Object) error {
	clusterlog.Info("validate update", "name", r.Name)
	oldCl, ok := old.(*Cluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Cluster but got a %T", old))
	}
	return r.validate(oldCl)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateDelete() error {
	clusterlog.Info("validate delete", "name", r.Name)
	return nil
}
