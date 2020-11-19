/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"github.com/getupio-undistro/undistro/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type moveOptions struct {
	fromKubeconfig        string
	fromKubeconfigContext string
	toKubeconfig          string
	toKubeconfigContext   string
	namespace             string
	skipInit              bool
}

var mo = &moveOptions{}

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move Cluster API objects and all dependencies between management clusters.",
	Long: LongDesc(`
		Move Cluster API objects and all dependencies between management clusters.

		Note: The destination cluster MUST have the required provider components installed.`),

	Example: Examples(`
		Move Cluster API objects and all dependencies between management clusters.
		undistro move --to-kubeconfig=target-kubeconfig.yaml`),
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMove()
	},
}

func init() {
	moveCmd.Flags().StringVar(&mo.fromKubeconfig, "kubeconfig", "",
		"Path to the kubeconfig file for the source management cluster. If unspecified, default discovery rules apply.")
	moveCmd.Flags().StringVar(&mo.toKubeconfig, "to-kubeconfig", "",
		"Path to the kubeconfig file to use for the destination management cluster.")
	moveCmd.Flags().StringVar(&mo.fromKubeconfigContext, "kubeconfig-context", "",
		"Context to be used within the kubeconfig file for the source management cluster. If empty, current context will be used.")
	moveCmd.Flags().StringVar(&mo.toKubeconfigContext, "to-kubeconfig-context", "",
		"Context to be used within the kubeconfig file for the destination management cluster. If empty, current context will be used.")
	moveCmd.Flags().StringVarP(&mo.namespace, "namespace", "n", "",
		"The namespace where the workload cluster is hosted. If unspecified, the current context's namespace is used.")
	moveCmd.Flags().BoolVar(&mo.skipInit, "skip-init", false, "skip new cluster initialization")

	RootCmd.AddCommand(moveCmd)
}

func runMove() error {
	if mo.toKubeconfig == "" {
		return errors.New("please specify a target cluster using the --to-kubeconfig flag")
	}

	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}

	if err := c.Move(client.MoveOptions{
		FromKubeconfig: client.Kubeconfig{Path: mo.fromKubeconfig, Context: mo.fromKubeconfigContext},
		ToKubeconfig:   client.Kubeconfig{Path: mo.toKubeconfig, Context: mo.toKubeconfigContext},
		Namespace:      mo.namespace,
		SkipInit:       mo.skipInit,
	}); err != nil {
		return err
	}
	return nil
}
