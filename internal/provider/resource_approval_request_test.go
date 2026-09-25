package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/terraform-registry/terraform-provider-registry/internal/client"
	"github.com/terraform-registry/terraform-provider-registry/internal/provider"
)

func TestAccApprovalRequest_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApprovalRequestConfig("hashicorp", "Need this mirror for prod"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_approval_request.test", "id"),
					resource.TestCheckResourceAttrSet("registry_approval_request.test", "mirror_id"),
					resource.TestCheckResourceAttr("registry_approval_request.test", "provider_namespace", "hashicorp"),
					resource.TestCheckResourceAttr("registry_approval_request.test", "justification", "Need this mirror for prod"),
					resource.TestCheckResourceAttr("registry_approval_request.test", "review_status", "pending"),
					resource.TestCheckResourceAttrSet("registry_approval_request.test", "created_at"),
				),
			},
			{
				ResourceName:      "registry_approval_request.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccApprovalRequestConfig(providerNamespace, justification string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_mirror" "test" {
  name                  = "acc-approval-mirror"
  upstream_registry_url = "https://registry.terraform.io"
}

resource "registry_approval_request" "test" {
  mirror_id          = registry_mirror.test.id
  provider_namespace = %q
  justification      = %q
}
`, providerNamespace, justification)
}

// TestApprovalRequestDelete_StateOnly pins that destroy sends nothing to the
// registry. Delete used to send DELETE /api/v1/admin/approvals/{id}, a route
// the backend has never served; Client.Delete treats a 404 on delete as
// success, so destroy "succeeded" while the request stayed pending. There is
// no delete or withdraw endpoint to call, so Delete must make no request at
// all, finish without an error (which is what drops it from state) and warn
// that the request is still in the registry.
func TestApprovalRequestDelete_StateOnly(t *testing.T) {
	const id = "11111111-1111-1111-1111-111111111111"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Delete must not call the registry, got %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c, err := client.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	r := testConfiguredResource(t, provider.NewApprovalRequestResource(), c)

	state := testResourceState(t, r, provider.ApprovalRequestResourceModel{
		ID:                types.StringValue(id),
		MirrorID:          types.StringValue("22222222-2222-2222-2222-222222222222"),
		ProviderNamespace: types.StringValue("hashicorp"),
		ReviewStatus:      types.StringValue("pending"),
	})
	resp := fwresource.DeleteResponse{State: state}
	r.Delete(context.Background(), fwresource.DeleteRequest{State: state}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete: unexpected error: %v", resp.Diagnostics)
	}
	warnings := resp.Diagnostics.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("warnings = %d, want 1 (%v)", len(warnings), resp.Diagnostics)
	}
	if detail := warnings[0].Detail(); !strings.Contains(detail, id) || !strings.Contains(detail, `"pending"`) {
		t.Errorf("warning detail should name the request and its status, got %q", detail)
	}
}
