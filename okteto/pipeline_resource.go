// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package okteto

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PipelineResource{}
var _ resource.ResourceWithImportState = &PipelineResource{}

func NewPipelineResource() resource.Resource {
	return &PipelineResource{}
}

// PipelineResource defines the resource implementation.
type PipelineResource struct {
	client *Client
}

// PipelineResourceModel describes the resource data model.
type pipelineResourceModel struct {
	Status      types.String   `tfsdk:"status"`
	Branch      types.String   `tfsdk:"branch"`
	RepoURL     types.String   `tfsdk:"repo_url"`
	Name        types.String   `tfsdk:"name"`
	Id          types.String   `tfsdk:"id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	Deployments types.Set      `tfsdk:"deployments"`
}

func (r *PipelineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

func (r *PipelineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Pipeline resource",

		Attributes: map[string]schema.Attribute{
			"status": schema.StringAttribute{
				MarkdownDescription: "Status",
				Computed:            true,
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "Branch",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo_url": schema.StringAttribute{
				MarkdownDescription: "RepoURL",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Pipeline identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployments": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"endpoints": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *PipelineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *pipelineResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 20*time.Minute)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	err := r.client.NewPipeline(r.client.Namespace, data.Name.ValueString(), data.RepoURL.ValueString(), data.Branch.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create pipeline, got error: %s", err))
		return
	}
	data.Id = data.Name
	data.Status = types.StringValue("Unknown")
	tflog.Trace(ctx, "created pipeline")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	err = waitPipelineState(ctx, createTimeout, r.client, data.Name.ValueString(), "error", "deployed")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to wait for pipeline state to be %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	err = waitDeploymentStates(ctx, createTimeout, r.client, data.Name.ValueString(), "error", "running")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to wait for pipeline state to be %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(data.refresh(ctx, r.client)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *PipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *pipelineResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(data.refresh(ctx, r.client)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *PipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *pipelineResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 20*time.Minute)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	tflog.Info(ctx, "Destroying pipeline...")
	err := destroyPipeline(ctx, r.client, deleteTimeout, data.Name.ValueString(), false)
	if err != nil {
		tflog.Info(ctx, fmt.Sprintf("Unable to destroy pipeline, got error: %s", err))
		tflog.Info(ctx, "Destroying pipeline with prejudice...")
		err = destroyPipeline(ctx, r.client, deleteTimeout, data.Name.ValueString(), true)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to force destroy pipeline, got error: %s", err))
			return
		}
	}
	tflog.Trace(ctx, "destroyed pipeline")
}

func (r *PipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func destroyPipeline(ctx context.Context, client *Client, timeout time.Duration, pipelineName string, force bool) error {
	err := client.DestroyPipeline(pipelineName, client.Namespace, force)
	if err == nil {
		tflog.Info(ctx, "Waiting for pipeline to be destroyed...")
		err = waitPipelineState(ctx, timeout, client, pipelineName, "destroy-error", "destroyed")
	}
	return err
}

func waitPipelineState(ctx context.Context, timeout time.Duration, client *Client, pipelineName string, errorState string, successState string) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		status, err := getPipelineState(client, pipelineName)
		if err == nil {
			switch status {
			case errorState:
				return retry.NonRetryableError(fmt.Errorf("pipeline failed. %s", status))
			case successState, "":
				return nil
			default:
				return retry.RetryableError(fmt.Errorf("expected instance to be created but was in state %s", status))
			}
		}
		return retry.NonRetryableError(fmt.Errorf("couldn't get pipeline by name. %s", err))
	})
}

func waitDeploymentStates(ctx context.Context, timeout time.Duration, client *Client, pipelineName string, errorState string, successState string) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		pipeline, err := client.GetPipeline(client.Namespace, pipelineName)
		if err != nil {
			fmt.Printf("waitDeploymentStates: error getting pipelin: %s \n", err)
		}
		if pipeline != nil {
			deployments, ok := pipeline["deployments"].([]map[string]interface{})
			if !ok {
				return retry.NonRetryableError(fmt.Errorf("could not get deployments data"))
			}
			for _, deployment := range deployments {
				name, ok := deployment["name"].(string)
				if !ok {
					return retry.NonRetryableError(fmt.Errorf("could not get deployment name"))
				}
				status, ok := deployment["status"].(string)
				if !ok {
					return retry.NonRetryableError(fmt.Errorf("could not get deployment state"))
				}
				fmt.Printf("Pipeline %s: Deployment %s: status: %s\n", pipelineName, name, status)
			}
			for _, deployment := range deployments {
				status, ok := deployment["status"].(string)
				if !ok {
					return retry.NonRetryableError(fmt.Errorf("could not get deployment state"))
				}

				switch deployment["status"] {
				case errorState:
					return retry.NonRetryableError(fmt.Errorf("pipeline deployment failed. %s", status))
				case successState:
				default:
					return retry.RetryableError(fmt.Errorf("retryable state: %s", status))
				}
			}
		}
		return nil
	})
}

func getPipelineState(client *Client, pipelineName string) (string, error) {
	pipeline, err := client.GetPipeline(client.Namespace, pipelineName)
	status := ""
	if err == nil && pipeline != nil {
		status, _ = pipeline["status"].(string)
		fmt.Printf("Pipeline %s status: %s\n", pipelineName, status)
	}
	return status, err
}

func flattenDeployments(ctx context.Context, set interface{}) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{}
	attributeTypes["endpoints"] = types.SetType{ElemType: types.StringType}
	elemType := types.ObjectType{AttrTypes: attributeTypes}

	deployments, ok := set.([]map[string]interface{})
	if !ok {
		diags.AddError("Client Error", "Could not get deployments data")
		return types.SetNull(elemType), diags
	}

	attrs := make([]attr.Value, 0, len(deployments))
	for _, d := range deployments {
		attr := map[string]attr.Value{}
		endpoints, _ := d["endpoints"].([]interface{})
		newEndpoints := make([]string, len(endpoints))
		for i, e := range endpoints {
			newEndpoint, _ := e.(map[string]interface{})
			url, _ := newEndpoint["url"].(string)
			newEndpoints[i] = url
		}
		//model.Endpoints = types.SetValueMust(types.StringType, newEndpoints)
		attr["endpoints"], diags = types.SetValueFrom(ctx, types.StringType, newEndpoints)
		val := types.ObjectValueMust(attributeTypes, attr)
		attrs = append(attrs, val)
	}

	return types.SetValueMust(elemType, attrs), diags
}

func (data *pipelineResourceModel) refresh(ctx context.Context, client *Client) diag.Diagnostics {
	var diags diag.Diagnostics

	pipeline, err := client.GetPipeline(client.Namespace, data.Name.ValueString())
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to get pipeline, got error: %s", err))
		return diags
	}

	v, _ := pipeline["status"].(string)
	data.Status = types.StringValue(v)
	data.Deployments, diags = flattenDeployments(ctx, pipeline["deployments"])

	return diags
}
