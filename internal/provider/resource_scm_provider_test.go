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

// testAccSCMProviderConfig creates its own organization and passes its id as
// organization_id: backend 4.18+ refuses a platform-admin create that names
// no organization (the acceptance-test admin has none by default — see
// provider_test.go's bootstrapAcceptanceToken), so this can no longer rely
// on organization_id's platform-wide default.
func testAccSCMProviderConfig(name, scmType string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "acc_scm" {
  name         = "acc-scm-org"
  display_name = "Acceptance SCM Org"
}

resource "registry_scm_provider" "test" {
  name            = %q
  type            = %q
  client_id       = "acc-test-client-id"
  client_secret   = "acc-test-client-secret"
  organization_id = registry_organization.acc_scm.id
}
`, name, scmType)
}
