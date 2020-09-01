package util_test

import (
	"path/filepath"
	"testing"

	uclient "github.com/getupcloud/undistro/client"
	"github.com/getupcloud/undistro/internal/scheme"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}
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

func TestUtil(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Util Suite",
		[]Reporter{printer.NewlineReporter{}})
}
