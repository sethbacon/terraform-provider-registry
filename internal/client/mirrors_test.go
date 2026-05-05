package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestUpdateMirrorRequest_OmitsUnsetFields guards against the silent zero-value
// PUT bug fixed in #8: backend treats omitted fields as "leave unchanged" but
// the previous Go struct used plain bool/int, so missing optionals shipped as
// `enabled: false` and `sync_interval_hours: 0`.
func TestUpdateMirrorRequest_OmitsUnsetFields(t *testing.T) {
	name := "renamed"
	req := UpdateMirrorRequest{Name: &name}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	wire := string(body)

	// Must contain only what was set.
	if !strings.Contains(wire, `"name":"renamed"`) {
		t.Errorf("payload missing name: %s", wire)
	}

	// Must NOT contain unset optional scalars — the fix for #8 is that these
	// get omitted entirely rather than shipped as zero values.
	for _, k := range []string{
		`"enabled"`,
		`"sync_interval_hours"`,
		`"upstream_registry_url"`,
		`"organization_id"`,
		`"description"`,
		`"version_filter"`,
	} {
		if strings.Contains(wire, k) {
			t.Errorf("unset field %s leaked into wire payload: %s", k, wire)
		}
	}
}

// TestUpdateTerraformMirrorRequest_OmitsUnsetFields covers the same guarantee
// for the Terraform/OpenTofu binary mirror update request.
func TestUpdateTerraformMirrorRequest_OmitsUnsetFields(t *testing.T) {
	name := "renamed"
	req := UpdateTerraformMirrorRequest{Name: &name}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	wire := string(body)

	if !strings.Contains(wire, `"name":"renamed"`) {
		t.Errorf("payload missing name: %s", wire)
	}

	for _, k := range []string{
		`"enabled"`,
		`"sync_interval_hours"`,
		`"gpg_verify"`,
		`"stable_only"`,
		`"tool"`,
		`"upstream_url"`,
	} {
		if strings.Contains(wire, k) {
			t.Errorf("unset field %s leaked into wire payload: %s", k, wire)
		}
	}
}

// TestMirror_DecodesPullThroughFields covers the addition in #14 — backend
// MirrorConfiguration carries pull_through_enabled and
// pull_through_cache_ttl_hours for opt-in pull-through caching of providers
// from the upstream registry. The previous client struct dropped both.
func TestMirror_DecodesPullThroughFields(t *testing.T) {
	body := []byte(`{
		"id": "11111111-1111-1111-1111-111111111111",
		"name": "hashicorp",
		"upstream_registry_url": "https://registry.terraform.io",
		"enabled": true,
		"sync_interval_hours": 24,
		"pull_through_enabled": true,
		"pull_through_cache_ttl_hours": 12,
		"created_at": "2026-04-30T00:00:00Z",
		"updated_at": "2026-04-30T00:00:00Z"
	}`)

	var m Mirror
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !m.PullThroughEnabled {
		t.Errorf("PullThroughEnabled = false, want true")
	}
	if m.PullThroughCacheTTLHours != 12 {
		t.Errorf("PullThroughCacheTTLHours = %d, want 12", m.PullThroughCacheTTLHours)
	}
}

// TestCreateMirrorRequest_RequiredFieldsPresent confirms required scalars
// (which are not pointers) still ship even when other fields are unset.
func TestCreateMirrorRequest_RequiredFieldsPresent(t *testing.T) {
	req := CreateMirrorRequest{
		Name:                "test",
		UpstreamRegistryURL: "https://registry.example.com",
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	wire := string(body)

	if !strings.Contains(wire, `"name":"test"`) {
		t.Errorf("required name missing: %s", wire)
	}
	if !strings.Contains(wire, `"upstream_registry_url":"https://registry.example.com"`) {
		t.Errorf("required upstream URL missing: %s", wire)
	}
	// Optional pointer scalars omitted.
	if strings.Contains(wire, `"enabled"`) {
		t.Errorf("unset enabled should be omitted: %s", wire)
	}
}
