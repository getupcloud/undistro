/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client/config"
	yaml "github.com/getupcloud/undistro/client/yamlprocessor"
	"github.com/getupcloud/undistro/internal/test"
)

func Test_newRepositoryClient_LocalFileSystemRepository(t *testing.T) {
	g := NewWithT(t)

	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	dst1 := createLocalTestProviderFile(t, tmpDir, "bootstrap-foo/v1.0.0/bootstrap-components.yaml", "")
	dst2 := createLocalTestProviderFile(t, tmpDir, "bootstrap-bar/v2.0.0/bootstrap-components.yaml", "")

	configClient, err := config.New("", config.InjectReader(test.NewFakeReader()))
	g.Expect(err).NotTo(HaveOccurred())

	type fields struct {
		provider config.Provider
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "successfully creates repository client with local filesystem backend and scheme == \"\"",
			fields: fields{
				provider: config.NewProvider("foo", dst1, undistrov1.BootstrapProviderType, nil, nil),
			},
		},
		{
			name: "successfully creates repository client with local filesystem backend and scheme == \"file\"",
			fields: fields{
				provider: config.NewProvider("bar", "file://"+dst2, undistrov1.BootstrapProviderType, nil, nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := NewWithT(t)

			repoClient, err := newRepositoryClient(tt.fields.provider, configClient)
			gs.Expect(err).NotTo(HaveOccurred())

			var expected *localRepository
			gs.Expect(repoClient.repository).To(BeAssignableToTypeOf(expected))
		})
	}
}

func Test_newRepositoryClient_YamlProcesor(t *testing.T) {
	tests := []struct {
		name   string
		opts   []Option
		assert func(*WithT, yaml.Processor)
	}{
		{
			name: "it creates a repository client with simple yaml processor by default",
			assert: func(g *WithT, p yaml.Processor) {
				_, ok := (p).(*yaml.SimpleProcessor)
				g.Expect(ok).To(BeTrue())
			},
		},
		{
			name: "it creates a repository client with specified yaml processor",
			opts: []Option{InjectYamlProcessor(test.NewFakeProcessor())},
			assert: func(g *WithT, p yaml.Processor) {
				_, ok := (p).(*yaml.SimpleProcessor)
				g.Expect(ok).To(BeFalse())
				_, ok = (p).(*test.FakeProcessor)
				g.Expect(ok).To(BeTrue())
			},
		},
		{
			name: "it creates a repository with simple yaml processor even if injected with nil processor",
			opts: []Option{InjectYamlProcessor(nil)},
			assert: func(g *WithT, p yaml.Processor) {
				g.Expect(p).ToNot(BeNil())
				_, ok := (p).(*yaml.SimpleProcessor)
				g.Expect(ok).To(BeTrue())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			configProvider := config.NewProvider("fakeProvider", "", undistrov1.CoreProviderType, nil, nil)
			configClient, err := config.New("", config.InjectReader(test.NewFakeReader()))
			g.Expect(err).NotTo(HaveOccurred())

			tt.opts = append(tt.opts, InjectRepository(test.NewFakeRepository()))

			repoClient, err := newRepositoryClient(
				configProvider,
				configClient,
				tt.opts...,
			)
			g.Expect(err).NotTo(HaveOccurred())
			tt.assert(g, repoClient.processor)
		})
	}
}
