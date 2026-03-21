package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMirror_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMirrorConfig("acc-mirror", "https://registry.terraform.io", 24),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_mirror.test", "id"),
					resource.TestCheckResourceAttr("registry_mirror.test", "name", "acc-mirror"),
					resource.TestCheckResourceAttr("registry_mirror.test", "upstream_registry_url", "https://registry.terraform.io"),
					resource.TestCheckResourceAttr("registry_mirror.test", "sync_interval_hours", "24"),
					resource.TestCheckResourceAttr("registry_mirror.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("registry_mirror.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_mirror.test", "updated_at"),
				),
			},
			{
				Config: testAccMirrorConfig("acc-mirror-updated", "https://registry.terraform.io", 12),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_mirror.test", "name", "acc-mirror-updated"),
					resource.TestCheckResourceAttr("registry_mirror.test", "sync_interval_hours", "12"),
				),
			},
			{
				ResourceName:      "registry_mirror.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMirrorConfig(name, upstreamURL string, syncInterval int) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_mirror" "test" {
  name                  = %q
  upstream_registry_url = %q
  sync_interval_hours   = %d
}
`, name, upstreamURL, syncInterval)
}
