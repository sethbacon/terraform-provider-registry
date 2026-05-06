package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetModuleVersion_FoundReturnsDeprecationFields verifies that the
// defensive lookup picks the correct entry out of the module-detail
// `versions[]` payload and surfaces the deprecation-related fields.
func TestGetModuleVersion_FoundReturnsDeprecationFields(t *testing.T) {
	expectedPath := "/api/v1/modules/acme/vpc/aws"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected request path: got %q, want %q", r.URL.Path, expectedPath)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "11111111-1111-1111-1111-111111111111",
			"namespace": "acme",
			"name": "vpc",
			"system": "aws",
			"versions": [
				{
					"id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
					"version": "1.0.0",
					"deprecated": false
				},
				{
					"id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
					"version": "1.1.0",
					"deprecated": true,
					"deprecated_at": "2026-05-01T00:00:00Z",
					"deprecation_message": "use 1.2.0",
					"replacement_source": "acme/vpc/aws"
				}
			]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	mv, err := c.GetModuleVersion(context.Background(), "acme", "vpc", "aws", "1.1.0")
	if err != nil {
		t.Fatalf("GetModuleVersion: %v", err)
	}
	if mv.Version != "1.1.0" {
		t.Errorf("Version = %q, want %q", mv.Version, "1.1.0")
	}
	if !mv.Deprecated {
		t.Errorf("Deprecated = false, want true")
	}
	if mv.DeprecationMessage == nil || *mv.DeprecationMessage != "use 1.2.0" {
		t.Errorf("DeprecationMessage = %v", mv.DeprecationMessage)
	}
	if mv.DeprecatedAt == nil || *mv.DeprecatedAt != "2026-05-01T00:00:00Z" {
		t.Errorf("DeprecatedAt = %v", mv.DeprecatedAt)
	}
	if mv.ReplacementSource == nil || *mv.ReplacementSource != "acme/vpc/aws" {
		t.Errorf("ReplacementSource = %v", mv.ReplacementSource)
	}
}

// TestGetModuleVersion_VersionNotInListReturnsNotFound verifies that when the
// module exists but the requested version is not in its `versions[]` list,
// a 404-style APIError is returned so IsNotFound recognises it.
func TestGetModuleVersion_VersionNotInListReturnsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "11111111-1111-1111-1111-111111111111",
			"namespace": "acme",
			"name": "vpc",
			"system": "aws",
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

	mv, err := c.GetModuleVersion(context.Background(), "acme", "vpc", "aws", "9.9.9")
	if err == nil {
		t.Fatalf("GetModuleVersion: expected error, got nil (mv=%+v)", mv)
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound(err) = false, want true (err=%v)", err)
	}
	if mv != nil {
		t.Errorf("mv = %+v, want nil", mv)
	}
}

// TestGetModuleVersion_ModuleNotFoundPropagates verifies that when the
// underlying module-detail GET returns 404, the same not-found error is
// propagated to the caller (so resource Read can RemoveResource correctly).
func TestGetModuleVersion_ModuleNotFoundPropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Module not found"}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	mv, err := c.GetModuleVersion(context.Background(), "missing", "mod", "aws", "1.0.0")
	if err == nil {
		t.Fatalf("GetModuleVersion: expected error, got nil (mv=%+v)", mv)
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound(err) = false, want true (err=%v)", err)
	}
	if mv != nil {
		t.Errorf("mv = %+v, want nil", mv)
	}
}
