/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"context"
	"sort"
	"testing"

	. "github.com/onsi/gomega"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client/cluster"
	"github.com/getupcloud/undistro/client/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

func Test_undistroClient_PlanUpgrade(t *testing.T) {
	type fields struct {
		client *fakeClient
	}
	type args struct {
		options PlanUpgradeOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "does not return error if cluster client is found",
			fields: fields{
				client: fakeClientForUpgrade(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: PlanUpgradeOptions{
					Kubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
				},
			},
			wantErr: false,
		},
		{
			name: "returns an error if cluster client is not found",
			fields: fields{
				client: fakeClientForUpgrade(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: PlanUpgradeOptions{
					Kubeconfig: Kubeconfig{Path: "kubeconfig", Context: "some-other-context"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			_, err := tt.fields.client.PlanUpgrade(tt.args.options)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
		})
	}
}

func Test_undistroClient_ApplyUpgrade(t *testing.T) {
	type fields struct {
		client *fakeClient
	}
	type args struct {
		options ApplyUpgradeOptions
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantProviders *undistrov1.ProviderList
		wantErr       bool
	}{
		{
			name: "apply a plan",
			fields: fields{
				client: fakeClientForUpgrade(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: ApplyUpgradeOptions{
					Kubeconfig:              Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ManagementGroup:         "cluster-api-system/cluster-api",
					Contract:                "v1alpha3",
					CoreProvider:            "",
					BootstrapProviders:      nil,
					ControlPlaneProviders:   nil,
					InfrastructureProviders: nil,
				},
			},
			wantProviders: &undistrov1.ProviderList{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "ProviderList",
				},
				ListMeta: metav1.ListMeta{},
				Items: []undistrov1.Provider{ // both providers should be upgraded
					fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.1", "cluster-api-system"),
					fakeProvider("infra", undistrov1.InfrastructureProviderType, "v2.0.1", "infra-system"),
				},
			},
			wantErr: false,
		},
		{
			name: "apply a custom plan - core provider only",
			fields: fields{
				client: fakeClientForUpgrade(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: ApplyUpgradeOptions{
					Kubeconfig:              Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ManagementGroup:         "cluster-api-system/cluster-api",
					Contract:                "",
					CoreProvider:            "cluster-api-system/cluster-api:v1.0.1",
					BootstrapProviders:      nil,
					ControlPlaneProviders:   nil,
					InfrastructureProviders: nil,
				},
			},
			wantProviders: &undistrov1.ProviderList{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "ProviderList",
				},
				ListMeta: metav1.ListMeta{},
				Items: []undistrov1.Provider{ // only one provider should be upgraded
					fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.1", "cluster-api-system"),
					fakeProvider("infra", undistrov1.InfrastructureProviderType, "v2.0.0", "infra-system"),
				},
			},
			wantErr: false,
		},
		{
			name: "apply a custom plan - infra provider only",
			fields: fields{
				client: fakeClientForUpgrade(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: ApplyUpgradeOptions{
					Kubeconfig:              Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ManagementGroup:         "cluster-api-system/cluster-api",
					Contract:                "",
					CoreProvider:            "",
					BootstrapProviders:      nil,
					ControlPlaneProviders:   nil,
					InfrastructureProviders: []string{"infra-system/infra:v2.0.1"},
				},
			},
			wantProviders: &undistrov1.ProviderList{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "ProviderList",
				},
				ListMeta: metav1.ListMeta{},
				Items: []undistrov1.Provider{ // only one provider should be upgraded
					fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.0", "cluster-api-system"),
					fakeProvider("infra", undistrov1.InfrastructureProviderType, "v2.0.1", "infra-system"),
				},
			},
			wantErr: false,
		},
		{
			name: "apply a custom plan - both providers",
			fields: fields{
				client: fakeClientForUpgrade(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: ApplyUpgradeOptions{
					Kubeconfig:              Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ManagementGroup:         "cluster-api-system/cluster-api",
					Contract:                "",
					CoreProvider:            "cluster-api-system/cluster-api:v1.0.1",
					BootstrapProviders:      nil,
					ControlPlaneProviders:   nil,
					InfrastructureProviders: []string{"infra-system/infra:v2.0.1"},
				},
			},
			wantProviders: &undistrov1.ProviderList{
				TypeMeta: metav1.TypeMeta{
					APIVersion: undistrov1.GroupVersion.String(),
					Kind:       "ProviderList",
				},
				ListMeta: metav1.ListMeta{},
				Items: []undistrov1.Provider{ // only one provider should be upgraded
					fakeProvider("cluster-api", undistrov1.CoreProviderType, "v1.0.1", "cluster-api-system"),
					fakeProvider("infra", undistrov1.InfrastructureProviderType, "v2.0.1", "infra-system"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := tt.fields.client.ApplyUpgrade(tt.args.options)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			// converting between client and cluster alias for Kubeconfig
			input := cluster.Kubeconfig(tt.args.options.Kubeconfig)
			proxy := tt.fields.client.clusters[input].Proxy()
			gotProviders := &undistrov1.ProviderList{}

			c, err := proxy.NewClient()
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(c.List(context.Background(), gotProviders)).To(Succeed())

			sort.Slice(gotProviders.Items, func(i, j int) bool {
				return gotProviders.Items[i].Name < gotProviders.Items[j].Name
			})
			sort.Slice(tt.wantProviders.Items, func(i, j int) bool {
				return tt.wantProviders.Items[i].Name < tt.wantProviders.Items[j].Name
			})
			for i := range gotProviders.Items {
				tt.wantProviders.Items[i].ResourceVersion = gotProviders.Items[i].ResourceVersion
			}
			g.Expect(gotProviders).To(Equal(tt.wantProviders))
		})
	}
}

func fakeClientForUpgrade() *fakeClient {
	core := config.NewProvider("cluster-api", "https://somewhere.com", undistrov1.CoreProviderType, nil)
	infra := config.NewProvider("infra", "https://somewhere.com", undistrov1.InfrastructureProviderType, nil)

	config1 := newFakeConfig().
		WithProvider(core).
		WithProvider(infra)

	repository1 := newFakeRepository(core, config1).
		WithPaths("root", "components.yaml").
		WithDefaultVersion("v1.0.1").
		WithFile("v1.0.1", "components.yaml", componentsYAML("ns2")).
		WithVersions("v1.0.0", "v1.0.1").
		WithMetadata("v1.0.1", &undistrov1.Metadata{
			ReleaseSeries: []undistrov1.ReleaseSeries{
				{Major: 1, Minor: 0, Contract: "v1alpha3"},
			},
		})
	repository2 := newFakeRepository(infra, config1).
		WithPaths("root", "components.yaml").
		WithDefaultVersion("v2.0.0").
		WithFile("v2.0.1", "components.yaml", componentsYAML("ns2")).
		WithVersions("v2.0.0", "v2.0.1").
		WithMetadata("v2.0.1", &undistrov1.Metadata{
			ReleaseSeries: []undistrov1.ReleaseSeries{
				{Major: 2, Minor: 0, Contract: "v1alpha3"},
			},
		})

	cluster1 := newFakeCluster(cluster.Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"}, config1).
		WithRepository(repository1).
		WithRepository(repository2).
		WithProviderInventory(core.Name(), core.Type(), "v1.0.0", "cluster-api-system", "").
		WithProviderInventory(infra.Name(), infra.Type(), "v2.0.0", "infra-system", "")

	client := newFakeClient(config1).
		WithRepository(repository1).
		WithRepository(repository2).
		WithCluster(cluster1)

	return client
}

func fakeProvider(name string, providerType undistrov1.ProviderType, version, targetNamespace string) undistrov1.Provider {
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
		WatchedNamespace: "",
	}
}

func Test_parseUpgradeItem(t *testing.T) {
	type args struct {
		provider string
	}
	tests := []struct {
		name    string
		args    args
		want    *cluster.UpgradeItem
		wantErr bool
	}{
		{
			name: "namespace/provider",
			args: args{
				provider: "namespace/provider",
			},
			want: &cluster.UpgradeItem{
				Provider: undistrov1.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "namespace",
						Name:      undistrov1.ManifestLabel("provider", undistrov1.CoreProviderType),
					},
					ProviderName: "provider",
					Type:         string(undistrov1.CoreProviderType),
				},
				NextVersion: "",
			},
			wantErr: false,
		},
		{
			name: "namespace/provider:version",
			args: args{
				provider: "namespace/provider:version",
			},
			want: &cluster.UpgradeItem{
				Provider: undistrov1.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "namespace",
						Name:      undistrov1.ManifestLabel("provider", undistrov1.CoreProviderType),
					},
					ProviderName: "provider",
					Type:         string(undistrov1.CoreProviderType),
				},
				NextVersion: "version",
			},
			wantErr: false,
		},
		{
			name: "namespace missing",
			args: args{
				provider: "provider:version",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "namespace empty",
			args: args{
				provider: "/provider:version",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got, err := parseUpgradeItem(tt.args.provider, undistrov1.CoreProviderType)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(got).To(Equal(tt.want))
		})
	}
}
