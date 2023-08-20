// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package okteto

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
type PipelineResourceModel struct {
	Status   types.String   `tfsdk:"status"`
	Branch   types.String   `tfsdk:"branch"`
	RepoURL  types.String   `tfsdk:"repo_url"`
	Name     types.String   `tfsdk:"name"`
	Id       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
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
	var data *PipelineResourceModel

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

	err = retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
		status, err := getPipelineStatus(r.client, data.Name.ValueString())
		if err == nil {
			switch status {
			case "error", "":
				return retry.NonRetryableError(fmt.Errorf("pipeline failed. %s", status))
			case "deployed":
				return nil
			default:
				return retry.RetryableError(fmt.Errorf("expected instance to be created but was in state %s", status))
			}
		}
		return retry.NonRetryableError(fmt.Errorf("couldn't get pipeline by name. %s", err))
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to wait for pipeline to complete, got error: %s", err))
		return
	}
}

func (r *PipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *PipelineResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	pipeline, err := r.client.GetPipeline(r.client.Namespace, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get pipeline, got error: %s", err))
		return
	}
	v, _ := pipeline["status"].(string)
	data.Status = types.StringValue(v)
	// tflog.Trace(ctx, "read secret")
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *PipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *PipelineResourceModel

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
	err := r.client.DestroyPipeline(data.Name.ValueString(), r.client.Namespace, false)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to destroy pipeline, got error: %s", err))
		return
	}
	tflog.Info(ctx, "Waiting for pipeline to be destroyed...")
	err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
		status, err := getPipelineStatus(r.client, data.Name.ValueString())
		if err == nil {
			switch status {
			case "destroy-error", "":
				return retry.NonRetryableError(fmt.Errorf("pipeline destroy failed. %s", status))
			case "destroyed":
				return nil
			default:
				return retry.RetryableError(fmt.Errorf("expected instance to be destroyed but was in state %s", status))
			}
		}
		return retry.NonRetryableError(fmt.Errorf("couldn't get pipeline by name. %s", err))
	})
	if err != nil {
		tflog.Info(ctx, "Destroying pipeline with prejudice...")
		err = r.client.DestroyPipeline(data.Name.ValueString(), r.client.Namespace, true)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to force destroy pipeline, got error: %s", err))
			return
		}
		tflog.Info(ctx, "Waiting for pipeline to be destroyed...")
		err = retry.RetryContext(ctx, deleteTimeout, func() *retry.RetryError {
			status, err := getPipelineStatus(r.client, data.Name.ValueString())
			if err == nil {
				switch status {
				case "destroy-error", "":
					return retry.NonRetryableError(fmt.Errorf("pipeline destroy failed. %s", status))
				case "destroyed":
					return nil
				default:
					return retry.RetryableError(fmt.Errorf("expected instance to be destroyed but was in state %s", status))
				}
			}
			return retry.NonRetryableError(fmt.Errorf("couldn't get pipeline by name. %s", err))
		})
		if err != nil {
			tflog.Info(ctx, "Destroying pipeline with prejudice failed...")
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to force destroy pipeline, got error: %s", err))
			return
		}
	}
	tflog.Trace(ctx, "destroyed pipeline")
}

func (r *PipelineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getPipelineStatus(client *Client, pipelineName string) (string, error) {
	pipeline, err := client.GetPipeline(client.Namespace, pipelineName)
	status := ""
	if err == nil && pipeline != nil {
		status, _ = pipeline["status"].(string)
	}
	fmt.Printf("getPipelineStatus: status: %s, error %s \n", status, err)
	return status, err
}
