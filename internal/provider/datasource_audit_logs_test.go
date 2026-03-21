package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataAuditLogs_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `
data "registry_audit_logs" "recent" {
  limit = 10
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_audit_logs.recent", "total"),
					resource.TestCheckResourceAttrSet("data.registry_audit_logs.recent", "audit_logs.#"),
				),
			},
		},
	})
}
