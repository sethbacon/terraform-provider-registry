package provider_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
	"github.com/terraform-registry/terraform-provider-registry/internal/provider"
)

// TestProviderVersionDeprecationRead drives Read against an httptest backend
// that serves only the provider-detail route. Read used to GET
// /api/v1/providers/{ns}/{type}/versions/{version}, which the backend has never
// served, and it read the resulting 404 as "gone" — so a deprecation dropped
// out of state on every refresh and was re-created on every plan. The
// "deprecated" case is that regression; the other two keep the legitimate
// removal paths (undeprecated out of band, version deleted) working.
func TestProviderVersionDeprecationRead(t *testing.T) {
	const detailPath = "/api/v1/providers/acme/aws"
	const detailBody = `{
		"id": "11111111-1111-1111-1111-111111111111",
		"namespace": "acme",
		"type": "aws",
		"versions": [
			{ "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "version": "1.0.0", "deprecated": false },
			{
				"id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
				"version": "1.1.0",
				"deprecated": true,
				"deprecated_at": "2026-05-01T00:00:00Z",
				"deprecation_message": "use 1.2.0"
			}
		]
	}`

	cases := []struct {
		name        string
		version     string
		wantRemoved bool
	}{
		{name: "deprecated version stays in state", version: "1.1.0"},
		{name: "undeprecated out of band is removed", version: "1.0.0", wantRemoved: true},
		{name: "deleted version is removed", version: "9.9.9", wantRemoved: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != detailPath {
					t.Errorf("unexpected request: got %s %s, want GET %s", r.Method, r.URL.Path, detailPath)
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(detailBody))
			}))
			defer srv.Close()

			c, err := client.NewClient(srv.URL, "test-token")
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			r := testConfiguredResource(t, provider.NewProviderVersionDeprecationResource(), c)

			// deprecated_at is empty in state because the deprecate endpoint's
			// response does not carry it; the first refresh fills it in.
			state := testResourceState(t, r, provider.ProviderVersionDeprecationResourceModel{
				Namespace:    types.StringValue("acme"),
				Type:         types.StringValue("aws"),
				Version:      types.StringValue(tc.version),
				Message:      types.StringValue("use 1.2.0"),
				DeprecatedAt: types.StringValue(""),
			})
			resp := fwresource.ReadResponse{State: state}
			r.Read(context.Background(), fwresource.ReadRequest{State: state}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Read: %v", resp.Diagnostics)
			}

			if removed := resp.State.Raw.IsNull(); removed != tc.wantRemoved {
				t.Fatalf("resource removed from state = %t, want %t", removed, tc.wantRemoved)
			}
			if tc.wantRemoved {
				return
			}

			var got provider.ProviderVersionDeprecationResourceModel
			if diags := resp.State.Get(context.Background(), &got); diags.HasError() {
				t.Fatalf("State.Get: %v", diags)
			}
			if got.Version.ValueString() != "1.1.0" {
				t.Errorf("version = %q, want %q", got.Version.ValueString(), "1.1.0")
			}
			if got.Message.ValueString() != "use 1.2.0" {
				t.Errorf("message = %q, want %q", got.Message.ValueString(), "use 1.2.0")
			}
			if got.DeprecatedAt.ValueString() != "2026-05-01T00:00:00Z" {
				t.Errorf("deprecated_at = %q, want %q", got.DeprecatedAt.ValueString(), "2026-05-01T00:00:00Z")
			}
		})
	}
}
