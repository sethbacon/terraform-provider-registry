package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig("acc-user@example.com", "Acc User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_user.test", "id"),
					resource.TestCheckResourceAttr("registry_user.test", "email", "acc-user@example.com"),
					resource.TestCheckResourceAttr("registry_user.test", "name", "Acc User"),
					resource.TestCheckResourceAttrSet("registry_user.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_user.test", "updated_at"),
				),
			},
			{
				Config: testAccUserConfig("acc-user@example.com", "Acc User Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_user.test", "name", "Acc User Updated"),
				),
			},
			{
				ResourceName:            "registry_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_at", "updated_at"},
			},
		},
	})
}

func testAccUserConfig(email, name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_user" "test" {
  email = %q
  name  = %q
}
`, email, name)
}
