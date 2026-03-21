package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProviderRecord_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderRecordConfig("initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_provider_record.test", "id"),
					resource.TestCheckResourceAttrSet("registry_provider_record.test", "organization_id"),
					resource.TestCheckResourceAttr("registry_provider_record.test", "namespace", "acc-provider-org"),
					resource.TestCheckResourceAttr("registry_provider_record.test", "type", "aws"),
					resource.TestCheckResourceAttr("registry_provider_record.test", "description", "initial description"),
					resource.TestCheckResourceAttrSet("registry_provider_record.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_provider_record.test", "updated_at"),
				),
			},
			{
				Config: testAccProviderRecordConfig("updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_provider_record.test", "description", "updated description"),
				),
			},
			{
				ResourceName:      "registry_provider_record.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProviderRecordConfig(description string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_organization" "test" {
  name         = "acc-provider-org"
  display_name = "Acc Provider Org"
}

resource "registry_provider_record" "test" {
  organization_id = registry_organization.test.id
  namespace       = registry_organization.test.name
  type            = "aws"
  description     = %q
}
`, description)
}
