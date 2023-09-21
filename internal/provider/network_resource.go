package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgennet "github.com/doublecloud/go-sdk/gen/network"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NetworkResource{}
var _ resource.ResourceWithImportState = &NetworkResource{}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

type NetworkResource struct {
	sdk            *dcsdk.SDK
	networkService *dcgennet.NetworkServiceClient
}

type NetworkResourceModel struct {
	Id            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	RegionID      types.String `tfsdk:"region_id"`
	CloudType     types.String `tfsdk:"cloud_type"`
	Ipv4CidrBlock types.String `tfsdk:"ipv4_cidr_block"`
	Ipv6CidrBlock types.String `tfsdk:"ipv6_cidr_block"`
}

func (r *NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network identifier",
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
			"cloud_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud type (aws, gcp, azure)",
				Validators:          []validator.String{cloudTypeValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of network",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of network",
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Region of network",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ipv4_cidr_block": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The IPv4 network range for the subnet, in CIDR notation. For example, 10.0.0.0/16.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ipv6_cidr_block": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The IPv6 network range for the subnet, it is known only after creation.",
			},
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.networkService = r.sdk.Network().Network()
}

func createNetworkRequest(m *NetworkResourceModel) (*network.CreateNetworkRequest, diag.Diagnostics) {
	rq := &network.CreateNetworkRequest{}
	rq.Name = m.Name.ValueString()
	rq.CloudType = m.CloudType.ValueString()
	rq.ProjectId = m.ProjectID.ValueString()
	rq.Description = m.Description.ValueString()
	rq.RegionId = m.RegionID.ValueString()
	rq.Ipv4CidrBlock = m.Ipv4CidrBlock.ValueString()
	return rq, nil
}

func deleteNetworkRequest(m *NetworkResourceModel) (*network.DeleteNetworkRequest, diag.Diagnostics) {
	rq := &network.DeleteNetworkRequest{NetworkId: m.Id.ValueString()}
	return rq, nil
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	opObj, err := r.networkService.Create(ctx, &network.CreateNetworkRequest{
		Name:          data.Name.ValueString(),
		CloudType:     data.CloudType.ValueString(),
		ProjectId:     data.ProjectID.ValueString(),
		Description:   data.Description.ValueString(),
		RegionId:      data.RegionID.ValueString(),
		Ipv4CidrBlock: data.Ipv4CidrBlock.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(opObj, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}

	data.Id = types.StringValue(op.ResourceId())

	net, err := r.networkService.Get(ctx, &network.GetNetworkRequest{NetworkId: op.ResourceId()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get network", fmt.Sprintf("failed request, error: %v", err))
		return
	}
	data.Ipv6CidrBlock = types.StringValue(net.Ipv6CidrBlock)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getNetworkResourceRequest(m *NetworkResourceModel) (*network.GetNetworkRequest, diag.Diagnostics) {
	if m.Id == types.StringNull() {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unknown network identifier", "missed one of required fields: network_id or name")}
	}
	return &network.GetNetworkRequest{NetworkId: m.Id.ValueString()}, nil
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *NetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	net, err := r.networkService.Get(ctx, &network.GetNetworkRequest{NetworkId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get network", fmt.Sprintf("failed request, error: %v", err))
		return
	}

	data.Id = types.StringValue(net.Id)
	data.Name = types.StringValue(net.Name)
	data.ProjectID = types.StringValue(net.ProjectId)
	data.Description = types.StringValue(net.Description)
	data.CloudType = types.StringValue(net.CloudType)
	data.RegionID = types.StringValue(net.RegionId)
	data.Ipv4CidrBlock = types.StringValue(net.Ipv4CidrBlock)
	data.Ipv6CidrBlock = types.StringValue(net.Ipv6CidrBlock)
	tflog.Info(ctx, fmt.Sprintf("read#5 %v", data.Id))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.AddError("Failed to update network", "networks doesn't support updates")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	net, err := r.networkService.Delete(ctx, &network.DeleteNetworkRequest{NetworkId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(net, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
	err = op.Wait(ctx)
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
