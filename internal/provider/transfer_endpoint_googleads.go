package provider

import (
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transferEndpointGoogleAdsSourceSettings struct {
	Credentials          *transferEndpointGoogleAdsSourceCredentials   `tfsdk:"credentials"`
	CustomerID           types.String                                  `tfsdk:"customer_id"`
	StartDate            types.String                                  `tfsdk:"start_date"`
	EndDate              types.String                                  `tfsdk:"end_date"`
	CustomQueries        []*transferEndpointGoogleAdsSourceCustomQuery `tfsdk:"custom_queries"`
	LoginCustomerId      types.String                                  `tfsdk:"login_customer_id"`
	ConversionWindowDays types.Float64                                 `tfsdk:"conversion_window_days"`
}

func transferEndpointGoogleAdsSourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"customer_id": schema.StringAttribute{
				Optional: true,
			},
			"start_date": schema.StringAttribute{
				Optional: true,
			},
			"end_date": schema.StringAttribute{
				Optional: true,
			},
			"custom_queries": schema.ListNestedAttribute{
				NestedObject: transferEndpointGoogleAdsSourceCustomQuerySchema(),
				Optional:     true,
			},
			"login_customer_id": schema.StringAttribute{
				Optional: true,
			},
			"conversion_window_days": schema.Float64Attribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"credentials": transferEndpointGoogleAdsSourceCredentialsSchema(),
		},
	}
}

func (m *transferEndpointGoogleAdsSourceSettings) parse(e *endpoint_airbyte.GoogleAdsSource) diag.Diagnostics {
	var diags diag.Diagnostics

	if crds := e.GetCredentials(); crds != nil {
		if m.Credentials == nil {
			m.Credentials = new(transferEndpointGoogleAdsSourceCredentials)
		}
		diags.Append(m.Credentials.parse(crds)...)
	}
	m.CustomerID = types.StringValue(e.GetCustomerId())
	m.StartDate = types.StringValue(e.GetStartDate())
	m.EndDate = types.StringValue(e.GetEndDate())
	if cqs := e.GetCustomQueries(); len(cqs) > 0 {
		for i := 0; i < len(cqs); i++ {
			if i >= len(m.CustomQueries) {
				m.CustomQueries = append(m.CustomQueries, new(transferEndpointGoogleAdsSourceCustomQuery))
			}
			diags.Append(m.CustomQueries[i].parse(cqs[i])...)
		}
		m.CustomQueries = m.CustomQueries[:len(cqs)]
	} else {
		m.CustomQueries = nil
	}
	m.LoginCustomerId = types.StringValue(e.GetLoginCustomerId())
	m.ConversionWindowDays = types.Float64Value(e.GetConversionWindowDays())

	return diags
}

func (m *transferEndpointGoogleAdsSourceSettings) convert(r *endpoint_airbyte.GoogleAdsSource) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.Credentials != nil {
		r.Credentials = new(endpoint_airbyte.GoogleAdsSource_Credentials)
		diags.Append(m.Credentials.convert(r.Credentials)...)
	}
	r.CustomerId = m.CustomerID.ValueString()
	r.StartDate = m.StartDate.ValueString()
	r.EndDate = m.EndDate.ValueString()
	if len(m.CustomQueries) > 0 {
		r.CustomQueries = make([]*endpoint_airbyte.GoogleAdsSource_CustomQuery, len(m.CustomQueries))
		for i := 0; i < len(m.CustomQueries); i++ {
			r.CustomQueries[i] = new(endpoint_airbyte.GoogleAdsSource_CustomQuery)
			diags.Append(m.CustomQueries[i].convert(r.CustomQueries[i])...)
		}
	}
	r.LoginCustomerId = m.LoginCustomerId.ValueString()
	r.ConversionWindowDays = m.ConversionWindowDays.ValueFloat64()

	return diags
}

type transferEndpointGoogleAdsSourceCredentials struct {
	DeveloperToken types.String `tfsdk:"developer_token"`
	ClientId       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	AccessToken    types.String `tfsdk:"access_token"`
	RefreshToken   types.String `tfsdk:"refresh_token"`
}

func transferEndpointGoogleAdsSourceCredentialsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"developer_token": schema.StringAttribute{
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				Optional: true,
			},
			"access_token": schema.StringAttribute{
				Optional: true,
			},
			"refresh_token": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (m *transferEndpointGoogleAdsSourceCredentials) parse(e *endpoint_airbyte.GoogleAdsSource_Credentials) diag.Diagnostics {
	m.DeveloperToken = types.StringValue(e.GetDeveloperToken())
	m.ClientId = types.StringValue(e.GetClientId())
	m.ClientSecret = types.StringValue(e.GetClientSecret())
	m.AccessToken = types.StringValue(e.GetAccessToken())
	m.RefreshToken = types.StringValue(e.GetRefreshToken())

	return nil
}

func (m *transferEndpointGoogleAdsSourceCredentials) convert(r *endpoint_airbyte.GoogleAdsSource_Credentials) diag.Diagnostics {
	r.DeveloperToken = m.DeveloperToken.ValueString()
	r.ClientId = m.ClientId.ValueString()
	r.ClientSecret = m.ClientSecret.ValueString()
	r.AccessToken = m.AccessToken.ValueString()
	r.RefreshToken = m.RefreshToken.ValueString()

	return nil
}

type transferEndpointGoogleAdsSourceCustomQuery struct {
	Query     types.String `tfsdk:"query"`
	TableName types.String `tfsdk:"table_name"`
}

func transferEndpointGoogleAdsSourceCustomQuerySchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{
				Optional: true,
			},
			"table_name": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (m *transferEndpointGoogleAdsSourceCustomQuery) parse(e *endpoint_airbyte.GoogleAdsSource_CustomQuery) diag.Diagnostics {
	m.Query = types.StringValue(e.GetQuery())
	m.TableName = types.StringValue(e.GetTableName())

	return nil
}

func (m *transferEndpointGoogleAdsSourceCustomQuery) convert(r *endpoint_airbyte.GoogleAdsSource_CustomQuery) diag.Diagnostics {
	r.Query = m.Query.ValueString()
	r.TableName = m.TableName.ValueString()

	return nil
}
