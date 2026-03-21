package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccModuleSCMLink_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModuleSCMlinkConfig("main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_module_scm_link.test", "module_id"),
					resource.TestCheckResourceAttrSet("registry_module_scm_link.test", "scm_provider_id"),
					resource.TestCheckResourceAttr("registry_module_scm_link.test", "owner", "my-org"),
					resource.TestCheckResourceAttr("registry_module_scm_link.test", "repo", "terraform-aws-vpc"),
					resource.TestCheckResourceAttr("registry_module_scm_link.test", "branch", "main"),
					resource.TestCheckResourceAttrSet("registry_module_scm_link.test", "created_at"),
				),
			},
			{
				Config: testAccModuleSCMlinkConfig("develop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_module_scm_link.test", "branch", "develop"),
				),
			},
			{
				ResourceName:                  "registry_module_scm_link.test",
				ImportState:                   true,
				ImportStateVerify:             true,
				ImportStateVerifyIdentifierAttribute: "module_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["registry_module_scm_link.test"]
					if !ok {
						return "", fmt.Errorf("not found: registry_module_scm_link.test")
					}
					return rs.Primary.Attributes["module_id"], nil
				},
			},
		},
	})
}

func testAccModuleSCMlinkConfig(branch string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "test" {
  name         = "acc-scmlink-org"
  display_name = "Acc SCM Link Org"
}

resource "registry_module" "test" {
  namespace       = registry_organization.test.name
  name            = "vpc"
  system          = "aws"
}

resource "registry_scm_provider" "test" {
  name          = "acc-scmlink-github"
  type          = "github"
  client_id     = "acc-test-client-id"
  client_secret = "acc-test-client-secret"
}

resource "registry_module_scm_link" "test" {
  module_id      = registry_module.test.id
  scm_provider_id = registry_scm_provider.test.id
  owner          = "my-org"
  repo           = "terraform-aws-vpc"
  branch         = %q
}
`, branch)
}
