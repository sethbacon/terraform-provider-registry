package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStorageConfig_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageConfigLocal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_storage_config.test", "id"),
					resource.TestCheckResourceAttr("registry_storage_config.test", "backend", "local"),
					resource.TestCheckResourceAttrSet("registry_storage_config.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_storage_config.test", "updated_at"),
				),
			},
			{
				ResourceName:            "registry_storage_config.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config", "activate"},
			},
		},
	})
}

func testAccStorageConfigLocal() string {
	return testAccProviderConfig() + `
resource "registry_storage_config" "test" {
  backend = "local"
  config  = {
    local_base_path = "/tmp/acc-test-storage"
  }
  activate = false
}
`
}
