/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package cluster

import (
	"testing"

	yaml "github.com/getupio-undistro/undistro/client/yamlprocessor"
	"github.com/getupio-undistro/undistro/internal/test"
	. "github.com/onsi/gomega"
)

func Test_newClusterClient_YamlProcessor(t *testing.T) {

	tests := []struct {
		name   string
		opts   []Option
		assert func(*WithT, yaml.Processor)
	}{
		{
			name: "it creates a cluster client with simple yaml processor by default",
			assert: func(g *WithT, p yaml.Processor) {
				_, ok := (p).(*yaml.SimpleProcessor)
				g.Expect(ok).To(BeTrue())
			},
		},
		{
			name: "it creates a cluster client with specified yaml processor",
			opts: []Option{InjectYamlProcessor(test.NewFakeProcessor())},
			assert: func(g *WithT, p yaml.Processor) {
				_, ok := (p).(*yaml.SimpleProcessor)
				g.Expect(ok).To(BeFalse())
				_, ok = (p).(*test.FakeProcessor)
				g.Expect(ok).To(BeTrue())
			},
		},
		{
			name: "it creates a cluster client with simple yaml processor even if injected with nil processor",
			opts: []Option{InjectYamlProcessor(nil)},
			assert: func(g *WithT, p yaml.Processor) {
				g.Expect(p).ToNot(BeNil())
				_, ok := (p).(*yaml.SimpleProcessor)
				g.Expect(ok).To(BeTrue())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			client := newClusterClient(Kubeconfig{}, &fakeConfigClient{}, tt.opts...)
			g.Expect(client).ToNot(BeNil())
			tt.assert(g, client.processor)
		})
	}
}
