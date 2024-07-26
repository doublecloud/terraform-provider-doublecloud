package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint"
	endpoint_airbyte "github.com/doublecloud/go-genproto/doublecloud/transfer/v1/endpoint/airbyte"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgentf "github.com/doublecloud/go-sdk/gen/transfer"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &TransferEndpointResource{}
	_ resource.ResourceWithImportState = &TransferEndpointResource{}
)

func NewTransferEndpointResource() resource.Resource {
	return &TransferEndpointResource{}
}

type TransferEndpointResource struct {
	sdk             *dcsdk.SDK
	endpointService *dcgentf.EndpointServiceClient
}

type TransferEndpointModel struct {
	Id          types.String      `tfsdk:"id"`
	ProjectID   types.String      `tfsdk:"project_id"`
	Name        types.String      `tfsdk:"name"`
	Description types.String      `tfsdk:"description"`
	Settings    *endpointSettings `tfsdk:"settings"`
}

type endpointSettings struct {
	ClickhouseSource        *endpointClickhouseSourceSettings                `tfsdk:"clickhouse_source"`
	KafkaSource             *endpointKafkaSourceSettings                     `tfsdk:"kafka_source"`
	KinesisSource           *endpointKinesisSourceSettings                   `tfsdk:"kinesis_source"`
	PostgresSource          *endpointPostgresSourceSettings                  `tfsdk:"postgres_source"`
	MetrikaSource           *endpointMetrikaSourceSettings                   `tfsdk:"metrika_source"`
	MysqlSource             *endpointMysqlSourceSettings                     `tfsdk:"mysql_source"`
	MongoSource             *endpointMongoSourceSettings                     `tfsdk:"mongo_source"`
	ObjectStorageSource     *endpointObjectStorageSourceSettings             `tfsdk:"object_storage_source"`
	S3Source                *endpointS3SourceSettings                        `tfsdk:"s3_source"`
	LinkedinAdsSource       *endpointLinkedinAdsSourceSettings               `tfsdk:"linkedinads_source"`
	AWSCloudTrailSource     *endpointAWSCloudTrailSourceSettings             `tfsdk:"aws_cloudtrail_source"`
	GoogleAdsSource         *transferEndpointGoogleAdsSourceSettings         `tfsdk:"googleads_source"`
	FacebookMarketingSource *transferEndpointFacebookMarketingSourceSettings `tfsdk:"facebookmarketing_source"`
	SnowflakeSource         *endpointSnowflakeSourceSettings                 `tfsdk:"snowflake_source"`
	JiraSource              *endpointJiraSourceSettings                      `tfsdk:"jira_source"`
	RedshiftSource          *endpointRedshiftSourceSettings                  `tfsdk:"redshift_source"`
	HubspotSource           *endpointHubspotSourceSettings                   `tfsdk:"hubspot_source"`
	BigquerySource          *endpointBigquerySourceSettings                  `tfsdk:"bigquery_source"`
	MssqlSource             *endpointMssqlSourceSettings                     `tfsdk:"mssql_source"`
	InstagramSource         *endpointInstagramSourceSettings                 `tfsdk:"instagram_source"`

	ClickhouseTarget    *endpointClickhouseTargetSettings    `tfsdk:"clickhouse_target"`
	KafkaTarget         *endpointKafkaTargetSettings         `tfsdk:"kafka_target"`
	PostgresTarget      *endpointPostgresTargetSettings      `tfsdk:"postgres_target"`
	MysqlTarget         *endpointMysqlTargetSettings         `tfsdk:"mysql_target"`
	MongoTarget         *endpointMongoTargetSettings         `tfsdk:"mongo_target"`
	ObjectStorageTarget *endpointObjectStorageTargetSettings `tfsdk:"object_storage_target"`
	BigqueryTarget      *endpointBigqueryTargetSettings      `tfsdk:"bigquery_target"`
}

func (r *TransferEndpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transfer_endpoint"
}

func (r *TransferEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Transfer endpoint resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Transfer endpoint ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Endpoint name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Endpoint description",
				Default:             stringdefault.StaticString(""),
			},
		},
		Blocks: map[string]schema.Block{
			"settings": schema.SingleNestedBlock{
				Description: "Settings",
				Blocks: map[string]schema.Block{
					"clickhouse_source":        transferEndpointChSourceSchema(),
					"redshift_source":          transferEndpointRedshiftSourceSchema(),
					"kafka_source":             transferEndpointKafkaSourceSchema(),
					"kinesis_source":           transferEndpointKinesisSourceSchema(),
					"postgres_source":          transferEndpointPostgresSourceSchema(),
					"mysql_source":             transferEndpointMysqlSourceSchema(),
					"mongo_source":             transferEndpointMongoSourceSchema(),
					"metrika_source":           transferEndpointMetrikaSourceSchema(),
					"object_storage_source":    transferEndpointObjectStorageSourceSchema(),
					"s3_source":                transferEndpointS3SourceSchema(),
					"linkedinads_source":       endpointLinkedinAdsSourceSettingsSchema(),
					"aws_cloudtrail_source":    endpointAWSCloudTrailSourceSettingsSchema(),
					"googleads_source":         transferEndpointGoogleAdsSourceSettingsSchema(),
					"facebookmarketing_source": transferEndpointFacebookMarketingSourceSettingsSchema(),
					"snowflake_source":         endpointSnowflakeSourceSettingsSchema(),
					"jira_source":              endpointJiraSourceSettingsSchema(),
					"hubspot_source":           transferEndpointHubspotSourceSettingsSchema(),
					"bigquery_source":          transferEndpointBigquerySourceSettingsSchema(),
					"mssql_source":             transferEndpointMssqlSourceSchema(),
					"instagram_source":         transferEndpointInstagramSourceSchema(),

					"clickhouse_target":     transferEndpointChTargetSchema(),
					"kafka_target":          transferEndpointKafkaTargetSchema(),
					"postgres_target":       transferEndpointPostgresTargetSchema(),
					"mysql_target":          transferEndpointMysqlTargetSchema(),
					"mongo_target":          transferEndpointMongoTargetSchema(),
					"object_storage_target": transferEndpointObjectStorageTargetSchema(),
					"bigquery_target":       transferEndpointBigqueryTargetSettingsSchema(),
				},
			},
		},
	}
}

func (r *TransferEndpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.sdk = sdk
	r.endpointService = r.sdk.Transfer().Endpoint()
}

func createEndpointRequest(m *TransferEndpointModel) (*transfer.CreateEndpointRequest, diag.Diagnostics) {
	var diag diag.Diagnostics

	rq := &transfer.CreateEndpointRequest{}
	rq.Name = m.Name.ValueString()
	rq.Description = m.Description.ValueString()
	rq.ProjectId = m.ProjectID.ValueString()

	settings, diag := transferEndpointSettings(m)
	rq.Settings = settings

	if diag.HasError() {
		return nil, diag
	}

	return rq, nil
}

func deleteEndpointRequest(m *TransferEndpointModel) (*transfer.DeleteEndpointRequest, diag.Diagnostics) {
	return &transfer.DeleteEndpointRequest{
		EndpointId: m.Id.ValueString(),
	}, nil
}

func (r *TransferEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TransferEndpointModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := createEndpointRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	net, err := r.endpointService.Create(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(net, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}

	data.Id = types.StringValue(op.ResourceId())
	// Update computed fields
	{
		rs, err := r.endpointService.Get(ctx, &transfer.GetEndpointRequest{EndpointId: data.Id.ValueString()})
		if err != nil {
			resp.Diagnostics.AddError("failed to get", err.Error())
			return
		}
		resp.Diagnostics.Append(data.parseTransferEndpoint(ctx, rs)...)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TransferEndpointModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := r.endpointService.Get(ctx, &transfer.GetEndpointRequest{EndpointId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	resp.Diagnostics.Append(data.parseTransferEndpoint(ctx, rs)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func updateEndpointRequest(m *TransferEndpointModel) (*transfer.UpdateEndpointRequest, diag.Diagnostics) {
	settings, diag := transferEndpointSettings(m)
	if diag.HasError() {
		return nil, diag
	}

	return &transfer.UpdateEndpointRequest{
		EndpointId:  m.Id.ValueString(),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Settings:    settings,
	}, nil
}

func (r *TransferEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TransferEndpointModel

	var diag diag.Diagnostics

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	rq, diag := updateEndpointRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	net, err := r.endpointService.Update(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(net, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TransferEndpointModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := deleteEndpointRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.endpointService.Delete(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted endpoint: %s", data.Id))
}

func (r *TransferEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func transferEndpointSettings(m *TransferEndpointModel) (*transfer.EndpointSettings, diag.Diagnostics) {
	var diag diag.Diagnostics

	if m.Settings == nil {
		diag.AddError("unknown settings", "specify settings block for transfer_endpoint")
		return nil, diag
	}

	if m.Settings.ClickhouseSource != nil {
		s, d := chSourceEndpointSettings(m.Settings.ClickhouseSource)
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.KafkaSource != nil {
		s, d := kafkaSourceEndpointSettings(m.Settings.KafkaSource)
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.PostgresSource != nil {
		s, d := postgresSourceEndpointSettings(m.Settings.PostgresSource)
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.MetrikaSource != nil {
		s, d := m.Settings.MetrikaSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.MysqlSource != nil {
		s, d := m.Settings.MysqlSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.MongoSource != nil {
		s, d := m.Settings.MongoSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.ObjectStorageSource != nil {
		s, d := m.Settings.ObjectStorageSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.S3Source != nil {
		s, d := m.Settings.S3Source.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.LinkedinAdsSource != nil {
		result := new(endpoint_airbyte.LinkedinAdsSource)
		diag.Append(m.Settings.LinkedinAdsSource.convert(result)...)
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_LinkedinAdsSource{LinkedinAdsSource: result},
		}, diag
	}
	if m.Settings.AWSCloudTrailSource != nil {
		result := new(endpoint_airbyte.AWSCloudTrailSource)
		diag.Append(m.Settings.AWSCloudTrailSource.convert(result)...)
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_AwsCloudtrailSource{AwsCloudtrailSource: result},
		}, diag
	}
	if m.Settings.GoogleAdsSource != nil {
		result := new(endpoint_airbyte.GoogleAdsSource)
		diag.Append(m.Settings.GoogleAdsSource.convert(result)...)
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_GoogleAdsSource{GoogleAdsSource: result},
		}, diag
	}
	if m.Settings.FacebookMarketingSource != nil {
		result := new(endpoint_airbyte.FacebookMarketingSource)
		diag.Append(m.Settings.FacebookMarketingSource.convert(result)...)
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_FacebookMarketingSource{FacebookMarketingSource: result},
		}, diag
	}
	if m.Settings.SnowflakeSource != nil {
		result := new(endpoint_airbyte.SnowflakeSource)
		diag.Append(m.Settings.SnowflakeSource.convert(result)...)
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_SnowflakeSource{SnowflakeSource: result},
		}, diag
	}
	if m.Settings.JiraSource != nil {
		result := new(endpoint_airbyte.JiraSource)
		diag.Append(m.Settings.JiraSource.convert(result)...)
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_JiraSource{JiraSource: result},
		}, diag
	}
	if m.Settings.RedshiftSource != nil {
		s, d := m.Settings.RedshiftSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.HubspotSource != nil {
		s, d := m.Settings.HubspotSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_HubspotSource{HubspotSource: s},
		}, diag
	}
	if m.Settings.BigquerySource != nil {
		result := new(endpoint_airbyte.BigQuerySource)
		diag.Append(m.Settings.BigquerySource.convert(result)...)
		return &transfer.EndpointSettings{
			Settings: &transfer.EndpointSettings_BigQuerySource{BigQuerySource: result},
		}, diag
	}
	if m.Settings.MssqlSource != nil {
		s, d := m.Settings.MssqlSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.InstagramSource != nil {
		s, d := m.Settings.InstagramSource.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}

	if m.Settings.ClickhouseTarget != nil {
		s, d := chTargetEndpointSettings(m.Settings.ClickhouseTarget)
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.KafkaTarget != nil {
		s, d := kafkaTargetEndpointSettings(m.Settings.KafkaTarget)
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag

	}
	if m.Settings.PostgresTarget != nil {
		s, d := postgresTargetEndpointSettings(m.Settings.PostgresTarget)
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.MysqlTarget != nil {
		s, d := m.Settings.MysqlTarget.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.MongoTarget != nil {
		s, d := m.Settings.MongoTarget.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.ObjectStorageTarget != nil {
		s, d := m.Settings.ObjectStorageTarget.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.BigqueryTarget != nil {
		s, d := m.Settings.BigqueryTarget.convert()
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}
	if m.Settings.KinesisSource != nil {
		s, d := kinesisSourceEndpointSettings(m.Settings.KinesisSource)
		if d.HasError() {
			diag.Append(d...)
		}
		return &transfer.EndpointSettings{Settings: s}, diag
	}

	diag.AddError("unknown endpoint settings", "would you mind to specify one of endpoint settings")
	return nil, diag
}

func (data *TransferEndpointModel) parseTransferEndpoint(ctx context.Context, e *transfer.Endpoint) diag.Diagnostics {
	var diag diag.Diagnostics
	data.Name = types.StringValue(e.Name)
	data.ProjectID = types.StringValue(e.ProjectId)
	data.Description = types.StringValue(e.Description)
	if data.Settings == nil {
		data.Settings = new(endpointSettings)
	}

	if settings := e.Settings.GetClickhouseTarget(); settings != nil {

		// TODO: move parsing to clickhouse file
		if data.Settings.ClickhouseTarget == nil {
			data.Settings.ClickhouseTarget = &endpointClickhouseTargetSettings{}
		}
		data.Settings.ClickhouseTarget.ClickhouseClusterName = types.StringValue(settings.ClickhouseClusterName)
		data.Settings.ClickhouseTarget.ClickhouseCleanupPolicy = types.StringValue(endpoint.CleanupPolicy_name[int32(settings.CleanupPolicy.Number())])

		if settings.AltNames != nil {
			data.Settings.ClickhouseTarget.AltNames = make([]altName, len(settings.AltNames))
			for i := 0; i < len(settings.AltNames); i++ {
				data.Settings.ClickhouseTarget.AltNames[i] = altName{
					FromName: types.StringValue(settings.AltNames[i].FromName),
					ToName:   types.StringValue(settings.AltNames[i].ToName),
				}
			}
		}
		diag.Append(parseTransferEndpointClickhouseConnection(ctx, settings.Connection, data.Settings.ClickhouseTarget.Connection)...)

	}
	if settings := e.Settings.GetClickhouseSource(); settings != nil {
		if data.Settings.ClickhouseSource == nil {
			data.Settings.ClickhouseSource = &endpointClickhouseSourceSettings{}
		}
		diag.Append(parseTransferEndpointClickhouseConnection(ctx, settings.Connection, data.Settings.ClickhouseSource.Connection)...)
		data.Settings.ClickhouseSource.IncludeTables = convertSliceToTFStrings(settings.IncludeTables)
		data.Settings.ClickhouseSource.ExcludeTables = convertSliceToTFStrings(settings.ExcludeTables)
	}
	if settings := e.Settings.GetKafkaSource(); settings != nil {
		if data.Settings.KafkaSource == nil {
			data.Settings.KafkaSource = &endpointKafkaSourceSettings{}
		}
		diag.Append(data.Settings.KafkaSource.parse(settings)...)
	}
	if settings := e.Settings.GetKafkaTarget(); settings != nil {
		if data.Settings.KafkaTarget == nil {
			data.Settings.KafkaTarget = &endpointKafkaTargetSettings{}
		}
		diag.Append(parseTransferEndpointKafkaTarget(ctx, settings, data.Settings.KafkaTarget)...)
	}
	if settings := e.Settings.GetPostgresSource(); settings != nil {
		if data.Settings.PostgresSource == nil {
			data.Settings.PostgresSource = &endpointPostgresSourceSettings{}
		}
		diag.Append(parseTransferEndpointPostgresSource(ctx, settings, data.Settings.PostgresSource)...)
	}
	if settings := e.Settings.GetPostgresTarget(); settings != nil {
		if data.Settings.PostgresTarget == nil {
			data.Settings.PostgresTarget = &endpointPostgresTargetSettings{}
		}
		diag.Append(parseTransferEndpointPostgresTarget(ctx, settings, data.Settings.PostgresTarget)...)
	}
	if settings := e.Settings.GetMetricaSource(); settings != nil {
		if data.Settings.MetrikaSource == nil {
			data.Settings.MetrikaSource = &endpointMetrikaSourceSettings{}
		}
		diag.Append(data.Settings.MetrikaSource.parse(settings)...)
	}
	if settings := e.Settings.GetMysqlSource(); settings != nil {
		if data.Settings.MysqlSource == nil {
			data.Settings.MysqlSource = &endpointMysqlSourceSettings{}
		}
		diag.Append(data.Settings.MysqlSource.parse(settings)...)
	}
	if settings := e.Settings.GetMysqlTarget(); settings != nil {
		if data.Settings.MysqlTarget == nil {
			data.Settings.MysqlTarget = &endpointMysqlTargetSettings{}
		}
		diag.Append(data.Settings.MysqlTarget.parse(settings)...)
	}
	if settings := e.Settings.GetMongoSource(); settings != nil {
		if data.Settings.MongoSource == nil {
			data.Settings.MongoSource = &endpointMongoSourceSettings{}
		}
		diag.Append(data.Settings.MongoSource.parse(settings)...)
	}
	if settings := e.Settings.GetMongoTarget(); settings != nil {
		if data.Settings.MongoTarget == nil {
			data.Settings.MongoTarget = &endpointMongoTargetSettings{}
		}
		diag.Append(data.Settings.MongoTarget.parse(settings)...)
	}
	if settings := e.Settings.GetObjectStorageSource(); settings != nil {
		if data.Settings.ObjectStorageSource == nil {
			data.Settings.ObjectStorageSource = &endpointObjectStorageSourceSettings{}
		}
		diag.Append(data.Settings.ObjectStorageSource.parse(settings)...)
	}
	if settings := e.Settings.GetObjectStorageTarget(); settings != nil {
		if data.Settings.ObjectStorageTarget == nil {
			data.Settings.ObjectStorageTarget = &endpointObjectStorageTargetSettings{}
		}
		diag.Append(data.Settings.ObjectStorageTarget.parse(settings)...)
	}
	if settings := e.Settings.GetS3Source(); settings != nil {
		if data.Settings.S3Source == nil {
			data.Settings.S3Source = &endpointS3SourceSettings{}
		}
		diag.Append(data.Settings.S3Source.parse(settings)...)
	}
	if settings := e.Settings.GetLinkedinAdsSource(); settings != nil {
		if data.Settings.LinkedinAdsSource == nil {
			data.Settings.LinkedinAdsSource = &endpointLinkedinAdsSourceSettings{}
		}
		diag.Append(data.Settings.LinkedinAdsSource.parse(settings)...)
	}
	if settings := e.Settings.GetAwsCloudtrailSource(); settings != nil {
		if data.Settings.AWSCloudTrailSource == nil {
			data.Settings.AWSCloudTrailSource = &endpointAWSCloudTrailSourceSettings{}
		}
		diag.Append(data.Settings.AWSCloudTrailSource.parse(settings)...)
	}
	if settings := e.Settings.GetGoogleAdsSource(); settings != nil {
		if data.Settings.GoogleAdsSource == nil {
			data.Settings.GoogleAdsSource = &transferEndpointGoogleAdsSourceSettings{}
		}
		diag.Append(data.Settings.GoogleAdsSource.parse(settings)...)
	}
	if settings := e.Settings.GetFacebookMarketingSource(); settings != nil {
		if data.Settings.FacebookMarketingSource == nil {
			data.Settings.FacebookMarketingSource = &transferEndpointFacebookMarketingSourceSettings{}
		}
		diag.Append(data.Settings.FacebookMarketingSource.parse(settings)...)
	}
	if settings := e.Settings.GetSnowflakeSource(); settings != nil {
		if data.Settings.SnowflakeSource == nil {
			data.Settings.SnowflakeSource = &endpointSnowflakeSourceSettings{}
		}
		diag.Append(data.Settings.SnowflakeSource.parse(settings)...)
	}
	if settings := e.Settings.GetJiraSource(); settings != nil {
		if data.Settings.JiraSource == nil {
			data.Settings.JiraSource = &endpointJiraSourceSettings{}
		}
		diag.Append(data.Settings.JiraSource.parse(settings)...)
	}
	if settings := e.Settings.GetRedshiftSource(); settings != nil {
		if data.Settings.RedshiftSource == nil {
			data.Settings.RedshiftSource = &endpointRedshiftSourceSettings{}
		}
		diag.Append(data.Settings.RedshiftSource.parse(settings)...)
	}
	if settings := e.Settings.GetBigQuerySource(); settings != nil {
		if data.Settings.BigquerySource == nil {
			data.Settings.BigquerySource = &endpointBigquerySourceSettings{}
		}
		diag.Append(data.Settings.BigquerySource.parse(settings)...)
	}
	if settings := e.Settings.GetBigqueryTarget(); settings != nil {
		if data.Settings.BigqueryTarget == nil {
			data.Settings.BigqueryTarget = &endpointBigqueryTargetSettings{}
		}
		diag.Append(data.Settings.BigqueryTarget.parse(settings)...)
	}
	if settings := e.Settings.GetHubspotSource(); settings != nil {
		if data.Settings.HubspotSource == nil {
			data.Settings.HubspotSource = &endpointHubspotSourceSettings{}
		}
		diag.Append(data.Settings.HubspotSource.parse(settings)...)
	}
	if settings := e.Settings.GetMssqlSource(); settings != nil {
		if data.Settings.MssqlSource == nil {
			data.Settings.MssqlSource = &endpointMssqlSourceSettings{}
		}
		diag.Append(data.Settings.MssqlSource.parse(settings)...)
	}
	if settings := e.Settings.GetInstagramSource(); settings != nil {
		if data.Settings.InstagramSource == nil {
			data.Settings.InstagramSource = &endpointInstagramSourceSettings{}
		}
		diag.Append(data.Settings.InstagramSource.parse(settings)...)
	}
	if settings := e.Settings.GetKinesisSource(); settings != nil {
		if data.Settings.KinesisSource == nil {
			data.Settings.KinesisSource = &endpointKinesisSourceSettings{}
		}
		diag.Append(data.Settings.KinesisSource.parse(settings)...)
	}
	if data.Settings == nil {
		diag.AddError("failed to parse", "unknown settings type")
	}

	return diag
}
