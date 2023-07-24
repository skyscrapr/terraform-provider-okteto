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
					resource.TestCheckResourceAttr("okteto_secret.test", "name", "test_secret"),
					resource.TestCheckResourceAttr("okteto_secret.test", "value", "value_one"),
				),
			},
			// // Update and Read testing
			// {
			// 	Config: testAccExampleResourceConfig("two"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttrSet("okteto_secret.test", "id"),
			// 		resource.TestCheckResourceAttr("okteto_secret.test", "name", "test_secret"),
			// 		resource.TestCheckResourceAttr("okteto_secret.test", "value", "value_two"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccExampleResourceConfig(name_suffix string) string {
	return fmt.Sprintf(`
provider okteto {
	namespace = "skyscrapr"
}

resource "okteto_secret" "test" {
  name = "test_secret"
  value = "value_%s"
}
`, name_suffix)
}
