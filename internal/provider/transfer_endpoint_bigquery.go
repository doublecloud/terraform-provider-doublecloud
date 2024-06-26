package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func transferEndpointBigquerySourceSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The GCP project ID for the project containing the target BigQuery dataset.",
			},
			"dataset_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The dataset ID to search for tables and views. If you are only loading data from one dataset, setting this option could result in much faster schema discovery.",
			},
			"credentials_json": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The contents of your Service Account Key JSON file. See the [documentation](https://docs.airbyte.io/integrations/sources/bigquery#setup-the-bigquery-source-in-airbyte) for more information on how to obtain this key.",
			},
		},
	}
}

func transferEndpointBigqueryTargetSettingsSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The GCP project ID for the project containing the target BigQuery dataset.",
			},
			"dataset_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The dataset ID to search for tables and views. If you are only loading data from one dataset, setting this option could result in much faster schema discovery.",
			},
			"credentials_json": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The contents of your Service Account Key JSON file. See the [documentation](https://docs.airbyte.io/integrations/sources/bigquery#setup-the-bigquery-source-in-airbyte) for more information on how to obtain this key.",
			},
		},
	}
}

type endpointBigquerySourceSettings struct {
	ProjectID       types.String `tfsdk:"project_id"`
	DatasetID       types.String `tfsdk:"dataset_id"`
	CredentialsJSON types.String `tfsdk:"credentials_json"`
}

type endpointBigqueryTargetSettings struct {
	ProjectID       types.String `tfsdk:"project_id"`
	DatasetID       types.String `tfsdk:"dataset_id"`
	CredentialsJSON types.String `tfsdk:"credentials_json"`
}

func (b *endpointBigquerySourceSettings) convert(e *endpoint_airbyte.BigQuerySource) diag.Diagnostics {
	e.ProjectId = b.ProjectID.ValueString()
	e.DatasetId = b.DatasetID.ValueString()
	e.CredentialsJson = b.CredentialsJSON.ValueString()

	return nil
}

func (b *endpointBigquerySourceSettings) parse(e *endpoint_airbyte.BigQuerySource) diag.Diagnostics {
	b.ProjectID = types.StringValue(e.GetProjectId())
	b.DatasetID = types.StringValue(e.GetDatasetId())
	if cred := e.GetCredentialsJson(); len(cred) > 0 {
		b.CredentialsJSON = types.StringValue(cred)
	}

	return nil
}

func (b *endpointBigqueryTargetSettings) convert() (*transfer.EndpointSettings_BigqueryTarget, diag.Diagnostics) {
	res := endpoint.BigQueryTarget{}
	res.ProjectId = b.ProjectID.ValueString()
	res.DatasetId = b.DatasetID.ValueString()
	res.CredentialsJson = b.CredentialsJSON.ValueString()

	return &transfer.EndpointSettings_BigqueryTarget{BigqueryTarget: &res}, nil
}

func (b *endpointBigqueryTargetSettings) parse(e *endpoint.BigQueryTarget) diag.Diagnostics {
	b.ProjectID = types.StringValue(e.GetProjectId())
	b.DatasetID = types.StringValue(e.GetDatasetId())
	if cred := e.GetCredentialsJson(); len(cred) > 0 {
		b.CredentialsJSON = types.StringValue(cred)
	}

	return nil
}
