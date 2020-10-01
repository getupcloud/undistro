/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"strings"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client/config"
	"github.com/getupcloud/undistro/internal/scheme"
	logf "github.com/getupcloud/undistro/log"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/cluster-api/util/yaml"
	apiyaml "sigs.k8s.io/yaml"
)

// MetadataClient has methods to work with metadata hosted on a provider repository.
// Metadata are yaml files providing additional information about provider's assets like e.g the version compatibility Matrix.
type MetadataClient interface {
	// Get returns the provider's metadata.
	Get() (*undistrov1.Metadata, error)
}

// metadataClient implements MetadataClient.
type metadataClient struct {
	configVarClient config.VariablesClient
	provider        config.Provider
	version         string
	repository      Repository
}

// ensure metadataClient implements MetadataClient.
var _ MetadataClient = &metadataClient{}

// newMetadataClient returns a metadataClient.
func newMetadataClient(provider config.Provider, version string, repository Repository, config config.VariablesClient) *metadataClient {
	return &metadataClient{
		configVarClient: config,
		provider:        provider,
		version:         version,
		repository:      repository,
	}
}

func (f *metadataClient) Get() (*undistrov1.Metadata, error) {
	log := logf.Log

	// gets the metadata file from the repository
	version := f.version
	name := "metadata.yaml"

	file, err := getLocalOverride(&newOverrideInput{
		configVariablesClient: f.configVarClient,
		provider:              f.provider,
		version:               version,
		filePath:              name,
	})
	if err != nil {
		return nil, err
	}
	if file == nil {
		log.V(5).Info("Fetching", "File", name, "Provider", f.provider.ManifestLabel(), "Version", version)
		file, err = f.repository.GetFile(version, name)
		if err != nil {
			if obj := f.getEmbeddedMetadata(); obj != nil {
				return obj, nil
			}
			return nil, errors.Wrapf(err, "failed to read %q from the repository for provider %q", name, f.provider.ManifestLabel())
		}
		objs, err := yaml.ToUnstructured(file)
		if err != nil {
			return nil, err
		}
		for _, o := range objs {
			if !strings.Contains(o.GetAPIVersion(), "getupcloud.com") {
				o.SetAPIVersion(undistrov1.GroupVersion.String())
				file, err = apiyaml.Marshal(o.Object)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		log.V(1).Info("Using", "Override", name, "Provider", f.provider.ManifestLabel(), "Version", version)
	}

	// Convert the yaml into a typed object
	obj := &undistrov1.Metadata{}
	codecFactory := serializer.NewCodecFactory(scheme.Scheme)

	if err := runtime.DecodeInto(codecFactory.UniversalDecoder(), file, obj); err != nil {
		return nil, errors.Wrapf(err, "error decoding %q for provider %q", name, f.provider.ManifestLabel())
	}

	//TODO: consider if to add metadata validation (TBD)

	return obj, nil
}

func (f *metadataClient) getEmbeddedMetadata() *undistrov1.Metadata {
	// undistro includes hard-coded metadata for cluster-API providers developed as a SIG-cluster-lifecycle project in order to
	// provide an option for simplifying the release process/the repository management of those projects.
	// Embedding metadata in undistro is optional, and the metadata.yaml file on the provider repository will always take precedence
	// on the embedded one.

	// if you are a developer of a SIG-cluster-lifecycle project, you can send a PR to extend the following list.
	switch f.provider.Type() {
	case undistrov1.UndistroProviderType:
		return &undistrov1.Metadata{
			TypeMeta: metav1.TypeMeta{
				APIVersion: undistrov1.GroupVersion.String(),
				Kind:       "Metadata",
			},
			ReleaseSeries: []undistrov1.ReleaseSeries{
				{Major: 0, Minor: 4, Contract: "v1alpha1"},
				{Major: 0, Minor: 5, Contract: "v1alpha1"},
				{Major: 0, Minor: 6, Contract: "v1alpha1"},
				{Major: 0, Minor: 7, Contract: "v1alpha1"},
				{Major: 0, Minor: 8, Contract: "v1alpha1"},
				{Major: 0, Minor: 9, Contract: "v1alpha1"},
				{Major: 0, Minor: 10, Contract: "v1alpha1"},
			},
		}
	case undistrov1.CoreProviderType:
		switch f.provider.Name() {
		case config.ClusterAPIProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 3, Contract: "v1alpha3"},
					// v1alpha2 release series are supported only for upgrades
					{Major: 0, Minor: 2, Contract: "v1alpha2"},
					// older version are not supported by undistro
				},
			}
		default:
			return nil
		}
	case undistrov1.BootstrapProviderType:
		switch f.provider.Name() {
		case config.KubeadmBootstrapProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 3, Contract: "v1alpha3"}, // From this release series CABPK version scheme is linked to CAPI; The 0.2 release series was skipped when doing this change.
					// v1alpha2 release series are supported only for upgrades
					{Major: 0, Minor: 1, Contract: "v1alpha2"}, // This release was hosted on a different repository
					// older version are not supported by undistro
				},
			}
		case config.TalosBootstrapProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 2, Contract: "v1alpha3"},
					// v1alpha2 release series are supported only for upgrades
					{Major: 0, Minor: 1, Contract: "v1alpha2"},
					// older version are not supported by undistro
				},
			}
		case config.EKSBootstrapProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 6, Contract: "v1alpha3"},
				},
			}
		default:
			return nil
		}
	case undistrov1.ControlPlaneProviderType:
		switch f.provider.Name() {
		case config.KubeadmControlPlaneProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 3, Contract: "v1alpha3"}, // KCP version scheme is linked to CAPI.
					// there are no older version for KCP
				},
			}
		case config.TalosControlPlaneProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 1, Contract: "v1alpha3"},
					// there are no older version for Talos controlplane
				},
			}
		default:
			return nil
		}
	case undistrov1.InfrastructureProviderType:
		switch f.provider.Name() {
		case config.AWSProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 6, Contract: "v1alpha3"},

					{Major: 0, Minor: 5, Contract: "v1alpha3"},
					// v1alpha2 release series are supported only for upgrades
					{Major: 0, Minor: 4, Contract: "v1alpha2"},
					// older version are not supported by undistro
				},
			}
		case config.AzureProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 4, Contract: "v1alpha3"},
					// v1alpha2 release series are supported only for upgrades
					{Major: 0, Minor: 3, Contract: "v1alpha2"},
					// older version are not supported by undistro
				},
			}
		case config.Metal3ProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 3, Contract: "v1alpha3"},
					// v1alpha2 release series are supported only for upgrades
					{Major: 0, Minor: 2, Contract: "v1alpha2"},
					// older version are not supported by undistro
				},
			}
		case config.PacketProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 3, Contract: "v1alpha3"},
					// older version are not supported by undistro
				},
			}
		case config.OpenStackProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 3, Contract: "v1alpha3"},
				},
			}
		case config.VSphereProviderName:
			return &undistrov1.Metadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "Metadata",
				},
				ReleaseSeries: []undistrov1.ReleaseSeries{
					// v1alpha3 release series
					{Major: 0, Minor: 7, Contract: "v1alpha3"},
					{Major: 0, Minor: 6, Contract: "v1alpha3"},
					// v1alpha2 release series are supported only for upgrades
					{Major: 0, Minor: 5, Contract: "v1alpha2"},
					// older version are not supported by undistro
				},
			}
		default:
			return nil
		}
	default:
		return nil
	}
}
