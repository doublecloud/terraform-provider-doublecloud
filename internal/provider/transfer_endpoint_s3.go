package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
)

type endpointS3SourceSettings struct {
	Dataset     types.String        `tfsdk:"dataset"`
	PathPattern types.String        `tfsdk:"path_pattern"`
	Schema      types.String        `tfsdk:"schema"`
	Format      *endpointS3Format   `tfsdk:"format"`
	Provider    *endpointS3Provider `tfsdk:"provider"`
}

type endpointS3Format struct {
	Csv     *endpointS3FormatCSV     `tfsdk:"csv"`
	Parquet *endpointS3FormatParquet `tfsdk:"parquet"`
	Avro    *endpointS3FormatAvro    `tfsdk:"avro"`
	Jsonl   *endpointS3FormatJsonl   `tfsdk:"jsonl"`
}

type endpointS3FormatCSV struct {
	Delimiter               types.String `tfsdk:"delimiter"`
	QuoteChar               types.String `tfsdk:"quote_char"`
	EscapeChar              types.String `tfsdk:"escape_char"`
	Encoding                types.String `tfsdk:"encoding"`
	DoubleQuote             types.Bool   `tfsdk:"double_quote"`
	NewlinesInValues        types.Bool   `tfsdk:"newlines_in_values"`
	BlockSize               types.Int64  `tfsdk:"block_size"`
	AdditionalReaderOptions types.String `tfsdk:"additional_reader_options"`
	AdvancedOptions         types.String `tfsdk:"advanced_options"`
}
type endpointS3FormatParquet struct {
	BufferSize types.Int64    `tfsdk:"buffer_size"`
	Columns    []types.String `tfsdk:"columns"`
	BatchSize  types.Int64    `tfsdk:"batch_size"`
}
type endpointS3FormatAvro struct{}
type endpointS3FormatJsonl struct {
	NewlinesInValues         types.Bool   `tfsdk:"newlines_in_values"`
	UnexpectedFieldBehaviour types.String `tfsdk:"unexpected_field_behavior"`
	BlockSize                types.Int64  `tfsdk:"block_size"`
}

type endpointS3Provider struct {
	Bucket             types.String `tfsdk:"bucket"`
	AwsAccessKeyId     types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	PathPrefix         types.String `tfsdk:"path_prefix"`
	Endpoint           types.String `tfsdk:"endpoint"`
	UseSSL             types.Bool   `tfsdk:"use_ssl"`
	VerifySSLCert      types.Bool   `tfsdk:"verify_ssl_cert"`
}

func transferUnexpectedFieldBehaviorValidator() validator.String {
	names := make([]string, len(endpoint_airbyte.S3Source_Jsonl_UnexpectedFieldBehavior_name))
	for i, v := range endpoint_airbyte.S3Source_Jsonl_UnexpectedFieldBehavior_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func transferEndpointS3SourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"dataset": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Dataset",
			},
			"path_pattern": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Path pattern",
			},
			"schema": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Schema",
			},
		},
		Blocks: map[string]schema.Block{
			"format": transferEndpointS3SourceFormatSchema(),
			"provider": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Bucket",
					},
					"aws_access_key_id": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Access key ID",
					},
					"aws_secret_access_key": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Secret access key",
					},
					"path_prefix": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Path prefix",
					},
					"endpoint": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Endpoint",
					},
					"use_ssl": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "",
					},
					"verify_ssl_cert": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "",
					},
				},
			},
		},
	}
}

func transferEndpointS3SourceFormatSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"csv": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"delimiter": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Delimiter",
					},
					"quote_char": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Quote character",
					},
					"escape_char": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Escape character",
					},
					"encoding": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "",
					},
					"double_quote": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Replace double quotes with single quotes",
					},
					"newlines_in_values": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Allow newline characters in values",
					},
					"block_size": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
						MarkdownDescription: "Block size",
					},
					"additional_reader_options": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "",
					},
					"advanced_options": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Advanced options",
					},
				},
			},
			"parquet": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"buffer_size": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
						MarkdownDescription: "Buffer size",
					},
					"columns": schema.ListAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "List of columns",
					},
					"batch_size": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
						MarkdownDescription: "Batch size",
					},
				},
			},
			"avro": schema.SingleNestedBlock{},
			"jsonl": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"newlines_in_values": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
						MarkdownDescription: "Allow newline characters in values",
					},
					"unexpected_field_behavior": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						Validators:          []validator.String{transferUnexpectedFieldBehaviorValidator()},
						MarkdownDescription: "",
					},
					"block_size": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
						MarkdownDescription: "Block size",
					},
				},
			},
		},
	}
}

func (m *endpointS3SourceSettings) convert() (*transfer.EndpointSettings_S3Source, diag.Diagnostics) {
	var diags diag.Diagnostics
	settings := &transfer.EndpointSettings_S3Source{S3Source: &endpoint_airbyte.S3Source{}}

	if v := m.Dataset; !v.IsNull() {
		settings.S3Source.Dataset = v.ValueString()
	}
	if v := m.PathPattern; !v.IsNull() {
		settings.S3Source.PathPattern = v.ValueString()
	}
	if v := m.Schema; !v.IsNull() {
		settings.S3Source.Schema = v.ValueString()
	}
	if v := m.Format; v != nil {
		format, d := v.convert()
		settings.S3Source.Format = format
		diags.Append(d...)
	}
	if v := m.Provider; v != nil {
		provider, d := v.convert()
		settings.S3Source.Provider = provider
		diags.Append(d...)
	}

	return settings, diags
}

func (m *endpointS3Format) convert() (*endpoint_airbyte.S3Source_Format, diag.Diagnostics) {
	var diags diag.Diagnostics

	format := endpoint_airbyte.S3Source_Format{}

	if v := m.Csv; v != nil {
		csv := endpoint_airbyte.S3Source_Format_Csv{Csv: &endpoint_airbyte.S3Source_Csv{}}
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
		if v := m.Csv.AdditionalReaderOptions; !v.IsNull() {
			csv.Csv.AdditionalReaderOptions = v.ValueString()
		}
		if v := m.Csv.AdvancedOptions; !v.IsNull() {
			csv.Csv.AdvancedOptions = v.ValueString()
		}
		return &endpoint_airbyte.S3Source_Format{
			Format: &csv,
		}, diags
	}
	if v := m.Parquet; v != nil {
		parquet := endpoint_airbyte.S3Source_Format_Parquet{Parquet: &endpoint_airbyte.S3Source_Parquet{}}

		if v := m.Parquet.BufferSize; !v.IsNull() {
			parquet.Parquet.BufferSize = v.ValueInt64()
		}
		if v := m.Parquet.Columns; v != nil {
			parquet.Parquet.Columns = convertSliceTFStrings(v)
		}
		if v := m.Parquet.BatchSize; !v.IsNull() {
			parquet.Parquet.BatchSize = v.ValueInt64()
		}
		return &endpoint_airbyte.S3Source_Format{
			Format: &parquet,
		}, diags
	}
	if v := m.Avro; v != nil {
		return &endpoint_airbyte.S3Source_Format{
			Format: &endpoint_airbyte.S3Source_Format_Avro{},
		}, diags
	}
	if v := m.Jsonl; v != nil {
		jsonl := endpoint_airbyte.S3Source_Format_Jsonl{Jsonl: &endpoint_airbyte.S3Source_Jsonl{}}

		if v := m.Jsonl.NewlinesInValues; !v.IsNull() {
			jsonl.Jsonl.NewlinesInValues = v.ValueBool()
		}
		if v := m.Jsonl.UnexpectedFieldBehaviour; !v.IsNull() {
			jsonl.Jsonl.UnexpectedFieldBehavior = endpoint_airbyte.S3Source_Jsonl_UnexpectedFieldBehavior(endpoint_airbyte.S3Source_Jsonl_UnexpectedFieldBehavior_value[v.ValueString()])
		}
		if v := m.Jsonl.BlockSize; !v.IsNull() {
			jsonl.Jsonl.BlockSize = v.ValueInt64()
		}
		return &endpoint_airbyte.S3Source_Format{
			Format: &jsonl,
		}, diags
	}
	diags.AddError("missed s3 source format", "missed one of block: csv, parquet, avro or jsonl")

	return &format, diags
}

func (m *endpointS3Provider) convert() (*endpoint_airbyte.S3Source_Provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	provider := endpoint_airbyte.S3Source_Provider{}
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
	if v := m.UseSSL; !v.IsNull() {
		provider.UseSsl = v.ValueBool()
	}
	if v := m.VerifySSLCert; !v.IsNull() {
		provider.VerifySslCert = v.ValueBool()
	}

	return &provider, diags
}

func (m *endpointS3SourceSettings) parse(e *endpoint_airbyte.S3Source) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Dataset = types.StringValue(e.Dataset)
	m.PathPattern = types.StringValue(e.PathPattern)
	m.Schema = types.StringValue(e.Schema)

	diags.Append(m.Format.parse(e.Format)...)
	diags.Append(m.Provider.parse(e.Provider)...)

	return diags
}

func (m *endpointS3Provider) parse(e *endpoint_airbyte.S3Source_Provider) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Bucket = types.StringValue(e.Bucket)
	m.AwsAccessKeyId = types.StringValue(e.AwsAccessKeyId)
	m.AwsSecretAccessKey = types.StringValue(e.AwsSecretAccessKey)
	m.PathPrefix = types.StringValue(e.PathPrefix)
	m.Endpoint = types.StringValue(e.Endpoint)
	m.UseSSL = types.BoolValue(e.UseSsl)
	m.VerifySSLCert = types.BoolValue(e.VerifySslCert)

	return diags
}

func (m *endpointS3Format) parse(e *endpoint_airbyte.S3Source_Format) diag.Diagnostics {
	var diags diag.Diagnostics

	if v := e.GetCsv(); v != nil {
		m.Csv = &endpointS3FormatCSV{}
		m.Csv.Delimiter = types.StringValue(v.Delimiter)
		m.Csv.QuoteChar = types.StringValue(v.QuoteChar)
		m.Csv.EscapeChar = types.StringValue(v.EscapeChar)
		m.Csv.Encoding = types.StringValue(v.Encoding)
		m.Csv.DoubleQuote = types.BoolValue(v.DoubleQuote)
		m.Csv.NewlinesInValues = types.BoolValue(v.NewlinesInValues)
		m.Csv.BlockSize = types.Int64Value(v.BlockSize)
		m.Csv.AdditionalReaderOptions = types.StringValue(v.AdditionalReaderOptions)
		m.Csv.AdvancedOptions = types.StringValue(v.AdvancedOptions)
	}
	if v := e.GetParquet(); v != nil {
		m.Parquet = &endpointS3FormatParquet{}
		m.Parquet.BufferSize = types.Int64Value(v.BufferSize)
		m.Parquet.Columns = convertSliceToTFStrings(v.Columns)
		m.Parquet.BatchSize = types.Int64Value(v.BatchSize)
	}
	if v := e.GetAvro(); v != nil {
		m.Avro = &endpointS3FormatAvro{}
	}
	if v := e.GetJsonl(); v != nil {
		m.Jsonl = &endpointS3FormatJsonl{}
		m.Jsonl.NewlinesInValues = types.BoolValue(v.NewlinesInValues)
		m.Jsonl.UnexpectedFieldBehaviour = types.StringValue(v.UnexpectedFieldBehavior.String())
		m.Jsonl.BlockSize = types.Int64Value(v.BlockSize)
	}

	return diags
}
