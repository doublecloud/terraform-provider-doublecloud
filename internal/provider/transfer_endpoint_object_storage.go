package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
)

type endpointObjectStorageSourceSettings struct {
	Provider     *endpointObjectStorageProvider     `tfsdk:"provider"`
	Format       *endpointObjectStorageFormat       `tfsdk:"format"`
	PathPattern  types.String                       `tfsdk:"path_pattern"`
	ResultTable  *endpointObjectStorageResultTable  `tfsdk:"result_table"`
	ResultSchema *endpointObjectStorageResultSchema `tfsdk:"result_schema"`
	EventSource  *endpointObjectStorageEventSource  `tfsdk:"event_source"`
}

type endpointObjectStorageFormat struct {
	Csv     *endpointObjectStorageFormatCSV     `tfsdk:"csv"`
	Parquet *endpointObjectStorageFormatParquet `tfsdk:"parquet"`
	Avro    *endpointObjectStorageFormatAvro    `tfsdk:"avro"`
	Jsonl   *endpointObjectStorageFormatJsonl   `tfsdk:"jsonl"`
}

type endpointObjectStorageFormatCSV struct {
	Delimiter               types.String                                           `tfsdk:"delimiter"`
	QuoteChar               types.String                                           `tfsdk:"quote_char"`
	EscapeChar              types.String                                           `tfsdk:"escape_char"`
	Encoding                types.String                                           `tfsdk:"encoding"`
	DoubleQuote             types.Bool                                             `tfsdk:"double_quote"`
	NewlinesInValues        types.Bool                                             `tfsdk:"newlines_in_values"`
	BlockSize               types.Int64                                            `tfsdk:"block_size"`
	AdvancedOptions         *endpointObjectStorageFormatCSVAdvancedOptions         `tfsdk:"advanced_options"`
	AdditionalReaderOptions *endpointObjectStorageFormatCSVAdditionalReaderOptions `tfsdk:"additional_options"`
}
type endpointObjectStorageFormatJsonl struct {
	NewlinesInValues        types.Bool   `tfsdk:"newlines_in_values"`
	UnexpectedFieldBehavior types.String `tfsdk:"unexpected_field_behavior"`
	BlockSize               types.Int64  `tfsdk:"block_size"`
}

type endpointObjectStorageFormatCSVAdvancedOptions struct {
	SkipRows                types.Int64    `tfsdk:"skip_rows"`
	SkipRowsAfterNames      types.Int64    `tfsdk:"skip_rows_after_names"`
	AutogenerateColumnNames types.Bool     `tfsdk:"autogenerate_column_names"`
	ColumnNames             []types.String `tfsdk:"column_names"`
}

type endpointObjectStorageFormatCSVAdditionalReaderOptions struct {
	NullValues             []types.String `tfsdk:"null_values"`
	TrueValues             []types.String `tfsdk:"true_values"`
	FalseValues            []types.String `tfsdk:"false_values"`
	DecimalPoint           types.String   `tfsdk:"decimal_point"`
	StringsCanBeNull       types.Bool     `tfsdk:"strings_can_be_null"`
	QuotedStringsCanBeNull types.Bool     `tfsdk:"quoted_strings_can_be_null"`
	IncludeColumns         []types.String `tfsdk:"include_columns"`
	IncludeMissingColumns  types.Bool     `tfsdk:"include_missing_columns"`
	TimestampParsers       []types.String `tfsdk:"timestamp_parsers"`
}

type (
	endpointObjectStorageFormatAvro        struct{}
	endpointObjectStorageFormatParquet     struct{}
	endpointObjectStorageResultSchemaInfer struct{}
	endpointObjectStorageEventSourceSNS    struct{}
	endpointObjectStorageEventSourcePubSub struct{}
)

type endpointObjectStorageProvider struct {
	Bucket             types.String `tfsdk:"bucket"`
	AwsAccessKeyId     types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	PathPrefix         types.String `tfsdk:"path_prefix"`
	Endpoint           types.String `tfsdk:"endpoint"`
	Region             types.String `tfsdk:"region"`
	UseSSL             types.Bool   `tfsdk:"use_ssl"`
	VerifySSLCert      types.Bool   `tfsdk:"verify_ssl_cert"`
}

type endpointObjectStorageResultTable struct {
	TableNamespace types.String `tfsdk:"table_namespace"`
	TableName      types.String `tfsdk:"table_name"`
	AddSystemCols  types.Bool   `tfsdk:"add_system_cols"`
}

type endpointObjectStorageResultSchema struct {
	Infer      *endpointObjectStorageResultSchemaInfer `tfsdk:"infer"`
	DataSchema *endpointObjectStorageDataSchema        `tfsdk:"data_schema"`
}

type endpointObjectStorageDataSchema struct {
	JsonFields *endpointObjetStorageDataSchemaJsonFields `tfsdk:"json_fields"`
	Fields     *transferParserSchemaFields               `tfsdk:"fields"`
}
type endpointObjetStorageDataSchemaJsonFields struct {
	JsonFields types.String `tfsdk:"json_fields"`
}

type endpointObjectStorageEventSource struct {
	SQS    *endpointObjectStorageEventSourceSQS    `tfsdk:"sqs"`
	SNS    *endpointObjectStorageEventSourceSNS    `tfsdk:"sns"`
	PubSub *endpointObjectStorageEventSourcePubSub `tfsdk:"pub_sub"`
}

type endpointObjectStorageEventSourceSQS struct {
	QueueName          types.String `tfsdk:"queue_name"`
	OwnerID            types.String `tfsdk:"owner_id"`
	AwsAccessKeyId     types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	Endpoint           types.String `tfsdk:"endpoint"`
	Region             types.String `tfsdk:"region"`
	UseSSL             types.Bool   `tfsdk:"use_ssl"`
	VerifySSLCert      types.Bool   `tfsdk:"verify_ssl_cert"`
}

type endpointObjectStorageTargetSettings struct {
	Bucket               types.String                           `tfsdk:"bucket"`
	ServiceAccountID     types.String                           `tfsdk:"service_account_id"`
	OutputFormat         types.String                           `tfsdk:"output_format"`
	BucketLayout         types.String                           `tfsdk:"bucket_layout"`
	BucketLayoutTimezone types.String                           `tfsdk:"bucket_layout_timezone"`
	BucketLayoutColumn   types.String                           `tfsdk:"bucket_layout_column"`
	BufferSize           types.String                           `tfsdk:"buffer_size"`
	BufferInterval       types.String                           `tfsdk:"buffer_interval"`
	OutputEncoding       types.String                           `tfsdk:"output_encoding"`
	Connection           *endpointObjectStorageConnection       `tfsdk:"connection"`
	SerializerConfig     *endpointObjectStorageSerializerConfig `tfsdk:"serializer_config"`
}

type endpointObjectStorageSerializerConfig struct {
	AnyAsString types.Bool `tfsdk:"any_as_string"`
}

type endpointObjectStorageConnection struct {
	AwsAccessKeyId     types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	Region             types.String `tfsdk:"region"`
	Endpoint           types.String `tfsdk:"endpoint"`
	UseSSL             types.Bool   `tfsdk:"use_ssl"`
	VerifySSLCert      types.Bool   `tfsdk:"verify_ssl_cert"`
}

func endpointObjetStorageDataSchemaJsonFieldsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"json_fields": schema.StringAttribute{Optional: true},
		},
	}
}

func endpointObjectStorageResultSchemaSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"infer": schema.SingleNestedBlock{
				MarkdownDescription: "Automatically infer schema",
			},
			"data_schema": endpointObjectStorageDataSchemaSchema(),
		},
	}
}

func endpointObjectStorageDataSchemaSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"json_fields": endpointObjetStorageDataSchemaJsonFieldsSchema(),
			"fields":      transferParserSchemaFieldsSchema(),
		},
	}
}

func transferEndpointObjectStorageUnexpectedFieldBehaviorValidator() validator.String {
	names := make([]string, len(endpoint.ObjectStorageReaderFormat_Jsonl_UnexpectedFieldBehavior_name))
	for i, v := range endpoint.ObjectStorageReaderFormat_Jsonl_UnexpectedFieldBehavior_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func transferEndpointObjectStorageSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"path_pattern": schema.StringAttribute{Optional: true},
		},
		Blocks: map[string]schema.Block{
			"format":        endpointObjectStorageSourceFormatSchema(),
			"event_source":  endpointObjectStorageSourceEventSourceSchema(),
			"result_schema": endpointObjectStorageResultSchemaSchema(),
			"provider": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"bucket":                schema.StringAttribute{Optional: true},
					"aws_access_key_id":     schema.StringAttribute{Optional: true, Sensitive: true},
					"aws_secret_access_key": schema.StringAttribute{Optional: true, Sensitive: true},
					"path_prefix":           schema.StringAttribute{Optional: true},
					"endpoint":              schema.StringAttribute{Optional: true},
					"region":                schema.StringAttribute{Optional: true},
					"use_ssl":               schema.BoolAttribute{Optional: true},
					"verify_ssl_cert":       schema.BoolAttribute{Optional: true},
				},
			},
			"result_table": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"table_namespace": schema.StringAttribute{Optional: true},
					"table_name":      schema.StringAttribute{Optional: true},
					"add_system_cols": schema.BoolAttribute{Optional: true},
				},
			},
		},
	}
}

func endpointObjectStorageSourceFormatSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"csv":     endpointObjectStorageSourceFormatCsvSchema(),
			"parquet": schema.SingleNestedBlock{},
			"avro":    schema.SingleNestedBlock{},
			"jsonl": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"newlines_in_values": schema.BoolAttribute{Optional: true},
					"unexpected_field_behavior": schema.StringAttribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						Validators:    []validator.String{transferEndpointObjectStorageUnexpectedFieldBehaviorValidator()},
					},
					"block_size": schema.Int64Attribute{Optional: true},
				},
			},
		},
	}
}

func endpointObjectStorageSourceFormatCsvSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"delimiter":          schema.StringAttribute{Optional: true},
			"quote_char":         schema.StringAttribute{Optional: true},
			"escape_char":        schema.StringAttribute{Optional: true},
			"encoding":           schema.StringAttribute{Optional: true},
			"double_quote":       schema.BoolAttribute{Optional: true},
			"newlines_in_values": schema.BoolAttribute{Optional: true},
			"block_size":         schema.Int64Attribute{Optional: true},
		},
		Blocks: map[string]schema.Block{
			"additional_options": endpointObjectStorageSourceFormatCsvAdditionalOptionsSchema(),
			"advanced_options":   endpointObjectStorageSourceFormatCsvAdvancedOptionsSchema(),
		},
	}
}

func endpointObjectStorageSourceFormatCsvAdditionalOptionsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"null_values":                schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"true_values":                schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"false_values":               schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"decimal_point":              schema.StringAttribute{Optional: true},
			"strings_can_be_null":        schema.BoolAttribute{Optional: true},
			"quoted_strings_can_be_null": schema.BoolAttribute{Optional: true},
			"include_columns":            schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"include_missing_columns":    schema.BoolAttribute{Optional: true},
			"timestamp_parsers":          schema.ListAttribute{ElementType: types.StringType, Optional: true},
		},
	}
}

func endpointObjectStorageSourceFormatCsvAdvancedOptionsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"skip_rows":                 schema.Int64Attribute{Optional: true},
			"skip_rows_after_names":     schema.Int64Attribute{Optional: true},
			"autogenerate_column_names": schema.BoolAttribute{Optional: true},
			"column_names":              schema.ListAttribute{ElementType: types.StringType, Optional: true},
		},
	}
}

func endpointObjectStorageSourceEventSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"sqs": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"queue_name":            schema.StringAttribute{Optional: true},
					"owner_id":              schema.StringAttribute{Optional: true},
					"aws_access_key_id":     schema.StringAttribute{Optional: true, Sensitive: true},
					"aws_secret_access_key": schema.StringAttribute{Optional: true, Sensitive: true},
					"endpoint":              schema.StringAttribute{Optional: true},
					"region":                schema.StringAttribute{Optional: true},
					"use_ssl":               schema.BoolAttribute{Optional: true},
					"verify_ssl_cert":       schema.BoolAttribute{Optional: true},
				},
			},
			"sns":     schema.SingleNestedBlock{},
			"pub_sub": schema.SingleNestedBlock{},
		},
	}
}

func transferEndpointObjectStorageTargetSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"bucket":             schema.StringAttribute{Optional: true},
			"service_account_id": schema.StringAttribute{Optional: true},
			"output_format": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators:    []validator.String{transferEndpointObjectStorageOutputFormatValidator()},
			},
			"bucket_layout":          schema.StringAttribute{Optional: true, Computed: true},
			"bucket_layout_timezone": schema.StringAttribute{Optional: true, Computed: true},
			"bucket_layout_column":   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"buffer_size":            schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"buffer_interval":        schema.StringAttribute{Optional: true},
			"output_encoding": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators:    []validator.String{transferEndpointObjectStorageOutputEncodingValidator()},
			},
		},
		Blocks: map[string]schema.Block{
			"connection":        endpointObjectStorageTargetConnectionSchema(),
			"serializer_config": endpointObjectStorageTargetSerializerConfigSchema(),
		},
	}
}

func endpointObjectStorageTargetConnectionSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"aws_access_key_id":     schema.StringAttribute{Optional: true},
			"aws_secret_access_key": schema.StringAttribute{Optional: true},
			"endpoint":              schema.StringAttribute{Optional: true},
			"region":                schema.StringAttribute{Optional: true},
			"use_ssl":               schema.BoolAttribute{Optional: true},
			"verify_ssl_cert":       schema.BoolAttribute{Optional: true},
		},
	}
}

func endpointObjectStorageTargetSerializerConfigSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"any_as_string": schema.BoolAttribute{Optional: true},
		},
	}
}

func transferEndpointObjectStorageOutputFormatValidator() validator.String {
	names := make([]string, len(endpoint.ObjectStorageSerializationFormat_name))
	for i, v := range endpoint.ObjectStorageSerializationFormat_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func transferEndpointObjectStorageOutputEncodingValidator() validator.String {
	names := make([]string, len(endpoint.ObjectStorageCodec_name))
	for i, v := range endpoint.ObjectStorageCodec_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func (m *endpointObjectStorageSourceSettings) convert() (*transfer.EndpointSettings_ObjectStorageSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := &transfer.EndpointSettings_ObjectStorageSource{ObjectStorageSource: &endpoint.ObjectStorageSource{}}

	if v := m.PathPattern; !v.IsNull() {
		settings.ObjectStorageSource.PathPattern = v.ValueString()
	}

	if v := m.Provider; v != nil {
		provider, d := v.convert()
		settings.ObjectStorageSource.Provider = provider
		diags.Append(d...)
	}
	if v := m.Format; v != nil {
		format, d := v.convert()
		settings.ObjectStorageSource.Format = format
		diags.Append(d...)
	}
	if v := m.ResultTable; v != nil {
		table, d := v.convert()
		settings.ObjectStorageSource.ResultTable = table
		diags.Append(d...)
	}
	if v := m.ResultSchema; v != nil {
		schema := &endpoint.ObjectStorageDataSchema{}
		diags.Append(v.convert(schema)...)
		settings.ObjectStorageSource.ResultSchema = schema
	}
	if v := m.EventSource; v != nil {
		event, d := v.convert()
		settings.ObjectStorageSource.EventSource = event
		diags.Append(d...)
	}
	return settings, diags
}

func (m *endpointObjectStorageSourceSettings) parse(e *endpoint.ObjectStorageSource) diag.Diagnostics {
	var diags diag.Diagnostics

	m.PathPattern = types.StringValue(e.GetPathPattern())

	if v := e.GetProvider(); v != nil {
		if m.Provider == nil {
			m.Provider = &endpointObjectStorageProvider{}
		}
		diags.Append(m.Provider.parse(v)...)
	}
	if v := e.GetFormat(); v != nil {
		if m.Format == nil {
			m.Format = &endpointObjectStorageFormat{}
		}
		diags.Append(m.Format.parse(v)...)
	}
	if v := e.GetResultTable(); v != nil {
		if m.ResultTable == nil {
			m.ResultTable = &endpointObjectStorageResultTable{}
		}
		diags.Append(m.ResultTable.parse(v)...)
	}
	if v := e.GetResultSchema(); v != nil {
		if m.ResultSchema == nil {
			m.ResultSchema = &endpointObjectStorageResultSchema{}
		}
		diags.Append(m.ResultSchema.parse(v)...)
	}
	if v := e.GetEventSource(); v != nil {
		if m.EventSource == nil {
			m.EventSource = &endpointObjectStorageEventSource{}
		}
		diags.Append(m.EventSource.parse(v)...)
	}

	return diags
}

func (m *endpointObjectStorageProvider) convert() (*endpoint.ObjectStorageProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	provider := endpoint.ObjectStorageProvider{}
	if v := m.Bucket; !v.IsNull() {
		provider.Bucket = v.ValueString()
	}
	if v := m.AwsAccessKeyId; !v.IsNull() {
		provider.AwsAccessKeyId = v.ValueString()
	}
	if v := m.AwsSecretAccessKey; !v.IsNull() {
		provider.AwsSecretAccessKey = v.ValueString()
	}
	if v := m.PathPrefix; !v.IsNull() {
		provider.PathPrefix = v.ValueString()
	}
	if v := m.Endpoint; !v.IsNull() {
		provider.Endpoint = v.ValueString()
	}
	if v := m.Region; !v.IsNull() {
		provider.Region = v.ValueString()
	}
	if v := m.UseSSL; !v.IsNull() {
		provider.UseSsl = v.ValueBool()
	}
	if v := m.VerifySSLCert; !v.IsNull() {
		provider.VerifySslCert = v.ValueBool()
	}

	return &provider, diags
}

func (m *endpointObjectStorageProvider) parse(e *endpoint.ObjectStorageProvider) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Bucket = types.StringValue(e.GetBucket())
	if v := e.GetAwsAccessKeyId(); len(v) > 0 {
		m.AwsAccessKeyId = types.StringValue(v)
	}
	if v := e.GetAwsSecretAccessKey(); len(v) > 0 {
		m.AwsSecretAccessKey = types.StringValue(v)
	}
	m.PathPrefix = types.StringValue(e.GetPathPrefix())
	m.Endpoint = types.StringValue(e.GetEndpoint())
	m.Region = types.StringValue(e.GetRegion())
	m.UseSSL = types.BoolValue(e.GetUseSsl())
	m.VerifySSLCert = types.BoolValue(e.GetVerifySslCert())

	return diags
}

func (m *endpointObjectStorageFormat) convert() (*endpoint.ObjectStorageReaderFormat, diag.Diagnostics) {
	var diags diag.Diagnostics

	format := endpoint.ObjectStorageReaderFormat{}

	if v := m.Csv; v != nil {
		csv := endpoint.ObjectStorageReaderFormat_Csv_{Csv: &endpoint.ObjectStorageReaderFormat_Csv{}}
		if v := m.Csv.Delimiter; !v.IsNull() {
			csv.Csv.Delimiter = v.ValueString()
		}
		if v := m.Csv.QuoteChar; !v.IsNull() {
			csv.Csv.QuoteChar = v.ValueString()
		}
		if v := m.Csv.EscapeChar; !v.IsNull() {
			csv.Csv.EscapeChar = v.ValueString()
		}
		if v := m.Csv.Encoding; !v.IsNull() {
			csv.Csv.Encoding = v.ValueString()
		}
		if v := m.Csv.DoubleQuote; !v.IsNull() {
			csv.Csv.DoubleQuote = v.ValueBool()
		}
		if v := m.Csv.NewlinesInValues; !v.IsNull() {
			csv.Csv.NewlinesInValues = v.ValueBool()
		}
		if v := m.Csv.BlockSize; !v.IsNull() {
			csv.Csv.BlockSize = v.ValueInt64()
		}
		if v := m.Csv.AdditionalReaderOptions; v != nil {
			event, d := v.convert()
			csv.Csv.AdditionalOptions = event
			diags.Append(d...)
		}
		if v := m.Csv.AdvancedOptions; v != nil {
			event, d := v.convert()
			csv.Csv.AdvancedOptions = event
			diags.Append(d...)
		}
		return &endpoint.ObjectStorageReaderFormat{
			Format: &csv,
		}, diags
	}
	if v := m.Jsonl; v != nil {
		jsonl := endpoint.ObjectStorageReaderFormat_Jsonl_{Jsonl: &endpoint.ObjectStorageReaderFormat_Jsonl{}}

		if v := m.Jsonl.NewlinesInValues; !v.IsNull() {
			jsonl.Jsonl.NewlinesInValues = v.ValueBool()
		}
		if v := m.Jsonl.UnexpectedFieldBehavior; !v.IsNull() {
			jsonl.Jsonl.UnexpectedFieldBehavior = endpoint.ObjectStorageReaderFormat_Jsonl_UnexpectedFieldBehavior(endpoint.ObjectStorageReaderFormat_Jsonl_UnexpectedFieldBehavior_value[v.ValueString()])
		}
		if v := m.Jsonl.BlockSize; !v.IsNull() {
			jsonl.Jsonl.BlockSize = v.ValueInt64()
		}
		return &endpoint.ObjectStorageReaderFormat{
			Format: &jsonl,
		}, diags
	}
	if v := m.Parquet; v != nil {
		return &endpoint.ObjectStorageReaderFormat{
			Format: &endpoint.ObjectStorageReaderFormat_Parquet_{},
		}, diags
	}
	if v := m.Avro; v != nil {
		return &endpoint.ObjectStorageReaderFormat{
			Format: &endpoint.ObjectStorageReaderFormat_Avro_{},
		}, diags
	}
	diags.AddError("missed s3 source format", "missed one of block: csv, parquet, avro or jsonl")

	return &format, diags
}

func (m *endpointObjectStorageFormat) parse(e *endpoint.ObjectStorageReaderFormat) diag.Diagnostics {
	var diags diag.Diagnostics

	if v := e.GetCsv(); v != nil {
		if m.Csv == nil {
			m.Csv = &endpointObjectStorageFormatCSV{}
		}
		m.Csv.Delimiter = types.StringValue(v.GetDelimiter())
		m.Csv.QuoteChar = types.StringValue(v.GetQuoteChar())
		m.Csv.EscapeChar = types.StringValue(v.GetEscapeChar())
		m.Csv.Encoding = types.StringValue(v.GetEncoding())
		m.Csv.DoubleQuote = types.BoolValue(v.GetDoubleQuote())
		m.Csv.NewlinesInValues = types.BoolValue(v.GetNewlinesInValues())
		m.Csv.BlockSize = types.Int64Value(v.GetBlockSize())
		if additionalOptions := v.GetAdditionalOptions(); additionalOptions != nil {
			if m.Csv.AdditionalReaderOptions == nil {
				m.Csv.AdditionalReaderOptions = &endpointObjectStorageFormatCSVAdditionalReaderOptions{}
			}
			diags.Append(m.Csv.AdditionalReaderOptions.parse(additionalOptions)...)
		}
		if advancedOptions := v.GetAdvancedOptions(); advancedOptions != nil {
			if m.Csv.AdvancedOptions == nil {
				m.Csv.AdvancedOptions = &endpointObjectStorageFormatCSVAdvancedOptions{}
			}
			diags.Append(m.Csv.AdvancedOptions.parse(advancedOptions)...)
		}
	}
	if v := e.GetParquet(); v != nil {
		m.Parquet = &endpointObjectStorageFormatParquet{}
	}
	if v := e.GetAvro(); v != nil {
		m.Avro = &endpointObjectStorageFormatAvro{}
	}
	if v := e.GetJsonl(); v != nil {
		m.Jsonl = &endpointObjectStorageFormatJsonl{}
		m.Jsonl.NewlinesInValues = types.BoolValue(v.GetNewlinesInValues())
		m.Jsonl.UnexpectedFieldBehavior = types.StringValue(v.GetUnexpectedFieldBehavior().String())
		m.Jsonl.BlockSize = types.Int64Value(v.GetBlockSize())
	}

	return diags
}

func (m *endpointObjectStorageFormatCSVAdvancedOptions) parse(e *endpoint.ObjectStorageReaderFormat_Csv_AdvancedOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	m.AutogenerateColumnNames = types.BoolValue(e.GetAutogenerateColumnNames())
	m.ColumnNames = convertSliceToTFStrings(e.GetColumnNames())
	m.SkipRows = types.Int64Value(e.GetSkipRows())
	m.SkipRowsAfterNames = types.Int64Value(e.GetSkipRowsAfterNames())

	return diags
}

func (m *endpointObjectStorageFormatCSVAdvancedOptions) convert() (*endpoint.ObjectStorageReaderFormat_Csv_AdvancedOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	advancedOptions := endpoint.ObjectStorageReaderFormat_Csv_AdvancedOptions{}
	if v := m.AutogenerateColumnNames; !v.IsNull() {
		advancedOptions.AutogenerateColumnNames = v.ValueBool()
	}
	if v := m.SkipRows; !v.IsNull() {
		advancedOptions.SkipRows = v.ValueInt64()
	}
	if v := m.SkipRowsAfterNames; !v.IsNull() {
		advancedOptions.SkipRowsAfterNames = v.ValueInt64()
	}
	if v := m.ColumnNames; v != nil {
		advancedOptions.ColumnNames = convertSliceTFStrings(v)
	}
	return &advancedOptions, diags
}

func (m *endpointObjectStorageFormatCSVAdditionalReaderOptions) parse(e *endpoint.ObjectStorageReaderFormat_Csv_AdditionalReaderOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	m.IncludeMissingColumns = types.BoolValue(e.GetIncludeMissingColumns())
	m.StringsCanBeNull = types.BoolValue(e.GetStringsCanBeNull())
	m.QuotedStringsCanBeNull = types.BoolValue(e.GetQuotedStringsCanBeNull())
	m.NullValues = convertSliceToTFStrings(e.GetNullValues())
	m.FalseValues = convertSliceToTFStrings(e.GetFalseValues())
	m.TrueValues = convertSliceToTFStrings(e.GetTrueValues())
	m.IncludeColumns = convertSliceToTFStrings(e.GetIncludeColumns())
	m.TimestampParsers = convertSliceToTFStrings(e.GetTimestampParsers())
	m.DecimalPoint = types.StringValue(e.GetDecimalPoint())

	return diags
}

func (m *endpointObjectStorageFormatCSVAdditionalReaderOptions) convert() (*endpoint.ObjectStorageReaderFormat_Csv_AdditionalReaderOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	additionalOptions := endpoint.ObjectStorageReaderFormat_Csv_AdditionalReaderOptions{}
	if v := m.IncludeMissingColumns; !v.IsNull() {
		additionalOptions.IncludeMissingColumns = v.ValueBool()
	}
	if v := m.StringsCanBeNull; !v.IsNull() {
		additionalOptions.StringsCanBeNull = v.ValueBool()
	}
	if v := m.QuotedStringsCanBeNull; !v.IsNull() {
		additionalOptions.QuotedStringsCanBeNull = v.ValueBool()
	}
	if v := m.DecimalPoint; !v.IsNull() {
		additionalOptions.DecimalPoint = v.ValueString()
	}
	if v := m.IncludeColumns; v != nil {
		additionalOptions.IncludeColumns = convertSliceTFStrings(v)
	}
	if v := m.NullValues; v != nil {
		additionalOptions.NullValues = convertSliceTFStrings(v)
	}
	if v := m.TrueValues; v != nil {
		additionalOptions.TrueValues = convertSliceTFStrings(v)
	}
	if v := m.FalseValues; v != nil {
		additionalOptions.FalseValues = convertSliceTFStrings(v)
	}
	if v := m.TimestampParsers; v != nil {
		additionalOptions.TimestampParsers = convertSliceTFStrings(v)
	}
	return &additionalOptions, diags
}

func (m *endpointObjectStorageResultTable) parse(e *endpoint.ObjectStorageResultTable) diag.Diagnostics {
	var diags diag.Diagnostics

	m.AddSystemCols = types.BoolValue(e.GetAddSystemCols())
	m.TableName = types.StringValue(e.GetTableName())
	m.TableNamespace = types.StringValue(e.GetTableNamespace())

	return diags
}

func (m *endpointObjectStorageResultTable) convert() (*endpoint.ObjectStorageResultTable, diag.Diagnostics) {
	var diags diag.Diagnostics

	table := endpoint.ObjectStorageResultTable{}
	if v := m.AddSystemCols; !v.IsNull() {
		table.AddSystemCols = v.ValueBool()
	}
	if v := m.TableName; !v.IsNull() {
		table.TableName = v.ValueString()
	}
	if v := m.TableNamespace; !v.IsNull() {
		table.TableNamespace = v.ValueString()
	}
	return &table, diags
}

func (m *endpointObjectStorageResultSchema) parse(e *endpoint.ObjectStorageDataSchema) diag.Diagnostics {
	var diags diag.Diagnostics
	if v := e.GetInfer(); v != nil {
		m.Infer = &endpointObjectStorageResultSchemaInfer{}
		m.DataSchema = nil
	}
	if v := e.GetDataSchema(); v != nil {
		m.DataSchema = &endpointObjectStorageDataSchema{}
		diags.Append(m.DataSchema.parse(v)...)
		m.Infer = nil
	}
	return diags
}

func (m *endpointObjectStorageResultSchema) convert(r *endpoint.ObjectStorageDataSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.DataSchema != nil {
		schema := &endpoint.DataSchema{}
		if schema != nil {
			diags.Append(m.DataSchema.convert(schema)...)
			r.Schema = &endpoint.ObjectStorageDataSchema_DataSchema{
				DataSchema: schema,
			}
		}
	}

	if m.Infer != nil {
		r.Schema = &endpoint.ObjectStorageDataSchema_Infer{}
	}

	return diags
}

func (m *endpointObjetStorageDataSchemaJsonFields) parse(json string) diag.Diagnostics {
	m.JsonFields = types.StringValue(json)
	return nil
}

func (m *endpointObjetStorageDataSchemaJsonFields) convert(r *string) diag.Diagnostics {
	*r = m.JsonFields.ValueString()
	return nil
}

func (m *endpointObjectStorageDataSchema) parse(e *endpoint.DataSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetFields() != nil:
		m.JsonFields = nil
		if m.Fields == nil {
			m.Fields = new(transferParserSchemaFields)
		}
		diags.Append(m.Fields.parse(e.GetFields())...)
	case e.GetJsonFields() != "":
		m.Fields = nil
		if m.JsonFields == nil {
			m.JsonFields = new(endpointObjetStorageDataSchemaJsonFields)
		}
		diags.Append(m.JsonFields.parse(e.GetJsonFields())...)
	default:
		diags.Append(diag.NewErrorDiagnostic("unknown schema type", fmt.Sprintf("%v", e.GetSchema())))
	}

	return diags
}

func (m *endpointObjectStorageDataSchema) convert(r *endpoint.DataSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case m.Fields != nil:
		fl := new(endpoint.FieldList)
		diags.Append(m.Fields.convert(fl)...)
		r.Schema = &endpoint.DataSchema_Fields{Fields: fl}
	case m.JsonFields != nil:
		jsn := new(string)
		diags.Append(m.JsonFields.convert(jsn)...)
		r.Schema = &endpoint.DataSchema_JsonFields{JsonFields: *jsn}
	}

	return diags
}

func (m *endpointObjectStorageEventSource) parse(e *endpoint.ObjectStorageEventSource) diag.Diagnostics {
	var diags diag.Diagnostics

	if v := e.GetSqs(); v != nil {
		if m.SQS == nil {
			m.SQS = &endpointObjectStorageEventSourceSQS{}
		}
		if name := v.GetQueueName(); name != "" {
			m.SQS.QueueName = types.StringValue(name)
		}
		if id := v.GetOwnerId(); id != "" {
			m.SQS.OwnerID = types.StringValue(id)
		}
		if keyID := v.GetAwsAccessKeyId(); keyID != "" {
			m.SQS.AwsAccessKeyId = types.StringValue(keyID)
		}
		if key := v.GetAwsSecretAccessKey(); key != "" {
			m.SQS.AwsSecretAccessKey = types.StringValue(key)
		}
		if endpoint := v.GetEndpoint(); endpoint != "" {
			m.SQS.Endpoint = types.StringValue(endpoint)
		}
		if region := v.GetRegion(); region != "" {
			m.SQS.Region = types.StringValue(region)
		}
		if use := v.GetUseSsl(); use {
			m.SQS.UseSSL = types.BoolValue(use)
		}
		if verify := v.GetVerifySslCert(); verify {
			m.SQS.VerifySSLCert = types.BoolValue(verify)
		}
	}
	if v := e.GetSns(); v != nil {
		if m.SNS == nil {
			m.SNS = &endpointObjectStorageEventSourceSNS{}
		}
	}
	if v := e.GetPubSub(); v != nil {
		if m.PubSub == nil {
			m.PubSub = &endpointObjectStorageEventSourcePubSub{}
		}
	}

	return diags
}

func (m *endpointObjectStorageEventSource) convert() (*endpoint.ObjectStorageEventSource, diag.Diagnostics) {
	var diags diag.Diagnostics

	event := endpoint.ObjectStorageEventSource{}
	if v := m.SQS; v != nil {
		sqs := endpoint.ObjectStorageEventSource_Sqs{Sqs: &endpoint.ObjectStorageEventSource_SQS{}}
		if v := m.SQS.QueueName; !v.IsNull() {
			sqs.Sqs.QueueName = v.ValueString()
		}
		if v := m.SQS.OwnerID; !v.IsNull() {
			sqs.Sqs.OwnerId = v.ValueString()
		}
		if v := m.SQS.AwsAccessKeyId; !v.IsNull() {
			sqs.Sqs.AwsAccessKeyId = v.ValueString()
		}
		if v := m.SQS.AwsSecretAccessKey; !v.IsNull() {
			sqs.Sqs.AwsSecretAccessKey = v.ValueString()
		}
		if v := m.SQS.Endpoint; !v.IsNull() {
			sqs.Sqs.Endpoint = v.ValueString()
		}
		if v := m.SQS.Region; !v.IsNull() {
			sqs.Sqs.Region = v.ValueString()
		}
		if v := m.SQS.UseSSL; !v.IsNull() {
			sqs.Sqs.UseSsl = v.ValueBool()
		}
		if v := m.SQS.VerifySSLCert; !v.IsNull() {
			sqs.Sqs.VerifySslCert = v.ValueBool()
		}
		return &endpoint.ObjectStorageEventSource{
			Source: &sqs,
		}, diags
	}

	if v := m.SNS; v != nil {
		return &endpoint.ObjectStorageEventSource{
			Source: &endpoint.ObjectStorageEventSource_Sns{},
		}, diags
	}
	if v := m.PubSub; v != nil {
		return &endpoint.ObjectStorageEventSource{
			Source: &endpoint.ObjectStorageEventSource_PubSub_{},
		}, diags
	}
	diags.AddError("missed s3 event source format", "missed one of block: sqs, sns, pub/sub")

	return &event, diags
}

func (m *endpointObjectStorageTargetSettings) parse(e *endpoint.ObjectStorageTarget) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Bucket = types.StringValue(e.GetBucket())
	m.BucketLayout = types.StringValue(e.GetBucketLayout())
	m.BucketLayoutColumn = types.StringValue(e.GetBucketLayoutColumn())
	m.BucketLayoutTimezone = types.StringValue(e.GetBucketLayoutTimezone())
	m.BufferInterval = types.StringValue(e.GetBufferInterval())
	m.BufferSize = types.StringValue(e.GetBufferSize())
	m.ServiceAccountID = types.StringValue(e.GetServiceAccountId())
	m.OutputFormat = types.StringValue(e.GetOutputFormat().String())
	m.OutputEncoding = types.StringValue(e.GetOutputEncoding().String())

	if v := e.GetConnection(); v != nil {
		if m.Connection == nil {
			m.Connection = &endpointObjectStorageConnection{}
		}
		diags.Append(m.Connection.parse(v)...)
	}
	if v := e.GetSerializerConfig(); v != nil {
		if m.SerializerConfig == nil {
			m.SerializerConfig = &endpointObjectStorageSerializerConfig{}
		}
		diags.Append(m.SerializerConfig.parse(v)...)
	}

	return diags
}

func (m *endpointObjectStorageTargetSettings) convert() (*transfer.EndpointSettings_ObjectStorageTarget, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := &transfer.EndpointSettings_ObjectStorageTarget{ObjectStorageTarget: &endpoint.ObjectStorageTarget{}}

	if v := m.Bucket; !v.IsNull() {
		settings.ObjectStorageTarget.Bucket = v.ValueString()
	}
	if v := m.BucketLayout; !v.IsNull() {
		settings.ObjectStorageTarget.BucketLayout = v.ValueString()
	}
	if v := m.BucketLayoutColumn; !v.IsNull() {
		settings.ObjectStorageTarget.BucketLayoutColumn = v.ValueString()
	}
	if v := m.BucketLayoutTimezone; !v.IsNull() {
		settings.ObjectStorageTarget.BucketLayoutTimezone = v.ValueString()
	}
	if v := m.ServiceAccountID; !v.IsNull() {
		settings.ObjectStorageTarget.ServiceAccountId = v.ValueString()
	}
	if v := m.BufferInterval; !v.IsNull() {
		settings.ObjectStorageTarget.BufferInterval = v.ValueString()
	}
	if v := m.BufferSize; !v.IsNull() {
		settings.ObjectStorageTarget.BufferSize = v.ValueString()
	}
	if v := m.OutputFormat; !v.IsNull() {
		settings.ObjectStorageTarget.OutputFormat = endpoint.ObjectStorageSerializationFormat(endpoint.ObjectStorageSerializationFormat_value[v.ValueString()])
	}
	if v := m.OutputEncoding; !v.IsNull() {
		settings.ObjectStorageTarget.OutputEncoding = endpoint.ObjectStorageCodec(endpoint.ObjectStorageCodec_value[v.ValueString()])
	}

	if v := m.Connection; v != nil {
		connection, d := v.convert()
		settings.ObjectStorageTarget.Connection = connection
		diags.Append(d...)
	}
	if v := m.SerializerConfig; v != nil {
		config, d := v.convert()
		settings.ObjectStorageTarget.SerializerConfig = config
		diags.Append(d...)
	}
	return settings, diags
}

func (m *endpointObjectStorageConnection) parse(e *endpoint.ObjectStorageConnection) diag.Diagnostics {
	var diags diag.Diagnostics

	if v := e.GetAwsAccessKeyId(); len(v) > 0 {
		m.AwsAccessKeyId = types.StringValue(v)
	}
	if v := e.GetAwsSecretAccessKey(); len(v) > 0 {
		m.AwsSecretAccessKey = types.StringValue(v)
	}
	m.Endpoint = types.StringValue(e.GetEndpoint())
	m.Region = types.StringValue(e.GetRegion())
	m.UseSSL = types.BoolValue(e.GetUseSsl())
	m.VerifySSLCert = types.BoolValue(e.GetVerifySslCert())

	return diags
}

func (m *endpointObjectStorageConnection) convert() (*endpoint.ObjectStorageConnection, diag.Diagnostics) {
	var diags diag.Diagnostics

	connection := endpoint.ObjectStorageConnection{}
	if v := m.AwsAccessKeyId; !v.IsNull() {
		connection.AwsAccessKeyId = v.ValueString()
	}
	if v := m.AwsSecretAccessKey; !v.IsNull() {
		connection.AwsSecretAccessKey = v.ValueString()
	}
	if v := m.Endpoint; !v.IsNull() {
		connection.Endpoint = v.ValueString()
	}
	if v := m.Region; !v.IsNull() {
		connection.Region = v.ValueString()
	}
	if v := m.UseSSL; !v.IsNull() {
		connection.UseSsl = v.ValueBool()
	}
	if v := m.VerifySSLCert; !v.IsNull() {
		connection.VerifySslCert = v.ValueBool()
	}

	return &connection, diags
}

func (m *endpointObjectStorageSerializerConfig) parse(e *endpoint.ObjectStorageSerializerConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	m.AnyAsString = types.BoolValue(e.GetAnyAsString())
	return diags
}

func (m *endpointObjectStorageSerializerConfig) convert() (*endpoint.ObjectStorageSerializerConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	config := endpoint.ObjectStorageSerializerConfig{}
	if v := m.AnyAsString; !v.IsNull() {
		config.AnyAsString = v.ValueBool()
	}
	return &config, diags
}
