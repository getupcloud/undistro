/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cmd

import (
	"fmt"

	"github.com/getupio-undistro/undistro/client"
	"github.com/spf13/cobra"
)

type getKubeconfigOptions struct {
	kubeconfig        string
	kubeconfigContext string
	namespace         string
}

var gk = &getKubeconfigOptions{}

var getKubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Gets the kubeconfig file for accessing a workload cluster",
	Long: LongDesc(`
		Gets the kubeconfig file for accessing a workload cluster`),

	Example: Examples(`
		# Get the workload cluster's kubeconfig.
		undistro get kubeconfig <name of workload cluster>
		# Get the workload cluster's kubeconfig in a particular namespace.
		undistro get kubeconfig <name of workload cluster> --namespace foo`),

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGetKubeconfig(args[0])
	},
}

func init() {
	getKubeconfigCmd.Flags().StringVarP(&gk.namespace, "namespace", "n", "",
		"Namespace where the workload cluster exist.")
	getKubeconfigCmd.Flags().StringVar(&gk.kubeconfig, "kubeconfig", "",
		"Path to the kubeconfig file to use for accessing the management cluster. If unspecified, default discovery rules apply.")
	getKubeconfigCmd.Flags().StringVar(&gk.kubeconfigContext, "kubeconfig-context", "",
		"Context to be used within the kubeconfig file. If empty, current context will be used.")
	getCmd.AddCommand(getKubeconfigCmd)
}

func runGetKubeconfig(workloadClusterName string) error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}

	options := client.GetKubeconfigOptions{
		Kubeconfig:          client.Kubeconfig{Path: gk.kubeconfig, Context: gk.kubeconfigContext},
		WorkloadClusterName: workloadClusterName,
		Namespace:           gk.namespace,
	}
	wc, err := c.GetWorkloadCluster(options.Kubeconfig)
	if err != nil {
		return err
	}
	out, err := wc.GetKubeconfig(options.WorkloadClusterName, options.Namespace)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
