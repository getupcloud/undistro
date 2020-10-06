/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/getupio-undistro/undistro/client/config"
	yaml "github.com/getupio-undistro/undistro/client/yamlprocessor"
	"github.com/getupio-undistro/undistro/internal/test"
)

// Nb.We are using core objects vs Machines/Cluster etc. because it is easier to test (you don't have to deal with CRDs
// or schema issues), but this is ok because a template can be any yaml that complies the undistro contract.
var templateMapYaml = []byte("apiVersion: v1\n" +
	"data:\n" +
	fmt.Sprintf("  variable: ${%s}\n", variableName) +
	"kind: ConfigMap\n" +
	"metadata:\n" +
	"  name: manager")

func Test_newTemplate(t *testing.T) {
	type args struct {
		rawYaml               []byte
		configVariablesClient config.VariablesClient
		processor             yaml.Processor
		targetNamespace       string
		listVariablesOnly     bool
	}
	type want struct {
		variables       []string
		targetNamespace string
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "variable is replaced and namespace fixed",
			args: args{
				rawYaml:               templateMapYaml,
				configVariablesClient: test.NewFakeVariableClient().WithVar(variableName, variableValue),
				processor:             yaml.NewSimpleProcessor(),
				targetNamespace:       "ns1",
				listVariablesOnly:     false,
			},
			want: want{
				variables:       []string{variableName},
				targetNamespace: "ns1",
			},
			wantErr: false,
		},
		{
			name: "List variable only",
			args: args{
				rawYaml:               templateMapYaml,
				configVariablesClient: test.NewFakeVariableClient(),
				processor:             yaml.NewSimpleProcessor(),
				targetNamespace:       "ns1",
				listVariablesOnly:     true,
			},
			want: want{
				variables:       []string{variableName},
				targetNamespace: "ns1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got, err := NewTemplate(TemplateInput{
				RawArtifact:           tt.args.rawYaml,
				ConfigVariablesClient: tt.args.configVariablesClient,
				Processor:             tt.args.processor,
				TargetNamespace:       tt.args.targetNamespace,
				ListVariablesOnly:     tt.args.listVariablesOnly,
			})
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(got.Variables()).To(Equal(tt.want.variables))
			g.Expect(got.TargetNamespace()).To(Equal(tt.want.targetNamespace))

			if tt.args.listVariablesOnly {
				return
			}

			// check variable replaced in components
			yaml, err := got.Yaml()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(yaml).To(ContainSubstring((fmt.Sprintf("variable: %s", variableValue))))
		})
	}
}
