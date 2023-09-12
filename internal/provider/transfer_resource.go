package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgentf "github.com/doublecloud/go-sdk/gen/transfer"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TransferResource{}
var _ resource.ResourceWithImportState = &TransferResource{}

func NewTransferResource() resource.Resource {
	return &TransferResource{}
}

type TransferResource struct {
	sdk *dcsdk.SDK
	// endpointService *dcgentf.EndpointServiceClient
	transferService *dcgentf.TransferServiceClient
}

type TransferResourceModel struct {
	Id          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	Target      types.String `tfsdk:"target"`
	Type        types.String `tfsdk:"type"`
	Activated   types.Bool   `tfsdk:"activated"`
}

func (r *TransferResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transfer"
}

func (r *TransferResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Transfer resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Transfer id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of transfer",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description",
				Default:             stringdefault.StaticString(""),
			},
			"type": schema.StringAttribute{
				Optional: true,
				// Computed:            true,
				MarkdownDescription: "Transfer type",
				Validators:          []validator.String{transferTypeValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Source endpoint_id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target endpoint_id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"activated": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Activation of transfer",
			},
		},
	}
}

func (r *TransferResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*dcsdk.SDK)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *dcsdk.SDK, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.sdk = sdk
	// r.endpointService = r.sdk.Transfer().Endpoint()
	r.transferService = r.sdk.Transfer().Transfer()
}

func createTransferRequest(m *TransferResourceModel) (*transfer.CreateTransferRequest, diag.Diagnostics) {
	var diag diag.Diagnostics

	rq := &transfer.CreateTransferRequest{}
	rq.Name = m.Name.ValueString()
	rq.Description = m.Description.ValueString()
	rq.ProjectId = m.ProjectID.ValueString()
	rq.SourceId = m.Source.ValueString()
	rq.TargetId = m.Target.ValueString()
	rq.Type = transfer.TransferType(transfer.TransferType_value[m.Type.ValueString()])

	if diag.HasError() {
		return nil, diag
	}

	return rq, nil
}

func deleteTransferRequest(m *TransferResourceModel) (*transfer.DeleteTransferRequest, diag.Diagnostics) {
	return &transfer.DeleteTransferRequest{
		TransferId: m.Id.ValueString(),
	}, diag.Diagnostics{}
}

func (r *TransferResource) setActivation(ctx context.Context, m *TransferResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	var dcOp *doublecloud.Operation
	var err error

	if m.Activated.IsNull() {
		return diags
	}

	if m.Activated.ValueBool() {
		dcOp, err = r.transferService.Activate(ctx, &transfer.ActivateTransferRequest{
			TransferId: m.Id.ValueString(),
		})
	} else {
		dcOp, err = r.transferService.Deactivate(ctx, &transfer.DeactivateTransferRequest{
			TransferId: m.Id.ValueString(),
		})
	}

	if err != nil {
		diags.AddError("failed to activate", err.Error())
		return diags
	}
	op, err := r.sdk.WrapOperation(dcOp, err)
	if err != nil {
		diags.AddError("failed to activate", err.Error())
		return diags
	}
	err = op.Wait(ctx)
	if err != nil {
		diags.AddError("failed to activate", err.Error())
	}
	return diags
}

func (r *TransferResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TransferResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := createTransferRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.transferService.Create(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}

	data.Id = types.StringValue(op.ResourceId())

	r.setActivation(ctx, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TransferResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := r.transferService.Get(ctx, &transfer.GetTransferRequest{
		TransferId: data.Id.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}
	data.Name = types.StringValue(rs.Name)
	data.Description = types.StringValue(rs.Description)
	data.Source = types.StringValue(rs.Source.Id)
	data.Target = types.StringValue(rs.Target.Id)
	data.Type = types.StringValue(rs.Type.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TransferResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	rs, err := r.transferService.Update(ctx, &transfer.UpdateTransferRequest{
		TransferId:  data.Id.ValueString(),
		Description: data.Description.ValueString(),
		Name:        data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}

	r.setActivation(ctx, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TransferResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := deleteTransferRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.transferService.Delete(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to wait for delete completion", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted endpoint: %s", data.Id))
}

func (r *TransferResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func transferTypeValidator() validator.String {
	names := make([]string, len(transfer.TransferType_name))
	for i, v := range transfer.TransferType_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}
