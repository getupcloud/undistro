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
package e2e_test

import (
	"context"
	"os"
	"time"

	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	"github.com/getupio-undistro/undistro/pkg/kube"
	"github.com/getupio-undistro/undistro/pkg/meta"
	"github.com/getupio-undistro/undistro/pkg/scheme"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-api/test/framework/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Create EC2 cluster", func() {
	var (
		clusterClient client.Client
	)
	It("Should generate recommend cluster spec", func() {
		cmd := exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("create", "cluster", "undistro-ec2-e2e", "-n", "e2e", "--infra", "aws", "--flavor", "ec2", "--ssh-key-name", "undistro", "--generate-file"),
		)
		_, _, err := cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		_, err = os.Stat("undistro-ec2-e2e.yaml")
		Expect(err).ToNot(HaveOccurred())

	})
	It("Should create EC2 cluster", func() {
		cmd := exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("apply", "-f", "./testdata/ec2-cluster.yaml"),
		)
		out, _, err := cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		klog.Info(string(out))
		cmd = exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("apply", "-f", "./testdata/ec2-policies.yaml"),
		)
		out, _, err = cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		klog.Info(string(out))
		Eventually(func() bool {
			cl := appv1alpha1.Cluster{}
			key := client.ObjectKey{
				Name:      "undistro-ec2-e2e",
				Namespace: "e2e",
			}
			err = k8sClient.Get(context.Background(), key, &cl)
			Expect(err).ToNot(HaveOccurred())
			klog.Info(cl)
			return meta.InReadyCondition(cl.Status.Conditions)
		}, 120*time.Minute, 2*time.Minute).Should(BeTrue())
		klog.Info("Get Kubeconfig")
		cmd = exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("get", "kubeconfig", "undistro-ec2-e2e", "-n", "e2e"),
		)
		out, _, err = cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		getter := kube.NewMemoryRESTClientGetter(out, "")
		cfg, err := getter.ToRESTConfig()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())
		clusterClient, err = client.New(cfg, client.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(clusterClient).ToNot(BeNil())
		Eventually(func() []corev1.Node {
			nodes := corev1.NodeList{}
			err = clusterClient.List(context.Background(), &nodes)
			Expect(err).ToNot(HaveOccurred())
			klog.Info(nodes.Items)
			return nodes.Items
		}, 120*time.Minute, 2*time.Minute).Should(HaveLen(7))
		klog.Info("check kyverno")
		Eventually(func() []unstructured.Unstructured {
			list := unstructured.UnstructuredList{}
			list.SetGroupVersionKind(schema.FromAPIVersionAndKind("kyverno.io/v1", "ClusterPolicyList"))
			err = clusterClient.List(context.Background(), &list)
			Expect(err).ToNot(HaveOccurred())
			Expect(list).ToNot(BeNil())
			klog.Info(err)
			klog.Info(list.Items)
			return list.Items
		}, 120*time.Minute, 2*time.Minute).Should(HaveLen(16))
		klog.Info("delete cluster")
		cmd = exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("delete", "-f", "./testdata/ec2-policies.yaml"),
		)
		out, _, err = cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		klog.Info(string(out))
		cmd = exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("delete", "-f", "./testdata/ec2-cluster.yaml"),
		)
		out, _, err = cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		klog.Info(string(out))
	}, float64(240*time.Minute))
})
