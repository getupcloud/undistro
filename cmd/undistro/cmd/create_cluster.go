package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/client/cluster"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	utilresource "sigs.k8s.io/cluster-api/util/resource"
	"sigs.k8s.io/cluster-api/util/yaml"
)

type createClusterOptions struct {
	url        string
	kubeconfig string
}

var ccOpts = &createClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:   "cluster",
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
	createClusterCmd.Flags().StringVarP(&ccOpts.url, "from", "f", "-",
		"The URL to read the template from. It defaults to '-' which reads from stdin.")
	createClusterCmd.Flags().StringVar(&ccOpts.kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig file to use for the management cluster. If empty, default discovery rules apply.")

	createCmd.AddCommand(createClusterCmd)
}

func createCluster(r io.Reader, w io.Writer) error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}
	options := client.ProcessYAMLOptions{
		ListVariablesOnly: false,
	}

	options.URLSource = &client.URLSourceOptions{
		URL: ccOpts.url,
	}
	if ccOpts.url == "-" {
		options.ReaderSource = &client.ReaderSourceOptions{
			Reader: r,
		}
		options.URLSource = nil
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
	nm := types.NamespacedName{}
	for _, o := range objs {
		if o.GetNamespace() == "" {
			o.SetNamespace("default")
		}
		if o.GetKind() == "Cluster" && o.GetAPIVersion() == undistrov1.GroupVersion.String() {
			nm.Name = o.GetName()
			nm.Namespace = o.GetNamespace()
		}
		err = k8sClient.Create(context.Background(), &o)
		if err != nil {
			return err
		}
		fmt.Printf("%s.%s %q created", strings.ToLower(o.GetKind()), o.GetObjectKind().GroupVersionKind().Group, o.GetName())
	}
	logStreamer, err := c.GetLogs(client.Kubeconfig{
		Path: ccOpts.kubeconfig,
	})
	if err != nil {
		return err
	}
	cfg, err := proxy.GetConfig()
	if err != nil {
		return err
	}
	err = logStreamer.Stream(context.Background(), cfg, os.Stdout, nm, cluster.IsReady)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "\n\nCluster %s is ready. \nRun undistro get kubeconfig %s -n %s to get the Kubeconfig\n\n", nm.String(), nm.Name, nm.Namespace)
	return nil
}
