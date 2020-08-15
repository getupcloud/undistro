/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create workload clusters.",
	Long:  `Create workload clusters.`,
}

func init() {
	RootCmd.AddCommand(createCmd)
}
