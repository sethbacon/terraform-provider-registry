package client

import (
	"encoding/json"
	"testing"
)

// TestOrganization_DecodesIdpFields covers the addition in #12 — backend
// model.Organization now carries idp_type and idp_name (nullable) for binding
// the org to a specific IdP. The previous client struct dropped both.
func TestOrganization_DecodesIdpFields(t *testing.T) {
	body := []byte(`{
		"id": "11111111-1111-1111-1111-111111111111",
		"name": "platform",
		"display_name": "Platform",
		"idp_type": "saml",
		"idp_name": "okta-prod",
		"created_at": "2026-04-30T00:00:00Z",
		"updated_at": "2026-04-30T00:00:00Z"
	}`)

	var o Organization
	if err := json.Unmarshal(body, &o); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if o.IdpType == nil || *o.IdpType != "saml" {
		t.Errorf("IdpType = %v, want pointer to 'saml'", o.IdpType)
	}
	if o.IdpName == nil || *o.IdpName != "okta-prod" {
		t.Errorf("IdpName = %v, want pointer to 'okta-prod'", o.IdpName)
	}
}
