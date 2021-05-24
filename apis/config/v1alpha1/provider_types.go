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
	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/meta"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProviderSpec defines the desired state of Provider
type ProviderSpec struct {
	// +kubebuilder:default=false
	Paused            bool                          `json:"paused,omitempty"`
	ProviderName      string                        `json:"providerName,omitempty"`
	ProviderVersion   string                        `json:"providerVersion,omitempty"`
	ProviderType      string                        `json:"providerType,omitempty"`
	Repository        Repository                    `json:"repository,omitempty"`
	ConfigurationFrom []appv1alpha1.ValuesReference `json:"configurationFrom,omitempty"`
	Configuration     *apiextensionsv1.JSON         `json:"configuration,omitempty"`
	// +kubebuilder:default=false
	AutoUpgrade bool `json:"autoUpgrade,omitempty"`
}

type Repository struct {
	// +kubebuilder:default="https://charts.undistro.io"
	URL       string                       `json:"url,omitempty"`
	SecretRef *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// ProviderStatus defines the observed state of Provider
type ProviderStatus struct {
	// ObservedGeneration is the last observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions           []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	HelmReleaseName      string             `json:"helmReleaseName,omitempty"`
	LastAppliedVersion   string             `json:"lastAppliedVersion,omitempty"`
	LastAttemptedVersion string             `json:"lastAttemptedVersion,omitempty"`
}

// ProviderProgressing resets any failures and registers progress toward
// reconciling the given Provider by setting the meta.ReadyCondition to
// 'Unknown' for meta.ProgressingReason.
func ProviderProgressing(p Provider) Provider {
	p.Status.Conditions = []metav1.Condition{}
	msg := "Reconciliation in progress"
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionUnknown, meta.ProgressingReason, msg)
	return p
}

// ProviderNotReady registers a failed reconciliation of the given Provider.
func ProviderNotReady(p Provider, reason, message string) Provider {
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionFalse, reason, message)
	return p
}

// ProviderReady registers a successful reconciliation of the given Provider.
func ProviderReady(p Provider) Provider {
	msg := "Provider reconciliation succeeded"
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionTrue, meta.ReconciliationSucceededReason, msg)
	p.Status.LastAppliedVersion = p.Status.LastAttemptedVersion
	return p
}

// ProviderAttempted registers an attempt of the given Provider with the given state.
// and returns the modified Provider and a boolean indicating a state change.
func ProviderAttempted(p Provider, releaseName, version string) Provider {
	p.Status.HelmReleaseName = releaseName
	p.Status.LastAppliedVersion = version
	return p
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=providers,scope=Namespaced
// +kubebuilder:printcolumn:name="Provider Name",type="string",JSONPath=".spec.providerName"
// +kubebuilder:printcolumn:name="Provider Version",type="string",JSONPath=".spec.providerVersion"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// Provider is the Schema for the providers API
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec   `json:"spec,omitempty"`
	Status ProviderStatus `json:"status,omitempty"`
}

func (p *Provider) GetStatusConditions() *[]metav1.Condition {
	return &p.Status.Conditions
}

func (p *Provider) GetNamespace() string {
	if p.Namespace == "" {
		return "default"
	}
	return p.Namespace
}

// +kubebuilder:object:root=true

// ProviderList contains a list of Provider
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
