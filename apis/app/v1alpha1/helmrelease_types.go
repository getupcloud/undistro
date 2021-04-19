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
	"strings"

	"github.com/getupio-undistro/undistro/pkg/meta"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValuesReference contains a reference to a resource containing Helm values,
// and optionally the key they can be found at.
type ValuesReference struct {
	// Kind of the values referent, valid values are ('Secret', 'ConfigMap').
	// +kubebuilder:validation:Enum=Secret;ConfigMap
	// +required
	Kind string `json:"kind"`

	// Name of the values referent. Should reside in the same namespace as the
	// referring resource.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Name string `json:"name"`

	// ValuesKey is the data key where the values.yaml or a specific value can be
	// found at. Defaults to 'values.yaml'.
	// +optional
	ValuesKey string `json:"valuesKey,omitempty"`

	// TargetPath is the YAML dot notation path the value should be merged at. When
	// set, the ValuesKey is expected to be a single flat value. Defaults to 'None',
	// which results in the values getting merged at the root.
	// +optional
	TargetPath string `json:"targetPath,omitempty"`

	// Optional marks this ValuesReference as optional. When set, a not found error
	// for the values reference is ignored, but any ValuesKey, TargetPath or
	// transient error will still result in a reconciliation failure.
	// +optional
	Optional bool `json:"optional,omitempty"`
}

type ChartSource struct {
	RepoChartSource `json:",inline,omitempty"`
	SecretRef       *corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// RepoChartSources describes a Helm chart sourced from a Helm
// repository.
type RepoChartSource struct {
	// RepoURL is the URL of the Helm repository, e.g.
	// `https://kubernetes-charts.storage.googleapis.com` or
	// `https://charts.example.com`.
	RepoURL string `json:"repository,omitempty"`
	// Name is the name of the Helm chart _without_ an alias, e.g.
	// redis (for `helm upgrade [flags] stable/redis`).

	Name string `json:"name,omitempty"`
	// Version is the targeted Helm chart version, e.g. 7.0.1.

	Version string `json:"version,omitempty"`
}

// CleanRepoURL returns the RepoURL but ensures it ends with a trailing
// slash.
func (s RepoChartSource) CleanRepoURL() string {
	cleanURL := strings.TrimRight(s.RepoURL, "/")
	return cleanURL + "/"
}

type Rollback struct {
	// Force will mark this Helm release to `--force` rollbacks. This
	// forces the resource updates through delete/recreate if needed.
	Force bool `json:"force,omitempty"`
	// Recreate will mark this Helm release to `--recreate-pods` for
	// if applicable. This performs pod restarts.
	Recreate bool `json:"recreate,omitempty"`
	// DisableHooks will mark this Helm release to prevent hooks from
	// running during the rollback.
	DisableHooks bool `json:"disableHooks,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during rollback.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// Wait will mark this Helm release to wait until all Pods,
	// PVCs, Services, and minimum number of Pods of a Deployment,
	// StatefulSet, or ReplicaSet are in a ready state before marking
	// the release as successful.
	Wait bool `json:"wait,omitempty"`
}

type Test struct {
	// Enable will mark this Helm release for tests.
	Enable bool `json:"enable,omitempty"`
	// IgnoreFailures will cause a Helm release to be rolled back
	// if it fails otherwise it will be left in a released state
	IgnoreFailures bool `json:"ignoreFailures,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during test.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// Cleanup, when targeting Helm 2, determines whether to delete
	// test pods between each test run initiated by the Helm Operator.
	Cleanup *bool `json:"cleanup,omitempty"`
}

type HelmReleaseSpec struct {
	Chart       ChartSource `json:"chart,omitempty"`
	ReleaseName string      `json:"releaseName,omitempty"`
	ClusterName string      `json:"clusterName,omitempty"`
	MaxHistory  *int        `json:"maxHistory,omitempty"`
	// TargetNamespace overrides the targeted namespace for the Helm
	// release. The default namespace equals to the namespace of the
	// HelmRelease resource.
	TargetNamespace string `json:"targetNamespace,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during installation and
	// upgrade operations.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ResetValues will mark this Helm release to reset the values
	// to the defaults of the targeted chart before performing
	// an upgrade. Not explicitly setting this to `false` equals
	// to `true` due to the declarative nature of the operator.
	ResetValues *bool `json:"resetValues,omitempty"`
	// SkipCRDs will mark this Helm release to skip the creation
	// of CRDs during a Helm 3 installation.
	SkipCRDs bool `json:"skipCRDs,omitempty"`
	// Wait will mark this Helm release to wait until all Pods,
	// PVCs, Services, and minimum number of Pods of a Deployment,
	// StatefulSet, or ReplicaSet are in a ready state before marking
	// the release as successful.
	Wait *bool `json:"wait,omitempty"`
	// Force will mark this Helm release to `--force` upgrades. This
	// forces the resource updates through delete/recreate if needed.
	ForceUpgrade *bool `json:"forceUpgrade,omitempty"`
	// The rollback settings for this Helm release.
	Rollback Rollback `json:"rollback,omitempty"`
	// The test settings for this Helm release.
	Test Test `json:"test,omitempty"`
	// Values holds the values for this Helm release.
	Values *apiextensionsv1.JSON `json:"values,omitempty"`
	// ValuesFrom holds references to resources containing Helm values for this HelmRelease,
	// and information about how they should be merged.
	ValuesFrom []ValuesReference `json:"valuesFrom,omitempty"`
	// BeforeApplyObjects holds the objects that will be applied
	// before this helm release installation
	BeforeApplyObjects []apiextensionsv1.JSON `json:"beforeApplyObjects,omitempty"`
	// AfterApplyObjects holds the objects that will be applied
	// after this helm release installation
	AfterApplyObjects []apiextensionsv1.JSON `json:"afterApplyObjects,omitempty"`
	// Dependencies holds the referencies of objects
	// this HelmRelease depends on
	Dependencies []corev1.ObjectReference `json:"dependencies,omitempty"`
	Paused       bool                     `json:"paused,omitempty"`
	AutoUpgrade  bool                     `json:"autoUpgrade,omitempty"`
}

// HelmReleaseStatus defines the observed state of HelmRelease// HelmReleaseStatus defines the observed state of a HelmRelease.
type HelmReleaseStatus struct {
	// ObservedGeneration is the last observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// LastAppliedRevision is the revision of the last successfully applied source.
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// LastAttemptedValuesChecksum is the SHA1 checksum of the values of the last
	// reconciliation attempt.
	LastAttemptedValuesChecksum string `json:"lastAttemptedValuesChecksum,omitempty"`

	// LastReleaseRevision is the revision of the last successful Helm release.
	LastReleaseRevision int `json:"lastReleaseRevision,omitempty"`

	// Failures is the reconciliation failure count against the latest desired
	// state. It is reset after a successful reconciliation.
	Failures int64 `json:"failures,omitempty"`

	// InstallFailures is the install failure count against the latest desired
	// state. It is reset after a successful reconciliation.
	InstallFailures int64 `json:"installFailures,omitempty"`

	// UpgradeFailures is the upgrade failure count against the latest desired
	// state. It is reset after a successful reconciliation.
	UpgradeFailures int64 `json:"upgradeFailures,omitempty"`
}

// HelmReleaseProgressing resets any failures and registers progress toward
// reconciling the given HelmRelease by setting the meta.ReadyCondition to
// 'Unknown' for meta.ProgressingReason.
func HelmReleaseProgressing(hr HelmRelease) HelmRelease {
	hr.Status.Conditions = []metav1.Condition{}
	msg := "Reconciliation in progress"
	meta.SetResourceCondition(&hr, meta.ReadyCondition, metav1.ConditionUnknown, meta.ProgressingReason, msg)
	resetFailureCounts(&hr)
	return hr
}

// HelmReleaseNotReady registers a failed reconciliation of the given HelmRelease.
func HelmReleaseNotReady(hr HelmRelease, reason, message string) HelmRelease {
	meta.SetResourceCondition(&hr, meta.ReadyCondition, metav1.ConditionFalse, reason, message)
	hr.Status.Failures++
	return hr
}

// HelmReleaseReady registers a successful reconciliation of the given HelmRelease.
func HelmReleaseReady(hr HelmRelease) HelmRelease {
	msg := "Release reconciliation succeeded"
	meta.SetResourceCondition(&hr, meta.ReadyCondition, metav1.ConditionTrue, meta.ReconciliationSucceededReason, msg)
	hr.Status.LastAppliedRevision = hr.Status.LastAttemptedRevision
	resetFailureCounts(&hr)
	return hr
}

// HelmReleaseAttempted registers an attempt of the given HelmRelease with the given state.
// and returns the modified HelmRelease and a boolean indicating a state change.
func HelmReleaseAttempted(hr HelmRelease, revision string, releaseRevision int, valuesChecksum string) (HelmRelease, bool) {
	changed := hr.Status.LastAttemptedRevision != revision ||
		hr.Status.LastReleaseRevision != releaseRevision ||
		hr.Status.LastAttemptedValuesChecksum != valuesChecksum
	hr.Status.LastAttemptedRevision = revision
	hr.Status.LastReleaseRevision = releaseRevision
	hr.Status.LastAttemptedValuesChecksum = valuesChecksum

	return hr, changed
}

func resetFailureCounts(hr *HelmRelease) {
	hr.Status.Failures = 0
	hr.Status.InstallFailures = 0
	hr.Status.UpgradeFailures = 0
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=hr,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterName",description=""
// +kubebuilder:printcolumn:name="Chart",type="string",JSONPath=".spec.chart.name",description=""
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.chart.version",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""

// HelmRelease is the Schema for the helmreleases API
type HelmRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmReleaseSpec   `json:"spec,omitempty"`
	Status HelmReleaseStatus `json:"status,omitempty"`
}

// GetStatusConditions returns a pointer to the Status.Conditions slice
func (hr *HelmRelease) GetStatusConditions() *[]metav1.Condition {
	return &hr.Status.Conditions
}

func (hr *HelmRelease) GetNamespace() string {
	if hr.Namespace == "" {
		return "default"
	}
	return hr.Namespace
}

// +kubebuilder:object:root=true

// HelmReleaseList contains a list of HelmRelease
type HelmReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmRelease{}, &HelmReleaseList{})
}
