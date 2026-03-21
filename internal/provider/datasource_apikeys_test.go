package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataAPIKeys_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_organization" "seed" {
  name         = "acc-ds-apikeys-org"
  display_name = "Acc DS API Keys Org"
}

resource "registry_api_key" "seed" {
  organization_id = registry_organization.seed.id
  name            = "acc-ds-apikey"
  scopes          = ["modules:read"]
}

data "registry_api_keys" "all" {
  depends_on = [registry_api_key.seed]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_api_keys.all", "api_keys.#"),
				),
			},
		},
	})
}
