package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationMember_basic(t *testing.T) {
	rSuffix := acctest.RandString(6)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationMemberConfig(rSuffix, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_organization_member.test", "id"),
					resource.TestCheckResourceAttrSet("registry_organization_member.test", "organization_id"),
					resource.TestCheckResourceAttrSet("registry_organization_member.test", "user_id"),
					resource.TestCheckResourceAttrSet("registry_organization_member.test", "user_email"),
					resource.TestCheckResourceAttrSet("registry_organization_member.test", "created_at"),
				),
			},
			{
				// Update: assign a role template
				Config: testAccOrganizationMemberConfig(rSuffix, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_organization_member.test", "role_template_id"),
				),
			},
			{
				ResourceName:      "registry_organization_member.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccOrganizationMemberConfig(suffix string, withRole bool) string {
	roleBlock := ""
	if withRole {
		roleBlock = `role_template_id = registry_role_template.member_role.id`
	}

	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "test" {
  name         = "acc-member-org-%s"
  display_name = "Acc Member Org"
}

resource "registry_user" "test" {
  email = "acc-member-%s@example.com"
  name  = "Acc Member User"
}

resource "registry_role_template" "member_role" {
  name         = "acc-member-role-%s"
  display_name = "Acc Member Role"
  scopes       = ["modules:read"]
}

resource "registry_organization_member" "test" {
  organization_id = registry_organization.test.id
  user_id         = registry_user.test.id
  %s
}
`, suffix, suffix, suffix, roleBlock)
}
