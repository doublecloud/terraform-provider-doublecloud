package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	dcgennet "github.com/doublecloud/go-sdk/gen/network"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetworkConnectionModel struct {
	ID          types.String `tfsdk:"id"`
	NetworkID   types.String `tfsdk:"network_id"`
	Description types.String `tfsdk:"description"`

	AWS    *awsNetworkConnectionInfo    `tfsdk:"aws"`
	Google *googleNetworkConnectionInfo `tfsdk:"google"`

	status       string
	statusReason string
}

type awsNetworkConnectionInfo struct {
	Peering *awsNetworkConnectionPeeringInfo `tfsdk:"peering"`
}

type awsNetworkConnectionPeeringInfo struct {
	VPCID         types.String `tfsdk:"vpc_id"`
	AccountID     types.String `tfsdk:"account_id"`
	RegionID      types.String `tfsdk:"region_id"`
	IPv4CIDRBlock types.String `tfsdk:"ipv4_cidr_block"`
	IPv6CIDRBlock types.String `tfsdk:"ipv6_cidr_block"`

	PeeringConnectionID  types.String `tfsdk:"peering_connection_id"`
	ManagedIPv4CIDRBlock types.String `tfsdk:"managed_ipv4_cidr_block"`
	ManagedIPv6CIDRBlock types.String `tfsdk:"managed_ipv6_cidr_block"`
}

type googleNetworkConnectionInfo struct {
	Name           types.String `tfsdk:"name"`
	PeerNetworkURL types.String `tfsdk:"peer_network_url"`

	ManagedNetworkURL types.String `tfsdk:"managed_network_url"`
}

var (
	networkConnectionResourceSchema = resourceschema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network Connection resource",

		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Network Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": resourceschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": resourceschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of network connection",
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws": resourceschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "AWS connection info",
				Attributes: map[string]resourceschema.Attribute{
					"peering": resourceschema.SingleNestedAttribute{
						Required:            true,
						MarkdownDescription: "VPC Peering connection info",
						Attributes: map[string]resourceschema.Attribute{
							"vpc_id": resourceschema.StringAttribute{
								Required:            true,
								MarkdownDescription: "ID of the VPC",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"account_id": resourceschema.StringAttribute{
								Required:            true,
								MarkdownDescription: "ID of the VPC owner account",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"region_id": resourceschema.StringAttribute{
								Required:            true,
								MarkdownDescription: "ID of the AWS region",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"ipv4_cidr_block": resourceschema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Customer IPv4 CIDR block.\nDoubleCloud will create route to this CIDR using Peering Connection.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"ipv6_cidr_block": resourceschema.StringAttribute{
								Optional:            true,
								Computed:            true,
								MarkdownDescription: "Customer IPv6 CIDR block.\nDoubleCloud will create route to this CIDR using Peering Connection.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
								Default: stringdefault.StaticString(""),
							},
							"peering_connection_id": resourceschema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Peering Connection ID.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"managed_ipv4_cidr_block": resourceschema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Managed AWS IPv4 CIDR block.\nCustomer should create route to this CIDR using Peering Connection.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"managed_ipv6_cidr_block": resourceschema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "Managed AWS IPv6 CIDR block.\nCustomer should create route to this CIDR using Peering Connection.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("aws"),
						path.MatchRoot("google"),
					}...),
				},
			},
			"google": resourceschema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Google Cloud connection info",
				Attributes: map[string]resourceschema.Attribute{
					"name": resourceschema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of this peering",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"peer_network_url": resourceschema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The URL of the peer network",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"managed_network_url": resourceschema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "The URL of the managed GCP network",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("aws"),
						path.MatchRoot("google"),
					}...),
				},
			},
		},
	}
)

func generateNetworkConnectionDatasourceSchema(diags diag.Diagnostics) dataschema.Schema {
	attrs := make(map[string]dataschema.Attribute)
	diags.Append(convertSchemaAttributes(networkConnectionResourceSchema.Attributes, attrs)...)
	res := dataschema.Schema{
		MarkdownDescription: "Network Connection datasource",
		Attributes:          attrs,
	}

	id := res.Attributes["id"].(*dataschema.StringAttribute)
	id.Computed = false
	id.Required = true

	return res
}

func (m *NetworkConnectionModel) FromProtobuf(nc *network.NetworkConnection) error {
	m.NetworkID = types.StringValue(nc.NetworkId)
	m.ID = types.StringValue(nc.Id)
	m.Description = types.StringValue(nc.Description)
	m.status = nc.Status.String()
	m.statusReason = nc.StatusReason

	switch {
	case nc.GetAws() != nil:
		if m.AWS == nil {
			m.AWS = &awsNetworkConnectionInfo{}
		}
		switch {
		case nc.GetAws().GetPeering() != nil:
			if m.AWS.Peering == nil {
				m.AWS.Peering = &awsNetworkConnectionPeeringInfo{}
			}
			params := nc.GetAws().GetPeering()
			peering := m.AWS.Peering

			peering.PeeringConnectionID = types.StringValue(params.PeeringConnectionId)
			peering.ManagedIPv4CIDRBlock = types.StringValue(params.ManagedIpv4CidrBlock)
			peering.ManagedIPv6CIDRBlock = types.StringValue(params.ManagedIpv6CidrBlock)
			peering.VPCID = types.StringValue(params.VpcId)
			peering.AccountID = types.StringValue(params.AccountId)
			peering.RegionID = types.StringValue(params.RegionId)
			peering.IPv4CIDRBlock = types.StringValue(params.Ipv4CidrBlock)
			peering.IPv6CIDRBlock = types.StringValue(params.Ipv6CidrBlock)
		default:
			return fmt.Errorf("unsupported type of AWS connection: %v", nc.GetAws().GetType())
		}
	case nc.GetGoogle() != nil:
		if m.Google == nil {
			m.Google = &googleNetworkConnectionInfo{}
		}
		params := nc.GetGoogle()
		peering := m.Google

		peering.Name = types.StringValue(params.Name)
		peering.PeerNetworkURL = types.StringValue(params.PeerNetworkUrl)
		peering.ManagedNetworkURL = types.StringValue(params.ManagedNetworkUrl)
	default:
		return fmt.Errorf("unsupported type of network connection: %v", nc.GetConnectionInfo())
	}

	return nil
}

func (m *NetworkConnectionModel) IsReady() bool {
	return m.status == network.NetworkConnection_NETWORK_CONNECTION_STATUS_ACTIVE.String() || m.status == network.NetworkConnection_NETWORK_CONNECTION_STATUS_ERROR.String()
}

func (m *NetworkConnectionModel) IsOK() (bool, string) {
	if m.status == network.NetworkConnection_NETWORK_CONNECTION_STATUS_ERROR.String() {
		return false, fmt.Sprintf("network connection %s is in ERROR state, reason: %s", m.ID, m.statusReason)
	}
	return true, ""
}

func (m *NetworkConnectionModel) Poll(ctx context.Context, client *dcgennet.NetworkConnectionServiceClient) diag.Diagnostics {
	return getNetworkConnection(ctx, client, m.ID.ValueString(), m)
}

func getNetworkConnection(
	ctx context.Context,
	client *dcgennet.NetworkConnectionServiceClient,
	id string,
	data *NetworkConnectionModel,
) diag.Diagnostics {
	var diags diag.Diagnostics
	nc, err := client.Get(ctx, &network.GetNetworkConnectionRequest{NetworkConnectionId: id})
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
