package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataTerraformMirrors_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_terraform_mirror" "seed" {
  name         = "acc-ds-tf-mirror"
  tool         = "terraform"
  upstream_url = "https://releases.hashicorp.com"
}

data "registry_terraform_mirrors" "all" {
  depends_on = [registry_terraform_mirror.seed]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_terraform_mirrors.all", "terraform_mirrors.#"),
				),
			},
		},
	})
}
