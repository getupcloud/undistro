/*
Copyright 2020 Getup Cloud. All rights reserved.
*/
package cluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getupcloud/undistro/client/cluster/helm"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	utilkubeconfig "github.com/getupcloud/undistro/internal/kubeconfig"
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
	cfg, err := p.GetKubeconfig(workloadClusterName, namespace)
	if err != nil {
		return nil, err
	}
	path := filepath.Join("configs", namespace, fmt.Sprintf("%s.kubeconfig", workloadClusterName))
	_, err = os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	_, derr := os.Stat(filepath.Dir(path))
	if os.IsNotExist(derr) {
		derr = os.MkdirAll(filepath.Dir(path), 0700)
		if derr != nil {
			return nil, derr
		}
	}
	if err == nil {
		// remove if already exists to ensure we will use the last version
		os.RemoveAll(path)
	}
	err = ioutil.WriteFile(path, []byte(cfg), 0666)
	if err != nil {
		return nil, err
	}
	return helm.New(path), nil
}
