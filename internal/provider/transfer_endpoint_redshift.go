package provider

import (
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointRedshiftSourceSettings struct {
	Host     types.String   `tfsdk:"host"`
	Port     types.Int64    `tfsdk:"port"`
	Database types.String   `tfsdk:"database"`
	Username types.String   `tfsdk:"username"`
	Password types.String   `tfsdk:"password"`
	Schemas  []types.String `tfsdk:"schemas"`
}

func transferEndpointRedshiftSourceSchema() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The hostname of the Redshift cluster.",
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The port number of the Redshift cluster.",
			},
			"database": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the database to connect to.",
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The username to use for connecting to the database.",
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The password to use for connecting to the database.",
			},
			"schemas": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "A list of schemas to include in the transfer.",
			},
		},
	}
}

func (m *endpointRedshiftSourceSettings) parse(e *endpoint_airbyte.RedshiftSource) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Host = types.StringValue(e.Host)
	m.Port = types.Int64Value(e.Port)
	m.Database = types.StringValue(e.Database)
	m.Username = types.StringValue(e.Username)

	if e.Schemas != nil {
		schemas := make([]types.String, len(e.Schemas))
		for i, schema := range e.Schemas {
			schemas[i] = types.StringValue(schema)
		}

		m.Schemas = schemas
	}

	return diags
}

func (m *endpointRedshiftSourceSettings) convert() (*transfer.EndpointSettings_RedshiftSource, diag.Diagnostics) {
	var diags diag.Diagnostics
	redshiftSource := endpoint_airbyte.RedshiftSource{Schemas: []string{}}

	if len(m.Schemas) > 0 {
		schemas := make([]string, len(m.Schemas))
		for i, schema := range m.Schemas {
			schemas[i] = schema.ValueString()
		}
		redshiftSource.Schemas = schemas
	}

	redshiftSource.Host = m.Host.ValueString()
	redshiftSource.Port = m.Port.ValueInt64()
	redshiftSource.Database = m.Database.ValueString()
	redshiftSource.Username = m.Username.ValueString()
	redshiftSource.Password = m.Password.ValueString()

	return &transfer.EndpointSettings_RedshiftSource{RedshiftSource: &redshiftSource}, diags
}
