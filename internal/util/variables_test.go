/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package util_test

import (
	"context"
	"encoding/base64"
	"testing"

	uclient "github.com/getupio-undistro/undistro/client"
	"github.com/getupio-undistro/undistro/internal/scheme"
	"github.com/getupio-undistro/undistro/internal/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg            *rest.Config
	k8sClient      client.Client
	testEnv        *envtest.Environment
	undistroClient uclient.Client
)

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}
	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())
	undistroClient, err = uclient.New("")
	Expect(err).ToNot(HaveOccurred())
	Expect(undistroClient).ToNot(BeNil())
	p, err := undistroClient.GetProxy(uclient.Kubeconfig{})
	Expect(err).ToNot(HaveOccurred())
	Expect(p).ToNot(BeNil())
	p.SetConfig(cfg)
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

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

func TestVariables(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Variables Suite",
		[]Reporter{printer.NewlineReporter{}})
}
