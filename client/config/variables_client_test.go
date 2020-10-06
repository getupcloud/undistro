/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package config

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/getupio-undistro/undistro/internal/test"
)

// Ensures FakeReader implements the Reader interface.
var _ Reader = &test.FakeReader{}

// Ensures the FakeVariableClient implements VariablesClient
var _ VariablesClient = &test.FakeVariableClient{}

func Test_variables_Get(t *testing.T) {
	reader := test.NewFakeReader().WithVar("foo", "bar")

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Returns value if the variable exists",
			args: args{
				key: "foo",
			},
			want:    "bar",
			wantErr: false,
		},
		{
			name: "Returns error if the variable does not exist",
			args: args{
				key: "baz",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			p := &variablesClient{
				reader: reader,
			}
			got, err := p.Get(tt.args.key)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(got).To(Equal(tt.want))
		})
	}
}
