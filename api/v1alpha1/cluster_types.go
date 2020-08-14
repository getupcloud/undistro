/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node struct {
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`
	// +kubebuilder:validation:MinLength=1
	MachineType string `json:"machineType,omitempty"`
}

type InfrastructureProvider struct {
	// +kubebuilder:validation:MinLength=1
	Name    string  `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
	// +kubebuilder:validation:UniqueItems=true
	Env []corev1.EnvVar `json:"env,omitempty"`
}

func (i *InfrastructureProvider) NameVersion() string {
	if i.Version != nil {
		return fmt.Sprintf("%s:%s", i.Name, *i.Version)
	}
	return i.Name
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// +kubebuilder:validation:MinLength=1
	KubernetesVersion      string                 `json:"kubernetesVersion,omitempty"`
	InfrastructureProvider InfrastructureProvider `json:"infrastructureProvider,omitempty"`
	ControlPlaneNode       Node                   `json:"controlPlaneNode,omitempty"`
	WorkerNode             Node                   `json:"workerNode,omitempty"`
}

type InstalledComponent struct {
	Name    string       `json:"name,omitempty"`
	Version string       `json:"version,omitempty"`
	URL     string       `json:"url,omitempty"`
	Type    ProviderType `json:"type,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	Phase               ClusterPhase         `json:"phase,omitempty"`
	InstalledComponents []InstalledComponent `json:"installedComponents,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clusters,shortName=cl,scope=Namespaced,categories=undistro
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Infra",type="string",JSONPath=".spec.infrastructureProvider.name"
// +kubebuilder:printcolumn:name="K8s",type="string",JSONPath=".spec.kubernetesVersion"
// +kubebuilder:printcolumn:name="Control Plane Replicas",type="integer",JSONPath=".spec.controlPlaneNode.replicas"
// +kubebuilder:printcolumn:name="Worker Replicas",type="integer",JSONPath=".spec.workerNode.replicas"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
