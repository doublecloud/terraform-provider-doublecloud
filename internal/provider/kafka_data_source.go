package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/kafka/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgen "github.com/doublecloud/go-sdk/gen/kafka"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClickhouseDataSource{}

func NewKafkaDataSource() datasource.DataSource {
	return &KafkaDataSource{}
}

type KafkaDataSource struct {
	sdk *dcsdk.SDK
	svc *dcgen.ClusterServiceClient
}

type KafkaDataSourceModel struct {
	Id                    types.String         `tfsdk:"id"`
	ProjectID             types.String         `tfsdk:"project_id"`
	Name                  types.String         `tfsdk:"name"`
	Description           types.String         `tfsdk:"description"`
	RegionID              types.String         `tfsdk:"region_id"`
	CloudType             types.String         `tfsdk:"cloud_type"`
	Version               types.String         `tfsdk:"version"`
	ConnectionInfo        *KafkaConnectionInfo `tfsdk:"connection_info"`
	PrivateConnectionInfo *KafkaConnectionInfo `tfsdk:"private_connection_info"`
}

type KafkaConnectionInfo struct {
	ConnectionString types.String `tfsdk:"connection_string"`
	User             types.String `tfsdk:"user"`
	Password         types.String `tfsdk:"password"`
}

func (d *KafkaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka"
}

func kafkaConnectionInfoSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"connection_string": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "String to use in clients",
		},
		"user": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Apache Kafka® user",
		},
		"password": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Password for Apache Kafka® user",
		},
	}
}

func (d *KafkaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kafka data source",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project identifier",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster identifier",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of cluster",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of cluster",
			},
			"region_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Region of cluster",
			},
			"cloud_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cloud type (aws, gcp, azure)",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Version of ClickHouse DBMS",
			},
			"connection_info": schema.SingleNestedAttribute{
				Optional:   true,
				Attributes: kafkaConnectionInfoSchema(),
			},
			"private_connection_info": schema.SingleNestedAttribute{
				Optional:   true,
				Attributes: kafkaConnectionInfoSchema(),
			},
		},
	}
}

func (d *KafkaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	sdk, ok := req.ProviderData.(*dcsdk.SDK)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *dcsdk.SDK, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.sdk = sdk
	d.svc = d.sdk.Kafka().Cluster()
}

func (d *KafkaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KafkaDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id == types.StringNull() && data.Name == types.StringNull() {
		resp.Diagnostics.AddError("missing attribute", "specify one of: id or name")
		return
	}

	if data.Id == types.StringNull() {
		it := d.svc.ClusterIterator(ctx, &kafka.ListClustersRequest{ProjectId: data.ProjectID.ValueString()})
		for it.Next() {
			c := it.Value()
			if c.Name == data.Name.ValueString() {
				data.Id = types.StringValue(c.Id)
				break
			}
		}
		if it.Error() != nil {
			resp.Diagnostics.AddError("iterator has failed", it.Error().Error())
		}
		if data.Id == types.StringNull() {
			resp.Diagnostics.AddError("cluster not found", fmt.Sprintf("clickhouse cluster `%v` haven't found", data.Name.ValueString()))
			return
		}
	}

	response, err := d.svc.Get(ctx, &kafka.GetClusterRequest{ClusterId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	data.Description = types.StringValue(response.Description)
	data.CloudType = types.StringValue(response.CloudType)
	data.RegionID = types.StringValue(response.RegionId)
	data.Version = types.StringValue(response.Version)
	data.ConnectionInfo = &KafkaConnectionInfo{
		ConnectionString: types.StringValue(response.ConnectionInfo.ConnectionString),
		User:             types.StringValue(response.ConnectionInfo.User),
		Password:         types.StringValue(response.ConnectionInfo.Password),
	}
	data.PrivateConnectionInfo = &KafkaConnectionInfo{
		ConnectionString: types.StringValue(response.PrivateConnectionInfo.ConnectionString),
		User:             types.StringValue(response.PrivateConnectionInfo.User),
		Password:         types.StringValue(response.PrivateConnectionInfo.Password),
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
