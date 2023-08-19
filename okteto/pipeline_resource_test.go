// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package okteto

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPipelineResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPipelineResourceConfig("main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("okteto_pipeline.test", "id"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "name", "okteto_aws_lambda"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "repo_url", "https://github.com/skyscrapr/okteto-pipeline-test.git"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "branch", "main"),
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

func TestAccPipelineResourceFailedDestroy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPipelineResourceConfig("error"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("okteto_pipeline.test", "id"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "name", "okteto_aws_lambda"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "repo_url", "https://github.com/skyscrapr/okteto-pipeline-test.git"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "branch", "error"),
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

func testAccPipelineResourceConfig(branch string) string {
	return fmt.Sprintf(`
provider okteto {
	namespace = "skyscrapr"
}

resource "okteto_pipeline" "test" {
  name = "okteto_aws_lambda"
  repo_url = "https://github.com/skyscrapr/okteto-pipeline-test.git"
  branch = "%s"
}
`, branch)
}
