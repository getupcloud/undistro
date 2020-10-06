/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	logf "github.com/getupio-undistro/undistro/log"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

const (
	// ConfigFolder defines the name of the config folder under $home
	ConfigFolder = ".undistro"
	// ConfigName defines the name of the config file under ConfigFolder
	ConfigName = "undistro"

	// ConfigFolder defines the name of the config folder under $home (capi)
	ConfigFolderCapi = ".cluster-api"
	// ConfigName defines the name of the config file under ConfigFolder (capi)
	ConfigNameCapi = "clusterctl"
)

var (
	configNames = []string{ConfigName, ConfigNameCapi}
)

// viperReader implements Reader using viper as backend for reading from environment variables
// and from a undistro config file.
type viperReader struct {
	configPaths []string
}

type viperReaderOption func(*viperReader)

func InjectConfigPaths(configPaths []string) viperReaderOption {
	return func(vr *viperReader) {
		vr.configPaths = configPaths
	}
}

// newViperReader returns a viperReader.
func newViperReader(opts ...viperReaderOption) Reader {
	vr := &viperReader{
		configPaths: []string{
			filepath.Join(homedir.HomeDir(), ConfigFolder),
			filepath.Join(homedir.HomeDir(), ConfigFolderCapi),
		},
	}
	for _, o := range opts {
		o(vr)
	}
	return vr
}

// Init initialize the viperReader.
func (v *viperReader) Init(path string) error {
	log := logf.Log

	// Configure viper for reading environment variables as well, and more specifically:
	// AutomaticEnv force viper to check for an environment variable any time a viper.Get request is made.
	// It will check for a environment variable with a name matching the key uppercased; in case name use the - delimiter,
	// the SetEnvKeyReplacer forces matching to name use the _ delimiter instead (- is not allowed in linux env variable names).
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()

	// Reads the undistro config file
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return err
		}
		// Use path file from the flag.
		viper.SetConfigFile(path)
	} else {
		// Checks if there is a default .undistro/undistro{.extension} file in home directory
		if !v.checkDefaultConfig() {
			// since there is no default config to read from, just skip
			// reading in config
			log.V(5).Info("No default config file available")
			return nil
		}
		// Configure viper for reading .undistro/undistro{.extension} in home directory
		viper.SetConfigName(ConfigName)
		for _, p := range v.configPaths {
			viper.AddConfigPath(p)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	log.V(5).Info("Using configuration", "File", viper.ConfigFileUsed())
	return nil
}

func (v *viperReader) Get(key string) (string, error) {
	if viper.Get(key) == nil {
		return "", errors.Errorf("Failed to get value for variable %q. Please set the variable value using os env variables or using the .undistro config file", key)
	}
	return viper.GetString(key), nil
}

func (v *viperReader) Set(key, value string) {
	viper.Set(key, value)
}

func (v *viperReader) UnmarshalKey(key string, rawval interface{}) error {
	return viper.UnmarshalKey(key, rawval)
}

// checkDefaultConfig checks the existence of the default config.
// Returns true if it finds a supported config file in the available config
// folders.
func (v *viperReader) checkDefaultConfig() bool {
	for _, path := range v.configPaths {
		for _, ext := range viper.SupportedExts {
			for _, cfgName := range configNames {
				f := filepath.Join(path, fmt.Sprintf("%s.%s", cfgName, ext))
				_, err := os.Stat(f)
				if err == nil {
					return true
				}
			}
		}
	}
	return false
}
