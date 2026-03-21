package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSCMProviders_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_scm_provider" "seed" {
  name          = "acc-ds-scm-github"
  type          = "github"
  client_id     = "acc-test-client-id"
  client_secret = "acc-test-client-secret"
}

data "registry_scm_providers" "all" {
  depends_on = [registry_scm_provider.seed]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_scm_providers.all", "scm_providers.#"),
				),
			},
		},
	})
}
