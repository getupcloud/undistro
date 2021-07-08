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
package e2e_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/getupio-undistro/undistro/pkg/cli"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/cluster-api/test/framework/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/yaml"
)

var (
	e2eRun    = flag.Bool("e2e", false, "set true to run e2e tests")
	k8sClient client.Client
)

func TestMain(m *testing.M) {
	flag.Parse()
	fmt.Println("E2E")
	runE2E := *e2eRun
	if !runE2E {
		fmt.Println("Skipping E2E")
		os.Exit(0)
	}
	ctx := context.Background()
	fmt.Println("Build docker image")
	sha := os.Getenv("GITHUB_SHA")
	image := fmt.Sprintf("localhost:5000/undistro:%s", sha)
	cmd := exec.NewCommand(
		exec.WithCommand("bash"),
		exec.WithArgs("-c", fmt.Sprintf("../testbin/docker-build-e2e.sh %s", image)),
	)
	stout, stderr, err := cmd.Run(ctx)
	if err != nil {
		fmt.Println(string(stderr))
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(stout))
	fmt.Println("Push docker image")
	cmd = exec.NewCommand(
		exec.WithCommand("docker"),
		exec.WithArgs("push", image),
	)
	stout, _, err = cmd.Run(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println(string(stout))
	cfg := cli.Config{
		Providers: []cli.Provider{
			{
				Name: "aws",
				Configuration: map[string]interface{}{
					"accessKeyID":     os.Getenv("E2E_AWS_ACCESS_KEY_ID"),
					"secretAccessKey": os.Getenv("E2E_AWS_SECRET_ACCESS_KEY"),
				},
			},
		},
		CoreProviders: []cli.Provider{
			{
				Name: "undistro",
				Configuration: map[string]interface{}{
					"image": map[string]interface{}{
						"repository": "localhost:5000/undistro",
						"tag":        sha,
					},
				},
			},
		},
	}
	byt, _ := yaml.Marshal(cfg)
	err = ioutil.WriteFile("undistro-config.yaml", byt, 0700)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	k8sClient, err = client.New(config, client.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Install UnDistro")
	cmd = exec.NewCommand(
		exec.WithCommand("undistro"),
		exec.WithArgs("--config", "undistro-config.yaml", "install"),
	)
	out, stderr, _ := cmd.Run(ctx)
	fmt.Println(string(out))
	if !bytes.Contains(out, []byte("Management cluster is ready to use.")) {
		msg := "failed to install undistro: " + string(stderr)
		fmt.Println(msg)
		cmd = exec.NewCommand(
			exec.WithCommand("kubectl"),
			exec.WithArgs("get", "pods", "--all-namespaces"),
		)
		out, stderr, _ = cmd.Run(ctx)
		fmt.Println(string(out))
		fmt.Println("err:", string(stderr))
		cmd = exec.NewCommand(
			exec.WithCommand("kubectl"),
			exec.WithArgs("describe", "nodes"),
		)
		out, stderr, _ = cmd.Run(ctx)
		fmt.Println(string(out))
		fmt.Println("err:", string(stderr))
		cmd = exec.NewCommand(
			exec.WithCommand("kubectl"),
			exec.WithArgs("describe", "pods", "-n", "undistro-system"),
		)
		out, stderr, _ = cmd.Run(ctx)
		fmt.Println(string(out))
		fmt.Println("err:", string(stderr))
		podList := corev1.PodList{}
		err = k8sClient.List(ctx, &podList, client.InNamespace("undistro-system"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, p := range podList.Items {
			if strings.Contains(p.Name, "undistro") {
				cmd = exec.NewCommand(
					exec.WithCommand("undistro"),
					exec.WithArgs("logs", p.Name, "-n", "undistro-system", "-c", "manager", "--previous"),
				)
				out, stderr, _ = cmd.Run(ctx)
				fmt.Println(string(out))
				fmt.Println("err:", string(stderr))
			}
		}
		os.Exit(1)
	}

	cmd = exec.NewCommand(
		exec.WithCommand("undistro"),
		exec.WithArgs("get", "pods", "-n", "undistro-system"),
	)
	out, _, err = cmd.Run(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(out))
	cmd = exec.NewCommand(
		exec.WithCommand("mv"),
		exec.WithArgs("aws-iam-authenticator", "./bin"),
	)
	out, _, err = cmd.Run(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(out))
	code := m.Run()
	os.Exit(code)
}

func TestGinkgoSuite(t *testing.T) {
	SetDefaultEventuallyPollingInterval(1 * time.Minute)
	SetDefaultEventuallyTimeout(120 * time.Minute)
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"E2E Suite",
		[]Reporter{printer.NewlineReporter{}})
}
