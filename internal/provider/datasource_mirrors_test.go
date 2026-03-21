package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataMirrors_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_mirror" "seed" {
  name                  = "acc-ds-mirror"
  upstream_registry_url = "https://registry.terraform.io"
}

data "registry_mirrors" "all" {
  depends_on = [registry_mirror.seed]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_mirrors.all", "mirrors.#"),
				),
			},
		},
	})
}
