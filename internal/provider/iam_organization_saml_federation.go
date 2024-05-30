package provider

import (
	"context"
	"fmt"
	"github.com/doublecloud/go-genproto/doublecloud/organizationmanager/saml/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcorganization "github.com/doublecloud/go-sdk/gen/organization"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IAMOrganizationSamlFederation{}
var _ resource.ResourceWithImportState = &IAMOrganizationSamlFederation{}
var _ resource.ResourceWithConfigure = &IAMOrganizationSamlFederation{}

func NewIAMOrganizationSamlFederation() resource.Resource {
	return &IAMOrganizationSamlFederation{}
}

type IAMOrganizationSamlFederation struct {
	sdk                 *dcsdk.SDK
	organizationService *dcorganization.Organization
}

type IAMOrganizationSamlFederationModel struct {
	ID                       types.String `tfsdk:"id"`
	OrganizationID           types.String `tfsdk:"organization_id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	CookieMaxAge             types.String `tfsdk:"cookie_max_age"`
	AutoCreateAccountOnLogin types.Bool   `tfsdk:"auto_create_account_on_login"`
	Issuer                   types.String `tfsdk:"issuer"`
	SsoBinding               types.String `tfsdk:"sso_binding"`
	SsoUrl                   types.String `tfsdk:"sso_url"`
	CaseInsensitiveNameIds   types.Bool   `tfsdk:"case_insensitive_name_ids"`
}

func (m *IAMOrganizationSamlFederationModel) FromProtobuf(nc *saml.Federation) error {
	m.ID = types.StringValue(nc.GetId())
	m.OrganizationID = types.StringValue(nc.GetOrganizationId())
	m.Name = types.StringValue(nc.GetName())
	m.Description = types.StringValue(nc.GetDescription())
	m.CookieMaxAge = types.StringValue(nc.GetCookieMaxAge().AsDuration().String())
	m.AutoCreateAccountOnLogin = types.BoolValue(nc.GetAutoCreateAccountOnLogin())
	m.Issuer = types.StringValue(nc.GetIssuer())
	m.SsoBinding = types.StringValue(nc.GetSsoBinding().String())
	m.SsoUrl = types.StringValue(nc.GetSsoUrl())
	m.CaseInsensitiveNameIds = types.BoolValue(nc.GetCaseInsensitiveNameIds())

	return nil
}

func (l *IAMOrganizationSamlFederation) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saml_federation"
}

func (l *IAMOrganizationSamlFederation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SAML Federation resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Group ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization ID",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organization group name",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Organization group description",
				Default:             stringdefault.StaticString(""),
			},
			"cookie_max_age": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Max age for cookies in federation",
				Default:             stringdefault.StaticString("12h"),
			},
			"auto_create_account_on_login": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable auto creation of accounts on login",
				Default:             booldefault.StaticBool(true),
			},
			"issuer": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Enable auto creation of accounts on login",
				Default:             stringdefault.StaticString(""),
			},
			"sso_binding": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "SSO Binding for federation",
				Default:             stringdefault.StaticString(""),
			},
			"sso_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "URL for SSO",
				Default:             stringdefault.StaticString(""),
			},
			"case_insensitive_name_ids": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Should use insensitive name ids",
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (l *IAMOrganizationSamlFederation) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	l.organizationService = l.sdk.Organization()
}

func (l *IAMOrganizationSamlFederation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IAMOrganizationSamlFederationModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	maxAge, err := time.ParseDuration(data.CookieMaxAge.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "failed to parse cookie_max_age", err.Error())
	}

	createReq := &saml.CreateFederationRequest{
		OrganizationId:           data.OrganizationID.ValueString(),
		Name:                     data.Name.ValueString(),
		Description:              data.Description.ValueString(),
		CookieMaxAge:             durationpb.New(maxAge),
		AutoCreateAccountOnLogin: data.AutoCreateAccountOnLogin.ValueBool(),
		Issuer:                   data.Issuer.ValueString(),
		SsoBinding:               saml.BindingType(saml.BindingType_value[data.SsoBinding.ValueString()]),
		SsoUrl:                   data.SsoUrl.ValueString(),
		SecuritySettings:         nil,
		CaseInsensitiveNameIds:   data.CaseInsensitiveNameIds.ValueBool(),
	}
	opObj, err := l.organizationService.SamlFederation().Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}

	data.ID = types.StringValue(opObj.GetResourceId())

	resp.Diagnostics.Append(getSamlFederation(ctx, l.organizationService, opObj.GetResourceId(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (l *IAMOrganizationSamlFederation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *IAMOrganizationSamlFederationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(getSamlFederation(ctx, l.organizationService, data.ID.ValueString(), data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getSamlFederation(
	ctx context.Context,
	client *dcorganization.Organization,
	id string,
	data *IAMOrganizationSamlFederationModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	nc, err := client.SamlFederation().Get(ctx, &saml.GetFederationRequest{FederationId: id})
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

func (l *IAMOrganizationSamlFederation) Update(ctx context.Context, request resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Failed to update logs export", "logs expport don't support updates")
}

func (l *IAMOrganizationSamlFederation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IAMOrganizationSamlFederationModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := l.organizationService.SamlFederation().Delete(ctx, &saml.DeleteFederationRequest{FederationId: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
}

func (l *IAMOrganizationSamlFederation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
