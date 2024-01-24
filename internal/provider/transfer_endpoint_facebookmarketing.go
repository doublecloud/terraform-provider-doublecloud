package provider

import (
	"fmt"
	"strings"

	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transferEndpointFacebookMarketingSourceSettings struct {
	StartDate            types.String                                      `tfsdk:"start_date"`
	AccountId            types.String                                      `tfsdk:"account_id"`
	EndDate              types.String                                      `tfsdk:"end_date"`
	AccessToken          types.String                                      `tfsdk:"access_token"`
	IncludeDeleted       types.Bool                                        `tfsdk:"include_deleted"`
	FetchThumbnailImages types.Bool                                        `tfsdk:"fetch_thumbnail_images"`
	CustomInsights       []*transferEndpointFacebookMarketingSourceInsight `tfsdk:"custom_insights"`
}

func transferEndpointFacebookMarketingSourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The date from which to replicate data for all incremental streams, in the format `YYYY-MM-DDT00:00:00Z`. All data generated after this date and before `end_date` (if set) will be replicated. Example: `2017-01-25T00:00:00Z`",
			},
			"account_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Facebook Ad account ID to use when pulling data from the Facebook Marketing API. Example: `111111111111111`",
			},
			"end_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The date until which you'd like to replicate data for all incremental streams, in the format `YYYY-MM-DDT00:00:00Z`. All data generated between `start_date` and this date will be replicated. Not setting this option will result in always syncing the latest data. Example: `2017-01-25T23:59:59Z`",
			},
			"access_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The value of the access token. See  [documentation](https://docs.airbyte.io/integrations/sources/facebook-marketing) for more information on the meaning of this token and how to obtain it",
			},
			"include_deleted": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Include data from deleted Campaigns, Ads, and AdSets",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"fetch_thumbnail_images": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "In each Ad Creative, fetch the `thumbnail_url` and store the result in `thumbnail_data_url`",
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"custom_insights": schema.ListNestedAttribute{
				NestedObject:        transferEndpointFacebookMarketingSourceInsightSchema(),
				Optional:            true,
				MarkdownDescription: "Insights. Each entry must have a name and can contains `fields`, `breakdowns`, or `action_breakdowns`",
			},
		},
	}
}

func (m *transferEndpointFacebookMarketingSourceSettings) parse(e *endpoint_airbyte.FacebookMarketingSource) diag.Diagnostics {
	var diags diag.Diagnostics

	m.StartDate = types.StringValue(e.GetStartDate())
	m.AccountId = types.StringValue(e.GetAccountId())
	m.EndDate = types.StringValue(e.GetEndDate())
	if tkn := e.GetAccessToken(); len(tkn) > 0 {
		m.AccessToken = types.StringValue(e.GetAccessToken())
	}
	m.IncludeDeleted = types.BoolValue(e.GetIncludeDeleted())
	m.FetchThumbnailImages = types.BoolValue(e.GetFetchThumbnailImages())
	if ins := e.GetCustomInsights(); len(ins) > 0 {
		for i := range ins {
			if i >= len(m.CustomInsights) {
				m.CustomInsights = append(m.CustomInsights, new(transferEndpointFacebookMarketingSourceInsight))
			}
			diags.Append(m.CustomInsights[i].parse(ins[i])...)
		}
		m.CustomInsights = m.CustomInsights[:len(ins)]
	} else {
		m.CustomInsights = nil
	}

	return diags
}

func (m *transferEndpointFacebookMarketingSourceSettings) convert(r *endpoint_airbyte.FacebookMarketingSource) diag.Diagnostics {
	var diags diag.Diagnostics

	r.StartDate = m.StartDate.ValueString()
	r.AccountId = m.AccountId.ValueString()
	r.EndDate = m.EndDate.ValueString()
	r.AccessToken = m.AccessToken.ValueString()
	r.IncludeDeleted = m.IncludeDeleted.ValueBool()
	r.FetchThumbnailImages = m.FetchThumbnailImages.ValueBool()
	if len(m.CustomInsights) > 0 {
		r.CustomInsights = make([]*endpoint_airbyte.FacebookMarketingSource_InsightConfig, len(m.CustomInsights))
		for i := 0; i < len(m.CustomInsights); i++ {
			r.CustomInsights[i] = new(endpoint_airbyte.FacebookMarketingSource_InsightConfig)
			diags.Append(m.CustomInsights[i].convert(r.CustomInsights[i])...)
		}
	}

	return diags
}

type transferEndpointFacebookMarketingSourceInsight struct {
	Name             types.String   `tfsdk:"name"`
	Fields           []types.String `tfsdk:"fields"`
	Breakdowns       []types.String `tfsdk:"breakdowns"`
	ActionBreakdowns []types.String `tfsdk:"action_breakdowns"`
}

func transferEndpointFacebookMarketingSourceInsightSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Insight name",
			},
			"fields": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "`fields` request parameter",
				Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(transferEndpointFacebookMarketingSourceInsightFieldOneofValues()...))},
			},
			"breakdowns": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "`breakdowns` request parameter",
				Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(transferEndpointFacebookMarketingSourceInsightBreakdownOneofValues()...))},
			},
			"action_breakdowns": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "`action_breakdowns` request parameter",
				Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(transferEndpointFacebookMarketingSourceInsightActionBreakdownOneofValues()...))},
			},
		},
	}
}

func transferEndpointFacebookMarketingSourceInsightFieldOneofValues() []string {
	result := make([]string, 0)
	for k, v := range endpoint_airbyte.FacebookMarketingSource_Field_value {
		if v == 0 {
			continue
		}
		result = append(
			result,
			strings.ToLower(k),
		)
	}
	return result
}

func transferEndpointFacebookMarketingSourceInsightBreakdownOneofValues() []string {
	result := make([]string, 0)
	for k, v := range endpoint_airbyte.FacebookMarketingSource_Breakdown_value {
		if v == 0 {
			continue
		}
		result = append(
			result,
			strings.ToLower(k),
		)
	}
	return result
}

func transferEndpointFacebookMarketingSourceInsightActionBreakdownOneofValues() []string {
	result := make([]string, 0)
	for k, v := range endpoint_airbyte.FacebookMarketingSource_ActionBreakdown_value {
		if v == 0 {
			continue
		}
		result = append(
			result,
			strings.ToLower(strings.TrimPrefix(k, "ACTION_")),
		)
	}
	return result
}

func (m *transferEndpointFacebookMarketingSourceInsight) parse(e *endpoint_airbyte.FacebookMarketingSource_InsightConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = types.StringValue(e.GetName())
	if flds := e.GetFields(); len(flds) > 0 {
		m.Fields = make([]types.String, 0)
		for i := range flds {
			m.Fields = append(m.Fields, types.StringValue(transferEndpointFacebookMarketingSourceInsightFieldToString(flds[i])))
		}
	} else {
		m.Fields = nil
	}
	if bdns := e.GetBreakdowns(); len(bdns) > 0 {
		m.Breakdowns = make([]types.String, 0)
		for i := range bdns {
			m.Breakdowns = append(m.Breakdowns, types.StringValue(transferEndpointFacebookMarketingSourceInsightBreakdownToString(bdns[i])))
		}
	} else {
		m.Breakdowns = nil
	}
	if abdns := e.GetActionBreakdowns(); len(abdns) > 0 {
		m.ActionBreakdowns = make([]types.String, 0)
		for i := range abdns {
			m.ActionBreakdowns = append(m.ActionBreakdowns, types.StringValue(transferEndpointFacebookMarketingSourceInsightActionBreakdownToString(abdns[i])))
		}
	} else {
		m.ActionBreakdowns = nil
	}

	return diags
}

func transferEndpointFacebookMarketingSourceInsightFieldToString(v endpoint_airbyte.FacebookMarketingSource_Field) string {
	result := endpoint_airbyte.FacebookMarketingSource_Field_name[int32(v)]
	result = strings.ToLower(result)
	return result
}

func transferEndpointFacebookMarketingSourceInsightBreakdownToString(v endpoint_airbyte.FacebookMarketingSource_Breakdown) string {
	result := endpoint_airbyte.FacebookMarketingSource_Breakdown_name[int32(v)]
	result = strings.ToLower(result)
	return result
}

func transferEndpointFacebookMarketingSourceInsightActionBreakdownToString(v endpoint_airbyte.FacebookMarketingSource_ActionBreakdown) string {
	result := endpoint_airbyte.FacebookMarketingSource_ActionBreakdown_name[int32(v)]
	result = strings.TrimPrefix(result, "ACTION_")
	result = strings.ToLower(result)
	return result
}

func (m *transferEndpointFacebookMarketingSourceInsight) convert(e *endpoint_airbyte.FacebookMarketingSource_InsightConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	e.Name = m.Name.ValueString()
	if len(m.Fields) > 0 {
		e.Fields = make([]endpoint_airbyte.FacebookMarketingSource_Field, len(m.Fields))
		for i := range m.Fields {
			var d diag.Diagnostic
			e.Fields[i], d = transferEndpointFacebookMarketingSourceInsightFieldToEnum(m.Fields[i].ValueString())
			diags.Append(d)
		}
	}
	if len(m.Breakdowns) > 0 {
		e.Breakdowns = make([]endpoint_airbyte.FacebookMarketingSource_Breakdown, len(m.Breakdowns))
		for i := range m.Breakdowns {
			var d diag.Diagnostic
			e.Breakdowns[i], d = transferEndpointFacebookMarketingSourceInsightBreakdownToEnum(m.Breakdowns[i].ValueString())
			diags.Append(d)
		}
	}
	if len(m.ActionBreakdowns) > 0 {
		e.ActionBreakdowns = make([]endpoint_airbyte.FacebookMarketingSource_ActionBreakdown, len(m.ActionBreakdowns))
		for i := range m.ActionBreakdowns {
			var d diag.Diagnostic
			e.ActionBreakdowns[i], d = transferEndpointFacebookMarketingSourceInsightActionBreakdownToEnum(m.ActionBreakdowns[i].ValueString())
			diags.Append(d)
		}
	}

	return diags
}

func transferEndpointFacebookMarketingSourceInsightFieldToEnum(v string) (endpoint_airbyte.FacebookMarketingSource_Field, diag.Diagnostic) {
	key := strings.ToUpper(v)
	result, ok := endpoint_airbyte.FacebookMarketingSource_Field_value[key]
	if !ok {
		return endpoint_airbyte.FacebookMarketingSource_FIELD_UNSPECIFIED, diag.NewAttributeErrorDiagnostic(path.Root("fields"), "unknown Field enum value", fmt.Sprintf("%q (enum key %q)", v, key))
	}
	return endpoint_airbyte.FacebookMarketingSource_Field(result), nil
}

func transferEndpointFacebookMarketingSourceInsightBreakdownToEnum(v string) (endpoint_airbyte.FacebookMarketingSource_Breakdown, diag.Diagnostic) {
	key := strings.ToUpper(v)
	result, ok := endpoint_airbyte.FacebookMarketingSource_Breakdown_value[key]
	if !ok {
		return endpoint_airbyte.FacebookMarketingSource_BREAKDOWN_UNSPECIFIED, diag.NewAttributeErrorDiagnostic(path.Root("breakdowns"), "unknown Breakdown enum value", fmt.Sprintf("%q (enum key %q)", v, key))
	}
	return endpoint_airbyte.FacebookMarketingSource_Breakdown(result), nil
}

func transferEndpointFacebookMarketingSourceInsightActionBreakdownToEnum(v string) (endpoint_airbyte.FacebookMarketingSource_ActionBreakdown, diag.Diagnostic) {
	key := "ACTION_" + strings.ToUpper(v)
	result, ok := endpoint_airbyte.FacebookMarketingSource_ActionBreakdown_value[key]
	if !ok {
		return endpoint_airbyte.FacebookMarketingSource_ACTION_BREAKDOWN_UNSPECIFIED, diag.NewAttributeErrorDiagnostic(path.Root("action_breakdowns"), "unknown ActionBreakdown enum value", fmt.Sprintf("%q (enum key %q)", v, key))
	}
	return endpoint_airbyte.FacebookMarketingSource_ActionBreakdown(result), nil
}
