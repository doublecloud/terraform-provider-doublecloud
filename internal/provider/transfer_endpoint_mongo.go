package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointMongoSourceSettings struct {
	Connection             *endpointMongoConnection  `tfsdk:"connection"`
	Collections            []endpointMongoCollection `tfsdk:"collection"`
	ExcludedCollections    []endpointMongoCollection `tfsdk:"excluded_collection"`
	SecondaryPreferredMode types.Bool                `tfsdk:"secondary_preferred_mode"`
}

type endpointMongoConnection struct {
	ConnectionType *endpointMongoConnectionType `tfsdk:"connection_type"`
	User           types.String                 `tfsdk:"user"`
	Password       types.String                 `tfsdk:"password"`
	AuthSource     types.String                 `tfsdk:"auth_source"`
}

type endpointMongoConnectionType struct {
	OnPremise  *endpointMongoConnectionOnPremise `tfsdk:"on_premise"`
	Srv        *endpointMongoConnectionSrv       `tfsdk:"srv"`
	TLSMode    *endpointTLSMode                  `tfsdk:"tls_mode"`
	ReplicaSet types.String                      `tfsdk:"replica_set"`
}

type endpointMongoConnectionOnPremise struct {
	Hosts []types.String `tfsdk:"hosts"`
	Port  types.Int64    `tfsdk:"port"`
}

type endpointMongoConnectionSrv struct {
	Hostname types.String `tfsdk:"hostname"`
}

type endpointMongoCollection struct {
	DatabaseName   types.String `tfsdk:"database_name"`
	CollectionName types.String `tfsdk:"collection_name"`
}

type endpointMongoTargetSettings struct {
	Connection    *endpointMongoConnection `tfsdk:"connection"`
	Database      types.String             `tfsdk:"database"`
	CleanupPolicy types.String             `tfsdk:"cleanup_policy"`
}

func transferEndpointMongoSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"secondary_preferred_mode": schema.BoolAttribute{
				MarkdownDescription: "Read mode of the MongoDB client",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"connection":          transferEndpointMongoConnectionSchema(),
			"collection":          transferEndpointMongoCollectionSchema(),
			"excluded_collection": transferEndpointMongoCollectionSchema(),
		},
	}
}

func transferEndpointMongoCollectionSchema() schema.Block {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"database_name": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Database name",
				},
				"collection_name": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: "Collection name",
				},
			},
		},
	}
}

func transferEndpointMongoConnectionSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Database user",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Database user password",
			},
			"auth_source": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Authentication database associated with the user",
			},
		},
		Blocks: map[string]schema.Block{
			"connection_type": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"replica_set": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Replica set",
					},
				},
				Blocks: map[string]schema.Block{
					"tls_mode":   transferEndpointTLSMode(),
					"srv":        transferEndpointMongoConnectionSrvSchema(),
					"on_premise": transferEndpointMongoConnectionOnPremiseSchema(),
				},
			},
		},
	}
}

func transferEndpointMongoConnectionOnPremiseSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"hosts": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of hosts",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Port",
			},
		},
	}
}
func transferEndpointMongoConnectionSrvSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"hostname": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "SRV hostname",
			},
		},
	}
}

func transferEndpointMongoTargetSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Database",
			},
			"cleanup_policy": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{transferEndpointCleanupPolicyValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Cleanup policy",
			},
		},
		Blocks: map[string]schema.Block{
			"connection": transferEndpointMongoConnectionSchema(),
		},
	}
}

func (m *endpointMongoConnection) convert() (*endpoint.MongoConnectionOptions, diag.Diagnostics) {
	var diags diag.Diagnostics
	options := &endpoint.MongoConnectionOptions{}

	if m == nil {
		diags.AddError("Connection block missing", "Specify a connection block")
		return nil, diags
	}

	options.User = m.User.ValueString()
	options.Password = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Password.ValueString()}}

	if !m.AuthSource.IsNull() {
		options.AuthSource = m.AuthSource.ValueString()
	}

	replicaSet := m.ConnectionType.ReplicaSet
	tlsMode := m.ConnectionType.TLSMode

	if srv := m.ConnectionType.Srv; srv != nil {
		opts := &endpoint.SrvMongo{}
		if hostname := srv.Hostname; !hostname.IsNull() {
			opts.Hostname = hostname.ValueString()
		}
		if !replicaSet.IsNull() {
			opts.ReplicaSet = replicaSet.ValueString()
		}
		if tlsMode != nil {
			opts.TlsMode = convertTLSMode(tlsMode)
		}

		options.Address = &endpoint.MongoConnectionOptions_Srv{Srv: opts}
	}

	if on_premise := m.ConnectionType.OnPremise; on_premise != nil {
		opts := &endpoint.OnPremiseMongo{}
		opts.Hosts = convertSliceTFStrings(on_premise.Hosts)
		if port := on_premise.Port; !port.IsNull() {
			opts.Port = port.ValueInt64()
		}
		if !replicaSet.IsNull() {
			opts.ReplicaSet = replicaSet.ValueString()
		}
		if tlsMode != nil {
			opts.TlsMode = convertTLSMode(tlsMode)
		}

		options.Address = &endpoint.MongoConnectionOptions_OnPremise{OnPremise: opts}
	}

	if options.Address == nil {
		diags.AddError("unknown connection", "required one of fields: srv or on_premise")
	}

	return options, diags
}

func (m *endpointMongoCollection) convert() *endpoint.MongoCollection {
	ret := endpoint.MongoCollection{}
	ret.DatabaseName = m.DatabaseName.ValueString()
	if !m.CollectionName.IsNull() {
		ret.CollectionName = m.CollectionName.ValueString()
	}
	return &ret
}

func (m *endpointMongoSourceSettings) convert() (*transfer.EndpointSettings_MongoSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	ret := endpoint.MongoSource{}

	cnn, d := m.Connection.convert()
	diags.Append(d...)
	ret.Connection = &endpoint.MongoConnection{
		Connection: &endpoint.MongoConnection_ConnectionOptions{
			ConnectionOptions: cnn,
		},
	}

	if len(m.Collections) != 0 {
		ret.Collections = make([]*endpoint.MongoCollection, len(m.Collections))
		for i := 0; i < len(m.Collections); i++ {
			ret.Collections[i] = m.Collections[i].convert()
		}
	}
	if len(m.ExcludedCollections) != 0 {
		ret.ExcludedCollections = make([]*endpoint.MongoCollection, len(m.ExcludedCollections))
		for i := 0; i < len(m.ExcludedCollections); i++ {
			ret.ExcludedCollections[i] = m.ExcludedCollections[i].convert()
		}
	}

	if !m.SecondaryPreferredMode.IsNull() {
		ret.SecondaryPreferredMode = m.SecondaryPreferredMode.ValueBool()
	}

	return &transfer.EndpointSettings_MongoSource{MongoSource: &ret}, diags
}

func (m *endpointMongoTargetSettings) convert() (*transfer.EndpointSettings_MongoTarget, diag.Diagnostics) {
	var diags diag.Diagnostics
	ret := endpoint.MongoTarget{}

	cnn, d := m.Connection.convert()
	diags.Append(d...)

	ret.Connection = &endpoint.MongoConnection{
		Connection: &endpoint.MongoConnection_ConnectionOptions{
			ConnectionOptions: cnn,
		},
	}

	if !m.Database.IsNull() {
		ret.Database = m.Database.ValueString()
	}
	if !m.CleanupPolicy.IsNull() {
		ret.CleanupPolicy = endpoint.CleanupPolicy(endpoint.CleanupPolicy_value[m.CleanupPolicy.ValueString()])
	}

	return &transfer.EndpointSettings_MongoTarget{MongoTarget: &ret}, diags
}

func (m *endpointMongoConnection) parse(e *endpoint.MongoConnection) diag.Diagnostics {
	var diags diag.Diagnostics
	if e == nil {
		m = nil
	}

	opts := e.GetConnectionOptions()
	m.User = types.StringValue(opts.User)
	m.AuthSource = types.StringValue(opts.AuthSource)

	var replicaSet types.String
	var tlsMode *endpointTLSMode
	if srv := opts.GetSrv(); srv != nil {
		if !m.ConnectionType.Srv.Hostname.IsNull() {
			m.ConnectionType.Srv.Hostname = types.StringValue(srv.Hostname)
		}

		replicaSet = types.StringValue(srv.ReplicaSet)

		if srv.TlsMode != nil {
			if disabled := srv.TlsMode.GetDisabled(); disabled != nil {
				tlsMode = nil
			}
			if config := srv.TlsMode.GetEnabled(); config != nil {
				tlsMode = &endpointTLSMode{CACertificate: types.StringValue(config.CaCertificate)}
			}
		}
	}
	if on_premise := opts.GetOnPremise(); on_premise != nil {
		if m.ConnectionType.OnPremise.Hosts != nil {
			m.ConnectionType.OnPremise.Hosts = convertSliceToTFStrings(on_premise.Hosts)
		}

		if !m.ConnectionType.OnPremise.Port.IsNull() {
			m.ConnectionType.OnPremise.Port = types.Int64Value(on_premise.Port)
		}

		replicaSet = types.StringValue(on_premise.ReplicaSet)

		if on_premise.TlsMode != nil {
			if disabled := on_premise.TlsMode.GetDisabled(); disabled != nil {
				tlsMode = nil
			}
			if config := on_premise.TlsMode.GetEnabled(); config != nil {
				tlsMode = &endpointTLSMode{CACertificate: types.StringValue(config.CaCertificate)}
			}
		}
	}

	if !m.ConnectionType.ReplicaSet.IsNull() {
		m.ConnectionType.ReplicaSet = replicaSet
	}
	if m.ConnectionType.TLSMode != nil {
		m.ConnectionType.TLSMode = tlsMode
	}

	return diags
}

func (m *endpointMongoSourceSettings) parse(e *endpoint.MongoSource) diag.Diagnostics {
	var diag diag.Diagnostics

	m.Connection.parse(e.Connection)
	if !m.SecondaryPreferredMode.IsNull() {
		m.SecondaryPreferredMode = types.BoolValue(e.SecondaryPreferredMode)
	}

	m.Collections = make([]endpointMongoCollection, len(e.Collections))
	for i := 0; i < len(e.Collections); i++ {
		m.Collections[i] = endpointMongoCollection{
			DatabaseName:   types.StringValue(e.Collections[i].DatabaseName),
			CollectionName: types.StringValue(e.Collections[i].CollectionName),
		}
	}
	m.ExcludedCollections = make([]endpointMongoCollection, len(e.ExcludedCollections))
	for i := 0; i < len(e.ExcludedCollections); i++ {
		m.ExcludedCollections[i] = endpointMongoCollection{
			DatabaseName:   types.StringValue(e.ExcludedCollections[i].DatabaseName),
			CollectionName: types.StringValue(e.ExcludedCollections[i].CollectionName),
		}
	}

	return diag
}

func (m *endpointMongoTargetSettings) parse(e *endpoint.MongoTarget) diag.Diagnostics {
	var diag diag.Diagnostics

	m.Connection.parse(e.Connection)
	m.Database = types.StringValue(e.Database)
	m.CleanupPolicy = types.StringValue(e.CleanupPolicy.String())
	return diag
}
