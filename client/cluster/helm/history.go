package helm

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/releaseutil"
)

func (h *HelmV3) History(releaseName string, opts HistoryOptions) ([]*Release, error) {
	cfg, err := newActionConfig(h.path, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return nil, err
	}

	history := action.NewHistory(cfg)
	opts.configure(history)

	hist, err := history.Run(releaseName)
	if err != nil {
		return nil, err
	}

	releaseutil.Reverse(hist, releaseutil.SortByRevision)

	var rels []*Release
	for i := 0; i < min(len(hist), history.Max); i++ {
		rels = append(rels, releaseToGenericRelease(hist[i]))
	}
	return rels, nil
}

func (opts HistoryOptions) configure(action *action.History) {
	action.Max = opts.Max
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
