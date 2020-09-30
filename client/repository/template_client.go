/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"github.com/getupcloud/undistro/client/config"
	yaml "github.com/getupcloud/undistro/client/yamlprocessor"
	logf "github.com/getupcloud/undistro/log"
	"github.com/pkg/errors"
)

// TemplateClient has methods to work with cluster templates hosted on a provider repository.
// Templates are yaml files to be used for creating a guest cluster.
type TemplateClient interface {
	Get(flavor, targetNamespace string, listVariablesOnly bool) (Template, error)
}

// templateClient implements TemplateClient.
type templateClient struct {
	provider              config.Provider
	version               string
	repository            Repository
	configVariablesClient config.VariablesClient
	processor             yaml.Processor
}

type TemplateClientInput struct {
	version               string
	provider              config.Provider
	repository            Repository
	configVariablesClient config.VariablesClient
	processor             yaml.Processor
}

// Ensure templateClient implements the TemplateClient interface.
var _ TemplateClient = &templateClient{}

// newTemplateClient returns a templateClient. It uses the SimpleYamlProcessor
// by default
func newTemplateClient(input TemplateClientInput) *templateClient {
	return &templateClient{
		provider:              input.provider,
		version:               input.version,
		repository:            input.repository,
		configVariablesClient: input.configVariablesClient,
		processor:             input.processor,
	}
}

// Get return the template for the flavor specified.
// In case the template does not exists, an error is returned.
// Get assumes the following naming convention for templates: cluster-template[-<flavor_name>].yaml
func (c *templateClient) Get(flavor, targetNamespace string, listVariablesOnly bool) (Template, error) {
	log := logf.Log
	if targetNamespace == "" {
		return nil, errors.New("invalid arguments: please provide a targetNamespace")
	}

	version := c.version
	name := c.processor.GetTemplateName(version, flavor)

	// read the component YAML, reading the local override file if it exists, otherwise read from the provider repository
	rawArtifact, err := getLocalOverride(&newOverrideInput{
		configVariablesClient: c.configVariablesClient,
		provider:              c.provider,
		version:               version,
		filePath:              name,
	})
	if err != nil {
		return nil, err
	}

	if rawArtifact == nil {
		log.V(5).Info("Fetching", "File", name, "Provider", c.provider.ManifestLabel(), "Version", version)
		rawArtifact, err = c.repository.GetFile(version, name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read %q from provider's repository %q", name, c.provider.ManifestLabel())
		}
	} else {
		log.V(1).Info("Using", "Override", name, "Provider", c.provider.ManifestLabel(), "Version", version)
	}

	return NewTemplate(TemplateInput{rawArtifact, c.configVariablesClient, c.processor, targetNamespace, listVariablesOnly})
}
