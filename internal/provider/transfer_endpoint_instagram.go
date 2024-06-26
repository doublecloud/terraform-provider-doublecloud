package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointInstagramSourceSettings struct {
	StartDate   types.String `tfsdk:"start_date"`
	AccessToken types.String `tfsdk:"access_token"`
}

func transferEndpointInstagramSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The date in format YYYY-MM-DDT00:00:00Z to start replicating data for User Insights. All data generated after this date will be replicated.",
			},
			"access_token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The value of the access token generated. See [Airbyte documentation](https://docs.airbyte.io/integrations/sources/instagram) for more information",
			},
		},
	}
}

func (i *endpointInstagramSourceSettings) convert() (*transfer.EndpointSettings_InstagramSource, diag.Diagnostics) {
	res := endpoint_airbyte.InstagramSource{}
	res.StartDate = i.StartDate.ValueString()
	res.AccessToken = i.AccessToken.ValueString()

	return &transfer.EndpointSettings_InstagramSource{InstagramSource: &res}, nil
}

func (i *endpointInstagramSourceSettings) parse(e *endpoint_airbyte.InstagramSource) diag.Diagnostics {
	i.StartDate = types.StringValue(e.GetStartDate())

	return nil
}
