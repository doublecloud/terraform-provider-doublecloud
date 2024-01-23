package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/doublecloud/go-genproto/doublecloud/kafka/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgen "github.com/doublecloud/go-sdk/gen/kafka"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaClusterResource{}
var _ resource.ResourceWithImportState = &KafkaClusterResource{}

func NewKafkaClusterResource() resource.Resource {
	return &KafkaClusterResource{}
}

type KafkaClusterResource struct {
	sdk            *dcsdk.SDK
	clusterService *dcgen.ClusterServiceClient
	userService    *dcgen.UserServiceClient
	topicService   *dcgen.TopicServiceClient
}

type KafkaClusterModel struct {
	Id                    types.String             `tfsdk:"id"`
	ProjectID             types.String             `tfsdk:"project_id"`
	CloudType             types.String             `tfsdk:"cloud_type"`
	RegionID              types.String             `tfsdk:"region_id"`
	Name                  types.String             `tfsdk:"name"`
	Description           types.String             `tfsdk:"description"`
	Version               types.String             `tfsdk:"version"`
	Resources             *KafkaResourcesModel     `tfsdk:"resources"`
	NetworkId             types.String             `tfsdk:"network_id"`
	SchemaRegistry        *schemaRegistryModel     `tfsdk:"schema_registry"`
	Access                *AccessModel             `tfsdk:"access"`
	ConnectionInfo        types.Object             `tfsdk:"connection_info"`
	PrivateConnectionInfo types.Object             `tfsdk:"private_connection_info"`
	Config                *KafkaClusterConfigModel `tfsdk:"config"`
}

type schemaRegistryModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type KafkaResourcesKafkaModel struct {
	ResourcePresetId types.String `tfsdk:"resource_preset_id"`
	DiskSize         types.Int64  `tfsdk:"disk_size"`
	MaxDiskSize      types.Int64  `tfsdk:"max_disk_size"`
	BrokerCount      types.Int64  `tfsdk:"broker_count"`
	ZoneCount        types.Int64  `tfsdk:"zone_count"`
}

type KafkaClusterConfigModel struct {
	MessageMaxBytes      types.Int64 `tfsdk:"message_max_bytes"`
	ReplicaFetchMaxBytes types.Int64 `tfsdk:"replica_fetch_max_bytes"`
	LogRetentionBytes    types.Int64 `tfsdk:"log_retention_bytes"`
	LogRetentionHours    types.Int64 `tfsdk:"log_retention_hours"`
	LogRetentionMinutes  types.Int64 `tfsdk:"log_retention_minutes"`
	LogRetentionMs       types.Int64 `tfsdk:"log_retention_ms"`
}

type KafkaResourcesModel struct {
	Kafka KafkaResourcesKafkaModel `tfsdk:"kafka"`
}

func (r *KafkaClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_cluster"
}

func (r *KafkaClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster Id",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project Id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud type (aws, gcp, azure)",
				Validators:          []validator.String{cloudTypeValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of cluster",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of cluster",
				Default:             stringdefault.StaticString(""),
			},
			"region_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Region of cluster",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network of cluster",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Version of Apache Kafka",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"connection_info": schema.SingleNestedAttribute{
				Computed:      true,
				Attributes:    kafkaConnectionInfoResSchema(),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			},
			"private_connection_info": schema.SingleNestedAttribute{
				Computed:      true,
				Attributes:    kafkaConnectionInfoResSchema(),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			},
		},
		Blocks: map[string]schema.Block{
			"resources": schema.SingleNestedBlock{
				Description: "Resources of cluster",
				Blocks: map[string]schema.Block{
					"kafka": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"resource_preset_id": schema.StringAttribute{Required: true},
							"disk_size": schema.Int64Attribute{
								Required:      true,
								PlanModifiers: []planmodifier.Int64{&suppressAutoscaledDiskDiff{}},
							},
							"max_disk_size": schema.Int64Attribute{Optional: true},
							"broker_count":  schema.Int64Attribute{Required: true},
							"zone_count":    schema.Int64Attribute{Required: true},
						},
					},
				},
			},
			"schema_registry": schema.SingleNestedBlock{
				Description: "Schema Registry configuration",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{Computed: true, Optional: true},
				},
			},
			"access": AccessSchemaBlock(),
			"config": schema.SingleNestedBlock{
				Description: "Cluster configuration",
				Attributes: map[string]schema.Attribute{
					"message_max_bytes":       schema.Int64Attribute{Optional: true},
					"replica_fetch_max_bytes": schema.Int64Attribute{Optional: true},
					"log_retention_bytes":     schema.Int64Attribute{Optional: true},
					"log_retention_hours":     schema.Int64Attribute{Optional: true},
					"log_retention_minutes":   schema.Int64Attribute{Optional: true},
					"log_retention_ms":        schema.Int64Attribute{Optional: true},
				},
			},
		},
		MarkdownDescription: "Kafka cluster resource",
		Version:             0,
	}
}

func (r *KafkaClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.clusterService = r.sdk.Kafka().Cluster()
	r.userService = r.sdk.Kafka().User()
	r.topicService = r.sdk.Kafka().Topic()
}

func createKafkaClusterRequest(m *KafkaClusterModel) (*kafka.CreateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	rq := &kafka.CreateClusterRequest{}
	rq.Name = m.Name.ValueString()
	rq.CloudType = m.CloudType.ValueString()
	rq.ProjectId = m.ProjectID.ValueString()
	rq.Description = m.Description.ValueString()
	rq.RegionId = m.RegionID.ValueString()
	rq.NetworkId = m.NetworkId.ValueString()
	rq.Version = m.Version.ValueString()

	rq.Resources = &kafka.ClusterResources{
		Kafka: &kafka.ClusterResources_Kafka{
			ResourcePresetId: m.Resources.Kafka.ResourcePresetId.ValueString(),
			DiskSize:         wrapperspb.Int64(m.Resources.Kafka.DiskSize.ValueInt64()),
			BrokerCount:      wrapperspb.Int64(m.Resources.Kafka.BrokerCount.ValueInt64()),
			ZoneCount:        wrapperspb.Int64(m.Resources.Kafka.ZoneCount.ValueInt64()),
		},
	}
	if v := m.Resources.Kafka.MaxDiskSize; !v.IsNull() {
		rq.Resources.Kafka.MaxDiskSize = wrapperspb.Int64(m.Resources.Kafka.MaxDiskSize.ValueInt64())
	}

	if m.SchemaRegistry != nil {
		enabled := m.SchemaRegistry.Enabled.ValueBool()
		rq.SchemaRegistryConfig = &kafka.SchemaRegistryConfig{
			Enabled: enabled,
		}
	}

	if m.Access != nil {
		access, d := m.Access.convert()
		diags.Append(d...)
		rq.Access = access
	}

	if m.Config != nil {
		config, d := m.Config.convert()
		diags.Append(d...)
		rq.KafkaConfig = config
	}

	return rq, diags
}

func deleteKafkaClusterRequest(m *KafkaClusterModel) (*kafka.DeleteClusterRequest, diag.Diagnostics) {
	rq := &kafka.DeleteClusterRequest{ClusterId: m.Id.ValueString()}
	return rq, nil
}

//nolint:unused
func kafkaAccessRoleValidator() validator.String {
	names := make([]string, len(kafka.Permission_AccessRole_name))
	for i, v := range kafka.Permission_AccessRole_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func cloudTypeValidator() validator.String {
	return stringvalidator.OneOfCaseInsensitive([]string{"aws", "gcp"}...)
}

func (r *KafkaClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *KafkaClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := createKafkaClusterRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.clusterService.Create(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}
	data.Id = types.StringValue(op.ResourceId())

	// TODO: make a parse server response into model
	getRq, diag := getKafkaClusterResourceRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	cluster, err := r.clusterService.Get(ctx, getRq)
	if err != nil {
		resp.Diagnostics.AddError("failed to read", err.Error())
		return
	}
	data.Version = types.StringValue(cluster.Version)
	if info := cluster.GetConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"connection_string": types.StringType,
			"user":              types.StringType,
			"password":          types.StringType,
		},
			map[string]attr.Value{
				"connection_string": types.StringValue(info.GetConnectionString()),
				"user":              types.StringValue(info.GetUser()),
				"password":          types.StringValue(info.GetPassword()),
			},
		)
		resp.Diagnostics.Append(d...)
		data.ConnectionInfo = o
	}
	if info := cluster.GetPrivateConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"connection_string": types.StringType,
			"user":              types.StringType,
			"password":          types.StringType,
		},
			map[string]attr.Value{
				"connection_string": types.StringValue(info.GetConnectionString()),
				"user":              types.StringValue(info.GetUser()),
				"password":          types.StringValue(info.GetPassword()),
			},
		)
		resp.Diagnostics.Append(d...)
		data.PrivateConnectionInfo = o
	}

	tflog.Info(ctx, fmt.Sprintf("doublecloud_kafka_cluster has been created: %s", op.ResourceId()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getKafkaClusterResourceRequest(m *KafkaClusterModel) (*kafka.GetClusterRequest, diag.Diagnostics) {
	if m.Id == types.StringNull() {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unknown identifier", "missed one of required fields: cluster_id or name")}
	}
	return &kafka.GetClusterRequest{
		ClusterId: m.Id.ValueString(),
		Sensitive: true,
	}, nil
}

func (r *KafkaClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *KafkaClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Warning or errors can be collected in a slice type
	// var diags diag.Diagnostics

	rq, diag := getKafkaClusterResourceRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.clusterService.Get(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	data.Id = types.StringValue(rs.Id)
	data.Name = types.StringValue(rs.Name)
	data.ProjectID = types.StringValue(rs.ProjectId)
	data.Description = types.StringValue(rs.Description)
	data.CloudType = types.StringValue(rs.CloudType)
	data.RegionID = types.StringValue(rs.RegionId)
	data.NetworkId = types.StringValue(rs.NetworkId)
	data.Version = types.StringValue(rs.Version)
	data.Resources = &KafkaResourcesModel{
		Kafka: KafkaResourcesKafkaModel{
			ResourcePresetId: types.StringValue(rs.GetResources().GetKafka().GetResourcePresetId()),
			DiskSize:         types.Int64Value(rs.GetResources().GetKafka().GetDiskSize().GetValue()),
			BrokerCount:      types.Int64Value(rs.GetResources().GetKafka().GetBrokerCount().GetValue()),
			ZoneCount:        types.Int64Value(rs.GetResources().GetKafka().GetZoneCount().GetValue()),
		},
	}
	if v := rs.GetResources().GetKafka().GetMaxDiskSize(); v != nil {
		data.Resources.Kafka.MaxDiskSize = types.Int64Value(v.GetValue())
	}

	if access := rs.GetAccess(); access != nil {
		if data.Access == nil {
			data.Access = new(AccessModel)
		}
		diag.Append(data.Access.parse(access)...)
	}

	if rs.SchemaRegistryConfig != nil {
		data.SchemaRegistry = &schemaRegistryModel{Enabled: types.BoolValue(rs.SchemaRegistryConfig.Enabled)}
	} else {
		data.SchemaRegistry = nil
	}

	if config := rs.GetKafkaConfig(); config != nil {
		data.Config = &KafkaClusterConfigModel{}
		diag.Append(data.Config.parse(config)...)
	} else {
		data.Config = nil
	}

	if info := rs.GetConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"connection_string": types.StringType,
			"user":              types.StringType,
			"password":          types.StringType,
		},
			map[string]attr.Value{
				"connection_string": types.StringValue(info.GetConnectionString()),
				"user":              types.StringValue(info.GetUser()),
				"password":          types.StringValue(info.GetPassword()),
			},
		)
		resp.Diagnostics.Append(d...)
		data.ConnectionInfo = o
	}
	if info := rs.GetPrivateConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"connection_string": types.StringType,
			"user":              types.StringType,
			"password":          types.StringType,
		},
			map[string]attr.Value{
				"connection_string": types.StringValue(info.GetConnectionString()),
				"user":              types.StringValue(info.GetUser()),
				"password":          types.StringValue(info.GetPassword()),
			},
		)
		resp.Diagnostics.Append(d...)
		data.PrivateConnectionInfo = o
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func updateKafkaClusterRequest(m *KafkaClusterModel) (*kafka.UpdateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	rq := &kafka.UpdateClusterRequest{}
	rq.ClusterId = m.Id.ValueString()
	rq.Name = m.Name.ValueString()
	rq.Description = m.Description.ValueString()
	rq.Version = m.Version.ValueString()

	rq.Resources = &kafka.ClusterResources{
		Kafka: &kafka.ClusterResources_Kafka{
			ResourcePresetId: m.Resources.Kafka.ResourcePresetId.ValueString(),
			DiskSize:         wrapperspb.Int64(m.Resources.Kafka.DiskSize.ValueInt64()),
			BrokerCount:      wrapperspb.Int64(m.Resources.Kafka.BrokerCount.ValueInt64()),
			ZoneCount:        wrapperspb.Int64(m.Resources.Kafka.ZoneCount.ValueInt64()),
		},
	}
	if v := m.Resources.Kafka.MaxDiskSize; !v.IsNull() {
		rq.Resources.Kafka.MaxDiskSize = wrapperspb.Int64(m.Resources.Kafka.MaxDiskSize.ValueInt64())
	}

	if m.SchemaRegistry != nil {
		rq.SchemaRegistryConfig = &kafka.SchemaRegistryConfig{
			Enabled: m.SchemaRegistry.Enabled.ValueBool(),
		}
	}
	if m.Access != nil {
		access, d := m.Access.convert()
		diags.Append(d...)
		rq.Access = access
	}

	if m.Config != nil {
		config, d := m.Config.convert()
		diags.Append(d...)
		rq.KafkaConfig = config
	}

	return rq, diags
}

func (r *KafkaClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *KafkaClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := updateKafkaClusterRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.clusterService.Update(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KafkaClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *KafkaClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := deleteKafkaClusterRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.clusterService.Delete(ctx, rq)
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
}

func (r *KafkaClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func kafkaConnectionInfoResSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"connection_string": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "String to use in clients",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"user": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Apache Kafka® user",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"password": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Password for Apache Kafka® user",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
	}
}

func (m *KafkaClusterConfigModel) convert() (*kafka.KafkaConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := kafka.KafkaConfig{}
	if v := m.MessageMaxBytes; !v.IsUnknown() && v.ValueInt64() != 0 {
		r.MessageMaxBytes = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.ReplicaFetchMaxBytes; !v.IsUnknown() && v.ValueInt64() != 0 {
		r.ReplicaFetchMaxBytes = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.LogRetentionBytes; !v.IsUnknown() && v.ValueInt64() != 0 {
		r.LogRetentionBytes = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.LogRetentionHours; !v.IsUnknown() && v.ValueInt64() != 0 {
		r.LogRetentionHours = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.LogRetentionMinutes; !v.IsUnknown() && v.ValueInt64() != 0 {
		r.LogRetentionMinutes = wrapperspb.Int64(v.ValueInt64())
	}
	if v := m.LogRetentionMs; !v.IsUnknown() && v.ValueInt64() != 0 {
		r.LogRetentionMs = wrapperspb.Int64(v.ValueInt64())
	}
	return &r, diags
}

func (m *KafkaClusterConfigModel) parse(v *kafka.KafkaConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if v := v.MessageMaxBytes; v != nil {
		m.MessageMaxBytes = types.Int64Value(v.GetValue())
	}
	if v := v.ReplicaFetchMaxBytes; v != nil {
		m.ReplicaFetchMaxBytes = types.Int64Value(v.GetValue())
	}
	if v := v.LogRetentionBytes; v != nil {
		m.LogRetentionBytes = types.Int64Value(v.GetValue())
	}
	if v := v.LogRetentionHours; v != nil {
		m.LogRetentionHours = types.Int64Value(v.GetValue())
	}
	if v := v.LogRetentionMinutes; v != nil {
		m.LogRetentionMinutes = types.Int64Value(v.GetValue())
	}
	if v := v.ReplicaFetchMaxBytes; v != nil {
		m.ReplicaFetchMaxBytes = types.Int64Value(v.GetValue())
	}
	return diags
}
