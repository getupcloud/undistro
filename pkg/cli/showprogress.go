/*
Copyright 2020 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/fs"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/getupio-undistro/undistro/pkg/template"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ShowProgressOptions struct {
	genericclioptions.IOStreams
	Namespace   string
	ClusterName string
}

func NewShowProgressOptions(streams genericclioptions.IOStreams) *ShowProgressOptions {
	return &ShowProgressOptions{
		IOStreams: streams,
	}
}

func (o *ShowProgressOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error
	o.Namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return errors.New("required 1 argument")
	}
	o.ClusterName = args[0]
	return nil
}

func (o *ShowProgressOptions) RunShowProgress(f cmdutil.Factory, cmd *cobra.Command) error {
	cfg, err := f.ToRESTConfig()
	if err != nil {
		return errors.Errorf("unable to get config: %v", err)
	}
	cfg.Timeout = 0
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Errorf("unable to create event client: %v", err)
	}
	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return errors.Errorf("unable to create client: %v", err)
	}
	key := client.ObjectKey{
		Name:      o.ClusterName,
		Namespace: o.Namespace,
	}
	obj := appv1alpha1.Cluster{}
	err = k8sClient.Get(cmd.Context(), key, &obj)
	if err != nil {
		return err
	}
	vars := map[string]interface{}{
		"ENV":     make(map[string]interface{}),
		"Cluster": &obj,
	}
	objs, err := template.GetObjs(fs.FS, "clustertemplates", obj.GetTemplate(), vars)
	if err != nil {
		return err
	}
	w, err := c.CoreV1().Events(o.Namespace).Watch(cmd.Context(), metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		return err
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func(ctx context.Context) {
		for e := range w.ResultChan() {
			ev, ok := e.Object.(*corev1.Event)
			if !ok {
				continue
			}
			if ev.CreationTimestamp.After(obj.GetCreationTimestamp().Time) && (ev.Reason != "" || ev.Message != "") && ev.Count == 1 {
				for _, item := range objs {
					if item.GetName() == ev.InvolvedObject.Name {
						fmt.Fprintf(o.IOStreams.Out, "Reason: %s Message: %s\n", ev.Reason, ev.Message)
						break
					}
				}
			}
		}
	}(cmd.Context())
	<-sig
	w.Stop()
	return nil
}

func NewCmdShowProgress(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewShowProgressOptions(streams)
	cmd := &cobra.Command{
		Use:                   "show-progress [cluster name]",
		DisableFlagsInUseLine: true,
		Short:                 "Show events of a clusters and child objects",
		Long:                  LongDesc(`Show events of a clusters and child objects`),
		Example: Examples(`
		# Show events of a clusters and child objects in default namespace
		undistro show-progress cool-cluster
		# Show events of a clusters and child objects in others namespace
		undistro show-progress cool-cluster -n cool-namespace
		`),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.RunShowProgress(f, cmd))
		},
	}
	return cmd
}
