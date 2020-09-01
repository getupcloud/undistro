package status

import (
	"context"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetReleaseStatus updates the status of the HelmRelease to the given
// release name and/or release status.
func SetReleaseStatus(ctx context.Context, client client.Client, hr *undistrov1.HelmRelease, releaseName, releaseStatus string) error {

	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if hr.Status.ReleaseName == releaseName && hr.Status.ReleaseStatus == releaseStatus {
			return
		}

		cHr := hr.DeepCopy()
		cHr.Status.ReleaseName = releaseName
		cHr.Status.ReleaseStatus = releaseStatus

		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

// SetReleaseRevision updates the revision in the status of the HelmRelease
// to the given revision, and sets the current revision as the previous one.
func SetReleaseRevision(ctx context.Context, client client.Client, hr *undistrov1.HelmRelease, revision string) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if revision == "" || hr.Status.Revision == revision {
			return
		}

		cHr := hr.DeepCopy()
		cHr.Status.Revision = revision

		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

// SetObservedGeneration updates the observed generation status of the
// HelmRelease to the given generation.
func SetObservedGeneration(ctx context.Context, client client.Client, hr *undistrov1.HelmRelease, generation int64) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if hr.Status.ObservedGeneration >= generation {
			return
		}

		cHr := hr.DeepCopy()
		cHr.Status.ObservedGeneration = generation

		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

func SetObservedGenerationCluster(ctx context.Context, client client.Client, hr *undistrov1.Cluster, generation int64) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if hr.Status.ObservedGeneration >= generation {
			return
		}

		cHr := hr.DeepCopy()
		cHr.Status.ObservedGeneration = generation

		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

// HasSynced returns if the HelmRelease has been processed by the
// controller.
func HasSynced(hr *undistrov1.HelmRelease) bool {
	return hr.Status.ObservedGeneration >= hr.Generation
}

// HasRolledBack returns if the current generation of the HelmRelease
// has been rolled back.
func HasRolledBack(hr *undistrov1.HelmRelease) bool {
	if !HasSynced(hr) {
		return false
	}

	rolledBack := GetCondition(hr.Status, undistrov1.HelmReleaseRolledBack)
	if rolledBack == nil {
		return false
	}

	return rolledBack.Status == undistrov1.ConditionTrue
}

// ShouldRetryUpgrade returns if the upgrade of a rolled back release should
// be retried.
func ShouldRetryUpgrade(hr *undistrov1.HelmRelease) bool {
	if !hr.Spec.Rollback.Retry {
		return false
	}
	return hr.Spec.Rollback.GetMaxRetries() == 0 || hr.Status.RollbackCount <= hr.Spec.Rollback.GetMaxRetries()
}

func SetClusterPhaseWithConditons(ctx context.Context, client client.Client, hr *undistrov1.Cluster, phase undistrov1.ClusterPhase, infraReady, controlplaneInit, controlplaneReady bool) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if hr.Status.InfrastructureReady == infraReady &&
			hr.Status.ControlPlaneInitialized == controlplaneInit &&
			hr.Status.ControlPlaneReady == controlplaneReady &&
			hr.Status.Phase == phase {
			return
		}

		cHr := hr.DeepCopy()
		hr.Status.Phase = phase
		cHr.Status.InfrastructureReady = infraReady
		cHr.Status.ControlPlaneInitialized = controlplaneInit
		cHr.Status.ControlPlaneReady = controlplaneReady
		cHr.Status.Ready = infraReady && controlplaneInit && controlplaneReady
		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

func SetClusterPhaseWithCompoments(ctx context.Context, client client.Client, hr *undistrov1.Cluster, phase undistrov1.ClusterPhase, components []undistrov1.InstalledComponent) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if hr.Status.Phase == phase {
			return
		}

		cHr := hr.DeepCopy()
		hr.Status.Phase = phase
		cHr.Status.InstalledComponents = components
		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

func SetClusterPhaseWithCapi(ctx context.Context, client client.Client, hr *undistrov1.Cluster, phase undistrov1.ClusterPhase, o unstructured.Unstructured) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if hr.Status.Phase == phase {
			return
		}

		cHr := hr.DeepCopy()
		hr.Status.Phase = phase
		cHr.Status.ClusterAPIRef = &corev1.ObjectReference{
			Kind:            o.GetKind(),
			Namespace:       o.GetNamespace(),
			Name:            o.GetName(),
			UID:             o.GetUID(),
			APIVersion:      o.GetAPIVersion(),
			ResourceVersion: o.GetResourceVersion(),
		}
		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

func SetClusterPhase(ctx context.Context, client client.Client, hr *undistrov1.Cluster, phase undistrov1.ClusterPhase) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			nm := types.NamespacedName{
				Name:      hr.Name,
				Namespace: hr.Namespace,
			}
			getErr := client.Get(ctx, nm, hr)
			if getErr != nil {
				return getErr
			}
		}

		if hr.Status.Phase == phase {
			return
		}

		cHr := hr.DeepCopy()
		hr.Status.Phase = phase
		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}
