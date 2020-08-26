/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"github.com/getupcloud/undistro/client/cluster"
	"github.com/getupcloud/undistro/client/config"
	"github.com/getupcloud/undistro/client/repository"
	yaml "github.com/getupcloud/undistro/client/yamlprocessor"
)

// Alias creates local aliases for types defined in the low-level libraries.
// By using a local alias, we ensure that users import and use undistro's high-level library.

// Provider defines a provider configuration.
type Provider config.Provider

// Components wraps a YAML file that defines the provider's components (CRDs, controller, RBAC rules etc.).
type Components repository.Components

// ComponentsOptions wraps inputs to get provider's components
type ComponentsOptions repository.ComponentsOptions

// Template wraps a YAML file that defines the cluster objects (Cluster, Machines etc.).
type Template repository.Template

// UpgradePlan defines a list of possible upgrade targets for a management group.
type UpgradePlan cluster.UpgradePlan

// Kubeconfig is a type that specifies inputs related to the actual kubeconfig.
type Kubeconfig cluster.Kubeconfig

// Processor defines the methods necessary for creating a specific yaml
// processor.
type Processor yaml.Processor

// Variables defines a list o variables necessary for config
type Variables config.VariablesClient

// Proxy is the cluster proxy
type Proxy cluster.Proxy

// Logs of providers controller
type Logs cluster.LogStreamer

type WorkloadCluster cluster.WorkloadCluster
