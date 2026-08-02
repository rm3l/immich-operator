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

	"k8s.io/utils/ptr"

	mediav1alpha1 "github.com/rm3l/immich-operator/api/v1alpha1"
)

// TestConfigSpecToMapPreservesCase ensures camelCase config keys (e.g.
// oauth.issuerUrl) are not lower-cased when the ConfigurationSpec is
// converted to a map for the generated ConfigMap/Secret. Immich rejects
// some settings when their key casing does not match exactly.
// See https://github.com/rm3l/immich-operator/issues/44.
func TestConfigSpecToMapPreservesCase(t *testing.T) {
	r := &ImmichReconciler{}

	spec := &mediav1alpha1.ConfigurationSpec{
		OAuth: &mediav1alpha1.OAuthConfig{
			Enabled:               ptr.To(true),
			IssuerURL:             ptr.To("https://example.com/auth"),
			ClientID:              ptr.To("client-id"),
			StorageLabel:          ptr.To("storageLabel"),
			StorageQuota:          ptr.To("storageQuota"),
			MobileOverrideEnabled: ptr.To(true),
			MobileRedirectURI:     ptr.To("app://callback"),
		},
		StorageTemplate: &mediav1alpha1.StorageTemplateConfig{
			Enabled:  ptr.To(true),
			Template: ptr.To("{{y}}/{{MM}}/{{filename}}"),
		},
	}

	result := r.configSpecToMap(spec)

	oauth, ok := result["oauth"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected oauth key in result, got: %v", result)
	}

	expectedOAuthKeys := []string{
		"enabled",
		"issuerUrl",
		"clientId",
		"storageLabelClaim",
		"storageQuotaClaim",
		"mobileOverrideEnabled",
		"mobileRedirectUri",
	}
	for _, key := range expectedOAuthKeys {
		if _, ok := oauth[key]; !ok {
			t.Errorf("expected oauth config to contain camelCase key %q, got keys: %v", key, mapKeys(oauth))
		}
	}

	if v, ok := oauth["issuerUrl"]; !ok || v != "https://example.com/auth" {
		t.Errorf("oauth.issuerUrl = %v, want %q", v, "https://example.com/auth")
	}

	storageTemplate, ok := result["storageTemplate"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected storageTemplate key in result, got: %v", result)
	}
	if _, ok := storageTemplate["template"]; !ok {
		t.Errorf("expected storageTemplate config to contain key %q, got keys: %v", "template", mapKeys(storageTemplate))
	}

	// Guard against regressing to lower-cased keys.
	lowerCasedKeys := []string{"issuerurl", "clientid", "storagelabelclaim", "storagequotaclaim", "mobileoverrideenabled", "mobileredirecturi"}
	for _, key := range lowerCasedKeys {
		if _, ok := oauth[key]; ok {
			t.Errorf("did not expect lower-cased key %q in oauth config", key)
		}
	}
}

// TestConfigSpecToMapExcludesNilFields ensures unset fields are omitted
// from the resulting map, matching the previous YAML-based behavior.
func TestConfigSpecToMapExcludesNilFields(t *testing.T) {
	r := &ImmichReconciler{}

	spec := &mediav1alpha1.ConfigurationSpec{
		OAuth: &mediav1alpha1.OAuthConfig{
			IssuerURL: ptr.To("https://example.com/auth"),
		},
	}

	result := r.configSpecToMap(spec)

	if _, ok := result["storageTemplate"]; ok {
		t.Errorf("expected storageTemplate to be omitted, got: %v", result)
	}

	oauth, ok := result["oauth"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected oauth key in result, got: %v", result)
	}
	if _, ok := oauth["clientId"]; ok {
		t.Errorf("expected oauth.clientId to be omitted, got: %v", oauth)
	}
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
