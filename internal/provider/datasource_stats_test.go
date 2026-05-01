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
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "users"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "organizations"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "scm_providers"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "modules.total"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "providers.total"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "provider_mirrors.total"),
					resource.TestCheckResourceAttrSet("data.registry_stats.dashboard", "binary_mirrors.total"),
				),
			},
		},
	})
}
