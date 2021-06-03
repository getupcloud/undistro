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
	"flag"
	"path/filepath"

	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/util"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ConfigFlags struct {
	ConfigFile *string
	*genericclioptions.ConfigFlags
}

// NewConfigFlags returns ConfigFlags with default values set
func NewConfigFlags() *ConfigFlags {
	return &ConfigFlags{
		ConfigFile:  stringptr(""),
		ConfigFlags: genericclioptions.NewConfigFlags(true),
	}
}

func (f *ConfigFlags) AddFlags(flags *pflag.FlagSet, goflags *flag.FlagSet) {
	klog.InitFlags(goflags)
	flags.AddGoFlagSet(goflags)
	if f.ConfigFile != nil {
		flags.StringVar(f.ConfigFile, "config", *f.ConfigFile, "Path to undistro configuration (default is `$HOME/.undistro/undistro.yaml`)")
	}
	f.ConfigFlags.AddFlags(flags)
}

func (f *ConfigFlags) Init() func() {
	return func() {
		home := homedir.HomeDir()
		cfgPath := filepath.Join(home, ".undistro", "undistro.yaml")
		if *f.ConfigFile != "" {
			viper.SetConfigFile(*f.ConfigFile)
		} else {
			viper.SetConfigFile(cfgPath)
		}
		viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err == nil {
			log.Log.Info("using config", "file", viper.ConfigFileUsed())
		}
	}
}

func stringptr(s string) *string {
	return &s
}

type Credentials struct {
	Username string `mapstructure:"username" json:"username,omitempty"`
	Password string `mapstructure:"password" json:"password,omitempty"`
}

type Provider struct {
	Name          string                 `mapstructure:"name" json:"name,omitempty"`
	Configuration map[string]interface{} `mapstructure:"configuration" json:"configuration,omitempty"`
}

type Config struct {
	Credentials   Credentials `mapstructure:"credentials" json:"credentials,omitempty"`
	CoreProviders []Provider  `mapstructure:"coreProviders" json:"coreProviders,omitempty"`
	Providers     []Provider  `mapstructure:"providers" json:"providers,omitempty"`
}

func defaultValues(ctx context.Context, c client.Client, name string) map[string]interface{} {
	isKind, _ := util.IsKindCluster(ctx, c)
	switch name {
	case "undistro":
		if isKind {
			return map[string]interface{}{
				"local": true,
			}
		}
	case "ingress-nginx":
		if isKind {
			return map[string]interface{}{
				"controller": map[string]interface{}{
					"hostPort": map[string]interface{}{
						"enabled": true,
					},
					"service": map[string]interface{}{
						"type": "NodePort",
					},
					"tolerations": []map[string]interface{}{
						{
							"effect":   "NoSchedule",
							"key":      meta.LabelK8sMaster,
							"operator": "Equal",
						},
					},
				},
			}
		}
	}
	return make(map[string]interface{})
}

func getConfigFrom(ctx context.Context, c client.Client, providers []Provider, name string) map[string]interface{} {
	m := defaultValues(ctx, c, name)
	cfg := make(map[string]interface{})
	for _, p := range providers {
		if p.Name == name {
			cfg = p.Configuration
			break
		}
	}
	return util.MergeMaps(m, cfg)
}
