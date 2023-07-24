// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package okteto

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ScaffoldingProvider{}

// ScaffoldingProvider defines the provider implementation.
type ScaffoldingProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type ScaffoldingProviderModel struct {
	ApiToken  types.String `tfsdk:"api_token"`
	Namespace types.String `tfsdk:"namespace"`
}

func (p *ScaffoldingProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "okteto"
	resp.Version = p.version
}

func (p *ScaffoldingProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				MarkdownDescription: "Okteto API Token",
				Optional:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Okteto Namespace",
				Required:            true,
			},
		},
	}
}

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	api_token := os.Getenv("OKTETO_API_TOKEN")
	if !data.ApiToken.IsNull() {
		api_token = data.ApiToken.ValueString()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	if data.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"No Okteto API Token",
			"The provider cannot create the Okteto API client as there is an unknown configuration value for the Okteto API token. "+
				"Either set the value statically in the configuration, or use the OKTETO_API_TOKEN environment variable.",
		)
	}

	// Example client configuration for data sources and resources
	client := NewClient(api_token, data.Namespace.ValueString())
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ScaffoldingProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSecretResource,
	}
}

func (p *ScaffoldingProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ScaffoldingProvider{
			version: version,
		}
	}
}
