package config

import (
	"reflect"
	"testing"
)

func TestProcessLabels(t *testing.T) {
	tests := []struct {
		name           string
		labelsStr      string
		expectedErr    bool
		expectedResult map[string]string
	}{
		{
			name:        "single valid label",
			labelsStr:   "acyl.dev/managed-by=nitro",
			expectedErr: false,
			expectedResult: map[string]string{
				"acyl.dev/managed-by": "nitro",
			},
		},
		{
			name:        "multiple valid labels",
			labelsStr:   "acyl.dev/managed-by=nitro,istio-injection=enabled",
			expectedErr: false,
			expectedResult: map[string]string{
				"acyl.dev/managed-by": "nitro",
				"istio-injection":     "enabled",
			},
		},
		{
			name:           "no label",
			labelsStr:      "",
			expectedErr:    true,
			expectedResult: map[string]string{},
		},
		{
			name:           "invalid labels",
			labelsStr:      "potato,onion",
			expectedErr:    true,
			expectedResult: map[string]string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var k8scfg K8sConfig
			err := k8scfg.ProcessLabels(tc.labelsStr)
			receivedErr := err != nil
			if receivedErr != tc.expectedErr {
				t.Fatalf("K8sConfig.ProcessLabels received unexpected error: %v", err)
			}
			if !reflect.DeepEqual(k8scfg.Labels, tc.expectedResult) {
				t.Fatalf("K8sConfig.ProcessLabels received unexpected results: %v", k8scfg.Labels)
			}
		})
	}
}
