package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointMetricaSourceSettings struct {
	CounterIDs     []types.Int64            `tfsdk:"counter_ids"`
	Token          types.String             `tfsdk:"token"`
	MetricaStreams []*endpointMetricaStream `tfsdk:"metrica_stream"`
}

type endpointMetricaStream struct {
	StreamType types.String `tfsdk:"stream_type"`
}

func transferEndpointMetricaStreamSchema() schema.Block {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"stream_type": schema.StringAttribute{
					Optional:    true,
					Description: "The type of the Metrica stream",
				},
			},
		},
		Description: "Configuration for Metrica streams",
	}
}
func transferEndpointMetricaSourceSchema() schema.Block {
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
			"metrica_stream": transferEndpointMetricaStreamSchema(),
		},
	}
}

func (m *endpointMetricaSourceSettings) parse(e *endpoint.MetricaSource) diag.Diagnostics {
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
		metricaStreams := make([]*endpointMetricaStream, len(e.GetStreams()))
		for i, stream := range e.GetStreams() {
			parsedStream := &endpointMetricaStream{}
			diags = append(diags, parsedStream.parse(stream)...)
			metricaStreams[i] = parsedStream
		}
		m.MetricaStreams = metricaStreams
	} else {
		m.MetricaStreams = []*endpointMetricaStream{}
	}

	return diags
}

func (m *endpointMetricaStream) parse(e *endpoint.MetricaStream) diag.Diagnostics {
	var diags diag.Diagnostics
	if e == nil {
		m = nil
	}

	if e.GetType() != endpoint.MetricaStreamType_METRICA_STREAM_TYPE_UNSPECIFIED {
		m.StreamType = types.StringValue(e.GetType().String())
	}

	return diags
}

func (m *endpointMetricaSourceSettings) convert() (*transfer.EndpointSettings_MetricaSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	metricaSource := endpoint.MetricaSource{}
	if len(m.CounterIDs) > 0 {
		counterIDs := make([]int64, len(m.CounterIDs))
		for i, id := range m.CounterIDs {
			counterIDs[i] = id.ValueInt64()
		}
		metricaSource.CounterIds = counterIDs
	} else {
		metricaSource.CounterIds = []int64{}
	}

	metricaSource.Token = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Token.ValueString()}}

	if len(m.MetricaStreams) > 0 {
		metricaStreams := make([]*endpoint.MetricaStream, len(m.MetricaStreams))
		for i, stream := range m.MetricaStreams {
			convertedStream, diag := stream.convert()
			diags = append(diags, diag...)
			metricaStreams[i] = convertedStream
		}
		metricaSource.Streams = metricaStreams
	} else {
		metricaSource.Streams = []*endpoint.MetricaStream{}
	}

	return &transfer.EndpointSettings_MetricaSource{MetricaSource: &metricaSource}, diags
}

func (m *endpointMetricaStream) convert() (*endpoint.MetricaStream, diag.Diagnostics) {
	return &endpoint.MetricaStream{Type: endpoint.MetricaStreamType(endpoint.MetricaStreamType_value[m.StreamType.ValueString()])}, diag.Diagnostics{}
}
