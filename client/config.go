/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package client

import (
	"io"
	"io/ioutil"
	"strconv"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client/cluster"
	"github.com/getupcloud/undistro/client/repository"
	yaml "github.com/getupcloud/undistro/client/yamlprocessor"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/utils/pointer"
)

func (c *undistroClient) GetVariables() Variables {
	return c.configClient.Variables()
}

func (c *undistroClient) GetProxy() (Proxy, error) {
	cluster, err := c.clusterClientFactory(
		ClusterClientFactoryInput{
			// use the default kubeconfig
			kubeconfig: Kubeconfig{},
		},
	)
	if err != nil {
		return nil, err
	}
	return cluster.Proxy(), nil
}

func (c *undistroClient) GetProvidersConfig() ([]Provider, error) {
	r, err := c.configClient.Providers().List()
	if err != nil {
		return nil, err
	}

	// Provider is an alias for config.Provider; this makes the conversion
	rr := make([]Provider, len(r))
	for i, provider := range r {
		rr[i] = provider
	}

	return rr, nil
}

func (c *undistroClient) GetProviderComponents(provider string, providerType undistrov1.ProviderType, options ComponentsOptions) (Components, error) {
	// ComponentsOptions is an alias for repository.ComponentsOptions; this makes the conversion
	inputOptions := repository.ComponentsOptions{
		Version:           options.Version,
		TargetNamespace:   options.TargetNamespace,
		WatchingNamespace: options.WatchingNamespace,
		SkipVariables:     options.SkipVariables,
	}
	components, err := c.getComponentsByName(provider, providerType, inputOptions)
	if err != nil {
		return nil, err
	}
	return components, nil
}

// ReaderSourceOptions define the options to be used when reading a template
// from an arbitrary reader
type ReaderSourceOptions struct {
	Reader io.Reader
}

// ProcessYAMLOptions are the options supported by ProcessYAML.
type ProcessYAMLOptions struct {
	ReaderSource *ReaderSourceOptions
	// URLSource to be used for reading the template
	URLSource *URLSourceOptions

	// ListVariablesOnly return the list of variables expected by the template
	// without executing any further processing.
	ListVariablesOnly bool
}

func (c *undistroClient) ProcessYAML(options ProcessYAMLOptions) (YamlPrinter, error) {
	if options.ReaderSource != nil {
		// NOTE: Beware of potentially reading in large files all at once
		// since this is inefficient and increases memory utilziation.
		content, err := ioutil.ReadAll(options.ReaderSource.Reader)
		if err != nil {
			return nil, err
		}
		return repository.NewTemplate(repository.TemplateInput{
			RawArtifact:           content,
			ConfigVariablesClient: c.configClient.Variables(),
			Processor:             yaml.NewSimpleProcessor(),
			TargetNamespace:       "",
			ListVariablesOnly:     options.ListVariablesOnly,
		})
	}

	// Technically we do not need to connect to the cluster. However, we are
	// leveraging the template client which exposes GetFromURL() is available
	// on the cluster client so we create a cluster client with default
	// configs to access it.
	cluster, err := c.clusterClientFactory(
		ClusterClientFactoryInput{
			// use the default kubeconfig
			kubeconfig: Kubeconfig{},
		},
	)
	if err != nil {
		return nil, err
	}

	if options.URLSource != nil {
		return c.getTemplateFromURL(cluster, *options.URLSource, "", options.ListVariablesOnly)
	}

	return nil, errors.New("unable to read custom template. Please specify a template source")
}

// GetClusterTemplateOptions carries the options supported by GetClusterTemplate.
type GetClusterTemplateOptions struct {
	// Kubeconfig defines the kubeconfig to use for accessing the management cluster. If empty,
	// default rules for kubeconfig discovery will be used.
	Kubeconfig Kubeconfig

	// ProviderRepositorySource to be used for reading the workload cluster template from a provider repository;
	// only one template source can be used at time; if not other source will be set, a ProviderRepositorySource
	// will be generated inferring values from the cluster.
	ProviderRepositorySource *ProviderRepositorySourceOptions

	// URLSource to be used for reading the workload cluster template; only one template source can be used at time.
	URLSource *URLSourceOptions

	// ConfigMapSource to be used for reading the workload cluster template; only one template source can be used at time.
	ConfigMapSource *ConfigMapSourceOptions

	// TargetNamespace where the objects describing the workload cluster should be deployed. If unspecified,
	// the current namespace will be used.
	TargetNamespace string

	// ClusterName to be used for the workload cluster.
	ClusterName string

	// KubernetesVersion to use for the workload cluster. If unspecified, the value from os env variables
	// or the .undistro/undistro.yaml config file will be used.
	KubernetesVersion string

	// ControlPlaneMachineCount defines the number of control plane machines to be added to the workload cluster.
	// It can be set through the cli flag, CONTROL_PLANE_MACHINE_COUNT environment variable or will default to 1
	ControlPlaneMachineCount *int64

	// WorkerMachineCount defines number of worker machines to be added to the workload cluster.
	// It can be set through the cli flag, WORKER_MACHINE_COUNT environment variable or will default to 0
	WorkerMachineCount *int64

	// ListVariablesOnly sets the GetClusterTemplate method to return the list of variables expected by the template
	// without executing any further processing.
	ListVariablesOnly bool

	// YamlProcessor defines the yaml processor to use for the cluster
	// template processing. If not defined, SimpleProcessor will be used.
	YamlProcessor Processor
}

// numSources return the number of template sources currently set on a GetClusterTemplateOptions.
func (o *GetClusterTemplateOptions) numSources() int {
	numSources := 0
	if o.ProviderRepositorySource != nil {
		numSources++
	}
	if o.ConfigMapSource != nil {
		numSources++
	}
	if o.URLSource != nil {
		numSources++
	}
	return numSources
}

// ProviderRepositorySourceOptions defines the options to be used when reading a workload cluster template
// from a provider repository.
type ProviderRepositorySourceOptions struct {
	// InfrastructureProvider to read the workload cluster template from. If unspecified, the default
	// infrastructure provider will be used if no other sources are specified.
	InfrastructureProvider string

	// Flavor defines The workload cluster template variant to be used when reading from the infrastructure
	// provider repository. If unspecified, the default cluster template will be used.
	Flavor string
}

// URLSourceOptions defines the options to be used when reading a workload cluster template from an URL.
type URLSourceOptions struct {
	// URL to read the workload cluster template from.
	URL string
}

// DefaultCustomTemplateConfigMapKey  where the workload cluster template is hosted.
const DefaultCustomTemplateConfigMapKey = "template"

// ConfigMapSourceOptions defines the options to be used when reading a workload cluster template from a ConfigMap.
type ConfigMapSourceOptions struct {
	// Namespace where the ConfigMap exists. If unspecified, the current namespace will be used.
	Namespace string

	// Name to read the workload cluster template from.
	Name string

	// DataKey where the workload cluster template is hosted. If unspecified, the
	// DefaultCustomTemplateConfigMapKey will be used.
	DataKey string
}

func (c *undistroClient) GetClusterTemplate(options GetClusterTemplateOptions) (Template, error) {
	// Checks that no more than on source is set
	numsSource := options.numSources()
	if numsSource > 1 {
		return nil, errors.New("invalid cluster template source: only one template can be used at time")
	}

	// If no source is set, defaults to using an empty ProviderRepositorySource so values will be
	// inferred from the cluster inventory.
	if numsSource == 0 {
		options.ProviderRepositorySource = &ProviderRepositorySourceOptions{}
	}

	// Gets  the client for the current management cluster
	cluster, err := c.clusterClientFactory(ClusterClientFactoryInput{options.Kubeconfig, options.YamlProcessor})
	if err != nil {
		return nil, err
	}

	// If the option specifying the targetNamespace is empty, try to detect it.
	if options.TargetNamespace == "" {
		currentNamespace, err := cluster.Proxy().CurrentNamespace()
		if err != nil {
			return nil, err
		}
		if currentNamespace == "" {
			return nil, errors.New("failed to identify the current namespace. Please specify a target namespace")
		}
		options.TargetNamespace = currentNamespace
	}

	// Inject some of the templateOptions into the configClient so they can be consumed as a variables from the template.
	if err := c.templateOptionsToVariables(options); err != nil {
		return nil, err
	}

	// Gets the workload cluster template from the selected source
	if options.ProviderRepositorySource != nil {
		return c.getTemplateFromRepository(cluster, options)
	}
	if options.ConfigMapSource != nil {
		return c.getTemplateFromConfigMap(cluster, *options.ConfigMapSource, options.TargetNamespace, options.ListVariablesOnly)
	}
	if options.URLSource != nil {
		return c.getTemplateFromURL(cluster, *options.URLSource, options.TargetNamespace, options.ListVariablesOnly)
	}

	return nil, errors.New("unable to read custom template. Please specify a template source")
}

// getTemplateFromRepository returns a workload cluster template from a provider repository.
func (c *undistroClient) getTemplateFromRepository(cluster cluster.Client, options GetClusterTemplateOptions) (Template, error) {
	source := *options.ProviderRepositorySource
	targetNamespace := options.TargetNamespace
	listVariablesOnly := options.ListVariablesOnly
	processor := options.YamlProcessor

	// If the option specifying the name of the infrastructure provider to get templates from is empty, try to detect it.
	provider := source.InfrastructureProvider
	ensureCustomResourceDefinitions := false
	if provider == "" {
		// ensure the custom resource definitions required by undistro are in place
		if err := cluster.ProviderInventory().EnsureCustomResourceDefinitions(); err != nil {
			return nil, errors.Wrapf(err, "failed to identify the default infrastructure provider. Please specify an infrastructure provider")
		}
		ensureCustomResourceDefinitions = true

		defaultProviderName, err := cluster.ProviderInventory().GetDefaultProviderName(undistrov1.InfrastructureProviderType)
		if err != nil {
			return nil, err
		}

		if defaultProviderName == "" {
			return nil, errors.New("failed to identify the default infrastructure provider. Please specify an infrastructure provider")
		}
		provider = defaultProviderName
	}

	// parse the abbreviated syntax for name[:version]
	name, version, err := parseProviderName(provider)
	if err != nil {
		return nil, err
	}

	// If the version of the infrastructure provider to get templates from is empty, try to detect it.
	if version == "" {
		// ensure the custom resource definitions required by undistro are in place (if not already done)
		if !ensureCustomResourceDefinitions {
			if err := cluster.ProviderInventory().EnsureCustomResourceDefinitions(); err != nil {
				return nil, errors.Wrapf(err, "failed to identify the default version for the provider %q. Please specify a version", name)
			}
		}

		defaultProviderVersion, err := cluster.ProviderInventory().GetDefaultProviderVersion(name, undistrov1.InfrastructureProviderType)
		if err != nil {
			return nil, err
		}

		if defaultProviderVersion == "" {
			return nil, errors.Errorf("failed to identify the default version for the provider %q. Please specify a version", name)
		}
		version = defaultProviderVersion
	}

	// Get the template from the template repository.
	providerConfig, err := c.configClient.Providers().Get(name, undistrov1.InfrastructureProviderType)
	if err != nil {
		return nil, err
	}

	repo, err := c.repositoryClientFactory(RepositoryClientFactoryInput{provider: providerConfig, processor: processor})
	if err != nil {
		return nil, err
	}

	template, err := repo.Templates(version).Get(source.Flavor, targetNamespace, listVariablesOnly)
	if err != nil {
		return nil, err
	}
	return template, nil
}

// getTemplateFromConfigMap returns a workload cluster template from a ConfigMap.
func (c *undistroClient) getTemplateFromConfigMap(cluster cluster.Client, source ConfigMapSourceOptions, targetNamespace string, listVariablesOnly bool) (Template, error) {
	// If the option specifying the configMapNamespace is empty, default it to the current namespace.
	if source.Namespace == "" {
		currentNamespace, err := cluster.Proxy().CurrentNamespace()
		if err != nil {
			return nil, err
		}
		source.Namespace = currentNamespace
	}

	// If the option specifying the configMapDataKey is empty, default it.
	if source.DataKey == "" {
		source.DataKey = DefaultCustomTemplateConfigMapKey
	}

	return cluster.Template().GetFromConfigMap(source.Namespace, source.Name, source.DataKey, targetNamespace, listVariablesOnly)
}

// getTemplateFromURL returns a workload cluster template from an URL.
func (c *undistroClient) getTemplateFromURL(cluster cluster.Client, source URLSourceOptions, targetNamespace string, listVariablesOnly bool) (Template, error) {
	return cluster.Template().GetFromURL(source.URL, targetNamespace, listVariablesOnly)
}

// templateOptionsToVariables injects some of the templateOptions to the configClient so they can be consumed as a variables from the template.
func (c *undistroClient) templateOptionsToVariables(options GetClusterTemplateOptions) error {

	// the TargetNamespace, if valid, can be used in templates using the ${ NAMESPACE } variable.
	if err := validateDNS1123Label(options.TargetNamespace); err != nil {
		return errors.Wrapf(err, "invalid target-namespace")
	}
	c.configClient.Variables().Set("NAMESPACE", options.TargetNamespace)

	// the ClusterName, if valid, can be used in templates using the ${ CLUSTER_NAME } variable.
	if err := validateDNS1123Domanin(options.ClusterName); err != nil {
		return errors.Wrapf(err, "invalid cluster name")
	}
	c.configClient.Variables().Set("CLUSTER_NAME", options.ClusterName)

	// the KubernetesVersion, if valid, can be used in templates using the ${ KUBERNETES_VERSION } variable.
	// NB. in case the KubernetesVersion from the templateOptions is empty, we are not setting any values so the
	// configClient is going to search into os env variables/the undistro config file as a fallback options.
	if options.KubernetesVersion != "" {
		if _, err := version.ParseSemantic(options.KubernetesVersion); err != nil {
			return errors.Errorf("invalid KubernetesVersion. Please use a semantic version number")
		}
		c.configClient.Variables().Set("KUBERNETES_VERSION", options.KubernetesVersion)
	}

	// the ControlPlaneMachineCount, if valid, can be used in templates using the ${ CONTROL_PLANE_MACHINE_COUNT } variable.
	if options.ControlPlaneMachineCount == nil {
		// Check if set through env variable and default to 1 otherwise
		if v, err := c.configClient.Variables().Get("CONTROL_PLANE_MACHINE_COUNT"); err != nil {
			options.ControlPlaneMachineCount = pointer.Int64Ptr(1)
		} else {
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return errors.Errorf("invalid value for CONTROL_PLANE_MACHINE_COUNT set")
			}
			options.ControlPlaneMachineCount = &i
		}
	}
	if *options.ControlPlaneMachineCount < 1 {
		return errors.Errorf("invalid ControlPlaneMachineCount. Please use a number greater or equal than 1")
	}
	c.configClient.Variables().Set("CONTROL_PLANE_MACHINE_COUNT", strconv.FormatInt(*options.ControlPlaneMachineCount, 10))

	// the WorkerMachineCount, if valid, can be used in templates using the ${ WORKER_MACHINE_COUNT } variable.
	if options.WorkerMachineCount == nil {
		// Check if set through env variable and default to 0 otherwise
		if v, err := c.configClient.Variables().Get("WORKER_MACHINE_COUNT"); err != nil {
			options.WorkerMachineCount = pointer.Int64Ptr(0)
		} else {
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return errors.Errorf("invalid value for WORKER_MACHINE_COUNT set")
			}
			options.WorkerMachineCount = &i
		}
	}
	if *options.WorkerMachineCount < 0 {
		return errors.Errorf("invalid WorkerMachineCount. Please use a number greater or equal than 0")
	}
	c.configClient.Variables().Set("WORKER_MACHINE_COUNT", strconv.FormatInt(*options.WorkerMachineCount, 10))

	return nil
}
