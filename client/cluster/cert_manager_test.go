/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cluster

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/getupio-undistro/undistro/client/config"
	manifests "github.com/getupio-undistro/undistro/config"
	"github.com/getupio-undistro/undistro/internal/scheme"
	"github.com/getupio-undistro/undistro/internal/test"
	. "github.com/onsi/gomega"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
)

func Test_VersionMarkerUpToDate(t *testing.T) {
	yaml, err := manifests.Asset(embeddedCertManagerManifestPath)
	if err != nil {
		t.Fatalf("Failed to get cert-manager.yaml asset data: %v", err)
	}

	actualHash := fmt.Sprintf("%x", sha256.Sum256(yaml))
	_, embeddedHash := embeddedCertManagerVersion()
	if actualHash != embeddedHash {
		t.Errorf("The cert-manager.yaml asset data has changed, but the version marker embeddedCertManagerManifestVersion has not been updated. Expected hash to be: %s", actualHash)
	}
}

func Test_certManagerClient_getManifestObjects(t *testing.T) {

	tests := []struct {
		name      string
		expectErr bool
		assert    func(*testing.T, []unstructured.Unstructured)
	}{
		{
			name:      "it should not contain the cert-manager-leaderelection ClusterRoleBinding",
			expectErr: false,
			assert: func(t *testing.T, objs []unstructured.Unstructured) {
				for _, o := range objs {
					if o.GetKind() == "ClusterRoleBinding" && o.GetName() == "cert-manager-leaderelection" {
						t.Error("should not find cert-manager-leaderelection ClusterRoleBinding")
					}
				}
			},
		},
		{
			name:      "the MutatingWebhookConfiguration should have sideEffects set to None ",
			expectErr: false,
			assert: func(t *testing.T, objs []unstructured.Unstructured) {
				found := false
				for i := range objs {
					o := objs[i]
					if o.GetKind() == "MutatingWebhookConfiguration" && o.GetName() == "cert-manager-webhook" {
						w := &admissionregistration.MutatingWebhookConfiguration{}
						err := scheme.Scheme.Convert(&o, w, nil)
						if err != nil {
							t.Errorf("did not expect err, got %s", err)
						}
						if len(w.Webhooks) != 1 {
							t.Error("expected 1 webhook to be configured")
						}
						wh := w.Webhooks[0]
						if wh.SideEffects != nil && *wh.SideEffects == admissionregistration.SideEffectClassNone {
							found = true
						}
					}
				}
				if !found {
					t.Error("Expected to find cert-manager-webhook MutatingWebhookConfiguration with sideEffects=None")
				}
			},
		},
		{
			name:      "the ValidatingWebhookConfiguration should have sideEffects set to None ",
			expectErr: false,
			assert: func(t *testing.T, objs []unstructured.Unstructured) {
				found := false
				for i := range objs {
					o := objs[i]
					if o.GetKind() == "ValidatingWebhookConfiguration" && o.GetName() == "cert-manager-webhook" {
						w := &admissionregistration.ValidatingWebhookConfiguration{}
						err := scheme.Scheme.Convert(&o, w, nil)
						if err != nil {
							t.Errorf("did not expect err, got %s", err)
						}
						if len(w.Webhooks) != 1 {
							t.Error("expected 1 webhook to be configured")
						}
						wh := w.Webhooks[0]
						if wh.SideEffects != nil && *wh.SideEffects == admissionregistration.SideEffectClassNone {
							found = true
						}
					}
				}
				if !found {
					t.Error("Expected to find cert-manager-webhook ValidatingWebhookConfiguration with sideEffects=None")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			pollImmediateWaiter := func(interval, timeout time.Duration, condition wait.ConditionFunc) error {
				return nil
			}
			fakeConfigClient := newFakeConfig("")

			cm := newCertMangerClient(fakeConfigClient, nil, pollImmediateWaiter)
			objs, err := cm.getManifestObjs()

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).ToNot(HaveOccurred())
			tt.assert(t, objs)
		})
	}

}

func Test_GetTimeout(t *testing.T) {

	pollImmediateWaiter := func(interval, timeout time.Duration, condition wait.ConditionFunc) error {
		return nil
	}

	tests := []struct {
		name    string
		timeout string
		want    time.Duration
	}{
		{
			name:    "no custom value set for timeout",
			timeout: "",
			want:    10 * time.Minute,
		},
		{
			name:    "a custom value of timeout is set",
			timeout: "5m",
			want:    5 * time.Minute,
		},
		{
			name:    "invalid custom value of timeout is set",
			timeout: "5",
			want:    10 * time.Minute,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			fakeConfigClient := newFakeConfig(tt.timeout)

			cm := newCertMangerClient(fakeConfigClient, nil, pollImmediateWaiter)
			tm := cm.getWaitTimeout()

			g.Expect(tm).To(Equal(tt.want))
		})
	}

}

func newFakeConfig(timeout string) fakeConfigClient {
	fakeReader := test.NewFakeReader().WithVar("cert-manager-timeout", timeout)

	client, _ := config.New("fake-config", config.InjectReader(fakeReader))
	return fakeConfigClient{
		fakeReader:         fakeReader,
		internalclient:     client,
		certManagerTimeout: timeout,
	}
}

type fakeConfigClient struct {
	fakeReader         *test.FakeReader
	internalclient     config.Client
	certManagerTimeout string
}

var _ config.Client = &fakeConfigClient{}

func (f fakeConfigClient) Providers() config.ProvidersClient {
	return f.internalclient.Providers()
}

func (f fakeConfigClient) Variables() config.VariablesClient {
	return f.internalclient.Variables()
}

func (f fakeConfigClient) ImageMeta() config.ImageMetaClient {
	return f.internalclient.ImageMeta()
}

func (f *fakeConfigClient) WithVar(key, value string) *fakeConfigClient {
	f.fakeReader.WithVar(key, value)
	return f
}

func (f *fakeConfigClient) WithProvider(provider config.Provider) *fakeConfigClient {
	f.fakeReader.WithProvider(provider.Name(), provider.Type(), provider.URL())
	return f
}
