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
package controllers

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	"github.com/getupio-undistro/undistro/pkg/test"
	// +kubebuilder:scaffold:imports
)

const (
	timeout = time.Second * 30
)

var (
	testEnv *test.Environment
	ctx     = ctrl.SetupSignalHandler()
)

func TestMain(m *testing.M) {
	fmt.Println("Creating new test environment")
	testEnv = test.NewEnvironment()
	if err := (&ClusterReconciler{
		Client: testEnv,
		Scheme: testEnv.GetScheme(),
		Log:    testEnv.GetLogger(),
	}).SetupWithManager(testEnv); err != nil {
		panic(fmt.Sprintf("Failed to start ClusterReconciler : %v", err))
	}
	if err := (&HelmReleaseReconciler{
		Client: testEnv,
		Scheme: testEnv.GetScheme(),
		Log:    testEnv.GetLogger(),
		config: testEnv.GetConfig(),
	}).SetupWithManager(testEnv); err != nil {
		panic(fmt.Sprintf("Failed to start HelmReleaseReconciler : %v", err))
	}
	go func() {
		fmt.Println("Starting the manager")
		if err := testEnv.StartManager(ctx); err != nil {
			panic(fmt.Sprintf("Failed to start the envtest manager: %v", err))
		}
	}()
	// wait for webhook port to be open prior to running tests
	testEnv.WaitForWebhooks()

	code := m.Run()

	fmt.Println("Tearing down test suite")
	if err := testEnv.Stop(); err != nil {
		panic(fmt.Sprintf("Failed to stop envtest: %v", err))
	}

	os.Exit(code)
}

// TestGinkgoSuite will run the ginkgo tests.
// This will run with the testEnv setup and teardown in TestMain.
func TestGinkgoSuite(t *testing.T) {
	SetDefaultEventuallyPollingInterval(100 * time.Millisecond)
	SetDefaultEventuallyTimeout(timeout)
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controllers Suite",
		[]Reporter{printer.NewlineReporter{}})
}
