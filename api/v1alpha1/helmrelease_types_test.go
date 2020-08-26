/*
Copyright 2020 Getup Cloud. All rights reserved.
*/

package v1alpha1

import (
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
)

func TestHelmValues(t *testing.T) {
	testCases := []struct {
		original         *HelmValues
		transformer      func(v *HelmValues) *HelmValues
		expectedCopy     *HelmValues
		expectedOriginal *HelmValues
	}{
		{
			original: &HelmValues{Data: map[string]interface{}{}},
			transformer: func(v *HelmValues) *HelmValues {
				v.Data["foo"] = "bar"
				return v
			},
			expectedCopy:     &HelmValues{Data: map[string]interface{}{"foo": "bar"}},
			expectedOriginal: &HelmValues{Data: map[string]interface{}{}},
		},
		{
			original: &HelmValues{Data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}},
			transformer: func(v *HelmValues) *HelmValues {
				v.Data["foo"] = map[string]interface{}{"bar": "oof"}
				return v
			},
			expectedCopy:     &HelmValues{Data: map[string]interface{}{"foo": map[string]interface{}{"bar": "oof"}}},
			expectedOriginal: &HelmValues{Data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := NewWithT(t)
			var out HelmValues
			tc.original.DeepCopyInto(&out)
			g.Expect(tc.transformer(&out)).To(Equal(tc.expectedCopy))
			g.Expect(tc.original).To(Equal(tc.expectedOriginal))
		})
	}
}

func TestRefOrDefault(t *testing.T) {
	testCases := []struct {
		chartSource      GitChartSource
		potentialDefault string
		expected         string
	}{
		{
			chartSource: GitChartSource{
				Ref: "master",
			},
			potentialDefault: "dev",
			expected:         "master",
		},
		{
			chartSource:      GitChartSource{},
			potentialDefault: "dev",
			expected:         "dev",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := NewWithT(t)
			got := tc.chartSource.RefOrDefault(tc.potentialDefault)
			g.Expect(got).To(Equal(tc.expected))
		})
	}
}
