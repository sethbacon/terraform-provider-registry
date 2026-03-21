package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTerraformMirror_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTerraformMirrorConfig("acc-tf-mirror", "terraform", "https://releases.hashicorp.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_terraform_mirror.test", "id"),
					resource.TestCheckResourceAttr("registry_terraform_mirror.test", "name", "acc-tf-mirror"),
					resource.TestCheckResourceAttr("registry_terraform_mirror.test", "tool", "terraform"),
					resource.TestCheckResourceAttr("registry_terraform_mirror.test", "upstream_url", "https://releases.hashicorp.com"),
					resource.TestCheckResourceAttr("registry_terraform_mirror.test", "enabled", "true"),
					resource.TestCheckResourceAttr("registry_terraform_mirror.test", "stable_only", "true"),
					resource.TestCheckResourceAttr("registry_terraform_mirror.test", "sync_interval_hours", "24"),
					resource.TestCheckResourceAttrSet("registry_terraform_mirror.test", "created_at"),
				),
			},
			{
				Config: testAccTerraformMirrorConfig("acc-tf-mirror-updated", "terraform", "https://releases.hashicorp.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_terraform_mirror.test", "name", "acc-tf-mirror-updated"),
				),
			},
			{
				ResourceName:      "registry_terraform_mirror.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTerraformMirrorConfig(name, tool, upstreamURL string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_terraform_mirror" "test" {
  name         = %q
  tool         = %q
  upstream_url = %q
}
`, name, tool, upstreamURL)
}
