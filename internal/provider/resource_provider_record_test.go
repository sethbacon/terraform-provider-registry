package provider_test

import (
	"testing"
)

func TestAccProviderRecord_basic(t *testing.T) {
	// Skip: the backend has no dedicated JSON create endpoint for provider records.
	// Providers are created implicitly via multipart file upload at POST /api/v1/providers.
	t.Skip("Provider records cannot be created via JSON API; requires multipart upload")
}
