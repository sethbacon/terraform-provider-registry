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

// TestUpdateTerraformMirrorRequest_GPGFields confirms custom_gpg_key and
// skip_gpg_verify are omitted when unset and present when set.
func TestUpdateTerraformMirrorRequest_GPGFields(t *testing.T) {
	key := "-----BEGIN PGP PUBLIC KEY BLOCK-----\ntest\n-----END PGP PUBLIC KEY BLOCK-----"
	skip := true
	req := UpdateTerraformMirrorRequest{
		CustomGPGKey:  &key,
		SkipGPGVerify: &skip,
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	wire := string(body)

	if !strings.Contains(wire, `"custom_gpg_key"`) {
		t.Errorf("custom_gpg_key missing from payload: %s", wire)
	}
	if !strings.Contains(wire, `"skip_gpg_verify":true`) {
		t.Errorf("skip_gpg_verify missing from payload: %s", wire)
	}

	// Unset fields must not appear.
	reqEmpty := UpdateTerraformMirrorRequest{}
	bodyEmpty, _ := json.Marshal(reqEmpty)
	wireEmpty := string(bodyEmpty)
	for _, k := range []string{`"custom_gpg_key"`, `"skip_gpg_verify"`} {
		if strings.Contains(wireEmpty, k) {
			t.Errorf("unset field %s leaked into wire payload: %s", k, wireEmpty)
		}
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
