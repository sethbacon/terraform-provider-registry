package client

import (
	"encoding/json"
	"testing"
)

// TestStorageConfig_DecodesS3IAMFields confirms new S3 IAM-role fields and
// credential _set indicators decode correctly from the backend response.
func TestStorageConfig_DecodesS3IAMFields(t *testing.T) {
	body := []byte(`{
		"id": "aaaa-0000",
		"backend_type": "s3",
		"is_active": false,
		"s3_region": "us-east-1",
		"s3_bucket": "my-bucket",
		"s3_auth_method": "iam_role",
		"s3_role_arn": "arn:aws:iam::123456789012:role/my-role",
		"s3_role_session_name": "registry",
		"s3_external_id": "ext-123",
		"s3_web_identity_token_file": "/var/run/secrets/token",
		"s3_access_key_id_set": true,
		"s3_secret_access_key_set": false,
		"azure_account_key_set": false,
		"gcs_credentials_json_set": false,
		"created_at": "2026-01-01T00:00:00Z",
		"updated_at": "2026-01-01T00:00:00Z"
	}`)

	var sc StorageConfig
	if err := json.Unmarshal(body, &sc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if sc.S3AuthMethod == nil || *sc.S3AuthMethod != "iam_role" {
		t.Errorf("S3AuthMethod = %v, want iam_role", sc.S3AuthMethod)
	}
	if sc.S3RoleARN == nil || *sc.S3RoleARN != "arn:aws:iam::123456789012:role/my-role" {
		t.Errorf("S3RoleARN missing or wrong")
	}
	if sc.S3ExternalID == nil || *sc.S3ExternalID != "ext-123" {
		t.Errorf("S3ExternalID missing or wrong")
	}
	if sc.S3WebIdentityTokenFile == nil || *sc.S3WebIdentityTokenFile != "/var/run/secrets/token" {
		t.Errorf("S3WebIdentityTokenFile missing or wrong")
	}
	if !sc.S3AccessKeyIDSet {
		t.Errorf("S3AccessKeyIDSet = false, want true")
	}
	if sc.S3SecretAccessKeySet {
		t.Errorf("S3SecretAccessKeySet = true, want false")
	}
}

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
