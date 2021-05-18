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
package util

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
)

func TestMergeMaps(t *testing.T) {
	nestedMap := map[string]interface{}{
		"foo": "bar",
		"baz": map[string]string{
			"cool": "stuff",
		},
	}
	anotherNestedMap := map[string]interface{}{
		"foo": "bar",
		"baz": map[string]string{
			"cool":    "things",
			"awesome": "stuff",
		},
	}
	flatMap := map[string]interface{}{
		"foo": "bar",
		"baz": "stuff",
	}
	anotherFlatMap := map[string]interface{}{
		"testing": "fun",
	}

	testMap := MergeMaps(flatMap, nestedMap)
	equal := reflect.DeepEqual(testMap, nestedMap)
	if !equal {
		t.Errorf("Expected a nested map to overwrite a flat value. Expected: %v, got %v", nestedMap, testMap)
	}

	testMap = MergeMaps(nestedMap, flatMap)
	equal = reflect.DeepEqual(testMap, flatMap)
	if !equal {
		t.Errorf("Expected a flat value to overwrite a map. Expected: %v, got %v", flatMap, testMap)
	}

	testMap = MergeMaps(nestedMap, anotherNestedMap)
	equal = reflect.DeepEqual(testMap, anotherNestedMap)
	if !equal {
		t.Errorf("Expected a nested map to overwrite another nested map. Expected: %v, got %v", anotherNestedMap, testMap)
	}

	testMap = MergeMaps(anotherFlatMap, anotherNestedMap)
	expectedMap := map[string]interface{}{
		"testing": "fun",
		"foo":     "bar",
		"baz": map[string]string{
			"cool":    "things",
			"awesome": "stuff",
		},
	}
	equal = reflect.DeepEqual(testMap, expectedMap)
	if !equal {
		t.Errorf("Expected a map with different keys to merge properly with another map. Expected: %v, got %v", expectedMap, testMap)
	}
}

func TestValuesChecksum(t *testing.T) {
	tests := []struct {
		name   string
		values chartutil.Values
		want   string
	}{
		{
			name:   "empty",
			values: chartutil.Values{},
			want:   "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name: "value map",
			values: chartutil.Values{
				"foo": "bar",
				"baz": map[string]string{
					"cool": "stuff",
				},
			},
			want: "496605d01c7847477b215f8f2a24798e3b385863",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValuesChecksum(tt.values); got != tt.want {
				t.Errorf("ValuesChecksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReleaseRevision(t *testing.T) {
	var rel *release.Release
	if rev := ReleaseRevision(rel); rev != 0 {
		t.Fatalf("ReleaseRevision() = %v, want %v", rev, 0)
	}
	rel = &release.Release{Version: 1}
	if rev := ReleaseRevision(rel); rev != 1 {
		t.Fatalf("ReleaseRevision() = %v, want %v", rev, 1)
	}
}

func TestToUnstructured(t *testing.T) {
	type args struct {
		rawyaml []byte
	}
	tests := []struct {
		name          string
		args          args
		wantObjsCount int
		wantErr       bool
		err           string
	}{
		{
			name: "single object",
			args: args{
				rawyaml: []byte("apiVersion: v1\n" +
					"kind: ConfigMap\n"),
			},
			wantObjsCount: 1,
			wantErr:       false,
		},
		{
			name: "multiple objects are detected",
			args: args{
				rawyaml: []byte("apiVersion: v1\n" +
					"kind: ConfigMap\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"kind: Secret\n"),
			},
			wantObjsCount: 2,
			wantErr:       false,
		},
		{
			name: "empty object are dropped",
			args: args{
				rawyaml: []byte("---\n" + //empty objects before
					"---\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"kind: ConfigMap\n" +
					"---\n" + // empty objects in the middle
					"---\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"kind: Secret\n" +
					"---\n" + //empty objects after
					"---\n" +
					"---\n"),
			},
			wantObjsCount: 2,
			wantErr:       false,
		},
		{
			name: "--- in the middle of objects are ignored",
			args: args{
				[]byte("apiVersion: v1\n" +
					"kind: ConfigMap\n" +
					"data: \n" +
					" key: |\n" +
					"  ··Several lines of text,\n" +
					"  ··with some --- \n" +
					"  ---\n" +
					"  ··in the middle\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"kind: Secret\n"),
			},
			wantObjsCount: 2,
			wantErr:       false,
		},
		{
			name: "returns error for invalid yaml",
			args: args{
				rawyaml: []byte("apiVersion: v1\n" +
					"kind: ConfigMap\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"foobar\n" +
					"kind: Secret\n"),
			},
			wantErr: true,
			err:     "failed to unmarshal the 2nd yaml document",
		},
		{
			name: "returns error for invalid yaml",
			args: args{
				rawyaml: []byte("apiVersion: v1\n" +
					"kind: ConfigMap\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"kind: Pod\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"kind: Deployment\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"foobar\n" +
					"kind: ConfigMap\n"),
			},
			wantErr: true,
			err:     "failed to unmarshal the 4th yaml document",
		},
		{
			name: "returns error for invalid yaml",
			args: args{
				rawyaml: []byte("apiVersion: v1\n" +
					"foobar\n" +
					"kind: ConfigMap\n" +
					"---\n" +
					"apiVersion: v1\n" +
					"kind: Secret\n"),
			},
			wantErr: true,
			err:     "failed to unmarshal the 1st yaml document",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToUnstructured(tt.args.rawyaml)
			if tt.wantErr {
				if err == nil {
					t.Error("err is nil")
					return
				}
				if !strings.Contains(err.Error(), tt.err) {
					t.Errorf("expected %v to contains %s", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Errorf("expected nil, but got %v", err)
			}
			if len(got) != tt.wantObjsCount {
				t.Errorf("expected %d, but got %d", tt.wantObjsCount, len(got))
			}
		})
	}
}

func TestOrdinalize(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0th"},
		{1, "1st"},
		{2, "2nd"},
		{43, "43rd"},
		{5, "5th"},
		{6, "6th"},
		{207, "207th"},
		{1008, "1008th"},
		{-109, "-109th"},
		{-0, "0th"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ordinalize %d", tt.input), func(t *testing.T) {
			got := Ordinalize(tt.input)
			if got != tt.expected {
				t.Errorf("expected %s, but got %s", tt.expected, got)
			}
		})
	}
}

func TestRemoveDuplicateTaints(t *testing.T) {
	type args struct {
		taints []corev1.Taint
	}
	tests := []struct {
		name string
		args args
		want []corev1.Taint
	}{
		{
			name: "remove dupicate",
			args: args{
				taints: []corev1.Taint{
					{
						Key:    "dedicated",
						Value:  "infra",
						Effect: corev1.TaintEffectNoSchedule,
					},
					{
						Key:    "dedicated",
						Value:  "infra",
						Effect: corev1.TaintEffectNoSchedule,
					},
					{
						Key:    "test",
						Value:  "test",
						Effect: corev1.TaintEffectNoSchedule,
					},
				},
			},
			want: []corev1.Taint{
				{
					Key:    "dedicated",
					Value:  "infra",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "test",
					Value:  "test",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveDuplicateTaints(tt.args.taints); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicateTaints() = %v, want %v", got, tt.want)
			}
		})
	}
}
