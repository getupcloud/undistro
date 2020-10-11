/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cluster

import (
	"testing"

	. "github.com/onsi/gomega"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/internal/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func Test_inventoryClient_GetManagementGroups(t *testing.T) {
	type fields struct {
		proxy Proxy
	}
	tests := []struct {
		name    string
		fields  fields
		want    ManagementGroupList
		wantErr bool
	}{
		{
			name: "Simple management cluster",
			fields: fields{ // 1 instance for each provider, watching all namespace
				proxy: test.NewFakeProxy().
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", "").
					WithProviderInventory("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system", "").
					WithProviderInventory("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system", ""),
			},
			want: ManagementGroupList{ // One Group
				{
					CoreProvider: fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", ""),
					Providers: []undistrov1.Provider{
						fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", ""),
						fakeProvider("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system", ""),
						fakeProvider("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system", ""),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "1 Core, many infra (1 ManagementGroup)",
			fields: fields{ // 1 instance of core and bootstrap provider, watching all namespace; more instances of infrastructure providers, each watching dedicated ns
				proxy: test.NewFakeProxy().
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", "").
					WithProviderInventory("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system", "").
					WithProviderInventory("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system1", "ns1").
					WithProviderInventory("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system2", "ns2"),
			},
			want: ManagementGroupList{ // One Group
				{
					CoreProvider: fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", ""),
					Providers: []undistrov1.Provider{
						fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system", ""),
						fakeProvider("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system", ""),
						fakeProvider("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system1", "ns1"),
						fakeProvider("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system2", "ns2"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "two ManagementGroups",
			fields: fields{ // more instances of core with related bootstrap, infrastructure
				proxy: test.NewFakeProxy().
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system1", "ns1").
					WithProviderInventory("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system1", "ns1").
					WithProviderInventory("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system1", "ns1").
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system2", "ns2").
					WithProviderInventory("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system2", "ns2").
					WithProviderInventory("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system2", "ns2"),
			},
			want: ManagementGroupList{ // Two Groups
				{
					CoreProvider: fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system1", "ns1"),
					Providers: []undistrov1.Provider{
						fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system1", "ns1"),
						fakeProvider("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system1", "ns1"),
						fakeProvider("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system1", "ns1"),
					},
				},
				{
					CoreProvider: fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system2", "ns2"),
					Providers: []undistrov1.Provider{
						fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system2", "ns2"),
						fakeProvider("bootstrap", undistrov1.BootstrapProviderType, "v1.0.0", "bootstrap-system2", "ns2"),
						fakeProvider("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system2", "ns2"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "fails with overlapping core providers",
			fields: fields{ //two core providers watching for the same namespaces
				proxy: test.NewFakeProxy().
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system1", "").
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system2", ""),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fails with overlapping core providers",
			fields: fields{ //a provider watching for objects controlled by more than one core provider
				proxy: test.NewFakeProxy().
					WithProviderInventory("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system", "").
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system1", "ns1").
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system2", "ns2"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "fails with orphan providers",
			fields: fields{ //a provider watching for objects not controlled any core provider
				proxy: test.NewFakeProxy().
					WithProviderInventory("infrastructure", undistrov1.InfrastructureProviderType, "v1.0.0", "infra-system", "ns1").
					WithProviderInventory("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system1", "ns2"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p := &inventoryClient{
				proxy: tt.fields.proxy,
			}
			got, err := p.GetManagementGroups()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(len(got)).To(Equal(len(tt.want)))
		})
	}
}

func fakeProvider(name string, providerType undistrov1.ProviderType, version, targetNamespace, watchingNamespace string) undistrov1.Provider {
	return undistrov1.Provider{
		TypeMeta: metav1.TypeMeta{
			APIVersion: undistrov1.GroupVersion.String(),
			Kind:       "Provider",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: targetNamespace,
			Name:      undistrov1.ManifestLabel(name, providerType),
			Labels: map[string]string{
				undistrov1.ClusterctlLabelName:     "",
				undistrov1.UndistroLabelName:       "",
				clusterv1.ProviderLabelName:        undistrov1.ManifestLabel(name, providerType),
				undistrov1.ClusterctlCoreLabelName: "inventory",
				undistrov1.UndistroCoreLabelName:   "inventory",
			},
		},
		ProviderName:     name,
		Type:             string(providerType),
		Version:          version,
		WatchedNamespace: watchingNamespace,
	}
}
