/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package helm

import (
	"fmt"

	"github.com/getupio-undistro/undistro/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"
)

const VERSION = "v3"

var (
	repositoryConfig = helmpath.ConfigPath("repositories.yaml")
	repositoryCache  = helmpath.CachePath("repository")
	pluginsDir       = helmpath.DataPath("plugins")
)

type HelmOptions struct {
	Driver    string
	Namespace string
}

type HelmV3 struct {
	path string
}

type infoLogFunc func(string, ...interface{})

// New creates a new HelmV3 client
func New(path string) Client {
	return &HelmV3{
		path: path,
	}
}

func (h *HelmV3) Version() string {
	return VERSION
}

// infoLogFunc allows us to pass our logger to components
// that expect a klog.Infof function.
func (h *HelmV3) infoLogFunc(namespace string, releaseName string) infoLogFunc {
	return func(format string, args ...interface{}) {
		message := fmt.Sprintf(format, args...)
		log.Log.Info(message, "targetNamespace", namespace, "release", releaseName)
	}
}

func newActionConfig(path string, logFunc infoLogFunc, namespace, driver string) (*action.Configuration, error) {

	restClientGetter := newConfigFlags(path, namespace)
	kubeClient := &kube.Client{
		Factory: util.NewFactory(restClientGetter),
		Log:     logFunc,
	}
	client, err := kubeClient.Factory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	store, err := newStorageDriver(client, logFunc, namespace, driver)
	if err != nil {
		return nil, err
	}

	return &action.Configuration{
		RESTClientGetter: restClientGetter,
		Releases:         store,
		KubeClient:       kubeClient,
		Log:              logFunc,
	}, nil
}

func newConfigFlags(path, namespace string) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		KubeConfig: &path,
		Namespace:  &namespace,
	}
}

func newStorageDriver(client *kubernetes.Clientset, logFunc infoLogFunc, namespace, d string) (*storage.Storage, error) {
	switch d {
	case "secret", "secrets", "":
		s := driver.NewSecrets(client.CoreV1().Secrets(namespace))
		s.Log = logFunc
		return storage.Init(s), nil
	case "configmap", "configmaps":
		c := driver.NewConfigMaps(client.CoreV1().ConfigMaps(namespace))
		c.Log = logFunc
		return storage.Init(c), nil
	case "memory":
		m := driver.NewMemory()
		return storage.Init(m), nil
	default:
		return nil, fmt.Errorf("unsupported storage driver '%s'", d)
	}
}

func getterProviders() getter.Providers {
	return getter.All(&cli.EnvSettings{
		RepositoryConfig: repositoryConfig,
		RepositoryCache:  repositoryCache,
		PluginsDirectory: pluginsDir,
	})
}
