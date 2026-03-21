package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSCMProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSCMProviderConfig("Acc GitHub SCM", "github"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_scm_provider.test", "id"),
					resource.TestCheckResourceAttr("registry_scm_provider.test", "name", "Acc GitHub SCM"),
					resource.TestCheckResourceAttr("registry_scm_provider.test", "type", "github"),
					resource.TestCheckResourceAttrSet("registry_scm_provider.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_scm_provider.test", "updated_at"),
				),
			},
			{
				Config: testAccSCMProviderConfig("Acc GitHub SCM Updated", "github"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_scm_provider.test", "name", "Acc GitHub SCM Updated"),
				),
			},
			{
				ResourceName:            "registry_scm_provider.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_id", "client_secret", "updated_at"},
			},
		},
	})
}

func testAccSCMProviderConfig(name, scmType string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_scm_provider" "test" {
  name          = %q
  type          = %q
  client_id     = "acc-test-client-id"
  client_secret = "acc-test-client-secret"
}
`, name, scmType)
}
