package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig("acc-policy", "allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("registry_policy.test", "id"),
					resource.TestCheckResourceAttr("registry_policy.test", "name", "acc-policy"),
					resource.TestCheckResourceAttr("registry_policy.test", "policy_type", "allow"),
					resource.TestCheckResourceAttr("registry_policy.test", "is_active", "true"),
					resource.TestCheckResourceAttrSet("registry_policy.test", "created_at"),
					resource.TestCheckResourceAttrSet("registry_policy.test", "updated_at"),
				),
			},
			{
				Config: testAccPolicyConfig("acc-policy-updated", "deny"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("registry_policy.test", "name", "acc-policy-updated"),
					resource.TestCheckResourceAttr("registry_policy.test", "policy_type", "deny"),
				),
			},
			{
				ResourceName:      "registry_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPolicyConfig(name, policyType string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "registry_policy" "test" {
  name              = %q
  policy_type       = %q
  namespace_pattern = "hashicorp/*"
  is_active         = true
}
`, name, policyType)
}
