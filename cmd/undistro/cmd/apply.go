package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/getupcloud/undistro/client"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	utilresource "sigs.k8s.io/cluster-api/util/resource"
	"sigs.k8s.io/cluster-api/util/yaml"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type applyClusterOptions struct {
	url        string
	kubeconfig string
}

var applyOpts = &applyClusterOptions{}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply change",
	Long:  "Apply change",

	Example: Examples(`
		undistro apply --from https://github.com/foo-org/foo-repository/blob/master/cluster-template.yaml

		undistro apply --from ~/workspace/cluster-template.yaml

		cat ~/workspace/cluster-template.yaml | undistro apply
`),

	RunE: func(cmd *cobra.Command, args []string) error {
		return applyCluster(os.Stdin, os.Stdout)
	},
}

func init() {
	// flags for the url source
	applyCmd.Flags().StringVarP(&applyOpts.url, "from", "f", "-",
		"The URL to read the template from. It defaults to '-' which reads from stdin.")
	applyCmd.Flags().StringVar(&applyOpts.kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig file to use for the management cluster. If empty, default discovery rules apply.")

	RootCmd.AddCommand(applyCmd)
}

func applyCluster(r io.Reader, w io.Writer) error {
	c, err := client.New(cfgFile)
	if err != nil {
		return err
	}
	options := client.ProcessYAMLOptions{
		ListVariablesOnly: false,
	}

	options.URLSource = &client.URLSourceOptions{
		URL: applyOpts.url,
	}
	if applyOpts.url == "-" {
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
	for _, o := range objs {
		if o.GetNamespace() == "" {
			o.SetNamespace("default")
		}
		nm := types.NamespacedName{
			Name:      o.GetName(),
			Namespace: o.GetNamespace(),
		}
		old := unstructured.Unstructured{}
		old.SetGroupVersionKind(o.GroupVersionKind())
		err = k8sClient.Get(context.Background(), nm, &old)
		if err != nil {
			return err
		}
		o.SetResourceVersion(old.GetResourceVersion())
		err = k8sClient.Patch(context.Background(), &o, ctrlclient.MergeFrom(&old))
		if err != nil {
			return err
		}
		fmt.Printf("%s.%s %q applied\n", strings.ToLower(o.GetKind()), o.GetObjectKind().GroupVersionKind().Group, o.GetName())
	}
	return nil
}
