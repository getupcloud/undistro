package helm

import (
	"fmt"
	"sort"

	"github.com/ncabatoff/go-seq/seq"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// releaseToGenericRelease transforms a v3 release structure
// into a generic `helm.Release`
func releaseToGenericRelease(r *release.Release) *Release {
	return &Release{
		Name:      r.Name,
		Namespace: r.Namespace,
		Chart:     chartToGenericChart(r.Chart),
		Info:      infoToGenericInfo(r.Info),
		Values:    configToGenericValues(r.Config),
		Manifest:  r.Manifest,
		Version:   r.Version,
	}
}

// chartToGenericChart transforms a v3 chart structure into
// a generic `helm.Chart`
func chartToGenericChart(c *chart.Chart) *Chart {
	return &Chart{
		Name:       c.Name(),
		Version:    formatVersion(c),
		AppVersion: c.AppVersion(),
		Values:     c.Values,
		Templates:  filesToGenericFiles(c.Templates),
	}
}

// filesToGenericFiles transforms a `chart.File` slice into
// an stable sorted slice with generic `helm.File`s
func filesToGenericFiles(f []*chart.File) []*File {
	gf := make([]*File, len(f))
	for i, ff := range f {
		gf[i] = &File{Name: ff.Name, Data: ff.Data}
	}
	sort.SliceStable(gf, func(i, j int) bool {
		return seq.Compare(gf[i], gf[j]) > 0
	})
	return gf
}

// infoToGenericInfo transforms a v3 info structure into
// a generic `helm.Info`
func infoToGenericInfo(i *release.Info) *Info {
	return &Info{
		LastDeployed: i.LastDeployed.Time,
		Description:  i.Description,
		Status:       lookUpGenericStatus(i.Status),
	}
}

// configToGenericValues forces the `chartutil.Values` to be parsed
// as YAML so that the value types of the nested map always equal to
// what they would be when returned from storage, as a dry-run release
// seems to skip this step, resulting in differences for e.g. float
// values (as they will be string values when returned from storage).
// TODO(hidde): this may not be necessary for v3.0.0 (stable).
func configToGenericValues(c chartutil.Values) map[string]interface{} {
	s, err := c.YAML()
	if err != nil {
		return c.AsMap()
	}
	gv, err := chartutil.ReadValues([]byte(s))
	if err != nil {
		return c.AsMap()
	}
	return gv.AsMap()
}

// formatVersion formats the chart version, while taking
// into account that the metadata may actually be missing
// due to unknown reasons.
// https://github.com/kubernetes/helm/issues/1347
func formatVersion(c *chart.Chart) string {
	if c.Metadata == nil {
		return ""
	}
	return c.Metadata.Version
}

// lookUpGenericStatus looks up the generic status for the
// given v3 release status.
func lookUpGenericStatus(s release.Status) Status {
	var statuses = map[release.Status]Status{
		release.StatusUnknown:         StatusUnknown,
		release.StatusDeployed:        StatusDeployed,
		release.StatusUninstalled:     StatusUninstalled,
		release.StatusSuperseded:      StatusSuperseded,
		release.StatusFailed:          StatusFailed,
		release.StatusUninstalling:    StatusUninstalling,
		release.StatusPendingInstall:  StatusPendingInstall,
		release.StatusPendingUpgrade:  StatusPendingUpgrade,
		release.StatusPendingRollback: StatusPendingRollback,
	}
	if status, ok := statuses[s]; ok {
		return status
	}
	return StatusUnknown
}

func (h *HelmV3) GetChartRevision(chartPath string) (string, error) {
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to load chart to determine revision: %w", err)
	}
	return chartRequested.Metadata.Version, nil
}

type (
	getOptions    GetOptions
	statusOptions GetOptions
)

func (h *HelmV3) Get(releaseName string, opts GetOptions) (*Release, error) {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return nil, err
	}

	get := action.NewGet(cfg)
	getOptions(opts).configure(get)

	res, err := get.Run(releaseName)
	switch err {
	case nil:
		return releaseToGenericRelease(res), nil
	case driver.ErrReleaseNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (opts getOptions) configure(action *action.Get) {
	action.Version = opts.Version
}

func (h *HelmV3) Status(releaseName string, opts StatusOptions) (Status, error) {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return "", err
	}

	status := action.NewStatus(cfg)
	statusOptions(opts).configure(status)

	res, err := status.Run(releaseName)
	switch err {
	case nil:
		return lookUpGenericStatus(res.Info.Status), nil
	default:
		return "", err
	}
}

func (opts statusOptions) configure(action *action.Status) {
	action.Version = opts.Version
}
