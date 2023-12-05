package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointKafkaSourceSettings struct {
	Connection *endpointKafkaConnectionOptions `tfsdk:"connection"`
	Auth       *endpointKafkaAuth              `tfsdk:"auth"`
	TopicName  types.String                    `tfsdk:"topic_name"`
	Parser     *endpointKafkaParser            `tfsdk:"parser"`
}

func (m *endpointKafkaSourceSettings) parse(e *endpoint.KafkaSource) diag.Diagnostics {
	var diags diag.Diagnostics

	if auth := e.GetAuth(); auth != nil {
		if m.Auth == nil {
			m.Auth = new(endpointKafkaAuth)
		}
		diags.Append(m.Auth.parse(auth)...)
	} else {
		m.Auth = nil
	}
	parseTransferEndpointKafkaConnection(e.Connection, m.Connection)
	if len(e.TopicNames) == 1 {
		m.TopicName = types.StringValue(e.TopicNames[0])
	} else {
		diags.AddError("Invalid topic name", "Should contain at exactly one topic for terraform managed endpoints")
	}

	if prsr := e.GetParser(); prsr != nil {
		if m.Parser == nil {
			m.Parser = new(endpointKafkaParser)
		}
		diags.Append(m.Parser.parse(prsr)...)
	} else {
		m.Parser = nil
	}

	return diags
}

type endpointKafkaConnectionOptions struct {
	ClusterId types.String            `tfsdk:"cluster_id"`
	OnPremise *endpointOnPremiseKafka `tfsdk:"on_premise"`
}

type endpointKafkaAuth struct {
	SASL   *endpointKafkaAuthSASL  `tfsdk:"sasl"`
	NoAuth *endpointKafkAuthNoAuth `tfsdk:"no_auth"`
}

func (m *endpointKafkaAuth) parse(e *endpoint.KafkaAuth) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetNoAuth() != nil:
		m.SASL = nil
		m.NoAuth = new(endpointKafkAuthNoAuth)
	case e.GetSasl() != nil:
		m.NoAuth = nil
		if m.SASL == nil {
			m.SASL = new(endpointKafkaAuthSASL)
		}
		diags.Append(m.SASL.parse(e.GetSasl())...)
	default:
		diags.Append(diag.NewErrorDiagnostic("unknown auth type", fmt.Sprintf("%v", e.GetSecurity())))
	}

	return diags
}

func (m *endpointKafkaAuth) convert(r *endpoint.KafkaAuth) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case m.NoAuth != nil:
		r.Security = &endpoint.KafkaAuth_NoAuth{NoAuth: new(endpoint.NoAuth)}
	case m.SASL != nil:
		sasl := new(endpoint.KafkaSaslSecurity)
		diags.Append(m.SASL.convert(sasl)...)
		r.Security = &endpoint.KafkaAuth_Sasl{Sasl: sasl}
	}

	return diags
}

type endpointKafkAuthNoAuth struct{}

type endpointKafkaAuthSASL struct {
	User      types.String `tfsdk:"user"`
	Password  types.String `tfsdk:"password"`
	Mechanism types.String `tfsdk:"mechanism"`
}

func (m *endpointKafkaAuthSASL) parse(e *endpoint.KafkaSaslSecurity) diag.Diagnostics {
	m.User = types.StringValue(e.GetUser())
	if e.GetMechanism() != endpoint.KafkaMechanism_KAFKA_MECHANISM_UNSPECIFIED {
		m.Mechanism = types.StringValue(e.GetMechanism().String())
	}

	return nil
}

func (m *endpointKafkaAuthSASL) convert(r *endpoint.KafkaSaslSecurity) diag.Diagnostics {
	r.User = m.User.ValueString()
	if len(m.Password.ValueString()) > 0 {
		r.Password = &endpoint.Secret{Value: &endpoint.Secret_Raw{Raw: m.Password.ValueString()}}
	}
	r.Mechanism = endpoint.KafkaMechanism(endpoint.KafkaMechanism_value[m.Mechanism.ValueString()])

	return nil
}

type endpointOnPremiseKafka struct {
	BrokerUrls []types.String   `tfsdk:"broker_urls"`
	TLSMode    *endpointTLSMode `tfsdk:"tls_mode"`
}

type endpointKafkaParser struct {
	JSON *transferParserGeneric `tfsdk:"json"`
	TSKV *transferParserGeneric `tfsdk:"tskv"`
}

func endpointKafkaParserSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"json": transferParserGenericSchema(),
			"tskv": transferParserGenericSchema(),
		},
	}
}

func (m *endpointKafkaParser) parse(e *endpoint.Parser) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetJsonParser() != nil:
		m.TSKV = nil
		if m.JSON == nil {
			m.JSON = new(transferParserGeneric)
		}
		diags.Append(m.JSON.parse(e.GetJsonParser())...)
	case e.GetTskvParser() != nil:
		m.JSON = nil
		if m.TSKV == nil {
			m.TSKV = new(transferParserGeneric)
		}
		diags.Append(m.TSKV.parse(e.GetTskvParser())...)
	default:
		diags.Append(diag.NewErrorDiagnostic("unknown parser type", fmt.Sprintf("%v", e.GetParser())))
	}

	return diags
}

func (m *endpointKafkaParser) convert(r *endpoint.Parser) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case m.JSON != nil:
		prsr := new(endpoint.GenericParserCommon)
		diags.Append(m.JSON.convert(prsr)...)
		r.Parser = &endpoint.Parser_JsonParser{JsonParser: prsr}
	case m.TSKV != nil:
		prsr := new(endpoint.GenericParserCommon)
		diags.Append(m.TSKV.convert(prsr)...)
		r.Parser = &endpoint.Parser_TskvParser{TskvParser: prsr}
	}

	return diags
}

type transferParserGeneric struct {
	Schema          *transferParserSchema `tfsdk:"schema"`
	NullKeysAllowed types.Bool            `tfsdk:"null_keys_allowed"`
	AddRestColumn   types.Bool            `tfsdk:"add_rest_column"`
}

func transferParserGenericSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"null_keys_allowed": schema.BoolAttribute{
				Optional: true,
			},
			"add_rest_column": schema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"schema": transferParserSchemaSchema(),
		},
	}
}

func (m *transferParserGeneric) parse(e *endpoint.GenericParserCommon) diag.Diagnostics {
	var diags diag.Diagnostics

	m.NullKeysAllowed = types.BoolValue(e.GetNullKeysAllowed())
	m.AddRestColumn = types.BoolValue(e.GetAddRestColumn())
	if sch := e.GetDataSchema(); sch != nil {
		if m.Schema == nil {
			m.Schema = new(transferParserSchema)
		}
		diags.Append(m.Schema.parse(sch)...)
	}

	return diags
}

func (m *transferParserGeneric) convert(r *endpoint.GenericParserCommon) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.Schema != nil {
		r.DataSchema = new(endpoint.DataSchema)
		diags.Append(m.Schema.convert(r.DataSchema)...)
	}
	r.NullKeysAllowed = m.NullKeysAllowed.ValueBool()
	r.AddRestColumn = m.AddRestColumn.ValueBool()

	return diags
}

type transferParserSchema struct {
	JSON   *transferParserSchemaJSON   `tfsdk:"json"`
	Fields *transferParserSchemaFields `tfsdk:"fields"`
}

func transferParserSchemaSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"json":   transferParserSchemaJSONSchema(),
			"fields": transferParserSchemaFieldsSchema(),
		},
	}
}

func (m *transferParserSchema) parse(e *endpoint.DataSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetFields() != nil:
		m.JSON = nil
		if m.Fields == nil {
			m.Fields = new(transferParserSchemaFields)
		}
		diags.Append(m.Fields.parse(e.GetFields())...)
	case e.GetJsonFields() != "":
		m.Fields = nil
		if m.JSON == nil {
			m.JSON = new(transferParserSchemaJSON)
		}
		diags.Append(m.JSON.parse(e.GetJsonFields())...)
	default:
		diags.Append(diag.NewErrorDiagnostic("unknown schema type", fmt.Sprintf("%v", e.GetSchema())))
	}

	return diags
}

func (m *transferParserSchema) convert(r *endpoint.DataSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case m.Fields != nil:
		fl := new(endpoint.FieldList)
		diags.Append(m.Fields.convert(fl)...)
		r.Schema = &endpoint.DataSchema_Fields{Fields: fl}
	case m.JSON != nil:
		jsn := new(string)
		diags.Append(m.JSON.convert(jsn)...)
		r.Schema = &endpoint.DataSchema_JsonFields{JsonFields: *jsn}
	}

	return diags
}

type transferParserSchemaJSON struct {
	Fields types.String `tfsdk:"fields"`
}

func transferParserSchemaJSONSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"fields": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (m *transferParserSchemaJSON) parse(json string) diag.Diagnostics {
	m.Fields = types.StringValue(json)
	return nil
}

func (m *transferParserSchemaJSON) convert(r *string) diag.Diagnostics {
	*r = m.Fields.ValueString()
	return nil
}

type transferParserSchemaFields struct {
	Fields []*transferParserSchemaFieldsField `tfsdk:"field"`
}

func transferParserSchemaFieldsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"field": schema.ListNestedBlock{
				NestedObject: transferParserSchemaFieldsFieldSchema(),
			},
		},
	}
}

func (m *transferParserSchemaFields) parse(e *endpoint.FieldList) diag.Diagnostics {
	var diags diag.Diagnostics

	flds := e.GetFields()
	if len(flds) == 0 {
		m.Fields = nil
		return nil
	}

	if len(m.Fields) == 0 {
		m.Fields = make([]*transferParserSchemaFieldsField, len(flds))
	}
	for i := range flds {
		if len(m.Fields) <= i {
			m.Fields = append(m.Fields, new(transferParserSchemaFieldsField))
		}
		if m.Fields[i] == nil {
			m.Fields[i] = new(transferParserSchemaFieldsField)
		}
		diags.Append(m.Fields[i].parse(flds[i])...)
	}

	return diags
}

func (m *transferParserSchemaFields) convert(r *endpoint.FieldList) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(m.Fields) == 0 {
		return nil
	}

	r.Fields = make([]*endpoint.ColSchema, len(m.Fields))
	for i := range m.Fields {
		r.Fields[i] = new(endpoint.ColSchema)
		diags.Append(m.Fields[i].convert(r.Fields[i])...)
	}

	return diags
}

type transferParserSchemaFieldsField struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"` // enum
	Key      types.Bool   `tfsdk:"key"`
	Required types.Bool   `tfsdk:"required"`
	Path     types.String `tfsdk:"path"`
}

func transferParserSchemaFieldsFieldSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
			"key": schema.BoolAttribute{
				Optional: true,
			},
			"required": schema.BoolAttribute{
				Optional: true,
			},
			"path": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (m *transferParserSchemaFieldsField) parse(e *endpoint.ColSchema) diag.Diagnostics {
	m.Name = types.StringValue(e.GetName())
	m.Type = types.StringValue(transferParserSchemaFieldsFieldTypeFromEnum(e.GetType()))
	m.Key = types.BoolValue(e.GetKey())
	m.Required = types.BoolValue(e.GetRequired())
	if pth := e.GetPath(); len(pth) > 0 {
		m.Path = types.StringValue(pth)
	} else {
		m.Path = types.StringNull()
	}
	return nil
}

func transferParserSchemaFieldsFieldTypeFromEnum(typ endpoint.ColumnType) string {
	result := endpoint.ColumnType_name[int32(typ)]
	result = strings.ToLower(result)
	return result
}

func (m *transferParserSchemaFieldsField) convert(r *endpoint.ColSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	r.Name = m.Name.ValueString()
	typ, err := transferParserSchemaFieldsFieldTypeToEnum(m.Type.ValueString())
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic(err.Error(), ""))
		return diags
	}
	r.Type = typ
	r.Key = m.Key.ValueBool()
	r.Required = m.Required.ValueBool()
	r.Path = m.Path.ValueString()

	return diags
}

func transferParserSchemaFieldsFieldTypeToEnum(typ string) (endpoint.ColumnType, error) {
	if len(typ) == 0 {
		return endpoint.ColumnType_COLUMN_TYPE_UNSPECIFIED, fmt.Errorf("column type must be set")
	}

	operationEnumString := strings.ToUpper(typ)

	typE, typValid := endpoint.ColumnType_value[operationEnumString]
	if !typValid {
		return endpoint.ColumnType_COLUMN_TYPE_UNSPECIFIED, fmt.Errorf("unknown column type %q", typ)
	}

	return endpoint.ColumnType(typE), nil
}

type endpointKafkaTargetSettings struct {
	Connection    *endpointKafkaConnectionOptions `tfsdk:"connection"`
	Auth          *endpointKafkaAuth              `tfsdk:"auth"`
	TopicSettings *endpointKafkaTopicSettings     `tfsdk:"topic_settings"`
	Serializer    *endpointSerializer             `tfsdk:"serializer"`
}

type endpointKafkaTopicSettings struct {
	Topic              *endpointKafkaTargetTopic        `tfsdk:"topic"`
	TopicPrefix        types.String                     `tfsdk:"topic_prefix"`
	TopicConfigEntries *[]endpointKafkaTopicConfigEntry `tfsdk:"topic_config_entries"`
}

type endpointKafkaTopicConfigEntry struct {
	ConfigName  types.String `tfsdk:"config_name"`
	ConfigValue types.String `tfsdk:"config_value"`
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

type (
	endpointSerializerAuto     struct{}
	endpointSerializerJSON     struct{}
	endpointSerializerDebezium struct {
		Parameter *[]endpointSerializerDebeziumParameter `tfsdk:"parameter"`
	}
)

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
					"user": schema.StringAttribute{
						Optional: true,
					},
					"password": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},
					"mechanism": schema.StringAttribute{
						Optional:      true,
						Validators:    []validator.String{transferEndpointKafkaMechanismValidator()},
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
			"parser":     endpointKafkaParserSchema(),
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
			"topic_config_entries": schema.ListNestedBlock{NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"config_name":  schema.StringAttribute{Required: true},
					"config_value": schema.StringAttribute{Required: true},
				},
			}},
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

func kafkaSourceEndpointSettings(m *endpointKafkaSourceSettings) (*transfer.EndpointSettings_KafkaSource, diag.Diagnostics) {
	settings := &transfer.EndpointSettings_KafkaSource{KafkaSource: &endpoint.KafkaSource{}}
	var diags diag.Diagnostics

	connection, d := convertKafkaConnectionOptions(m.Connection)
	diags.Append(d...)
	settings.KafkaSource.Connection = connection
	if m.Auth != nil {
		settings.KafkaSource.Auth = new(endpoint.KafkaAuth)
		diags.Append(m.Auth.convert(settings.KafkaSource.Auth)...)
	}
	settings.KafkaSource.TopicNames = []string{m.TopicName.ValueString()}
	if m.Parser != nil {
		settings.KafkaSource.Parser = new(endpoint.Parser)
		diags.Append(m.Parser.convert(settings.KafkaSource.Parser)...)
	}

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
	if m.TopicConfigEntries != nil {
		for _, entry := range *m.TopicConfigEntries {
			settings.TopicConfigEntries = append(settings.TopicConfigEntries, &endpoint.TopicConfigEntry{
				ConfigName:  entry.ConfigName.ValueString(),
				ConfigValue: entry.ConfigValue.ValueString(),
			})
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
				SerializerParameters: parameters,
			},
		}
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
	if m.Auth != nil {
		settings.KafkaTarget.Auth = new(endpoint.KafkaAuth)
		diags.Append(m.Auth.convert(settings.KafkaTarget.Auth)...)
	}
	settings.KafkaTarget.TopicSettings, d = convertKafkaTargetTopicSettings(m.TopicSettings)
	diags.Append(d...)
	settings.KafkaTarget.Serializer, d = convertSerializer(m.Serializer)
	diags.Append(d...)

	return settings, diags
}

func parseTransferEndpointKafkaConnection(e *endpoint.KafkaConnectionOptions, m *endpointKafkaConnectionOptions) {
	if e == nil {
		m = nil
		return
	}
	if m == nil {
		m = new(endpointKafkaConnectionOptions)
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

func parseTransferEndpointKafkaTarget(ctx context.Context, e *endpoint.KafkaTarget, c *endpointKafkaTargetSettings) diag.Diagnostics {
	var diags diag.Diagnostics

	parseTransferEndpointKafkaConnection(e.Connection, c.Connection)

	if auth := e.GetAuth(); auth != nil {
		if c.Auth == nil {
			c.Auth = new(endpointKafkaAuth)
		}
		diags.Append(c.Auth.parse(auth)...)
	}

	if e.Serializer == nil {
		c.Serializer = nil
	} else {
		if c.Serializer == nil {
			c.Serializer = new(endpointSerializer)
		}
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
				c.Serializer.Debezium.Parameter = &p
			}
		}
	}

	if e.TopicSettings == nil {
		c.TopicSettings = nil
	} else {
		if c.TopicSettings == nil {
			c.TopicSettings = new(endpointKafkaTopicSettings)
		}
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
		if len(e.TopicSettings.TopicConfigEntries) != 0 {
			p := make([]endpointKafkaTopicConfigEntry, len(e.TopicSettings.TopicConfigEntries))
			for i, entry := range e.TopicSettings.TopicConfigEntries {
				p[i] = endpointKafkaTopicConfigEntry{
					ConfigName:  types.StringValue(entry.ConfigName),
					ConfigValue: types.StringValue(entry.ConfigValue),
				}
			}
			c.TopicSettings.TopicConfigEntries = &p
		}
	}

	return diags
}
