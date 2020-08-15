/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCluster_Default(t *testing.T) {
	g := NewWithT(t)

	c := &Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: ClusterSpec{
			ControlPlaneNode: Node{
				MachineType: "test",
			},
			WorkerNode: Node{
				MachineType: "testWorker",
			},
		},
	}
	c.Default()

	g.Expect(c.Spec.KubernetesVersion).To(Equal(defaultKubernetesVersion))
	g.Expect(*c.Spec.ControlPlaneNode.Replicas).To(Equal(defaultControlPlaneReplicas))
	g.Expect(*c.Spec.WorkerNode.Replicas).To(Equal(defaultWorkerReplicas))
}