/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"os"
	"path/filepath"
	"testing"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client/config"
	"github.com/getupcloud/undistro/internal/test"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/util/homedir"
)

func TestOverrides(t *testing.T) {
	tests := []struct {
		name            string
		configVarClient config.VariablesClient
		expectedPath    string
	}{
		{
			name:            "returns default overrides path if no config provided",
			configVarClient: test.NewFakeVariableClient(),
			expectedPath:    filepath.Join(homedir.HomeDir(), config.ConfigFolder, overrideFolder, "infrastructure-myinfra", "v1.0.1", "infra-comp.yaml"),
		},
		{
			name:            "returns default overrides path if config variable is empty",
			configVarClient: test.NewFakeVariableClient().WithVar(overrideFolderKey, ""),
			expectedPath:    filepath.Join(homedir.HomeDir(), config.ConfigFolder, overrideFolder, "infrastructure-myinfra", "v1.0.1", "infra-comp.yaml"),
		},
		{
			name:            "returns default overrides path if config variable is whitespace",
			configVarClient: test.NewFakeVariableClient().WithVar(overrideFolderKey, "   "),
			expectedPath:    filepath.Join(homedir.HomeDir(), config.ConfigFolder, overrideFolder, "infrastructure-myinfra", "v1.0.1", "infra-comp.yaml"),
		},
		{
			name:            "uses overrides folder from the config variables",
			configVarClient: test.NewFakeVariableClient().WithVar(overrideFolderKey, "/Users/foobar/workspace/releases"),
			expectedPath:    "/Users/foobar/workspace/releases/infrastructure-myinfra/v1.0.1/infra-comp.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			provider := config.NewProvider("myinfra", "", undistrov1.InfrastructureProviderType, nil, nil)
			override := newOverride(&newOverrideInput{
				configVariablesClient: tt.configVarClient,
				provider:              provider,
				version:               "v1.0.1",
				filePath:              "infra-comp.yaml",
			})

			g.Expect(override.Path()).To(Equal(tt.expectedPath))
		})
	}
}

func TestGetLocalOverrides(t *testing.T) {
	t.Run("returns contents of file successfully", func(t *testing.T) {
		g := NewWithT(t)
		tmpDir := createTempDir(t)
		defer os.RemoveAll(tmpDir)

		createLocalTestProviderFile(t, tmpDir, "infrastructure-myinfra/v1.0.1/infra-comp.yaml", "foo: bar")

		info := &newOverrideInput{
			configVariablesClient: test.NewFakeVariableClient().WithVar(overrideFolderKey, tmpDir),
			provider:              config.NewProvider("myinfra", "", undistrov1.InfrastructureProviderType, nil, nil),
			version:               "v1.0.1",
			filePath:              "infra-comp.yaml",
		}

		b, err := getLocalOverride(info)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(string(b)).To(Equal("foo: bar"))
	})

	t.Run("doesn't return error if file does not exist", func(t *testing.T) {
		g := NewWithT(t)

		info := &newOverrideInput{
			configVariablesClient: test.NewFakeVariableClient().WithVar(overrideFolderKey, "do-not-exist"),
			provider:              config.NewProvider("myinfra", "", undistrov1.InfrastructureProviderType, nil, nil),
			version:               "v1.0.1",
			filePath:              "infra-comp.yaml",
		}

		_, err := getLocalOverride(info)
		g.Expect(err).ToNot(HaveOccurred())
	})
}
