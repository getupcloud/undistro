/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package helm

import (
	"time"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Client is the generic interface for Helm (v2 and v3) clients.
type Client interface {
	Get(releaseName string, opts GetOptions) (*Release, error)
	Status(releaseName string, opts StatusOptions) (Status, error)
	UpgradeFromPath(chartPath string, releaseName string, values []byte, opts UpgradeOptions) (*Release, error)
	History(releaseName string, opts HistoryOptions) ([]*Release, error)
	Rollback(releaseName string, opts RollbackOptions) (*Release, error)
	Test(releaseName string, opts TestOptions) error
	DependencyUpdate(chartPath string) error
	RepositoryIndex() error
	RepositoryAdd(name, url, username, password, certFile, keyFile, caFile string) error
	RepositoryRemove(name string) error
	RepositoryImport(path string) error
	Pull(ref, version, dest string) (string, error)
	PullWithRepoURL(repoURL, name, version, dest string) (string, error)
	Uninstall(releaseName string, opts UninstallOptions) error
	GetChartRevision(chartPath string) (string, error)
	Version() string
	EnsureChartFetched(base string, source *undistrov1.RepoChartSource) (string, bool, error)
}

func Diff(j *Release, k *Release) string {
	return cmp.Diff(j.Values, k.Values) + cmp.Diff(j.Chart, k.Chart)
}

// GetOptions holds the options available for Helm get
// operations, the version implementation _must_ implement all
// fields supported by that version but can (silently) ignore
// unsupported set values.
type GetOptions struct {
	Namespace string
	Version   int
}

// StatusOptions holds the options available for Helm status
// operations, the version implementation _must_ implement all
// fields supported by that version but can (silently) ignore
// unsupported set values.
type StatusOptions struct {
	Namespace string
	Version   int
}

// UpgradeOptions holds the options available for Helm upgrade
// operations, the version implementation _must_ implement all
// fields supported by that version but can (silently) ignore
// unsupported set values.
type UpgradeOptions struct {
	Namespace         string
	Timeout           time.Duration
	Wait              bool
	Install           bool
	DisableHooks      bool
	DryRun            bool
	ClientOnly        bool
	Force             bool
	ResetValues       bool
	SkipCRDs          bool
	ReuseValues       bool
	Recreate          bool
	MaxHistory        int
	Atomic            bool
	DisableValidation bool
}

// RollbackOptions holds the options available for Helm rollback
// operations, the version implementation _must_ implement all
// fields supported by that version but can (silently) ignore
// unsupported set values.
type RollbackOptions struct {
	Namespace    string
	Version      int
	Timeout      time.Duration
	Wait         bool
	DisableHooks bool
	DryRun       bool
	Recreate     bool
	Force        bool
}

// TestOptions holds the options available for Helm test
// operations, the version implementation _must_ implement all
// fields supported by that version but can (silently) ignore
// unsupported set values.
type TestOptions struct {
	Namespace string
	Cleanup   bool
	Timeout   time.Duration
}

// UninstallOptions holds the options available for Helm uninstall
// operations, the version implementation _must_ implement all
// fields supported by that version but can (silently) ignore
// unsupported set values.
type UninstallOptions struct {
	Namespace    string
	DisableHooks bool
	DryRun       bool
	KeepHistory  bool
	Timeout      time.Duration
}

// HistoryOption holds the options available for Helm history
// operations, the version implementation _must_ implement all
// fields supported by that version but can (silently) ignore
// unsupported set values.
type HistoryOptions struct {
	Namespace string
	Max       int
}

// Define release statuses
const (
	// StatusUnknown indicates that a release is in an uncertain state
	StatusUnknown Status = "unknown"
	// StatusDeployed indicates that the release has been pushed to Kubernetes
	StatusDeployed Status = "deployed"
	// StatusUninstalled indicates that a release has been uninstalled from Kubernetes
	StatusUninstalled Status = "uninstalled"
	// StatusSuperseded indicates that this release object is outdated and a newer one exists
	StatusSuperseded Status = "superseded"
	// StatusFailed indicates that the release was not successfully deployed
	StatusFailed Status = "failed"
	// StatusUninstalling indicates that a uninstall operation is underway
	StatusUninstalling Status = "uninstalling"
	// StatusPendingInstall indicates that an install operation is underway
	StatusPendingInstall Status = "pending-install"
	// StatusPendingUpgrade indicates that an upgrade operation is underway
	StatusPendingUpgrade Status = "pending-upgrade"
	// StatusPendingRollback indicates that an rollback operation is underway
	StatusPendingRollback Status = "pending-rollback"
)

// Release describes a generic chart deployment
type Release struct {
	Name      string
	Namespace string
	Chart     *Chart
	Info      *Info
	Values    map[string]interface{}
	Manifest  string
	Version   int
}

// Info holds metadata of a chart deployment
type Info struct {
	LastDeployed time.Time
	Description  string
	Status       Status
}

// Chart describes the chart for a release
type Chart struct {
	Name       string
	Version    string
	AppVersion string
	Values     chartutil.Values
	Templates  []*File
}

// File represents a file as a name/value pair.
// The name is a relative path within the scope
// of the chart's base directory.
type File struct {
	Name string
	Data []byte
}

// Status holds the status of a release
type Status string

// AllowsUpgrade returns true if the status allows the release
// to be upgraded. This is currently only the case if it equals
// `StatusDeployed`.
func (s Status) AllowsUpgrade() bool {
	return s == StatusDeployed
}

// String returns the Status as a string
func (s Status) String() string {
	return string(s)
}
