/*
Copyright 2020 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"
)

func Test_isValidNameForAWS(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "name with _",
			args: args{
				name: "cluster_name",
			},
			want: true,
		},
		{
			name: "name with -",
			args: args{
				name: "cluster-name",
			},
			want: true,
		},
		{
			name: "name with @",
			args: args{
				name: "cluster@name",
			},
			want: false,
		},
		{
			name: "name with .",
			args: args{
				name: "cluster.name",
			},
			want: false,
		},
		{
			name: "name with space",
			args: args{
				name: "cluster name",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidNameForAWS(tt.args.name); got != tt.want {
				t.Errorf("isValidNameForAWS() = %v, want %v", got, tt.want)
			}
		})
	}
}
