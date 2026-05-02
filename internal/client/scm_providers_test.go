package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSCMProvider_DecodesV1Fields guards against the model drift fixed in #9:
// backend's scm.SCMProvider now exposes organization_id, is_active, tenant_id,
// and client_id, while the previous provider model expected a phantom
// oauth_status field that does not exist in the backend.
func TestSCMProvider_DecodesV1Fields(t *testing.T) {
	body := []byte(`{
		"id": "11111111-1111-1111-1111-111111111111",
		"organization_id": "22222222-2222-2222-2222-222222222222",
		"provider_type": "github",
		"name": "main-github",
		"base_url": "https://github.example.com",
		"client_id": "abc123",
		"is_active": true,
		"created_at": "2026-04-30T00:00:00Z",
		"updated_at": "2026-04-30T00:00:00Z"
	}`)

	var p SCMProvider
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !p.IsActive {
		t.Errorf("IsActive = false, want true")
	}
	if p.OrganizationID == nil || *p.OrganizationID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("OrganizationID = %v, want pointer to UUID", p.OrganizationID)
	}
	if p.ClientID != "abc123" {
		t.Errorf("ClientID = %q, want abc123", p.ClientID)
	}
}

// TestUpdateSCMProviderRequest_OmitsUnsetFields confirms the pointer-typed
// update request omits unset optional fields from the wire payload, matching
// the backend semantics where missing fields are left unchanged.
func TestUpdateSCMProviderRequest_OmitsUnsetFields(t *testing.T) {
	name := "renamed"
	req := UpdateSCMProviderRequest{Name: &name}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	wire := string(body)

	if !strings.Contains(wire, `"name":"renamed"`) {
		t.Errorf("payload missing name: %s", wire)
	}

	for _, k := range []string{
		`"base_url"`,
		`"tenant_id"`,
		`"client_id"`,
		`"client_secret"`,
		`"webhook_secret"`,
		`"is_active"`,
	} {
		if strings.Contains(wire, k) {
			t.Errorf("unset field %s leaked into wire payload: %s", k, wire)
		}
	}
}
