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
package template_test

import (
	"context"
	"encoding/base64"

	"github.com/getupio-undistro/undistro/pkg/template"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Varaables", func() {
	var (
		vi  template.VariablesInput
		ctx context.Context
	)
	BeforeEach(func() {
		ctx = context.Background()
		vi = template.VariablesInput{
			Variables: make(map[string]interface{}),
			ClientSet: testEnv.GetClient(),
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
				err := template.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, ok := vi.Variables["UNDISTRO_TEST"]
				Eventually(func() bool {
					return ok && value == "test"
				}).Should(BeTrue())
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
				Expect(testEnv.GetClient().Create(ctx, &cfgMap)).To(BeNil())
				defer func() {
					err := testEnv.GetClient().Delete(ctx, &cfgMap)
					Expect(err).NotTo(HaveOccurred())
				}()
				err := template.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, ok := vi.Variables["UNDISTRO_TEST"]
				Eventually(func() bool {
					return ok && value == "testConfigMap"
				}).Should(BeTrue())
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
				Expect(testEnv.GetClient().Create(ctx, &secret)).To(BeNil())
				defer func() {
					err := testEnv.GetClient().Delete(ctx, &secret)
					Expect(err).NotTo(HaveOccurred())
				}()
				err := template.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, ok := vi.Variables["UNDISTRO_TEST"]
				Expect(value).To(Equal("testSecret"))
				Eventually(func() bool {
					return ok && value == "testSecret"
				}).Should(BeTrue())
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
				Expect(testEnv.GetClient().Create(ctx, &secret)).To(BeNil())
				defer func() {
					err := testEnv.GetClient().Delete(ctx, &secret)
					Expect(err).NotTo(HaveOccurred())
				}()
				err := template.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, ok := vi.Variables["UNDISTRO_TEST"]
				Eventually(func() bool {
					return ok && value == "testSecret"
				}).Should(BeTrue())
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
				err := template.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, ok := vi.Variables["UNDISTRO_TEST"]
				Eventually(func() bool {
					return !ok && value == nil
				}).Should(BeTrue())
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
				err := template.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).ToNot(HaveOccurred())
				value, ok := vi.Variables["UNDISTRO_TEST"]
				Eventually(func() bool {
					return !ok && value == nil
				}).Should(BeTrue())
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
				err := template.SetVariablesFromEnvVar(ctx, vi)
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
				err := template.SetVariablesFromEnvVar(ctx, vi)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
