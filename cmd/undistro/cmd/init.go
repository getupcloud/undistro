/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"fmt"

	"github.com/getupio-undistro/undistro/client"
	"github.com/spf13/cobra"
)

type initOptions struct {
	kubeconfig              string
	kubeconfigContext       string
	coreProvider            string
	bootstrapProviders      []string
	controlPlaneProviders   []string
	infrastructureProviders []string
	targetNamespace         string
	watchingNamespace       string
	listImages              bool
	clusterctl              bool
}

var initOpts = &initOptions{}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a management cluster.",
	Long: LongDesc(`
		Initialize a management cluster.

		Installs Cluster API core components, the kubeadm bootstrap provider,
		and the selected bootstrap and infrastructure providers.

		The management cluster must be an existing Kubernetes cluster, make sure
		to have enough privileges to install the desired components.

		Use 'undistro config providers' to get a list of available providers; if necessary, edit
		$HOME/.undistro/undistro.yaml file to add new provider or to customize existing ones.

		Some providers require environment variables to be set before running undistro init.
		Refer to the provider documentation, or use 'undistro config provider [name]' to get a list of required variables.

		See https://cluster-api.sigs.k8s.io for more details.`),

	Example: Examples(`
		# Initialize a management cluster, by installing the given infrastructure provider.
		#
		# Note: when this command is executed on an empty management cluster,
 		#       it automatically triggers the installation of the Cluster API core provider.
		undistro init --infrastructure=aws

		# Initialize a management cluster with a specific version of the given infrastructure provider.
		undistro init --infrastructure=aws:v0.4.1

		# Initialize a management cluster with a custom kubeconfig path and the given infrastructure provider.
		undistro init --kubeconfig=foo.yaml  --infrastructure=aws

		# Initialize a management cluster with multiple infrastructure providers.
		undistro init --infrastructure=aws,vsphere

		# Initialize a management cluster with a custom target namespace for the provider resources.
		undistro init --infrastructure aws --target-namespace foo

		# Initialize a management cluster with a custom watching namespace for the given provider.
		undistro init --infrastructure aws --watching-namespace=foo

		# Lists the container images required for initializing the management cluster.
		#
		# Note: This command is a dry-run; it won't perform any action other than printing to screen.
		undistro init --infrastructure aws --list-images`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	initCmd.Flags().StringVar(&initOpts.kubeconfig, "kubeconfig", "",
		"Path to the kubeconfig for the management cluster. If unspecified, default discovery rules apply.")
	initCmd.Flags().StringVar(&initOpts.kubeconfigContext, "kubeconfig-context", "",
		"Context to be used within the kubeconfig file. If empty, current context will be used.")
	initCmd.Flags().StringVar(&initOpts.coreProvider, "core", "",
		"Core provider version (e.g. cluster-api:v0.3.0) to add to the management cluster. If unspecified, Cluster API's latest release is used.")
	initCmd.Flags().StringSliceVarP(&initOpts.infrastructureProviders, "infrastructure", "i", nil,
		"Infrastructure providers and versions (e.g. aws:v0.5.0) to add to the management cluster.")
	initCmd.Flags().StringSliceVarP(&initOpts.bootstrapProviders, "bootstrap", "b", nil,
		"Bootstrap providers and versions (e.g. kubeadm:v0.3.0) to add to the management cluster. If unspecified, Kubeadm bootstrap provider's latest release is used.")
	initCmd.Flags().StringSliceVarP(&initOpts.controlPlaneProviders, "control-plane", "c", nil,
		"Control plane providers and versions (e.g. kubeadm:v0.3.0) to add to the management cluster. If unspecified, the Kubeadm control plane provider's latest release is used.")
	initCmd.Flags().StringVar(&initOpts.targetNamespace, "target-namespace", "undistro-system",
		"The target namespace where the providers should be deployed. If unspecified, the provider components' default namespace is used.")
	initCmd.Flags().StringVar(&initOpts.watchingNamespace, "watching-namespace", "",
		"Namespace the providers should watch when reconciling objects. If unspecified, all namespaces are watched.")
	initCmd.Flags().BoolVar(&initOpts.listImages, "list-images", false,
		"Lists the container images required for initializing the management cluster (without actually installing the providers)")
	initCmd.Flags().BoolVar(&initOpts.clusterctl, "clusterctl", false,
		"change the init behavior to be the same of clusterctl")

	RootCmd.AddCommand(initCmd)
}

func runInit() error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}

	if initOpts.clusterctl {
		initOpts.targetNamespace = ""
	}

	options := client.InitOptions{
		Kubeconfig:              client.Kubeconfig{Path: initOpts.kubeconfig, Context: initOpts.kubeconfigContext},
		CoreProvider:            initOpts.coreProvider,
		BootstrapProviders:      initOpts.bootstrapProviders,
		ControlPlaneProviders:   initOpts.controlPlaneProviders,
		InfrastructureProviders: initOpts.infrastructureProviders,
		TargetNamespace:         initOpts.targetNamespace,
		WatchingNamespace:       initOpts.watchingNamespace,
		LogUsageInstructions:    true,
	}

	if initOpts.listImages {
		images, err := c.InitImages(options)
		if err != nil {
			return err
		}

		for _, i := range images {
			fmt.Println(i)
		}
		return nil
	}

	if _, err := c.Init(options); err != nil {
		return err
	}
	return nil
}
