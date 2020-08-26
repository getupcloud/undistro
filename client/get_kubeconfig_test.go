/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"testing"

	"github.com/getupcloud/undistro/client/cluster"
	"github.com/getupcloud/undistro/internal/test"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func Test_clusterctlClient_GetKubeconfig(t *testing.T) {

	configClient := newFakeConfig()
	kubeconfig := cluster.Kubeconfig{Path: "kubeconfig", Context: "mgmt-context"}
	clusterClient := &fakeClusterClient{
		kubeconfig: kubeconfig,
		workloadCluster: &fakeWorkloadCluster{
			Error: errors.Errorf("Cluster for kubeconfig %q and/or context %q does not exist.", "", ""),
		},
	}
	// create a clusterctl client where the proxy returns an empty namespace
	clusterClient.fakeProxy = test.NewFakeProxy().WithNamespace("")
	badClient := newFakeClient(configClient).WithCluster(clusterClient)

	tests := []struct {
		name      string
		client    *fakeClient
		options   GetKubeconfigOptions
		expectErr bool
	}{
		{
			name:      "returns error if unable namespace is empty",
			client:    badClient,
			options:   GetKubeconfigOptions{Kubeconfig: Kubeconfig(kubeconfig)},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			wc, err := tt.client.GetWorkloadCluster(tt.options.Kubeconfig)
			g.Expect(err).NotTo(HaveOccurred())
			config, err := wc.GetKubeconfig(tt.options.WorkloadClusterName, tt.options.Namespace)
			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(config).ToNot(BeEmpty())
		})
	}
}
