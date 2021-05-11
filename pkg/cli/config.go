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
	"encoding/json"
	"flag"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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

func getConfigFrom(providers []Provider, name string) *apiextensionsv1.JSON {
	for _, p := range providers {
		if p.Name == name {
			byt, _ := json.Marshal(p.Configuration) // nolint
			return &apiextensionsv1.JSON{Raw: byt}
		}
	}
	return &apiextensionsv1.JSON{}
}
