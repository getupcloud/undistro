/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package repository

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"github.com/getupcloud/undistro/client/config"
	yaml "github.com/getupcloud/undistro/client/yamlprocessor"
	"github.com/getupcloud/undistro/internal/test"
)

func Test_templates_Get(t *testing.T) {
	p1 := config.NewProvider("p1", "", undistrov1.BootstrapProviderType, nil)

	type fields struct {
		version               string
		provider              config.Provider
		repository            Repository
		configVariablesClient config.VariablesClient
		processor             yaml.Processor
	}
	type args struct {
		flavor            string
		targetNamespace   string
		listVariablesOnly bool
	}
	type want struct {
		variables       []string
		targetNamespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "pass if default template exists",
			fields: fields{
				version:  "v1.0",
				provider: p1,
				repository: test.NewFakeRepository().
					WithPaths("root", "").
					WithDefaultVersion("v1.0").
					WithFile("v1.0", "cluster-template.yaml", templateMapYaml),
				configVariablesClient: test.NewFakeVariableClient().WithVar(variableName, variableValue),
				processor:             yaml.NewSimpleProcessor(),
			},
			args: args{
				flavor:            "",
				targetNamespace:   "ns1",
				listVariablesOnly: false,
			},
			want: want{
				variables:       []string{variableName},
				targetNamespace: "ns1",
			},
			wantErr: false,
		},
		{
			name: "pass if template for a flavor exists",
			fields: fields{
				version:  "v1.0",
				provider: p1,
				repository: test.NewFakeRepository().
					WithPaths("root", "").
					WithDefaultVersion("v1.0").
					WithFile("v1.0", "cluster-template-prod.yaml", templateMapYaml),
				configVariablesClient: test.NewFakeVariableClient().WithVar(variableName, variableValue),
				processor:             yaml.NewSimpleProcessor(),
			},
			args: args{
				flavor:            "prod",
				targetNamespace:   "ns1",
				listVariablesOnly: false,
			},
			want: want{
				variables:       []string{variableName},
				targetNamespace: "ns1",
			},
			wantErr: false,
		},
		{
			name: "fails if template does not exists",
			fields: fields{
				version:  "v1.0",
				provider: p1,
				repository: test.NewFakeRepository().
					WithPaths("root", "").
					WithDefaultVersion("v1.0"),
				configVariablesClient: test.NewFakeVariableClient().WithVar(variableName, variableValue),
				processor:             yaml.NewSimpleProcessor(),
			},
			args: args{
				flavor:            "",
				targetNamespace:   "ns1",
				listVariablesOnly: false,
			},
			wantErr: true,
		},
		{
			name: "fails if variables does not exists",
			fields: fields{
				version:  "v1.0",
				provider: p1,
				repository: test.NewFakeRepository().
					WithPaths("root", "").
					WithDefaultVersion("v1.0").
					WithFile("v1.0", "cluster-template.yaml", templateMapYaml),
				configVariablesClient: test.NewFakeVariableClient(),
				processor:             yaml.NewSimpleProcessor(),
			},
			args: args{
				flavor:            "",
				targetNamespace:   "ns1",
				listVariablesOnly: false,
			},
			wantErr: true,
		},
		{
			name: "pass if variables does not exists but listVariablesOnly flag is set",
			fields: fields{
				version:  "v1.0",
				provider: p1,
				repository: test.NewFakeRepository().
					WithPaths("root", "").
					WithDefaultVersion("v1.0").
					WithFile("v1.0", "cluster-template.yaml", templateMapYaml),
				configVariablesClient: test.NewFakeVariableClient(),
				processor:             yaml.NewSimpleProcessor(),
			},
			args: args{
				flavor:            "",
				targetNamespace:   "ns1",
				listVariablesOnly: true,
			},
			want: want{
				variables:       []string{variableName},
				targetNamespace: "ns1",
			},
			wantErr: false,
		},
		{
			name: "returns error if processor is unable to get variables",
			fields: fields{
				version:  "v1.0",
				provider: p1,
				repository: test.NewFakeRepository().
					WithPaths("root", "").
					WithDefaultVersion("v1.0").
					WithFile("v1.0", "cluster-template.yaml", templateMapYaml),
				configVariablesClient: test.NewFakeVariableClient().WithVar(variableName, variableValue),
				processor:             test.NewFakeProcessor().WithGetVariablesErr(errors.New("cannot get vars")).WithTemplateName("cluster-template.yaml"),
			},
			args: args{
				targetNamespace:   "ns1",
				listVariablesOnly: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			f := newTemplateClient(
				TemplateClientInput{
					version:               tt.fields.version,
					provider:              tt.fields.provider,
					repository:            tt.fields.repository,
					configVariablesClient: tt.fields.configVariablesClient,
					processor:             tt.fields.processor,
				},
			)
			got, err := f.Get(tt.args.flavor, tt.args.targetNamespace, tt.args.listVariablesOnly)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(got.Variables()).To(Equal(tt.want.variables))
			g.Expect(got.TargetNamespace()).To(Equal(tt.want.targetNamespace))

			// check variable replaced in yaml
			yaml, err := got.Yaml()
			g.Expect(err).NotTo(HaveOccurred())

			if !tt.args.listVariablesOnly {
				g.Expect(yaml).To(ContainSubstring((fmt.Sprintf("variable: %s", variableValue))))
			}

			// check if target namespace is set
			for _, o := range got.Objs() {
				g.Expect(o.GetNamespace()).To(Equal(tt.want.targetNamespace))
			}
		})
	}
}
