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
	"github.com/getupcloud/undistro/internal/util"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	utilresource "sigs.k8s.io/cluster-api/util/resource"
	"sigs.k8s.io/cluster-api/util/yaml"
)

type deleteClusterOptions struct {
	url        string
	kubeconfig string
}

var ddOpts = &deleteClusterOptions{}

var deleteClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Delete a cluster",
	Long:  "Delete a cluster",

	Example: Examples(`
		# Deletes a cluster with variable values using
		a template from a specific URL.
		undistro delete cluster --from https://github.com/foo-org/foo-repository/blob/master/cluster-template.yaml

		# Deletes a cluster with variable values using
		a template stored locally.
		undistro delete cluster --from ~/workspace/cluster-template.yaml

		# Deletes a cluster from template passed in via stdin
		cat ~/workspace/cluster-template.yaml | undistro delete cluster
`),

	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteCluster(os.Stdin, os.Stdout)
	},
}

func init() {
	// flags for the url source
	deleteClusterCmd.Flags().StringVarP(&ddOpts.url, "from", "f", "-",
		"The URL to read the template from. It defaults to '-' which reads from stdin.")
	deleteClusterCmd.Flags().StringVar(&ddOpts.kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig file to use for the management cluster. If empty, default discovery rules apply.")

	deleteCmd.AddCommand(deleteClusterCmd)
}

func deleteCluster(r io.Reader, w io.Writer) error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}
	options := client.ProcessYAMLOptions{
		ListVariablesOnly: false,
	}

	options.URLSource = &client.URLSourceOptions{
		URL: ddOpts.url,
	}
	if ddOpts.url == "-" {
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
	nm := types.NamespacedName{}
	objs = util.ReverseObjs(utilresource.SortForCreate(objs))
	for _, o := range objs {
		if o.GetNamespace() == "" {
			o.SetNamespace("default")
		}
		if o.GetKind() == "Cluster" && o.GetAPIVersion() == undistrov1.GroupVersion.String() {
			nm.Name = o.GetName()
			nm.Namespace = o.GetNamespace()
		}
		err = k8sClient.Delete(context.Background(), &o)
		if err != nil {
			return err
		}
		fmt.Printf("%s.%s %q deleted", strings.ToLower(o.GetKind()), o.GetObjectKind().GroupVersionKind().Group, o.GetName())
	}
	logStreamer, err := c.GetLogs(client.Kubeconfig{
		Path: ddOpts.kubeconfig,
	})
	if err != nil {
		return err
	}
	cfg, err := proxy.GetConfig()
	if err != nil {
		return err
	}
	if nm.Namespace == "" {
		nm.Namespace = "default"
	}
	err = logStreamer.Stream(context.Background(), cfg, os.Stdout, nm, cluster.IsDeleted)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "\n\nCluster %s is deleted.\n\n", nm.String())
	return nil
}
