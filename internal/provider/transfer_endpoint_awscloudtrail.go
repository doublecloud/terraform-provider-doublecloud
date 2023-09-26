package provider

import (
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func endpointAWSCloudTrailSourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"key_id": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "AWS CloudTrail Access Key ID. See [documentation](https://docs.airbyte.io/integrations/sources/aws-cloudtrail) for information on how to obtain this value.",
			},
			"secret_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "AWS CloudTrail Secret Key. See [documentation](https://docs.airbyte.io/integrations/sources/aws-cloudtrail) for information on how to obtain this value.",
			},
			"region_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The default AWS region; for example, `us-west-1`.",
			},
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The date from which replication should start. Note that in AWS CloudTrail, historical data are available for the last 90 days only. Format `YYYY-MM-DD`; for example, `2021-01-25`.",
			},
		},
	}
}

type endpointAWSCloudTrailSourceSettings struct {
	KeyID      types.String `tfsdk:"key_id"`
	SecretKey  types.String `tfsdk:"secret_key"`
	RegionName types.String `tfsdk:"region_name"`
	StartDate  types.String `tfsdk:"start_date"`
}

func (s *endpointAWSCloudTrailSourceSettings) parse(e *endpoint_airbyte.AWSCloudTrailSource) diag.Diagnostics {
	if sv := e.GetAwsKeyId(); len(sv) > 0 {
		s.KeyID = types.StringValue(sv)
	}
	if sv := e.GetAwsSecretKey(); len(sv) > 0 {
		s.SecretKey = types.StringValue(sv)
	}
	s.RegionName = types.StringValue(e.GetAwsRegionName())
	s.StartDate = types.StringValue(e.GetStartDate())

	return nil
}

func (s *endpointAWSCloudTrailSourceSettings) convert(r *endpoint_airbyte.AWSCloudTrailSource) diag.Diagnostics {
	r.AwsKeyId = s.KeyID.ValueString()
	r.AwsSecretKey = s.SecretKey.ValueString()
	r.AwsRegionName = s.RegionName.ValueString()
	r.StartDate = s.StartDate.ValueString()

	return nil
}
