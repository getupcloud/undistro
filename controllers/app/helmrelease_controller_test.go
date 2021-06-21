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
package app

import (
	appv1alpha1 "github.com/getupio-undistro/undistro/apis/app/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("HelmRelease Reconciler", func() {
	It("Should create a HelmRelease", func() {
		instance := &appv1alpha1.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-helm-",
				Namespace:    "default",
			},
			Spec: appv1alpha1.HelmReleaseSpec{
				Chart: appv1alpha1.ChartSource{
					RepoChartSource: appv1alpha1.RepoChartSource{
						RepoURL: "https://test",
						Name:    "test",
						Version: "1.0.0",
					},
				},
			},
		}
		Expect(testEnv.Create(ctx, instance)).ToNot(HaveOccurred())
		key := client.ObjectKey{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		}
		defer func() {
			err := testEnv.Delete(ctx, instance)
			Expect(err).NotTo(HaveOccurred())
		}()

		Eventually(func() bool {
			if err := testEnv.Get(ctx, key, instance); err != nil {
				return false
			}
			return len(instance.Finalizers) > 0
		}, timeout).Should(BeTrue())
	})

	It("Should require a chart", func() {
		instance := &appv1alpha1.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-helm-",
				Namespace:    "default",
			},
			Spec: appv1alpha1.HelmReleaseSpec{},
		}
		Expect(testEnv.Create(ctx, instance)).To(HaveOccurred())
	})
})
