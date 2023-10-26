// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package okteto

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)



func TestAccPipelineResource_basic(t *testing.T) {
	testExternalProviders := map[string]resource.ExternalProvider{
		"aws": {
			Source:            "hashicorp/aws",
			// VersionConstraint: "~> 2.3",
		},
	}
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: testExternalProviders,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPipelineResourceConfig_basic("main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("okteto_pipeline.test", "id"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "name", "okteto_aws_s3"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "repo_url", "https://github.com/skyscrapr/oktetodo-terraform-s3.git"),
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
				Config: testAccPipelineResourceConfig_basic("error"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("okteto_pipeline.test", "id"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "name", "okteto_aws_lambda"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "repo_url", "https://github.com/skyscrapr/okteto-pipeline-test.git"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "branch", "error"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccPipelineResource_complex(t *testing.T) {
	t.Skip("Skip due to time to test")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"aws": {
				Source: "hashicorp/aws",
			},
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPipelineResourceConfig_complex("main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("okteto_pipeline.test", "id"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "name", "okteto_aws_lambda"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "repo_url", "https://github.com/skyscrapr/oktetodo-terraform-s3.git"),
					resource.TestCheckResourceAttr("okteto_pipeline.test", "branch", "main"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// func testAccPipelineResourceConfig_basic(branch string) string {
// 	return fmt.Sprintf(`
// provider okteto {
// 	namespace = "skyscrapr"
// }

// resource "okteto_pipeline" "test" {
//   name = "okteto_aws_lambda"
//   repo_url = "https://github.com/skyscrapr/okteto-pipeline-test.git"
//   branch = "%s"
// }
// `, branch)
// }

func testAccPipelineResourceConfig_basic(branch string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

provider okteto {
	namespace = "skyscrapr"
}

resource "okteto_pipeline" "test" {
  name = "okteto_aws_s3"
  repo_url = "https://github.com/skyscrapr/oktetodo-terraform-s3.git"
  branch = "%s"
}
`, branch)
}

func testAccPipelineResourceConfig_complex(branch string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
	}

provider okteto {
	namespace = "skyscrapr"
}

resource "okteto_pipeline" "test" {
  name = "okteto_aws_lambda"
  repo_url = "https://github.com/skyscrapr/oktetodo-terraform-s3.git"
  branch = "%s"
  depends_on = [
	okteto_secret.aws_access_key_id,
	okteto_secret.aws_secret_access_key,
	okteto_secret.aws_region
  ]
}

resource okteto_secret "aws_access_key_id" {
    name = "AWS_ACCESS_KEY_ID"
    value = aws_iam_access_key.okteto_deploy.id
}

resource okteto_secret "aws_secret_access_key" {
    name = "AWS_SECRET_ACCESS_KEY"
    value = aws_iam_access_key.okteto_deploy.secret
    depends_on = [
        okteto_secret.aws_access_key_id
    ]
}

resource okteto_secret "aws_region" {
    name = "AWS_REGION"
    value = "us-east-1"
    depends_on = [
        okteto_secret.aws_secret_access_key
    ]
}

resource "aws_iam_user" "okteto_deploy" {
	name = "okteto_deploy"
  }
  
  resource "aws_iam_user_policy_attachment" "cloudformation_fullaccess" {
	user       = aws_iam_user.okteto_deploy.name
	policy_arn = "arn:aws:iam::aws:policy/AWSCloudFormationFullAccess"
  }
  
  resource "aws_iam_user_policy_attachment" "iam_full_access" {
	user       = aws_iam_user.okteto_deploy.name
	policy_arn = "arn:aws:iam::aws:policy/IAMFullAccess"
  }
  
  resource "aws_iam_user_policy_attachment" "lambda_full_access" {
	user       = aws_iam_user.okteto_deploy.name
	policy_arn = "arn:aws:iam::aws:policy/AWSLambda_FullAccess"
  }
  
  resource "aws_iam_user_policy_attachment" "api_gateway_admin" {
	user       = aws_iam_user.okteto_deploy.name
	policy_arn = "arn:aws:iam::aws:policy/AmazonAPIGatewayAdministrator"
  }
  
  resource "aws_iam_user_policy_attachment" "s3_full_access" {
	user       = aws_iam_user.okteto_deploy.name
	policy_arn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
  }
  
  resource "aws_iam_access_key" "okteto_deploy" {
	user = aws_iam_user.okteto_deploy.name
  }
`, branch)
}
