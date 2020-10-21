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
			ControlPlaneNode: ControlPlaneNode{
				Node: Node{
					MachineType: "test",
				},
			},
			WorkerNodes: []Node{
				{
					MachineType: "testWorker",
				},
			},
		},
	}
	c.Default()

	g.Expect(c.Spec.KubernetesVersion).To(Equal(defaultKubernetesVersion))
	g.Expect(*c.Spec.ControlPlaneNode.Replicas).To(Equal(defaultControlPlaneReplicas))
	g.Expect(*c.Spec.WorkerNodes[0].Replicas).To(Equal(defaultWorkerReplicas))
}
