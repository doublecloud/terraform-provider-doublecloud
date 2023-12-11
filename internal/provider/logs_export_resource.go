package provider

import (
	"context"
	"fmt"
	"github.com/doublecloud/go-genproto/doublecloud/logs/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dclogs "github.com/doublecloud/go-sdk/gen/logs"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LogsExportResource{}
var _ resource.ResourceWithImportState = &LogsExportResource{}
var _ resource.ResourceWithConfigure = &LogsExportResource{}

func NewLogsExportResource() resource.Resource {
	return &LogsExportResource{}
}

type LogsExportResource struct {
	sdk               *dcsdk.SDK
	logsExportService *dclogs.ExportServiceClient
}

type LogsExportResourceModel struct {
	ID          types.String                    `tfsdk:"id"`
	ProjectID   types.String                    `tfsdk:"project_id"`
	Name        types.String                    `tfsdk:"name"`
	Description types.String                    `tfsdk:"description"`
	Sources     []*logExportSourceResourceModel `tfsdk:"sources"`

	S3      *s3LogsExportResourceModel             `tfsdk:"s3"`
	Datadog *datadogLogsExportNetworkResourceModel `tfsdk:"datadog"`
}

func (m *LogsExportResourceModel) FromProtobuf(nc *logs.LogsExport) error {
	m.ID = types.StringValue(nc.GetId())
	m.ProjectID = types.StringValue(nc.GetProjectId())
	m.Name = types.StringValue(nc.GetName())
	m.Description = types.StringValue(nc.GetDescription())
	m.Sources = []*logExportSourceResourceModel{}
	for _, s := range nc.GetSources() {
		m.Sources = append(m.Sources, &logExportSourceResourceModel{
			Type: types.StringValue(s.GetType().String()),
			ID:   types.StringValue(s.GetId()),
		})
	}
	switch v := nc.Target.Target.(type) {
	case *logs.LogsTarget_S3:
		m.S3 = &s3LogsExportResourceModel{
			Bucket:             types.StringValue(v.S3.Bucket),
			BucketLayout:       types.StringValue(v.S3.BucketLayout),
			AWSAccessKeyID:     types.StringValue(v.S3.AwsAccessKeyId),
			AWSSecretAccessKey: types.StringValue(v.S3.AwsSecretAccessKey),
			Region:             types.StringValue(v.S3.Region),
			Endpoint:           types.StringValue(v.S3.Endpoint),
			DisableSSL:         types.BoolValue(v.S3.DisableSsl),
			SkipVerifySSLCert:  types.BoolValue(v.S3.SkipVerifySslCert),
		}
	case *logs.LogsTarget_Datadog:
		m.Datadog = &datadogLogsExportNetworkResourceModel{
			APIKey:      types.StringValue(v.Datadog.ApiKey),
			DatadogHost: types.StringValue(v.Datadog.DatadogHost.String()),
		}
	default:
		return fmt.Errorf("unknown type: %T", nc.Target.Target)
	}
	return nil
}

type logExportSourceResourceModel struct {
	Type types.String `tfsdk:"type"`
	ID   types.String `tfsdk:"id"`
}

type s3LogsExportResourceModel struct {
	Bucket             types.String `tfsdk:"bucket"`
	BucketLayout       types.String `tfsdk:"bucket_layout"`
	AWSAccessKeyID     types.String `tfsdk:"aws_access_key_id"`
	AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	Region             types.String `tfsdk:"region"`
	Endpoint           types.String `tfsdk:"endpoint"`
	DisableSSL         types.Bool   `tfsdk:"disable_ssl"`
	SkipVerifySSLCert  types.Bool   `tfsdk:"skip_verify_ssl_cert"`
}

type datadogLogsExportNetworkResourceModel struct {
	APIKey      types.String `tfsdk:"api_key"`
	DatadogHost types.String `tfsdk:"datadog_host"`
}

func (l *LogsExportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logs_export"
}

func (l *LogsExportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Log export ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Log export name",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Log export description",
				Default:             stringdefault.StaticString(""),
			},
			"sources": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Type of log export source",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
							Validators:          []validator.String{protoEnumValidator(logs.LogSourceType_name)},
						},
						"id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Resource ID",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
					},
				},
			},
			"s3": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Name of the S3 bucket to export logs to",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"bucket_layout": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Folder where logs will be exported. Can include the date as a template variable in the Go date format, such as \"2006/01/02/some_folder\"",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"aws_access_key_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Access key ID",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"aws_secret_access_key": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Secret access key",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"region": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Region where the bucket is located",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"endpoint": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Endpoint of the S3-compatible service. Leave blank if you're using AWS.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"skip_verify_ssl_cert": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Skip verifying SSL certificate. Enable if the bucket allows self-signed certificates",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"disable_ssl": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Allow connections without SSL. Enable if you're connecting to an S3-compatible service that doesn't use SSL/TLS",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
				},
				Validators: []validator.Object{},
			},
			"datadog": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Datadog export target",
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Datadog API Key",
					},
					"datadog_host": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Datadog site. Make sure to specify the correct site because Datadog sites are independent and data isn't shared across them by default",
						Validators:          []validator.String{protoEnumValidator(logs.LogsTargetDatadog_DatadogHost_name)},
					},
				},
				Validators: []validator.Object{},
			},
		},
	}
}

func (l *LogsExportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	l.sdk = sdk
	l.logsExportService = l.sdk.Logs().Export()
}

func (l *LogsExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *LogsExportResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &logs.CreateExportRequest{
		ProjectId:   data.ProjectID.ValueString(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Sources:     nil,
		Target:      nil,
	}
	for _, s := range data.Sources {
		createReq.Sources = append(createReq.Sources, &logs.LogSource{
			Type: logs.LogSourceType(logs.LogSourceType_value[s.Type.ValueString()]),
			Id:   s.ID.ValueString(),
		})
	}
	switch {
	case data.S3 != nil:
		createReq.Target = &logs.LogsTarget{Target: &logs.LogsTarget_S3{S3: &logs.LogsTargetS3{
			Bucket:             data.S3.Bucket.ValueString(),
			BucketLayout:       data.S3.BucketLayout.ValueString(),
			AwsAccessKeyId:     data.S3.AWSAccessKeyID.ValueString(),
			AwsSecretAccessKey: data.S3.AWSSecretAccessKey.ValueString(),
			Region:             data.S3.Region.ValueString(),
			Endpoint:           data.S3.Endpoint.ValueString(),
			DisableSsl:         data.S3.DisableSSL.ValueBool(),
			SkipVerifySslCert:  data.S3.SkipVerifySSLCert.ValueBool(),
		}}}
	case data.Datadog != nil:
		createReq.Target = &logs.LogsTarget{Target: &logs.LogsTarget_Datadog{Datadog: &logs.LogsTargetDatadog{
			ApiKey:      data.Datadog.APIKey.ValueString(),
			DatadogHost: logs.LogsTargetDatadog_DatadogHost(logs.LogsTargetDatadog_DatadogHost_value[data.Datadog.DatadogHost.ValueString()]),
		}}}
	default:
		resp.Diagnostics.AddError("misconfiguration", "at least one of \"s3\" or \"datadog\" must be specified")
		return
	}
	opObj, err := l.logsExportService.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}

	data.ID = types.StringValue(opObj.GetResourceId())

	resp.Diagnostics.Append(getLogsExport(ctx, l.logsExportService, opObj.GetResourceId(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (l *LogsExportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *LogsExportResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(getLogsExport(ctx, l.logsExportService, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getLogsExport(
	ctx context.Context,
	client *dclogs.ExportServiceClient,
	id string,
	data *LogsExportResourceModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	nc, err := client.Get(ctx, &logs.GetExportRequest{Id: id})
	if err != nil {
		diags.AddError("Failed to get network connection", fmt.Sprintf("failed request, error: %v", err))
		return diags
	}

	if err = data.FromProtobuf(nc); err != nil {
		diags.AddError("Failed to get network connection", fmt.Sprintf("failed parse, error: %v", err))
		return diags
	}

	return diags
}

func (l *LogsExportResource) Update(ctx context.Context, request resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Failed to update logs export", "logs expport don't support updates")
}

func (l *LogsExportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *LogsExportResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := l.logsExportService.Delete(ctx, &logs.DeleteExportRequest{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
}

func (l *LogsExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
