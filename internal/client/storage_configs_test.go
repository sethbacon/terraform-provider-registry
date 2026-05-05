package client

import (
	"encoding/json"
	"testing"
)

// TestStorageConfig_DecodesIsActive guards against the JSON-tag drift fixed
// in #6: backend renamed "active" to "is_active" before the provider's first
// release, leaving the resource showing perpetual diff on the active attribute.
func TestStorageConfig_DecodesIsActive(t *testing.T) {
	// Sample response shape taken from backend StorageConfigResponse
	// (backend/internal/db/models/storage_config.go).
	body := []byte(`{
		"id": "11111111-1111-1111-1111-111111111111",
		"backend_type": "s3",
		"is_active": true,
		"s3_region": "us-east-1",
		"s3_bucket": "example",
		"created_at": "2026-04-30T00:00:00Z",
		"updated_at": "2026-04-30T00:00:00Z"
	}`)

	var sc StorageConfig
	if err := json.Unmarshal(body, &sc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !sc.IsActive {
		t.Errorf("IsActive = false, want true (JSON tag must be is_active)")
	}
	if sc.BackendType != "s3" {
		t.Errorf("BackendType = %q, want s3", sc.BackendType)
	}
}
