package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataModules_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_organization" "seed" {
  name         = "acc-ds-modules-org"
  display_name = "Acc DS Modules Org"
}

resource "registry_module" "seed" {
  namespace       = registry_organization.seed.name
  name            = "vpc"
  system          = "aws"
}

data "registry_modules" "all" {
  depends_on = [registry_module.seed]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_modules.all", "modules.#"),
				),
			},
		},
	})
}
