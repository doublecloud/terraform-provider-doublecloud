package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointMetrikaSourceSettings struct {
	CounterIDs     []types.Int64            `tfsdk:"counter_ids"`
	Token          types.String             `tfsdk:"token"`
	MetrikaStreams []*endpointMetrikaStream `tfsdk:"metrika_stream"`
}

type endpointMetrikaStream struct {
	StreamType types.String   `tfsdk:"stream_type"`
	Columns    []types.String `tfsdk:"columns"`
}

func transferEndpointMetrikaStreamSchema() schema.Block {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"stream_type": schema.StringAttribute{
					Optional:    true,
					Description: "The type of the Metrika stream",
				},
				"columns": schema.ListAttribute{
					ElementType: types.StringType,
					Optional:    true,
					Description: "The columns included in the Metrika stream",
				},
			},
		},
		Description: "Configuration for Metrika streams",
	}
}
func transferEndpointMetrikaSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"counter_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Description: "List of counter IDs",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Access token",
				Sensitive:           true,
			},
		},
		Blocks: map[string]schema.Block{
			"metrika_stream": transferEndpointMetrikaStreamSchema(),
		},
	}
}

func (m *endpointMetrikaSourceSettings) parse(e *endpoint.MetrikaSource) diag.Diagnostics {
	var diags diag.Diagnostics
	if len(e.GetCounterIds()) > 0 {
		counterIDs := make([]types.Int64, len(e.CounterIds))
		for i, id := range e.CounterIds {
			counterIDs[i] = types.Int64Value(id)
		}
		m.CounterIDs = counterIDs
	} else {
		m.CounterIDs = []types.Int64{}
	}

	if len(e.GetStreams()) > 0 {
		metrikaStreams := make([]*endpointMetrikaStream, len(e.GetStreams()))
		for i, stream := range e.GetStreams() {
			parsedStream := &endpointMetrikaStream{}
			diags = append(diags, parsedStream.parse(stream)...)
			metrikaStreams[i] = parsedStream
		}
		m.MetrikaStreams = metrikaStreams
	} else {
		m.MetrikaStreams = []*endpointMetrikaStream{}
	}

	return diags
}

func (m *endpointMetrikaStream) parse(e *endpoint.MetrikaStream) diag.Diagnostics {
	var diags diag.Diagnostics
	if e == nil {
		m = nil
	}
	if len(e.GetColumns()) > 0 {
		columns := make([]types.String, len(e.Columns))
		for i, column := range e.Columns {
			columns[i] = types.StringValue(column)
		}
		m.Columns = columns
	} else {
		m.Columns = []types.String{}
	}

	if e.GetType() != endpoint.MetrikaStreamType_METRIKA_STREAM_TYPE_UNSPECIFIED {
		m.StreamType = types.StringValue(e.GetType().String())
	}

	return diags
}

func (m *endpointMetrikaSourceSettings) convert() (*transfer.EndpointSettings_MetrikaSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	metrikaSource := endpoint.MetrikaSource{}
	if len(m.CounterIDs) > 0 {
		counterIDs := make([]int64, len(m.CounterIDs))
		for i, id := range m.CounterIDs {
			counterIDs[i] = id.ValueInt64()
		}
		metrikaSource.CounterIds = counterIDs
	} else {
		metrikaSource.CounterIds = []int64{}
	}

	metrikaSource.Token = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Token.ValueString()}}

	if len(m.MetrikaStreams) > 0 {
		metrikaStreams := make([]*endpoint.MetrikaStream, len(m.MetrikaStreams))
		for i, stream := range m.MetrikaStreams {
			convertedStream, diag := stream.convert()
			diags = append(diags, diag...)
			metrikaStreams[i] = convertedStream
		}
		metrikaSource.Streams = metrikaStreams
	} else {
		metrikaSource.Streams = []*endpoint.MetrikaStream{}
	}

	return &transfer.EndpointSettings_MetrikaSource{MetrikaSource: &metrikaSource}, diags
}

func (m *endpointMetrikaStream) convert() (*endpoint.MetrikaStream, diag.Diagnostics) {
	var diags diag.Diagnostics
	metrikaStream := &endpoint.MetrikaStream{}
	if len(m.Columns) > 0 {
		columns := make([]string, len(m.Columns))
		for i, column := range m.Columns {
			columns[i] = column.ValueString()
		}
		metrikaStream.Columns = columns
	} else {
		metrikaStream.Columns = []string{}
	}
	metrikaStream.Type = endpoint.MetrikaStreamType(endpoint.MetrikaStreamType_value[m.StreamType.ValueString()])

	return metrikaStream, diags
}
