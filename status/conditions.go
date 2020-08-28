package status

import (
	"context"
	"fmt"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Clock is defined as a var so it can be stubbed during tests.
var Clock clock.Clock = clock.RealClock{}

func GetCondition(status undistrov1.HelmReleaseStatus, conditionType undistrov1.HelmReleaseConditionType) *undistrov1.HelmReleaseCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

func SetConditions(ctx context.Context, client client.Client, hr *undistrov1.HelmRelease, conditions []undistrov1.HelmReleaseCondition, setters ...func(*undistrov1.HelmRelease)) error {
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

		cHr := hr.DeepCopy()
		for _, condition := range conditions {
			currCondition := GetCondition(hr.Status, condition.Type)
			if currCondition != nil && currCondition.Status == condition.Status {
				condition.LastTransitionTime = currCondition.LastTransitionTime
			}

			cHr.Status.Conditions = append(filterOutCondition(cHr.Status.Conditions, condition.Type), condition)
			switch {
			case condition.Type == undistrov1.HelmReleaseReleased && condition.Status == undistrov1.ConditionTrue:
				cHr.Status.Conditions = filterOutCondition(cHr.Status.Conditions, undistrov1.HelmReleaseRolledBack)
				cHr.Status.RollbackCount = 0
			case condition.Type == undistrov1.HelmReleaseRolledBack && condition.Status == undistrov1.ConditionTrue:
				cHr.Status.RollbackCount = cHr.Status.RollbackCount + 1
			}
		}
		for _, setter := range setters {
			setter(cHr)
		}

		err = client.Status().Update(ctx, cHr)
		firstTry = false
		return
	})
	return err
}

func SetStatusPhase(ctx context.Context, client client.Client, hr *undistrov1.HelmRelease, phase undistrov1.HelmReleasePhase, setters ...func(*undistrov1.HelmRelease)) error {
	conditions, ok := ConditionsForPhase(hr, phase)
	if !ok {
		return nil
	}
	setters = append(setters, func(cHr *undistrov1.HelmRelease) {
		cHr.Status.Phase = phase
	})
	return SetConditions(ctx, client, hr, conditions, setters...)
}

func SetStatusPhaseWithRevision(ctx context.Context, client client.Client, hr *undistrov1.HelmRelease, phase undistrov1.HelmReleasePhase, revision string) error {
	return SetStatusPhase(ctx, client, hr, phase, func(cHr *undistrov1.HelmRelease) {
		switch {
		case phase == undistrov1.HelmReleasePhaseInstalling || phase == undistrov1.HelmReleasePhaseUpgrading:
			cHr.Status.LastAttemptedRevision = revision
		case phase == undistrov1.HelmReleasePhaseSucceeded:
			cHr.Status.Revision = revision
		}
	})
}

// ConditionsForPhrase returns conditions for the given phase.
func ConditionsForPhase(hr *undistrov1.HelmRelease, phase undistrov1.HelmReleasePhase) ([]undistrov1.HelmReleaseCondition, bool) {
	condition := &undistrov1.HelmReleaseCondition{}
	conditions := []*undistrov1.HelmReleaseCondition{condition}
	switch phase {
	case undistrov1.HelmReleasePhaseInstalling:
		condition.Type = undistrov1.HelmReleaseDeployed
		condition.Status = undistrov1.ConditionUnknown
		condition.Message = fmt.Sprintf(`Running installation for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseUpgrading:
		condition.Type = undistrov1.HelmReleaseDeployed
		condition.Status = undistrov1.ConditionUnknown
		condition.Message = fmt.Sprintf(`Running upgrade for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseDeployed:
		condition.Type = undistrov1.HelmReleaseDeployed
		condition.Status = undistrov1.ConditionTrue
		condition.Message = fmt.Sprintf(`Installation or upgrade succeeded for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseDeployFailed:
		message := fmt.Sprintf(`Installation or upgrade failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
		condition.Type = undistrov1.HelmReleaseDeployed
		condition.Status = undistrov1.ConditionFalse
		condition.Message = message
		conditions = append(conditions, &undistrov1.HelmReleaseCondition{
			Type:    undistrov1.HelmReleaseReleased,
			Status:  undistrov1.ConditionFalse,
			Message: message,
		})
	case undistrov1.HelmReleasePhaseSucceeded:
		condition.Type = undistrov1.HelmReleaseReleased
		condition.Status = undistrov1.ConditionTrue
		condition.Message = fmt.Sprintf(`Release was successful for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseFailed:
		condition.Type = undistrov1.HelmReleaseReleased
		condition.Status = undistrov1.ConditionFalse
		condition.Message = fmt.Sprintf(`Release failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseTesting:
		condition.Type = undistrov1.HelmReleaseTested
		condition.Status = undistrov1.ConditionUnknown
		condition.Message = fmt.Sprintf(`Testing Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseTested:
		condition.Type = undistrov1.HelmReleaseTested
		condition.Status = undistrov1.ConditionTrue
		condition.Message = fmt.Sprintf(`Tested Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseTestFailed:
		message := fmt.Sprintf(`Test failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
		condition.Type = undistrov1.HelmReleaseTested
		condition.Status = undistrov1.ConditionFalse
		condition.Message = message
		if !hr.Spec.Test.GetIgnoreFailures() {
			conditions = append(conditions, &undistrov1.HelmReleaseCondition{
				Type:    undistrov1.HelmReleaseReleased,
				Status:  undistrov1.ConditionFalse,
				Message: message,
			})
		}
	case undistrov1.HelmReleasePhaseRollingBack:
		condition.Type = undistrov1.HelmReleaseRolledBack
		condition.Status = undistrov1.ConditionUnknown
		condition.Message = fmt.Sprintf(`Rolling back Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseRolledBack:
		condition.Type = undistrov1.HelmReleaseRolledBack
		condition.Status = undistrov1.ConditionTrue
		condition.Message = fmt.Sprintf(`Rolled back Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseRollbackFailed:
		condition.Type = undistrov1.HelmReleaseRolledBack
		condition.Status = undistrov1.ConditionFalse
		condition.Message = fmt.Sprintf(`Rollback failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseChartFetched:
		condition.Type = undistrov1.HelmReleaseChartFetched
		condition.Status = undistrov1.ConditionTrue
		condition.Message = fmt.Sprintf(`Chart fetch was successful for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case undistrov1.HelmReleasePhaseChartFetchFailed:
		message := fmt.Sprintf(`Chart fetch failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
		condition.Type = undistrov1.HelmReleaseChartFetched
		condition.Status = undistrov1.ConditionFalse
		condition.Message = message
		conditions = append(conditions, &undistrov1.HelmReleaseCondition{
			Type:    undistrov1.HelmReleaseReleased,
			Status:  undistrov1.ConditionFalse,
			Message: message,
		})
	default:
		return []undistrov1.HelmReleaseCondition{}, false
	}
	nowTime := metav1.NewTime(Clock.Now())
	updatedConditions := []undistrov1.HelmReleaseCondition{}
	for _, c := range conditions {
		c.Reason = string(phase)
		c.LastUpdateTime = &nowTime
		c.LastTransitionTime = &nowTime
		updatedConditions = append(updatedConditions, *c)
	}

	return updatedConditions, true
}

// filterOutCondition returns a new slice of condition without the
// condition of the given type.
func filterOutCondition(conditions []undistrov1.HelmReleaseCondition,
	conditionType undistrov1.HelmReleaseConditionType) []undistrov1.HelmReleaseCondition {

	var newConditions []undistrov1.HelmReleaseCondition
	for _, c := range conditions {
		if c.Type == conditionType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
