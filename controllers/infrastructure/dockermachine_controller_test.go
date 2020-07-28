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

package controllers

import (
	"testing"

	. "github.com/onsi/gomega"

	clusterv1 "github.com/getupcloud/undistro/apis/cluster/v1alpha1"
	infrav1 "github.com/getupcloud/undistro/apis/infrastructure/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

func setupScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	if err := clusterv1.AddToScheme(s); err != nil {
		panic(err)
	}
	if err := infrav1.AddToScheme(s); err != nil {
		panic(err)
	}
	return s
}

func TestDockerMachineReconciler_DockerClusterToDockerMachines(t *testing.T) {
	g := NewWithT(t)

	clusterName := "my-cluster"
	dockerCluster := newDockerCluster(clusterName, "my-docker-cluster")
	dockerMachine1 := newDockerMachine("my-docker-machine-0")
	dockerMachine2 := newDockerMachine("my-docker-machine-1")
	objects := []runtime.Object{
		newCluster(clusterName),
		dockerCluster,
		newMachine(clusterName, "my-machine-0", dockerMachine1),
		newMachine(clusterName, "my-machine-1", dockerMachine2),
		// Intentionally omitted
		newMachine(clusterName, "my-machine-2", nil),
	}
	c := fake.NewFakeClientWithScheme(setupScheme(), objects...)
	r := DockerMachineReconciler{
		Client: c,
		Log:    klogr.New(),
	}
	mo := handler.MapObject{
		Object: dockerCluster,
	}
	out := r.DockerClusterToDockerMachines(mo)
	machineNames := make([]string, len(out))
	for i := range out {
		machineNames[i] = out[i].Name
	}
	g.Expect(out).To(HaveLen(2))
	g.Expect(machineNames).To(ConsistOf("my-machine-0", "my-machine-1"))
}

func newCluster(clusterName string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
	}
}

func newDockerCluster(clusterName, dockerName string) *infrav1.DockerCluster {
	return &infrav1.DockerCluster{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: dockerName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: clusterv1.GroupVersion.String(),
					Kind:       "Cluster",
					Name:       clusterName,
				},
			},
		},
	}
}

func newMachine(clusterName, machineName string, dockerMachine *infrav1.DockerMachine) *clusterv1.Machine {
	machine := &clusterv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name: machineName,
			Labels: map[string]string{
				clusterv1.ClusterLabelName: clusterName,
			},
		},
	}
	if dockerMachine != nil {
		machine.Spec.InfrastructureRef = v1.ObjectReference{
			Name:       dockerMachine.Name,
			Namespace:  dockerMachine.Namespace,
			Kind:       dockerMachine.Kind,
			APIVersion: dockerMachine.GroupVersionKind().GroupVersion().String(),
		}
	}
	return machine
}

func newDockerMachine(name string) *infrav1.DockerMachine {
	return &infrav1.DockerMachine{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec:   infrav1.DockerMachineSpec{},
		Status: infrav1.DockerMachineStatus{},
	}
}
