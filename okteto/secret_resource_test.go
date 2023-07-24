// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package okteto

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecretResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExampleResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("okteto_secret.test", "id"),
					resource.TestCheckResourceAttr("okteto_secret.test", "key", "test_one"),
					resource.TestCheckResourceAttr("okteto_secret.test", "value", "value"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "okteto_secret.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				Config: testAccExampleResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("okteto_secret.test", "key", "test_two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccExampleResourceConfig(key_suffix string) string {
	return fmt.Sprintf(`
provider okteto {
	namespace = "skyscrapr"
}

resource "okteto_secret" "test" {
  key = "test_%[1]s"
  value = "value"
}
`, key_suffix)
}
