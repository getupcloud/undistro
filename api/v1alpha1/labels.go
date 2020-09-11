/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package v1alpha1

import "fmt"

const (
	HelmApplyAfter  = "getupcloud.com/apply-after-deploy"
	HelmApplyBefore = "getupcloud.com/apply-before-deploy"
)

const (
	// ClusterctlLabelName is applied to all components managed by undistro.
	ClusterctlLabelName = "clusterctl.cluster.x-k8s.io"

	// ClusterctlCoreLabelName is applied to all the core objects managed by undistro.
	ClusterctlCoreLabelName = "clusterctl.cluster.x-k8s.io/core"

	// ClusterctlResourceLifecyleLabelName describes the lifecyle for a specific resource.
	//
	// Example: resources shared between instances of the same provider:  CRDs,
	// ValidatingWebhookConfiguration, MutatingWebhookConfiguration, and so on.
	ClusterctlResourceLifecyleLabelName = "clusterctl.cluster.x-k8s.io/lifecycle"

	// ClusterctlMoveLabelName can be set on CRDs that providers wish to move that are not part of a cluster
	ClusterctlMoveLabelName = "clusterctl.cluster.x-k8s.io/move"

	// UndisroLabelName is applied to all components managed by undistro.
	UndistroLabelName = "getupcloud.com"

	// UndistroCoreLabelName is applied to all the core objects managed by undistro.
	UndistroCoreLabelName = "getupcloud.com/core"

	// UndistroResourceLifecyleLabelName describes the lifecyle for a specific resource.
	//
	// Example: resources shared between instances of the same provider:  CRDs,
	// ValidatingWebhookConfiguration, MutatingWebhookConfiguration, and so on.
	UndistroResourceLifecyleLabelName = "getupcloud.com/lifecycle"

	// UndistroMoveLabelName can be set on CRDs that providers wish to move that are not part of a cluster
	UndistroMoveLabelName = "getupcloud.com/move"
)

// ResourceLifecycle configures the lifecycle of a resource
type ResourceLifecycle string

const (
	// ResourceLifecycleShared is used to indicate that a resource is shared between
	// multiple instances of a provider.
	ResourceLifecycleShared = ResourceLifecycle("shared")
)

// ManifestLabel returns the cluster.x-k8s.io/provider label value for a provider/type.
//
// Note: the label uniquely describes the provider type and its kind (e.g. bootstrap-kubeadm);
// it's not meant to be used to describe each instance of a particular provider.
func ManifestLabel(name string, providerType ProviderType) string {
	switch providerType {
	case CoreProviderType, UndistroProviderType:
		return name
	case BootstrapProviderType:
		return fmt.Sprintf("bootstrap-%s", name)
	case ControlPlaneProviderType:
		return fmt.Sprintf("control-plane-%s", name)
	case InfrastructureProviderType:
		return fmt.Sprintf("infrastructure-%s", name)
	default:
		return fmt.Sprintf("unknown-type-%s", name)
	}
}
