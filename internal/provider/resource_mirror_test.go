package provider_test

import (
	"fmt"
	"regexp"
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
				ResourceName:            "registry_mirror.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

// TestAccMirror_noOrganizationID_refused pins backend 4.18+'s breaking
// change (#1012): a platform-admin create that names no organization, in
// the body or the X-Organization-Id header, and has no in-scope organization
// of its own to fall back to, is refused rather than silently landing in a
// default organization. The acceptance-test admin has no organization
// membership by design (see provider_test.go's bootstrapAcceptanceToken),
// so it always hits this path when organization_id is left unset.
func TestAccMirror_noOrganizationID_refused(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
resource "registry_mirror" "test" {
  name                  = "acc-mirror-no-org"
  upstream_registry_url = "https://registry.terraform.io"
  sync_interval_hours   = 24
}
`,
				ExpectError: regexp.MustCompile(`(?i)no organization context`),
			},
		},
	})
}

// testAccMirrorConfig creates its own organization and passes its id as
// organization_id: backend 4.18+ refuses a platform-admin create that names
// no organization (the acceptance-test admin has none by default — see
// provider_test.go's bootstrapAcceptanceToken), so this can no longer rely
// on organization_id's platform-wide default.
func testAccMirrorConfig(name, upstreamURL string, syncInterval int) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "acc_mirror" {
  name         = "acc-mirror-org"
  display_name = "Acceptance Mirror Org"
}

resource "registry_mirror" "test" {
  name                  = %q
  upstream_registry_url = %q
  sync_interval_hours   = %d
  organization_id       = registry_organization.acc_mirror.id
}
`, name, upstreamURL, syncInterval)
}
