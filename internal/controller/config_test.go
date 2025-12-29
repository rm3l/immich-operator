/*
Copyright 2025.

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

package controller

import (
	"testing"
)

func TestRemoveNullValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "no null values",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
		{
			name: "with null values",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": nil,
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key3": "value3",
			},
		},
		{
			name: "nested map with null values",
			input: map[string]interface{}{
				"key1": "value1",
				"nested": map[string]interface{}{
					"inner1": "innerValue",
					"inner2": nil,
				},
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"nested": map[string]interface{}{
					"inner1": "innerValue",
				},
			},
		},
		{
			name: "nested map becomes empty after removal",
			input: map[string]interface{}{
				"key1": "value1",
				"nested": map[string]interface{}{
					"inner": nil,
				},
			},
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name: "deeply nested",
			input: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"keep":   "value",
						"remove": nil,
					},
				},
			},
			expected: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"keep": "value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeNullValues(tt.input)
			if !mapsEqual(tt.input, tt.expected) {
				t.Errorf("removeNullValues() = %v, expected %v", tt.input, tt.expected)
			}
		})
	}
}

func TestDeepMergeMap(t *testing.T) {
	tests := []struct {
		name     string
		dst      map[string]interface{}
		src      map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "both empty",
			dst:      map[string]interface{}{},
			src:      map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name: "src empty",
			dst: map[string]interface{}{
				"key1": "value1",
			},
			src: map[string]interface{}{},
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name: "dst empty",
			dst:  map[string]interface{}{},
			src: map[string]interface{}{
				"key1": "value1",
			},
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name: "src overrides dst",
			dst: map[string]interface{}{
				"key1": "oldValue",
			},
			src: map[string]interface{}{
				"key1": "newValue",
			},
			expected: map[string]interface{}{
				"key1": "newValue",
			},
		},
		{
			name: "merge non-overlapping keys",
			dst: map[string]interface{}{
				"key1": "value1",
			},
			src: map[string]interface{}{
				"key2": "value2",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "deep merge nested maps",
			dst: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
			},
			src: map[string]interface{}{
				"nested": map[string]interface{}{
					"key2": "newValue2",
					"key3": "value3",
				},
			},
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "value1",
					"key2": "newValue2",
					"key3": "value3",
				},
			},
		},
		{
			name: "src nil value ignored",
			dst: map[string]interface{}{
				"key1": "value1",
			},
			src: map[string]interface{}{
				"key1": nil,
			},
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name: "src replaces non-map with map",
			dst: map[string]interface{}{
				"key1": "value1",
			},
			src: map[string]interface{}{
				"key1": map[string]interface{}{
					"nested": "value",
				},
			},
			expected: map[string]interface{}{
				"key1": map[string]interface{}{
					"nested": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepMergeMap(tt.dst, tt.src)
			if !mapsEqual(result, tt.expected) {
				t.Errorf("deepMergeMap() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// mapsEqual compares two maps recursively
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		aMap, aIsMap := va.(map[string]interface{})
		bMap, bIsMap := vb.(map[string]interface{})
		if aIsMap && bIsMap {
			if !mapsEqual(aMap, bMap) {
				return false
			}
		} else if aIsMap != bIsMap {
			return false
		} else if va != vb {
			return false
		}
	}
	return true
}
