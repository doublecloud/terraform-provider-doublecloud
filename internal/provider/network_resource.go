package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
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
	IsExternal    types.Bool   `tfsdk:"is_external"`

	AWS *awsExternalNetworkResourceModel    `tfsdk:"aws"`
	GCP *googleExternalNetworkResourceModel `tfsdk:"gcp"`
}

type awsExternalNetworkResourceModel struct {
	VPCID          types.String `tfsdk:"vpc_id"`
	AccountID      types.String `tfsdk:"account_id"`
	IAMRoleARN     types.String `tfsdk:"iam_role_arn"`
	PrivateSubnets types.Bool   `tfsdk:"private_subnets"`
}

type googleExternalNetworkResourceModel struct {
	NetworkName    types.String `tfsdk:"network_name"`
	SubnetworkName types.String `tfsdk:"subnetwork_name"`
	ProjectName    types.String `tfsdk:"project_name"`
	SAEmail        types.String `tfsdk:"service_account_email"`
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
				MarkdownDescription: "Network ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud provider (`aws`, `gcp`, or `azure`)",
				Validators:          []validator.String{cloudTypeValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network description",
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network region",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ipv4_cidr_block": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Subnet IPv4 network range in CIDR notation, such as `10.0.0.0/16`.\n" +
					"    Required for non-BYOC networks.\n" +
					"    For BYOC, it's read from the provided VPC (AWS) or Subnetwork (GCP).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("aws"),
						path.MatchRoot("gcp"),
					}...),
				},
			},
			"ipv6_cidr_block": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Subnet IPv6 network rabge. Available only after the network is created.",
			},
			"is_external": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "True if the network was imported using BYOC.",
			},
			"aws": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"vpc_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "VPC ID",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"account_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "VPC owner account ID",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"iam_role_arn": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "ARN of an IAM role with permissions to create resources",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"private_subnets": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Create private subnets instead of the default public ones",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("ipv4_cidr_block"),
						path.MatchRoot("gcp"),
					}...),
				},
			},
			"gcp": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "BYOC parameters for GCP",
				Attributes: map[string]schema.Attribute{
					"network_name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of the network to import",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"subnetwork_name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of the subnetwork to import",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"project_name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of the project where is the imported network is located",
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
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("ipv4_cidr_block"),
						path.MatchRoot("aws"),
					}...),
				},
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

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var opObj *doublecloud.Operation
	var err error
	isExternal := data.AWS != nil || data.GCP != nil
	if isExternal {
		importReq := &network.ImportNetworkRequest{
			Name:        data.Name.ValueString(),
			ProjectId:   data.ProjectID.ValueString(),
			Description: data.Description.ValueString(),
		}

		switch {
		case data.AWS != nil:
			importReq.Params = &network.ImportNetworkRequest_Aws{
				Aws: &network.ImportAWSVPCRequest{
					RegionId:       data.RegionID.ValueString(),
					VpcId:          data.AWS.VPCID.ValueString(),
					AccountId:      data.AWS.AccountID.ValueString(),
					IamRoleArn:     data.AWS.IAMRoleARN.ValueString(),
					PrivateSubnets: data.AWS.PrivateSubnets.ValueBool(),
				},
			}
		case data.GCP != nil:
			importReq.Params = &network.ImportNetworkRequest_Google{
				Google: &network.ImportGoogleVPCRequest{
					ServiceAccountEmail: data.GCP.SAEmail.ValueString(),
					ProjectName:         data.GCP.ProjectName.ValueString(),
					NetworkName:         data.GCP.NetworkName.ValueString(),
					RegionId:            data.RegionID.ValueString(),
					SubnetworkName:      data.GCP.SubnetworkName.ValueString(),
				},
			}
		}

		opObj, err = r.networkService.Import(ctx, importReq)
	} else {
		opObj, err = r.networkService.Create(ctx, &network.CreateNetworkRequest{
			Name:          data.Name.ValueString(),
			CloudType:     data.CloudType.ValueString(),
			ProjectId:     data.ProjectID.ValueString(),
			Description:   data.Description.ValueString(),
			RegionId:      data.RegionID.ValueString(),
			Ipv4CidrBlock: data.Ipv4CidrBlock.ValueString(),
		})
	}

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
	data.IsExternal = types.BoolValue(isExternal)
	if isExternal {
		data.Ipv4CidrBlock = types.StringValue(net.Ipv4CidrBlock)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
	data.IsExternal = types.BoolValue(net.IsExternal)

	if net.IsExternal {
		switch {
		case net.GetAws() != nil:
			awsER := net.GetAws()
			if data.AWS == nil {
				data.AWS = &awsExternalNetworkResourceModel{}
			}
			data.AWS.IAMRoleARN = types.StringValue(awsER.IamRoleArn.GetValue())
			data.AWS.AccountID = types.StringValue(awsER.AccountId.GetValue())
			data.AWS.VPCID = types.StringValue(awsER.VpcId)
			data.AWS.PrivateSubnets = types.BoolValue(awsER.PrivateSubnets.GetValue())
		case net.GetGcp() != nil:
			gcpER := net.GetGcp()
			if data.GCP == nil {
				data.GCP = &googleExternalNetworkResourceModel{}
			}
			data.GCP.SAEmail = types.StringValue(gcpER.ServiceAccountEmail.GetValue())
			data.GCP.ProjectName = types.StringValue(gcpER.ProjectName.GetValue())
			data.GCP.NetworkName = types.StringValue(gcpER.NetworkName.GetValue())
			data.GCP.SubnetworkName = types.StringValue(gcpER.SubnetworkName.GetValue())
		default:
			resp.Diagnostics.AddError("unsupported network cloud provider", fmt.Sprintf("got: %v", net.GetExternalResources()))
			return
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Failed to update network", "networks don't support updates")
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
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NetworkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data *NetworkResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AWS != nil {
		if data.CloudType.ValueString() != "aws" && data.CloudType.ValueString() != "AWS" {
			resp.Diagnostics.AddAttributeError(
				path.Root("cloud_type"),
				"BYOC and \"cloud_type\" mismatch",
				fmt.Sprintf("Provided BYOC AWS configuration, but \"cloud_type\" is set to %q.", data.CloudType.ValueString()),
			)
		}
	}

	if data.GCP != nil {
		if data.CloudType.ValueString() != "gcp" && data.CloudType.ValueString() != "GCP" {
			resp.Diagnostics.AddAttributeError(
				path.Root("cloud_type"),
				"BYOC and \"cloud_type\" mismatch",
				fmt.Sprintf("Provided BYOC GCP configuration, but \"cloud_type\" is set to %q.", data.CloudType.ValueString()),
			)
		}
	}

	if data.AWS == nil && data.GCP == nil && data.Ipv4CidrBlock.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ipv4_cidr_block"),
			"ipv4_cidr_block must be specified.",
			"IPv4 CIDR block is required for non-BYOC networks.",
		)
	}
}
