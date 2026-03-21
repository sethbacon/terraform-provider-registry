package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataOrganizations_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_organization" "seed" {
  name         = "acc-ds-org"
  display_name = "Acc DS Org"
}

data "registry_organizations" "all" {
  depends_on = [registry_organization.seed]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_organizations.all", "organizations.#"),
				),
			},
		},
	})
}
