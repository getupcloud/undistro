/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/client/cluster"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Test_undistroClient_Delete(t *testing.T) {
	type fields struct {
		client *fakeClient
	}
	type args struct {
		options DeleteOptions
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantProviders sets.String
		wantErr       bool
	}{
		{
			name: "Delete all the providers",
			fields: fields{
				client: fakeClusterForDelete(),
			},
			args: args{
				options: DeleteOptions{
					Kubeconfig:              Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					IncludeNamespace:        false,
					IncludeCRDs:             false,
					Namespace:               "",
					CoreProvider:            "",
					BootstrapProviders:      nil,
					InfrastructureProviders: nil,
					ControlPlaneProviders:   nil,
					DeleteAll:               true, // delete all the providers
				},
			},
			wantProviders: sets.NewString(),
			wantErr:       false,
		},
		{
			name: "Delete single provider",
			fields: fields{
				client: fakeClusterForDelete(),
			},
			args: args{
				options: DeleteOptions{
					Kubeconfig:              Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					IncludeNamespace:        false,
					IncludeCRDs:             false,
					Namespace:               "capbpk-system",
					CoreProvider:            "",
					BootstrapProviders:      []string{bootstrapProviderConfig.Name()},
					InfrastructureProviders: nil,
					ControlPlaneProviders:   nil,
					DeleteAll:               false,
				},
			},
			wantProviders: sets.NewString(capiProviderConfig.Name()),
			wantErr:       false,
		},
		{
			name: "Delete single provider auto-detect namespace",
			fields: fields{
				client: fakeClusterForDelete(),
			},
			args: args{
				options: DeleteOptions{
					Kubeconfig:              Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					IncludeNamespace:        false,
					IncludeCRDs:             false,
					Namespace:               "", // empty namespace triggers namespace auto detection
					CoreProvider:            "",
					BootstrapProviders:      []string{bootstrapProviderConfig.Name()},
					InfrastructureProviders: nil,
					ControlPlaneProviders:   nil,
					DeleteAll:               false,
				},
			},
			wantProviders: sets.NewString(capiProviderConfig.Name()),
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := tt.fields.client.Delete(tt.args.options)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			input := cluster.Kubeconfig(tt.args.options.Kubeconfig)
			proxy := tt.fields.client.clusters[input].Proxy()
			gotProviders := &undistrov1.ProviderList{}

			c, err := proxy.NewClient()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(c.List(context.Background(), gotProviders)).To(Succeed())

			gotProvidersSet := sets.NewString()
			for _, gotProvider := range gotProviders.Items {
				gotProvidersSet.Insert(gotProvider.Name)
			}

			g.Expect(gotProvidersSet).To(Equal(tt.wantProviders))
		})
	}
}

// undistro client for a management cluster with capi and bootstrap provider
func fakeClusterForDelete() *fakeClient {
	config1 := newFakeConfig().
		WithVar("var", "value").
		WithProvider(capiProviderConfig).
		WithProvider(bootstrapProviderConfig)

	repository1 := newFakeRepository(capiProviderConfig, config1).
		WithPaths("root", "components.yaml").
		WithDefaultVersion("v1.0.0").
		WithFile("v1.0.0", "components.yaml", componentsYAML("ns1")).
		WithFile("v1.1.0", "components.yaml", componentsYAML("ns1"))
	repository2 := newFakeRepository(bootstrapProviderConfig, config1).
		WithPaths("root", "components.yaml").
		WithDefaultVersion("v2.0.0").
		WithFile("v2.0.0", "components.yaml", componentsYAML("ns2")).
		WithFile("v2.1.0", "components.yaml", componentsYAML("ns2"))

	cluster1 := newFakeCluster(cluster.Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"}, config1)
	cluster1.fakeProxy.WithProviderInventory(capiProviderConfig.Name(), capiProviderConfig.Type(), "v1.0.0", "capi-system", "")
	cluster1.fakeProxy.WithProviderInventory(bootstrapProviderConfig.Name(), bootstrapProviderConfig.Type(), "v1.0.0", "capbpk-system", "")

	client := newFakeClient(config1).
		// fake repository for capi, bootstrap and infra provider (matching provider's config)
		WithRepository(repository1).
		WithRepository(repository2).
		// fake empty cluster
		WithCluster(cluster1)

	return client
}
