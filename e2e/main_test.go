package e2e_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/getupio-undistro/undistro/pkg/cli"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-api/test/framework/exec"
	"sigs.k8s.io/yaml"
)

var e2eRun = flag.Bool("e2e", false, "set true to run e2e tests")

func TestMain(m *testing.M) {
	klog.Info("E2E")
	runE2E := *e2eRun
	if !runE2E {
		klog.Info("Skiping E2E")
		os.Exit(0)
	}
	ctx := context.Background()
	klog.Info("Build docker image")
	sha := os.Getenv("GITHUB_SHA")
	image := fmt.Sprintf("localhost:5000/undistro:%s", sha)
	cmd := exec.NewCommand(
		exec.WithCommand("docker"),
		exec.WithArgs("build", "-t", image, "../"),
	)
	_, _, err := cmd.Run(ctx)
	if err != nil {
		klog.Info(err)
		os.Exit(1)
	}
	klog.Info("Push docker image")
	cmd = exec.NewCommand(
		exec.WithCommand("docker"),
		exec.WithArgs("push", image),
	)
	_, _, err = cmd.Run(ctx)
	if err != nil {
		klog.Info(err)
		os.Exit(1)
	}
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
		klog.Info(err)
		os.Exit(1)
	}
	klog.Info("Install UnDistro")
	cmd = exec.NewCommand(
		exec.WithCommand("undistro"),
		exec.WithArgs("--config", "undistro-config.yaml", "install"),
	)
	out, errOut, err := cmd.Run(ctx)
	if err != nil {
		klog.Info(err)
		os.Exit(1)
	}
	if len(errOut) > 0 {
		klog.Info(string(errOut))
		os.Exit(1)
	}
	if !bytes.Contains(out, []byte("Management cluster is ready to use.")) {
		klog.Info("failed to install undistro")
		os.Exit(1)
	}
	code := m.Run()
	os.Exit(code)
}
