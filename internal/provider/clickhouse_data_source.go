package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgen "github.com/doublecloud/go-sdk/gen/clickhouse"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClickhouseDataSource{}

func NewClickhouseDataSource() datasource.DataSource {
	return &ClickhouseDataSource{}
}

type ClickhouseDataSource struct {
	sdk *dcsdk.SDK
	svc *dcgen.ClusterServiceClient
}

type ClickhouseDataSourceModel struct {
	Id                    types.String              `tfsdk:"id"`
	ProjectID             types.String              `tfsdk:"project_id"`
	Name                  types.String              `tfsdk:"name"`
	Description           types.String              `tfsdk:"description"`
	RegionID              types.String              `tfsdk:"region_id"`
	CloudType             types.String              `tfsdk:"cloud_type"`
	Version               types.String              `tfsdk:"version"`
	ConnectionInfo        *ClickhoiseConnectionInfo `tfsdk:"connection_info"`
	PrivateConnectionInfo *ClickhoiseConnectionInfo `tfsdk:"private_connection_info"`
}

type ClickhoiseConnectionInfo struct {
	Host           types.String `tfsdk:"host"`
	User           types.String `tfsdk:"user"`
	Password       types.String `tfsdk:"password"`
	HttpsPort      types.Int64  `tfsdk:"https_port"`
	TcpPortSecure  types.Int64  `tfsdk:"tcp_port_secure"`
	NativeProtocol types.String `tfsdk:"native_protocol"`
	HttpsUri       types.String `tfsdk:"https_uri"`
	JdbcUri        types.String `tfsdk:"jdbc_uri"`
	OdbcUri        types.String `tfsdk:"odbc_uri"`
}

func (d *ClickhouseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse"
}

func clickhouseConenctionInfoSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"host": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Host to connect",
		},
		"user": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "ClickHouse user",
		},
		"password": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Password for ClickHouse user",
		},
		"https_port": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Port to connect using HTTPS protocol",
		},
		"tcp_port_secure": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: "Port to connect using TCP/native protocol",
		},
		"native_protocol": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Connection string for ClickHouse native protocol",
		},
		"https_uri": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "URI to connect using HTTPS protocol",
		},
		"jdbc_uri": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "URI to connect using JDBC protocol",
		},
		"odbc_uri": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "URI to connect using ODBC protocol",
		},
	}
}

func (d *ClickhouseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Clickhouse data soruce",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project identifier",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster identifier",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of cluster",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of cluster",
			},
			"region_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Region of cluster",
			},
			"cloud_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cloud type (aws, gcp, azure)",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Version of ClickHouse DBMS",
			},
			"connection_info": schema.SingleNestedAttribute{
				Optional:   true,
				Attributes: clickhouseConenctionInfoSchema(),
			},
			"private_connection_info": schema.SingleNestedAttribute{
				Optional:   true,
				Attributes: clickhouseConenctionInfoSchema(),
			},
		},
	}
}

func (d *ClickhouseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.svc = d.sdk.ClickHouse().Cluster()
}

func parseClickhouseConnectionInfo(r *clickhouse.ConnectionInfo) *ClickhoiseConnectionInfo {
	if r == nil {
		return nil
	}
	c := &ClickhoiseConnectionInfo{}
	c.Host = types.StringValue(r.Host)
	c.User = types.StringValue(r.User)
	c.Password = types.StringValue(r.Password)
	c.HttpsPort = types.Int64Value(r.HttpsPort.Value)
	c.TcpPortSecure = types.Int64Value(r.TcpPortSecure.Value)
	c.NativeProtocol = types.StringValue(r.NativeProtocol)
	c.HttpsUri = types.StringValue(r.HttpsUri)
	c.JdbcUri = types.StringValue(r.JdbcUri)
	c.OdbcUri = types.StringValue(r.OdbcUri)
	return c
}

func parseClickhousePrivateConnectionInfo(r *clickhouse.PrivateConnectionInfo) *ClickhoiseConnectionInfo {
	if r == nil {
		return nil
	}
	c := &ClickhoiseConnectionInfo{}
	c.Host = types.StringValue(r.Host)
	c.User = types.StringValue(r.User)
	c.Password = types.StringValue(r.Password)
	c.HttpsPort = types.Int64Value(r.HttpsPort.Value)
	c.TcpPortSecure = types.Int64Value(r.TcpPortSecure.Value)
	c.NativeProtocol = types.StringValue(r.NativeProtocol)
	c.HttpsUri = types.StringValue(r.HttpsUri)
	c.JdbcUri = types.StringValue(r.JdbcUri)
	c.OdbcUri = types.StringValue(r.OdbcUri)
	return c
}

func (d *ClickhouseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClickhouseDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id == types.StringNull() && data.Name == types.StringNull() {
		resp.Diagnostics.AddError("missing attribute", "specify one of: id or name")
		return
	}

	if data.Id == types.StringNull() {
		it := d.svc.ClusterIterator(ctx, &clickhouse.ListClustersRequest{ProjectId: data.ProjectID.ValueString()})
		for it.Next() {
			c := it.Value()
			if c.Name == data.Name.ValueString() {
				data.Id = types.StringValue(c.Id)
				break
			}
		}
		if data.Id == types.StringNull() {
			resp.Diagnostics.AddError("cluster not found", fmt.Sprintf("clickhouse cluster `%v` haven't found", data.Name.ValueString()))
			return
		}
	}

	response, err := d.svc.Get(ctx, &clickhouse.GetClusterRequest{ClusterId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	data.Description = types.StringValue(response.Description)
	data.CloudType = types.StringValue(response.CloudType)
	data.RegionID = types.StringValue(response.RegionId)
	data.Version = types.StringValue(response.Version)
	data.ConnectionInfo = parseClickhouseConnectionInfo(response.ConnectionInfo)
	data.PrivateConnectionInfo = parseClickhousePrivateConnectionInfo(response.PrivateConnectionInfo)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
