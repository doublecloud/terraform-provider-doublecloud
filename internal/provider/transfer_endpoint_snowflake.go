package provider

import (
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointSnowflakeSourceSettings struct {
	Host          types.String                        `tfsdk:"host"`
	Role          types.String                        `tfsdk:"role"`
	Warehouse     types.String                        `tfsdk:"warehouse"`
	Database      types.String                        `tfsdk:"database"`
	Schema        types.String                        `tfsdk:"schema"`
	JDBCUrlParams types.String                        `tfsdk:"jdbc_url_params"`
	Credentials   *endpointSnowflakeSourceCredentials `tfsdk:"credentials"`
}

type endpointSnowflakeSourceCredentials struct {
	OAuth     *endpointSnowflakeSourceCredentialsOauth     `tfsdk:"oauth"`
	BasicAuth *endpointSnowflakeSourceCredentialsBasicAuth `tfsdk:"basic_auth"`
}

type endpointSnowflakeSourceCredentialsOauth struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	AccessToken  types.String `tfsdk:"access_token"`
	RefreshToken types.String `tfsdk:"refresh_token"`
}
type endpointSnowflakeSourceCredentialsBasicAuth struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func endpointSnowflakeSourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"host":            schema.StringAttribute{Optional: true},
			"role":            schema.StringAttribute{Optional: true},
			"warehouse":       schema.StringAttribute{Optional: true},
			"database":        schema.StringAttribute{Optional: true},
			"schema":          schema.StringAttribute{Optional: true},
			"jdbc_url_params": schema.StringAttribute{Optional: true},
		},
		Blocks: map[string]schema.Block{
			"credentials": transferEndpointSnowflakeSourceCredentialsSchema(),
		},
	}
}

func transferEndpointSnowflakeSourceCredentialsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"oauth":      transferEndpointSnowflakeSourceCredentialsOauthSchema(),
			"basic_auth": transferEndpointSnowflakeSourceCredentialsBasicAuthSchema(),
		},
	}
}

func transferEndpointSnowflakeSourceCredentialsBasicAuthSchema() schema.Block {
	return &schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{Optional: true},
			"password": schema.StringAttribute{Optional: true, Sensitive: true},
		},
	}
}

func transferEndpointSnowflakeSourceCredentialsOauthSchema() schema.Block {
	return &schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"client_id":     schema.StringAttribute{Optional: true, Sensitive: true},
			"client_secret": schema.StringAttribute{Optional: true, Sensitive: true},
			"access_token":  schema.StringAttribute{Optional: true, Sensitive: true},
			"refresh_token": schema.StringAttribute{Optional: true, Sensitive: true},
		},
	}
}

func (s *endpointSnowflakeSourceSettings) parse(e *endpoint_airbyte.SnowflakeSource) diag.Diagnostics {
	var diags diag.Diagnostics

	s.Host = types.StringValue(e.GetHost())
	s.Role = types.StringValue(e.GetRole())
	s.Warehouse = types.StringValue(e.GetWarehouse())
	s.Database = types.StringValue(e.GetDatabase())
	s.Schema = types.StringValue(e.GetSchema())
	s.JDBCUrlParams = types.StringValue(e.GetJdbcUrlParams())

	if e.GetCredentials() != nil {
		if s.Credentials == nil {
			s.Credentials = new(endpointSnowflakeSourceCredentials)
		}
		diags.Append(s.Credentials.parse(e.GetCredentials())...)
	}

	return diags
}

func (s *endpointSnowflakeSourceSettings) convert(r *endpoint_airbyte.SnowflakeSource) diag.Diagnostics {
	var diags diag.Diagnostics

	r.Host = s.Host.ValueString()
	r.Role = s.Role.ValueString()
	r.Warehouse = s.Warehouse.ValueString()
	r.Database = s.Database.ValueString()
	r.Schema = s.Schema.ValueString()
	r.JdbcUrlParams = s.JDBCUrlParams.ValueString()

	if s.Credentials != nil {
		r.Credentials = new(endpoint_airbyte.SnowflakeSource_Credentials)
		diags.Append(s.Credentials.convert(r.Credentials)...)
	}

	return diags
}

func (c *endpointSnowflakeSourceCredentials) parse(e *endpoint_airbyte.SnowflakeSource_Credentials) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetOauth() != nil:
		if c.OAuth == nil {
			c.OAuth = new(endpointSnowflakeSourceCredentialsOauth)
		}
		diags.Append(c.OAuth.parse(e.GetOauth())...)
	case e.GetBasicAuth() != nil:
		if c.BasicAuth == nil {
			c.BasicAuth = new(endpointSnowflakeSourceCredentialsBasicAuth)
		}
		diags.Append(c.BasicAuth.parse(e.GetBasicAuth())...)
	}

	return diags
}

func (c *endpointSnowflakeSourceCredentials) convert(r *endpoint_airbyte.SnowflakeSource_Credentials) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case c.OAuth != nil:
		credentials := new(endpoint_airbyte.SnowflakeSource_Credentials_Oauth)
		diags.Append(c.OAuth.convert(credentials)...)
		r.Credentials = credentials
	case c.BasicAuth != nil:
		credentials := new(endpoint_airbyte.SnowflakeSource_Credentials_BasicAuth_)
		diags.Append(c.BasicAuth.convert(credentials)...)
		r.Credentials = credentials
	}

	return diags
}

func (o *endpointSnowflakeSourceCredentialsOauth) parse(e *endpoint_airbyte.SnowflakeSource_Credentials_OAuth) diag.Diagnostics {
	if len(e.GetClientId()) > 0 {
		o.ClientID = types.StringValue(e.GetClientId())
	}
	if len(e.GetClientSecret()) > 0 {
		o.ClientSecret = types.StringValue(e.GetClientSecret())
	}
	if len(e.GetAccessToken()) > 0 {
		o.AccessToken = types.StringValue(e.GetAccessToken())
	}
	if len(e.GetRefreshToken()) > 0 {
		o.RefreshToken = types.StringValue(e.GetRefreshToken())
	}
	return nil
}

func (o *endpointSnowflakeSourceCredentialsOauth) convert(r *endpoint_airbyte.SnowflakeSource_Credentials_Oauth) diag.Diagnostics {
	r.Oauth = &endpoint_airbyte.SnowflakeSource_Credentials_OAuth{
		ClientId:     o.ClientID.ValueString(),
		ClientSecret: o.ClientSecret.ValueString(),
		AccessToken:  o.AccessToken.ValueString(),
		RefreshToken: o.RefreshToken.ValueString(),
	}
	return nil
}

func (b *endpointSnowflakeSourceCredentialsBasicAuth) parse(e *endpoint_airbyte.SnowflakeSource_Credentials_BasicAuth) diag.Diagnostics {
	if len(e.GetUsername()) > 0 {
		b.Username = types.StringValue(e.GetUsername())
	}
	if len(e.GetPassword()) > 0 {
		b.Password = types.StringValue(e.GetPassword())
	}
	return nil
}

func (b *endpointSnowflakeSourceCredentialsBasicAuth) convert(r *endpoint_airbyte.SnowflakeSource_Credentials_BasicAuth_) diag.Diagnostics {
	r.BasicAuth = &endpoint_airbyte.SnowflakeSource_Credentials_BasicAuth{
		Username: b.Username.ValueString(),
		Password: b.Password.ValueString(),
	}
	return nil
}
