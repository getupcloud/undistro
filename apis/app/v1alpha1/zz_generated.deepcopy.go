// +build !ignore_autogenerated

/*
Copyright 2021 The UnDistro authors

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Autoscaling) DeepCopyInto(out *Autoscaling) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Autoscaling.
func (in *Autoscaling) DeepCopy() *Autoscaling {
	if in == nil {
		return nil
	}
	out := new(Autoscaling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Bastion) DeepCopyInto(out *Bastion) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	if in.AllowedCIDRBlocks != nil {
		in, out := &in.AllowedCIDRBlocks, &out.AllowedCIDRBlocks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Bastion.
func (in *Bastion) DeepCopy() *Bastion {
	if in == nil {
		return nil
	}
	out := new(Bastion)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartSource) DeepCopyInto(out *ChartSource) {
	*out = *in
	out.RepoChartSource = in.RepoChartSource
	if in.SecretRef != nil {
		in, out := &in.SecretRef, &out.SecretRef
		*out = new(v1.LocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartSource.
func (in *ChartSource) DeepCopy() *ChartSource {
	if in == nil {
		return nil
	}
	out := new(ChartSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cluster) DeepCopyInto(out *Cluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cluster.
func (in *Cluster) DeepCopy() *Cluster {
	if in == nil {
		return nil
	}
	out := new(Cluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Cluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterList) DeepCopyInto(out *ClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Cluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterList.
func (in *ClusterList) DeepCopy() *ClusterList {
	if in == nil {
		return nil
	}
	out := new(ClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterSpec) DeepCopyInto(out *ClusterSpec) {
	*out = *in
	in.Network.DeepCopyInto(&out.Network)
	in.InfrastructureProvider.DeepCopyInto(&out.InfrastructureProvider)
	if in.Bastion != nil {
		in, out := &in.Bastion, &out.Bastion
		*out = new(Bastion)
		(*in).DeepCopyInto(*out)
	}
	if in.ControlPlane != nil {
		in, out := &in.ControlPlane, &out.ControlPlane
		*out = new(ControlPlaneNode)
		(*in).DeepCopyInto(*out)
	}
	if in.Workers != nil {
		in, out := &in.Workers, &out.Workers
		*out = make([]WorkerNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterSpec.
func (in *ClusterSpec) DeepCopy() *ClusterSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterStatus) DeepCopyInto(out *ClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.BastionConfig != nil {
		in, out := &in.BastionConfig, &out.BastionConfig
		*out = new(Bastion)
		(*in).DeepCopyInto(*out)
	}
	in.ControlPlane.DeepCopyInto(&out.ControlPlane)
	if in.Workers != nil {
		in, out := &in.Workers, &out.Workers
		*out = make([]WorkerNode, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterStatus.
func (in *ClusterStatus) DeepCopy() *ClusterStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneNode) DeepCopyInto(out *ControlPlaneNode) {
	*out = *in
	in.Node.DeepCopyInto(&out.Node)
	out.Endpoint = in.Endpoint
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneNode.
func (in *ControlPlaneNode) DeepCopy() *ControlPlaneNode {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneNode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DefaultPolicies) DeepCopyInto(out *DefaultPolicies) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefaultPolicies.
func (in *DefaultPolicies) DeepCopy() *DefaultPolicies {
	if in == nil {
		return nil
	}
	out := new(DefaultPolicies)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DefaultPolicies) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DefaultPoliciesList) DeepCopyInto(out *DefaultPoliciesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DefaultPolicies, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefaultPoliciesList.
func (in *DefaultPoliciesList) DeepCopy() *DefaultPoliciesList {
	if in == nil {
		return nil
	}
	out := new(DefaultPoliciesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DefaultPoliciesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DefaultPoliciesSpec) DeepCopyInto(out *DefaultPoliciesSpec) {
	*out = *in
	if in.ExcludePolicies != nil {
		in, out := &in.ExcludePolicies, &out.ExcludePolicies
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefaultPoliciesSpec.
func (in *DefaultPoliciesSpec) DeepCopy() *DefaultPoliciesSpec {
	if in == nil {
		return nil
	}
	out := new(DefaultPoliciesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DefaultPoliciesStatus) DeepCopyInto(out *DefaultPoliciesStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AppliedPolicies != nil {
		in, out := &in.AppliedPolicies, &out.AppliedPolicies
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefaultPoliciesStatus.
func (in *DefaultPoliciesStatus) DeepCopy() *DefaultPoliciesStatus {
	if in == nil {
		return nil
	}
	out := new(DefaultPoliciesStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmRelease) DeepCopyInto(out *HelmRelease) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmRelease.
func (in *HelmRelease) DeepCopy() *HelmRelease {
	if in == nil {
		return nil
	}
	out := new(HelmRelease)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HelmRelease) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmReleaseList) DeepCopyInto(out *HelmReleaseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HelmRelease, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmReleaseList.
func (in *HelmReleaseList) DeepCopy() *HelmReleaseList {
	if in == nil {
		return nil
	}
	out := new(HelmReleaseList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HelmReleaseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmReleaseSpec) DeepCopyInto(out *HelmReleaseSpec) {
	*out = *in
	in.Chart.DeepCopyInto(&out.Chart)
	if in.MaxHistory != nil {
		in, out := &in.MaxHistory, &out.MaxHistory
		*out = new(int)
		**out = **in
	}
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.ResetValues != nil {
		in, out := &in.ResetValues, &out.ResetValues
		*out = new(bool)
		**out = **in
	}
	if in.Wait != nil {
		in, out := &in.Wait, &out.Wait
		*out = new(bool)
		**out = **in
	}
	if in.ForceUpgrade != nil {
		in, out := &in.ForceUpgrade, &out.ForceUpgrade
		*out = new(bool)
		**out = **in
	}
	in.Rollback.DeepCopyInto(&out.Rollback)
	in.Test.DeepCopyInto(&out.Test)
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = new(apiextensionsv1.JSON)
		(*in).DeepCopyInto(*out)
	}
	if in.ValuesFrom != nil {
		in, out := &in.ValuesFrom, &out.ValuesFrom
		*out = make([]ValuesReference, len(*in))
		copy(*out, *in)
	}
	if in.BeforeApplyObjects != nil {
		in, out := &in.BeforeApplyObjects, &out.BeforeApplyObjects
		*out = make([]apiextensionsv1.JSON, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AfterApplyObjects != nil {
		in, out := &in.AfterApplyObjects, &out.AfterApplyObjects
		*out = make([]apiextensionsv1.JSON, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Dependencies != nil {
		in, out := &in.Dependencies, &out.Dependencies
		*out = make([]v1.ObjectReference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmReleaseSpec.
func (in *HelmReleaseSpec) DeepCopy() *HelmReleaseSpec {
	if in == nil {
		return nil
	}
	out := new(HelmReleaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HelmReleaseStatus) DeepCopyInto(out *HelmReleaseStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HelmReleaseStatus.
func (in *HelmReleaseStatus) DeepCopy() *HelmReleaseStatus {
	if in == nil {
		return nil
	}
	out := new(HelmReleaseStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InfrastructureProvider) DeepCopyInto(out *InfrastructureProvider) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InfrastructureProvider.
func (in *InfrastructureProvider) DeepCopy() *InfrastructureProvider {
	if in == nil {
		return nil
	}
	out := new(InfrastructureProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LaunchTemplateReference) DeepCopyInto(out *LaunchTemplateReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LaunchTemplateReference.
func (in *LaunchTemplateReference) DeepCopy() *LaunchTemplateReference {
	if in == nil {
		return nil
	}
	out := new(LaunchTemplateReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Network) DeepCopyInto(out *Network) {
	*out = *in
	in.ClusterNetwork.DeepCopyInto(&out.ClusterNetwork)
	out.VPC = in.VPC
	if in.Subnets != nil {
		in, out := &in.Subnets, &out.Subnets
		*out = make([]NetworkSpec, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Network.
func (in *Network) DeepCopy() *Network {
	if in == nil {
		return nil
	}
	out := new(Network)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkSpec) DeepCopyInto(out *NetworkSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkSpec.
func (in *NetworkSpec) DeepCopy() *NetworkSpec {
	if in == nil {
		return nil
	}
	out := new(NetworkSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Node) DeepCopyInto(out *Node) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Taints != nil {
		in, out := &in.Taints, &out.Taints
		*out = make([]v1.Taint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ProviderTags != nil {
		in, out := &in.ProviderTags, &out.ProviderTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Node.
func (in *Node) DeepCopy() *Node {
	if in == nil {
		return nil
	}
	out := new(Node)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepoChartSource) DeepCopyInto(out *RepoChartSource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepoChartSource.
func (in *RepoChartSource) DeepCopy() *RepoChartSource {
	if in == nil {
		return nil
	}
	out := new(RepoChartSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rollback) DeepCopyInto(out *Rollback) {
	*out = *in
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(metav1.Duration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rollback.
func (in *Rollback) DeepCopy() *Rollback {
	if in == nil {
		return nil
	}
	out := new(Rollback)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Test) DeepCopyInto(out *Test) {
	*out = *in
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(metav1.Duration)
		**out = **in
	}
	if in.Cleanup != nil {
		in, out := &in.Cleanup, &out.Cleanup
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Test.
func (in *Test) DeepCopy() *Test {
	if in == nil {
		return nil
	}
	out := new(Test)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ValuesReference) DeepCopyInto(out *ValuesReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ValuesReference.
func (in *ValuesReference) DeepCopy() *ValuesReference {
	if in == nil {
		return nil
	}
	out := new(ValuesReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkerNode) DeepCopyInto(out *WorkerNode) {
	*out = *in
	in.Node.DeepCopyInto(&out.Node)
	out.Autoscale = in.Autoscale
	out.LaunchTemplateReference = in.LaunchTemplateReference
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkerNode.
func (in *WorkerNode) DeepCopy() *WorkerNode {
	if in == nil {
		return nil
	}
	out := new(WorkerNode)
	in.DeepCopyInto(out)
	return out
}
