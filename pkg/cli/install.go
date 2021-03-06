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
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/capi"
	"github.com/getupio-undistro/undistro/pkg/certmanager"
	"github.com/getupio-undistro/undistro/pkg/cloud"
	"github.com/getupio-undistro/undistro/pkg/helm"
	"github.com/getupio-undistro/undistro/pkg/internalautohttps"
	"github.com/getupio-undistro/undistro/pkg/kube"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/retry"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/getupio-undistro/undistro/pkg/undistro"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	getters = getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
	}
)

const (
	undistroRepo = "https://registry.undistro.io/chartrepo/library"
	ns           = undistro.Namespace
)

type InstallOptions struct {
	ConfigPath  string
	ClusterName string
	genericclioptions.IOStreams
}

func NewInstallOptions(streams genericclioptions.IOStreams) *InstallOptions {
	return &InstallOptions{
		IOStreams: streams,
	}
}

func (o *InstallOptions) Complete(f *ConfigFlags, cmd *cobra.Command, args []string) error {
	o.ConfigPath = *f.ConfigFile
	switch len(args) {
	case 0:
		// do nothing
	case 1:
		o.ClusterName = args[0]
	default:
		return cmdutil.UsageErrorf(cmd, "%s", "too many arguments")
	}
	return nil
}

func (o *InstallOptions) Validate() error {
	if o.ConfigPath != "" {
		_, err := os.Stat(o.ConfigPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *InstallOptions) installProviders(ctx context.Context, streams genericclioptions.IOStreams, c client.Client, providers []Provider, indexFile *repo.IndexFile, secretRef *corev1.LocalObjectReference) error {
	for _, p := range providers {
		chart := fmt.Sprintf("undistro-%s", p.Name)
		versions := indexFile.Entries[chart]
		if versions.Len() == 0 {
			return errors.Errorf("chart %s not found", chart)
		}
		version := versions[0]
		secretName := fmt.Sprintf("%s-config", chart)
		fmt.Fprintf(streams.Out, "Installing provider %s version %s\n", p.Name, version.AppVersion)
		fmt.Fprintf(streams.Out, "Installing provider %s required tools\n", p.Name)
		err := cloud.InstallTools(ctx, streams, p.Name)
		if err != nil {
			return errors.Errorf("unable to install required tools for provider %s: %v", p.Name, err)
		}
		secretData := make(map[string][]byte)
		valuesRef := make([]appv1alpha1.ValuesReference, 0)
		for k, v := range p.Configuration {
			str, ok := v.(string)
			if !ok {
				continue
			}
			secretData[k] = []byte(str)
			valuesRef = append(valuesRef, appv1alpha1.ValuesReference{
				Kind:       "Secret",
				Name:       secretName,
				ValuesKey:  k,
				TargetPath: k,
			})
		}
		s := corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ns,
			},
			Data: secretData,
		}
		hasDiff, err := util.CreateOrUpdate(ctx, c, &s)
		if err != nil {
			return err
		}
		provider := configv1alpha1.Provider{
			TypeMeta: metav1.TypeMeta{
				APIVersion: configv1alpha1.GroupVersion.String(),
				Kind:       "Provider",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.Name,
				Namespace: ns,
			},
			Spec: configv1alpha1.ProviderSpec{
				ProviderName:      chart,
				ProviderVersion:   version.Version,
				ProviderType:      string(configv1alpha1.InfraProviderType),
				ConfigurationFrom: valuesRef,
				Repository: configv1alpha1.Repository{
					SecretRef: secretRef,
				},
			},
		}
		if hasDiff {
			provider, err = cloud.Init(ctx, c, provider)
			if err != nil {
				return err
			}
		}
		_, err = util.CreateOrUpdate(ctx, c, &provider)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *InstallOptions) installChart(restGetter genericclioptions.RESTClientGetter, chartRepo *helm.ChartRepository, secretRef *corev1.LocalObjectReference, chartName string, overrideValuesMap map[string]interface{}) (*configv1alpha1.Provider, error) {
	overrideValues := &apiextensionsv1.JSON{}
	if len(overrideValuesMap) > 0 {
		byt, err := json.Marshal(overrideValuesMap)
		if err != nil {
			return nil, err
		}
		overrideValues.Raw = byt
	}
	versions := chartRepo.Index.Entries[chartName]
	if versions.Len() == 0 {
		return nil, errors.Errorf("chart %s not found", chartName)
	}
	version := versions[0]
	ch, err := chartRepo.Get(chartName, version.Version)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(o.IOStreams.Out, "Downloading required resources to perform %s installation\n", chartName)
	res, err := chartRepo.DownloadChart(ch)
	if err != nil {
		return nil, err
	}
	chart, err := loader.LoadArchive(res)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(o.IOStreams.Out, "Installing %s version %s\n", chart.Name(), chart.AppVersion())
	for _, dep := range chart.Dependencies() {
		fmt.Fprintf(o.IOStreams.Out, "Installing %s version %s\n", dep.Name(), dep.AppVersion())
	}
	p := configv1alpha1.Provider{
		TypeMeta: metav1.TypeMeta{
			APIVersion: configv1alpha1.GroupVersion.String(),
			Kind:       "Provider",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      chartName,
			Namespace: ns,
			Labels: map[string]string{
				meta.LabelProviderType: string(configv1alpha1.CoreProviderType),
			},
		},
		Spec: configv1alpha1.ProviderSpec{
			ProviderName:    chartName,
			ProviderVersion: version.Version,
			ProviderType:    string(configv1alpha1.CoreProviderType),
			Repository: configv1alpha1.Repository{
				SecretRef: secretRef,
			},
			Configuration: overrideValues,
		},
	}
	err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
		runner, err := helm.NewRunner(restGetter, ns, log.Log)
		if err != nil {
			return err
		}
		wait := true
		forceUpgrade := true
		reset := false
		history := 0
		hr := appv1alpha1.HelmRelease{
			Spec: appv1alpha1.HelmReleaseSpec{
				ReleaseName:     chartName,
				TargetNamespace: ns,
				Wait:            &wait,
				MaxHistory:      &history,
				ResetValues:     &reset,
				ForceUpgrade:    &forceUpgrade,
				Timeout: &metav1.Duration{
					Duration: 1 * time.Minute,
				},
			},
		}
		m := make(map[string]interface{})
		if overrideValues.Raw != nil {
			err = json.Unmarshal(overrideValues.Raw, &m)
			if err != nil {
				return err
			}
		}
		chart.Values = util.MergeMaps(chart.Values, m)
		rel, _ := runner.ObserveLastRelease(hr)
		if rel == nil {
			_, err = runner.Install(hr, chart, chart.Values)
			if err != nil {
				return err
			}
		} else if rel.Info.Status == release.StatusDeployed {
			_, err = runner.Upgrade(hr, chart, chart.Values)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return &p, err
}

func (o *InstallOptions) RunInstall(f cmdutil.Factory, cmd *cobra.Command) error {
	cfg := Config{}
	if o.ConfigPath != "" {
		err := viper.Unmarshal(&cfg)
		if err != nil {
			return errors.Errorf("unable to unmarshal config: %v", err)
		}
	}
	restCfg, err := f.ToRESTConfig()
	if err != nil {
		return err
	}
	secretName := "undistro-config"
	c, err := client.New(restCfg, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return err
	}
	restGetter := kube.NewInClusterRESTClientGetter(restCfg, ns)
	if o.ClusterName != "" {
		byt, err := kubeconfig.FromSecret(cmd.Context(), c, util.ObjectKeyFromString(o.ClusterName))
		if err != nil {
			return err
		}
		restGetter = kube.NewMemoryRESTClientGetter(byt, ns)
		restCfg, err = restGetter.ToRESTConfig()
		if err != nil {
			return err
		}
		c, err = client.New(restCfg, client.Options{
			Scheme: scheme.Scheme,
		})
		if err != nil {
			return err
		}
	}
	_, err = util.CreateOrUpdate(cmd.Context(), c, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	})
	if err != nil {
		return err
	}
	var clientOpts []getter.Option
	var secretRef *corev1.LocalObjectReference
	if cfg.Credentials.Username != "" && cfg.Credentials.Password != "" {
		secretRef = &corev1.LocalObjectReference{
			Name: secretName,
		}
		userb64 := base64.StdEncoding.EncodeToString([]byte(cfg.Credentials.Username))
		passb64 := base64.StdEncoding.EncodeToString([]byte(cfg.Credentials.Password))
		s := corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ns,
			},
			Data: map[string][]byte{
				"username": []byte(userb64),
				"password": []byte(passb64),
			},
		}
		_, err = util.CreateOrUpdate(cmd.Context(), c, &s)
		if err != nil {
			return err
		}
		opts, cleanup, err := helm.ClientOptionsFromSecret(s)
		if err != nil {
			return err
		}
		defer cleanup()
		clientOpts = opts
	}
	chartRepo, err := helm.NewChartRepository(undistroRepo, getters, clientOpts)
	if err != nil {
		return err
	}
	fmt.Fprintf(o.IOStreams.Out, "Downloading repository index\n")
	err = chartRepo.DownloadIndex()
	if err != nil {
		return errors.Wrap(err, "failed to download repository index")
	}
	fmt.Fprintf(o.IOStreams.Out, "Ensure cert-manager is installed\n")
	certObjs, err := util.ToUnstructured([]byte(certmanager.TestResources))
	if err != nil {
		return err
	}
	installCert := false
	for _, o := range certObjs {
		_, err = util.CreateOrUpdate(cmd.Context(), c, &o)
		if err != nil {
			installCert = true
			break
		}
	}
	providers := make([]*configv1alpha1.Provider, 0)
	if installCert {
		provider, err := o.installChart(restGetter, chartRepo, secretRef, "cert-manager", getConfigFrom(cmd.Context(), c, cfg.CoreProviders, "cert-manager"))
		if err != nil {
			return err
		}
		providers = append(providers, provider)
		err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			for _, o := range certObjs {
				_, errCert := util.CreateOrUpdate(cmd.Context(), c, &o)
				if errCert != nil {
					return errCert
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	installCapi := false
	capiObjs, err := util.ToUnstructured([]byte(capi.TestResources))
	if err != nil {
		return err
	}
	for _, o := range capiObjs {
		_, errCapi := util.CreateOrUpdate(cmd.Context(), c, &o)
		if errCapi != nil {
			installCapi = true
			break
		}
	}
	if installCapi {
		providerCapi, err := o.installChart(restGetter, chartRepo, secretRef, "cluster-api", getConfigFrom(cmd.Context(), c, cfg.CoreProviders, "cluster-api"))
		if err != nil {
			return err
		}
		providers = append(providers, providerCapi)
		err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			for _, o := range capiObjs {
				_, errCert := util.CreateOrUpdate(cmd.Context(), c, &o)
				if errCert != nil {
					return errCert
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	installUndistro := false
	undistroObjs, err := util.ToUnstructured([]byte(undistro.TestResources))
	if err != nil {
		return err
	}
	for _, o := range undistroObjs {
		_, errUndistro := util.CreateOrUpdate(cmd.Context(), c, &o)
		if errUndistro != nil {
			installUndistro = true
			break
		}
	}
	if installUndistro {
		undistroChartValues := getConfigFrom(cmd.Context(), c, cfg.CoreProviders, "undistro")
		providerUndistro, err := o.installChart(restGetter, chartRepo, secretRef, "undistro", undistroChartValues)
		if err != nil {
			return err
		}
		providers = append(providers, providerUndistro)

		providerNginx, err := o.installChart(restGetter, chartRepo, secretRef, "ingress-nginx", getConfigFrom(cmd.Context(), c, cfg.CoreProviders, "ingress-nginx"))
		if err != nil {
			return err
		}
		providers = append(providers, providerNginx)
		err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			for _, o := range undistroObjs {
				_, errCert := util.CreateOrUpdate(cmd.Context(), c, &o)
				if errCert != nil {
					return errCert
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		// install cert in local environments
		isLocal := undistroChartValues["local"].(bool)
		if isLocal {
			fmt.Fprintf(o.IOStreams.Out, "Installing local certificates\n")
			err = internalautohttps.InstallLocalCert(cmd.Context(), c)
			if err != nil {
				return err
			}
		}
	}
	err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
		for _, p := range providers {
			if p.Labels == nil {
				p.Labels = make(map[string]string)
			}
			p.Labels[meta.LabelProviderType] = "core"
			_, err = util.CreateOrUpdate(cmd.Context(), c, p)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
		return o.installProviders(cmd.Context(), o.IOStreams, c, cfg.Providers, chartRepo.Index, secretRef)
	})
	if err != nil {
		return err
	}

	fmt.Fprint(o.IOStreams.Out, "Waiting all providers to be ready")
	for {
		list := configv1alpha1.ProviderList{}
		err = retry.WithExponentialBackoff(retry.NewBackoff(), func() error {
			err = c.List(cmd.Context(), &list, client.InNamespace(ns))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		ready := true
		for _, item := range list.Items {
			if !meta.InReadyCondition(item.Status.Conditions) {
				ready = false
				break
			}
			err = cloud.PostInstall(cmd.Context(), c, item)
			if err != nil {
				return err
			}
		}

		// retest objects because certmanager update certs when new pod is added
		for _, o := range certObjs {
			_, err = util.CreateOrUpdate(cmd.Context(), c, &o)
			if err != nil {
				ready = false
				break
			}
		}
		for _, o := range undistroObjs {
			_, err = util.CreateOrUpdate(cmd.Context(), c, &o)
			if err != nil {
				ready = false
				break
			}
		}
		for _, o := range capiObjs {
			_, err = util.CreateOrUpdate(cmd.Context(), c, &o)
			if err != nil {
				ready = false
				break
			}
		}
		if ready {
			fmt.Fprintln(o.IOStreams.Out, "\n\nManagement cluster is ready to use.")
			return nil
		}
		fmt.Fprint(o.IOStreams.Out, ".")
		<-time.After(15 * time.Second)
	}
}

func NewCmdInstall(f *ConfigFlags, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewInstallOptions(streams)
	cmd := &cobra.Command{
		Use:                   "install [cluster namespace/cluster name]",
		DisableFlagsInUseLine: true,
		Short:                 "Install UnDistro",
		Long: LongDesc(`Install UnDistro.
		If cluster argument exists UnDistro will be installed in this remote cluster.
		If config file exists UnDistro will be installed using file's configurations`),
		Example: Examples(`
		# Install UnDistro in local cluster
		undistro install
		# Install UnDistro in remote cluster
		undistro install undistro-production/cool-product-cluster
		# Install UnDistro with configuration file
		undistro --config undistro-config.yaml install undistro-production/cool-product-cluster
		`),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			err := o.RunInstall(cmdutil.NewFactory(f), cmd)
			if err != nil {
				fmt.Fprintf(o.ErrOut, "error: %v\n", err)
			}
		},
	}
	return cmd
}
