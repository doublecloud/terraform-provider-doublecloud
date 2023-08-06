package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointTLSMode struct {
	CACertificate types.String `tfsdk:"ca_certificate"`
}

func convertSliceTFStrings(s []types.String) []string {
	if s == nil {
		return nil
	}
	ret := make([]string, len(s))
	for i := 0; i < len(s); i++ {
		ret[i] = s[i].ValueString()
	}
	return ret
}

func convertSliceToTFStrings(s []string) []types.String {
	if s == nil {
		return nil
	}
	ret := make([]types.String, len(s))
	for i := 0; i < len(s); i++ {
		ret[i] = types.StringValue(s[i])
	}
	return ret
}

func transferEndpointCleanupPolicyValidator() validator.String {
	names := make([]string, len(endpoint.CleanupPolicy_name))
	for i, v := range endpoint.CleanupPolicy_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func transferEndpointParserSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"json_parser": transferEndpointGenericParserSchema(),
			"tskv_parser": transferEndpointGenericParserSchema(),
		},
	}
}

func transferEndpointTLSMode() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"ca_certificate": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "X.509 certificate of the certificate authority which issued the server's certificate, in PEM format. When CA certificate is specified TLS is used to connect to the server",
			},
		},
	}
}

func convertTLSMode(m *endpointTLSMode) *endpoint.TLSMode {
	if m == nil {
		return &endpoint.TLSMode{TlsMode: &endpoint.TLSMode_Disabled{}}
	}
	if m.CACertificate.IsNull() {
		return &endpoint.TLSMode{TlsMode: &endpoint.TLSMode_Enabled{Enabled: &endpoint.TLSConfig{}}}
	}
	return &endpoint.TLSMode{TlsMode: &endpoint.TLSMode_Enabled{Enabled: &endpoint.TLSConfig{CaCertificate: m.CACertificate.ValueString()}}}
}

func transferEndpointGenericParserSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			// "data_schema": ...endpoint
			"null_keys_allowed": schema.BoolAttribute{
				MarkdownDescription: "Allow null keys, if no - null keys will be putted to unparsed data",
				Optional:            true,
			},
			"add_rest_column": schema.BoolAttribute{
				MarkdownDescription: "Will add _rest column for all unknown fields",
				Optional:            true,
			},
		},
	}
}

func transferEndpointDataSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"json_fields": schema.StringAttribute{
				Optional: true,
			},
		},
		// Blocks: map[string]schema.Block{
		// 	"fields": schema.SingleNestedBlock{
		// 		"field":
		// 	}
		// }
	}
}

func transferEndpointSerializerSchemaBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: "Data serialization format",
		Blocks: map[string]schema.Block{
			"auto": schema.SingleNestedBlock{MarkdownDescription: "Select the serialization format automatically"},
			"json": schema.SingleNestedBlock{MarkdownDescription: "Serialize data in json format"},
			"debezium": schema.SingleNestedBlock{
				MarkdownDescription: "Serialize data in json format",
				Blocks: map[string]schema.Block{
					"parameter": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"key":   schema.StringAttribute{Optional: true},
								"value": schema.StringAttribute{Optional: true},
							},
						},
					},
				},
			},
		},
	}
}

func transferObjectTransferStageValidator() validator.String {
	names := make([]string, len(endpoint.ObjectTransferStage_name))
	for i, v := range endpoint.ObjectTransferStage_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}
