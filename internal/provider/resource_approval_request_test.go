package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
