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
	"strconv"
	"strings"

	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

type Node struct {
	Replicas     *int32            `json:"replicas,omitempty"`
	MachineType  string            `json:"machineType,omitempty"`
	Subnet       string            `json:"subnet,omitempty"`
	Taints       []corev1.Taint    `json:"taints,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	ProviderTags map[string]string `json:"providerTags,omitempty"`
}

func (n Node) TaintTmpl() string {
	has := len(n.Taints) > 0
	if !has {
		return ""
	}
	tstr := make([]string, len(n.Taints))
	for i, t := range n.Taints {
		tstr[i] = fmt.Sprintf("%s=%s:%v", t.Key, t.Value, t.Effect)
	}
	return strings.Join(tstr, ",")
}

func (n Node) LabelsTmpl() string {
	has := len(n.Labels) > 0
	if !has {
		return ""
	}
	lstr := make([]string, len(n.Labels))
	i := 0
	for k, v := range n.Labels {
		lstr[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return strings.Join(lstr, ",")
}

func (n Node) HasKubeletArgs() bool {
	return n.LabelsTmpl() != "" || n.TaintTmpl() != ""
}

type ControlPlaneNode struct {
	Node       `json:",inline,omitempty"`
	Endpoint   capi.APIEndpoint `json:"endpoint,omitempty"`
	InternalLB bool             `json:"internalLB,omitempty"`
}

type WorkerNode struct {
	Node      `json:",inline,omitempty"`
	Autoscale Autoscaling `json:"autoscaling,omitempty"`
	InfraNode bool        `json:"infraNode,omitempty"`
}

type Autoscaling struct {
	Enabled bool `json:"enabled,omitempty"`
	// The minimum size of the group.
	MinSize int32 `json:"minSize,omitempty"`
	// The maximum size of the group.
	MaxSize int32 `json:"maxSize,omitempty"`
}

type InfrastructureProvider struct {
	Name   string          `json:"name,omitempty"`
	SSHKey string          `json:"sshKey,omitempty"`
	Flavor string          `json:"flavor,omitempty"`
	Region string          `json:"region,omitempty"`
	Env    []corev1.EnvVar `json:"env,omitempty"`
}

func (i InfrastructureProvider) Flavors() []string {
	switch i.Name {
	case "aws":
		return []string{"ec2", "eks"}
	}
	return nil
}

func (i InfrastructureProvider) IsManaged() bool {
	return i.Name == "aws" && i.Flavor == "eks"
}

type NetworkSpec struct {
	ID        string `json:"id,omitempty"`
	CIDRBlock string `json:"cidrBlock,omitempty"`
	Zone      string `json:"zone,omitempty"`
	IsPublic  bool   `json:"isPublic,omitempty"`
}

type Network struct {
	capi.ClusterNetwork `json:",inline"`
	VPC                 NetworkSpec   `json:"vpc,omitempty"`
	Subnets             []NetworkSpec `json:"subnets,omitempty"`
	MultiZone           bool          `json:"multiZone,omitempty"`
}

type Bastion struct {
	Enabled             *bool    `json:"enabled,omitempty"`
	DisableIngressRules bool     `json:"disableIngressRules,omitempty"`
	AllowedCIDRBlocks   []string `json:"allowedCIDRBlocks,omitempty"`
	InstanceType        string   `json:"instanceType,omitempty"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	Paused                 bool                   `json:"paused,omitempty"`
	Network                Network                `json:"network,omitempty"`
	InfrastructureProvider InfrastructureProvider `json:"infrastructureProvider,omitempty"`
	KubernetesVersion      string                 `json:"kubernetesVersion,omitempty"`
	Bastion                *Bastion               `json:"bastion,omitempty"`
	ControlPlane           *ControlPlaneNode      `json:"controlPlane,omitempty"`
	Workers                []WorkerNode           `json:"workers,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// ObservedGeneration is the last observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions          []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	TotalWorkerReplicas int32              `json:"totalWorkerReplicas,omitempty"`
	TotalWorkerPools    int32              `json:"totalWorkerPools,omitempty"`
	BastionPublicIP     string             `json:"bastionPublicIP,omitempty"`
	LastUsedUID         string             `json:"lastUsedUID,omitempty"`
	BastionConfig       *Bastion           `json:"bastionConfig,omitempty"`
	KubernetesVersion   string             `json:"kubernetesVersion,omitempty"`
	ControlPlane        ControlPlaneNode   `json:"controlPlane,omitempty"`
	Workers             []WorkerNode       `json:"workers,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=cl,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="k8s",type="string",JSONPath=".spec.kubernetesVersion",description=""
// +kubebuilder:printcolumn:name="Infra",type="string",JSONPath=".spec.infrastructureProvider.name",description=""
// +kubebuilder:printcolumn:name="Worker Pools",type="integer",JSONPath=".status.totalWorkerPools",description=""
// +kubebuilder:printcolumn:name="Worker Replicas",type="integer",JSONPath=".status.totalWorkerReplicas",description=""
// +kubebuilder:printcolumn:name="ControlPlane Replicas",type="integer",JSONPath=".spec.controlPlane.replicas",description=""
// +kubebuilder:printcolumn:name="Bastion IP",type="string",JSONPath=".status.bastionPublicIP",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

func (c Cluster) GetTemplate() string {
	return fmt.Sprintf("%s/%s", c.Spec.InfrastructureProvider.Name, c.Spec.InfrastructureProvider.Flavor)
}

func (c *Cluster) GetNamespace() string {
	if c.Namespace == "" {
		return "default"
	}
	return c.Namespace
}

func (c *Cluster) GetStatusConditions() *[]metav1.Condition {
	return &c.Status.Conditions
}

func (c *Cluster) GetWorkerRefByMachinePool(mpName string) (WorkerNode, error) {
	split := strings.Split(mpName, "-")
	indexStr := split[len(split)-1]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return WorkerNode{}, err
	}
	if index > len(c.Spec.Workers)-1 {
		return WorkerNode{}, InvalidMP
	}
	return c.Spec.Workers[index], nil
}

var InvalidMP = errors.New("invalid machinepool")

// ClusterProgressing resets any failures and registers progress toward
// reconciling the given Cluster by setting the meta.ReadyCondition to
// 'Unknown' for meta.ProgressingReason.
func ClusterProgressing(p Cluster) Cluster {
	p.Status.Conditions = []metav1.Condition{}
	msg := "Reconciliation in progress"
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionUnknown, meta.ProgressingReason, msg)
	return p
}

// ClusterNotReady registers a failed reconciliation of the given Cluster.
func ClusterNotReady(p Cluster, reason, message string) Cluster {
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionFalse, reason, message)
	return p
}

// ClusterReady registers a successful reconciliation of the given Cluster.
func ClusterReady(p Cluster) Cluster {
	msg := "Cluster reconciliation succeeded"
	meta.SetResourceCondition(&p, meta.ReadyCondition, metav1.ConditionTrue, meta.ReconciliationSucceededReason, msg)
	return p
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
