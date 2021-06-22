/*
Copyright 2020-2021 The UnDistro authors

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
	"fmt"

	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/helm"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/getter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type UpgradeOptions struct {
	genericclioptions.IOStreams
	Version      string
	ProviderName string
}

func NewUpgradeOptions(streams genericclioptions.IOStreams) *UpgradeOptions {
	return &UpgradeOptions{
		IOStreams: streams,
	}
}

func (o *UpgradeOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Version, "version", "", "provider version to upgrade")
}

func (o *UpgradeOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("required 1 argument")
	}
	o.ProviderName = args[0]
	return nil
}

func (o *UpgradeOptions) RunUpgrade(f cmdutil.Factory, cmd *cobra.Command) error {
	var clientOpts []getter.Option
	cfg, err := f.ToRESTConfig()
	if err != nil {
		return errors.Errorf("unable to get config: %v", err)
	}
	workloadClient, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return errors.Errorf("unable to create client: %v", err)
	}
	key := client.ObjectKey{
		Name:      o.ProviderName,
		Namespace: ns,
	}
	p := configv1alpha1.Provider{}
	err = workloadClient.Get(cmd.Context(), key, &p)
	if err != nil {
		return err
	}
	if o.Version == "" {
		chartRepo, err := helm.NewChartRepository(undistroRepo, getters, clientOpts)
		if err != nil {
			return err
		}
		p.TypeMeta = metav1.TypeMeta{
			Kind:       "Provider",
			APIVersion: configv1alpha1.GroupVersion.String(),
		}
		fmt.Fprintf(o.IOStreams.Out, "Downloading repository index\n")
		err = chartRepo.DownloadIndex()
		if err != nil {
			return errors.Wrap(err, "failed to download repository index")
		}
		versions := chartRepo.Index.Entries[o.ProviderName]
		if versions.Len() == 0 {
			return errors.Errorf("chart %s not found", o.ProviderName)
		}
		version := versions[0]
		_, err = chartRepo.Get(o.ProviderName, version.Version)
		if err != nil {
			return err
		}
		o.Version = version.Version
	}
	p.Spec.ProviderVersion = o.Version
	_, err = util.CreateOrUpdate(cmd.Context(), workloadClient, &p)
	return err
}

func NewCmdUpgrade(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewUpgradeOptions(streams)
	cmd := &cobra.Command{
		Use:                   "upgrade [provider name]",
		DisableFlagsInUseLine: true,
		Short:                 "Upgrade a provider",
		Long:                  LongDesc(`Upgrade a provider to specified version`),
		Example: Examples(`
		# Upgrade provider to specified version
		undistro upgrade undistro --version 0.1.17
		# Upgrade provider to latest version
		undistro upgrade undistro
		`),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.RunUpgrade(f, cmd))
		},
	}
	o.AddFlags(cmd.Flags())
	return cmd
}
