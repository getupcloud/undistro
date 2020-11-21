package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	proxy, err := c.GetProxy(client.Kubeconfig{
		Path: ccOpts.kubeconfig,
	})
	if err != nil {
		return err
	}
	k8sClient, err := proxy.NewClient()
	if err != nil {
		return err
	}
	cfg, err := proxy.GetConfig()
	if err != nil {
		return err
	}
	objs = utilresource.SortForCreate(objs)
	nm := types.NamespacedName{}
	var (
		dd       time.Time
		objNames = make([]string, 0)
	)
	for _, o := range objs {
		if o.GetNamespace() == "" {
			o.SetNamespace("default")
		}
		if o.GetKind() == "Cluster" && o.GetAPIVersion() == undistrov1.GroupVersion.String() {
			nm.Name = o.GetName()
			nm.Namespace = o.GetNamespace()
			cl := undistrov1.Cluster{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, &cl)
			if err != nil {
				return err
			}
			objNames = append(objNames, cl.Name)
			for i := 0; i < len(cl.Spec.WorkerNodes); i++ {
				objNames = append(objNames, fmt.Sprintf("%s-mp-%d", cl.Name, i))
			}
			dd = time.Now()
		}
		err = k8sClient.Create(context.Background(), &o)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "%s.%s %q created\n", strings.ToLower(o.GetKind()), o.GetObjectKind().GroupVersionKind().Group, o.GetName())
	}
	listener, err := c.GetEventListener(client.Kubeconfig{
		Path: ccOpts.kubeconfig,
	})
	if err != nil {
		return err
	}
	watchch, err := listener.Listen(context.Background(), cfg, objNames...)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, color.GreenString("\u2714 %s", "Cluster creation started"))
	for e := range watchch.ResultChan() {
		ev, ok := e.Object.(*corev1.Event)
		if !ok {
			return errors.New("not an event")
		}
		if ev.GetCreationTimestamp().After(dd) && ev.Count == 1 {
			switch ev.Type {
			case corev1.EventTypeNormal:
				fmt.Fprintln(os.Stdout, color.GreenString("\u2714 %s", ev.Message))
			case corev1.EventTypeWarning:
				fmt.Fprintln(os.Stdout, color.RedString("\u271d %s", ev.Message))
			}
			if ev.Reason == "ClusterReady" {
				watchch.Stop()
			}
		}
	}
	fmt.Fprintf(os.Stdout, "\n\nCluster %s is ready. Run command below to get the Kubeconfig\n\nundistro get kubeconfig %s -n %s\n\n", nm.String(), nm.Name, nm.Namespace)
	return nil
}
