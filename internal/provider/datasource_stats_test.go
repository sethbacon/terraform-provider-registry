package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataStats_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + `data "registry_stats" "dashboard" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "total_modules"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "total_providers"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "total_organizations"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "total_users"),
				),
			},
		},
	})
}
