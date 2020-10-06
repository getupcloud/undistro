/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"strings"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/client/repository"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

// getComponentsByName is a utility method that returns components
// for a given provider with options including targetNamespace, and watchingNamespace.
func (c *undistroClient) getComponentsByName(provider string, providerType undistrov1.ProviderType, options repository.ComponentsOptions) (repository.Components, error) {

	// Parse the abbreviated syntax for name[:version]
	name, version, err := parseProviderName(provider)
	if err != nil {
		return nil, err
	}
	options.Version = version

	// Gets the provider configuration (that includes the location of the provider repository)
	providerConfig, err := c.configClient.Providers().Get(name, providerType)
	if err != nil {
		return nil, err
	}

	// Get a client for the provider repository and read the provider components;
	// during the process, provider components will be processed performing variable substitution, customization of target
	// and watching namespace etc.
	// Currently we are not supporting custom yaml processors for the provider
	// components. So we revert to using the default SimpleYamlProcessor.
	repositoryClientFactory, err := c.repositoryClientFactory(RepositoryClientFactoryInput{provider: providerConfig})
	if err != nil {
		return nil, err
	}

	components, err := repositoryClientFactory.Components().Get(options)
	if err != nil {
		return nil, err
	}
	return components, nil
}

// parseProviderName defines a utility function that parses the abbreviated syntax for name[:version]
func parseProviderName(provider string) (name string, version string, err error) {
	t := strings.Split(strings.ToLower(provider), ":")
	if len(t) > 2 {
		return "", "", errors.Errorf("invalid provider name %q. Provider name should be in the form name[:version]", provider)
	}

	if t[0] == "" {
		return "", "", errors.Errorf("invalid provider name %q. Provider name should be in the form name[:version] and name cannot be empty", provider)
	}

	name = t[0]
	if err := validateDNS1123Label(name); err != nil {
		return "", "", errors.Wrapf(err, "invalid provider name %q. Provider name should be in the form name[:version] and the name should be valid", provider)
	}

	version = ""
	if len(t) > 1 {
		if t[1] == "" {
			return "", "", errors.Errorf("invalid provider name %q. Provider name should be in the form name[:version] and version cannot be empty", provider)
		}
		version = t[1]
	}

	return name, version, nil
}

func validateDNS1123Label(label string) error {
	errs := validation.IsDNS1123Label(label)
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func validateDNS1123Domanin(subdomain string) error {
	errs := validation.IsDNS1123Subdomain(subdomain)
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
