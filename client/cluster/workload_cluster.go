/*
Copyright 2020 Getup Cloud. All rights reserved.
*/
package cluster

import (
	"github.com/getupcloud/undistro/client/cluster/helm"
	"github.com/getupcloud/undistro/log"
	"github.com/pkg/errors"
	utilkubeconfig "sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkloadCluster has methods for fetching kubeconfig of workload cluster from management cluster.
type WorkloadCluster interface {
	// GetKubeconfig returns the kubeconfig of the workload cluster.
	GetKubeconfig(workloadClusterName string, namespace string) (string, error)
	// GetHelm instance
	GetHelm() helm.Client
}

// workloadCluster implements WorkloadCluster.
type workloadCluster struct {
	proxy      Proxy
	helmClient helm.Client
}

// newWorkloadCluster returns a workloadCluster.
func newWorkloadCluster(proxy Proxy) *workloadCluster {
	cfg, err := proxy.GetConfig()
	if err != nil {
		log.Log.Error(err, "couldn't get rest config")
	}
	return &workloadCluster{
		proxy:      proxy,
		helmClient: helm.New(cfg),
	}
}

func (p *workloadCluster) GetKubeconfig(workloadClusterName string, namespace string) (string, error) {
	cs, err := p.proxy.NewClient()
	if err != nil {
		return "", err
	}

	obj := client.ObjectKey{
		Namespace: namespace,
		Name:      workloadClusterName,
	}
	dataBytes, err := utilkubeconfig.FromSecret(ctx, cs, obj)
	if err != nil {
		return "", errors.Wrapf(err, "\"%s-kubeconfig\" not found in namespace %q", workloadClusterName, namespace)
	}
	return string(dataBytes), nil
}

func (p *workloadCluster) GetHelm() helm.Client {
	return p.helmClient
}
