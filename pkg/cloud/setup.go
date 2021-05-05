/*
Copyright 2020 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cloud

import (
	"context"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/cloud/aws"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Account interface {
	GetID() string
	GetUsername() string
	IsRoot() bool
}

// ReconcileNetwork from clouds
func ReconcileNetwork(ctx context.Context, r client.Client, cl *appv1alpha1.Cluster, capiCluster *capi.Cluster) error {
	if capiCluster.Spec.InfrastructureRef == nil || capiCluster.Spec.ControlPlaneRef == nil {
		return nil
	}
	if capiCluster.Spec.ClusterNetwork != nil {
		cl.Spec.Network.ClusterNetwork = *capiCluster.Spec.ClusterNetwork
	}
	switch cl.Spec.InfrastructureProvider.Name {
	case "aws":
		return aws.ReconcileNetwork(ctx, r, cl, capiCluster)
	}
	return nil
}

func PostInstall(ctx context.Context, c client.Client, p configv1alpha1.Provider) error {
	switch p.Spec.ProviderName {
	case "undistro-aws":
		return aws.PostInstall(ctx, c)
	}
	return nil
}

// Init providers
func Init(ctx context.Context, c client.Client, p configv1alpha1.Provider) (configv1alpha1.Provider, error) {
	var err error
	switch p.Spec.ProviderName {
	case "undistro-aws":
		p.Spec.ConfigurationFrom, err = aws.Init(ctx, c, p.Spec.ConfigurationFrom, p.Spec.ProviderVersion)
		if err != nil {
			return p, err
		}
	}
	return p, nil
}

// Upgrade providers
func Upgrade(ctx context.Context, c client.Client, p configv1alpha1.Provider) (configv1alpha1.Provider, error) {
	var err error
	switch p.Spec.ProviderName {
	case "undistro-aws":
		p.Spec.ConfigurationFrom, err = aws.Upgrade(ctx, c, p.Spec.ConfigurationFrom, p.Spec.ProviderVersion)
		if err != nil {
			return p, err
		}
	}
	return p, nil
}

// InstallTools install required tools for provider
func InstallTools(ctx context.Context, streams genericclioptions.IOStreams, providerName string) error {
	var err error
	switch providerName {
	case "aws":
		err = aws.InstallTools(ctx, streams)
		if err != nil {
			return err
		}
	}
	return nil
}

func DefaultRegion(infra string) string {
	switch infra {
	case "aws":
		return aws.DefaultAWSRegion
	}
	return ""
}

func GetAccount(ctx context.Context, c client.Client, cl *appv1alpha1.Cluster) (Account, error) {
	switch cl.Spec.InfrastructureProvider.Name {
	case "aws":
		return aws.NewAccount(ctx, c)
	}
	return nil, nil
}
