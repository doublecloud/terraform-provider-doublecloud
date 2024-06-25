package provider

import (
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointHubspotSourceSettings struct {
	Credentials               *endpointHubspotSourceCredentials `tfsdk:"credentials"`
	StartDate                 types.String                      `tfsdk:"start_date"`
	EnableExperimentalStreams types.Bool                        `tfsdk:"enable_experimental_streams"`
}

type endpointHubspotSourceCredentials struct {
	PrivateApp *endpointHubspotSourceCredentialsPrivateApp `tfsdk:"private_app"`
}

type endpointHubspotSourceCredentialsPrivateApp struct {
	AccessToken types.String `tfsdk:"access_token"`
}

func transferEndpointHubspotSourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated.",
			},
			"enable_experimental_streams": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "If enabled then experimental streams become available for sync.",
			},
		},
		Blocks: map[string]schema.Block{
			"credentials": transferEndpointHubspotSourceCredentialsSchema(),
		},
	}
}

func transferEndpointHubspotSourceCredentialsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Blocks: map[string]schema.Block{
			"private_app": transferEndpointHubspotSourceCredentialsPrivateAppSchema(),
		},
		MarkdownDescription: "Choose how to authenticate to HubSpot",
	}
}

func transferEndpointHubspotSourceCredentialsPrivateAppSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Access token",
			},
		},
	}
}

func (h *endpointHubspotSourceSettings) convert() (*endpoint_airbyte.HubspotSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	res := &endpoint_airbyte.HubspotSource{}
	res.StartDate = h.StartDate.ValueString()

	if !h.EnableExperimentalStreams.IsNull() {
		res.EnableExperimentalStreams = h.EnableExperimentalStreams.ValueBool()
	}

	if h.Credentials != nil {
		credentials, d := h.Credentials.convert()
		diags.Append(d...)
		res.Credentials = credentials
	}

	return res, diags
}

func (h *endpointHubspotSourceCredentials) convert() (*endpoint_airbyte.HubspotSource_Credentials, diag.Diagnostics) {
	var diags diag.Diagnostics
	res := &endpoint_airbyte.HubspotSource_Credentials{}
	switch {
	case h.PrivateApp != nil:
		privateApp, d := h.PrivateApp.convert()
		diags.Append(d...)
		res.AuthMethod = privateApp
	default:
		diags.AddError("Invalid Credentials", "No valid credentials provided. Please provide valid credentials.")
	}

	return res, diags
}

func (h *endpointHubspotSourceCredentialsPrivateApp) convert() (*endpoint_airbyte.HubspotSource_Credentials_PrivateApp_, diag.Diagnostics) {
	res := &endpoint_airbyte.HubspotSource_Credentials_PrivateApp_{
		PrivateApp: &endpoint_airbyte.HubspotSource_Credentials_PrivateApp{
			AccessToken: h.AccessToken.ValueString(),
		},
	}

	return res, nil
}

func (h *endpointHubspotSourceSettings) parse(e *endpoint_airbyte.HubspotSource) diag.Diagnostics {
	var diags diag.Diagnostics

	h.StartDate = types.StringValue(e.GetStartDate())
	h.EnableExperimentalStreams = types.BoolValue(e.GetEnableExperimentalStreams())

	if e.GetCredentials() != nil {
		if h.Credentials == nil {
			h.Credentials = new(endpointHubspotSourceCredentials)
		}
		diags.Append(h.Credentials.parse(e.GetCredentials())...)
	}

	return diags
}

func (h *endpointHubspotSourceCredentials) parse(e *endpoint_airbyte.HubspotSource_Credentials) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case e.GetPrivateApp() != nil:
		if h.PrivateApp == nil {
			h.PrivateApp = new(endpointHubspotSourceCredentialsPrivateApp)
		}
		diags.Append(h.PrivateApp.parse(e.GetPrivateApp())...)
	}

	return diags
}

func (h *endpointHubspotSourceCredentialsPrivateApp) parse(e *endpoint_airbyte.HubspotSource_Credentials_PrivateApp) diag.Diagnostics {
	if len(e.GetAccessToken()) > 0 {
		h.AccessToken = types.StringValue(e.GetAccessToken())
	}
	return nil
}
