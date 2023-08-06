package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
)

type endpointClickhouseConnectionOptions struct {
	Database types.String                         `tfsdk:"database"`
	User     types.String                         `tfsdk:"user"`
	Password types.String                         `tfsdk:"password"`
	Address  *endpointClickhouseConnectionAddress `tfsdk:"address"`
}

type endpointClickhouseConnectionAddress struct {
	ClusterId types.String         `tfsdk:"cluster_id"`
	OnPremise *onPremiseClickhouse `tfsdk:"on_premise"`
}

type endpointClickhouseShards struct {
	Name  types.String   `tfsdk:"name"`
	Hosts []types.String `tfsdk:"hosts"`
}

type onPremiseClickhouse struct {
	Shards     []endpointClickhouseShards `tfsdk:"shard"`
	HttpPort   types.Int64                `tfsdk:"http_port"`
	NativePort types.Int64                `tfsdk:"native_port"`
	TLSMode    *endpointTLSMode           `tfsdk:"tls_mode"`
}

type endpointClickhouseSourceSettings struct {
	IncludeTables []types.String                       `tfsdk:"include_tables"`
	ExcludeTables []types.String                       `tfsdk:"exclude_tables"`
	Connection    *endpointClickhouseConnectionOptions `tfsdk:"connection"`
}

type altName struct {
	FromName types.String `tfsdk:"from_name"`
	ToName   types.String `tfsdk:"to_name"`
}

type endpointClickhouseTargetSettings struct {
	Connection              *endpointClickhouseConnectionOptions `tfsdk:"connection"`
	ClickhouseClusterName   types.String                         `tfsdk:"clickhouse_cluster_name"`
	AltNames                []altName                            `tfsdk:"alt_name"`
	ClickhouseCleanupPolicy types.String                         `tfsdk:"clickhouse_cleanup_policy"`
	// TODO: sharding
}

func transferEndpointClickhouseConnectionSchemaBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{Optional: true},
			"user":     schema.StringAttribute{Optional: true},
			"password": schema.StringAttribute{Optional: true, Sensitive: true},
		},
		Blocks: map[string]schema.Block{
			"address": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"cluster_id": schema.StringAttribute{Optional: true},
				},
				Blocks: map[string]schema.Block{
					"on_premise": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"http_port":   schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(8443)},
							"native_port": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(8443)},
						},
						Blocks: map[string]schema.Block{
							"shard": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name":  schema.StringAttribute{Optional: true},
										"hosts": schema.ListAttribute{ElementType: types.StringType, Optional: true},
									},
								},
							},
							"tls_mode": transferEndpointTLSMode(),
						},
					},
				},
			},
		},
	}
}

func transferEndpointChSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"include_tables": schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"exclude_tables": schema.ListAttribute{ElementType: types.StringType, Optional: true},
		},
		Blocks: map[string]schema.Block{
			"connection": transferEndpointClickhouseConnectionSchemaBlock(),
		},
	}
}

func convertConnectionOptions(m *endpointClickhouseConnectionOptions) (*endpoint.ClickhouseConnectionOptions, diag.Diagnostics) {
	var diag diag.Diagnostics

	options := &endpoint.ClickhouseConnectionOptions{}
	options.Database = m.Database.ValueString()
	options.User = m.User.ValueString()
	options.Password = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Password.ValueString()}}

	if cluster_id := m.Address.ClusterId; !cluster_id.IsNull() {
		options.Address = &endpoint.ClickhouseConnectionOptions_MdbClusterId{MdbClusterId: cluster_id.ValueString()}
	}
	if on_premise := m.Address.OnPremise; on_premise != nil {
		opts := &endpoint.OnPremiseClickhouse{}
		options.Address = &endpoint.ClickhouseConnectionOptions_OnPremise{OnPremise: opts}
		opts.HttpPort = m.Address.OnPremise.HttpPort.ValueInt64()
		opts.NativePort = m.Address.OnPremise.NativePort.ValueInt64()
		opts.TlsMode = convertTLSMode(m.Address.OnPremise.TLSMode)

		shards := m.Address.OnPremise.Shards
		opts.Shards = make([]*endpoint.ClickhouseShard, len(shards))
		for i := 0; i < len(shards); i++ {
			opts.Shards[i] = &endpoint.ClickhouseShard{
				Name:  shards[i].Name.ValueString(),
				Hosts: convertSliceTFStrings(shards[i].Hosts),
			}
		}
	}

	if options.Address == nil {
		diag.AddError("unknown connection", "required one of fields: cluster_id or on_premise")
	}
	return options, diag
}

func chSourceEndpointSettings(m *endpointClickhouseSourceSettings) (*transfer.EndpointSettings_ClickhouseSource, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_ClickhouseSource{ClickhouseSource: &endpoint.ClickhouseSource{}}
	var diag diag.Diagnostics
	if m.IncludeTables != nil {
		settings.ClickhouseSource.IncludeTables = convertSliceTFStrings(m.IncludeTables)
	}
	if m.ExcludeTables != nil {
		settings.ClickhouseSource.ExcludeTables = convertSliceTFStrings(m.ExcludeTables)
	}
	options, diag := convertConnectionOptions(m.Connection)

	if diag.HasError() {
		return nil, diag
	}

	settings.ClickhouseSource.Connection = &endpoint.ClickhouseConnection{
		Connection: &endpoint.ClickhouseConnection_ConnectionOptions{
			ConnectionOptions: options,
		},
	}
	return settings, nil
}

func transferEndpointChTargetSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"clickhouse_cluster_name": schema.StringAttribute{
				MarkdownDescription: "clickhouse_cluster_name",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"clickhouse_cleanup_policy": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{transferEndpointCleanupPolicyValidator()},
				Default:    stringdefault.StaticString("DISABLED"),
			},
		},
		Blocks: map[string]schema.Block{
			"connection": transferEndpointClickhouseConnectionSchemaBlock(),
			"alt_name": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"from_name": schema.StringAttribute{Optional: true},
						"to_name":   schema.StringAttribute{Optional: true},
					},
				},
			},
		},
	}
}

func chTargetEndpointSettings(m *endpointClickhouseTargetSettings) (*transfer.EndpointSettings_ClickhouseTarget, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_ClickhouseTarget{ClickhouseTarget: &endpoint.ClickhouseTarget{}}
	var diag diag.Diagnostics

	if !m.ClickhouseClusterName.IsUnknown() {
		settings.ClickhouseTarget.ClickhouseClusterName = m.ClickhouseClusterName.ValueString()
	}
	if m.AltNames != nil {
		altNames := make([]*endpoint.AltName, len(m.AltNames))
		for i := 0; i < len(m.AltNames); i++ {
			v := m.AltNames[i]
			altNames[i] = &endpoint.AltName{FromName: v.FromName.ValueString(), ToName: v.ToName.ValueString()}
		}
		settings.ClickhouseTarget.AltNames = altNames
	}

	if v := m.ClickhouseCleanupPolicy; !v.IsUnknown() {
		settings.ClickhouseTarget.CleanupPolicy = endpoint.ClickhouseCleanupPolicy(endpoint.CleanupPolicy_value[v.ValueString()])
	}

	options, diag := convertConnectionOptions(m.Connection)

	if diag.HasError() {
		return nil, diag
	}

	settings.ClickhouseTarget.Connection = &endpoint.ClickhouseConnection{
		Connection: &endpoint.ClickhouseConnection_ConnectionOptions{
			ConnectionOptions: options,
		},
	}

	return settings, nil
}

func parseTransferEndpointClickhouseConnection(ctx context.Context, e *endpoint.ClickhouseConnection, c *endpointClickhouseConnectionOptions) diag.Diagnostics {
	var diag diag.Diagnostics

	opts := e.GetConnectionOptions()
	c.User = types.StringValue(opts.User)
	c.Database = types.StringValue(opts.Database)
	if addr := opts.GetMdbClusterId(); addr != "" {
		c.Address.ClusterId = types.StringValue(addr)
	}
	if addr := opts.GetOnPremise(); addr != nil {
		on_prem := c.Address.OnPremise
		on_prem.HttpPort = types.Int64Value(addr.HttpPort)
		on_prem.NativePort = types.Int64Value(addr.NativePort)
		if addr.TlsMode != nil {
			if disabled := addr.TlsMode.GetDisabled(); disabled != nil {
				on_prem.TLSMode = nil
			}
			if config := addr.TlsMode.GetEnabled(); config != nil {
				on_prem.TLSMode = &endpointTLSMode{CACertificate: types.StringValue(config.CaCertificate)}
			}
		}

		if addr.Shards != nil {
			on_prem.Shards = make([]endpointClickhouseShards, len(addr.Shards))
			for i := 0; i < len(addr.Shards); i++ {
				on_prem.Shards[i] = endpointClickhouseShards{
					Name:  types.StringValue(addr.Shards[i].Name),
					Hosts: convertSliceToTFStrings(addr.Shards[i].Hosts),
				}
			}
		}
	}

	return diag
}
