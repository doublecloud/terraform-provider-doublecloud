package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointMysqlSourceSettings struct {
	Connection             *endpointMysqlConnection             `tfsdk:"connection"`
	Database               types.String                         `tfsdk:"database"`
	ServiceDatabase        types.String                         `tfsdk:"service_database"`
	User                   types.String                         `tfsdk:"user"`
	Password               types.String                         `tfsdk:"password"`
	IncludeTablesRegex     []types.String                       `tfsdk:"include_tables_regex"`
	ExcludeTablesRegex     []types.String                       `tfsdk:"exclude_tables_regex"`
	Timezone               types.String                         `tfsdk:"timezone"`
	ObjectTransferSettings *endpointMysqlObjectTransferSettings `tfsdk:"object_transfer_settings"`
}

type endpointMysqlConnection struct {
	OnPremise *endpointMysqlOnPremise `tfsdk:"on_premise"`
}

type endpointMysqlOnPremise struct {
	Hosts   []types.String   `tfsdk:"hosts"`
	Port    types.Int64      `tfsdk:"port"`
	TLSMode *endpointTLSMode `tfsdk:"tls_mode"`
}

type endpointMysqlObjectTransferSettings struct {
	View    types.String `tfsdk:"view"`
	Routine types.String `tfsdk:"routine"`
	Trigger types.String `tfsdk:"trigger"`
	Tables  types.String `tfsdk:"tables"`
}

type endpointMysqlTargetSettings struct {
	Connection *endpointMysqlConnection `tfsdk:"connection"`
	// SecurityGroups      []types.String           `tfsdk:"security_groups"`
	Database            types.String `tfsdk:"database"`
	User                types.String `tfsdk:"user"`
	Password            types.String `tfsdk:"password"`
	SqlMode             types.String `tfsdk:"sql_mode"`
	SkipConstraintCheck types.Bool   `tfsdk:"skip_constraint_checks"`
	Timezone            types.String `tfsdk:"timezone"`
	CleanupPolicy       types.String `tfsdk:"cleanup_policy"`
	ServiceDatabase     types.String `tfsdk:"service_database"`
}

func transferEndpointMysqlSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "Database name",
				Optional:            true,
			},
			"service_database": schema.StringAttribute{
				MarkdownDescription: "Service database name",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"include_tables_regex": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"exclude_tables_regex": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "Used for parsing timestamps for saving source timezones. Accepts values from the IANA timezone database. Default is the local timezone.",
				Optional:            true,
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"connection":               transferEndpointMysqlConnectionSchema(),
			"object_transfer_settings": transferEndpointMysqlObjectTransferSchemaBlock(),
		},
	}
}

func transferEndpointMysqlConnectionSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"on_premise": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"hosts": schema.ListAttribute{
						MarkdownDescription: "List of MySQL hosts",
						ElementType:         types.StringType,
						Optional:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "MySQL port",
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

func transferEndpointMysqlObjectTransferSchemaBlock() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"view": schema.StringAttribute{
				MarkdownDescription: "CREATE VIEW ...",
				Computed:            true,
				Optional:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"routine": schema.StringAttribute{
				MarkdownDescription: "CREATE PROCEDURE ... ; CREATE FUNCTION ... ;",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "CREATE TRIGGER ...",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tables": schema.StringAttribute{
				MarkdownDescription: "CREATE TABLE ...",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{transferObjectTransferStageValidator()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func transferEndpointMysqlTargetSchema() schema.Block {
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
			"sql_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "SQL mode",
			},
			"skip_constraint_checks": schema.BoolAttribute{
				MarkdownDescription: "Disable constraint checks",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "Used for parsing timestamps for saving source timezones. Accepts values from the IANA timezone database. Default is the local timezone.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cleanup_policy": schema.StringAttribute{
				MarkdownDescription: "Cleanup policy for activating, reactivating, and reuploading processes. Default is `truncate`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{transferEndpointCleanupPolicyValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_database": schema.StringAttribute{
				MarkdownDescription: "Database schema for the service table",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"connection": transferEndpointPostgresConnectionSchema(),
		},
	}
}

func (m *endpointMysqlSourceSettings) convert() (*transfer.EndpointSettings_MysqlSource, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_MysqlSource{MysqlSource: &endpoint.MysqlSource{}}
	var diags diag.Diagnostics

	connection, d := m.Connection.convert()
	diags.Append(d...)
	settings.MysqlSource.Connection = connection
	settings.MysqlSource.Database = m.Database.ValueString()
	if !m.ServiceDatabase.IsNull() {
		settings.MysqlSource.ServiceDatabase = m.ServiceDatabase.ValueString()
	}
	settings.MysqlSource.User = m.User.ValueString()
	settings.MysqlSource.Password = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Password.ValueString()}}
	if m.IncludeTablesRegex != nil {
		settings.MysqlSource.IncludeTablesRegex = convertSliceTFStrings(m.IncludeTablesRegex)
	}
	if m.ExcludeTablesRegex != nil {
		settings.MysqlSource.ExcludeTablesRegex = convertSliceTFStrings(m.ExcludeTablesRegex)
	}
	if !m.Timezone.IsNull() {
		settings.MysqlSource.Timezone = m.Timezone.ValueString()
	}
	settings.MysqlSource.ObjectTransferSettings = &endpoint.MysqlObjectTransferSettings{}

	if m.ObjectTransferSettings != nil {
		settings.MysqlSource.ObjectTransferSettings = m.ObjectTransferSettings.convert()
	}

	return settings, diags
}

func (m *endpointMysqlConnection) convert() (*endpoint.MysqlConnection, diag.Diagnostics) {
	var diag diag.Diagnostics

	options := &endpoint.MysqlConnection{}

	if on_premise := m.OnPremise; on_premise != nil {
		tlsMode := convertTLSMode(m.OnPremise.TLSMode)

		options.Connection = &endpoint.MysqlConnection_OnPremise{OnPremise: &endpoint.OnPremiseMysql{
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

func (m *endpointMysqlObjectTransferSettings) convert() *endpoint.MysqlObjectTransferSettings {
	stage := &endpoint.MysqlObjectTransferSettings{}

	if m == nil {
		return stage
	}

	stage.View = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.View.ValueString()])
	stage.Routine = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Routine.ValueString()])
	stage.Trigger = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Trigger.ValueString()])
	stage.Tables = endpoint.ObjectTransferStage(endpoint.ObjectTransferStage_value[m.Tables.ValueString()])

	return stage
}

func (m *endpointMysqlTargetSettings) convert() (*transfer.EndpointSettings_MysqlTarget, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_MysqlTarget{MysqlTarget: &endpoint.MysqlTarget{}}
	var diags diag.Diagnostics

	connection, d := m.Connection.convert()
	diags.Append(d...)

	settings.MysqlTarget.Connection = connection
	settings.MysqlTarget.Database = m.Database.ValueString()
	settings.MysqlTarget.User = m.User.ValueString()
	settings.MysqlTarget.Password = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Password.ValueString()}}

	// if m.SecurityGroups != nil {
	// 	settings.MysqlTarget.SecurityGroups = convertSliceTFStrings(m.SecurityGroups)
	// }
	if !m.SqlMode.IsNull() {
		settings.MysqlTarget.SqlMode = m.SqlMode.ValueString()
	}
	if !m.SkipConstraintCheck.IsNull() {
		settings.MysqlTarget.SkipConstraintChecks = m.SkipConstraintCheck.ValueBool()
	}
	if !m.Timezone.IsNull() {
		settings.MysqlTarget.Timezone = m.Timezone.ValueString()
	}
	if !m.CleanupPolicy.IsNull() {
		settings.MysqlTarget.CleanupPolicy = endpoint.CleanupPolicy(endpoint.CleanupPolicy_value[m.CleanupPolicy.ValueString()])
	}
	if !m.ServiceDatabase.IsNull() {
		settings.MysqlTarget.ServiceDatabase = m.ServiceDatabase.ValueString()
	}

	return settings, diags
}

func (m *endpointMysqlConnection) parse(e *endpoint.MysqlConnection) {
	if e == nil {
		m = nil
	}

	if on_premise := e.GetOnPremise(); on_premise != nil {
		if m == nil {
			m = &endpointMysqlConnection{}
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

func (m *endpointMysqlObjectTransferSettings) parse(e *endpoint.MysqlObjectTransferSettings) {
	if e == nil {
		m = nil
	}
	m.View = types.StringValue(e.GetView().String())
	m.Routine = types.StringValue(e.GetRoutine().String())
	m.Trigger = types.StringValue(e.GetView().String())
	m.Tables = types.StringValue(e.GetTables().String())
}

func (m *endpointMysqlSourceSettings) parse(e *endpoint.MysqlSource) diag.Diagnostics {
	var diag diag.Diagnostics

	m.Connection.parse(e.Connection)
	m.Database = types.StringValue(e.Database)
	m.ServiceDatabase = types.StringValue(e.ServiceDatabase)
	m.User = types.StringValue(e.User)
	m.IncludeTablesRegex = convertSliceToTFStrings(e.IncludeTablesRegex)
	m.ExcludeTablesRegex = convertSliceToTFStrings(e.ExcludeTablesRegex)
	m.Timezone = types.StringValue(e.Timezone)

	if e.ObjectTransferSettings != nil {
		if m.ObjectTransferSettings == nil {
			m.ObjectTransferSettings = &endpointMysqlObjectTransferSettings{}
		}
		m.ObjectTransferSettings.parse(e.ObjectTransferSettings)
	} else {
		m.ObjectTransferSettings = nil
	}

	return diag
}

func (m *endpointMysqlTargetSettings) parse(e *endpoint.MysqlTarget) diag.Diagnostics {
	var diag diag.Diagnostics

	m.Connection.parse(e.Connection)
	// m.SecurityGroups = convertSliceToTFStrings(e.SecurityGroups)

	m.Database = types.StringValue(e.Database)
	m.User = types.StringValue(e.User)
	m.SqlMode = types.StringValue(e.SqlMode)
	m.SkipConstraintCheck = types.BoolValue(e.SkipConstraintChecks)
	m.Timezone = types.StringValue(e.Timezone)
	m.CleanupPolicy = types.StringValue(e.CleanupPolicy.String())
	m.ServiceDatabase = types.StringValue(e.ServiceDatabase)

	return diag
}
