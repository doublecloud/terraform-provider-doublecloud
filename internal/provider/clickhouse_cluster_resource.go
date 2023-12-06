package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgen "github.com/doublecloud/go-sdk/gen/clickhouse"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type clickhouseClusterModel struct {
	Id          types.String                `tfsdk:"id"`
	ProjectId   types.String                `tfsdk:"project_id"`
	CloudType   types.String                `tfsdk:"cloud_type"`
	RegionId    types.String                `tfsdk:"region_id"`
	Name        types.String                `tfsdk:"name"`
	Description types.String                `tfsdk:"description"`
	Version     types.String                `tfsdk:"version"`
	Resources   *clickhouseClusterResources `tfsdk:"resources"`

	Access    *AccessModel      `tfsdk:"access"`
	NetworkId types.String      `tfsdk:"network_id"`
	Config    *clickhouseConfig `tfsdk:"config"`

	// TODO: support mw
	// https://github.com/doublecloud/api/blob/main/doublecloud/v1/maintenance.proto
	// MaintenanceWindow *maintenanceWindow          `tfsdk:"maintenance_window"`
}

type clickhouseClusterResources struct {
	Clickhouse *clickhouseClusterResourcesClickhouse `tfsdk:"clickhouse"`
	Keeper     *clickhouseClusterResourcesKeeper     `tfsdk:"dedicated_keeper"`
}

func (m *clickhouseClusterResources) convert() (*clickhouse.ClusterResources, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := clickhouse.ClusterResources{}

	if v := m.Clickhouse; v != nil {
		r.Clickhouse = &clickhouse.ClusterResources_Clickhouse{}
		if v := m.Clickhouse.ResourcePresetId; !v.IsNull() {
			r.Clickhouse.ResourcePresetId = v.ValueString()
		} else {
			diags.AddError("missed resource_preset_id", "clickhouse resource_preset_id must be set")
		}
		if v := m.Clickhouse.DiskSize; !v.IsNull() {
			r.Clickhouse.DiskSize = wrapperspb.Int64(v.ValueInt64())
		} else {
			diags.AddError("missed disk_size", "clickhouse disk_size must be set")
		}
		if v := m.Clickhouse.MaxDiskSize; !v.IsNull() {
			r.Clickhouse.MaxDiskSize = wrapperspb.Int64(v.ValueInt64())
		}
		if v := m.Clickhouse.ReplicaCount; !v.IsNull() {
			r.Clickhouse.ReplicaCount = wrapperspb.Int64(v.ValueInt64())
		} else {
			diags.AddError("missed replica_count", "clickhouse replica count must be set")
		}
		if v := m.Clickhouse.ShardCount; !v.IsNull() {
			r.Clickhouse.ShardCount = wrapperspb.Int64(v.ValueInt64())
		}
	}
	if r.Clickhouse == nil {
		diags.AddError("missed block clickhouse", "specify clickhouse block in resources block")
	}
	if v := m.Keeper; v != nil {
		r.DedicatedKeeper = &clickhouse.ClusterResources_Keeper{}
		if v := m.Keeper.ResourcePresetId; !v.IsNull() {
			r.DedicatedKeeper.ResourcePresetId = v.ValueString()
		} else {
			diags.AddError("missed resource_preset_id", "keeper resource_preset_id must be set")
		}
		if v := m.Keeper.DiskSize; !v.IsNull() {
			r.DedicatedKeeper.DiskSize = wrapperspb.Int64(v.ValueInt64())
		} else {
			diags.AddError("missed disk_size", "keeper disk_size must be set")
		}
		if v := m.Keeper.MaxDiskSize; !v.IsNull() {
			r.DedicatedKeeper.MaxDiskSize = wrapperspb.Int64(v.ValueInt64())
		}
		if v := m.Keeper.ReplicaCount; !v.IsNull() {
			r.DedicatedKeeper.ReplicaCount = wrapperspb.Int64(v.ValueInt64())
		} else {
			diags.AddError("missed replica_count", "keeper replica count must be set")
		}
	}

	return &r, diags
}

type clickhouseClusterResourcesClickhouse struct {
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	MaxDiskSize      types.Int64  `tfsdk:"max_disk_size"`
	ReplicaCount     types.Int64  `tfsdk:"replica_count"`
	ShardCount       types.Int64  `tfsdk:"shard_count"`
}

type clickhouseClusterResourcesKeeper struct {
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	MaxDiskSize      types.Int64  `tfsdk:"max_disk_size"`
	ReplicaCount     types.Int64  `tfsdk:"replica_count"`
}

type clickhouseConfig struct {
	LogLevel                                  types.String  `tfsdk:"log_level"`
	MaxConnections                            types.Int64   `tfsdk:"max_connections"`
	MaxConcurrentQueries                      types.Int64   `tfsdk:"max_concurrent_queries"`
	KeepAliveTimeout                          types.String  `tfsdk:"keep_alive_timeout"`
	UncompressedCacheSize                     types.Int64   `tfsdk:"uncompressed_cache_size"`
	MarkCacheSize                             types.Int64   `tfsdk:"mark_cache_size"`
	MaxTableSizeToDrop                        types.Int64   `tfsdk:"max_table_size_to_drop"`
	MaxPartitionSizeToDrop                    types.Int64   `tfsdk:"max_partition_size_to_drop"`
	Timezone                                  types.String  `tfsdk:"timezone"`
	BackgroundPoolSize                        types.Int64   `tfsdk:"background_pool_size"`
	BackgroundSchedulePoolSize                types.Int64   `tfsdk:"background_schedule_pool_size"`
	BackgroundFetchesPoolSize                 types.Int64   `tfsdk:"background_fetches_pool_size"`
	BackgroundMovePoolSize                    types.Int64   `tfsdk:"background_move_pool_size"`
	BackgroundCommonPoolSize                  types.Int64   `tfsdk:"background_common_pool_size"`
	BackgroundMergesMutationsConcurrencyRatio types.Int64   `tfsdk:"background_merges_mutations_concurrency_ratio"`
	TotalMemoryProfilerStep                   types.Int64   `tfsdk:"total_memory_profiler_step"`
	TotalMemoryTrackerSampleProbability       types.Float64 `tfsdk:"total_memory_tracker_sample_probability"`
	BackgroundMessageBrokerSchedulePoolSize   types.Int64   `tfsdk:"background_message_broker_schedule_pool_size"`
	// MergeTree                                 *clickhouseConfigMergeTree    `tfsdk:"merge_tree"`
	// Compression                               []clickhouseConfigCompression `tfsdk:"compression"`
	Kafka *clickhouseConfigKafka `tfsdk:"kafka"`
	// KafkaTopics types.Map              `tfsdk:"kafka_topics"`
	// Rabbitmq                                  *clickhouseConfigRabbitmq     `tfsdk:"rabbitmq"`
	QueryLogRetentionSize              types.Int64  `tfsdk:"query_log_retention_size"`
	QueryLogRetentionTime              types.String `tfsdk:"query_log_retention_time"`
	QueryThreadLogEnabled              types.Bool   `tfsdk:"query_thread_log_enabled"`
	QueryThreadLogRetentionSize        types.Int64  `tfsdk:"query_thread_log_retention_size"`
	QueryThreadLogRetentionTime        types.String `tfsdk:"query_thread_log_retention_time"`
	QueryViewsLogEnabled               types.Bool   `tfsdk:"query_views_log_enabled"`
	QueryViewsLogRetentionSize         types.Int64  `tfsdk:"query_views_log_retention_size"`
	QueryViewsLogRetentionTime         types.String `tfsdk:"query_views_log_retention_time"`
	PartLogRetentionSize               types.Int64  `tfsdk:"part_log_retention_size"`
	PartLogRetentionTime               types.String `tfsdk:"part_log_retention_time"`
	MetricLogEnabled                   types.Bool   `tfsdk:"metric_log_enabled"`
	MetricLogRetentionSize             types.Int64  `tfsdk:"metric_log_retention_size"`
	MetricLogRetentionTime             types.String `tfsdk:"metric_log_retention_time"`
	AsynchronousMetricLogEnabled       types.Bool   `tfsdk:"asynchronous_metric_log_enabled"`
	AsynchronousMetricLogRetentionSize types.Int64  `tfsdk:"asynchronous_metric_log_retention_size"`
	AsynchronousMetricLogRetentionTime types.String `tfsdk:"asynchronous_metric_log_retention_time"`
	TraceLogEnabled                    types.Bool   `tfsdk:"trace_log_enabled"`
	TraceLogRetentionSize              types.Int64  `tfsdk:"trace_log_retention_size"`
	TraceLogRetentionTime              types.String `tfsdk:"trace_log_retention_time"`
	TextLogEnabled                     types.Bool   `tfsdk:"text_log_enabled"`
	TextLogRetentionSize               types.Int64  `tfsdk:"text_log_retention_size"`
	TextLogRetentionTime               types.String `tfsdk:"text_log_retention_time"`
	TextLogLevel                       types.String `tfsdk:"text_log_level"`
	OpentelemetrySpanLogEnabled        types.Bool   `tfsdk:"opentelemetry_span_log_enabled"`
	OpentelemetrySpanLogRetentionSize  types.Int64  `tfsdk:"opentelemetry_span_log_retention_size"`
	OpentelemetrySpanLogRetentionTime  types.String `tfsdk:"opentelemetry_span_log_retention_time"`
	SessionLogEnabled                  types.Bool   `tfsdk:"session_log_enabled"`
	SessionLogRetentionSize            types.Int64  `tfsdk:"session_log_retention_size"`
	SessionLogRetentionTime            types.String `tfsdk:"session_log_retention_time"`
	ZookeeperLogEnabled                types.Bool   `tfsdk:"zookeeper_log_enabled"`
	ZookeeperLogRetentionSize          types.Int64  `tfsdk:"zookeeper_log_retention_size"`
	ZookeeperLogRetentionTime          types.String `tfsdk:"zookeeper_log_retention_time"`
	AsynchronousInsertLogEnabled       types.Bool   `tfsdk:"asynchronous_insert_log_enabled"`
	AsynchronousInsertLogRetentionSize types.Int64  `tfsdk:"asynchronous_insert_log_retention_size"`
	AsynchronousInsertLogRetentionTime types.String `tfsdk:"asynchronous_insert_log_retention_time"`
	//     map<string,GraphiteRollup> graphite_rollup = 19;
}

//nolint:unused
type clickhouseConfigMergeTree struct {
	ReplicatedDeduplicationWindow                  types.Int64  `tfsdk:"replicated_deduplication_window"`
	ReplicatedDeduplicationWindowSeconds           types.String `tfsdk:"replicated_deduplication_window_seconds"`
	PartsToDelayInsert                             types.Int64  `tfsdk:"parts_to_delay_insert"`
	PartsToThrowInsert                             types.Int64  `tfsdk:"parts_to_throw_insert"`
	InactivePartsToDelayInsert                     types.Int64  `tfsdk:"inactive_parts_to_delay_insert"`
	InactivePartsToThrowInsert                     types.Int64  `tfsdk:"inactive_parts_to_throw_insert"`
	MaxReplicatedMergesInQueue                     types.Int64  `tfsdk:"max_replicated_merges_in_queue"`
	NumberOfFreeEntriesInPoolToLowerMaxSizeOfMerge types.Int64  `tfsdk:"number_of_free_entries_in_pool_to_lower_max_size_of_merge"`
	MaxBytesToMergeAtMinSpaceInPool                types.Int64  `tfsdk:"max_bytes_to_merge_at_min_space_in_pool"`
	MaxBytesToMergeAtMaxSpaceInPool                types.Int64  `tfsdk:"max_bytes_to_merge_at_max_space_in_pool"`
	MinBytesForWidePart                            types.Int64  `tfsdk:"min_bytes_for_wide_part"`
	MinRowsForWidePart                             types.Int64  `tfsdk:"min_rows_for_wide_part"`
	TtlOnlyDropParts                               types.Bool   `tfsdk:"ttl_only_drop_parts"`
	AllowRemoteFsZeroCopyReplication               types.Bool   `tfsdk:"allow_remote_fs_zero_copy_replication"`
	MergeWithTtlTimeout                            types.String `tfsdk:"merge_with_ttl_timeout"`
	MergeWithRecompressionTtlTimeout               types.String `tfsdk:"merge_with_recompression_ttl_timeout"`
	MaxPartsInTotal                                types.Int64  `tfsdk:"max_parts_in_total"`
	MaxNumberOfMergesWithTtlInPool                 types.Int64  `tfsdk:"max_number_of_merges_with_ttl_in_pool"`
	CleanupDelayPeriod                             types.String `tfsdk:"cleanup_delay_period"`
	NumberOfFreeEntriesInPoolToExecuteMutation     types.Int64  `tfsdk:"number_of_free_entries_in_pool_to_execute_mutation"`
	MaxAvgPartSizeForTooManyParts                  types.Int64  `tfsdk:"max_avg_part_size_for_too_many_parts"`
	MinAgeToForceMergeSeconds                      types.String `tfsdk:"min_age_to_force_merge_seconds"`
	MinAgeToForceMergeOnPartitionOnly              types.Bool   `tfsdk:"min_age_to_force_merge_on_partition_only"`
	MergeSelectingSleepMs                          types.String `tfsdk:"merge_selecting_sleep_ms"`
}

//nolint:unused
type clickhouseConfigCompression struct {
	Method           types.String  `tfsdk:"method"`
	MinPartSize      types.Int64   `tfsdk:"min_part_size"`
	MinPartSizeRatio types.Float64 `tfsdk:"min_part_size_ratio"`
	Level            types.Int64   `tfsdk:"level"`
}

type clickhouseConfigKafka struct {
	SecurityProtocol                 types.String `tfsdk:"security_protocol"`
	SaslMechanism                    types.String `tfsdk:"sasl_mechanism"`
	SaslUsername                     types.String `tfsdk:"sasl_username"`
	SaslPassword                     types.String `tfsdk:"sasl_password"`
	EnableSslCertificateVerification types.Bool   `tfsdk:"enable_ssl_certificate_verification"`
	MaxPoolIntervalMs                types.String `tfsdk:"max_poll_interval_ms"`
	SessionTimeoutMs                 types.String `tfsdk:"session_timeout_ms"`
}

//nolint:unused
type clickhouseConfigRabbitmq struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Vhost    types.String `tfsdk:"vhost"`
}

func clickhouseConfigLogLevelValidator() validator.String {
	names := make([]string, len(clickhouse.ClickhouseConfig_LogLevel_name))
	for i, v := range clickhouse.ClickhouseConfig_LogLevel_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func clickhouseConfigKafkaSecurityProtocolValidator() validator.String {
	names := make([]string, 0)
	for k, v := range clickhouse.ClickhouseConfig_Kafka_SecurityProtocol_value {
		if v == 0 {
			continue
		}
		names = append(names, strings.ToUpper(strings.TrimPrefix(k, "SECURITY_PROTOCOL_")))
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func clickhouseConfigKafkaSaslMechanismValidator() validator.String {
	names := make([]string, 0)
	for k, v := range clickhouse.ClickhouseConfig_Kafka_SaslMechanism_value {
		if v == 0 {
			continue
		}
		names = append(names, strings.ToUpper(strings.TrimPrefix(k, "SASL_MECHANISM_")))
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

//nolint:unused
func clickhouseConfigCompressionMethodValidator() validator.String {
	names := make([]string, len(clickhouse.ClickhouseConfig_Compression_Method_name))
	for i, v := range clickhouse.ClickhouseConfig_Compression_Method_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ClickhouseClusterResource{}
var _ resource.ResourceWithImportState = &ClickhouseClusterResource{}

func NewClickhouseClusterResource() resource.Resource {
	return &ClickhouseClusterResource{}
}

type ClickhouseClusterResource struct {
	sdk *dcsdk.SDK
	svc *dcgen.ClusterServiceClient
}

func (r *ClickhouseClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse_cluster"
}

func (r *ClickhouseClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClickHouse Cluster resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the project where the ClickHouse cluster is created",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cloud_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud provider where the cluster is created. Possible values: `aws` and `gcp`",
				Validators:          []validator.String{cloudTypeValidator()},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the region where resources are created",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster description",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Version of the ClickHouse DBMS.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"network_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the network where the cluster is created",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
		Blocks: map[string]schema.Block{
			"resources": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"clickhouse": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"resource_preset_id": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "ID of the computational resources preset available to a host (CPU, memory, etc.)",
							},
							"disk_size": schema.Int64Attribute{
								Optional:            true,
								PlanModifiers:       []planmodifier.Int64{&suppressAutoscaledDiskDiff{}},
								MarkdownDescription: "Storage volume available to a host in bytes",
							},
							"max_disk_size": schema.Int64Attribute{
								Optional:            true,
								MarkdownDescription: "Limit for automatical storage volume scale, in bytes. Autoscaling disabled if not set.",
							},
							"replica_count": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
								MarkdownDescription: "Number of hosts per shard",
							},
							"shard_count": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
								MarkdownDescription: "Number of shards in the cluster",
							},
						},
					},
					"dedicated_keeper": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"resource_preset_id": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "ID of the computational resources preset available to a host (CPU, memory, etc.)",
							},
							"disk_size": schema.Int64Attribute{
								Optional:            true,
								PlanModifiers:       []planmodifier.Int64{&suppressAutoscaledDiskDiff{}},
								MarkdownDescription: "Volume of the storage available to a host, in bytes",
							},
							"max_disk_size": schema.Int64Attribute{
								Optional:            true,
								MarkdownDescription: "Limit for automatical storage volume scale, in bytes. Autoscaling disabled if not set.",
							},
							"replica_count": schema.Int64Attribute{
								Optional:            true,
								Computed:            true,
								Default:             int64default.StaticInt64(1),
								MarkdownDescription: "Number of keeper hosts",
							},
						},
					},
				},
			},
			"access": AccessSchemaBlock(),
			"config": clickhouseConfigSchemaBlock(),
			// maintenance window
		},
	}
}

func (r *ClickhouseClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.svc = r.sdk.ClickHouse().Cluster()
}

func createClickhouseClusterRequest(m *clickhouseClusterModel) (*clickhouse.CreateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	rq := &clickhouse.CreateClusterRequest{}

	rq.ProjectId = m.ProjectId.ValueString()
	rq.CloudType = m.CloudType.ValueString()
	rq.RegionId = m.RegionId.ValueString()
	rq.Name = m.Name.ValueString()
	rq.NetworkId = m.NetworkId.ValueString()
	if v := m.Description; !v.IsNull() {
		rq.Description = v.ValueString()
	}
	if v := m.Version; !v.IsNull() {
		rq.Version = v.ValueString()
	}
	resources, d := m.Resources.convert()
	diags.Append(d...)
	rq.Resources = resources
	if m.Access != nil {
		access, d := m.Access.convert()
		diags.Append(d...)
		rq.Access = access
	}
	if m.Config != nil {
		rq.ClickhouseConfig, d = m.Config.convert()
		diags.Append(d...)
	}
	// TODO: mw

	return rq, diags
}

func (r *ClickhouseClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *clickhouseClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := createClickhouseClusterRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	dcOperation, err := r.svc.Create(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(dcOperation, err)
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
		response, err := r.svc.Get(ctx, &clickhouse.GetClusterRequest{ClusterId: data.Id.ValueString()})
		if err != nil {
			resp.Diagnostics.AddError("failed to get", err.Error())
			return
		}
		resp.Diagnostics.Append(data.parse(response)...)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClickhouseClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *clickhouseClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.svc.Get(ctx, &clickhouse.GetClusterRequest{ClusterId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}
	resp.Diagnostics.Append(data.parse(response)...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func updateClickhouseCluster(m *clickhouseClusterModel) (*clickhouse.UpdateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	rq := &clickhouse.UpdateClusterRequest{ClusterId: m.Id.ValueString()}
	rq.Name = m.Name.ValueString()
	rq.Description = m.Description.ValueString()
	rq.Version = m.Version.ValueString()

	resources, d := m.Resources.convert()
	diags.Append(d...)
	rq.Resources = resources

	config, d := m.Config.convert()
	diags.Append(d...)
	rq.ClickhouseConfig = config

	if m.Access != nil {
		access, d := m.Access.convert()
		diags.Append(d...)
		rq.Access = access
	}

	return rq, diags
}

func (r *ClickhouseClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *clickhouseClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	rq, diag := updateClickhouseCluster(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	dcOperation, err := r.svc.Update(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(dcOperation, err)
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

func (r *ClickhouseClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *clickhouseClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dcOperation, err := r.svc.Delete(ctx, &clickhouse.DeleteClusterRequest{ClusterId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(dcOperation, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
}

func (r *ClickhouseClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *clickhouseClusterModel) parse(rs *clickhouse.Cluster) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ProjectId = types.StringValue(rs.ProjectId)
	m.CloudType = types.StringValue(rs.CloudType)
	m.RegionId = types.StringValue(rs.RegionId)
	m.Name = types.StringValue(rs.Name)
	m.Description = types.StringValue(rs.Description)
	m.Version = types.StringValue(rs.Version)
	m.NetworkId = types.StringValue(rs.NetworkId)

	if m.Resources == nil {
		m.Resources = &clickhouseClusterResources{}
	}
	diags.Append(m.Resources.parse(rs.Resources)...)
	if m.Config == nil {
		m.Config = &clickhouseConfig{}
	}
	diags.Append(m.Config.parse(rs.ClickhouseConfig)...)
	if access := rs.GetAccess(); access != nil {
		if m.Access == nil {
			m.Access = new(AccessModel)
		}
		diags.Append(m.Access.parse(access)...)
	}

	// parse MW
	return diags
}

func (m *clickhouseClusterResources) parse(rs *clickhouse.ClusterResources) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.Clickhouse == nil {
		m.Clickhouse = &clickhouseClusterResourcesClickhouse{}
	}
	m.Clickhouse.ResourcePresetId = types.StringValue(rs.Clickhouse.ResourcePresetId)
	m.Clickhouse.DiskSize = types.Int64Value(rs.Clickhouse.DiskSize.GetValue())
	m.Clickhouse.ReplicaCount = types.Int64Value(rs.Clickhouse.ReplicaCount.GetValue())
	m.Clickhouse.ShardCount = types.Int64Value(rs.Clickhouse.ShardCount.GetValue())
	if v := rs.Clickhouse.MaxDiskSize; v != nil {
		m.Clickhouse.MaxDiskSize = types.Int64Value(v.GetValue())
	}

	if v := rs.GetDedicatedKeeper(); v != nil {
		// temporary fix
		// server should respond with `nil` structure in case of unused keepers
		// This if-clause needed for handle situation with current empty structure
		if v.GetDiskSize().GetValue() == 0 || v.GetReplicaCount().GetValue() == 0 {
			m.Keeper = nil
		} else {
			m.Keeper = new(clickhouseClusterResourcesKeeper)
			m.Keeper.ResourcePresetId = types.StringValue(v.ResourcePresetId)
			m.Keeper.DiskSize = types.Int64Value(v.DiskSize.GetValue())
			m.Keeper.ReplicaCount = types.Int64Value(v.ReplicaCount.GetValue())
			if v := rs.Clickhouse.MaxDiskSize; v != nil {
				m.Keeper.MaxDiskSize = types.Int64Value(v.GetValue())
			}
		}
	} else {
		m.Keeper = nil
	}

	return diags
}

func (m *clickhouseConfig) convert() (*clickhouse.ClickhouseConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	config := &clickhouse.ClickhouseConfig{}

	if m == nil {
		return config, diags
	}

	if v := m.LogLevel; !v.IsUnknown() {
		config.LogLevel = clickhouse.ClickhouseConfig_LogLevel(clickhouse.ClickhouseConfig_LogLevel_value[v.ValueString()])
	}
	if v := m.MaxConnections; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.MaxConnections = wrapperspb.Int64(v.ValueInt64())
	}

	if v := m.MaxConcurrentQueries; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.MaxConcurrentQueries = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.KeepAliveTimeout; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse keep_alive_timeout", err.Error())
		}
		config.KeepAliveTimeout = durationpb.New(duration)
	}
	if v := m.UncompressedCacheSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.UncompressedCacheSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.MarkCacheSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.MarkCacheSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.MaxTableSizeToDrop; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.MaxTableSizeToDrop = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.MaxPartitionSizeToDrop; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.MaxPartitionSizeToDrop = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.Timezone; !v.IsUnknown() && v.ValueString() != "" {
		config.Timezone = wrapperspb.String(v.ValueString())
	}
	if v := m.BackgroundPoolSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.BackgroundPoolSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.BackgroundSchedulePoolSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.BackgroundSchedulePoolSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.BackgroundFetchesPoolSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.BackgroundFetchesPoolSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.BackgroundMovePoolSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.BackgroundMovePoolSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.BackgroundCommonPoolSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.BackgroundCommonPoolSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.BackgroundMergesMutationsConcurrencyRatio; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.BackgroundMergesMutationsConcurrencyRatio = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.TotalMemoryProfilerStep; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.TotalMemoryProfilerStep = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.TotalMemoryTrackerSampleProbability; !v.IsUnknown() && v.ValueFloat64() != 0 {
		config.TotalMemoryTrackerSampleProbability = wrapperspb.Double(v.ValueFloat64())
	}
	if v := m.BackgroundMessageBrokerSchedulePoolSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.BackgroundMessageBrokerSchedulePoolSize = wrapperspb.Int64(v.ValueInt64())
	}
	// merge_tree
	// compression
	// graphite_rollup
	if v := m.Kafka; v != nil {
		k, d := m.Kafka.convert()
		diags.Append(d...)
		config.Kafka = k
	}
	// kafka_topics
	// rabbitmq
	if v := m.QueryLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.QueryLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.QueryLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse query_log_retention_time", err.Error())
		}
		config.QueryLogRetentionTime = durationpb.New(duration)
	}

	if v := m.QueryThreadLogEnabled; !v.IsNull() {
		config.QueryThreadLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.QueryThreadLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.QueryThreadLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.QueryThreadLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse query_thread_log_retention_time", err.Error())
		}
		config.QueryThreadLogRetentionTime = durationpb.New(duration)
	}

	if v := m.QueryViewsLogEnabled; !v.IsNull() {
		config.QueryViewsLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.QueryViewsLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.QueryViewsLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.QueryViewsLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse query_views_log_retention_time", err.Error())
		}
		config.QueryViewsLogRetentionTime = durationpb.New(duration)
	}

	if v := m.PartLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.PartLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.PartLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse part_log_retention_time", err.Error())
		}
		config.PartLogRetentionTime = durationpb.New(duration)
	}

	if v := m.MetricLogEnabled; !v.IsNull() {
		config.MetricLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.MetricLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.MetricLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.MetricLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse metric_log_retention_time", err.Error())
		}
		config.MetricLogRetentionTime = durationpb.New(duration)
	}

	if v := m.AsynchronousMetricLogEnabled; !v.IsNull() {
		config.AsynchronousMetricLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.AsynchronousMetricLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.AsynchronousMetricLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.AsynchronousMetricLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse asynchronous_metric_log_retention_time", err.Error())
		}
		config.AsynchronousMetricLogRetentionTime = durationpb.New(duration)
	}

	if v := m.TraceLogEnabled; !v.IsNull() {
		config.TraceLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.TraceLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.TraceLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.TraceLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse trace_log_retention_time", err.Error())
		}
		config.TraceLogRetentionTime = durationpb.New(duration)
	}

	if v := m.TextLogEnabled; !v.IsNull() {
		config.TextLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.TextLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.TextLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.TextLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse text_log_retention_time", err.Error())
		}
		config.TextLogRetentionTime = durationpb.New(duration)
	}

	if v := m.TextLogLevel; !v.IsUnknown() && v.ValueString() != "" {
		config.TextLogLevel = clickhouse.ClickhouseConfig_LogLevel(clickhouse.ClickhouseConfig_LogLevel_value[v.ValueString()])
	}

	if v := m.OpentelemetrySpanLogEnabled; !v.IsNull() {
		config.OpentelemetrySpanLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.OpentelemetrySpanLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.OpentelemetrySpanLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.OpentelemetrySpanLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse opentelemetry_span_log_retention_time", err.Error())
		}
		config.OpentelemetrySpanLogRetentionTime = durationpb.New(duration)
	}

	if v := m.SessionLogEnabled; !v.IsNull() {
		config.SessionLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.SessionLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.SessionLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.SessionLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse session_log_retention_time", err.Error())
		}
		config.SessionLogRetentionTime = durationpb.New(duration)
	}

	if v := m.ZookeeperLogEnabled; !v.IsNull() {
		config.ZookeeperLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.ZookeeperLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.ZookeeperLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.ZookeeperLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse zookeeper_log_retention_time", err.Error())
		}
		config.ZookeeperLogRetentionTime = durationpb.New(duration)
	}

	if v := m.AsynchronousInsertLogEnabled; !v.IsNull() {
		config.AsynchronousInsertLogEnabled = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.AsynchronousInsertLogRetentionSize; !v.IsUnknown() && v.ValueInt64() != 0 {
		config.AsynchronousInsertLogRetentionSize = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.AsynchronousInsertLogRetentionTime; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("config"), "failed to parse asynchronous_insert_log_retention_time", err.Error())
		}
		config.AsynchronousInsertLogRetentionTime = durationpb.New(duration)
	}

	return config, diags
}

func (m *clickhouseConfig) parse(rs *clickhouse.ClickhouseConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	m.LogLevel = types.StringValue(rs.LogLevel.String())
	if v := rs.MaxConnections; v != nil {
		m.MaxConnections = types.Int64Value(v.Value)
	}
	if v := rs.MaxConnections; v != nil {
		m.MaxConnections = types.Int64Value(v.Value)
	}
	if v := rs.MaxConcurrentQueries; v != nil {
		m.MaxConcurrentQueries = types.Int64Value(v.Value)
	}
	if v := rs.KeepAliveTimeout; v != nil {
		m.KeepAliveTimeout = types.StringValue(v.String())
	}
	if v := rs.UncompressedCacheSize; v != nil {
		m.UncompressedCacheSize = types.Int64Value(v.Value)
	}
	if v := rs.MarkCacheSize; v != nil {
		m.MarkCacheSize = types.Int64Value(v.Value)
	}
	if v := rs.MaxTableSizeToDrop; v != nil {
		m.MaxTableSizeToDrop = types.Int64Value(v.Value)
	}
	if v := rs.MaxPartitionSizeToDrop; v != nil {
		m.MaxPartitionSizeToDrop = types.Int64Value(v.Value)
	}
	if v := rs.Timezone; v != nil {
		m.Timezone = types.StringValue(v.Value)
	}
	if v := rs.BackgroundPoolSize; v != nil {
		m.BackgroundPoolSize = types.Int64Value(v.Value)
	}
	if v := rs.BackgroundSchedulePoolSize; v != nil {
		m.BackgroundSchedulePoolSize = types.Int64Value(v.Value)
	}
	if v := rs.BackgroundFetchesPoolSize; v != nil {
		m.BackgroundFetchesPoolSize = types.Int64Value(v.Value)
	}
	if v := rs.BackgroundMovePoolSize; v != nil {
		m.BackgroundMovePoolSize = types.Int64Value(v.Value)
	}
	if v := rs.BackgroundCommonPoolSize; v != nil {
		m.BackgroundCommonPoolSize = types.Int64Value(v.Value)
	}
	if v := rs.BackgroundMergesMutationsConcurrencyRatio; v != nil {
		m.BackgroundMergesMutationsConcurrencyRatio = types.Int64Value(v.Value)
	}
	if v := rs.TotalMemoryProfilerStep; v != nil {
		m.TotalMemoryProfilerStep = types.Int64Value(v.Value)
	}
	if v := rs.TotalMemoryTrackerSampleProbability; v != nil {
		m.TotalMemoryTrackerSampleProbability = types.Float64Value(v.Value)
	}
	if v := rs.BackgroundMessageBrokerSchedulePoolSize; v != nil {
		m.BackgroundMessageBrokerSchedulePoolSize = types.Int64Value(v.Value)
	}
	// merge_tree
	// compression
	// graphite_rollup
	if v := rs.GetKafka(); v != nil {
		if m.Kafka == nil {
			m.Kafka = &clickhouseConfigKafka{}
		}
		diags.Append(m.Kafka.parse(v)...)
	}
	// kafka topics
	// rabbit_mq
	if v := rs.QueryLogRetentionSize; v != nil {
		m.QueryLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.QueryLogRetentionTime; v != nil {
		m.QueryLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.QueryThreadLogEnabled; v != nil {
		m.QueryThreadLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.QueryThreadLogRetentionSize; v != nil {
		m.QueryThreadLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.QueryThreadLogRetentionTime; v != nil {
		m.QueryThreadLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.QueryViewsLogEnabled; v != nil {
		m.QueryViewsLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.QueryViewsLogRetentionSize; v != nil {
		m.QueryViewsLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.QueryViewsLogRetentionTime; v != nil {
		m.QueryViewsLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.PartLogRetentionSize; v != nil {
		m.PartLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.PartLogRetentionTime; v != nil {
		m.PartLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.MetricLogEnabled; v != nil {
		m.MetricLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.MetricLogRetentionSize; v != nil {
		m.MetricLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.MetricLogRetentionTime; v != nil {
		m.MetricLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.AsynchronousMetricLogEnabled; v != nil {
		m.AsynchronousMetricLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.AsynchronousMetricLogRetentionSize; v != nil {
		m.AsynchronousMetricLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.AsynchronousMetricLogRetentionTime; v != nil {
		m.AsynchronousMetricLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.TraceLogEnabled; v != nil {
		m.TraceLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.TraceLogRetentionSize; v != nil {
		m.TraceLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.TraceLogRetentionTime; v != nil {
		m.TraceLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.TextLogEnabled; v != nil {
		m.TextLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.TextLogRetentionSize; v != nil {
		m.TextLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.TextLogRetentionTime; v != nil {
		m.TextLogRetentionTime = types.StringValue(v.String())
	}
	m.TextLogLevel = types.StringValue(rs.TextLogLevel.String())

	if v := rs.OpentelemetrySpanLogEnabled; v != nil {
		m.OpentelemetrySpanLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.OpentelemetrySpanLogRetentionSize; v != nil {
		m.OpentelemetrySpanLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.OpentelemetrySpanLogRetentionTime; v != nil {
		m.OpentelemetrySpanLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.SessionLogEnabled; v != nil {
		m.SessionLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.SessionLogRetentionSize; v != nil {
		m.SessionLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.SessionLogRetentionTime; v != nil {
		m.SessionLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.ZookeeperLogEnabled; v != nil {
		m.ZookeeperLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.ZookeeperLogRetentionSize; v != nil {
		m.ZookeeperLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.ZookeeperLogRetentionTime; v != nil {
		m.ZookeeperLogRetentionTime = types.StringValue(v.String())
	}

	if v := rs.AsynchronousInsertLogEnabled; v != nil {
		m.AsynchronousInsertLogEnabled = types.BoolValue(v.Value)
	}
	if v := rs.AsynchronousInsertLogRetentionSize; v != nil {
		m.AsynchronousInsertLogRetentionSize = types.Int64Value(v.Value)
	}
	if v := rs.AsynchronousInsertLogRetentionTime; v != nil {
		m.AsynchronousInsertLogRetentionTime = types.StringValue(v.String())
	}

	return diags
}

func clickhouseConfigSchemaBlock() schema.Block {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"log_level": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(clickhouse.ClickhouseConfig_LOG_LEVEL_INFORMATION.String()),
				Validators:          []validator.String{clickhouseConfigLogLevelValidator()},
				MarkdownDescription: "Level of logged events, such as `ERROR` or `TRACE`",
			},
			"max_connections": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of inbound client connections",
			},
			"max_concurrent_queries": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum number of requests processed simultaneously",
			},
			"keep_alive_timeout": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Time in seconds for which ClickHouse waits for incoming requests before closing the connection",
			},
			"uncompressed_cache_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Cache size in bytes for uncompressed data used by table engines in the MergeTree family",
			},
			"mark_cache_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Approximate size in bytes of the mark cache used by table engines in the MergeTree family",
			},
			"max_table_size_to_drop": schema.Int64Attribute{Optional: true,
				MarkdownDescription: "Maximum size in bytes of a table in the MergeTree family that can be deleted using the `DROP TABLE` query",
			},
			"max_partition_size_to_drop": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum partition size in bytes for the MergeTree family at which a table can be deleted using the `DROP TABLE` query",
			},
			"timezone": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cluster time zone from the IANA Time Zone Database, such as `Africa/Abidjan`",
			},
			"background_pool_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of threads performing background merges and mutations for tables with MergeTree engines",
			},
			"background_schedule_pool_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of threads for background jobs for replicated tables, streams in Apache Kafka, and DNS cache updates",
			},
			"background_fetches_pool_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of threads performing background fetches for replicated tables",
			},
			"background_move_pool_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of threads performing background moves of data parts for tables with MergeTree engines",
			},
			"background_common_pool_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of threads performing various operations (mostly garbage collection) for tables with MergeTree engines",
			},
			"background_merges_mutations_concurrency_ratio": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Ratio of the number of threads to the number of background merges and mutations that can be executed concurrently",
			},
			"total_memory_profiler_step": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "RAM in bytes for a stack trace at each memory allocation step",
			},
			"total_memory_tracker_sample_probability": schema.Float64Attribute{
				Optional:            true,
				MarkdownDescription: "Allows collecting and logging informaiton about random memory allocation and release with the specified probability",
			},
			"background_message_broker_schedule_pool_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of threads for executing background message translation operations",
			},
			// merge_tree, compression, ...
			"query_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the query log table in bytes",
			},
			"query_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the query log table in the duration string format, such as `2h45m`",
			},

			"query_thread_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging query threads",
			},
			"query_thread_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the query thread log table in bytes",
			},
			"query_thread_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the query thread log table in the duration string format, such as `2h45m`",
			},

			"query_views_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging query views",
			},
			"query_views_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the query views log table in bytes",
			},
			"query_views_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the query views log table in the duration string format, such as `2h45m`",
			},

			"part_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the part log table in bytes",
			},
			"part_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the part log table in the duration string format, such as `2h45m`",
			},

			"metric_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging metric values from the `system.metrics` and the `system.events` tables to `system.metric_log`",
			},
			"metric_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the metric log table in bytes",
			},
			"metric_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the metric log table in the duration string format, such as `2h45m`",
			},

			"asynchronous_metric_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging asynchronous metrics",
			},
			"asynchronous_metric_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the asynchronous insert log table in bytes",
			},
			"asynchronous_metric_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the asynchronous insert log table in the duration string format, such as `2h45m`",
			},

			"trace_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging stack traces collected by the query profiler",
			},
			"trace_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the trace log table in bytes",
			},
			"trace_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the trace log table in the duration string format, such as `2h45m`",
			},

			"text_log_enabled": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "Enable system logs",
			},
			"text_log_retention_size": schema.Int64Attribute{
				Optional: true,
				MarkdownDescription: "Maximum size of the text log table in bytes",
			},
			"text_log_retention_time": schema.StringAttribute{Optional: true},
			"text_log_level": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(clickhouse.ClickhouseConfig_LOG_LEVEL_INFORMATION.String()),
				Validators: []validator.String{clickhouseConfigLogLevelValidator()},
				MarkdownDescription: "Level of logging, such as `ERROR` or `TRACE`",
			},

			"opentelemetry_span_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable log trace and metric values from a distributed application",
			},
			"opentelemetry_span_log_retention_size": schema.Int64Attribute{
				Optional: true,
				MarkdownDescription: "Maximum size of the opentelemetry span log table in bytes",
			},
			"opentelemetry_span_log_retention_time": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Retention time of the opentelemetry span log table in the duration string format, such as `2h45m`",
			},

			"session_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging successful and failed login/logout events",
			},
			"session_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the session log table in bytes",
			},
			"session_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the session log in the duration string format, such as `2h45m`",
			},

			"zookeeper_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging parameters of requests to the ZooKeeper server and responses from it",
			},
			"zookeeper_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the ZooKeeper log table in bytes",
			},
			"zookeeper_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the ZooKeeper log table in the duration string format, such as `2h45m`",
			},

			"asynchronous_insert_log_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable logging asynchronous inserts",
			},
			"asynchronous_insert_log_retention_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum size of the asynchronous insert log table in bytes",
			},
			"asynchronous_insert_log_retention_time": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Retention time of the asynchronous insert log table in the duration string format, such as `2h45m`",
			},
		},
		Blocks: map[string]schema.Block{
			"kafka": clickhouseKafkaSchemaBlock(),
		},
		Validators: []validator.Object{objectvalidator.IsRequired()},
	}
}

func clickhouseKafkaSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"security_protocol": schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			Validators:    []validator.String{clickhouseConfigKafkaSecurityProtocolValidator()},
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			MarkdownDescription: "Security protocol used for authentication",
		},
		"sasl_mechanism": schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			Validators:    []validator.String{clickhouseConfigKafkaSaslMechanismValidator()},
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			MarkdownDescription: "SASL authentication mechanism",
		},
		"sasl_username": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "Apache Kafka® account username",
		},
		"sasl_password": schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
			MarkdownDescription: "Apache Kafka® account password",
		},
		"enable_ssl_certificate_verification": schema.BoolAttribute{
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			MarkdownDescription: "Enable SSL certificate verification",
		},
		"max_poll_interval_ms": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "Maximum interval in milliseconds between making poll calls to get messages for high-level consumers",
		},
		"session_timeout_ms": schema.StringAttribute{
			Optional: true,
			MarkdownDescription: "Timeout to maintain a client group session",
		},
	}
}

func clickhouseKafkaSchemaBlock() schema.Block {
	return schema.SingleNestedBlock{Attributes: clickhouseKafkaSchemaAttributes()}
}

func (m *clickhouseConfigKafka) parse(r *clickhouse.ClickhouseConfig_Kafka) diag.Diagnostics {
	var diags diag.Diagnostics

	m.SecurityProtocol = types.StringValue(strings.TrimPrefix(r.GetSecurityProtocol().String(), "SECURITY_PROTOCOL_"))
	m.SaslMechanism = types.StringValue(strings.TrimPrefix(r.GetSaslMechanism().String(), "SASL_MECHANISM_"))
	if v := r.GetSaslUsername(); v != nil {
		m.SaslUsername = types.StringValue(v.GetValue())
	}
	if v := r.GetSaslPassword(); v != nil {
		m.SaslPassword = types.StringValue(v.GetValue())
	}
	if v := r.GetEnableSslCertificateVerification(); v != nil {
		m.EnableSslCertificateVerification = types.BoolValue(v.GetValue())
	}
	if v := r.GetMaxPollIntervalMs(); v != nil {
		m.MaxPoolIntervalMs = types.StringValue(v.AsDuration().String())
	}
	if v := r.GetSessionTimeoutMs(); v != nil {
		m.SessionTimeoutMs = types.StringValue(v.AsDuration().String())
	}

	return diags
}

func (m *clickhouseConfigKafka) convert() (*clickhouse.ClickhouseConfig_Kafka, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := &clickhouse.ClickhouseConfig_Kafka{}

	{
		key := fmt.Sprintf("SECURITY_PROTOCOL_%v", strings.Replace(m.SecurityProtocol.ValueString(), "\"", "", -1))
		securityProtocol := clickhouse.ClickhouseConfig_Kafka_SecurityProtocol_value[key]
		r.SecurityProtocol = clickhouse.ClickhouseConfig_Kafka_SecurityProtocol(securityProtocol)
	}
	{
		key := fmt.Sprintf("SASL_MECHANISM_%v", strings.Replace(m.SaslMechanism.ValueString(), "\"", "", -1))
		SaslMechanism := clickhouse.ClickhouseConfig_Kafka_SaslMechanism_value[key]
		r.SaslMechanism = clickhouse.ClickhouseConfig_Kafka_SaslMechanism(SaslMechanism)
	}
	if v := m.SaslUsername; !v.IsUnknown() && v.ValueString() != "" {
		r.SaslUsername = wrapperspb.String(v.ValueString())
	}
	if v := m.SaslPassword; !v.IsUnknown() && v.ValueString() != "" {
		r.SaslPassword = wrapperspb.String(v.ValueString())
	}
	if v := m.EnableSslCertificateVerification; !v.IsNull() {
		r.EnableSslCertificateVerification = wrapperspb.Bool(v.ValueBool())
	}
	if v := m.MaxPoolIntervalMs; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("kafka"), "failed to parse max_pool_interval_ms", err.Error())
		}
		r.MaxPollIntervalMs = durationpb.New(duration)
	}
	if v := m.SessionTimeoutMs; !v.IsUnknown() && v.ValueString() != "" {
		duration, err := time.ParseDuration(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("kafka"), "failed to parse session_timeout_ms", err.Error())
		}
		r.SessionTimeoutMs = durationpb.New(duration)
	}

	return r, diags
}
