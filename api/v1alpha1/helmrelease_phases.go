package v1alpha1

// HelmReleasePhase represents the phase a HelmRelease is in.
// +optional
type HelmReleasePhase string

const (
	// ChartFetched means the chart to which the HelmRelease refers
	// has been fetched successfully
	HelmReleasePhaseChartFetched HelmReleasePhase = "ChartFetched"
	// ChartFetchedFailed means the chart to which the HelmRelease
	// refers could not be fetched.
	HelmReleasePhaseChartFetchFailed HelmReleasePhase = "ChartFetchFailed"

	// Installing means the installation for the HelmRelease is running.
	HelmReleasePhaseInstalling HelmReleasePhase = "Installing"
	// Migrating means the HelmRelease is converting from one version to another
	HelmReleasePhaseMigrating HelmReleasePhase = "Migrating"
	// Upgrading means the upgrade for the HelmRelease is running.
	HelmReleasePhaseUpgrading HelmReleasePhase = "Upgrading"
	// Deployed means the dry-run, installation, or upgrade for the
	// HelmRelease succeeded.
	HelmReleasePhaseDeployed HelmReleasePhase = "Deployed"
	// DeployFailed means the dry-run, installation, or upgrade for the
	// HelmRelease failed.
	HelmReleasePhaseDeployFailed HelmReleasePhase = "DeployFailed"

	// Testing means a test for the HelmRelease is running.
	HelmReleasePhaseTesting HelmReleasePhase = "Testing"
	// TestFailed means the test for the HelmRelease failed.
	HelmReleasePhaseTestFailed HelmReleasePhase = "TestFailed"
	// Tested means the test for the HelmRelease succeeded.
	HelmReleasePhaseTested HelmReleasePhase = "Tested"

	HelmReleasePhaseUninstalling HelmReleasePhase = "Uninstalling"
	HelmReleasePhaseUninstalled  HelmReleasePhase = "Uninstalled"

	// Succeeded means the chart release, as specified in this
	// HelmRelease, has been processed by Helm.
	HelmReleasePhaseSucceeded HelmReleasePhase = "Succeeded"
	// Failed means the installation or upgrade for the HelmRelease
	// failed.
	HelmReleasePhaseFailed HelmReleasePhase = "Failed"

	// RollingBack means a rollback for the HelmRelease is running.
	HelmReleasePhaseRollingBack HelmReleasePhase = "RollingBack"
	// RolledBack means the HelmRelease has been rolled back.
	HelmReleasePhaseRolledBack HelmReleasePhase = "RolledBack"
	// RolledBackFailed means the rollback for the HelmRelease failed.
	HelmReleasePhaseRollbackFailed HelmReleasePhase = "RollbackFailed"

	// WaitClusterReady wait cluster to be ready
	HelmReleasePhaseWaitClusterReady HelmReleasePhase = "ClusterNotReady"
	// HelmReleasePhaseWaitDependency wait dependencies to be installed
	HelmReleasePhaseWaitDependency HelmReleasePhase = "WaitingDependencies"
)
