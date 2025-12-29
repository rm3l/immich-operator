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
	"unicode"
)

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		override map[string]string
		expected map[string]string
	}{
		{
			name:     "both empty",
			base:     map[string]string{},
			override: map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "override empty",
			base: map[string]string{
				"key1": "value1",
			},
			override: map[string]string{},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "base empty",
			base: map[string]string{},
			override: map[string]string{
				"key1": "value1",
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "override takes precedence",
			base: map[string]string{
				"key1": "oldValue",
			},
			override: map[string]string{
				"key1": "newValue",
			},
			expected: map[string]string{
				"key1": "newValue",
			},
		},
		{
			name: "merge non-overlapping",
			base: map[string]string{
				"key1": "value1",
			},
			override: map[string]string{
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "complex merge",
			base: map[string]string{
				"app":     "immich",
				"version": "1.0",
			},
			override: map[string]string{
				"version": "2.0",
				"env":     "production",
			},
			expected: map[string]string{
				"app":     "immich",
				"version": "2.0",
				"env":     "production",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeMaps(tt.base, tt.override)
			if !mapsEqualStr(result, tt.expected) {
				t.Errorf("mergeMaps() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "length 8",
			length: 8,
		},
		{
			name:   "length 16",
			length: 16,
		},
		{
			name:   "length 32",
			length: 32,
		},
		{
			name:   "length 64",
			length: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := generateRandomPassword(tt.length)
			if err != nil {
				t.Errorf("generateRandomPassword() error = %v", err)
				return
			}
			if len(password) != tt.length {
				t.Errorf("generateRandomPassword() length = %d, expected %d", len(password), tt.length)
			}
			// Verify all characters are alphanumeric
			for _, c := range password {
				if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
					t.Errorf("generateRandomPassword() contains invalid character: %c", c)
				}
			}
		})
	}

	// Test uniqueness
	t.Run("passwords are unique", func(t *testing.T) {
		passwords := make(map[string]bool)
		for i := 0; i < 100; i++ {
			p, err := generateRandomPassword(32)
			if err != nil {
				t.Errorf("generateRandomPassword() error = %v", err)
				return
			}
			if passwords[p] {
				t.Errorf("generateRandomPassword() generated duplicate password")
				return
			}
			passwords[p] = true
		}
	})
}

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

// mapsEqualStr compares two string maps
func mapsEqualStr(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || va != vb {
			return false
		}
	}
	return true
}
