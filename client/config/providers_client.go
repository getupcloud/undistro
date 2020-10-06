/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	gotemplate "text/template"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/internal/template"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/yaml"
)

const (
	// Core providers
	ClusterAPIProviderName = "cluster-api"
	UndistroProviderName   = "undistro"

	// Infra providers
	AWSProviderName       = "aws"
	AzureProviderName     = "azure"
	Metal3ProviderName    = "metal3"
	OpenStackProviderName = "openstack"
	PacketProviderName    = "packet"
	VSphereProviderName   = "vsphere"

	// Bootstrap providers
	KubeadmBootstrapProviderName = "kubeadm"
	TalosBootstrapProviderName   = "talos"
	EKSBootstrapProviderName     = "eks"

	// ControlPlane providers
	KubeadmControlPlaneProviderName = "kubeadm"
	TalosControlPlaneProviderName   = "talos"
	EKSControlPlaneProviderName     = "eks"

	// Other
	ProvidersConfigKey = "providers"
)

// ProvidersClient has methods to work with provider configurations.
type ProvidersClient interface {
	// List returns all the provider configurations, including provider configurations hard-coded in undistro
	// and user-defined provider configurations read from the undistro configuration file.
	// In case of conflict, user-defined provider override the hard-coded configurations.
	List() ([]Provider, error)

	// Get returns the configuration for the provider with a given name/type.
	// In case the name/type does not correspond to any existing provider, an error is returned.
	Get(name string, providerType undistrov1.ProviderType) (Provider, error)
}

// providersClient implements ProvidersClient.
type providersClient struct {
	reader Reader
}

// ensure providersClient implements ProvidersClient.
var _ ProvidersClient = &providersClient{}

func newProvidersClient(reader Reader) *providersClient {
	return &providersClient{
		reader: reader,
	}
}

func (p *providersClient) defaults() []Provider {
	// undistro includes a predefined list of Cluster API providers sponsored by SIG-cluster-lifecycle to provide users the simplest
	// out-of-box experience. This is an opt-in feature; other providers can be added by using the undistro configuration file.

	defaults := []Provider{
		// cluster API core provider
		&provider{
			name:         ClusterAPIProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api/releases/latest/core-components.yaml",
			providerType: undistrov1.CoreProviderType,
		},

		&provider{
			name:         UndistroProviderName,
			url:          "https://github.com/getupio-undistro/undistro/releases/latest/core-components.yaml",
			providerType: undistrov1.UndistroProviderType,
		},

		// Infrastructure providers
		&provider{
			name:          AWSProviderName,
			url:           "https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases/latest/infrastructure-components.yaml",
			providerType:  undistrov1.InfrastructureProviderType,
			preConfigFunc: awsPreConfig,
			initFunc:      awsInit,
		},
		&provider{
			name:         AzureProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api-provider-azure/releases/latest/infrastructure-components.yaml",
			providerType: undistrov1.InfrastructureProviderType,
		},
		&provider{
			name:         PacketProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api-provider-packet/releases/latest/infrastructure-components.yaml",
			providerType: undistrov1.InfrastructureProviderType,
		},
		&provider{
			name:         Metal3ProviderName,
			url:          "https://github.com/metal3-io/cluster-api-provider-metal3/releases/latest/infrastructure-components.yaml",
			providerType: undistrov1.InfrastructureProviderType,
		},
		&provider{
			name:         OpenStackProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api-provider-openstack/releases/latest/infrastructure-components.yaml",
			providerType: undistrov1.InfrastructureProviderType,
		},
		&provider{
			name:         VSphereProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/releases/latest/infrastructure-components.yaml",
			providerType: undistrov1.InfrastructureProviderType,
		},

		// Bootstrap providers
		&provider{
			name:         KubeadmBootstrapProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api/releases/latest/bootstrap-components.yaml",
			providerType: undistrov1.BootstrapProviderType,
		},
		&provider{
			name:         TalosBootstrapProviderName,
			url:          "https://github.com/talos-systems/cluster-api-bootstrap-provider-talos/releases/latest/bootstrap-components.yaml",
			providerType: undistrov1.BootstrapProviderType,
		},
		&provider{
			name:         EKSBootstrapProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases/latest/eks-bootstrap-components.yaml",
			providerType: undistrov1.BootstrapProviderType,
		},

		// ControlPlane providers
		&provider{
			name:         KubeadmControlPlaneProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api/releases/latest/control-plane-components.yaml",
			providerType: undistrov1.ControlPlaneProviderType,
		},
		&provider{
			name:         TalosControlPlaneProviderName,
			url:          "https://github.com/talos-systems/cluster-api-control-plane-provider-talos/releases/latest/control-plane-components.yaml",
			providerType: undistrov1.ControlPlaneProviderType,
		},
		&provider{
			name:         EKSControlPlaneProviderName,
			url:          "https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases/latest/eks-controlplane-components.yaml",
			providerType: undistrov1.ControlPlaneProviderType,
		},
	}

	return defaults
}

// configProvider mirrors config.Provider interface and allows serialization of the corresponding info
type configProvider struct {
	Name string                  `json:"name,omitempty"`
	URL  string                  `json:"url,omitempty"`
	Type undistrov1.ProviderType `json:"type,omitempty"`
}

func SetupTemplates(cl *undistrov1.Cluster, v VariablesClient) (template.Options, error) {
	variable := func(vr string) string {
		s, err := v.Get(vr)
		if err != nil {
			return ""
		}
		return s
	}
	opts := template.Options{
		Funcs: []gotemplate.FuncMap{{
			"variable": variable,
		}},
	}
	pvs := make([]configProvider, 0)
	if cl.Spec.InfrastructureProvider.File != nil {
		pvs = append(pvs, configProvider{
			Name: cl.Spec.InfrastructureProvider.Name,
			URL:  *cl.Spec.InfrastructureProvider.File,
			Type: undistrov1.InfrastructureProviderType,
		})
	}
	if cl.Spec.BootstrapProvider != nil {
		if cl.Spec.BootstrapProvider.File != nil {
			pvs = append(pvs, configProvider{
				Name: cl.Spec.BootstrapProvider.Name,
				URL:  *cl.Spec.BootstrapProvider.File,
				Type: undistrov1.BootstrapProviderType,
			})
		}
	}
	if cl.Spec.ControlPlaneProvider != nil {
		if cl.Spec.ControlPlaneProvider.File != nil {
			pvs = append(pvs, configProvider{
				Name: cl.Spec.ControlPlaneProvider.Name,
				URL:  *cl.Spec.ControlPlaneProvider.File,
				Type: undistrov1.ControlPlaneProviderType,
			})
		}
	}
	if len(pvs) > 0 {
		byt, err := yaml.Marshal(pvs)
		if err != nil {
			return opts, err
		}
		v.Set(ProvidersConfigKey, string(byt))
	}
	if cl.Spec.Template != nil {
		res, err := http.Get(*cl.Spec.Template)
		if err != nil {
			return opts, err
		}
		if res.StatusCode != http.StatusOK {
			return opts, errors.Errorf("bad response code sended by %s: %d", *cl.Spec.Template, res.StatusCode)
		}
		byt, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return opts, err
		}
		defer res.Body.Close()
		name := fmt.Sprintf("clustertemplates/%s.yaml", cl.Spec.InfrastructureProvider.Name)
		opts.Asset = func(file string) ([]byte, error) {
			switch file {
			case name:
				return byt, nil
			default:
				return nil, errors.Errorf("file not found: %s", file)
			}
		}
		opts.AssetNames = func() []string {
			return []string{name}
		}
	}
	return opts, nil
}

func (p *providersClient) List() ([]Provider, error) {
	// Creates a maps with all the defaults provider configurations
	providers := p.defaults()

	// Gets user defined provider configurations, validate them, and merges with
	// hard-coded configurations handling conflicts (user defined take precedence on hard-coded)

	userDefinedProviders := []configProvider{}
	if err := p.reader.UnmarshalKey(ProvidersConfigKey, &userDefinedProviders); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal providers from the undistro configuration file")
	}

	for _, u := range userDefinedProviders {
		provider := NewProvider(u.Name, u.URL, u.Type, nil, nil)
		if err := validateProvider(provider); err != nil {
			return nil, errors.Wrapf(err, "error validating configuration for the %s with name %s. Please fix the providers value in undistro configuration file", provider.Type(), provider.Name())
		}

		override := false
		for i := range providers {
			if providers[i].SameAs(provider) {
				providers[i] = provider
				override = true
			}
		}

		if !override {
			providers = append(providers, provider)
		}
	}

	// ensure provider configurations are consistently sorted
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Less(providers[j])
	})

	return providers, nil
}

func (p *providersClient) Get(name string, providerType undistrov1.ProviderType) (Provider, error) {
	l, err := p.List()
	if err != nil {
		return nil, err
	}

	provider := NewProvider(name, "", providerType, nil, nil) //Nb. Having the url empty is fine because the url is not considered by SameAs.
	for _, r := range l {
		if r.SameAs(provider) {
			return r, nil
		}
	}

	return nil, errors.Errorf("failed to get configuration for the %s with name %s. Please check the provider name and/or add configuration for new providers using the .undistro config file", providerType, name)
}

func validateProvider(r Provider) error {
	if r.Name() == "" {
		return errors.New("name value cannot be empty")
	}

	if (r.Name() == ClusterAPIProviderName) != (r.Type() == undistrov1.CoreProviderType) {
		return errors.Errorf("name %s must be used with the %s type (name: %s, type: %s)", ClusterAPIProviderName, undistrov1.CoreProviderType, r.Name(), r.Type())
	}

	errMsgs := validation.IsDNS1123Subdomain(r.Name())
	if len(errMsgs) != 0 {
		return errors.Errorf("invalid provider name: %s", strings.Join(errMsgs, "; "))
	}
	if r.URL() == "" {
		return errors.New("provider URL value cannot be empty")
	}

	_, err := url.Parse(r.URL())
	if err != nil {
		return errors.Wrap(err, "error parsing provider URL")
	}

	switch r.Type() {
	case undistrov1.CoreProviderType,
		undistrov1.BootstrapProviderType,
		undistrov1.InfrastructureProviderType,
		undistrov1.ControlPlaneProviderType,
		undistrov1.UndistroProviderType:
		break
	default:
		return errors.Errorf("invalid provider type. Allowed values are [%s, %s, %s, %s, %s]",
			undistrov1.CoreProviderType,
			undistrov1.BootstrapProviderType,
			undistrov1.InfrastructureProviderType,
			undistrov1.ControlPlaneProviderType,
			undistrov1.UndistroProviderType)
	}
	return nil
}
