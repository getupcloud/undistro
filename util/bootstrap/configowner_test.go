/*
Copyright 2019 The Kubernetes Authors.

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

package util

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	bootstrapv1 "github.com/getupcloud/undistro/apis/bootstrap/v1alpha1"
	clusterv1 "github.com/getupcloud/undistro/apis/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetConfigOwner(t *testing.T) {
	g := NewWithT(t)

	scheme := runtime.NewScheme()
	g.Expect(clusterv1.AddToScheme(scheme)).To(Succeed())

	t.Run("should get the owner when present (Machine)", func(t *testing.T) {
		g := NewWithT(t)
		myMachine := &clusterv1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-machine",
				Namespace: "my-ns",
				Labels: map[string]string{
					clusterv1.MachineControlPlaneLabelName: "",
				},
			},
			Spec: clusterv1.MachineSpec{
				ClusterName: "my-cluster",
				Bootstrap: clusterv1.Bootstrap{
					DataSecretName: pointer.StringPtr("my-data-secret"),
				},
			},
			Status: clusterv1.MachineStatus{
				InfrastructureReady: true,
			},
		}

		c := fake.NewFakeClientWithScheme(scheme, myMachine)
		obj := &bootstrapv1.KubeadmConfig{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       "Machine",
						APIVersion: clusterv1.GroupVersion.String(),
						Name:       "my-machine",
					},
				},
				Namespace: "my-ns",
				Name:      "my-resource-owned-by-machine",
			},
		}
		configOwner, err := GetConfigOwner(context.TODO(), c, obj)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(configOwner).ToNot(BeNil())
		g.Expect(configOwner.ClusterName()).To(BeEquivalentTo("my-cluster"))
		g.Expect(configOwner.IsInfrastructureReady()).To(BeTrue())
		g.Expect(configOwner.IsControlPlaneMachine()).To(BeTrue())
		g.Expect(*configOwner.DataSecretName()).To(BeEquivalentTo("my-data-secret"))
	})

	t.Run("return an error when not found", func(t *testing.T) {
		g := NewWithT(t)
		c := fake.NewFakeClientWithScheme(scheme)
		obj := &bootstrapv1.KubeadmConfig{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       "Machine",
						APIVersion: clusterv1.GroupVersion.String(),
						Name:       "my-machine",
					},
				},
				Namespace: "my-ns",
				Name:      "my-resource-owned-by-machine",
			},
		}
		_, err := GetConfigOwner(context.TODO(), c, obj)
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("return nothing when there is no owner", func(t *testing.T) {
		g := NewWithT(t)
		c := fake.NewFakeClientWithScheme(scheme)
		obj := &bootstrapv1.KubeadmConfig{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{},
				Namespace:       "my-ns",
				Name:            "my-resource-owned-by-machine",
			},
		}
		configOwner, err := GetConfigOwner(context.TODO(), c, obj)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(configOwner).To(BeNil())
	})
}
