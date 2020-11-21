/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"sort"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/client/cluster"
	"github.com/getupio-undistro/undistro/client/config"
	"github.com/getupio-undistro/undistro/client/repository"
	logf "github.com/getupio-undistro/undistro/log"
	"github.com/pkg/errors"
)

const NoopProvider = "-"

// InitOptions carries the options supported by Init.
type InitOptions struct {
	// Kubeconfig defines the kubeconfig to use for accessing the management cluster. If empty,
	// default rules for kubeconfig discovery will be used.
	Kubeconfig Kubeconfig

	// CoreProvider version (e.g. cluster-api:v0.3.0) to add to the management cluster. If unspecified, the
	// cluster-api core provider's latest release is used.
	CoreProvider string

	// UndistroProvider version (e.g. undistro:v0.3.0) to add to the management cluster. If unspecified, the
	// undistro provider's latest release is used.
	UndistroProvider string

	// BootstrapProviders and versions (e.g. kubeadm:v0.3.0) to add to the management cluster.
	// If unspecified, the kubeadm bootstrap provider's latest release is used.
	BootstrapProviders []string

	// InfrastructureProviders and versions (e.g. aws:v0.5.0) to add to the management cluster.
	InfrastructureProviders []string

	// ControlPlaneProviders and versions (e.g. kubeadm:v0.3.0) to add to the management cluster.
	// If unspecified, the kubeadm control plane provider latest release is used.
	ControlPlaneProviders []string

	// TargetNamespace defines the namespace where the providers should be deployed. If unspecified, each provider
	// will be installed in a provider's default namespace.
	TargetNamespace string

	// WatchingNamespace defines the namespace the providers should watch to reconcile Cluster API objects.
	// If unspecified, the providers watches for Cluster API objects across all namespaces.
	WatchingNamespace string

	// LogUsageInstructions instructs the init command to print the usage instructions in case of first run.
	LogUsageInstructions bool

	// skipVariables skips variable parsing in the provider components yaml.
	// It is set to true for listing images of provider components.
	skipVariables bool
}

func (c *undistroClient) defaultVariables() {
	v := c.GetVariables()
	v.Set("EXP_EKS", "true")
	v.Set("EXP_EKS_IAM", "true")
	v.Set("EXP_EKS_ADD_ROLES", "true")
	v.Set("EXP_MACHINE_POOL", "true")
	v.Set("EXP_CLUSTER_RESOURCE_SET", "true")
}

// Init initializes a management cluster by adding the requested list of providers.
func (c *undistroClient) Init(options InitOptions) ([]Components, error) {
	log := logf.Log
	// set default variables
	c.defaultVariables()
	// gets access to the management cluster
	cluster, err := c.clusterClientFactory(ClusterClientFactoryInput{kubeconfig: options.Kubeconfig})
	if err != nil {
		return nil, err
	}

	// ensure the custom resource definitions required by undistro are in place
	if err := cluster.ProviderInventory().EnsureCustomResourceDefinitions(); err != nil {
		return nil, err
	}

	// checks if the cluster already contains a Core provider.
	// if not we consider this the first time init is executed, and thus we enforce the installation of a core provider,
	// a bootstrap provider and a control-plane provider (if not already explicitly requested by the user)
	log.Info("Fetching providers")
	firstRun := c.addDefaultProviders(cluster, &options)

	// create an installer service, add the requested providers to the install queue and then perform validation
	// of the target state of the management cluster before starting the installation.
	installer, err := c.setupInstaller(cluster, options)
	if err != nil {
		return nil, err
	}

	// Before installing the providers, validates the management cluster resulting by the planned installation. The following checks are performed:
	// - There should be only one instance of the same provider per namespace.
	// - Instances of the same provider should not be fighting for objects (no watching overlap).
	// - Providers combines in valid management groups
	//   - All the providers should belong to one/only one management groups
	//   - All the providers in a management group must support the same API Version of Cluster API (contract)
	if err := installer.Validate(); err != nil {
		return nil, err
	}

	cm, err := cluster.CertManager()
	if err != nil {
		return nil, err
	}

	// Before installing the providers, ensure the cert-manager Webhook is in place.
	if err := cm.EnsureInstalled(); err != nil {
		return nil, err
	}

	components, err := installer.Install()
	if err != nil {
		return nil, err
	}

	// If this is the firstRun, then log the usage instructions.
	if firstRun && options.LogUsageInstructions {
		log.Info("")
		log.Info("Your management cluster has been initialized successfully!")
		log.Info("")
	}

	// Components is an alias for repository.Components; this makes the conversion from the two types
	aliasComponents := make([]Components, len(components))
	for i, components := range components {
		aliasComponents[i] = components
	}
	return aliasComponents, nil
}

// Init returns the list of images required for init.
func (c *undistroClient) InitImages(options InitOptions) ([]string, error) {
	// gets access to the management cluster
	cluster, err := c.clusterClientFactory(ClusterClientFactoryInput{kubeconfig: options.Kubeconfig})
	if err != nil {
		return nil, err
	}

	// checks if the cluster already contains a Core provider.
	// if not we consider this the first time init is executed, and thus we enforce the installation of a core provider,
	// a bootstrap provider and a control-plane provider (if not already explicitly requested by the user)
	c.addDefaultProviders(cluster, &options)

	// skip variable parsing when listing images
	options.skipVariables = true

	// create an installer service, add the requested providers to the install queue and then perform validation
	// of the target state of the management cluster before starting the installation.
	installer, err := c.setupInstaller(cluster, options)
	if err != nil {
		return nil, err
	}

	cm, err := cluster.CertManager()
	if err != nil {
		return nil, err
	}

	// Gets the list of container images required for the cert-manager (if not already installed).
	images, err := cm.Images()
	if err != nil {
		return nil, err
	}

	// Appends the list of container images required for the selected providers.
	images = append(images, installer.Images()...)

	sort.Strings(images)
	return images, nil
}

func (c *undistroClient) setupInstaller(cluster cluster.Client, options InitOptions) (cluster.ProviderInstaller, error) {
	installedProviders, err := cluster.ProviderInventory().List()
	if err != nil {
		return nil, err
	}
	installer := cluster.ProviderInstaller()

	addOptions := addToInstallerOptions{
		installer:          installer,
		targetNamespace:    options.TargetNamespace,
		watchingNamespace:  options.WatchingNamespace,
		skipVariables:      options.skipVariables,
		installedProviders: installedProviders,
	}

	if options.UndistroProvider != "" {
		if err := c.addToInstaller(addOptions, undistrov1.UndistroProviderType, options.UndistroProvider); err != nil {
			return nil, err
		}
	}

	if options.CoreProvider != "" {
		if err := c.addToInstaller(addOptions, undistrov1.CoreProviderType, options.CoreProvider); err != nil {
			return nil, err
		}
	}

	if err := c.addToInstaller(addOptions, undistrov1.BootstrapProviderType, options.BootstrapProviders...); err != nil {
		return nil, err
	}

	if err := c.addToInstaller(addOptions, undistrov1.ControlPlaneProviderType, options.ControlPlaneProviders...); err != nil {
		return nil, err
	}

	if err := c.addToInstaller(addOptions, undistrov1.InfrastructureProviderType, options.InfrastructureProviders...); err != nil {
		return nil, err
	}

	return installer, nil
}

func (c *undistroClient) addDefaultProviders(cluster cluster.Client, options *InitOptions) bool {
	firstRun := false
	// Check if there is already a core provider installed in the cluster
	// Nb. we are ignoring the error so this operation can support listing images even if there is no an existing management cluster;
	// in case there is no an existing management cluster, we assume there are no core providers installed in the cluster.
	currentCoreProvider, _ := cluster.ProviderInventory().GetDefaultProviderName(undistrov1.CoreProviderType)

	// If there are no core providers installed in the cluster, consider this a first run and add default providers to the list
	// of providers to be installed.
	if currentCoreProvider == "" {
		firstRun = true
		if options.CoreProvider == "" {
			options.CoreProvider = config.ClusterAPIProviderName
		}
		if options.UndistroProvider == "" {
			options.UndistroProvider = config.UndistroProviderName
		}
		if len(options.BootstrapProviders) == 0 {
			options.BootstrapProviders = append(options.BootstrapProviders, config.KubeadmBootstrapProviderName)
		}
		if len(options.ControlPlaneProviders) == 0 {
			options.ControlPlaneProviders = append(options.ControlPlaneProviders, config.KubeadmControlPlaneProviderName)
		}
	}
	return firstRun
}

type addToInstallerOptions struct {
	installer          cluster.ProviderInstaller
	installedProviders *undistrov1.ProviderList
	targetNamespace    string
	watchingNamespace  string
	skipVariables      bool
}

// addToInstaller adds the components to the install queue and checks that the actual provider type match the target group
func (c *undistroClient) addToInstaller(options addToInstallerOptions, providerType undistrov1.ProviderType, providers ...string) error {
	for _, provider := range providers {
		// It is possible to opt-out from automatic installation of bootstrap/control-plane providers using '-' as a provider name (NoopProvider).
		if provider == NoopProvider {
			if providerType == undistrov1.CoreProviderType {
				return errors.New("the '-' value can not be used for the core provider")
			}
			continue
		}
		installedProviders := options.installedProviders.FilterByProviderNameAndType(provider, providerType)
		p, err := c.configClient.Providers().Get(provider, providerType)
		if err != nil {
			logf.Log.V(5).Info("failed to get provider config:", "provider", provider, "type", providerType, "error", err)
		}
		if p != nil {
			initFunc := p.GetInitFunc()
			if initFunc != nil {
				err = initFunc(c.configClient, len(installedProviders) == 0)
				if err != nil {
					return errors.Errorf("failed to init func for %s: %v", provider, err)
				}
			}
		}
		componentsOptions := repository.ComponentsOptions{
			TargetNamespace:   options.targetNamespace,
			WatchingNamespace: options.watchingNamespace,
			SkipVariables:     options.skipVariables,
		}
		components, err := c.getComponentsByName(provider, providerType, componentsOptions)
		if err != nil {
			return errors.Wrapf(err, "failed to get provider components for the %q provider", provider)
		}

		if components.Type() != providerType {
			return errors.Errorf("can't use %q provider as an %q, it is a %q", provider, providerType, components.Type())
		}

		options.installer.Add(components)
	}
	return nil
}
