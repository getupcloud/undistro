module github.com/getupio-undistro/undistro

go 1.16

require (
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/aws/aws-sdk-go v1.36.12
	github.com/go-logr/logr v0.4.0
	github.com/gorilla/mux v1.8.0
	github.com/json-iterator/go v1.1.10
	github.com/kyverno/kyverno v1.3.6
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	helm.sh/helm/v3 v3.1.0-rc.1.0.20210519153047-b1e247643251 // branch main
	k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/cli-runtime v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog/v2 v2.8.0
	k8s.io/kubectl v0.21.1
	sigs.k8s.io/cluster-api v0.3.11-0.20210519202651-f5dec18def70 // branch master
	sigs.k8s.io/controller-runtime v0.9.0-beta.5
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
