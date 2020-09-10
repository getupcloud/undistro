/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/getupcloud/undistro/log"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type HelmAction string

const (
	InstallAction       HelmAction = "install"
	UpgradeAction       HelmAction = "upgrade"
	SkipAction          HelmAction = "skip"
	RollbackAction      HelmAction = "rollback"
	UninstallAction     HelmAction = "uninstall"
	DryRunCompareAction HelmAction = "dry-run-compare"
	TestAction          HelmAction = "test"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmRelease is a type to represent a Helm release.
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Release",type="string",JSONPath=".status.releaseName",description="Release is the name of the Helm release, as given by Helm."
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase is the current release phase being performed for the HelmRelease."
// +kubebuilder:printcolumn:name="ReleaseStatus",type="string",JSONPath=".status.releaseStatus",description="ReleaseStatus is the status of the Helm release, as given by Helm."
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=helmreleases,shortName=hr;hrs
type HelmRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   HelmReleaseSpec   `json:"spec"`
	Status HelmReleaseStatus `json:"status,omitempty"`
}

func (hr HelmRelease) GetClusterNamespacedName() types.NamespacedName {
	nm := types.NamespacedName{}
	if hr.Spec.ClusterName != "" {
		spl := strings.Split(hr.Spec.ClusterName, "/")
		if len(spl) == 2 {
			if spl[0] != "default" {
				nm.Namespace = spl[0]
			}
			nm.Name = spl[1]
			return nm
		}
		nm.Name = spl[0]
	}
	return nm
}

// GetReleaseName returns the configured release name, or constructs and
// returns one based on the namespace and name of the HelmRelease.
// When the HelmRelease's metadata.namespace and spec.targetNamespace
// differ, both are used in the generated name.
// This name is used for naming and operating on the release in Helm.
func (hr HelmRelease) GetReleaseName() string {
	if hr.Spec.ReleaseName == "" {
		namespace := hr.GetDefaultedNamespace()
		targetNamespace := hr.GetTargetNamespace()

		if namespace != targetNamespace {
			// prefix the releaseName with the administering HelmRelease namespace as well
			return fmt.Sprintf("%s-%s-%s", namespace, targetNamespace, hr.Name)
		}
		return fmt.Sprintf("%s-%s", targetNamespace, hr.Name)
	}

	return hr.Spec.ReleaseName
}

// GetDefaultedNamespace returns the HelmRelease's namespace
// defaulting to the "default" if not set.
func (hr HelmRelease) GetDefaultedNamespace() string {
	if hr.GetNamespace() == "" {
		return "default"
	}
	return hr.Namespace
}

// GetTargetNamespace returns the configured release targetNamespace
// defaulting to the namespace of the HelmRelease if not set.
func (hr HelmRelease) GetTargetNamespace() string {
	if hr.Spec.TargetNamespace == "" {
		return hr.GetDefaultedNamespace()
	}
	return hr.Spec.TargetNamespace
}

func (hr HelmRelease) GetHelmVersion(defaultVersion string) string {
	if defaultVersion != "" {
		return defaultVersion
	}
	return string(HelmV3)
}

// GetTimeout returns the install or upgrade timeout (defaults to 300s)
func (hr HelmRelease) GetTimeout() time.Duration {
	if hr.Spec.Timeout == nil {
		return 300 * time.Second
	}
	return time.Duration(*hr.Spec.Timeout) * time.Second
}

// GetMaxHistory returns the maximum number of release
// revisions to keep (defaults to 10)
func (hr HelmRelease) GetMaxHistory() int {
	if hr.Spec.MaxHistory == nil {
		return 10
	}
	return *hr.Spec.MaxHistory
}

// GetReuseValues returns if the values of the previous release should
// be reused based on the value of `ResetValues`. When this value is
// not explicitly set, it is assumed values should not be reused, as
// this aligns with the declarative behaviour of the operator.
func (hr HelmRelease) GetReuseValues() bool {
	switch hr.Spec.ResetValues {
	case nil:
		return false
	default:
		return !*hr.Spec.ResetValues
	}
}

// GetWait returns if wait should be enabled. If not explicitly set
// it is true if either rollbacks or tests are enabled, and false
// otherwise.
func (hr HelmRelease) GetWait() bool {
	switch hr.Spec.Wait {
	case nil:
		return hr.Spec.Rollback.Enable || hr.Spec.Test.Enable
	default:
		return *hr.Spec.Wait
	}
}

// GetValuesFromSources maintains backwards compatibility with
// ValueFileSecrets by merging them into the ValuesFrom array.
func (hr HelmRelease) GetValuesFromSources() []ValuesFromSource {
	return hr.Spec.ValuesFrom
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmReleaseList is a list of HelmReleases
type HelmReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []HelmRelease `json:"items"`
}

type ChartSource struct {
	// +optional
	*RepoChartSource `json:",inline"`
}

// RepoChartSources describes a Helm chart sourced from a Helm
// repository.
type RepoChartSource struct {
	// RepoURL is the URL of the Helm repository, e.g.
	// `https://kubernetes-charts.storage.googleapis.com` or
	// `https://charts.example.com`.
	// +kubebuilder:validation:Optional
	RepoURL string `json:"repository"`
	// Name is the name of the Helm chart _without_ an alias, e.g.
	// redis (for `helm upgrade [flags] stable/redis`).
	// +kubebuilder:validation:Optional
	Name string `json:"name"`
	// Version is the targeted Helm chart version, e.g. 7.0.1.
	// +kubebuilder:validation:Optional
	Version string `json:"version"`
}

// CleanRepoURL returns the RepoURL but ensures it ends with a trailing
// slash.
func (s RepoChartSource) CleanRepoURL() string {
	cleanURL := strings.TrimRight(s.RepoURL, "/")
	return cleanURL + "/"
}

type ValuesFromSource struct {
	// The reference to a config map with release values.
	// +optional
	ConfigMapKeyRef *OptionalConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// The reference to a secret with release values.
	// +optional
	SecretKeyRef *OptionalSecretKeySelector `json:"secretKeyRef,omitempty"`
	// The reference to an external source with release values.
	// +optional
	ExternalSourceRef *ExternalSourceSelector `json:"externalSourceRef,omitempty"`
	// The reference to a local chart file with release values.
	// +optional
	ChartFileRef *ChartFileSelector `json:"chartFileRef,omitempty"`
}

type ChartFileSelector struct {
	// Path is the file path to the source relative to the chart root.
	Path string `json:"path"`
	// Optional will mark this ChartFileSelector as optional.
	// The result of this are that operations are permitted without
	// the source, due to it e.g. being temporarily unavailable.
	// +optional
	Optional *bool `json:"optional,omitempty"`
}

type ExternalSourceSelector struct {
	// URL is the URL of the external source.
	URL string `json:"url"`
	// Optional will mark this ExternalSourceSelector as optional.
	// The result of this are that operations are permitted without
	// the source, due to it e.g. being temporarily unavailable.
	// +optional
	Optional *bool `json:"optional,omitempty"`
}

type Rollback struct {
	// Enable will mark this Helm release for rollbacks.
	// +optional
	Enable bool `json:"enable,omitempty"`
	// Retry will mark this Helm release for upgrade retries after a
	// rollback.
	// +optional
	Retry bool `json:"retry,omitempty"`
	// MaxRetries is the maximum amount of upgrade retries the operator
	// should make before bailing.
	// +optional
	MaxRetries *int64 `json:"maxRetries,omitempty"`
	// Force will mark this Helm release to `--force` rollbacks. This
	// forces the resource updates through delete/recreate if needed.
	// +optional
	Force bool `json:"force,omitempty"`
	// Recreate will mark this Helm release to `--recreate-pods` for
	// if applicable. This performs pod restarts.
	// +optional
	Recreate bool `json:"recreate,omitempty"`
	// DisableHooks will mark this Helm release to prevent hooks from
	// running during the rollback.
	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during rollback.
	// +optional
	Timeout *int64 `json:"timeout,omitempty"`
	// Wait will mark this Helm release to wait until all Pods,
	// PVCs, Services, and minimum number of Pods of a Deployment,
	// StatefulSet, or ReplicaSet are in a ready state before marking
	// the release as successful.
	// +optional
	Wait bool `json:"wait,omitempty"`
}

// GetTimeout returns the configured timout for the Helm release,
// or the default of 300s.
func (r Rollback) GetTimeout() time.Duration {
	if r.Timeout == nil {
		return 300 * time.Second
	}
	return time.Duration(*r.Timeout) * time.Second
}

// GetMaxRetries returns the configured max retries for the Helm
// release, or the default of 5.
func (r Rollback) GetMaxRetries() int64 {
	if r.MaxRetries == nil {
		return 5
	}
	return *r.MaxRetries
}

type Test struct {
	// Enable will mark this Helm release for tests.
	// +optional
	Enable bool `json:"enable,omitempty"`
	// IgnoreFailures will cause a Helm release to be rolled back
	// if it fails otherwise it will be left in a released state
	// +optional
	IgnoreFailures *bool `json:"ignoreFailures,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during test.
	// +optional
	Timeout *int64 `json:"timeout,omitempty"`
	// Cleanup, when targeting Helm 2, determines whether to delete
	// test pods between each test run initiated by the Helm Operator.
	// +optional
	Cleanup *bool `json:"cleanup,omitempty"`
}

// IgnoreFailures returns the configured ignoreFailures flag,
// or the default of false to preserve backwards compatible
func (t Test) GetIgnoreFailures() bool {
	switch t.IgnoreFailures {
	case nil:
		return false
	default:
		return *t.IgnoreFailures
	}
}

// GetTimeout returns the configured timout for the Helm release,
// or the default of 300s.
func (t Test) GetTimeout() time.Duration {
	if t.Timeout == nil {
		return 300 * time.Second
	}
	return time.Duration(*t.Timeout) * time.Second
}

// GetCleanup returns the configured test cleanup flag, or the
// default of true.
func (t Test) GetCleanup() bool {
	switch t.Cleanup {
	case nil:
		return true
	default:
		return *t.Cleanup
	}
}

// HelmVersion is the version of Helm to target. If not supplied,
// the lowest _enabled Helm version_ will be targeted.
// Valid HelmVersion values are:
// "v3"
// +kubebuilder:validation:Enum="v3"
// +optional
type HelmVersion string

const (
	HelmV3 HelmVersion = "v3"
)

func (hr HelmRelease) GetValues() map[string]interface{} {
	var values map[string]interface{}
	if hr.Spec.Values != nil {
		err := json.Unmarshal(hr.Spec.Values.Raw, &values)
		if err != nil {
			log.Log.Error(err, "is not a json")
		}
	}
	return values
}

type HelmReleaseSpec struct {
	// +kubebuilder:validation:Required
	ChartSource `json:"chart,omitempty"`
	// ReleaseName is the name of the The Helm release. If not supplied,
	// it will be generated by affixing the namespace to the resource
	// name.
	ReleaseName string `json:"releaseName,omitempty"`
	// +kubebuilder:validation:Required
	ClusterName string `json:"clusterName,omitempty"`
	// MaxHistory is the maximum amount of revisions to keep for the
	// Helm release. If not supplied, it defaults to 10.
	MaxHistory *int               `json:"maxHistory,omitempty"`
	ValuesFrom []ValuesFromSource `json:"valuesFrom,omitempty"`
	// TargetNamespace overrides the targeted namespace for the Helm
	// release. The default namespace equals to the namespace of the
	// HelmRelease resource.
	// +optional
	TargetNamespace string `json:"targetNamespace,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during installation and
	// upgrade operations.
	// +optional
	Timeout *int64 `json:"timeout,omitempty"`
	// ResetValues will mark this Helm release to reset the values
	// to the defaults of the targeted chart before performing
	// an upgrade. Not explicitly setting this to `false` equals
	// to `true` due to the declarative nature of the operator.
	// +optional
	ResetValues *bool `json:"resetValues,omitempty"`
	// SkipCRDs will mark this Helm release to skip the creation
	// of CRDs during a Helm 3 installation.
	// +optional
	SkipCRDs bool `json:"skipCRDs,omitempty"`
	// Wait will mark this Helm release to wait until all Pods,
	// PVCs, Services, and minimum number of Pods of a Deployment,
	// StatefulSet, or ReplicaSet are in a ready state before marking
	// the release as successful.
	// +optional
	Wait *bool `json:"wait,omitempty"`
	// Force will mark this Helm release to `--force` upgrades. This
	// forces the resource updates through delete/recreate if needed.
	// +optional
	ForceUpgrade bool `json:"forceUpgrade,omitempty"`
	// The rollback settings for this Helm release.
	// +optional
	Rollback Rollback `json:"rollback,omitempty"`
	// The test settings for this Helm release.
	// +optional
	Test Test `json:"test,omitempty"`
	// Values holds the values for this Helm release.
	// +optional
	Values *apiextensionsv1.JSON `json:"values,omitempty"`
}

// HelmReleaseStatus contains status information about an HelmRelease.
type HelmReleaseStatus struct {
	// Phase the release is in, one of ('ChartFetched',
	// 'ChartFetchFailed', 'Installing', 'Upgrading', 'Deployed',
	// 'DeployFailed', 'Testing', 'TestFailed', 'Tested', 'Succeeded',
	// 'RollingBack', 'RolledBack', 'RollbackFailed')
	// +optional
	Phase HelmReleasePhase `json:"phase,omitempty"`

	// ReleaseName is the name as either supplied or generated.
	// +optional
	ReleaseName string `json:"releaseName,omitempty"`

	// ReleaseStatus is the status as given by Helm for the release
	// managed by this resource.
	// +optional
	ReleaseStatus string `json:"releaseStatus,omitempty"`

	// Revision holds the Git hash or version of the chart currently
	// deployed.
	// +optional
	Revision string `json:"revision,omitempty"`

	// LastAttemptedRevision is the revision of the latest chart
	// sync, and may be of a failed release.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// RollbackCount records the amount of rollback attempts made,
	// it is incremented after a rollback failure and reset after a
	// successful upgrade or revision change.
	// +optional
	RollbackCount int64 `json:"rollbackCount,omitempty"`
}

func init() {
	SchemeBuilder.Register(&HelmRelease{}, &HelmReleaseList{})
}
