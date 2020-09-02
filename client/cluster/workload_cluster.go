/*
Copyright 2020 Getup Cloud. All rights reserved.
*/
package cluster

import (
	"github.com/getupcloud/undistro/client/cluster/helm"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	utilkubeconfig "sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkloadCluster has methods for fetching kubeconfig of workload cluster from management cluster.
type WorkloadCluster interface {
	// GetKubeconfig returns the kubeconfig of the workload cluster.
	GetKubeconfig(workloadClusterName string, namespace string) (string, error)
	// GetRestConfig returns the *rest.Config of the workload cluster.
	GetRestConfig(workloadClusterName string, namespace string) (*rest.Config, error)
	// GetHelm returns helm.Client
	GetHelm(workloadClusterName string, namespace string) (helm.Client, error)
}

// workloadCluster implements WorkloadCluster.
type workloadCluster struct {
	proxy Proxy
}

// newWorkloadCluster returns a workloadCluster.
func newWorkloadCluster(proxy Proxy) *workloadCluster {
	return &workloadCluster{
		proxy: proxy,
	}
}

func (p *workloadCluster) GetKubeconfig(workloadClusterName string, namespace string) (string, error) {
	cs, err := p.proxy.NewClient()
	if err != nil {
		return "", err
	}
	if namespace == "" {
		namespace = "default"
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

func (p *workloadCluster) GetRestConfig(workloadClusterName string, namespace string) (*rest.Config, error) {
	k, err := p.GetKubeconfig(workloadClusterName, namespace)
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.NewClientConfigFromBytes([]byte(k))
	if err != nil {
		return nil, err
	}
	workloadCfg, err := cfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	return workloadCfg, nil
}

func (p *workloadCluster) GetHelm(workloadClusterName string, namespace string) (helm.Client, error) {
	cfg, err := p.GetRestConfig(workloadClusterName, namespace)
	if err != nil {
		return nil, err
	}
	return helm.New(cfg), nil
}
