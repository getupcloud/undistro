package cluster

import (
	"context"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
)

func ChooseAction(ctx context.Context, cl *undistrov1.Cluster) undistrov1.ClusterAction {
	if cl.Status.Phase == undistrov1.NewPhase {
		return undistrov1.InitClusterAction
	}
	if !cl.DeletionTimestamp.IsZero() {
		return undistrov1.DeleteClusterAction
	}
	if cl.Status.ClusterAPIRef == nil {
		return undistrov1.ProvisionClusterAction
	}
	if cl.Status.ControlPlaneInitialized && !cl.Status.ControlPlaneReady {
		return undistrov1.InstallNetClusterAction
	}
	if cl.Status.Ready {
		return undistrov1.UpgradeClusterAction
	}
	return undistrov1.StatusClusterAction
}
