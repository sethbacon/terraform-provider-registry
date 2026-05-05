package client

import (
	"encoding/json"
	"testing"
)

// TestModule_DecodesDeprecationFields covers the addition in #13 — backend
// model.Module carries module-level deprecation state (deprecated, deprecated_at,
// deprecation_message, successor_module_id) which the previous client struct
// dropped silently.
func TestModule_DecodesDeprecationFields(t *testing.T) {
	body := []byte(`{
		"id": "11111111-1111-1111-1111-111111111111",
		"organization_id": "22222222-2222-2222-2222-222222222222",
		"namespace": "acme",
		"name": "vpc",
		"system": "aws",
		"deprecated": true,
		"deprecated_at": "2026-04-30T00:00:00Z",
		"deprecation_message": "Use acme/networking/aws instead",
		"successor_module_id": "33333333-3333-3333-3333-333333333333",
		"created_at": "2026-01-01T00:00:00Z",
		"updated_at": "2026-04-30T00:00:00Z"
	}`)

	var m Module
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !m.Deprecated {
		t.Errorf("Deprecated = false, want true")
	}
	if m.DeprecatedAt == nil || *m.DeprecatedAt != "2026-04-30T00:00:00Z" {
		t.Errorf("DeprecatedAt = %v", m.DeprecatedAt)
	}
	if m.DeprecationMessage == nil || *m.DeprecationMessage != "Use acme/networking/aws instead" {
		t.Errorf("DeprecationMessage = %v", m.DeprecationMessage)
	}
	if m.SuccessorModuleID == nil || *m.SuccessorModuleID != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("SuccessorModuleID = %v", m.SuccessorModuleID)
	}
}
