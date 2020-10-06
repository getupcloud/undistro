/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"testing"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/client/cluster"
	"github.com/getupio-undistro/undistro/client/config"
	. "github.com/onsi/gomega"
)

func Test_undistroClient_Move(t *testing.T) {
	type fields struct {
		client *fakeClient
	}
	type args struct {
		options MoveOptions
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
				client: fakeClientForMove(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: MoveOptions{
					FromKubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ToKubeconfig:   Kubeconfig{Path: "kubeconfig", Context: "worker-context"},
				},
			},
			wantErr: false,
		},
		{
			name: "returns an error if from cluster client is not found",
			fields: fields{
				client: fakeClientForMove(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: MoveOptions{
					FromKubeconfig: Kubeconfig{Path: "kubeconfig", Context: "does-not-exist"},
					ToKubeconfig:   Kubeconfig{Path: "kubeconfig", Context: "worker-context"},
				},
			},
			wantErr: true,
		},
		{
			name: "returns an error if to cluster client is not found",
			fields: fields{
				client: fakeClientForMove(), // core v1.0.0 (v1.0.1 available), infra v2.0.0 (v2.0.1 available)
			},
			args: args{
				options: MoveOptions{
					FromKubeconfig: Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"},
					ToKubeconfig:   Kubeconfig{Path: "kubeconfig", Context: "does-not-exist"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := tt.fields.client.Move(tt.args.options)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
		})
	}
}

func fakeClientForMove() *fakeClient {
	core := config.NewProvider("cluster-api", "https://somewhere.com", undistrov1.CoreProviderType, nil, nil)
	infra := config.NewProvider("infra", "https://somewhere.com", undistrov1.InfrastructureProviderType, nil, nil)

	config1 := newFakeConfig().
		WithProvider(core).
		WithProvider(infra)

	cluster1 := newFakeCluster(cluster.Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"}, config1).
		WithProviderInventory(core.Name(), core.Type(), "v1.0.0", "cluster-api-system", "").
		WithProviderInventory(infra.Name(), infra.Type(), "v2.0.0", "infra-system", "").
		WithObjectMover(&fakeObjectMover{})

	// Creating this cluster for move_test
	cluster2 := newFakeCluster(cluster.Kubeconfig{Path: "kubeconfig", Context: "worker-context"}, config1).
		WithProviderInventory(core.Name(), core.Type(), "v1.0.0", "cluster-api-system", "").
		WithProviderInventory(infra.Name(), infra.Type(), "v2.0.0", "infra-system", "")

	client := newFakeClient(config1).
		WithCluster(cluster1).
		WithCluster(cluster2)

	return client
}

type fakeObjectMover struct {
	moveErr error
}

func (f *fakeObjectMover) Move(namespace string, toCluster cluster.Client) error {
	return f.moveErr
}
