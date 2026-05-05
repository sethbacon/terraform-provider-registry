package client

import (
	"encoding/json"
	"testing"
)

func TestBackendVersion_Decode(t *testing.T) {
	body := []byte(`{
		"version": "1.0.0",
		"build_date": "2026-04-01T00:00:00Z",
		"api_version": "1",
		"default_language": "en",
		"protocols": ["terraform.io/v1"],
		"capabilities": ["modules", "providers"]
	}`)

	var v BackendVersion
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if v.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", v.Version)
	}
	if v.APIVersion != "1" {
		t.Errorf("APIVersion = %q, want 1", v.APIVersion)
	}
	if len(v.Protocols) != 1 || v.Protocols[0] != "terraform.io/v1" {
		t.Errorf("Protocols = %v, want [terraform.io/v1]", v.Protocols)
	}
}
