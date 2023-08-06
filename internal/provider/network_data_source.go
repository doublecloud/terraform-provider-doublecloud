package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgennet "github.com/doublecloud/go-sdk/gen/network"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NetworkDataSource{}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

// NetworkDataSource defines the data source implementation.
type NetworkDataSource struct {
	sdk            *dcsdk.SDK
	networkService *dcgennet.NetworkServiceClient
}

type NetworkDataSourceModel struct {
	Id            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	RegionID      types.String `tfsdk:"region_id"`
	CloudType     types.String `tfsdk:"cloud_type"`
	Ipv4CidrBlock types.String `tfsdk:"ipv4_cidr_block"`
}

func (d *NetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network data source",

		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project identifier",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network identifier",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of network",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of network",
			},
			"region_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Region of network",
			},
			"ipv4_cidr_block": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The IPv4 network range for the subnet, in CIDR notation. For example, 10.0.0.0/16.",
			},
			"cloud_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cloud type (aws, gcp, azure)",
			},
		},
	}
}

func (d *NetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.sdk = sdk
	d.networkService = d.sdk.Network().Network()
}

func (d *NetworkDataSource) getNetworkIdByName(ctx context.Context, m *NetworkDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	it := d.networkService.NetworkIterator(ctx, &network.ListNetworksRequest{ProjectId: m.ProjectID.ValueString()})
	for it.Next() {
		n := it.Value()
		if n.Name == m.Name.ValueString() {
			m.Id = types.StringValue(n.Id)
			return diags
		}
	}
	diags.AddError("network not found", fmt.Sprintf("network with name `%v` haven't found", m.Name.ValueString()))

	return diags
}

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id == types.StringNull() {
		diag := d.getNetworkIdByName(ctx, &data)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	}

	net, err := d.networkService.Get(ctx, &network.GetNetworkRequest{NetworkId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	data.Id = types.StringValue(net.Id)
	data.Name = types.StringValue(net.Name)
	data.ProjectID = types.StringValue(net.ProjectId)
	data.Description = types.StringValue(net.Description)
	data.CloudType = types.StringValue(net.CloudType)
	data.RegionID = types.StringValue(net.RegionId)
	data.Ipv4CidrBlock = types.StringValue(net.Ipv4CidrBlock)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
