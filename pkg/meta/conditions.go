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
package meta

import (
	"github.com/getupio-undistro/undistro/pkg/record"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// ReadyCondition is the name of the Ready condition implemented by all toolkit
	// resources.
	ReadyCondition string = "Ready"

	// ReconciliationSucceededReason represents the fact that the reconciliation of
	// a toolkit resource has succeeded.
	ReconciliationSucceededReason string = "ReconciliationSucceeded"

	// ReconciliationFailedReason represents the fact that the reconciliation of a
	// toolkit resource has failed.
	ReconciliationFailedReason string = "ReconciliationFailed"

	// ProgressingReason represents the fact that the reconciliation of a toolkit
	// resource is underway.
	ProgressingReason string = "Progressing"

	// DependencyNotReadyReason represents the fact that one of the toolkit resource
	// dependencies is not ready.
	DependencyNotReadyReason string = "DependencyNotReady"

	// PausedReason represents the fact that the reconciliation of a toolkit
	// resource is Paused.
	PausedReason string = "Paused"

	// URLInvalidReason represents the fact that a given source has an invalid URL.
	URLInvalidReason string = "URLInvalid"

	// StorageOperationFailedReason signals a failure caused by a storage operation.
	StorageOperationFailedReason string = "StorageOperationFailed"

	// AuthenticationFailedReason represents the fact that a given secret does not
	// have the required fields or the provided credentials do not match.
	AuthenticationFailedReason string = "AuthenticationFailed"

	// VerificationFailedReason represents the fact that the cryptographic
	// provenance verification for the source failed.
	VerificationFailedReason string = "VerificationFailed"

	// IndexationFailedReason represents the fact that the indexation of the given
	// Helm repository failed.
	IndexationFailedReason string = "IndexationFailed"

	// IndexationSucceededReason represents the fact that the indexation of the
	// given Helm repository succeeded.
	IndexationSucceededReason string = "IndexationSucceed"

	// ChartPullFailedReason represents the fact that the pull of the Helm chart
	// failed.
	ChartPullFailedReason string = "ChartPullFailed"

	// ChartPullSucceededReason represents the fact that the pull of the Helm chart
	// succeeded.
	ChartPullSucceededReason string = "ChartPullSucceeded"

	// ChartPackageFailedReason represent the fact that the package of the Helm
	// chart failed.
	ChartPackageFailedReason string = "ChartPackageFailed"

	// ChartPackageSucceededReason represents the fact that the package of the Helm
	// chart succeeded.
	ChartPackageSucceededReason string = "ChartPackageSucceeded"

	// InstallSucceededReason represents the fact that the Helm install for the
	// HelmRelease succeeded.
	InstallSucceededReason string = "InstallSucceeded"

	// InstallFailedReason represents the fact that the Helm install for the
	// HelmRelease failed.
	InstallFailedReason string = "InstallFailed"

	// UpgradeSucceededReason represents the fact that the Helm upgrade for the
	// HelmRelease succeeded.
	UpgradeSucceededReason string = "UpgradeSucceeded"

	// UpgradeFailedReason represents the fact that the Helm upgrade for the
	// HelmRelease failed.
	UpgradeFailedReason string = "UpgradeFailed"

	// TestSucceededReason represents the fact that the Helm tests for the
	// HelmRelease succeeded.
	TestSucceededReason string = "TestSucceeded"

	// TestFailedReason represents the fact that the Helm tests for the HelmRelease
	// failed.
	TestFailedReason string = "TestFailed"

	// RollbackSucceededReason represents the fact that the Helm rollback for the
	// HelmRelease succeeded.
	RollbackSucceededReason string = "RollbackSucceeded"

	// RollbackFailedReason represents the fact that the Helm test for the
	// HelmRelease failed.
	RollbackFailedReason string = "RollbackFailed"

	// UninstallSucceededReason represents the fact that the Helm uninstall for the
	// HelmRelease succeeded.
	UninstallSucceededReason string = "UninstallSucceeded"

	// UninstallFailedReason represents the fact that the Helm uninstall for the
	// HelmRelease failed.
	UninstallFailedReason string = "UninstallFailed"

	// ArtifactFailedReason represents the fact that the artifact download for the
	// HelmRelease failed.
	ArtifactFailedReason string = "ArtifactFailed"

	// InitFailedReason represents the fact that the initialization of the Helm
	// configuration failed.
	InitFailedReason string = "InitFailed"

	// GetLastReleaseFailedReason represents the fact that observing the last
	// release failed.
	GetLastReleaseFailedReason string = "GetLastReleaseFailed"

	// ReleasedCondition represents the status of the last release attempt
	// (install/upgrade/test) against the latest desired state.
	ReleasedCondition string = "Released"

	// TestSuccessCondition represents the status of the last test attempt against
	// the latest desired state.
	TestSuccessCondition string = "TestSuccess"

	// RemediatedCondition represents the status of the last remediation attempt
	// (uninstall/rollback) due to a failure of the last release attempt against the
	// latest desired state.
	RemediatedCondition string = "Remediated"

	ObjectsAppliedCondition     string = "ObjectApplied"
	ObjectsAppliedSuccessReason string = "ObjectAppliedSuccess"
	ObjectsApliedFailedReason   string = "ObjectAppliedFailed"

	ChartAppliedCondition     string = "ChartApplied"
	ChartAppliedSuccessReason string = "ChartAppliedSuccess"
	ChartAppliedFailedReason  string = "ChartAppliedFailed"

	WaitChartReason     string = "WaitChart"
	WaitProvisionReason string = "WaitClusterProvision"

	CNIInstalledCondition     string = "CNIInstalled"
	CNIInstalledSuccessReason string = "CNIInstalledSuccess"
	CNIInstalledFailedReason  string = "CNIInstalledFailed"
	TemplateAppliedFailed     string = "TemplateAppliedFailed"
	ReconcileNodesFailed      string = "ReconcileNodesFailed"
)

// InReadyCondition returns if the given Condition slice has a ReadyCondition
// with a 'True' condition status.
func InReadyCondition(conditions []metav1.Condition) bool {
	return apimeta.IsStatusConditionTrue(conditions, ReadyCondition)
}

// ObjectWithStatusConditions is an interface that describes kubernetes resource
// type structs with Status Conditions
type ObjectWithStatusConditions interface {
	runtime.Object
	GetStatusConditions() *[]metav1.Condition
}

// SetResourceCondition sets the given condition with the given status,
// reason and message on a resource.
func SetResourceCondition(obj ObjectWithStatusConditions, condition string, status metav1.ConditionStatus, reason, message string) {
	conditions := obj.GetStatusConditions()

	newCondition := metav1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: message,
	}

	apimeta.SetStatusCondition(conditions, newCondition)
	if status == metav1.ConditionFalse {
		record.Warn(obj, reason, message)
	} else {
		record.Event(obj, reason, message)
	}
}
