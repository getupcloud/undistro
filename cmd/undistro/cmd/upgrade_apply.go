/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"github.com/getupio-undistro/undistro/client"
	"github.com/spf13/cobra"
)

type upgradeApplyOptions struct {
	kubeconfig        string
	kubeconfigContext string
	managementGroup   string
	contract          string
}

var ua = &upgradeApplyOptions{}

var upgradeApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply new versions of Cluster API core and providers in a management cluster",
	Long: LongDesc(`
		The upgrade apply command applies new versions of Cluster API providers as defined by undistro upgrade plan.

		New version should be applied for each management groups, ensuring all the providers on the same cluster API version
		in order to guarantee the proper functioning of the management cluster.`),

	Example: Examples(`
		# Upgrades all the providers in the capi-system/cluster-api management group to the latest version available which is compliant
		# to the v1alpha3 API Version of Cluster API (contract).
		undistro upgrade apply --management-group capi-system/cluster-api  --contract v1alpha3

		# Upgrades only the capa-system/aws provider instance in the capi-system/cluster-api management group to the v0.5.0 version.
		undistro upgrade apply --management-group capi-system/cluster-api  --infrastructure capa-system/aws:v0.5.0`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpgradeApply()
	},
}

func init() {
	upgradeApplyCmd.Flags().StringVar(&ua.kubeconfig, "kubeconfig", "",
		"Path to the kubeconfig file to use for accessing the management cluster. If unspecified, default discovery rules apply.")
	upgradeApplyCmd.Flags().StringVar(&ua.kubeconfigContext, "kubeconfig-context", "",
		"Context to be used within the kubeconfig file. If empty, current context will be used.")
	upgradeApplyCmd.Flags().StringVar(&ua.managementGroup, "management-group", "undistro-system/cluster-api",
		"The management group that should be upgraded (e.g. capi-system/cluster-api)")
	upgradeApplyCmd.Flags().StringVar(&ua.contract, "contract", "ndistro-system/cluster-api",
		"The API Version of Cluster API (contract, e.g. v1alpha3) the management group should upgrade to")
}

func runUpgradeApply() error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}

	if err := c.ApplyUpgrade(client.ApplyUpgradeOptions{
		Kubeconfig:      client.Kubeconfig{Path: ua.kubeconfig, Context: ua.kubeconfigContext},
		ManagementGroup: ua.managementGroup,
		Contract:        ua.contract,
	}); err != nil {
		return err
	}
	return nil
}
