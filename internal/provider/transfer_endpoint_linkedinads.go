package provider

import (
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func endpointLinkedinAdsSourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "UTC date in the `YYYY-MM-DD` format. Any data before this date will not be replicated",
			},
			"account_ids": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Optional:            true,
				MarkdownDescription: "Space-separated account IDs to pull the data from. Leave empty if you want to pull data from all the associated accounts",
			},
		},
		Blocks: map[string]schema.Block{
			"credentials": endpointLinkedinAdsSourceSettingsCredentialsSchema(),
		},
	}
}

type endpointLinkedinAdsSourceSettings struct {
	StartDate   types.String                                  `tfsdk:"start_date"`
	AccountIds  []types.Int64                                 `tfsdk:"account_ids"`
	Credentials *endpointLinkedinAdsSourceSettingsCredentials `tfsdk:"credentials"`
}

func (s *endpointLinkedinAdsSourceSettings) parse(e *endpoint_airbyte.LinkedinAdsSource) diag.Diagnostics {
	var diags diag.Diagnostics

	s.StartDate = types.StringValue(e.StartDate)
	s.AccountIds = int64SliceValue(e.GetAccountIds())

	if e.GetCredentials() != nil {
		if s.Credentials == nil {
			s.Credentials = new(endpointLinkedinAdsSourceSettingsCredentials)
		}
		diags.Append(s.Credentials.parse(e.GetCredentials())...)
	}

	return diags
}

func int64SliceValue(values []int64) []types.Int64 {
	if len(values) == 0 {
		return nil
	}
	result := make([]types.Int64, len(values))
	for i, value := range values {
		result[i] = types.Int64Value(value)
	}
	return result
}

func (s *endpointLinkedinAdsSourceSettings) convert(r *endpoint_airbyte.LinkedinAdsSource) diag.Diagnostics {
	var diags diag.Diagnostics

	r.StartDate = s.StartDate.ValueString()
	r.AccountIds = int64ValueSlice(s.AccountIds)
	if s.Credentials != nil {
		r.Credentials = new(endpoint_airbyte.LinkedinAdsSource_Credentials)
		diags.Append(s.Credentials.convert(r.Credentials)...)
	}

	return diags
}

func int64ValueSlice(slice []types.Int64) []int64 {
	if len(slice) == 0 {
		return nil
	}
	result := make([]int64, len(slice))
	for i, v := range slice {
		result[i] = v.ValueInt64()
	}
	return result
}

func endpointLinkedinAdsSourceSettingsCredentialsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"oauth":        endpointLinkedinAdsSourceSettingsCredentialsOAuthSchema(),
			"access_token": endpointLinkedinAdsSourceSettingsCredentialsAccessTokenSchema(),
		},
		MarkdownDescription: "Authentication method",
	}
}

type endpointLinkedinAdsSourceSettingsCredentials struct {
	OAuth       *endpointLinkedinAdsSourceSettingsCredentialsOAuth       `tfsdk:"oauth"`
	AccessToken *endpointLinkedinAdsSourceSettingsCredentialsAccessToken `tfsdk:"access_token"`
}

func (c *endpointLinkedinAdsSourceSettingsCredentials) parse(e *endpoint_airbyte.LinkedinAdsSource_Credentials) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetOauth() != nil:
		if c.OAuth == nil {
			c.OAuth = new(endpointLinkedinAdsSourceSettingsCredentialsOAuth)
		}
		diags.Append(c.OAuth.parse(e.GetOauth())...)
	case len(e.GetAccessToken()) > 0:
		if c.AccessToken == nil {
			c.AccessToken = new(endpointLinkedinAdsSourceSettingsCredentialsAccessToken)
		}
		diags.Append(c.AccessToken.parse(e.GetAccessToken())...)
	}

	return diags
}

func (c *endpointLinkedinAdsSourceSettingsCredentials) convert(r *endpoint_airbyte.LinkedinAdsSource_Credentials) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case c.OAuth != nil:
		credentials := new(endpoint_airbyte.LinkedinAdsSource_Credentials_Oauth)
		diags.Append(c.OAuth.convert(credentials)...)
		r.Credentials = credentials
	case c.AccessToken != nil:
		credentials := new(endpoint_airbyte.LinkedinAdsSource_Credentials_AccessToken)
		diags.Append(c.AccessToken.convert(credentials)...)
		r.Credentials = credentials
	}

	return diags
}

func endpointLinkedinAdsSourceSettingsCredentialsOAuthSchema() schema.Block {
	return &schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Client ID of the LinkedIn Ads developer application",
			},
			"client_secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Client Secret for the LinkedIn Ads developer application",
			},
			"refresh_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Key to refresh the expired access token",
			},
		},
	}
}

type endpointLinkedinAdsSourceSettingsCredentialsOAuth struct {
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	RefreshToken types.String `tfsdk:"refresh_token"`
}

func (c *endpointLinkedinAdsSourceSettingsCredentialsOAuth) parse(e *endpoint_airbyte.LinkedinAdsSource_Credentials_OAuth) diag.Diagnostics {
	if len(e.GetClientId()) > 0 {
		c.ClientId = types.StringValue(e.GetClientId())
	}
	if len(e.GetClientSecret()) > 0 {
		c.ClientSecret = types.StringValue(e.GetClientSecret())
	}
	if len(e.GetRefreshToken()) > 0 {
		c.RefreshToken = types.StringValue(e.GetRefreshToken())
	}
	return nil
}

func (c *endpointLinkedinAdsSourceSettingsCredentialsOAuth) convert(r *endpoint_airbyte.LinkedinAdsSource_Credentials_Oauth) diag.Diagnostics {
	r.Oauth = &endpoint_airbyte.LinkedinAdsSource_Credentials_OAuth{
		ClientId:     c.ClientId.ValueString(),
		ClientSecret: c.ClientSecret.ValueString(),
		RefreshToken: c.RefreshToken.ValueString(),
	}
	return nil
}

func endpointLinkedinAdsSourceSettingsCredentialsAccessTokenSchema() schema.Block {
	return &schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Access token",
			},
		},
	}
}

type endpointLinkedinAdsSourceSettingsCredentialsAccessToken struct {
	AccessToken types.String `tfsdk:"access_token"`
}

func (c *endpointLinkedinAdsSourceSettingsCredentialsAccessToken) parse(accessToken string) diag.Diagnostics {
	if len(accessToken) > 0 {
		c.AccessToken = types.StringValue(accessToken)
	}
	return nil
}

func (c *endpointLinkedinAdsSourceSettingsCredentialsAccessToken) convert(r *endpoint_airbyte.LinkedinAdsSource_Credentials_AccessToken) diag.Diagnostics {
	r.AccessToken = c.AccessToken.ValueString()
	return nil
}
