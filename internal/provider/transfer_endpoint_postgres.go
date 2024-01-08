package provider

import (
	"context"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointPostgresSourceSettings struct {
	Connection             *endpointPostgresConnection             `tfsdk:"connection"`
	Database               types.String                            `tfsdk:"database"`
	User                   types.String                            `tfsdk:"user"`
	Password               types.String                            `tfsdk:"password"`
	IncludeTables          []types.String                          `tfsdk:"include_tables"`
	ExcludeTables          []types.String                          `tfsdk:"exclude_tables"`
	SlotByteLagLimit       types.Int64                             `tfsdk:"slot_byte_lag_limit"`
	ServiceSchema          types.String                            `tfsdk:"service_schema"`
	ObjectTransferSettings *endpointPostgresObjectTransferSettings `tfsdk:"object_transfer_settings"`
}

type endpointPostgresTargetSettings struct {
	Connection *endpointPostgresConnection `tfsdk:"connection"`
	// SecurityGroups []types.String              `tfsdk:"security_groups"`
	Database      types.String `tfsdk:"database"`
	User          types.String `tfsdk:"user"`
	Password      types.String `tfsdk:"password"`
	CleanupPolicy types.String `tfsdk:"cleanup_policy"`
}

type endpointPostgresConnection struct {
	OnPremise *endpointPostgresConnectionOnPremise `tfsdk:"on_premise"`
}

type endpointPostgresConnectionOnPremise struct {
	Hosts   []types.String   `tfsdk:"hosts"`
	Port    types.Int64      `tfsdk:"port"`
	TLSMode *endpointTLSMode `tfsdk:"tls_mode"`
}

type endpointPostgresObjectTransferSettings struct {
	Sequence         types.String `tfsdk:"sequence"`
	SequenceOwnedBy  types.String `tfsdk:"sequence_owned_by"`
	SequenceSet      types.String `tfsdk:"sequence_set"`
	Table            types.String `tfsdk:"table"`
	PrimaryKey       types.String `tfsdk:"primary_key"`
	FkConstraint     types.String `tfsdk:"fk_constraint"`
	DefaultValues    types.String `tfsdk:"default_values"`
	Constraint       types.String `tfsdk:"constraint"`
	Index            types.String `tfsdk:"index"`
	View             types.String `tfsdk:"view"`
	MaterializedView types.String `tfsdk:"materialized_view"`
	Function         types.String `tfsdk:"function"`
	Trigger          types.String `tfsdk:"trigger"`
	Type             types.String `tfsdk:"type"`
	Rule             types.String `tfsdk:"rule"`
	Collation        types.String `tfsdk:"collation"`
	Policy           types.String `tfsdk:"policy"`
	Cast             types.String `tfsdk:"cast"`
}

func transferEndpointPostgresSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Optional:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "Database user",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Database user password",
				Optional:            true,
				Sensitive:           true,
			},
			"include_tables": schema.ListAttribute{
				MarkdownDescription: "List of tables to be replicated. Table names must be full and contain schemas. Can contain `schema_name.*` patterns. If the setting isn't specified or contains an empty list, all tables are replicated",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"exclude_tables": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of tables to be excluded from replication",
			},
			"slot_byte_lag_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum lag of replication slots (in bytes). When this limit is exceeded,replication is aborted",
				Optional:            true,
				Computed:            true,
			},
			"service_schema": schema.StringAttribute{
				MarkdownDescription: "Database schema for service tables (`__consumer_keeper` and `__data_transfer_mole_finder`). Default is `public`",
				Optional:            true,
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"connection":               transferEndpointPostgresConnectionSchema(),
			"object_transfer_settings": transferEndpointPostgresObjectTransferSchemaBlock(),
		},
	}
}

func transferEndpointPostgresConnectionSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"on_premise": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"hosts": schema.ListAttribute{
						MarkdownDescription: "List of PostgreSQL hosts",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "Port of the PostgreSQL instance",
						Optional:            true,
					},
				},
				Blocks: map[string]schema.Block{
					"tls_mode": transferEndpointTLSMode(),
				},
			},
		},
	}
}

func transferEndpointPostgresObjectTransferSchemaBlock() schema.Block {
	return schema.SingleNestedBlock{
		PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
		Attributes: map[string]schema.Attribute{
			"sequence": schema.StringAttribute{
				MarkdownDescription: "CREATE SEQUENCE ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"sequence_owned_by": schema.StringAttribute{
				MarkdownDescription: "CREATE SEQUENCE ... OWNED BY ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"sequence_set": schema.StringAttribute{
				Optional:   true,
				Validators: []validator.String{transferObjectTransferStageValidator()},
			},
			"table": schema.StringAttribute{
				MarkdownDescription: "CREATE TABLE ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"primary_key": schema.StringAttribute{
				MarkdownDescription: "ALTER TABLE ... ADD PRIMARY KEY ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"fk_constraint": schema.StringAttribute{
				MarkdownDescription: "ALTER TABLE ... ADD FOREIGN KEY ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"default_values": schema.StringAttribute{
				MarkdownDescription: "ALTER TABLE ... ALTER COLUMN ... SET DEFAULT ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"constraint": schema.StringAttribute{
				MarkdownDescription: "ALTER TABLE ... ADD CONSTRAINT ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"index": schema.StringAttribute{
				MarkdownDescription: "CREATE INDEX ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"view": schema.StringAttribute{
				MarkdownDescription: "CREATE VIEW ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"materialized_view": schema.StringAttribute{
				MarkdownDescription: "CREATE MATERIALIZED VIEW ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"function": schema.StringAttribute{
				MarkdownDescription: "CREATE FUNCTION ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "CREATE TRIGGER ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "CREATE TYPE ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"rule": schema.StringAttribute{
				MarkdownDescription: "CREATE RULE ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"collation": schema.StringAttribute{
				MarkdownDescription: "CREATE COLLATION ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"policy": schema.StringAttribute{
				MarkdownDescription: "CREATE POLICY ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
			"cast": schema.StringAttribute{
				MarkdownDescription: "CREATE CAST ...",
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
			},
		},
	}
}

func transferEndpointPostgresTargetSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Optional:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "Database user",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Database user password",
				Optional:            true,
				Sensitive:           true,
			},
			// "security_groups": schema.ListAttribute{
			// 	MarkdownDescription: "Security groups",
			// 	ElementType:         types.StringType,
			// 	Optional:            true,
			// },
			"cleanup_policy": schema.StringAttribute{
				MarkdownDescription: "Cleanup policy for activating, reactivating, and reuploading processes. Default is `truncate`",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{transferEndpointCleanupPolicyValidator()},
			},
		},
		Blocks: map[string]schema.Block{
			"connection": transferEndpointPostgresConnectionSchema(),
		},
	}
}

func postgresSourceEndpointSettings(m *endpointPostgresSourceSettings) (*transfer.EndpointSettings_PostgresSource, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_PostgresSource{PostgresSource: &endpoint.PostgresSource{}}
	var diags diag.Diagnostics

	connection, d := convertPostgresConnection(m.Connection)
	diags.Append(d...)
	settings.PostgresSource.Connection = connection
	settings.PostgresSource.Database = m.Database.ValueString()
	settings.PostgresSource.User = m.User.ValueString()
	settings.PostgresSource.Password = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Password.ValueString()}}
	settings.PostgresSource.ObjectTransferSettings = &endpoint.PostgresObjectTransferSettings{}

	if m.IncludeTables != nil {
		settings.PostgresSource.IncludeTables = convertSliceTFStrings(m.IncludeTables)
	}
	if m.ExcludeTables != nil {
		settings.PostgresSource.ExcludeTables = convertSliceTFStrings(m.ExcludeTables)
	}
	if !m.SlotByteLagLimit.IsNull() {
		settings.PostgresSource.SlotByteLagLimit = m.SlotByteLagLimit.ValueInt64()
	}
	if !m.ServiceSchema.IsNull() {
		settings.PostgresSource.ServiceSchema = m.ServiceSchema.ValueString()
	}

	if m.ObjectTransferSettings != nil {
		settings.PostgresSource.ObjectTransferSettings = convertObjectTransferSettings(m.ObjectTransferSettings)
	}

	return settings, diags
}

func convertPostgresConnection(m *endpointPostgresConnection) (*endpoint.PostgresConnection, diag.Diagnostics) {
	var diag diag.Diagnostics

	options := &endpoint.PostgresConnection{}

	if on_premise := m.OnPremise; on_premise != nil {
		tlsMode := convertTLSMode(m.OnPremise.TLSMode)

		options.Connection = &endpoint.PostgresConnection_OnPremise{OnPremise: &endpoint.OnPremisePostgres{
			Hosts:   convertSliceTFStrings(m.OnPremise.Hosts),
			Port:    m.OnPremise.Port.ValueInt64(),
			TlsMode: tlsMode,
		}}
	}

	if options.Connection == nil {
		diag.AddError("unknown connection", "required on_premise block")
	}
	return options, diag
}

func convertObjectTransferSettings(m *endpointPostgresObjectTransferSettings) *endpoint.PostgresObjectTransferSettings {
	stage := &endpoint.PostgresObjectTransferSettings{}

	if m == nil {
		return stage
	}

	if !m.Sequence.IsNull() {
		stage.Sequence = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Sequence.ValueString()])
	}
	if !m.SequenceOwnedBy.IsNull() {
		stage.SequenceOwnedBy = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.SequenceOwnedBy.ValueString()])
	}
	if !m.SequenceSet.IsNull() {
		stage.SequenceSet = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.SequenceSet.ValueString()])
	}
	if !m.Table.IsNull() {
		stage.Table = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Table.ValueString()])
	}
	if !m.PrimaryKey.IsNull() {
		stage.PrimaryKey = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.PrimaryKey.ValueString()])
	}
	if !m.FkConstraint.IsNull() {
		stage.FkConstraint = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.FkConstraint.ValueString()])
	}
	if !m.DefaultValues.IsNull() {
		stage.DefaultValues = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.DefaultValues.ValueString()])
	}
	if !m.DefaultValues.IsNull() {
		stage.DefaultValues = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.DefaultValues.ValueString()])
	}
	if !m.Constraint.IsNull() {
		stage.Constraint = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Constraint.ValueString()])
	}
	if !m.Index.IsNull() {
		stage.Index = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Index.ValueString()])
	}
	if !m.View.IsNull() {
		stage.View = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.View.ValueString()])
	}
	if !m.MaterializedView.IsNull() {
		stage.MaterializedView = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.MaterializedView.ValueString()])
	}
	if !m.Function.IsNull() {
		stage.Function = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Function.ValueString()])
	}
	if !m.Trigger.IsNull() {
		stage.Trigger = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Trigger.ValueString()])
	}
	if !m.Type.IsNull() {
		stage.Type = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Type.ValueString()])
	}
	if !m.Rule.IsNull() {
		stage.Rule = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Rule.ValueString()])
	}
	if !m.Collation.IsNull() {
		stage.Collation = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Collation.ValueString()])
	}
	if !m.Policy.IsNull() {
		stage.Policy = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Policy.ValueString()])
	}
	if !m.Cast.IsNull() {
		stage.Cast = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Cast.ValueString()])
	}
	return stage
}

func postgresTargetEndpointSettings(m *endpointPostgresTargetSettings) (*transfer.EndpointSettings_PostgresTarget, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_PostgresTarget{PostgresTarget: &endpoint.PostgresTarget{}}
	var diags diag.Diagnostics

	connection, d := convertPostgresConnection(m.Connection)
	diags.Append(d...)
	settings.PostgresTarget.Connection = connection
	settings.PostgresTarget.Database = m.Database.ValueString()
	settings.PostgresTarget.User = m.User.ValueString()
	settings.PostgresTarget.Password = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Password.ValueString()}}

	// if m.SecurityGroups != nil {
	// 	settings.PostgresTarget.SecurityGroups = convertSliceTFStrings(m.SecurityGroups)
	// }

	if !m.CleanupPolicy.IsNull() {
		settings.PostgresTarget.CleanupPolicy = endpoint.CleanupPolicy(endpoint.CleanupPolicy_value[m.CleanupPolicy.ValueString()])
	}

	return settings, diags
}

func parseTransferEndpointPostgresSource(ctx context.Context, e *endpoint.PostgresSource, c *endpointPostgresSourceSettings) diag.Diagnostics {
	var diag diag.Diagnostics

	parseTransferEndpointPostgresConnection(e.Connection, c.Connection)
	c.Database = types.StringValue(e.Database)
	c.User = types.StringValue(e.User)

	c.IncludeTables = convertSliceToTFStrings(e.IncludeTables)
	c.ExcludeTables = convertSliceToTFStrings(e.ExcludeTables)
	c.SlotByteLagLimit = types.Int64Value(e.SlotByteLagLimit)
	c.ServiceSchema = types.StringValue(e.ServiceSchema)

	// TODO: Fix bug with default empty block
	// parse ObjectTransferSettings
	return diag
}

func parseTransferEndpointPostgresTarget(ctx context.Context, e *endpoint.PostgresTarget, c *endpointPostgresTargetSettings) diag.Diagnostics {
	var diag diag.Diagnostics

	parseTransferEndpointPostgresConnection(e.Connection, c.Connection)
	// c.SecurityGroups = convertSliceToTFStrings(e.SecurityGroups)
	c.Database = types.StringValue(e.Database)
	c.User = types.StringValue(e.User)
	c.CleanupPolicy = types.StringValue(e.CleanupPolicy.String())

	return diag
}

func parseTransferEndpointPostgresConnection(e *endpoint.PostgresConnection, m *endpointPostgresConnection) {
	if e == nil {
		m = nil
	}

	if on_premise := e.GetOnPremise(); on_premise != nil {
		if m == nil {
			m = &endpointPostgresConnection{}
		}
		if m.OnPremise.Hosts != nil {
			m.OnPremise.Hosts = convertSliceToTFStrings(on_premise.Hosts)
		}
		if !m.OnPremise.Port.IsNull() {
			m.OnPremise.Port = types.Int64Value(on_premise.Port)
		}
		if m.OnPremise.TLSMode != nil {
			if disabled := on_premise.TlsMode.GetDisabled(); disabled != nil {
				m.OnPremise.TLSMode = nil
			}
			if config := on_premise.TlsMode.GetEnabled(); config != nil {
				m.OnPremise.TLSMode = &endpointTLSMode{CACertificate: types.StringValue(config.CaCertificate)}
			}
		}
	}
}
