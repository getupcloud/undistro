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
	"github.com/getupio-undistro/undistro/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultPoliciesSpec defines the desired state of DefaultPolicies
type DefaultPoliciesSpec struct {
	Paused          bool     `json:"paused,omitempty"`
	ClusterName     string   `json:"clusterName,omitempty"`
	ExcludePolicies []string `json:"excludePolicies,omitempty"`
}

// DefaultPoliciesStatus defines the observed state of DefaultPolicies
type DefaultPoliciesStatus struct {
	// ObservedGeneration is the last observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	AppliedPolicies []string `json:"appliedPolicies,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// DefaultPolicies is the Schema for the defaultpolicies API
type DefaultPolicies struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DefaultPoliciesSpec   `json:"spec,omitempty"`
	Status DefaultPoliciesStatus `json:"status,omitempty"`
}

func (p *DefaultPolicies) GetStatusConditions() *[]metav1.Condition {
	return &p.Status.Conditions
}

func DefaultPoliciesNotReady(p DefaultPolicies, reason, message string) DefaultPolicies {
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionFalse, reason, message)
	return p
}

func DefaultPoliciesReady(p DefaultPolicies) DefaultPolicies {
	msg := "Default policies reconciliation succeeded"
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionTrue, meta.ReconciliationSucceededReason, msg)
	return p
}

//+kubebuilder:object:root=true

// DefaultPoliciesList contains a list of DefaultPolicies
type DefaultPoliciesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DefaultPolicies `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DefaultPolicies{}, &DefaultPoliciesList{})
}
