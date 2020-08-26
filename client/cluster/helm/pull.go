package helm

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/internal/urlutil"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
)

func (h *HelmV3) Rollback(releaseName string, opts RollbackOptions) (*Release, error) {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return nil, err
	}

	rollback := action.NewRollback(cfg)
	opts.configure(rollback)

	if err := rollback.Run(releaseName); err != nil {
		return nil, err
	}

	// As rolling back does no longer return information about
	// the release in v3, we need to make an additional call to
	// get the release we rolled back to.
	return h.Get(releaseName, GetOptions{Namespace: opts.Namespace})
}

func (opts RollbackOptions) configure(action *action.Rollback) {
	action.Timeout = opts.Timeout
	action.Version = opts.Version
	action.Wait = opts.Wait
	action.DisableHooks = opts.DisableHooks
	action.DryRun = opts.DryRun
	action.Recreate = opts.Recreate
}

func (h *HelmV3) Pull(ref, version, dest string) (string, error) {
	repositoryConfigLock.RLock()
	defer repositoryConfigLock.RUnlock()

	c := downloader.ChartDownloader{
		Out:              os.Stdout,
		Verify:           downloader.VerifyNever,
		RepositoryConfig: repositoryConfig,
		RepositoryCache:  repositoryCache,
		Getters:          getterProviders(),
	}
	d, _, err := c.DownloadTo(ref, version, dest)
	return d, err
}

func (h *HelmV3) PullWithRepoURL(repoURL, name, version, dest string) (string, error) {
	// This first attempts to look up the repository name by the given
	// `repoURL`, if found the repository name and given chart name
	// are used to construct a `chartRef` Helm understands.
	//
	// If no repository is found it attempts to resolve the absolute
	// URL to the chart by making a request to the given `repoURL`,
	// this absolute URL is then used to instruct Helm to pull the
	// chart.

	repositoryConfigLock.RLock()
	repoFile, err := loadRepositoryConfig()
	repositoryConfigLock.RUnlock()
	if err != nil {
		return "", err
	}

	// Here we attempt to find an entry for the repository. If found the
	// entry's name is used to construct a `chartRef` Helm understands.
	var chartRef string
	for _, entry := range repoFile.Repositories {
		if urlutil.Equal(repoURL, entry.URL) {
			chartRef = entry.Name + "/" + name
			// Ensure we have the repository index as this is
			// later used by Helm.
			if r, err := newChartRepository(entry); err == nil {
				r.DownloadIndexFile()
			}
			break
		}
	}

	if chartRef == "" {
		// We were unable to find an entry so we need to make a request
		// to the repository to get the absolute URL of the chart.
		chartRef, err = repo.FindChartInRepoURL(repoURL, name, version, "", "", "", getterProviders())
		if err != nil {
			return "", err
		}

		// As Helm also attempts to find credentials for the absolute URL
		// we give to it, and does not ignore missing index files, we need
		// to be sure all indexes files are present, and we are only able
		// to do so by updating our indexes.
		if err := downloadMissingRepositoryIndexes(repoFile.Repositories); err != nil {
			return "", err
		}
	}

	return h.Pull(chartRef, version, dest)
}

func downloadMissingRepositoryIndexes(repositories []*repo.Entry) error {
	var wg sync.WaitGroup
	for _, c := range repositories {
		r, err := newChartRepository(c)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func(r *repo.ChartRepository) {
			f := filepath.Join(r.CachePath, helmpath.CacheIndexFile(r.Config.Name))
			if _, err := os.Stat(f); os.IsNotExist(err) {
				r.DownloadIndexFile()
			}
			wg.Done()
		}(r)
	}
	wg.Wait()
	return nil
}

func (h *HelmV3) EnsureChartFetched(base string, source *undistrov1.RepoChartSource) (string, bool, error) {
	repoPath, filename, err := makeChartPath(base, h.Version(), source)
	if err != nil {
		return "", false, ChartUnavailableError{err}
	}
	chartPath := filepath.Join(repoPath, filename)
	stat, err := os.Stat(chartPath)
	switch {
	case os.IsNotExist(err):
		chartPath, err = h.PullWithRepoURL(source.RepoURL, source.Name, source.Version, chartPath)
		if err != nil {
			return chartPath, false, ChartUnavailableError{err}
		}
		return chartPath, true, nil
	case err != nil:
		return chartPath, false, ChartUnavailableError{err}
	case stat.IsDir():
		return chartPath, false, ChartUnavailableError{errors.New("path to chart exists but is a directory")}
	}
	return chartPath, false, nil
}

// makeChartPath gives the expected filesystem location for a chart,
// without testing whether the file exists or not.
func makeChartPath(base string, clientVersion string, source *undistrov1.RepoChartSource) (string, string, error) {
	// We don't need to obscure the location of the charts in the
	// filesystem; but we do need a stable, filesystem-friendly path
	// to them that is based on the URL and the client version.
	repoPath := filepath.Join("helmcharts", base, clientVersion, base64.URLEncoding.EncodeToString([]byte(source.CleanRepoURL())))
	if err := os.MkdirAll(repoPath, 00750); err != nil {
		return "", "", err
	}
	filename := fmt.Sprintf("%s-%s.tgz", source.Name, source.Version)
	return repoPath, filename, nil
}
