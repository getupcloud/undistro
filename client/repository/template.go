/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"github.com/getupio-undistro/undistro/client/config"
	yaml "github.com/getupio-undistro/undistro/client/yamlprocessor"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "sigs.k8s.io/cluster-api/util/yaml"
)

// Template wraps a YAML file that defines the cluster objects (Cluster, Machines etc.).
// It is important to notice that undistro applies a set of processing steps to the “raw” cluster template YAML read
// from the provider repositories:
// 1. Checks for all the variables in the cluster template YAML file and replace with corresponding config values
// 2. Ensure all the cluster objects are deployed in the target namespace
type Template interface {
	// Variables required by the template.
	// This value is derived by the template YAML.
	Variables() []string

	// TargetNamespace where the template objects will be installed.
	TargetNamespace() string

	// Yaml returns yaml defining all the cluster template objects as a byte array.
	Yaml() ([]byte, error)

	// Objs returns the cluster template as a list of Unstructured objects.
	Objs() []unstructured.Unstructured
}

// template implements Template.
type template struct {
	variables       []string
	targetNamespace string
	objs            []unstructured.Unstructured
}

// Ensures template implements the Template interface.
var _ Template = &template{}

func (t *template) Variables() []string {
	return t.variables
}

func (t *template) TargetNamespace() string {
	return t.targetNamespace
}

func (t *template) Objs() []unstructured.Unstructured {
	return t.objs
}

func (t *template) Yaml() ([]byte, error) {
	return utilyaml.FromUnstructured(t.objs)
}

type TemplateInput struct {
	RawArtifact           []byte
	ConfigVariablesClient config.VariablesClient
	Processor             yaml.Processor
	TargetNamespace       string
	ListVariablesOnly     bool
}

// NewTemplate returns a new objects embedding a cluster template YAML file.
func NewTemplate(input TemplateInput) (*template, error) {
	variables, err := input.Processor.GetVariables(input.RawArtifact)
	if err != nil {
		return nil, err
	}

	if input.ListVariablesOnly {
		return &template{
			variables:       variables,
			targetNamespace: input.TargetNamespace,
		}, nil
	}

	processedYaml, err := input.Processor.Process(input.RawArtifact, input.ConfigVariablesClient.Get)
	if err != nil {
		return nil, err
	}

	// Transform the yaml in a list of objects, so following transformation can work on typed objects (instead of working on a string/slice of bytes).
	objs, err := utilyaml.ToUnstructured(processedYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse yaml")
	}

	// Ensures all the template components are deployed in the target namespace (applies only to namespaced objects)
	// This is required in order to ensure a cluster and all the related objects are in a single namespace, that is a requirement for
	// the undistro move operation (and also for many controller reconciliation loops).
	objs = fixTargetNamespace(objs, input.TargetNamespace)

	return &template{
		variables:       variables,
		targetNamespace: input.TargetNamespace,
		objs:            objs,
	}, nil
}
