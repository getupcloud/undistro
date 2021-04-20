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

	"github.com/getupio-undistro/undistro/pkg/cloud"
	"github.com/getupio-undistro/undistro/pkg/fs"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/getupio-undistro/undistro/pkg/template"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/create"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type ClusterOptions struct {
	genericclioptions.IOStreams
	Namespace    string
	ClusterName  string
	Infra        string
	Flavor       string
	SshKeyName   string
	GenerateFile bool
	K8sVersion   string
	Region       string
}

func NewClusterOptions(streams genericclioptions.IOStreams) *ClusterOptions {
	return &ClusterOptions{
		IOStreams:  streams,
		K8sVersion: "v1.19.8",
	}
}

func (o *ClusterOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error
	o.Namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}
	if o.Namespace == "" {
		o.Namespace = "dafault"
	}
	if len(args) != 1 {
		return errors.New("required 1 argument")
	}
	if o.Infra == "" {
		return errors.New("required flag: infra")
	}
	if o.Flavor == "" {
		return errors.New("required flag: flavor")
	}
	err = o.validateInfraFlavor()
	if err != nil {
		return err
	}
	o.ClusterName = args[0]
	return nil
}

func (o *ClusterOptions) validateInfraFlavor() error {
	switch o.Infra {
	case "aws":
		switch o.Flavor {
		case "ec2", "eks":
			return nil
		default:
			return errors.Errorf("unknown flavor: %s", o.Flavor)
		}
	default:
		return errors.Errorf("unknown infrastructure: %s", o.Infra)
	}
}

func (o *ClusterOptions) setRegionByInfra(ctx context.Context, c client.Client) error {
	sname := fmt.Sprintf("undistro-%s-config", o.Infra)
	key := client.ObjectKey{
		Name:      sname,
		Namespace: ns,
	}
	s := corev1.Secret{}
	err := c.Get(ctx, key, &s)
	if err != nil {
		return err
	}
	byt, ok := s.Data["region"]
	if !ok {
		byt = []byte(cloud.DefaultRegion(o.Infra))
	}
	o.Region = string(byt)
	if o.Region == "" {
		return errors.New("default region is not set. Use --region to set a region to create a cluster")
	}
	return nil
}

func (o *ClusterOptions) RunCreateCluster(f cmdutil.Factory, cmd *cobra.Command) error {
	cfg, err := f.ToRESTConfig()
	if err != nil {
		return errors.Errorf("unable to get kubeconfig: %v", err)
	}
	c, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return errors.Errorf("unable to create client: %v", err)
	}
	if o.Region == "" {
		err = o.setRegionByInfra(cmd.Context(), c)
		if err != nil {
			return err
		}
	}
	vars := map[string]interface{}{
		"Flavor":     o.Flavor,
		"SSHKey":     o.SshKeyName,
		"Namespace":  o.Namespace,
		"Name":       o.ClusterName,
		"K8sVersion": o.K8sVersion,
		"Region":     o.Region,
	}
	objs, err := template.GetObjs(fs.DefaultArchFS, "defaultarch", o.Infra, vars)
	if err != nil {
		return err
	}
	for _, o := range objs {
		_, err = util.CreateOrUpdate(cmd.Context(), c, &o)
		if err != nil {
			return err
		}
	}
	if o.GenerateFile {
		f, err := os.Create(fmt.Sprintf("%s.yaml", o.ClusterName))
		if err != nil {
			return err
		}
		defer f.Close()
		for _, o := range objs {
			f.WriteString("---\n")
			byt, err := yaml.Marshal(o.Object)
			if err != nil {
				return err
			}
			f.Write(byt)
		}
	}
	return nil
}

func (o *ClusterOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Infra, "infra", o.Infra, "the infrastructure where cluster will be created")
	flags.StringVar(&o.Flavor, "flavor", o.Flavor, "the flavor used to create cluster in selected infrastructure")
	flags.StringVar(&o.SshKeyName, "ssh-key-name", o.ClusterName, "the name of SSH Key in provider")
	flags.StringVar(&o.K8sVersion, "k8s-version", o.K8sVersion, "the Kubernetes version (default v1.19.8)")
	flags.StringVar(&o.Region, "region", o.Region, "the region where cluster will be created")
	flags.BoolVar(&o.GenerateFile, "generate-file", o.GenerateFile, "Generate cluster YAML file")
}

func NewCmdCluster(f cmdutil.Factory, flags *pflag.FlagSet, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewClusterOptions(streams)
	o.AddFlags(flags)
	cmd := &cobra.Command{
		Use:                   "cluster [cluster name]",
		DisableFlagsInUseLine: true,
		Short:                 "Create a cluster based on recommended spec",
		Long:                  LongDesc(`Create a cluster based on spec recommend by Getup`),
		Example: Examples(`
		undistro create cluster cool-cluster -n cool-namespace --infra aws --flavor ec2
		`),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.RunCreateCluster(f, cmd))
		},
	}
	return cmd
}

func NewCmdCreate(f cmdutil.Factory, flags *pflag.FlagSet, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := create.NewCmdCreate(f, streams)
	cmd.AddCommand(NewCmdCluster(f, flags, streams))
	return cmd
}
