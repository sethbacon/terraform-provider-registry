package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestModuleSCMLink_DecodesV1Fields covers the fix in #10 — backend's response
// model carries module_path and auto_publish_enabled, which the previous
// provider client struct was missing entirely.
func TestModuleSCMLink_DecodesV1Fields(t *testing.T) {
	body := []byte(`{
		"id": "11111111-1111-1111-1111-111111111111",
		"module_id": "22222222-2222-2222-2222-222222222222",
		"scm_provider_id": "33333333-3333-3333-3333-333333333333",
		"repository_owner": "acme",
		"repository_name": "vpc",
		"default_branch": "main",
		"module_path": "modules/vpc",
		"tag_pattern": "v*",
		"auto_publish_enabled": true,
		"created_at": "2026-04-30T00:00:00Z",
		"updated_at": "2026-04-30T00:00:00Z"
	}`)

	var l ModuleSCMLink
	if err := json.Unmarshal(body, &l); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if l.ModulePath != "modules/vpc" {
		t.Errorf("ModulePath = %q, want modules/vpc (json tag must be module_path)", l.ModulePath)
	}
	if !l.AutoPublish {
		t.Errorf("AutoPublish = false, want true (json tag must be auto_publish_enabled)")
	}
}

// TestCreateModuleSCMLinkRequest_UsesRepositoryPath asserts the asymmetric
// naming on the request side: backend's LinkSCMRequest accepts
// repository_path (the request renames module_path on persist) and
// auto_publish_enabled.
func TestCreateModuleSCMLinkRequest_UsesRepositoryPath(t *testing.T) {
	req := CreateModuleSCMLinkRequest{
		SCMProviderID:   "p",
		RepositoryOwner: "acme",
		RepositoryName:  "vpc",
		DefaultBranch:   "main",
		RepositoryPath:  "modules/vpc",
		AutoPublish:     true,
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	wire := string(body)

	if !strings.Contains(wire, `"repository_path":"modules/vpc"`) {
		t.Errorf("missing repository_path on the wire: %s", wire)
	}
	if !strings.Contains(wire, `"auto_publish_enabled":true`) {
		t.Errorf("missing auto_publish_enabled on the wire: %s", wire)
	}
	if strings.Contains(wire, `"module_path"`) {
		t.Errorf("request body should not use module_path (response-side name): %s", wire)
	}
}
