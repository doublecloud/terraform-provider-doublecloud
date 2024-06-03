package provider

import (
	"context"
	"fmt"
	"github.com/doublecloud/go-genproto/doublecloud/organizationmanager/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcorganization "github.com/doublecloud/go-sdk/gen/organization"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IAMOrganizationGroup{}
var _ resource.ResourceWithImportState = &IAMOrganizationGroup{}
var _ resource.ResourceWithConfigure = &IAMOrganizationGroup{}

func NewIAMOrganizationGroup() resource.Resource {
	return &IAMOrganizationGroup{}
}

type IAMOrganizationGroup struct {
	sdk                 *dcsdk.SDK
	organizationService *dcorganization.Organization
}

type IAMOrganizationGroupModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

func (m *IAMOrganizationGroupModel) FromProtobuf(nc *organizationmanager.Group) error {
	m.ID = types.StringValue(nc.GetId())
	m.OrganizationID = types.StringValue(nc.GetOrganizationId())
	m.Name = types.StringValue(nc.GetName())
	m.Description = types.StringValue(nc.GetDescription())

	return nil
}

func (l *IAMOrganizationGroup) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group"
}

func (l *IAMOrganizationGroup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Organization Group resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Group ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization group name",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Organization group description",
				Default:             stringdefault.StaticString(""),
			},
		},
	}
}

func (l *IAMOrganizationGroup) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	l.sdk = sdk
	l.organizationService = l.sdk.Organization()
}

func (l *IAMOrganizationGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IAMOrganizationGroupModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &organizationmanager.CreateGroupRequest{
		OrganizationId: data.OrganizationID.ValueString(),
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
	}
	opObj, err := l.organizationService.Group().Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}

	data.ID = types.StringValue(opObj.GetResourceId())

	resp.Diagnostics.Append(getGroup(ctx, l.organizationService, opObj.GetResourceId(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (l *IAMOrganizationGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *IAMOrganizationGroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(getGroup(ctx, l.organizationService, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getGroup(
	ctx context.Context,
	client *dcorganization.Organization,
	id string,
	data *IAMOrganizationGroupModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	nc, err := client.Group().Get(ctx, &organizationmanager.GetGroupRequest{GroupId: id})
	if err != nil {
		diags.AddError("Failed to get network connection", fmt.Sprintf("failed request, error: %v", err))
		return diags
	}

	if err = data.FromProtobuf(nc); err != nil {
		diags.AddError("Failed to get network connection", fmt.Sprintf("failed parse, error: %v", err))
		return diags
	}

	return diags
}

func (l *IAMOrganizationGroup) Update(ctx context.Context, request resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Failed to update logs export", "logs expport don't support updates")
}

func (l *IAMOrganizationGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IAMOrganizationGroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := l.organizationService.Group().Delete(ctx, &organizationmanager.DeleteGroupRequest{GroupId: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
}

func (l *IAMOrganizationGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
