package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataProviders_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_organization" "seed" {
  name         = "acc-ds-providers-org"
  display_name = "Acc DS Providers Org"
}

resource "registry_provider_record" "seed" {
  organization_id = registry_organization.seed.id
  namespace       = registry_organization.seed.name
  type            = "mycloud"
}

data "registry_providers" "all" {
  depends_on = [registry_provider_record.seed]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_providers.all", "providers.#"),
				),
			},
		},
	})
}
