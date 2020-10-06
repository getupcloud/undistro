/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/getupio-undistro/undistro/client/config"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/homedir"
)

const (
	overrideFolder    = "overrides"
	overrideFolderKey = "overridesFolder"
)

// Overrider provides behavior to determine the overrides layer.
type Overrider interface {
	Path() string
}

// overrides implements the Overrider interface.
type overrides struct {
	configVariablesClient config.VariablesClient
	providerLabel         string
	version               string
	filePath              string
}

type newOverrideInput struct {
	configVariablesClient config.VariablesClient
	provider              config.Provider
	version               string
	filePath              string
}

// newOverride returns an Overrider.
func newOverride(o *newOverrideInput) Overrider {
	return &overrides{
		configVariablesClient: o.configVariablesClient,
		providerLabel:         o.provider.ManifestLabel(),
		version:               o.version,
		filePath:              o.filePath,
	}
}

// Path returns the fully formed path to the file within the specified
// overrides config.
func (o *overrides) Path() string {
	basepath := filepath.Join(homedir.HomeDir(), config.ConfigFolder, overrideFolder)
	f, err := o.configVariablesClient.Get(overrideFolderKey)
	if err == nil && len(strings.TrimSpace(f)) != 0 {
		basepath = f
	}

	return filepath.Join(
		basepath,
		o.providerLabel,
		o.version,
		o.filePath,
	)
}

// getLocalOverride return local override file from the config folder, if it exists.
// This is required for development purposes, but it can be used also in production as a workaround for problems on the official repositories
func getLocalOverride(info *newOverrideInput) ([]byte, error) {
	overridePath := newOverride(info).Path()
	// it the local override exists, use it
	_, err := os.Stat(overridePath)
	if err == nil {
		content, err := ioutil.ReadFile(overridePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read local override for %s", overridePath)
		}
		return content, nil
	}

	// it the local override does not exists, return (so files from the provider's repository could be used)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// blocks for any other error
	return nil, err
}
