package cmd

import (
	"context"
	"io"

	"github.com/getupio-undistro/undistro/client"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
)

type logOptions struct {
	kubeconfig string
}

var logOpts = &logOptions{}

var logCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs in undistro namespace",
	Long:  "Get logs in undistro namespace",

	RunE: func(cmd *cobra.Command, args []string) error {
		return logsStream(cmd.Context(), cmd.OutOrStdout())
	},
}

func init() {
	// flags for the url source
	logCmd.Flags().StringVar(&logOpts.kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig file to use for the management cluster. If empty, default discovery rules apply.")

	RootCmd.AddCommand(logCmd)
}

func logsStream(ctx context.Context, w io.Writer) error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}
	kcfg := client.Kubeconfig{
		Path: logOpts.kubeconfig,
	}
	proxy, err := c.GetProxy(kcfg)
	if err != nil {
		return err
	}
	cfg, err := proxy.GetConfig()
	if err != nil {
		return err
	}
	steamer, err := c.GetLogs(kcfg)
	if err != nil {
		return err
	}
	return steamer.Stream(ctx, cfg, w, types.NamespacedName{}, nil, true)
}
