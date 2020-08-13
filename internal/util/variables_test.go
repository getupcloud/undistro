/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package util_test

import (
	"context"
	"testing"

	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/internal/scheme"
	"github.com/getupcloud/undistro/internal/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
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
	p, err := undistroClient.GetProxy()
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
	It("should set variable value in EnvVar", func() {
		ctx := context.Background()
		vi := util.VariablesInput{
			VariablesClient: undistroClient.GetVariables(),
			ClientSet:       k8sClient,
			NamespacedName: types.NamespacedName{
				Namespace: "default",
			},
		}
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
})

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}
