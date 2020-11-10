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
	Replicas *int64 `json:"replicas,omitempty"`
	// +kubebuilder:validation:MinLength=1
	MachineType string `json:"machineType,omitempty"`
	Subnet      string `json:"subnet,omitempty"`
}

type ControlPlaneNode struct {
	Node       `json:",inline,omitempty"`
	EndpointIP string `json:"endpointIP,omitempty"`
	InternalLB bool   `json:"internalLB,omitempty"`
}

type WorkerNode struct {
	Node      `json:",inline,omitempty"`
	Autoscale Autoscale `json:"autoscale,omitempty"`
}

type Autoscale struct {
	Enabled bool `json:"enabled,omitempty"`
	// The minimum size of the group.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	MinSize int64 `json:"minSize,omitempty"`

	// The maximum size of the group.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	MaxSize int64 `json:"maxSize,omitempty"`
}

type InfrastructureProvider struct {
	// +kubebuilder:validation:MinLength=1
	Name    string  `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
	File    *string `json:"file,omitempty"`
	Managed bool    `json:"managed,omitempty"`
	// +kubebuilder:validation:MinLength=1
	SSHKey string          `json:"sshKey,omitempty"`
	Env    []corev1.EnvVar `json:"env,omitempty"`
}

func (i *InfrastructureProvider) NameVersion() string {
	if i.Version != nil {
		return fmt.Sprintf("%s:%s", i.Name, *i.Version)
	}
	return i.Name
}

// +kubebuilder:validation:Enum=calico;provider
type CNI string

const (
	CalicoCNI   = CNI("calico")
	EmptyCNI    = CNI("")
	ProviderCNI = CNI("provider")
	Finalizer   = "undistro.io"
)

var cniMapAddr = map[CNI]string{
	CalicoCNI: "https://docs.projectcalico.org/v3.16/manifests/calico.yaml",
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// +kubebuilder:validation:MinLength=1
	KubernetesVersion      string                 `json:"kubernetesVersion,omitempty"`
	Template               *string                `json:"template,omitempty"`
	InfrastructureProvider InfrastructureProvider `json:"infrastructureProvider,omitempty"`
	ControlPlaneNode       ControlPlaneNode       `json:"controlPlaneNode,omitempty"`
	WorkerNodes            []WorkerNode           `json:"workerNodes,omitempty"`
	CniName                CNI                    `json:"cniName,omitempty"`
	Network                *Network               `json:"network,omitempty"`
	Bastion                *Bastion               `json:"bastion,omitempty"`
}

type Bastion struct {
	Enabled             bool     `json:"enabled,omitempty"`
	DisableIngressRules bool     `json:"disableIngressRules,omitempty"`
	AllowedCIDRBlocks   []string `json:"allowedCIDRBlocks,omitempty"`
	InstanceType        string   `json:"instanceType,omitempty"`
}

type Network struct {
	VPC          string   `json:"vpc,omitempty"`
	Subnets      []string `json:"subnets,omitempty"`
	PodsCIDR     []string `json:"podsCIDR,omitempty"`
	ServicesCIDR string   `json:"servicesCIDR,omitempty"`
}

type InstalledComponent struct {
	Name    string       `json:"name,omitempty"`
	Version string       `json:"version,omitempty"`
	URL     string       `json:"url,omitempty"`
	Type    ProviderType `json:"type,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	Phase               ClusterPhase            `json:"phase,omitempty"`
	InstalledComponents []InstalledComponent    `json:"installedComponents,omitempty"`
	Ready               bool                    `json:"ready,omitempty"`
	ClusterAPIRef       *corev1.ObjectReference `json:"clusterAPIRef,omitempty"`
	KubernetesVersion   string                  `json:"kubernetesVersion,omitempty"`
	ControlPlaneNode    ControlPlaneNode        `json:"controlPlaneNode,omitempty"`
	WorkerNodes         []WorkerNode            `json:"workerNodes,omitempty"`
	InfrastructureName  string                  `json:"infrastructureName,omitempty"`
	TotalWorkerReplicas int64                   `json:"totalWorkerReplicas,omitempty"`
	TotalWorkerPools    int64                   `json:"totalWorkerPools,omitempty"`
	BastionPublicIP     string                  `json:"bastionPublicIP,omitempty"`
	BastionConfig       *Bastion                `json:"bastionConfig,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clusters,shortName=cl,scope=Cluster,categories=undistro
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Infra",type="string",JSONPath=".status.infrastructureName"
// +kubebuilder:printcolumn:name="K8s",type="string",JSONPath=".status.kubernetesVersion"
// +kubebuilder:printcolumn:name="Worker Pools",type="integer",JSONPath=".status.totalWorkerPools"
// +kubebuilder:printcolumn:name="Control Plane Replicas",type="integer",JSONPath=".status.controlPlaneNode.replicas"
// +kubebuilder:printcolumn:name="Worker Replicas",type="integer",JSONPath=".status.totalWorkerReplicas"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Bastion IP",type="string",JSONPath=".status.bastionPublicIP"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready"

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

func (c *Cluster) GetManagedProvidersInfra() (bootstrap []string, controlplane []string) {
	switch c.Spec.InfrastructureProvider.Name {
	case "aws":
		return []string{"eks"}, []string{"eks"}
	}
	return nil, nil
}

func (c *Cluster) GetCNITemplateURL() string {
	return cniMapAddr[c.Spec.CniName]
}

func (c Cluster) GetBastion() Bastion {
	if c.Spec.Bastion == nil {
		return Bastion{
			Enabled:             true,
			DisableIngressRules: true,
		}
	}
	return *c.Spec.Bastion
}

func (c Cluster) IsManaged() bool {
	return c.Spec.InfrastructureProvider.Managed
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
