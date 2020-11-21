/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"context"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/internal/util"
	"k8s.io/apimachinery/pkg/types"
)

// MoveOptions carries the options supported by move.
type MoveOptions struct {
	// FromKubeconfig defines the kubeconfig to use for accessing the source management cluster. If empty,
	// default rules for kubeconfig discovery will be used.
	FromKubeconfig Kubeconfig

	// ToKubeconfig defines the kubeconfig to use for accessing the target management cluster. If empty,
	// default rules for kubeconfig discovery will be used.
	ToKubeconfig Kubeconfig

	// Namespace where the objects describing the workload cluster exists. If unspecified, the current
	// namespace will be used.
	Namespace string

	SkipInit bool
}

func (c *undistroClient) Move(options MoveOptions) error {
	// Get the client for interacting with the source management cluster.
	fromCluster, err := c.clusterClientFactory(ClusterClientFactoryInput{kubeconfig: options.FromKubeconfig})
	if err != nil {
		return err
	}

	// Ensures the custom resource definitions required by undistro are in place.
	if err := fromCluster.ProviderInventory().EnsureCustomResourceDefinitions(); err != nil {
		return err
	}

	// Get the client for interacting with the target management cluster.
	toCluster, err := c.clusterClientFactory(ClusterClientFactoryInput{kubeconfig: options.ToKubeconfig})
	if err != nil {
		return err
	}
	if !options.SkipInit {
		clClient, err := fromCluster.Proxy().NewClient()
		if err != nil {
			return err
		}
		ctx := context.Background()
		clList := undistrov1.ClusterList{}
		if err := clClient.List(ctx, &clList); err != nil {
			return err
		}
		newClClient, err := toCluster.Proxy().NewClient()
		if err != nil {
			return err
		}
		for _, cl := range clList.Items {
			err = util.SetVariablesFromEnvVar(ctx, util.VariablesInput{
				VariablesClient: c.GetVariables(),
				ClientSet:       newClClient,
				NamespacedName: types.NamespacedName{
					Name:      cl.Name,
					Namespace: cl.Namespace,
				},
				EnvVars: cl.Spec.InfrastructureProvider.Env,
			})
			if err != nil {
				return err
			}
			initOpts := InitOptions{
				Kubeconfig:              options.ToKubeconfig,
				TargetNamespace:         "undistro-system",
				InfrastructureProviders: []string{cl.Spec.InfrastructureProvider.NameVersion()},
			}
			firstRun := c.addDefaultProviders(toCluster, &initOpts)
			if firstRun {
				bp, cp := cl.GetManagedProvidersInfra()
				if bp != nil {
					initOpts.BootstrapProviders = append(initOpts.BootstrapProviders, bp...)
				}
				if cp != nil {
					initOpts.ControlPlaneProviders = append(initOpts.ControlPlaneProviders, cp...)
				}
				if _, err := c.Init(initOpts); err != nil {
					return err
				}
			}
		}
	}

	// Ensures the custom resource definitions required by undistro are in place
	if err := toCluster.ProviderInventory().EnsureCustomResourceDefinitions(); err != nil {
		return err
	}

	// If the option specifying the Namespace is empty, try to detect it.
	if options.Namespace == "" {
		currentNamespace, err := fromCluster.Proxy().CurrentNamespace()
		if err != nil {
			return err
		}
		options.Namespace = currentNamespace
	}
	if err := fromCluster.ObjectMover().Move(options.Namespace, toCluster); err != nil {
		return err
	}

	return nil
}
