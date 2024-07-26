package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointKinesisSourceSettings struct {
	StreamName types.String    `tfsdk:"stream_name"`
	Region     types.String    `tfsdk:"region"`
	AccessKey  types.String    `tfsdk:"aws_access_key_id"`
	SecretKey  types.String    `tfsdk:"aws_secret_access_key"`
	Parser     *endpointParser `tfsdk:"parser"`
}

func transferEndpointKinesisSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"stream_name": schema.StringAttribute{
				MarkdownDescription: "Name of AWS Kinesis Data Stream",
				Optional:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Name of AWS Region where stream is deployed",
				Optional:            true,
			},
			"aws_access_key_id": schema.StringAttribute{
				MarkdownDescription: "AWS Access Key with access to this stream",
				Sensitive:           true,
				Optional:            true,
			},
			"aws_secret_access_key": schema.StringAttribute{
				MarkdownDescription: "AWS Secret Access Key with access to this stream",
				Sensitive:           true,
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"parser": endpointKafkaParserSchema(),
		},
	}
}

func (m *endpointKinesisSourceSettings) parse(e *endpoint.KinesisSource) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Region = types.StringValue(e.Region)
	m.StreamName = types.StringValue(e.StreamName)
	if e.AwsSecretAccessKey != "" {
		m.AccessKey = types.StringValue(e.AwsAccessKeyId)
	}
	if e.AwsSecretAccessKey != "" {
		m.SecretKey = types.StringValue(e.AwsSecretAccessKey)
	}

	if prsr := e.GetParser(); prsr != nil {
		if m.Parser == nil {
			m.Parser = new(endpointParser)
		}
		diags.Append(m.Parser.parse(prsr)...)
	} else {
		m.Parser = nil
	}

	return diags
}

func kinesisSourceEndpointSettings(m *endpointKinesisSourceSettings) (*transfer.EndpointSettings_KinesisSource, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_KinesisSource{KinesisSource: &endpoint.KinesisSource{}}
	var diags diag.Diagnostics
	settings.KinesisSource.Region = m.Region.ValueString()
	settings.KinesisSource.StreamName = m.StreamName.ValueString()
	settings.KinesisSource.AwsAccessKeyId = m.AccessKey.ValueString()
	settings.KinesisSource.AwsSecretAccessKey = m.SecretKey.ValueString()

	if m.Parser != nil {
		settings.KinesisSource.Parser = new(endpoint.Parser)
		diags.Append(m.Parser.convert(settings.KinesisSource.Parser)...)
	}

	return settings, diags
}
