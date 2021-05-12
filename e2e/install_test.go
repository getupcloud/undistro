package e2e_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Validate UnDistro Installation", func() {
	It("Verify if pods not crash", func() {
		Eventually(func() []corev1.Pod {
			podList := corev1.PodList{}
			err := k8sClient.List(context.Background(), &podList, client.InNamespace("undistro-system"))
			Expect(err).ToNot(HaveOccurred())
			return podList.Items
		}, 10*time.Minute, 1*time.Minute).Should(HaveLen(12))
		Eventually(func() bool {
			podList := corev1.PodList{}
			err := k8sClient.List(context.Background(), &podList, client.InNamespace("undistro-system"))
			Expect(err).ToNot(HaveOccurred())
			running := true
			for _, p := range podList.Items {
				if p.Status.Phase != corev1.PodRunning {
					running = false
				}
			}
			return running
		}, 10*time.Minute, 1*time.Minute).Should(BeTrue())
		It("Verify if UnDistro AWS is correctly installed", func() {
			Eventually(func() string {
				s := corev1.Secret{}
				key := client.ObjectKey{
					Name:      "undistro-aws-config",
					Namespace: "undistro-system",
				}
				err := k8sClient.Get(context.Background(), key, &s)
				Expect(err).ToNot(HaveOccurred())
				return string(s.Data["credentials"])
			}, 10*time.Minute, 1*time.Minute).ShouldNot(BeEmpty())
		})
	})
})
