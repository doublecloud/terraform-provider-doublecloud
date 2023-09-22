package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ExternalNetworkResource{}
var _ resource.ResourceWithImportState = &ExternalNetworkResource{}

func NewExternalNetworkResource() resource.Resource {
	return &ExternalNetworkResource{}
}

type ExternalNetworkResource struct {
	NetworkResource
}

type ExternalNetworkResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Ipv4CidrBlock types.String `tfsdk:"ipv4_cidr_block"`
	Ipv6CidrBlock types.String `tfsdk:"ipv6_cidr_block"`

	AWS    *awsExternalNetworkResourceModel    `tfsdk:"aws"`
	Google *googleExternalNetworkResourceModel `tfsdk:"google"`
}

type awsExternalNetworkResourceModel struct {
	VPCID          types.String `tfsdk:"vpc_id"`
	RegionID       types.String `tfsdk:"region_id"`
	AccountID      types.String `tfsdk:"account_id"`
	IAMRoleARN     types.String `tfsdk:"iam_role_arn"`
	PrivateSubnets types.Bool   `tfsdk:"private_subnets"`
}

type googleExternalNetworkResourceModel struct {
	NetworkName types.String `tfsdk:"network_name"`
	RegionID    types.String `tfsdk:"region_id"`
	ProjectName types.String `tfsdk:"project_name"`
	SAEmail     types.String `tfsdk:"service_account_email"`
}

func (r *ExternalNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_network"
}

func (r *ExternalNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"ipv4_cidr_block": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The IPv4 network range for the subnet.",
			},
			"ipv6_cidr_block": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The IPv6 network range for the subnet.",
			},
			"aws": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"vpc_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "ID of the VPC",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"region_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "ID of the region to place instances",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"account_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "ID of the VPC owner account",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"iam_role_arn": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "IAM role ARN to use for resource creations",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"private_subnets": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Create private subnets instead of default public",
					},
				},
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("aws"),
						path.MatchRoot("google"),
					}...),
				},
			},
			"google": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"network_name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of a network to import",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"region_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "ID of the region to place instances",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"project_name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of a project where is an imported network is located",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"service_account_email": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Service account email",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

func (r *ExternalNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ExternalNetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	importReq := &network.ImportNetworkRequest{
		Name:        data.Name.ValueString(),
		ProjectId:   data.ProjectID.ValueString(),
		Description: data.Description.ValueString(),
	}

	switch {
	case data.AWS != nil:
		importReq.Params = &network.ImportNetworkRequest_Aws{
			Aws: &network.ImportAWSVPCRequest{
				VpcId:      data.AWS.VPCID.ValueString(),
				RegionId:   data.AWS.RegionID.ValueString(),
				AccountId:  data.AWS.AccountID.ValueString(),
				IamRoleArn: data.AWS.IAMRoleARN.ValueString(),
			},
		}
	case data.Google != nil:
		// TODO export google to public API
	}

	opObj, err := r.networkService.Import(ctx, importReq)
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

	data.ID = types.StringValue(op.ResourceId())

	net, err := r.networkService.Get(ctx, &network.GetNetworkRequest{NetworkId: op.ResourceId()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get network", fmt.Sprintf("failed request, error: %v", err))
		return
	}
	data.Ipv4CidrBlock = types.StringValue(net.Ipv4CidrBlock)
	data.Ipv6CidrBlock = types.StringValue(net.Ipv6CidrBlock)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExternalNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ExternalNetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	net, err := r.networkService.Get(ctx, &network.GetNetworkRequest{NetworkId: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get network", fmt.Sprintf("failed request, error: %v", err))
		return
	}

	data.ID = types.StringValue(net.Id)
	data.Name = types.StringValue(net.Name)
	data.ProjectID = types.StringValue(net.ProjectId)
	data.Description = types.StringValue(net.Description)
	data.Ipv4CidrBlock = types.StringValue(net.Ipv4CidrBlock)
	data.Ipv6CidrBlock = types.StringValue(net.Ipv6CidrBlock)

	switch er := net.ExternalResources.(type) {
	case *network.Network_Aws:
		data.AWS = &awsExternalNetworkResourceModel{
			VPCID:      types.StringValue(er.Aws.VpcId),
			RegionID:   types.StringValue(net.RegionId),
			AccountID:  types.StringValue(er.Aws.AccountId.GetValue()),
			IAMRoleARN: types.StringValue(er.Aws.IamRoleArn.GetValue()),
			// TODO export PrivateSubnets to public API
			//PrivateSubnets: types.BoolValue(er.Aws.),
		}
	case *network.Network_Gcp:
		data.Google = &googleExternalNetworkResourceModel{
			// TODO export GCP params
			//NetworkName: types.StringValue(er.Gcp),
			//RegionID:    types.StringValue(),
			//ProjectName: types.StringValue(),
			//SAEmail:     types.StringValue(),
		}
	}

	tflog.Info(ctx, fmt.Sprintf("read#5 %v", data.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExternalNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ExternalNetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	net, err := r.networkService.Delete(ctx, &network.DeleteNetworkRequest{NetworkId: data.ID.ValueString()})
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
