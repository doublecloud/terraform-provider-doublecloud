package provider

import (
	"context"
	"fmt"
	"github.com/doublecloud/go-genproto/doublecloud/airflow/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgen "github.com/doublecloud/go-sdk/gen/airflow"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AirflowClusterResource{}
var _ resource.ResourceWithImportState = &AirflowClusterResource{}

func NewAirflowClusterResource() resource.Resource {
	return &AirflowClusterResource{}
}

type AirflowClusterResource struct {
	sdk            *dcsdk.SDK
	airflowService *dcgen.ClusterServiceClient
}

func (a *AirflowClusterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (a *AirflowClusterResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_airflow_cluster"
}

func createAirflowClusterRequest(a *AirflowClusterModel) (*airflow.CreateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	rq := &airflow.CreateClusterRequest{}
	rq.Name = a.Name.ValueString()
	rq.CloudType = a.CloudType.ValueString()
	rq.ProjectId = a.ProjectID.ValueString()
	rq.Description = a.Description.ValueString()
	rq.RegionId = a.RegionID.ValueString()
	rq.NetworkId = a.NetworkId.ValueString()

	rq.Resources = &airflow.ClusterResources{
		Airflow: &airflow.ClusterResources_Airflow{
			MaxWorkerCount:    wrapperspb.Int64(a.Resources.Airflow.MaxWorkerCount.ValueInt64()),
			EnvironmentFlavor: a.Resources.Airflow.EnvironmentFlavor.ValueString(),
			MinWorkerCount:    wrapperspb.Int64(a.Resources.Airflow.MinWorkerCount.ValueInt64()),
			WorkerConcurrency: wrapperspb.Int64(a.Resources.Airflow.WorkerConcurrency.ValueInt64()),
			WorkerDiskSize:    wrapperspb.Int64(a.Resources.Airflow.WorkerDiskSize.ValueInt64()),
			WorkerPreset:      a.Resources.Airflow.WorkerPreset.ValueString(),
		},
	}

	if a.Access != nil {
		access, d := a.Access.convert()
		diags.Append(d...)
		rq.Access = access
	}

	if a.Config != nil {
		config, d := a.Config.convert()
		diags.Append(d...)
		rq.Config = config
	}

	return rq, diags
}

func (a *AirflowClusterConfigModel) convertUpdateConfig() (*airflow.UpdateClusterRequest_UpdateAirflowConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := &airflow.UpdateClusterRequest_UpdateAirflowConfig{}

	if v := a.CustomImageDigest.ValueString(); v != "" {
		r.CustomImage = &airflow.UpdateClusterRequest_UpdateAirflowConfig_CustomImageDigest{CustomImageDigest: wrapperspb.String(v)}
	} else if v = a.ManagedRequirementsTxt.ValueString(); v != "" {
		r.CustomImage = &airflow.UpdateClusterRequest_UpdateAirflowConfig_RequirementsTxt{RequirementsTxt: wrapperspb.String(v)}
	}

	if v := a.UserServiceAccount.ValueString(); v != "" {
		r.UserServiceAccountId = wrapperspb.String(v)
	}

	if a.syncConfig != nil {
		r.GitSync = &airflow.UpdateClusterRequest_UpdateAirflowConfig_UpdateGitSyncConfig{}

		if v := a.syncConfig.RepoUrl.ValueString(); v != "" {
			r.GitSync.GetGitSync().RepoUrl = v
		} else {
			diags.AddError("Invalid Value", "RepoUrl cannot be empty")
		}

		if v := a.syncConfig.Branch.ValueString(); v != "" {
			r.GitSync.GetGitSync().Branch = v
		} else {
			diags.AddError("Invalid Value", "Branch cannot be empty")
		}

		if v := a.syncConfig.Revision.ValueString(); v != "" {
			r.GitSync.GetGitSync().Revision = v
		} else {
			diags.AddWarning("Empty Value", "Revision is not provided")
		}

		if v := a.syncConfig.DagsPath.ValueString(); v != "" {
			r.GitSync.GetGitSync().DagsPath = v
		} else {
			diags.AddError("Invalid Value", "DagsPath cannot be empty")
		}

		if a.syncConfig.credentials != nil && a.syncConfig.credentials.ApiCredentials != nil {
			creds := &airflow.SyncConfig_ApiCredentials{
				ApiCredentials: &airflow.GitApiCredentials{
					Username: a.syncConfig.credentials.ApiCredentials.Username.ValueString(),
					Password: a.syncConfig.credentials.ApiCredentials.Password.ValueString(),
				},
			}
			r.GitSync.GetGitSync().Credentials = creds
		}
	}
	if len(a.AirflowEnvVariableModel) > 0 {
		envVars := make([]*airflow.AirflowEnvVariable, len(a.AirflowEnvVariableModel))
		for i, envVarModel := range a.AirflowEnvVariableModel {
			envVars[i] = &airflow.AirflowEnvVariable{
				Name:  envVarModel.Name.ValueString(),
				Value: envVarModel.Value.ValueString(),
			}
		}
		r.EnvVars = envVars
	}

	return r, diags
}

func (a *AirflowClusterConfigModel) convert() (*airflow.CreateClusterRequest_AirflowConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := &airflow.CreateClusterRequest_AirflowConfig{}

	if v := a.VersionId.ValueString(); v == "" {
		diags.AddError("Invalid Value", "VersionId cannot be empty")
	} else {
		r.VersionId = v
	}

	if v := a.CustomImageDigest.ValueString(); v != "" {
		r.RequirementsTxt = v
	}

	if v := a.ManagedRequirementsTxt.ValueString(); v != "" {
		r.RequirementsTxt = v
	}

	if v := a.UserServiceAccount.ValueString(); v != "" {
		r.UserServiceAccountId = v
	}

	if a.syncConfig != nil {
		r.GitSync = &airflow.SyncConfig{}

		if v := a.syncConfig.RepoUrl.ValueString(); v != "" {
			r.GitSync.RepoUrl = v
		} else {
			diags.AddError("Invalid Value", "RepoUrl cannot be empty")
		}

		if v := a.syncConfig.Branch.ValueString(); v != "" {
			r.GitSync.Branch = v
		} else {
			diags.AddError("Invalid Value", "Branch cannot be empty")
		}

		if v := a.syncConfig.Revision.ValueString(); v != "" {
			r.GitSync.Revision = v
		} else {
			diags.AddWarning("Empty Value", "Revision is not provided")
		}

		if v := a.syncConfig.DagsPath.ValueString(); v != "" {
			r.GitSync.DagsPath = v
		} else {
			diags.AddError("Invalid Value", "DagsPath cannot be empty")
		}

		if a.syncConfig.credentials != nil && a.syncConfig.credentials.ApiCredentials != nil {
			creds := &airflow.SyncConfig_ApiCredentials{
				ApiCredentials: &airflow.GitApiCredentials{
					Username: a.syncConfig.credentials.ApiCredentials.Username.ValueString(),
					Password: a.syncConfig.credentials.ApiCredentials.Password.ValueString(),
				},
			}
			r.GitSync.Credentials = creds
		}
	}
	if len(a.AirflowEnvVariableModel) > 0 {
		envVars := make([]*airflow.AirflowEnvVariable, len(a.AirflowEnvVariableModel))
		for i, envVarModel := range a.AirflowEnvVariableModel {
			envVars[i] = &airflow.AirflowEnvVariable{
				Name:  envVarModel.Name.ValueString(),
				Value: envVarModel.Value.ValueString(),
			}
		}
		r.EnvVars = envVars
	}

	return r, diags
}

func (a *AirflowClusterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data *AirflowClusterModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	req, diag := createAirflowClusterRequest(data)
	if diag.HasError() {
		response.Diagnostics.Append(diag...)
	}

	rs, err := a.airflowService.Create(ctx, req)
	if err != nil {
		response.Diagnostics.AddError("failed to create", err.Error())
		return
	}

	op, err := a.sdk.WrapOperation(rs, err)
	if err != nil {
		response.Diagnostics.AddError("failed to create", err.Error())
	}

	err = op.Wait(ctx)
	if err != nil {
		response.Diagnostics.AddError("failed to create", err.Error())
	}

	data.Id = types.StringValue(op.ResourceId())

	getRq, diag := getAirflowClusterResourceRequest(data)
	if diag.HasError() {
		response.Diagnostics.Append(diag...)
		return
	}

	cluster, err := a.airflowService.Get(ctx, getRq)
	if err != nil {
		response.Diagnostics.AddError("failed to read", err.Error())
		return
	}

	if info := cluster.GetConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"host":     types.StringType,
			"user":     types.StringType,
			"password": types.StringType,
		},
			map[string]attr.Value{
				"host":     types.StringValue(info.GetHost()),
				"user":     types.StringValue(info.GetUser()),
				"password": types.StringValue(info.GetPassword()),
			},
		)
		response.Diagnostics.Append(d...)
		data.ConnectionInfo = o
	}
	if info := cluster.GetCrConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"host":     types.StringType,
			"user":     types.StringType,
			"password": types.StringType,
		},
			map[string]attr.Value{
				"host":     types.StringValue(info.GetRemoteImagePath()),
				"user":     types.StringValue(info.GetUser()),
				"password": types.StringValue(info.GetPassword()),
			},
		)
		response.Diagnostics.Append(d...)
		data.ConnectionInfo = o
	}
	tflog.Info(ctx, fmt.Sprintf("doublecloud_airflow_cluster has been created: %s", op.ResourceId()))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func getAirflowClusterResourceRequest(a *AirflowClusterModel) (*airflow.GetClusterRequest, diag.Diagnostics) {
	if a.Id == types.StringNull() {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Unknown identifier", "missed one of required fields: cluster_id or name")}
	}
	return &airflow.GetClusterRequest{
		ClusterId: a.Id.ValueString(),
		Sensitive: true,
	}, nil
}
func (a *AirflowClusterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data *AirflowClusterModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}
	rq, diag := getAirflowClusterResourceRequest(data)
	if diag.HasError() {
		response.Diagnostics.Append(diag...)
		return
	}

	rs, err := a.airflowService.Get(ctx, rq)
	if err != nil {
		response.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	data.Id = types.StringValue(rs.Id)
	data.Name = types.StringValue(rs.Name)
	data.ProjectID = types.StringValue(rs.ProjectId)
	data.Description = types.StringValue(rs.Description)
	data.CloudType = types.StringValue(rs.CloudType)
	data.RegionID = types.StringValue(rs.RegionId)
	data.NetworkId = types.StringValue(rs.NetworkId)
	data.Resources = &AirflowResourcesModel{
		Airflow: AirflowResourcesAirflowModel{
			MaxWorkerCount:    types.Int64Value(rs.GetResources().GetAirflow().GetMaxWorkerCount().GetValue()),
			MinWorkerCount:    types.Int64Value(rs.GetResources().GetAirflow().GetMinWorkerCount().GetValue()),
			EnvironmentFlavor: types.StringValue(rs.GetResources().GetAirflow().GetEnvironmentFlavor()),
			WorkerConcurrency: types.Int64Value(rs.GetResources().GetAirflow().GetWorkerConcurrency().GetValue()),
			WorkerDiskSize:    types.Int64Value(rs.GetResources().GetAirflow().GetWorkerDiskSize().GetValue()),
			WorkerPreset:      types.StringValue(rs.GetResources().GetAirflow().GetWorkerPreset()),
		},
	}

	if access := rs.GetAccess(); access != nil {
		if data.Access == nil {
			data.Access = new(AccessModel)
		}
		diag.Append(data.Access.parse(access)...)
	}

	if config := rs.GetConfig(); config != nil {
		data.Config = &AirflowClusterConfigModel{}
		diag.Append(data.Config.parse(config)...)
	} else {
		data.Config = nil
	}

	if info := rs.GetConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"host":     types.StringType,
			"user":     types.StringType,
			"password": types.StringType,
		},
			map[string]attr.Value{
				"host":     types.StringValue(info.GetHost()),
				"user":     types.StringValue(info.GetUser()),
				"password": types.StringValue(info.GetPassword()),
			},
		)
		response.Diagnostics.Append(d...)
		data.ConnectionInfo = o
	}
	if info := rs.GetCrConnectionInfo(); info != nil {
		o, d := types.ObjectValue(map[string]attr.Type{
			"host":     types.StringType,
			"user":     types.StringType,
			"password": types.StringType,
		},
			map[string]attr.Value{
				"host":     types.StringValue(info.GetRemoteImagePath()),
				"user":     types.StringValue(info.GetUser()),
				"password": types.StringValue(info.GetPassword()),
			},
		)
		response.Diagnostics.Append(d...)
		data.ConnectionInfo = o
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (a *AirflowClusterConfigModel) parse(v *airflow.AirflowConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	if v.VersionId != "" {
		a.VersionId = types.StringValue(v.VersionId)
	} else {
		diags.AddError("Invalid Value", "VersionId is missing in the response")
	}

	if v.GitSync != nil {
		a.syncConfig = &AirflowClusterSyncConfigModel{}

		if v.GitSync.RepoUrl != "" {
			a.syncConfig.RepoUrl = types.StringValue(v.GitSync.RepoUrl)
		} else {
			diags.AddError("Invalid Value", "RepoUrl is missing in the response")
		}

		if v.GitSync.Branch != "" {
			a.syncConfig.Branch = types.StringValue(v.GitSync.Branch)
		} else {
			diags.AddError("Invalid Value", "Branch is missing in the response")
		}

		if v.GitSync.Revision != "" {
			a.syncConfig.Revision = types.StringValue(v.GitSync.Revision)
		} else {
			diags.AddWarning("Empty Value", "Revision is not provided in the response")
		}

		if v.GitSync.DagsPath != "" {
			a.syncConfig.DagsPath = types.StringValue(v.GitSync.DagsPath)
		} else {
			diags.AddError("Invalid Value", "DagsPath is missing in the response")
		}

		// Handle credentials if provided
		if creds, ok := v.GitSync.Credentials.(*airflow.SyncConfig_ApiCredentials); ok {
			a.syncConfig.credentials = &Credentials{
				ApiCredentials: &GitApiCredentials{
					Username: types.StringValue(creds.ApiCredentials.Username),
					Password: types.StringValue(creds.ApiCredentials.Password),
				},
			}
		} else {
			diags.AddWarning("Missing Credentials", "No API credentials provided in the response")
		}
	}

	if v.ManagedRequirementsTxt != "" {
		a.ManagedRequirementsTxt = types.StringValue(v.ManagedRequirementsTxt)
	}

	if v.UserServiceAccountId != "" {
		a.UserServiceAccount = types.StringValue(v.UserServiceAccountId)
	}

	if v.CustomImageDigest != "" {
		a.CustomImageDigest = types.StringValue(v.CustomImageDigest)
	}

	if len(v.EnvVars) > 0 {
		envVars := make([]*AirflowEnvVariableModel, len(v.EnvVars))
		for i, envVar := range v.EnvVars {
			envVars[i] = &AirflowEnvVariableModel{
				Name:  types.StringValue(envVar.Name),
				Value: types.StringValue(envVar.Value),
			}
		}
		a.AirflowEnvVariableModel = envVars
	}

	return diags
}

func (a *AirflowClusterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data *AirflowClusterModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	rq, diag := updateAirflowClusterRequest(data)
	if diag.HasError() {
		response.Diagnostics.Append(diag...)
		return
	}

	rs, err := a.airflowService.Update(ctx, rq)
	if err != nil {
		response.Diagnostics.AddError("failed to update", err.Error())
		return
	}

	op, err := a.sdk.WrapOperation(rs, err)
	if err != nil {
		response.Diagnostics.AddError("failed to update", err.Error())
	}

	err = op.Wait(ctx)
	if err != nil {
		response.Diagnostics.AddError("failed to update", err.Error())
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func updateAirflowClusterRequest(a *AirflowClusterModel) (*airflow.UpdateClusterRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	rq := &airflow.UpdateClusterRequest{}
	rq.ClusterId = a.Id.ValueString()
	rq.Description = wrapperspb.String(a.Description.ValueString())

	rq.Resources = &airflow.UpdateClusterRequest_UpdateClusterResources{
		Airflow: &airflow.UpdateClusterRequest_UpdateClusterResources_Airflow{
			MaxWorkerCount:    wrapperspb.Int64(a.Resources.Airflow.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:    wrapperspb.Int64(a.Resources.Airflow.MaxWorkerCount.ValueInt64()),
			WorkerConcurrency: wrapperspb.Int64(a.Resources.Airflow.WorkerConcurrency.ValueInt64()),
			WorkerDiskSize:    wrapperspb.Int64(a.Resources.Airflow.WorkerDiskSize.ValueInt64()),
			WorkerPreset:      wrapperspb.String(a.Resources.Airflow.MaxWorkerCount.String()),
		},
	}
	if a.Access != nil {
		access, d := a.Access.convert()
		diags.Append(d...)
		rq.Access = access
	}

	if a.Config != nil {
		config, d := a.Config.convertUpdateConfig()
		diags.Append(d...)
		rq.Config = config
	}

	return rq, diags
}

func deleteAirflowClusterRequest(a *AirflowClusterModel) (*airflow.DeleteClusterRequest, diag.Diagnostics) {
	rq := &airflow.DeleteClusterRequest{ClusterId: a.Id.ValueString()}
	return rq, nil
}

func (a *AirflowClusterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data *AirflowClusterModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	rq, diag := deleteAirflowClusterRequest(data)
	if diag.HasError() {
		response.Diagnostics.Append(diag...)
		return
	}
	rs, err := a.airflowService.Delete(ctx, rq)
	if err != nil {
		response.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	op, err := a.sdk.WrapOperation(rs, err)
	if err != nil {
		response.Diagnostics.AddError("failed to delete", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		response.Diagnostics.AddError("failed to delete", err.Error())
	}
}

type AirflowClusterModel struct {
	Id               types.String               `tfsdk:"id"`
	ProjectID        types.String               `tfsdk:"project_id"`
	CloudType        types.String               `tfsdk:"cloud_type"`
	RegionID         types.String               `tfsdk:"region_id"`
	Name             types.String               `tfsdk:"name"`
	Description      types.String               `tfsdk:"description"`
	NetworkId        types.String               `tfsdk:"network_id"`
	Resources        *AirflowResourcesModel     `tfsdk:"resources"`
	ConnectionInfo   types.Object               `tfsdk:"connection_info"`
	CrConnectionInfo types.Object               `tfsdk:"custom_remote_connection_info"`
	Access           *AccessModel               `tfsdk:"access"`
	Config           *AirflowClusterConfigModel `tfsdk:"config"`
}

type AirflowResourcesModel struct {
	Airflow AirflowResourcesAirflowModel `tfsdk:"airflow"`
}

type AirflowResourcesAirflowModel struct {
	MaxWorkerCount    types.Int64  `tfsdk:"max_worker_count"`
	MinWorkerCount    types.Int64  `tfsdk:"min_worker_count"`
	EnvironmentFlavor types.String `tfsdk:"environment_flavor"`
	WorkerConcurrency types.Int64  `tfsdk:"worker_concurrency"`
	WorkerDiskSize    types.Int64  `tfsdk:"worker_disk_size"`
	WorkerPreset      types.String `tfsdk:"worker_preset"`
}

type AirflowClusterConfigModel struct {
	VersionId               types.String                   `tfsdk:"version_id"`
	syncConfig              *AirflowClusterSyncConfigModel `tfsdk:"sync_config"`
	CustomImageDigest       types.String                   `tfsdk:"custom_image_digest"`
	ManagedRequirementsTxt  types.String                   `tfsdk:"managed_requirements_txt"`
	UserServiceAccount      types.String                   `tfsdk:"user_service_account"`
	AirflowEnvVariableModel []*AirflowEnvVariableModel     `tfsdk:"airflow_env_variable"`
}

type AirflowClusterSyncConfigModel struct {
	RepoUrl     types.String `tfsdk:"repo_url"`
	Branch      types.String `tfsdk:"branch"`
	Revision    types.String `tfsdk:"revision"`
	DagsPath    types.String `tfsdk:"dags_path"`
	credentials *Credentials `tfsdk:"credentials"`
}

type Credentials struct {
	ApiCredentials *GitApiCredentials `tfsdk:"credentials"`
}

type GitApiCredentials struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type AirflowEnvVariableModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (a AirflowClusterResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloud provider (`aws`, `gcp`, or `azure`)",
				Validators:          []validator.String{cloudTypeValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Cluster description",
				Default:             stringdefault.StaticString(""),
			},
			"region_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Region where the cluster is located",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster network",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"connection_info": schema.SingleNestedAttribute{
				Computed:            true,
				Attributes:          airflowConnectionInfoResSchema(),
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Public connection info",
			},
			//"custom_remote_connection_info": schema.SingleNestedAttribute{
			//	Computed:            true,
			//	Attributes:          airflowCustomRemoteConnectionInfoResSchema(),
			//	PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			//	MarkdownDescription: "Remote container registry connection information.",
			//},
		},
		Blocks: map[string]schema.Block{
			"resources": schema.SingleNestedBlock{
				Description: "Cluster resources",
				Blocks: map[string]schema.Block{
					"airflow": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"max_worker_count": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Maximum number of Airflow workers.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
									int64validator.AtMost(10),
								},
							},
							"min_worker_count": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Minimum number of Airflow workers.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
									int64validator.AtMost(10),
								},
							},
							"environment_flavor": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Environment flavor for Airflow.",
								Validators: []validator.String{
									stringvalidator.OneOf("small", "medium", "large"),
								},
							},
							"worker_concurrency": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Concurrency level for Airflow workers.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
									int64validator.AtMost(30),
								},
							},
							"worker_disk_size": schema.Int64Attribute{
								Required:            true,
								MarkdownDescription: "Disk size for Airflow workers.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
									int64validator.AtMost(10),
								},
							},
							"worker_preset": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Worker preset configuration.",
								Validators: []validator.String{
									stringvalidator.OneOf("small", "medium", "large"),
								},
							},
						},
					},
				},
			},
			"access": AccessSchemaBlock(),
			"config": schema.SingleNestedBlock{
				Description: "Cluster configuration",
				Attributes: map[string]schema.Attribute{
					"version_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Version ID of the Airflow cluster.",
					},
					"custom_image_digest": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Custom image digest for the Airflow cluster.",
					},
					"managed_requirements_txt": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Path to the managed `requirements.txt` file.",
					},
					"user_service_account": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "User service account for the Airflow cluster.",
					},
				},
				Blocks: map[string]schema.Block{
					"airflow_env_variable": schema.ListNestedBlock{
						Description: "Environment variables for the Airflow cluster.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Name of the environment variable.",
								},
								"value": schema.StringAttribute{
									Required:            true,
									MarkdownDescription: "Value of the environment variable.",
								},
							},
						},
					},
					"sync_config": schema.SingleNestedBlock{
						Description: "Synchronization configuration for the Airflow cluster.",
						Attributes: map[string]schema.Attribute{
							"repo_url": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Repository URL for DAGs.",
							},
							"branch": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Branch name for the DAGs repository.",
							},
							"revision": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Revision (commit hash) for the DAGs repository.",
							},
							"dags_path": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Path to DAGs in the repository.",
							},
						},
						Blocks: map[string]schema.Block{
							"credentials": schema.SingleNestedBlock{
								Description: "Credentials for the DAGs repository.",
								Blocks: map[string]schema.Block{
									"api_credentials": schema.SingleNestedBlock{
										Description: "API credentials for accessing the DAGs repository.",
										Attributes: map[string]schema.Attribute{
											"username": schema.StringAttribute{
												Required:            true,
												MarkdownDescription: "Username for API credentials.",
											},
											"password": schema.StringAttribute{
												Required:            true,
												Sensitive:           true,
												MarkdownDescription: "Password for API credentials.",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		MarkdownDescription: "Airflow Cluster resource",
		Version:             0,
	}
}

func airflowConnectionInfoResSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"host": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "host to use in clients",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"user": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Airflow user",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"password": schema.StringAttribute{
			Computed:            true,
			Sensitive:           true,
			MarkdownDescription: "Password for the Airflow user",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
	}
}

func airflowCustomRemoteConnectionInfoResSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"remote_image_path": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "host to use in clients",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"user": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Airflow user",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"password": schema.StringAttribute{
			Computed:            true,
			Sensitive:           true,
			MarkdownDescription: "Password for the Airflow user",
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
	}
}
