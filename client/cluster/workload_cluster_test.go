/*
Copyright 2020 Getup Cloud. All rights reserved.
*/
package cluster

import (
	"testing"

	"github.com/getupio-undistro/undistro/internal/test"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/secret"
)

func Test_WorkloadCluster_GetKubeconfig(t *testing.T) {

	var (
		validKubeConfig = `
clusters:
- cluster:
    certificate-authority-data: stuff
    server: https://test-cluster-api:6443
  name: test1
contexts:
- context:
    cluster: test1
    user: test1-admin
  name: test1-admin@test1
current-context: test1-admin@test1
kind: Config
preferences: {}
users:
- name: test1-admin
  user:
    client-certificate-data: stuff-cert-data
    client-key-data: stuff-key-data
`

		validSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1-kubeconfig",
				Namespace: "test",
				Labels:    map[string]string{clusterv1.ClusterLabelName: "test1"},
			},
			Data: map[string][]byte{
				secret.KubeconfigDataName: []byte(validKubeConfig),
			},
		}
	)

	tests := []struct {
		name      string
		expectErr bool
		proxy     Proxy
	}{
		{
			name:      "return secret data",
			expectErr: false,
			proxy:     test.NewFakeProxy().WithObjs(validSecret),
		},
		{
			name:      "return error if cannot find secert",
			expectErr: true,
			proxy:     test.NewFakeProxy(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			wc := newWorkloadCluster(tt.proxy)
			data, err := wc.GetKubeconfig("test1", "test")

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(data).To(Equal(string(validSecret.Data[secret.KubeconfigDataName])))
		})
	}

}
