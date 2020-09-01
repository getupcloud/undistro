package util_test

import (
	"context"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/internal/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Create or patch object", func() {
	Context("Create object", func() {
		It("should create an object that doesn't have spec", func() {
			s := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch",
					Namespace: "default",
				},
				StringData: map[string]string{
					"undistro": "testSecret",
				},
			}
			m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
			Expect(err).ToNot(HaveOccurred())
			u := unstructured.Unstructured{Object: m}
			err = util.CreateOrPatch(context.Background(), k8sClient, u)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create object", func() {
			s := &undistrov1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "getupcloud.com/v1alpha1",
					Kind:       "Cluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch",
					Namespace: "default",
				},
				Spec: undistrov1.ClusterSpec{},
			}
			m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
			Expect(err).ToNot(HaveOccurred())
			u := unstructured.Unstructured{Object: m}
			err = util.CreateOrPatch(context.Background(), k8sClient, u)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("Patch object", func() {
		It("should patch an object that doesn't have spec", func() {
			s := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch",
					Namespace: "default",
				},
				StringData: map[string]string{
					"undistro": "testSecret2",
				},
			}
			cl := corev1.Secret{}
			m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
			Expect(err).ToNot(HaveOccurred())
			u := unstructured.Unstructured{Object: m}
			err = util.CreateOrPatch(context.Background(), k8sClient, u)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "patch", Namespace: "default"}, &cl)
			Expect(err).ToNot(HaveOccurred())
			v := cl.Data["undistro"]
			Expect(string(v)).To(Equal("testSecret2"))
		})

		It("should exec without errors if object doesn't have changes", func() {
			s := &undistrov1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "getupcloud.com/v1alpha1",
					Kind:       "Cluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch",
					Namespace: "default",
				},
				Spec: undistrov1.ClusterSpec{},
			}
			m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
			Expect(err).ToNot(HaveOccurred())
			u := unstructured.Unstructured{Object: m}
			err = util.CreateOrPatch(context.Background(), k8sClient, u)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should patch object", func() {
			s := &undistrov1.Cluster{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "getupcloud.com/v1alpha1",
					Kind:       "Cluster",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "patch",
					Namespace: "default",
				},
				Spec: undistrov1.ClusterSpec{
					KubernetesVersion: "v1.19.0",
				},
			}
			cl := undistrov1.Cluster{}
			m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
			Expect(err).ToNot(HaveOccurred())
			u := unstructured.Unstructured{Object: m}
			err = util.CreateOrPatch(context.Background(), k8sClient, u)
			Expect(err).ToNot(HaveOccurred())
			err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "patch", Namespace: "default"}, &cl)
			Expect(err).ToNot(HaveOccurred())
			Expect(cl.Spec.KubernetesVersion).To(Equal("v1.19.0"))
		})
	})
})
