package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

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
	Id                    types.String                 `tfsdk:"id"`
	ProjectID             types.String                 `tfsdk:"project_id"`
	Name                  types.String                 `tfsdk:"name"`
	Description           types.String                 `tfsdk:"description"`
	RegionID              types.String                 `tfsdk:"region_id"`
	CloudType             types.String                 `tfsdk:"cloud_type"`
	Version               types.String                 `tfsdk:"version"`
	ConnectionInfo        *ClickhouseConnectionInfo    `tfsdk:"connection_info"`
	PrivateConnectionInfo *ClickhouseConnectionInfo    `tfsdk:"private_connection_info"`
	CustomCertificate     *ClickhouseCustomCertificate `tfsdk:"custom_certificate"`
}

type ClickhouseConnectionInfo struct {
	Host               types.String `tfsdk:"host"`
	User               types.String `tfsdk:"user"`
	Password           types.String `tfsdk:"password"`
	HttpsPort          types.Int64  `tfsdk:"https_port"`
	TcpPortSecure      types.Int64  `tfsdk:"tcp_port_secure"`
	NativeProtocol     types.String `tfsdk:"native_protocol"`
	HttpsUri           types.String `tfsdk:"https_uri"`
	JdbcUri            types.String `tfsdk:"jdbc_uri"`
	OdbcUri            types.String `tfsdk:"odbc_uri"`
	HttpsPortCTLS      types.Int64  `tfsdk:"https_port_ctls"`
	TcpPortSecureCTLS  types.Int64  `tfsdk:"tcp_port_secure_ctls"`
	NativeProtocolCTLS types.String `tfsdk:"native_protocol_ctls"`
	HttpsUriCTLS       types.String `tfsdk:"https_uri_ctls"`
}

func (ci ClickhouseConnectionInfo) convert(diags diag.Diagnostics) types.Object {
	res, d := types.ObjectValue(map[string]attr.Type{
		"host":                 types.StringType,
		"user":                 types.StringType,
		"password":             types.StringType,
		"https_port":           types.Int64Type,
		"tcp_port_secure":      types.Int64Type,
		"native_protocol":      types.StringType,
		"https_uri":            types.StringType,
		"jdbc_uri":             types.StringType,
		"odbc_uri":             types.StringType,
		"https_port_ctls":      types.Int64Type,
		"tcp_port_secure_ctls": types.Int64Type,
		"native_protocol_ctls": types.StringType,
		"https_uri_ctls":       types.StringType,
	},
		map[string]attr.Value{
			"host":                 ci.Host,
			"user":                 ci.User,
			"password":             ci.Password,
			"https_port":           ci.HttpsPort,
			"tcp_port_secure":      ci.TcpPortSecure,
			"native_protocol":      ci.NativeProtocol,
			"https_uri":            ci.HttpsUri,
			"jdbc_uri":             ci.JdbcUri,
			"odbc_uri":             ci.OdbcUri,
			"https_port_ctls":      ci.HttpsPortCTLS,
			"tcp_port_secure_ctls": ci.TcpPortSecureCTLS,
			"native_protocol_ctls": ci.NativeProtocolCTLS,
			"https_uri_ctls":       ci.HttpsUriCTLS,
		},
	)
	diags.Append(d...)
	return res
}

type ClickhouseCustomCertificate struct {
	Certificate types.String `tfsdk:"certificate"`
	Key         types.String `tfsdk:"key"`
	RootCA      types.String `tfsdk:"root_ca"`
}

func (cc *ClickhouseCustomCertificate) convert() *clickhouseCustomCertificate {
	if cc == nil {
		return nil
	}
	res := clickhouseCustomCertificate{
		Certificate: cc.Certificate,
		Key:         cc.Key,
		RootCA:      cc.RootCA,
	}
	return &res
}

func (d *ClickhouseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse"
}

func (d *ClickhouseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	connInfo := make(map[string]schema.Attribute)
	resp.Diagnostics.Append(convertSchemaAttributes(clickhouseConenctionInfoSchema(), connInfo)...)
	customCertificate := make(map[string]schema.Attribute)
	resp.Diagnostics.Append(convertSchemaAttributes(clickhouseCustomCertificateSchema(), customCertificate)...)
	resp.Schema = schema.Schema{
		MarkdownDescription: "Clickhouse data source",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project ID",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster ID",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster name",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster description",
			},
			"region_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Region where the cluster is located",
			},
			"cloud_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cloud provider (`aws`, `gcp`, or `azure`)",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Version of the ClickHouse DBMS",
			},
			"connection_info": schema.SingleNestedAttribute{
				Computed:            true,
				Attributes:          connInfo,
				MarkdownDescription: "Public connection info",
			},
			"private_connection_info": schema.SingleNestedAttribute{
				Computed:            true,
				Attributes:          connInfo,
				MarkdownDescription: "Private connection info",
			},
			"custom_certificate": schema.SingleNestedAttribute{
				Computed:            true,
				Attributes:          customCertificate,
				MarkdownDescription: "Custom TLS certificate",
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

func parseClickhouseConnectionInfo(r *clickhouse.ConnectionInfo) *ClickhouseConnectionInfo {
	if r == nil {
		return nil
	}
	c := &ClickhouseConnectionInfo{}
	c.Host = types.StringValue(r.Host)
	c.User = types.StringValue(r.User)
	c.Password = types.StringValue(r.Password)
	c.HttpsPort = types.Int64Value(r.HttpsPort.Value)
	c.TcpPortSecure = types.Int64Value(r.TcpPortSecure.Value)
	c.NativeProtocol = types.StringValue(r.NativeProtocol)
	c.HttpsUri = types.StringValue(r.HttpsUri)
	c.JdbcUri = types.StringValue(r.JdbcUri)
	c.OdbcUri = types.StringValue(r.OdbcUri)
	c.HttpsPortCTLS = types.Int64Value(r.HttpsPortCtls.Value)
	c.TcpPortSecureCTLS = types.Int64Value(r.TcpPortSecureCtls.Value)
	c.NativeProtocolCTLS = types.StringValue(r.NativeProtocolCtls)
	c.HttpsUriCTLS = types.StringValue(r.HttpsUriCtls)
	return c
}

func parseClickhousePrivateConnectionInfo(r *clickhouse.PrivateConnectionInfo) *ClickhouseConnectionInfo {
	if r == nil {
		return nil
	}
	c := &ClickhouseConnectionInfo{}
	c.Host = types.StringValue(r.Host)
	c.User = types.StringValue(r.User)
	c.Password = types.StringValue(r.Password)
	c.HttpsPort = types.Int64Value(r.HttpsPort.Value)
	c.TcpPortSecure = types.Int64Value(r.TcpPortSecure.Value)
	c.NativeProtocol = types.StringValue(r.NativeProtocol)
	c.HttpsUri = types.StringValue(r.HttpsUri)
	c.JdbcUri = types.StringValue(r.JdbcUri)
	c.OdbcUri = types.StringValue(r.OdbcUri)
	c.HttpsPortCTLS = types.Int64Value(r.HttpsPortCtls.Value)
	c.TcpPortSecureCTLS = types.Int64Value(r.TcpPortSecureCtls.Value)
	c.NativeProtocolCTLS = types.StringValue(r.NativeProtocolCtls)
	c.HttpsUriCTLS = types.StringValue(r.HttpsUriCtls)
	return c
}

func parseClickhouseCustomCertificate(r *clickhouse.CustomCertificate, oldKey string, diags diag.Diagnostics) *ClickhouseCustomCertificate {
	if r == nil {
		return nil
	}

	if !r.GetEnabled() {
		return nil
	}

	certRaw := string(r.Certificate.GetValue()[:])
	if len(certRaw) == 0 {
		return nil
	}

	certificate := types.StringValue(certRaw)
	s, err := strconv.Unquote(oldKey)
	if err != nil {
		diags.AddError("Can't unquote TLS certificate", err.Error())
		return nil
	}
	key := basetypes.NewStringValue(s)
	rootCa := types.StringNull()
	rootRaw := string(r.RootCa.GetValue()[:])
	if len(rootRaw) > 0 {
		rootCa = types.StringValue(rootRaw)
	}
	c := &ClickhouseCustomCertificate{
		Certificate: certificate,
		Key:         key,
		RootCA:      rootCa,
	}
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
		if it.Error() != nil {
			resp.Diagnostics.AddError("iterator has failed", it.Error().Error())
		}
		if data.Id == types.StringNull() {
			resp.Diagnostics.AddError("cluster not found", fmt.Sprintf("clickhouse cluster `%v` haven't found", data.Name.ValueString()))
			return
		}
	}

	response, err := d.svc.Get(ctx, &clickhouse.GetClusterRequest{
		ClusterId: data.Id.ValueString(),
		Sensitive: true,
	})
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
	oldKey := ""
	if data.CustomCertificate != nil {
		oldKey = data.CustomCertificate.Key.String()
	}
	data.CustomCertificate = parseClickhouseCustomCertificate(response.CustomCertificate, oldKey, resp.Diagnostics)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
