package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig("acc-test-key", `["modules:read"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_api_key.test", "id"),
					resource.TestCheckResourceAttr("registry_api_key.test", "name", "acc-test-key"),
					resource.TestCheckResourceAttrSet("registry_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("registry_api_key.test", "key_prefix"),
					resource.TestCheckResourceAttr("registry_api_key.test", "scopes.#", "1"),
					resource.TestCheckResourceAttrSet("registry_api_key.test", "created_at"),
				),
			},
			{
				Config: testAccAPIKeyConfig("acc-test-key-updated", `["modules:read", "modules:write"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_api_key.test", "name", "acc-test-key-updated"),
					resource.TestCheckResourceAttr("registry_api_key.test", "scopes.#", "2"),
				),
			},
			{
				// Raw key is not recoverable on import — ignore it.
				ResourceName:            "registry_api_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func testAccAPIKeyConfig(name, scopes string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "test" {
  name         = "acc-apikey-org"
  display_name = "Acc API Key Org"
}

resource "registry_api_key" "test" {
  organization_id = registry_organization.test.id
  name            = %q
  scopes          = %s
}
`, name, scopes)
}
