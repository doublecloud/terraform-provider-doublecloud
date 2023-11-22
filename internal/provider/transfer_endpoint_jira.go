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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointJiraSourceSettings struct {
	ApiToken                  types.String   `tfsdk:"api_token"`
	Domain                    types.String   `tfsdk:"domain"`
	Email                     types.String   `tfsdk:"email"`
	Projects                  []types.String `tfsdk:"projects"`
	StartDate                 types.String   `tfsdk:"start_date"`
	IssuesStreamExpandWith    []types.String `tfsdk:"issues_stream_expand_with"`
	EnableExperimentalStreams types.Bool     `tfsdk:"enable_experimental_streams"`
}

func endpointJiraSourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"api_token":  schema.StringAttribute{Optional: true, Sensitive: true},
			"domain":     schema.StringAttribute{Optional: true},
			"email":      schema.StringAttribute{Optional: true},
			"projects":   schema.ListAttribute{ElementType: types.StringType, Optional: true},
			"start_date": schema.StringAttribute{Optional: true},
			"issues_stream_expand_with": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "`breakdowns` request parameter",
				Validators:          []validator.List{listvalidator.ValueStringsAre(stringvalidator.OneOf(transferEndpointJiraSourceIssuesStreamExpandWithOneofValues()...))},
			},
			"enable_experimental_streams": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		},
	}
}

func transferEndpointJiraSourceIssuesStreamExpandWithOneofValues() []string {
	result := make([]string, 0)
	for k, v := range endpoint_airbyte.JiraSource_Expand_value {
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

func (s *endpointJiraSourceSettings) parse(e *endpoint_airbyte.JiraSource) diag.Diagnostics {
	var diags diag.Diagnostics

	s.ApiToken = types.StringValue(e.GetApiToken())
	s.Domain = types.StringValue(e.GetDomain())
	s.Email = types.StringValue(e.GetEmail())
	s.Projects = toTypesStringSlice(e.GetProjects())
	s.StartDate = types.StringValue(e.GetStartDate())
	s.EnableExperimentalStreams = types.BoolValue(e.GetEnableExperimentalStreams())
	if expands := e.GetIssuesStreamExpandWith(); len(expands) > 0 {
		s.IssuesStreamExpandWith = make([]types.String, 0)
		for i := range expands {
			s.IssuesStreamExpandWith = append(s.IssuesStreamExpandWith, types.StringValue(transferEndpointJiraSourceIssuesStreamExpandWithToString(expands[i])))
		}
	} else {
		s.IssuesStreamExpandWith = nil
	}

	return diags
}

func transferEndpointJiraSourceIssuesStreamExpandWithToString(v endpoint_airbyte.JiraSource_Expand) string {
	result := endpoint_airbyte.JiraSource_Expand_name[int32(v)]
	result = strings.ToLower(result)
	return result
}

func transferEndpointJiraSourceIssuesStreamExpandWithToEnum(v string) (endpoint_airbyte.JiraSource_Expand, diag.Diagnostic) {
	key := strings.ToUpper(v)
	result, ok := endpoint_airbyte.JiraSource_Expand_value[key]
	if !ok {
		return endpoint_airbyte.JiraSource_EXPAND_UNSPECIFIED, diag.NewAttributeErrorDiagnostic(path.Root("fields"), "unknown Field enum value", fmt.Sprintf("%q (enum key %q)", v, key))
	}
	return endpoint_airbyte.JiraSource_Expand(result), nil
}

func toTypesStringSlice(values []string) []types.String {
	if len(values) == 0 {
		return nil
	}
	result := make([]types.String, len(values))
	for i, value := range values {
		result[i] = types.StringValue(value)
	}
	return result
}

func toStringSlice(values []types.String) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.String()
	}
	return result
}

func (s *endpointJiraSourceSettings) convert(r *endpoint_airbyte.JiraSource) diag.Diagnostics {
	var diags diag.Diagnostics

	r.ApiToken = s.ApiToken.ValueString()
	r.Domain = s.Domain.ValueString()
	r.Email = s.Email.ValueString()
	r.Projects = toStringSlice(s.Projects)
	r.StartDate = s.StartDate.ValueString()
	if len(s.IssuesStreamExpandWith) > 0 {
		r.IssuesStreamExpandWith = make([]endpoint_airbyte.JiraSource_Expand, len(s.IssuesStreamExpandWith))
		for i := range s.IssuesStreamExpandWith {
			var d diag.Diagnostic
			r.IssuesStreamExpandWith[i], d = transferEndpointJiraSourceIssuesStreamExpandWithToEnum(s.IssuesStreamExpandWith[i].ValueString())
			diags.Append(d)
		}
	}
	r.EnableExperimentalStreams = s.EnableExperimentalStreams.ValueBool()

	return diags
}
