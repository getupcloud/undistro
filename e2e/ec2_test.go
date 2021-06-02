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
	"fmt"
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
		fmt.Println(string(out))
		cmd = exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("apply", "-f", "./testdata/ec2-policies.yaml"),
		)
		out, _, err = cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(string(out))
		Eventually(func() bool {
			cl := appv1alpha1.Cluster{}
			key := client.ObjectKey{
				Name:      "undistro-ec2-e2e",
				Namespace: "e2e",
			}
			err = k8sClient.Get(context.Background(), key, &cl)
			Expect(err).ToNot(HaveOccurred())
			fmt.Println(cl)
			return meta.InReadyCondition(cl.Status.Conditions)
		}, 120*time.Minute, 2*time.Minute).Should(BeTrue())
		fmt.Println("Get Kubeconfig")
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
			if err != nil {
				return []corev1.Node{}
			}
			fmt.Println(nodes.Items)
			fmt.Println(len(nodes.Items))
			return nodes.Items
		}, 120*time.Minute, 2*time.Minute).Should(HaveLen(7))
		Eventually(func() []corev1.Node {
			cpNodes := make([]corev1.Node, 0)
			nodes := corev1.NodeList{}
			err = clusterClient.List(context.Background(), &nodes)
			if err != nil {
				return []corev1.Node{}
			}
			fmt.Println(nodes.Items)
			fmt.Println(len(nodes.Items))
			for _, n := range nodes.Items {
				labels := n.GetLabels()
				if labels != nil {
					_, okCP := labels[meta.LabelK8sCP]
					_, okMaster := labels[meta.LabelK8sMaster]
					if okCP || okMaster {
						cpNodes = append(cpNodes, n)
					}
				}
			}
			fmt.Println(cpNodes)
			fmt.Println(len(cpNodes))
			return cpNodes
		}, 120*time.Minute, 2*time.Minute).Should(HaveLen(3))
		fmt.Println("check kyverno")
		Eventually(func() []unstructured.Unstructured {
			list := unstructured.UnstructuredList{}
			list.SetGroupVersionKind(schema.FromAPIVersionAndKind("kyverno.io/v1", "ClusterPolicyList"))
			err = clusterClient.List(context.Background(), &list)
			if err != nil {
				fmt.Println(err)
				return []unstructured.Unstructured{}
			}
			fmt.Println(list.Items)
			fmt.Println(len(list.Items))
			if undistroPodName != "" {
				cmd := exec.NewCommand(
					exec.WithCommand("undistro"),
					exec.WithArgs("logs", undistroPodName, "-n", "undistro-system", "-c", "manager"),
				)
				out, stderr, err := cmd.Run(context.Background())
				if err != nil {
					fmt.Println(string(stderr))
					fmt.Println(err)
					fmt.Println(string(out))
					return list.Items
				}
				fmt.Println(string(out))
			}
			return list.Items
		}, 120*time.Minute, 2*time.Minute).Should(HaveLen(16))
		fmt.Println("delete cluster")

		cmd = exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("delete", "-f", "./testdata/ec2-policies.yaml"),
		)
		out, _, err = cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(string(out))
		cmd = exec.NewCommand(
			exec.WithCommand("undistro"),
			exec.WithArgs("delete", "-f", "./testdata/ec2-cluster.yaml"),
		)
		out, _, err = cmd.Run(context.Background())
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(string(out))
	}, float64(240*time.Minute))
})
