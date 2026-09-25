package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetProviderVersion_ReadsProviderDetail pins the route GetProviderVersion
// calls. It used to GET /api/v1/providers/{ns}/{type}/versions/{version},
// which the backend has never served (that path is DELETE-only), so every
// refresh of registry_provider_version_deprecation got a 404 and dropped the
// resource from state. The only read that carries version deprecation state is
// the provider-detail GET, so the request must be exactly that.
func TestGetProviderVersion_ReadsProviderDetail(t *testing.T) {
	expectedPath := "/api/v1/providers/acme/aws"

	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodGet || r.URL.Path != expectedPath {
			t.Errorf("unexpected request: got %s %s, want GET %s", r.Method, r.URL.Path, expectedPath)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "11111111-1111-1111-1111-111111111111",
			"namespace": "acme",
			"type": "aws",
			"description": "AWS provider",
			"source": "https://github.com/acme/terraform-provider-aws",
			"versions": [
				{
					"id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"version": "1.0.0",
					"protocols": ["6.0"],
					"platforms": [
						{ "os": "linux", "arch": "amd64", "filename": "p.zip", "shasum": "abc", "download_count": 3 }
					],
					"deprecated": false,
					"signed": true,
					"created_at": "2026-04-01T00:00:00Z"
				},
				{
					"id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"version": "1.1.0",
					"protocols": ["6.0"],
					"platforms": [],
					"deprecated": true,
					"signed": false,
					"created_at": "2026-04-15T00:00:00Z",
					"deprecated_at": "2026-05-01T00:00:00Z",
					"deprecation_message": "use 1.2.0"
				}
			],
			"created_at": "2026-04-01T00:00:00Z",
			"updated_at": "2026-05-01T00:00:00Z"
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pv, err := c.GetProviderVersion(context.Background(), "acme", "aws", "1.1.0")
	if err != nil {
		t.Fatalf("GetProviderVersion: %v", err)
	}
	if requests != 1 {
		t.Errorf("requests = %d, want 1", requests)
	}
	if pv.ID != "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb" {
		t.Errorf("ID = %q, want the 1.1.0 entry", pv.ID)
	}
	if pv.Version != "1.1.0" {
		t.Errorf("Version = %q, want %q", pv.Version, "1.1.0")
	}
	if !pv.Deprecated {
		t.Errorf("Deprecated = false, want true")
	}
	if pv.DeprecationMessage == nil || *pv.DeprecationMessage != "use 1.2.0" {
		t.Errorf("DeprecationMessage = %v", pv.DeprecationMessage)
	}
	if pv.DeprecatedAt == nil || *pv.DeprecatedAt != "2026-05-01T00:00:00Z" {
		t.Errorf("DeprecatedAt = %v", pv.DeprecatedAt)
	}
}

// TestGetProviderVersion_NotDeprecatedVersion verifies that a version the
// backend reports as not deprecated comes back with Deprecated=false and no
// deprecation fields (the backend omits deprecated_at/deprecation_message
// rather than sending null), which is what lets the resource Read tell an
// out-of-band undeprecation apart from a missing version.
func TestGetProviderVersion_NotDeprecatedVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "11111111-1111-1111-1111-111111111111",
			"namespace": "acme",
			"type": "aws",
			"versions": [
				{ "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "version": "1.0.0", "deprecated": false }
			]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pv, err := c.GetProviderVersion(context.Background(), "acme", "aws", "1.0.0")
	if err != nil {
		t.Fatalf("GetProviderVersion: %v", err)
	}
	if pv.Deprecated {
		t.Errorf("Deprecated = true, want false")
	}
	if pv.DeprecatedAt != nil || pv.DeprecationMessage != nil {
		t.Errorf("DeprecatedAt = %v, DeprecationMessage = %v, want both nil", pv.DeprecatedAt, pv.DeprecationMessage)
	}
}

// TestGetProviderVersion_VersionNotInListReturnsNotFound verifies that when
// the provider exists but the requested version is not in its `versions[]`
// list, a 404-style APIError is returned so IsNotFound recognises it.
func TestGetProviderVersion_VersionNotInListReturnsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "11111111-1111-1111-1111-111111111111",
			"namespace": "acme",
			"type": "aws",
			"versions": [
				{ "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "version": "1.0.0", "deprecated": false }
			]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pv, err := c.GetProviderVersion(context.Background(), "acme", "aws", "9.9.9")
	if err == nil {
		t.Fatalf("GetProviderVersion: expected error, got nil (pv=%+v)", pv)
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound(err) = false, want true (err=%v)", err)
	}
	if pv != nil {
		t.Errorf("pv = %+v, want nil", pv)
	}
}

// TestGetProviderVersion_ProviderNotFoundPropagates verifies that when the
// underlying provider-detail GET returns 404, the same not-found error is
// propagated to the caller (so resource Read can RemoveResource correctly).
func TestGetProviderVersion_ProviderNotFoundPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Provider not found"}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	pv, err := c.GetProviderVersion(context.Background(), "missing", "prov", "1.0.0")
	if err == nil {
		t.Fatalf("GetProviderVersion: expected error, got nil (pv=%+v)", pv)
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound(err) = false, want true (err=%v)", err)
	}
	if pv != nil {
		t.Errorf("pv = %+v, want nil", pv)
	}
}
