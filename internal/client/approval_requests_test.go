package client

import (
	"encoding/json"
	"testing"
)

// TestApprovalRequest_DecodesV1Fields covers the field expansion in #11 —
// backend's MirrorApprovalRequest now returns organization_id, requested_by,
// reviewed_at, expires_at, plus joined requested_by_name / reviewed_by_name /
// mirror_name fields, none of which the previous client struct knew about.
func TestApprovalRequest_DecodesV1Fields(t *testing.T) {
	body := []byte(`{
		"id": "11111111-1111-1111-1111-111111111111",
		"mirror_config_id": "22222222-2222-2222-2222-222222222222",
		"organization_id": "33333333-3333-3333-3333-333333333333",
		"requested_by": "44444444-4444-4444-4444-444444444444",
		"provider_namespace": "hashicorp",
		"provider_name": "aws",
		"reason": "needed for prod",
		"status": "approved",
		"reviewed_by": "55555555-5555-5555-5555-555555555555",
		"reviewed_at": "2026-04-30T01:00:00Z",
		"review_notes": "ok",
		"auto_approved": false,
		"created_at": "2026-04-30T00:00:00Z",
		"updated_at": "2026-04-30T01:00:00Z",
		"expires_at": "2027-04-30T00:00:00Z",
		"requested_by_name": "Alice",
		"reviewed_by_name": "Bob",
		"mirror_name": "hashicorp-prod"
	}`)

	var ar ApprovalRequest
	if err := json.Unmarshal(body, &ar); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ar.OrganizationID == nil || *ar.OrganizationID != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("OrganizationID = %v, want UUID pointer", ar.OrganizationID)
	}
	if ar.RequestedBy == nil || *ar.RequestedBy != "44444444-4444-4444-4444-444444444444" {
		t.Errorf("RequestedBy = %v", ar.RequestedBy)
	}
	if ar.ReviewedAt == nil || *ar.ReviewedAt != "2026-04-30T01:00:00Z" {
		t.Errorf("ReviewedAt = %v", ar.ReviewedAt)
	}
	if ar.ExpiresAt == nil || *ar.ExpiresAt != "2027-04-30T00:00:00Z" {
		t.Errorf("ExpiresAt = %v", ar.ExpiresAt)
	}
	if ar.RequestedByName != "Alice" {
		t.Errorf("RequestedByName = %q", ar.RequestedByName)
	}
	if ar.ReviewedByName != "Bob" {
		t.Errorf("ReviewedByName = %q", ar.ReviewedByName)
	}
	if ar.MirrorName != "hashicorp-prod" {
		t.Errorf("MirrorName = %q", ar.MirrorName)
	}
}
