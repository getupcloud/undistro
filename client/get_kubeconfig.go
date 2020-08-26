/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

//GetKubeconfigOptions carries all the options supported by GetKubeconfig
type GetKubeconfigOptions struct {
	// Kubeconfig defines the kubeconfig to use for accessing the management cluster. If empty,
	// default rules for kubeconfig discovery will be used.
	Kubeconfig Kubeconfig

	// Namespace is the namespace in which secret is placed.
	Namespace string

	// WorkloadClusterName is the name of the workload cluster.
	WorkloadClusterName string
}

func (c *undistroClient) GetWorkloadCluster(cfg Kubeconfig) (WorkloadCluster, error) {
	// gets access to the management cluster
	clusterClient, err := c.clusterClientFactory(ClusterClientFactoryInput{kubeconfig: cfg})
	if err != nil {
		return nil, err
	}
	return clusterClient.WorkloadCluster(), nil
}
