package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleTemplate_basic(t *testing.T) {
	rName := "acc-pub-" + acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleTemplateConfig(rName, "Acc Publisher", `["modules:read", "modules:write"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_role_template.test", "id"),
					resource.TestCheckResourceAttr("registry_role_template.test", "name", rName),
					resource.TestCheckResourceAttr("registry_role_template.test", "display_name", "Acc Publisher"),
					resource.TestCheckResourceAttr("registry_role_template.test", "scopes.#", "2"),
					resource.TestCheckResourceAttr("registry_role_template.test", "is_system", "false"),
					resource.TestCheckResourceAttrSet("registry_role_template.test", "created_at"),
				),
			},
			{
				Config: testAccRoleTemplateConfig(rName, "Acc Publisher Updated", `["modules:read"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_role_template.test", "display_name", "Acc Publisher Updated"),
					resource.TestCheckResourceAttr("registry_role_template.test", "scopes.#", "1"),
				),
			},
			{
				ResourceName:            "registry_role_template.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

func testAccRoleTemplateConfig(name, displayName, scopes string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_role_template" "test" {
  name         = %q
  display_name = %q
  scopes       = %s
}
`, name, displayName, scopes)
}
