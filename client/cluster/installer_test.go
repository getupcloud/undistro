/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cluster

import (
	"testing"

	. "github.com/onsi/gomega"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client/config"
	"github.com/getupcloud/undistro/client/repository"
	"github.com/getupcloud/undistro/internal/test"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_providerInstaller_Validate(t *testing.T) {
	fakeReader := test.NewFakeReader().
		WithProvider("cluster-api", undistrov1.CoreProviderType, "https://somewhere.com").
		WithProvider("infra1", undistrov1.InfrastructureProviderType, "https://somewhere.com").
		WithProvider("infra2", undistrov1.InfrastructureProviderType, "https://somewhere.com")

	repositoryMap := map[string]repository.Repository{
		"cluster-api": test.NewFakeRepository().
			WithVersions("v1.0.0", "v1.0.1").
			WithMetadata("v1.0.0", &undistrov1.Metadata{
				ReleaseSeries: []undistrov1.ReleaseSeries{
					{Major: 1, Minor: 0, Contract: "v1alpha3"},
				},
			}).
			WithMetadata("v2.0.0", &undistrov1.Metadata{
				ReleaseSeries: []undistrov1.ReleaseSeries{
					{Major: 1, Minor: 0, Contract: "v1alpha3"},
					{Major: 2, Minor: 0, Contract: "v1alpha4"},
				},
			}),
		"infrastructure-infra1": test.NewFakeRepository().
			WithVersions("v1.0.0", "v1.0.1").
			WithMetadata("v1.0.0", &undistrov1.Metadata{
				ReleaseSeries: []undistrov1.ReleaseSeries{
					{Major: 1, Minor: 0, Contract: "v1alpha3"},
				},
			}).
			WithMetadata("v2.0.0", &undistrov1.Metadata{
				ReleaseSeries: []undistrov1.ReleaseSeries{
					{Major: 1, Minor: 0, Contract: "v1alpha3"},
					{Major: 2, Minor: 0, Contract: "v1alpha4"},
				},
			}),
		"infrastructure-infra2": test.NewFakeRepository().
			WithVersions("v1.0.0", "v1.0.1").
			WithMetadata("v1.0.0", &undistrov1.Metadata{
				ReleaseSeries: []undistrov1.ReleaseSeries{
					{Major: 1, Minor: 0, Contract: "v1alpha3"},
				},
			}).
			WithMetadata("v2.0.0", &undistrov1.Metadata{
				ReleaseSeries: []undistrov1.ReleaseSeries{
					{Major: 1, Minor: 0, Contract: "v1alpha3"},
					{Major: 2, Minor: 0, Contract: "v1alpha4"},
				},
			}),
	}

	type fields struct {
		proxy        Proxy
		installQueue []repository.Components
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "install core + infra1 on an empty cluster",
			fields: fields{
				proxy: test.NewFakeProxy(), //empty cluster
				installQueue: []repository.Components{ // install core + infra1, v1alpha3 contract
					newFakeComponents("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", ""),
					newFakeComponents("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "infra1-system", ""),
				},
			},
			wantErr: false,
		},
		{
			name: "install infra2 on a cluster already initialized with core + infra1",
			fields: fields{
				proxy: test.NewFakeProxy(). // cluster with core + infra1, v1alpha3 contract
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", "").
								WithProviderInventory("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "infra1-system", ""),
				installQueue: []repository.Components{ // install infra2, v1alpha3 contract
					newFakeComponents("infra2", undistrov1.InfrastructureProviderType, "v1.0.0", "infra2-system", ""),
				},
			},
			wantErr: false,
		},
		{
			name: "install another instance of infra1 on a cluster already initialized with core + infra1, no overlaps",
			fields: fields{
				proxy: test.NewFakeProxy(). // cluster with core + infra1, v1alpha3 contract
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", "").
								WithProviderInventory("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "ns1", "ns1"),
				installQueue: []repository.Components{ // install infra2, v1alpha3 contract
					newFakeComponents("infra2", undistrov1.InfrastructureProviderType, "v1.0.0", "ns2", "ns2"),
				},
			},
			wantErr: false,
		},
		{
			name: "install another instance of infra1 on a cluster already initialized with core + infra1, same namespace of the existing infra1",
			fields: fields{
				proxy: test.NewFakeProxy(). // cluster with core + infra1, v1alpha3 contract
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", "").
								WithProviderInventory("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "n1", ""),
				installQueue: []repository.Components{ // install infra1, v1alpha3 contract
					newFakeComponents("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "n1", ""),
				},
			},
			wantErr: true,
		},
		{
			name: "install another instance of infra1 on a cluster already initialized with core + infra1, watching overlap with the existing infra1",
			fields: fields{
				proxy: test.NewFakeProxy(). // cluster with core + infra1, v1alpha3 contract
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", "").
								WithProviderInventory("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "infra1-system", ""),
				installQueue: []repository.Components{ // install infra1, v1alpha3 contract
					newFakeComponents("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "infra2-system", ""),
				},
			},
			wantErr: true,
		},
		{
			name: "install another instance of infra1 on a cluster already initialized with core + infra1, not part of the existing management group",
			fields: fields{
				proxy: test.NewFakeProxy(). // cluster with core + infra1, v1alpha3 contract
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "ns1", "ns1").
								WithProviderInventory("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "ns1", "ns1"),
				installQueue: []repository.Components{ // install infra1, v1alpha3 contract
					newFakeComponents("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "ns2", "ns2"),
				},
			},
			wantErr: true,
		},
		{
			name: "install an instance of infra1 on a cluster already initialized with two core, but it is part of two management group",
			fields: fields{
				proxy: test.NewFakeProxy(). // cluster with two core (two management groups)
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "ns1", "ns1").
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "ns2", "ns2"),
				installQueue: []repository.Components{ // install infra1, v1alpha3 contract
					newFakeComponents("infra1", undistrov1.InfrastructureProviderType, "v1.0.0", "infra1-system", ""),
				},
			},
			wantErr: true,
		},
		{
			name: "install core@v1alpha3 + infra1@v1alpha4 on an empty cluster",
			fields: fields{
				proxy: test.NewFakeProxy(), //empty cluster
				installQueue: []repository.Components{ // install core + infra1, v1alpha3 contract
					newFakeComponents("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", ""),
					newFakeComponents("infra1", undistrov1.InfrastructureProviderType, "v2.0.0", "infra1-system", ""),
				},
			},
			wantErr: true,
		},
		{
			name: "install infra1@v1alpha4 on a cluster already initialized with core@v1alpha3 +",
			fields: fields{
				proxy: test.NewFakeProxy(). // cluster with one core, v1alpha3 contract
								WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "ns1", "ns1"),
				installQueue: []repository.Components{ // install infra1, v1alpha4 contract
					newFakeComponents("infra1", undistrov1.InfrastructureProviderType, "v2.0.0", "infra1-system", ""),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			configClient, _ := config.New("", config.InjectReader(fakeReader))

			i := &providerInstaller{
				configClient:      configClient,
				proxy:             tt.fields.proxy,
				providerInventory: newInventoryClient(tt.fields.proxy, nil),
				repositoryClientFactory: func(provider config.Provider, configClient config.Client, options ...repository.Option) (repository.Client, error) {
					return repository.New(provider, configClient, repository.InjectRepository(repositoryMap[provider.ManifestLabel()]))
				},
				installQueue: tt.fields.installQueue,
			}

			err := i.Validate()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

type fakeComponents struct {
	config.Provider
	inventoryObject undistrov1.Provider
}

func (c *fakeComponents) Version() string {
	panic("not implemented")
}

func (c *fakeComponents) Variables() []string {
	panic("not implemented")
}

func (c *fakeComponents) Images() []string {
	panic("not implemented")
}

func (c *fakeComponents) TargetNamespace() string {
	panic("not implemented")
}

func (c *fakeComponents) WatchingNamespace() string {
	panic("not implemented")
}

func (c *fakeComponents) InventoryObject() undistrov1.Provider {
	return c.inventoryObject
}

func (c *fakeComponents) InstanceObjs() []unstructured.Unstructured {
	panic("not implemented")
}

func (c *fakeComponents) SharedObjs() []unstructured.Unstructured {
	panic("not implemented")
}

func (c *fakeComponents) Yaml() ([]byte, error) {
	panic("not implemented")
}

func newFakeComponents(name string, providerType undistrov1.ProviderType, version, targetNamespace, watchingNamespace string) repository.Components {
	inventoryObject := fakeProvider(name, providerType, version, targetNamespace, watchingNamespace)
	return &fakeComponents{
		Provider:        config.NewProvider(inventoryObject.ProviderName, "", undistrov1.ProviderType(inventoryObject.Type), nil, nil),
		inventoryObject: inventoryObject,
	}
}

func Test_shouldInstallSharedComponents(t *testing.T) {
	type args struct {
		providerList *undistrov1.ProviderList
		provider     undistrov1.Provider
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "First instance of the provider, must install shared components",
			args: args{
				providerList: &undistrov1.ProviderList{Items: []undistrov1.Provider{}}, // no core provider installed
				provider:     fakeProvider("core", undistrov1.CoreProviderType, "v2.0.0", "", ""),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Second instance of the provider, same version, must NOT install shared components",
			args: args{
				providerList: &undistrov1.ProviderList{Items: []undistrov1.Provider{
					fakeProvider("core", undistrov1.CoreProviderType, "v2.0.0", "", ""),
				}},
				provider: fakeProvider("core", undistrov1.CoreProviderType, "v2.0.0", "", ""),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Second instance of the provider, older version, must NOT install shared components",
			args: args{
				providerList: &undistrov1.ProviderList{Items: []undistrov1.Provider{
					fakeProvider("core", undistrov1.CoreProviderType, "v2.0.0", "", ""),
				}},
				provider: fakeProvider("core", undistrov1.CoreProviderType, "v1.0.0", "", ""),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Second instance of the provider, newer version, must install shared components",
			args: args{
				providerList: &undistrov1.ProviderList{Items: []undistrov1.Provider{
					fakeProvider("core", undistrov1.CoreProviderType, "v2.0.0", "", ""),
				}},
				provider: fakeProvider("core", undistrov1.CoreProviderType, "v3.0.0", "", ""),
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got, err := shouldInstallSharedComponents(tt.args.providerList, tt.args.provider)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(got).To(Equal(tt.want))
		})
	}
}
