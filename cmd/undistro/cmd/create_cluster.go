package cmd

import (
	"context"
	"io"
	"os"

	"github.com/getupcloud/undistro/client"
	"github.com/spf13/cobra"
	utilresource "sigs.k8s.io/cluster-api/util/resource"
	"sigs.k8s.io/cluster-api/util/yaml"
)

type createClusterOptions struct {
	url string
}

var ccOpts = &createClusterOptions{}

var createCluterCmd = &cobra.Command{
	Use:   "yaml",
	Short: "Create a cluster",
	Long:  "Create a cluster",

	Example: Examples(`
		# Creates a cluster with variable values using
		a template from a specific URL.
		undistro create cluster --from https://github.com/foo-org/foo-repository/blob/master/cluster-template.yaml

		# Creates a cluster with variable values using
		a template stored locally.
		undistro create cluster --from ~/workspace/cluster-template.yaml

		# Creates a cluster from template passed in via stdin
		cat ~/workspace/cluster-template.yaml | undistro create cluster
`),

	RunE: func(cmd *cobra.Command, args []string) error {
		return createCluster(os.Stdin, os.Stdout)
	},
}

func init() {
	// flags for the url source
	createCluterCmd.Flags().StringVarP(&ccOpts.url, "from", "f", "-",
		"The URL to read the template from. It defaults to '-' which reads from stdin.")

	createCmd.AddCommand(createCluterCmd)
}

func createCluster(r io.Reader, w io.Writer) error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}
	options := client.ProcessYAMLOptions{
		ListVariablesOnly: false,
	}
	if ccOpts.url != "" {
		if ccOpts.url == "-" {
			options.ReaderSource = &client.ReaderSourceOptions{
				Reader: r,
			}
		} else {
			options.URLSource = &client.URLSourceOptions{
				URL: ccOpts.url,
			}
		}
	}
	printer, err := c.ProcessYAML(options)
	if err != nil {
		return err
	}
	b, err := printer.Yaml()
	if err != nil {
		return err
	}
	objs, err := yaml.ToUnstructured(b)
	if err != nil {
		return err
	}
	proxy, err := c.GetProxy()
	if err != nil {
		return err
	}
	k8sClient, err := proxy.NewClient()
	if err != nil {
		return err
	}
	objs = utilresource.SortForCreate(objs)
	for _, o := range objs {
		err = k8sClient.Create(context.TODO(), &o)
		if err != nil {
			return err
		}
	}
	return nil
}
