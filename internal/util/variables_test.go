/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package util_test

import (
	"context"
	"encoding/base64"

	"github.com/getupcloud/undistro/internal/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Varaables", func() {
	var (
		vi  util.VariablesInput
		ctx context.Context
	)
	BeforeEach(func() {
		ctx = context.Background()
		vi = util.VariablesInput{
			VariablesClient: undistroClient.GetVariables(),
			ClientSet:       k8sClient,
			NamespacedName: types.NamespacedName{
				Namespace: "default",
			},
		}
	})
	Describe("set variables", func() {
		Context("set variables with success", func() {
			It("should set variable using basic EnvVar", func() {
				e := corev1.EnvVar{
					Name:  "UNDISTRO_TEST",
					Value: "test",
				}
				vi.EnvVars = append(vi.EnvVars, e)
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, err := undistroClient.GetVariables().Get("UNDISTRO_TEST")
				Expect(err).ToNot(HaveOccurred())
				Expect(value).To(Equal("test"))
			})

			It("should set variable when EnvVar using configMap", func() {
				e := corev1.EnvVar{
					Name: "UNDISTRO_TEST",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "configmaptest",
							},
							Key: "undistro",
						},
					},
				}
				vi.EnvVars = append(vi.EnvVars, e)
				cfgMap := corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "configmaptest",
						Namespace: "default",
					},
					Data: map[string]string{
						"undistro": "testConfigMap",
					},
				}
				Expect(k8sClient.Create(ctx, &cfgMap)).To(BeNil())
				defer func() {
					err := k8sClient.Delete(ctx, &cfgMap)
					Expect(err).NotTo(HaveOccurred())
				}()
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, err := undistroClient.GetVariables().Get("UNDISTRO_TEST")
				Expect(err).ToNot(HaveOccurred())
				Expect(value).To(Equal("testConfigMap"))
			})

			It("should set variable when EnvVar using secret stringData", func() {
				e := corev1.EnvVar{
					Name: "UNDISTRO_TEST",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secrettest",
							},
							Key: "undistro",
						},
					},
				}
				vi.EnvVars = append(vi.EnvVars, e)
				secret := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secrettest",
						Namespace: "default",
					},
					StringData: map[string]string{
						"undistro": "testSecret",
					},
				}
				Expect(k8sClient.Create(ctx, &secret)).To(BeNil())
				defer func() {
					err := k8sClient.Delete(ctx, &secret)
					Expect(err).NotTo(HaveOccurred())
				}()
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, err := undistroClient.GetVariables().Get("UNDISTRO_TEST")
				Expect(err).ToNot(HaveOccurred())
				Expect(value).To(Equal("testSecret"))
			})

			It("should set variable when EnvVar using secret data", func() {
				e := corev1.EnvVar{
					Name: "UNDISTRO_TEST",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secrettest",
							},
							Key: "undistro",
						},
					},
				}
				vi.EnvVars = append(vi.EnvVars, e)
				b64 := base64.StdEncoding.EncodeToString([]byte("testSecret"))
				secret := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secrettest",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"undistro": []byte(b64),
					},
				}
				Expect(k8sClient.Create(ctx, &secret)).To(BeNil())
				defer func() {
					err := k8sClient.Delete(ctx, &secret)
					Expect(err).NotTo(HaveOccurred())
				}()
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, err := undistroClient.GetVariables().Get("UNDISTRO_TEST")
				Expect(err).ToNot(HaveOccurred())
				Expect(value).To(Equal("testSecret"))
			})
		})

		Context("Should return an empty string", func() {
			It("should return an empty string when configMap is not found", func() {
				e := corev1.EnvVar{
					Name: "UNDISTRO_TEST",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secrettest",
							},
							Key: "undistro",
						},
					},
				}
				vi.EnvVars = append(vi.EnvVars, e)
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, err := undistroClient.GetVariables().Get("UNDISTRO_TEST")
				Expect(err).ToNot(HaveOccurred())
				Expect(value).To(BeEmpty())
			})

			It("should return an empty string when secret is not found", func() {
				e := corev1.EnvVar{
					Name: "UNDISTRO_TEST",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secrettest",
							},
							Key: "undistro",
						},
					},
				}
				vi.EnvVars = append(vi.EnvVars, e)
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, err := undistroClient.GetVariables().Get("UNDISTRO_TEST")
				Expect(err).ToNot(HaveOccurred())
				Expect(value).To(BeEmpty())
			})
		})

		Context("should return an error", func() {
			It("should return an error when fieldRef is not nil", func() {
				e := corev1.EnvVar{
					Name: "UNDISTRO_TEST",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{},
					},
				}
				vi.EnvVars = append(vi.EnvVars, e)
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).To(HaveOccurred())
			})

			It("should return an error when resourceFieldRef is not nil", func() {
				e := corev1.EnvVar{
					Name: "UNDISTRO_TEST",
					ValueFrom: &corev1.EnvVarSource{
						ResourceFieldRef: &corev1.ResourceFieldSelector{},
					},
				}
				vi.EnvVars = append(vi.EnvVars, e)
				err := util.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
