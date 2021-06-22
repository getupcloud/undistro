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
package test

import (
	"context"
	"io/ioutil"
	"net"
	"path"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	configv1alpha1 "github.com/getupio-undistro/undistro/apis/config/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	"github.com/getupio-undistro/undistro/pkg/util"
	"github.com/onsi/ginkgo"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	klog.InitFlags(nil)
	logger := klogr.New()
	// additionally force all of the controllers to use the Ginkgo logger.
	ctrl.SetLogger(logger)
	// add logger for ginkgo
	klog.SetOutput(ginkgo.GinkgoWriter)
}

var (
	env *envtest.Environment
)

func init() {
	// Get the root of the current file to use in CRD paths.
	_, filename, _, _ := goruntime.Caller(0) //nolint
	root := path.Join(path.Dir(filename), "..", "..")

	// Create the test environment.
	env = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join(root, "config", "crd", "bases"),
			filepath.Join(root, "charts", "cluster-api", "crds"),
		},
	}
}

// Environment encapsulates a Kubernetes local test environment.
type Environment struct {
	manager.Manager
	client.Client
	Config *rest.Config
	cancel context.CancelFunc
}

func NewEnvironment() *Environment {
	// initialize webhook here to be able to test the envtest install via webhookOptions
	// This should set LocalServingCertDir and LocalServingPort that are used below.
	initializeWebhookInEnvironment()
	if _, err := env.Start(); err != nil {
		err = kerrors.NewAggregate([]error{err, env.Stop()})
		panic(err)
	}
	options := manager.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
		CertDir:            env.WebhookInstallOptions.LocalServingCertDir,
		Port:               env.WebhookInstallOptions.LocalServingPort,
	}
	mgr, err := ctrl.NewManager(env.Config, options)
	if err != nil {
		klog.Fatalf("Failed to start testenv manager: %v", err)
	}
	if err = (&configv1alpha1.Provider{}).SetupWebhookWithManager(mgr); err != nil {
		klog.Fatalln(err, "unable to create webhook", "webhook", "Provider")
	}
	if err = (&appv1alpha1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		klog.Fatalln(err, "unable to create webhook", "webhook", "Cluster")
	}
	if err = (&appv1alpha1.HelmRelease{}).SetupWebhookWithManager(mgr); err != nil {
		klog.Fatalln(err, "unable to create webhook", "webhook", "HelmRelease")
	}
	return &Environment{
		Manager: mgr,
		Client:  mgr.GetClient(),
		Config:  mgr.GetConfig(),
	}
}

func (t *Environment) StartManager(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel
	return t.Manager.Start(ctx)
}

func (t *Environment) WaitForWebhooks() {
	port := env.WebhookInstallOptions.LocalServingPort

	klog.V(2).Infof("Waiting for webhook port %d to be open prior to running tests", port)
	timeout := 1 * time.Second
	for {
		time.Sleep(1 * time.Second)
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), timeout)
		if err != nil {
			klog.V(2).Infof("Webhook port is not ready, will retry in %v: %s", timeout, err)
			continue
		}
		conn.Close()
		klog.V(2).Info("Webhook port is now open. Continuing with tests...")
		return
	}
}

func (t *Environment) Stop() error {
	t.cancel()
	return env.Stop()
}

func initializeWebhookInEnvironment() {
	validatingWebhooks := []client.Object{}
	mutatingWebhooks := []client.Object{}

	// Get the root of the current file to use in CRD paths.
	_, filename, _, _ := goruntime.Caller(0) //nolint
	root := path.Join(path.Dir(filename), "..", "..")
	configyamlFile, err := ioutil.ReadFile(filepath.Join(root, "config", "webhook", "manifests.yaml"))
	if err != nil {

		klog.Fatalf("Failed to read core webhook configuration file: %v ", err)
	}
	if err != nil {
		klog.Fatalf("failed to parse yaml")
	}
	// append the webhook with suffix to avoid clashing webhooks. repeated for every webhook
	mutatingWebhooks, validatingWebhooks, err = appendWebhookConfiguration(mutatingWebhooks, validatingWebhooks, configyamlFile, "config")
	if err != nil {
		klog.Fatalf("Failed to append core controller webhook config: %v", err)
	}

	env.WebhookInstallOptions = envtest.WebhookInstallOptions{
		MaxTime:            20 * time.Second,
		PollInterval:       time.Second,
		ValidatingWebhooks: validatingWebhooks,
		MutatingWebhooks:   mutatingWebhooks,
	}
}

const (
	mutatingWebhookKind   = "MutatingWebhookConfiguration"
	validatingWebhookKind = "ValidatingWebhookConfiguration"
	mutatingwebhook       = "mutating-webhook-configuration"
	validatingwebhook     = "validating-webhook-configuration"
)

// Mutate the name of each webhook, because kubebuilder generates the same name for all controllers.
// In normal usage, kustomize will prefix the controller name, which we have to do manually here.
func appendWebhookConfiguration(mutatingWebhooks []client.Object, validatingWebhooks []client.Object, configyamlFile []byte, tag string) ([]client.Object, []client.Object, error) {

	objs, err := util.ToUnstructured(configyamlFile)
	if err != nil {
		klog.Fatalf("failed to parse yaml")
	}
	// look for resources of kind MutatingWebhookConfiguration
	for i := range objs {
		o := objs[i]
		if o.GetKind() == mutatingWebhookKind {
			// update the name in metadata
			if o.GetName() == mutatingwebhook {
				o.SetName(strings.Join([]string{mutatingwebhook, "-", tag}, ""))
				mutatingWebhooks = append(mutatingWebhooks, &o)
			}
		}
		if o.GetKind() == validatingWebhookKind {
			// update the name in metadata
			if o.GetName() == validatingwebhook {
				o.SetName(strings.Join([]string{validatingwebhook, "-", tag}, ""))
				validatingWebhooks = append(validatingWebhooks, &o)
			}
		}
	}
	return mutatingWebhooks, validatingWebhooks, err
}
