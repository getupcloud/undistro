/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"sort"

	"github.com/getupio-undistro/undistro/client"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade core and provider components in a management cluster.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	upgradeCmd.AddCommand(upgradePlanCmd)
	upgradeCmd.AddCommand(upgradeApplyCmd)
	RootCmd.AddCommand(upgradeCmd)
}

func sortUpgradeItems(plan client.UpgradePlan) {
	sort.Slice(plan.Providers, func(i, j int) bool {
		return plan.Providers[i].Provider.Type < plan.Providers[j].Provider.Type ||
			(plan.Providers[i].Provider.Type == plan.Providers[j].Provider.Type && plan.Providers[i].Provider.Name < plan.Providers[j].Provider.Name) ||
			(plan.Providers[i].Provider.Type == plan.Providers[j].Provider.Type && plan.Providers[i].Provider.Name == plan.Providers[j].Provider.Name && plan.Providers[i].Provider.Namespace < plan.Providers[j].Provider.Namespace)
	})
}

func sortUpgradePlans(upgradePlans []client.UpgradePlan) {
	sort.Slice(upgradePlans, func(i, j int) bool {
		return upgradePlans[i].CoreProvider.Namespace < upgradePlans[j].CoreProvider.Namespace ||
			(upgradePlans[i].CoreProvider.Namespace == upgradePlans[j].CoreProvider.Namespace && upgradePlans[i].Contract < upgradePlans[j].Contract)
	})
}

func prettifyTargetVersion(version string) string {
	if version == "" {
		return "Already up to date"
	}
	return version
}
