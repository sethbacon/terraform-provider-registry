package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccModule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModuleConfig("acc-module-org", "vpc", "aws", "A VPC module"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_module.test", "id"),
					resource.TestCheckResourceAttr("registry_module.test", "name", "vpc"),
					resource.TestCheckResourceAttr("registry_module.test", "system", "aws"),
					resource.TestCheckResourceAttr("registry_module.test", "description", "A VPC module"),
					resource.TestCheckResourceAttrSet("registry_module.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_module.test", "updated_at"),
				),
			},
			{
				Config: testAccModuleConfig("acc-module-org", "vpc", "aws", "Updated VPC module"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_module.test", "description", "Updated VPC module"),
				),
			},
			{
				ResourceName:            "registry_module.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

func testAccModuleConfig(namespace, name, system, description string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "test" {
  name         = %q
  display_name = "Acc Module Org"
}

resource "registry_module" "test" {
  namespace       = registry_organization.test.name
  name            = %q
  system          = %q
  description     = %q
}
`, namespace, name, system, description)
}
