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
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
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
