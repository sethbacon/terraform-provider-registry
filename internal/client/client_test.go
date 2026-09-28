package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDo_SetsOrganizationIDHeaderFromBody guards the X-Organization-Id
// header the backend's shared tenant-scope rule reads as a fallback when a
// create request names no organization any other way (backend #1012): the
// client forwards whatever organization_id a request body carries so a
// platform-admin caller with no in-scope organization of its own is not
// refused as ambiguous.
func TestDo_SetsOrganizationIDHeaderFromBody(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Organization-Id")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	orgID := "22222222-2222-2222-2222-222222222222"
	body := struct {
		OrganizationID *string `json:"organization_id"`
	}{OrganizationID: &orgID}

	resp, err := c.Do(context.Background(), http.MethodPost, "/x", body)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	_ = resp.Body.Close()

	if gotHeader != orgID {
		t.Errorf("X-Organization-Id = %q, want %q", gotHeader, orgID)
	}
}

// TestDo_OmitsOrganizationIDHeaderWhenAbsent covers the complementary case:
// a request body with no organization_id field (or a nil one) must not send
// the header at all, since an empty header value is itself meaningful to
// the backend's resolution order (it falls through to the caller's
// single in-scope organization) rather than "no opinion".
func TestDo_OmitsOrganizationIDHeaderWhenAbsent(t *testing.T) {
	tests := []struct {
		name string
		body interface{}
	}{
		{"no body", nil},
		{"nil organization_id", struct {
			OrganizationID *string `json:"organization_id"`
		}{}},
		{"no organization_id field", struct {
			Name string `json:"name"`
		}{Name: "acc-test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sawHeader bool
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				sawHeader = r.Header.Get("X-Organization-Id") != ""
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c, err := NewClient(srv.URL, "test-token")
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			resp, err := c.Do(context.Background(), http.MethodPost, "/x", tt.body)
			if err != nil {
				t.Fatalf("Do: %v", err)
			}
			_ = resp.Body.Close()

			if sawHeader {
				t.Errorf("X-Organization-Id header sent, want omitted")
			}
		})
	}
}
