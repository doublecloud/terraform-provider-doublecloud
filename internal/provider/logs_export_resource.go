package provider

import (
	"context"
	"fmt"
	"github.com/doublecloud/go-genproto/doublecloud/logs/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dclogs "github.com/doublecloud/go-sdk/gen/logs"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

func NewLogsExportResource() resource.Resource {
	return &LogsExportResource{}
}

type LogsExportResource struct {
	sdk               *dcsdk.SDK
	logsExportService *dclogs.ExportServiceClient
}

type LogsExportResourceModel struct {
	Id          types.String                    `tfsdk:"id"`
	ProjectID   types.String                    `tfsdk:"project_id"`
	Name        types.String                    `tfsdk:"name"`
	Description types.String                    `tfsdk:"description"`
	Sources     []*logExportSourceResourceModel `tfsdk:"sources"`

	S3      *s3LogsExportResourceModel             `tfsdk:"s3"`
	Datadog *datadogLogsExportNetworkResourceModel `tfsdk:"datadog"`
}

type logExportSourceResourceModel struct {
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

func (l LogsExportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logs_export"
}

func (l LogsExportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Logs Export identifier",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project identifier",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of logs export",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of logs export",
				Default:             stringdefault.StaticString(""),
			},
			"s3": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Bucket.\nName of the S3 bucket to export logs to.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"bucket_layout": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Folder name\nFolder where logs will be exported. \n Can include the date as a template variable in the Go date format, such as \"2006/01/02/some_folder\".",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"aws_access_key_id": schema.StringAttribute{
						Required:            false,
						MarkdownDescription: "Access key ID",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"aws_secret_access_key": schema.StringAttribute{
						Required:            false,
						MarkdownDescription: "Secret access key",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"region": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Region\nRegion where the bucket is located.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"endpoint": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Endpoint\nEndpoint of the S3-compatible service. Leave blank if you're using AWS.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"skip_verify_ssl_cert": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Skip verifying SSL certificate\nSelect if the bucket allows self-signed certificates.",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"disable_ssl": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Allow connections without SSL\nSelect if you're connecting to an S3-compatible service that doesn't use SSL/TLS.",
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
						MarkdownDescription: "API Key for Datadog",
					},
					"datadog_host": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Host name for Datadog",
						Validators:          []validator.String{datadogHostOptions()},
					},
				},
				Validators: []validator.Object{},
			},
		},
	}
}

func datadogHostOptions() validator.String {
	names := make([]string, len(logs.LogsTargetDatadog_DatadogHost_name))
	for i, v := range logs.LogsTargetDatadog_DatadogHost_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func (l LogsExportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (l LogsExportResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (l LogsExportResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (l LogsExportResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (l LogsExportResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

func (l LogsExportResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	//TODO implement me
	panic("implement me")
}
