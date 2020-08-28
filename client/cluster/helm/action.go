package helm

import (
	"context"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/pkg/errors"
)

func ChooseAction(ctx context.Context, h Client, hr *undistrov1.HelmRelease, chart ChartState, synced, hasRollback, retry bool) (undistrov1.HelmAction, *Release, error) {
	curRel, err := h.Get(hr.GetReleaseName(), GetOptions{Namespace: hr.GetTargetNamespace()})
	if err != nil {
		return undistrov1.SkipAction, nil, errors.Errorf("couldn't to retrieve Helm release: %v", err)
	}
	if curRel == nil {
		return undistrov1.InstallAction, nil, nil
	}
	// If the current state of the release does not allow us to safely
	// upgrade, we skip.
	if s := curRel.Info.Status; !s.AllowsUpgrade() {
		return undistrov1.SkipAction, nil, errors.Errorf("status '%s' of release does not allow a safe upgrade", s.String())
	}

	if !hr.DeletionTimestamp.IsZero() {
		return undistrov1.UninstallAction, curRel, nil
	}

	// If this revision of the `HelmRelease` has not been synchronized
	// yet, we attempt an upgrade.
	if !synced {
		return undistrov1.UpgradeAction, curRel, nil
	}
	// The release has been rolled back, inspect state.
	if hasRollback {
		if chart.Changed || retry {
			return undistrov1.UpgradeAction, curRel, nil
		}
		hist, err := h.History(hr.GetReleaseName(), HistoryOptions{Namespace: hr.GetTargetNamespace(), Max: hr.GetMaxHistory()})
		if err != nil {
			return undistrov1.SkipAction, nil, errors.Errorf("couldn't to retreive history for rolled back release: %v", err)
		}
		for _, r := range hist {
			if r.Info.Status == StatusFailed || r.Info.Status == StatusSuperseded {
				curRel = r
				break
			}
		}
	} else if chart.Changed {
		return undistrov1.UpgradeAction, curRel, nil
	}
	return undistrov1.DryRunCompareAction, curRel, nil
}
