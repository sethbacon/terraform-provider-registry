package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganization_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig("acc-test-org", "Acc Test Org"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_organization.test", "id"),
					resource.TestCheckResourceAttr("registry_organization.test", "name", "acc-test-org"),
					resource.TestCheckResourceAttr("registry_organization.test", "display_name", "Acc Test Org"),
					resource.TestCheckResourceAttrSet("registry_organization.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_organization.test", "updated_at"),
				),
			},
			{
				Config: testAccOrganizationConfig("acc-test-org", "Acc Test Org Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_organization.test", "display_name", "Acc Test Org Updated"),
				),
			},
			{
				ResourceName:            "registry_organization.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
		},
	})
}

func testAccOrganizationConfig(name, displayName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "test" {
  name         = %q
  display_name = %q
}
`, name, displayName)
}
