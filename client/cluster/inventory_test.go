/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cluster

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	undistrov1 "github.com/getupio-undistro/undistro/api/v1alpha1"
	"github.com/getupio-undistro/undistro/internal/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

func fakePollImmediateWaiter(interval, timeout time.Duration, condition wait.ConditionFunc) error {
	return nil
}

func Test_inventoryClient_EnsureCustomResourceDefinitions(t *testing.T) {
	type fields struct {
		alreadyHasCRD bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Has not CRD",
			fields: fields{
				alreadyHasCRD: false,
			},
			wantErr: false,
		},
		{
			name: "Already has CRD",
			fields: fields{
				alreadyHasCRD: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p := newInventoryClient(test.NewFakeProxy(), fakePollImmediateWaiter)
			if tt.fields.alreadyHasCRD {
				//forcing creation of metadata before test
				g.Expect(p.EnsureCustomResourceDefinitions()).To(Succeed())
			}

			err := p.EnsureCustomResourceDefinitions()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

var fooProvider = undistrov1.Provider{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "ns1"}}

func Test_inventoryClient_List(t *testing.T) {
	type fields struct {
		initObjs []runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		want    []undistrov1.Provider
		wantErr bool
	}{
		{
			name: "Get list",
			fields: fields{
				initObjs: []runtime.Object{
					&fooProvider,
				},
			},
			want: []undistrov1.Provider{
				fooProvider,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p := newInventoryClient(test.NewFakeProxy().WithObjs(tt.fields.initObjs...), fakePollImmediateWaiter)
			got, err := p.List()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(got.Items).To(ConsistOf(tt.want))
		})
	}
}

func Test_inventoryClient_Create(t *testing.T) {
	type fields struct {
		proxy Proxy
	}
	type args struct {
		m undistrov1.Provider
	}
	providerV2 := fakeProvider("infra", undistrov1.InfrastructureProviderType, "v0.2.0", "", "")
	providerV3 := fakeProvider("infra", undistrov1.InfrastructureProviderType, "v0.3.0", "", "")

	tests := []struct {
		name          string
		fields        fields
		args          args
		wantProviders []undistrov1.Provider
		wantErr       bool
	}{
		{
			name: "Creates a provider",
			fields: fields{
				proxy: test.NewFakeProxy(),
			},
			args: args{
				m: providerV2,
			},
			wantProviders: []undistrov1.Provider{
				providerV2,
			},
			wantErr: false,
		},
		{
			name: "Patches a provider",
			fields: fields{
				proxy: test.NewFakeProxy().WithObjs(&providerV2),
			},
			args: args{
				m: providerV3,
			},
			wantProviders: []undistrov1.Provider{
				providerV3,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p := &inventoryClient{
				proxy: tt.fields.proxy,
			}
			err := p.Create(tt.args.m)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).NotTo(HaveOccurred())

			got, err := p.List()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).NotTo(HaveOccurred())

			for i := range got.Items {
				tt.wantProviders[i].ResourceVersion = got.Items[i].ResourceVersion
			}

			g.Expect(got.Items).To(ConsistOf(tt.wantProviders))
		})
	}
}
