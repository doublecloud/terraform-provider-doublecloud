package provider

import (
	"context"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointKafkaSourceSettings struct {
	Connection *endpointKafkaConnectionOptions `tfsdk:"connection"`
	Auth       *endpointKafkaAuth              `tfsdk:"auth"`
	TopicName  types.String                    `tfsdk:"topic_name"`
	// TODO: Parser
}

type endpointKafkaConnectionOptions struct {
	ClusterId types.String            `tfsdk:"cluster_id"`
	OnPremise *endpointOnPremiseKafka `tfsdk:"on_premise"`
}

type endpointKafkaAuth struct {
	SASL   *endpointKafkaAuthSASL  `tfsdk:"sasl"`
	NoAuth *endpointKafkAuthNoAuth `tfsdk:"no_auth"`
}

type endpointKafkaAuthSASL struct {
	User      types.String `tfsdk:"user"`
	Password  types.String `tfsdk:"password"`
	Mechanism types.String `tfsdk:"mechanism"`
}

type endpointKafkAuthNoAuth struct{}

type endpointOnPremiseKafka struct {
	BrokerUrls []types.String   `tfsdk:"broker_urls"`
	TLSMode    *endpointTLSMode `tfsdk:"tls_mode"`
}

type endpointKafkaTargetSettings struct {
	Connection    *endpointKafkaConnectionOptions `tfsdk:"connection"`
	Auth          *endpointKafkaAuth              `tfsdk:"auth"`
	TopicSettings *endpointKafkaTopicSettings     `tfsdk:"topic_settings"`
	Serializer    *endpointSerializer             `tfsdk:"serializer"`
}

type endpointKafkaTopicSettings struct {
	Topic       *endpointKafkaTargetTopic `tfsdk:"topic"`
	TopicPrefix types.String              `tfsdk:"topic_prefix"`
}

type endpointKafkaTargetTopic struct {
	TopicName   types.String `tfsdk:"topic_name"`
	SaveTxOrder types.Bool   `tfsdk:"save_tx_order"`
}

type endpointSerializer struct {
	Auto     *endpointSerializerAuto     `tfsdk:"auto"`
	JSON     *endpointSerializerJSON     `tfsdk:"json"`
	Debezium *endpointSerializerDebezium `tfsdk:"debezium"`
}

type endpointSerializerAuto struct{}
type endpointSerializerJSON struct{}
type endpointSerializerDebezium struct {
	Parameter *[]endpointSerializerDebeziumParameter `tfsdk:"parameter"`
}

type endpointSerializerDebeziumParameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func transferEndpointKafkaConnectionSchemaBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		// MarkdownDescription: ,
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{Optional: true},
		},
		Blocks: map[string]schema.Block{
			"on_premise": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"broker_urls": schema.ListAttribute{
						MarkdownDescription: "Kafka broker URLs",
						ElementType:         types.StringType,
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

func transferEndpointKafkaMechanismValidator() validator.String {
	names := make([]string, len(endpoint.KafkaMechanism_name))
	for i, v := range endpoint.KafkaMechanism_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func transferEndpointKafkaAuthSchemaBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"sasl": schema.SingleNestedBlock{
				MarkdownDescription: "Authentication with SASL",
				Attributes: map[string]schema.Attribute{
					"user":     schema.StringAttribute{Optional: true},
					"password": schema.StringAttribute{Optional: true, Sensitive: true},
					"mechanism": schema.StringAttribute{
						Optional:   true,
						Validators: []validator.String{transferEndpointKafkaMechanismValidator()},
					},
				},
			},
			"no_auth": schema.SingleNestedBlock{
				MarkdownDescription: "No authentication",
			},
		},
	}
}

func transferEndpointKafkaSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"topic_name": schema.StringAttribute{
				MarkdownDescription: "Full source topic name",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"connection": transferEndpointKafkaConnectionSchemaBlock(),
			"auth":       transferEndpointKafkaAuthSchemaBlock(),
		},
	}
}

func transferEndpointKafkaTargetTopicSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"topic_name": schema.StringAttribute{
				MarkdownDescription: "Topic name",
				Optional:            true,
			},
			"save_tx_order": schema.BoolAttribute{
				MarkdownDescription: "Save transactions order. Not to split events queue into separate per-table queues.",
				Optional:            true,
			},
		},
	}
}

func transferEndpointKafkaTargetTopicSettingsSchemaBlock() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"topic_prefix": schema.StringAttribute{
				MarkdownDescription: "Analogue of the Debezium setting database.server.name. Messages will be sent to topic with name <topic_prefix>.<schema>.<table_name>.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"topic": transferEndpointKafkaTargetTopicSchema(),
		},
	}
}

func transferEndpointKafkaTargetSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"connection":     transferEndpointKafkaConnectionSchemaBlock(),
			"auth":           transferEndpointKafkaAuthSchemaBlock(),
			"topic_settings": transferEndpointKafkaTargetTopicSettingsSchemaBlock(),
			"serializer":     transferEndpointSerializerSchemaBlock(),
		},
	}
}

func convertKafkaConnectionOptions(m *endpointKafkaConnectionOptions) (*endpoint.KafkaConnectionOptions, diag.Diagnostics) {
	var diag diag.Diagnostics

	options := &endpoint.KafkaConnectionOptions{}

	if cluster_id := m.ClusterId; !cluster_id.IsNull() {
		options.Connection = &endpoint.KafkaConnectionOptions_ClusterId{ClusterId: cluster_id.ValueString()}
	}
	if on_premise := m.OnPremise; on_premise != nil {
		tlsMode := convertTLSMode(m.OnPremise.TLSMode)

		options.Connection = &endpoint.KafkaConnectionOptions_OnPremise{
			OnPremise: &endpoint.OnPremiseKafka{
				BrokerUrls: convertSliceTFStrings(m.OnPremise.BrokerUrls),
				TlsMode:    tlsMode,
			},
		}
	}

	if options.Connection == nil {
		diag.AddError("unknown connection", "required one of fields: cluster_id or on_premise")
	}
	return options, diag
}

func convertKafkaAuth(m *endpointKafkaAuth) (*endpoint.KafkaAuth, diag.Diagnostics) {
	var diag diag.Diagnostics

	options := &endpoint.KafkaAuth{}

	if m.NoAuth != nil {
		options.Security = &endpoint.KafkaAuth_NoAuth{}
	}
	if m.SASL != nil && m.NoAuth == nil {
		sasl := &endpoint.KafkaSaslSecurity{
			User:     m.SASL.User.ValueString(),
			Password: &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.SASL.Password.ValueString()}},
		}
		if !m.SASL.Mechanism.IsNull() {
			sasl.Mechanism = endpoint.KafkaMechanism(endpoint.KafkaMechanism_value[m.SASL.Mechanism.ValueString()])
		}

		options.Security = &endpoint.KafkaAuth_Sasl{
			Sasl: sasl,
		}
	}

	if options.Security == nil {
		diag.AddError("unknown auth", "required one of blocks: no_auth or sasl")
	}
	return options, diag
}

func kafkaSourceEndpointSettings(m *endpointKafkaSourceSettings) (*transfer.EndpointSettings_KafkaSource, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_KafkaSource{KafkaSource: &endpoint.KafkaSource{}}
	var diags diag.Diagnostics

	connection, d := convertKafkaConnectionOptions(m.Connection)
	diags.Append(d...)
	auth, d := convertKafkaAuth(m.Auth)
	diags.Append(d...)
	settings.KafkaSource.Connection = connection
	settings.KafkaSource.Auth = auth
	settings.KafkaSource.TopicName = m.TopicName.ValueString()
	// TODO: describe settings.KafkaSource.Parser

	return settings, diags
}

func convertKafkaTargetTopicSettings(m *endpointKafkaTopicSettings) (*endpoint.KafkaTargetTopicSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	settings := &endpoint.KafkaTargetTopicSettings{}
	if m.Topic != nil {
		settings.TopicSettings = &endpoint.KafkaTargetTopicSettings_Topic{
			Topic: &endpoint.KafkaTargetTopic{
				TopicName:   m.Topic.TopicName.ValueString(),
				SaveTxOrder: m.Topic.SaveTxOrder.ValueBool(),
			},
		}
	}
	if m.Topic == nil && !m.TopicPrefix.IsNull() {
		settings.TopicSettings = &endpoint.KafkaTargetTopicSettings_TopicPrefix{
			TopicPrefix: m.TopicPrefix.ValueString(),
		}
	}
	if settings.TopicSettings == nil {
		diags.AddError("unknown kafka_target.topic_settings", "specify oneof: topic block or topic_prefix attribut")
	}

	return settings, diags
}

func convertSerializer(m *endpointSerializer) (*endpoint.Serializer, diag.Diagnostics) {
	var diags diag.Diagnostics

	if m == nil {
		diags.AddError("unknown serializer", "specify serializer block")
		return nil, diags
	}

	s := &endpoint.Serializer{}
	if m.Auto != nil && s.Serializer == nil {
		s.Serializer = &endpoint.Serializer_SerializerAuto{}
	}
	if m.JSON != nil && s.Serializer == nil {
		s.Serializer = &endpoint.Serializer_SerializerJson{}
	}
	if m.Debezium != nil && s.Serializer == nil {
		parameters := make([]*endpoint.DebeziumSerializerParameter, 0)
		if m.Debezium.Parameter != nil {
			p := *m.Debezium.Parameter
			parameters = make([]*endpoint.DebeziumSerializerParameter, len(p))
			for i := 0; i < len(p); i++ {
				parameters[i] = &endpoint.DebeziumSerializerParameter{
					Key:   p[i].Key.ValueString(),
					Value: p[i].Value.ValueString(),
				}
			}
		}
		s.Serializer = &endpoint.Serializer_SerializerDebezium{
			SerializerDebezium: &endpoint.SerializerDebezium{
				SerializerParameters: parameters}}
	}
	if s.Serializer == nil {
		diags.AddError("unknown kafka_target.serializer", "specify one of blocks: auto, json or debezium")
	}
	return s, diags
}

func kafkaTargetEndpointSettings(m *endpointKafkaTargetSettings) (*transfer.EndpointSettings_KafkaTarget, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_KafkaTarget{KafkaTarget: &endpoint.KafkaTarget{}}
	var diags, d diag.Diagnostics

	settings.KafkaTarget.Connection, d = convertKafkaConnectionOptions(m.Connection)
	diags.Append(d...)
	settings.KafkaTarget.Auth, d = convertKafkaAuth(m.Auth)
	diags.Append(d...)
	settings.KafkaTarget.TopicSettings, d = convertKafkaTargetTopicSettings(m.TopicSettings)
	diags.Append(d...)
	settings.KafkaTarget.Serializer, d = convertSerializer(m.Serializer)
	diags.Append(d...)

	return settings, diags
}

func parseTransferEndpointKafkaAuth(e *endpoint.KafkaAuth, m *endpointKafkaAuth) {
	if e == nil {
		m = nil
	}

	if no_auth := e.GetNoAuth(); no_auth != nil {
		m = &endpointKafkaAuth{NoAuth: &endpointKafkAuthNoAuth{}}
	}

	if sasl := e.GetSasl(); sasl != nil {
		if m == nil {
			m = &endpointKafkaAuth{SASL: &endpointKafkaAuthSASL{}}
		}
		if !m.SASL.User.IsNull() {
			m.SASL.User = types.StringValue(sasl.User)
		}
		if !m.SASL.Mechanism.IsNull() {
			m.SASL.Mechanism = types.StringValue(sasl.Mechanism.String())
		}
	}
}

func parseTransferEndpointKafkaConnection(e *endpoint.KafkaConnectionOptions, m *endpointKafkaConnectionOptions) {
	if e == nil {
		m = nil
	}

	if cluster_id := e.GetClusterId(); cluster_id != "" {
		m.ClusterId = types.StringValue(cluster_id)
	}
	if on_premise := e.GetOnPremise(); on_premise != nil {
		if m == nil {
			m = &endpointKafkaConnectionOptions{}
		}
		if m.OnPremise.BrokerUrls != nil {
			m.OnPremise.BrokerUrls = convertSliceToTFStrings(on_premise.BrokerUrls)
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

func parseTransferEndpointKafkaSource(ctx context.Context, e *endpoint.KafkaSource, c *endpointKafkaSourceSettings) diag.Diagnostics {
	var diag diag.Diagnostics

	parseTransferEndpointKafkaAuth(e.Auth, c.Auth)
	parseTransferEndpointKafkaConnection(e.Connection, c.Connection)
	c.TopicName = types.StringValue(e.TopicName)

	return diag
}

func parseTransferEndpointKafkaTarget(ctx context.Context, e *endpoint.KafkaTarget, c *endpointKafkaTargetSettings) diag.Diagnostics {
	var diag diag.Diagnostics

	parseTransferEndpointKafkaAuth(e.Auth, c.Auth)
	parseTransferEndpointKafkaConnection(e.Connection, c.Connection)

	if e.Serializer == nil {
		c.Serializer = nil
	} else {
		if auto := e.Serializer.GetSerializerAuto(); auto != nil {
			c.Serializer.Auto = &endpointSerializerAuto{}
		}
		if json := e.Serializer.GetSerializerJson(); json != nil {
			c.Serializer.JSON = &endpointSerializerJSON{}
		}
		if debezium := e.Serializer.GetSerializerDebezium(); debezium != nil {
			c.Serializer.Debezium = &endpointSerializerDebezium{}
			if len(debezium.SerializerParameters) != 0 {
				p := make([]endpointSerializerDebeziumParameter, len(debezium.SerializerParameters))
				for i := 0; i < len(debezium.SerializerParameters); i++ {
					p[i] = endpointSerializerDebeziumParameter{
						Key:   types.StringValue(debezium.SerializerParameters[i].Key),
						Value: types.StringValue(debezium.SerializerParameters[i].Value),
					}
				}
			}
		}
	}

	if e.TopicSettings == nil {
		c.TopicSettings = nil
	} else {
		if prefix := e.TopicSettings.GetTopicPrefix(); prefix != "" {
			c.TopicSettings.TopicPrefix = types.StringValue(prefix)
		}
		if topic := e.TopicSettings.GetTopic(); topic != nil {
			if c.TopicSettings.Topic == nil {
				c.TopicSettings.Topic = &endpointKafkaTargetTopic{}
			}
			if topic.TopicName != "" {
				c.TopicSettings.Topic.TopicName = types.StringValue(topic.TopicName)
			}
			if topic.SaveTxOrder {
				c.TopicSettings.Topic.SaveTxOrder = types.BoolValue(topic.SaveTxOrder)
			}
		}
	}

	return diag
}
