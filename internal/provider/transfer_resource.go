package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgentf "github.com/doublecloud/go-sdk/gen/transfer"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TransferResource{}
var _ resource.ResourceWithImportState = &TransferResource{}

func NewTransferResource() resource.Resource {
	return &TransferResource{}
}

type TransferResource struct {
	sdk *dcsdk.SDK
	// endpointService *dcgentf.EndpointServiceClient
	transferService *dcgentf.TransferServiceClient
}

func (r *TransferResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transfer"
}

func (r *TransferResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Transfer resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Transfer id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of transfer",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description",
				Default:             stringdefault.StaticString(""),
			},
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Transfer type",
				Validators:          []validator.String{transferTypeValidator()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Source endpoint_id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target endpoint_id",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"activated": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Activation of transfer",
			},
			"transformation": transferTransformationSchema(),
			"runtime":        transferRuntimeSchema(),
		},
	}
}

func (r *TransferResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.transferService = r.sdk.Transfer().Transfer()
}

func (r *TransferResource) setActivation(ctx context.Context, m *transferResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	var dcOp *doublecloud.Operation
	var err error

	if m.Activated.IsNull() {
		return diags
	}

	if m.Activated.ValueBool() {
		dcOp, err = r.transferService.Activate(ctx, &transfer.ActivateTransferRequest{
			TransferId: m.Id.ValueString(),
		})
	} else {
		dcOp, err = r.transferService.Deactivate(ctx, &transfer.DeactivateTransferRequest{
			TransferId: m.Id.ValueString(),
		})
	}

	if err != nil {
		diags.AddError("failed to activate", err.Error())
		return diags
	}
	op, err := r.sdk.WrapOperation(dcOp, err)
	if err != nil {
		diags.AddError("failed to activate", err.Error())
		return diags
	}
	err = op.Wait(ctx)
	if err != nil {
		diags.AddError("failed to activate", err.Error())
	}
	return diags
}

func (r *TransferResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *transferResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := data.CreateRequest()
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.transferService.Create(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to call Create", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to wrap Create operation", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to Create", err.Error())
	}

	data.Id = types.StringValue(op.ResourceId())

	r.setActivation(ctx, data)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *transferResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := r.transferService.Get(ctx, &transfer.GetTransferRequest{
		TransferId: data.Id.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}
	diags := data.parse(rs)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *transferResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	rq, diag := data.UpdateRequest()
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	rs, err := r.transferService.Update(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to call Update", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to wrap Update operation", err.Error())
		return
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to Update", err.Error())
		return
	}

	r.setActivation(ctx, data)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TransferResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *transferResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := data.DeleteRequest()
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.transferService.Delete(ctx, rq)
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
		resp.Diagnostics.AddError("failed to wait for delete completion", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted endpoint: %s", data.Id))
}

func (r *TransferResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func transferTypeValidator() validator.String {
	names := make([]string, len(transfer.TransferType_name))
	for i, v := range transfer.TransferType_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

type transferResourceModel struct {
	Id             types.String            `tfsdk:"id"`
	ProjectID      types.String            `tfsdk:"project_id"`
	Name           types.String            `tfsdk:"name"`
	Description    types.String            `tfsdk:"description"`
	Source         types.String            `tfsdk:"source"`
	Target         types.String            `tfsdk:"target"`
	Type           types.String            `tfsdk:"type"`
	Activated      types.Bool              `tfsdk:"activated"`
	Transformation *transferTransformation `tfsdk:"transformation"`
	Runtime        *transferRuntime        `tfsdk:"runtime"`
}

type requestType int

const (
	requestTypeCreate requestType = 0
	requestTypeUpdate requestType = 1
)

func (m *transferResourceModel) CreateRequest() (*transfer.CreateTransferRequest, diag.Diagnostics) {
	r := new(transfer.CreateTransferRequest)
	var diags diag.Diagnostics

	r.Name = m.Name.ValueString()
	r.Description = m.Description.ValueString()
	r.ProjectId = m.ProjectID.ValueString()
	r.SourceId = m.Source.ValueString()
	r.TargetId = m.Target.ValueString()
	r.Type = transfer.TransferType(transfer.TransferType_value[m.Type.ValueString()])
	if m.Transformation != nil {
		r.Transformation = new(transfer.Transformation)
		diags.Append(m.Transformation.convert(requestTypeCreate, r.Transformation)...)
	}
	if m.Runtime != nil && m.Runtime.Dedicated != nil {
		settings := &transfer.Settings{Settings: &transfer.Settings_AutoSettings{AutoSettings: &transfer.AutoSettings{}}}
		if m.Runtime.Dedicated.VPCID.ValueString() != "" {
			settings = &transfer.Settings{Settings: &transfer.Settings_ManualSettings{ManualSettings: &transfer.ManualSettings{
				NetworkId: m.Runtime.Dedicated.VPCID.ValueString(),
			}}}
		}
		r.Runtime = &transfer.Runtime{Runtime: &transfer.Runtime_DedicatedRuntime{DedicatedRuntime: &transfer.DedicatedRuntime{
			Flavor:   transfer.Flavor(transfer.Flavor_value[m.Runtime.Dedicated.Flavor.ValueString()]),
			Settings: settings,
		}}}
	}

	return r, diags
}

func (m *transferResourceModel) parse(t *transfer.Transfer) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Id = types.StringValue(t.GetId())
	m.ProjectID = types.StringValue(t.GetProjectId())
	m.Name = types.StringValue(t.GetName())
	m.Description = types.StringValue(t.GetDescription())
	m.Source = types.StringValue(t.GetSource().GetId())
	m.Target = types.StringValue(t.GetTarget().GetId())
	m.Type = types.StringValue(t.GetType().String())

	if t.GetTransformation() != nil && len(t.GetTransformation().GetTransformers()) > 0 {
		if m.Transformation == nil {
			m.Transformation = new(transferTransformation)
		}
		diags.Append(m.Transformation.parse(t.Transformation)...)
	} else {
		m.Transformation = nil
	}
	if t.GetRuntime().GetDedicatedRuntime() != nil {
		m.Runtime = new(transferRuntime)
		m.Runtime.Dedicated = new(transferDedicatedRuntime)
		m.Runtime.Dedicated.VPCID = types.StringValue(t.GetRuntime().GetDedicatedRuntime().GetSettings().GetManualSettings().GetNetworkId())
		m.Runtime.Dedicated.Flavor = types.StringValue(t.GetRuntime().GetDedicatedRuntime().Flavor.String())
	} else {
		m.Runtime = nil
	}

	return diags
}

func (m *transferResourceModel) UpdateRequest() (*transfer.UpdateTransferRequest, diag.Diagnostics) {
	r := new(transfer.UpdateTransferRequest)
	var diags diag.Diagnostics

	r.TransferId = m.Id.ValueString()
	r.Name = m.Name.ValueString()
	r.Description = m.Description.ValueString()
	if m.Transformation != nil {
		r.Transformation = new(transfer.Transformation)
		diags.Append(m.Transformation.convert(requestTypeUpdate, r.Transformation)...)
	}

	return r, diags
}

func (m *transferResourceModel) DeleteRequest() (*transfer.DeleteTransferRequest, diag.Diagnostics) {
	return &transfer.DeleteTransferRequest{
		TransferId: m.Id.ValueString(),
	}, nil
}

type transferTransformation struct {
	Transformers []transferTransformer `tfsdk:"transformers"`
}

type transferRuntime struct {
	Dedicated *transferDedicatedRuntime `tfsdk:"dedicated"`
}

type transferDedicatedRuntime struct {
	VPCID  types.String `tfsdk:"vpc_id"`
	Flavor types.String `tfsdk:"flavor"`
}

func transferTransformationSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"transformers": transferTransformerSchema(),
		},
		Optional: true,
	}
}

func transferRuntimeSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"dedicated": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"vpc_id": schema.StringAttribute{Optional: true},
					"flavor": schema.StringAttribute{Required: true, Validators: []validator.String{transferRuntimeFlavorValidator()}},
				},
				Optional: true,
			},
		},
		Optional: true,
	}
}

func transferRuntimeFlavorValidator() validator.String {
	names := make([]string, len(transfer.Flavor_name))
	for i, v := range transfer.Flavor_name {
		names[i] = v
	}
	return stringvalidator.OneOfCaseInsensitive(names...)
}

func (m *transferTransformation) convert(rqt requestType, r *transfer.Transformation) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(m.Transformers) > 0 {
		r.Transformers = make([]*transfer.Transformer, len(m.Transformers))
		for i := range m.Transformers {
			r.Transformers[i] = new(transfer.Transformer)
			diags.Append(m.Transformers[i].convert(rqt, r.Transformers[i])...)
		}
	}

	return diags
}

func (m *transferTransformation) parse(t *transfer.Transformation) diag.Diagnostics {
	var diags diag.Diagnostics

	tTransformers := t.GetTransformers()
	if len(tTransformers) > 0 {
		if len(m.Transformers) == 0 {
			m.Transformers = make([]transferTransformer, 0)
		}
		for i := range tTransformers {
			if i >= len(m.Transformers) {
				m.Transformers = append(m.Transformers, *new(transferTransformer))
			}
			diags.Append(m.Transformers[i].parse(tTransformers[i])...)
		}
	} else {
		m.Transformers = nil
	}

	return diags
}

type transferTransformer struct {
	ReplacePrimaryKey *transferTransformerReplacePrimaryKey `tfsdk:"replace_primary_key"`
	ConvertToString   *transferTransformerConvertToString   `tfsdk:"convert_to_string"`
	DBT               *transferTransformerDBT               `tfsdk:"dbt"`
	TableSplitter     *transferTransformerTableSplitter     `tfsdk:"table_splitter"`
}

func transferTransformerSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"replace_primary_key": transferTransformerReplacePrimaryKeySchema(),
				"convert_to_string":   transferTransformerConvertToStringSchema(),
				"dbt":                 transferTransformerDBTSchema(),
				"table_splitter":      transferTransformerTableSplitterSchema(),
			},
		},
		Optional: true,
	}
}

func (m *transferTransformer) convert(rqt requestType, r *transfer.Transformer) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case m.ReplacePrimaryKey != nil:
		tr := new(transfer.ReplacePrimaryKeyTransformer)
		diags.Append(m.ReplacePrimaryKey.convert(rqt, tr)...)
		r.Transformer = &transfer.Transformer_ReplacePrimaryKey{ReplacePrimaryKey: tr}
	case m.ConvertToString != nil:
		tr := new(transfer.ToStringTransformer)
		diags.Append(m.ConvertToString.convert(rqt, tr)...)
		r.Transformer = &transfer.Transformer_ConvertToString{ConvertToString: tr}
	case m.DBT != nil:
		tr := new(transfer.DBTTransformer)
		diags.Append(m.DBT.convert(rqt, tr)...)
		r.Transformer = &transfer.Transformer_Dbt{Dbt: tr}
	case m.TableSplitter != nil:
		tr := new(transfer.TableSplitterTransformer)
		m.TableSplitter.convert(rqt, tr)
		r.Transformer = &transfer.Transformer_TableSplitterTransformer{TableSplitterTransformer: tr}
	default:
		diags.Append(diag.NewErrorDiagnostic("a transformer is present, but not set to any oneof value", ""))
	}

	return diags
}

func (m *transferTransformer) parse(t *transfer.Transformer) diag.Diagnostics {
	var diags diag.Diagnostics

	switch {
	case t.GetReplacePrimaryKey() != nil:
		if m.ReplacePrimaryKey == nil {
			m.clear()
			m.ReplacePrimaryKey = new(transferTransformerReplacePrimaryKey)
		}
		diags.Append(m.ReplacePrimaryKey.parse(t.GetReplacePrimaryKey())...)
	case t.GetConvertToString() != nil:
		if m.ConvertToString == nil {
			m.clear()
			m.ConvertToString = new(transferTransformerConvertToString)
		}
		diags.Append(m.ConvertToString.parse(t.GetConvertToString())...)
	case t.GetDbt() != nil:
		if m.DBT == nil {
			m.clear()
			m.DBT = new(transferTransformerDBT)
		}
		diags.Append(m.DBT.parse(t.GetDbt())...)
	case t.GetTableSplitterTransformer() != nil:
		if m.TableSplitter == nil {
			m.clear()
			m.TableSplitter = new(transferTransformerTableSplitter)
		}
		diags.Append(m.TableSplitter.parse(t.GetTableSplitterTransformer())...)
	default:
		m.clear()
	}

	return diags
}

func (m *transferTransformer) clear() {
	m.ReplacePrimaryKey = nil
	m.ConvertToString = nil
	m.DBT = nil
	m.TableSplitter = nil
}

type transferTransformerReplacePrimaryKey struct {
	Tables *transferTransformerTablesFilter `tfsdk:"tables"`
	Keys   []types.String                   `tfsdk:"keys"`
}

func transferTransformerReplacePrimaryKeySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Columns to mark as PRIMARY KEYs.",
			},
			"tables": transferTransformerTablesFilterSchema(),
		},
		Optional:            true,
		MarkdownDescription: "Replace the set of columns marked as PRIMARY KEYs.",
	}
}

func (m *transferTransformerReplacePrimaryKey) convert(rqt requestType, r *transfer.ReplacePrimaryKeyTransformer) diag.Diagnostics {
	var diags diag.Diagnostics

	r.Keys = convertSliceTFStrings(m.Keys)
	if m.Tables != nil {
		r.Tables = new(transfer.TablesFilter)
		diags.Append(m.Tables.convert(rqt, r.Tables)...)
	}

	return diags
}

func (m *transferTransformerReplacePrimaryKey) parse(t *transfer.ReplacePrimaryKeyTransformer) diag.Diagnostics {
	var diags diag.Diagnostics

	if tables := t.GetTables(); tables != nil {
		if m.Tables == nil {
			m.Tables = new(transferTransformerTablesFilter)
		}
		diags.Append(m.Tables.parse(tables)...)
	} else {
		m.Tables = nil
	}

	m.Keys = convertSliceToTFStrings(t.GetKeys())

	return diags
}

type transferTransformerTablesFilter struct {
	Include []types.String `tfsdk:"include"`
	Exclude []types.String `tfsdk:"exclude"`
}

func transferTransformerTablesFilterSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"include": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Included tables (regular expressions). Start every name with `^` and finish with `$` to avoid unexpected side effects.",
			},
			"exclude": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Excluded tables (regular expressions). Start every name with `^` and finish with `$` to avoid unexpected side effects.",
			},
		},
		Optional:            true,
		MarkdownDescription: "Tables.",
	}
}

func (m *transferTransformerTablesFilter) convert(rqt requestType, r *transfer.TablesFilter) diag.Diagnostics {
	r.IncludeTables = convertSliceTFStrings(m.Include)
	r.ExcludeTables = convertSliceTFStrings(m.Exclude)

	return nil
}

func (m *transferTransformerTablesFilter) parse(t *transfer.TablesFilter) diag.Diagnostics {
	m.Include = convertSliceToTFStrings(t.GetIncludeTables())
	m.Exclude = convertSliceToTFStrings(t.GetExcludeTables())

	return nil
}

type transferTransformerConvertToString struct {
	Tables  *transferTransformerTablesFilter  `tfsdk:"tables"`
	Columns *transferTransformerColumnsFilter `tfsdk:"columns"`
}

func transferTransformerConvertToStringSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"tables":  transferTransformerTablesFilterSchema(),
			"columns": transferTransformerColumnsFilterSchema(),
		},
		Optional:            true,
		MarkdownDescription: "Convert columns' values to strings.",
	}
}

func (m *transferTransformerConvertToString) convert(rqt requestType, r *transfer.ToStringTransformer) diag.Diagnostics {
	var diags diag.Diagnostics

	if m.Tables != nil {
		r.Tables = new(transfer.TablesFilter)
		diags.Append(m.Tables.convert(rqt, r.Tables)...)
	}
	if m.Columns != nil {
		r.Columns = new(transfer.ColumnsFilter)
		diags.Append(m.Columns.convert(rqt, r.Columns)...)
	}

	return diags
}

func (m *transferTransformerConvertToString) parse(t *transfer.ToStringTransformer) diag.Diagnostics {
	var diags diag.Diagnostics

	if tables := t.GetTables(); tables != nil {
		if m.Tables == nil {
			m.Tables = new(transferTransformerTablesFilter)
		}
		diags.Append(m.Tables.parse(tables)...)
	} else {
		m.Tables = nil
	}

	if columns := t.GetColumns(); columns != nil {
		if m.Columns == nil {
			m.Columns = new(transferTransformerColumnsFilter)
		}
		diags.Append(m.Columns.parse(columns)...)
	} else {
		m.Columns = nil
	}

	return diags
}

type transferTransformerColumnsFilter struct {
	Include []types.String `tfsdk:"include"`
	Exclude []types.String `tfsdk:"exclude"`
}

func transferTransformerColumnsFilterSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"include": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Included columns (regular expressions). Start every name with `^` and finish with `$` to avoid unexpected side effects.",
			},
			"exclude": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Excluded columns (regular expressions). Start every name with `^` and finish with `$` to avoid unexpected side effects.",
			},
		},
		Optional: true,
	}
}

func (m *transferTransformerColumnsFilter) convert(rqt requestType, r *transfer.ColumnsFilter) diag.Diagnostics {
	r.IncludeColumns = convertSliceTFStrings(m.Include)
	r.ExcludeColumns = convertSliceTFStrings(m.Exclude)

	return nil
}

func (m *transferTransformerColumnsFilter) parse(t *transfer.ColumnsFilter) diag.Diagnostics {
	m.Include = convertSliceToTFStrings(t.GetIncludeColumns())
	m.Exclude = convertSliceToTFStrings(t.GetExcludeColumns())

	return nil
}

type transferTransformerDBT struct {
	GitRepositoryLink types.String `tfsdk:"git_repository_link"`
	GitBranch         types.String `tfsdk:"git_branch"`
	ProfileName       types.String `tfsdk:"profile_name"`
	Operation         types.String `tfsdk:"operation"` // is a oneof
}

func transferTransformerDBTSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"git_repository_link": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A link to a git repository with a DBT project. Must start with `https://`. The root directory of the repository must contain a `dbt_project.yml` file.",
			},
			"git_branch": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A branch or a tag of the git repository with the DBT project.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"profile_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The name for a profile which will be created automatically using the settings of the destination endpoint. The name must match the `profile` property in the `dbt_project.yml` file.",
			},
			"operation": schema.StringAttribute{
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf(transferTransformerDBTOperationValidatorOneofValues()...)},
				MarkdownDescription: "Operation; for example, `run`.",
			},
		},
		Optional:            true,
		MarkdownDescription: "Run DBT after snapshot finish.",
	}
}

func transferTransformerDBTOperationValidatorOneofValues() []string {
	result := make([]string, 0)
	for opcode := range transfer.DBTTransformer_Operation_name {
		if opcode == int32(transfer.DBTTransformer_OPERATION_UNSPECIFIED) {
			continue
		}
		result = append(result, opcodeToOperation(transfer.DBTTransformer_Operation(opcode)))
	}
	return result
}

func opcodeToOperation(opcode transfer.DBTTransformer_Operation) string {
	result := transfer.DBTTransformer_Operation_name[int32(opcode)]
	result = strings.ToLower(strings.TrimPrefix(result, "OPERATION_"))
	return result
}

func (m *transferTransformerDBT) convert(rqt requestType, r *transfer.DBTTransformer) diag.Diagnostics {
	var diags diag.Diagnostics

	r.GitRepositoryLink = m.GitRepositoryLink.ValueString()
	r.GitBranch = m.GitBranch.ValueString()
	r.ProfileName = m.ProfileName.ValueString()

	opcode, err := operationToOpcode(m.Operation.ValueString())
	if err != nil {
		diags.AddAttributeError(path.Root("operation"), "unsupported operation type", err.Error())
		return diags
	}
	r.Operation = opcode

	return diags
}

func operationToOpcode(operation string) (transfer.DBTTransformer_Operation, error) {
	if len(operation) == 0 {
		return transfer.DBTTransformer_OPERATION_UNSPECIFIED, errors.New("DBT operation must be set")
	}

	operationEnumString := "OPERATION_" + strings.ToUpper(operation)

	opcode, opValid := transfer.DBTTransformer_Operation_value[operationEnumString]
	if !opValid {
		return transfer.DBTTransformer_OPERATION_UNSPECIFIED, fmt.Errorf("unknown DBT operation %q", operation)
	}

	return transfer.DBTTransformer_Operation(opcode), nil
}

func (m *transferTransformerDBT) parse(t *transfer.DBTTransformer) diag.Diagnostics {
	m.GitRepositoryLink = types.StringValue(t.GetGitRepositoryLink())
	if b := t.GetGitBranch(); len(b) > 0 {
		m.GitBranch = types.StringValue(b)
	} else {
		m.GitBranch = types.StringNull()
	}
	m.ProfileName = types.StringValue(t.GetProfileName())
	m.Operation = types.StringValue(opcodeToOperation(t.GetOperation()))

	return nil
}

type transferTransformerTableSplitter struct {
	Tables   *transferTransformerTablesFilter `tfsdk:"tables"`
	Columns  []types.String                   `tfsdk:"columns"`
	Splitter types.String                     `tfsdk:"splitter"`
}

func transferTransformerTableSplitterSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"columns": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Columns with values to use as a new table name.",
			},
			"splitter": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A string separating the parts of the new table name.",
			},
			"tables": transferTransformerTablesFilterSchema(),
		},
		Optional:            true,
		MarkdownDescription: "Replace the name of the table to a value composed of values of columns of a row.",
	}
}

func (m *transferTransformerTableSplitter) convert(rqt requestType, r *transfer.TableSplitterTransformer) diag.Diagnostics {
	var diags diag.Diagnostics

	r.Columns = convertSliceTFStrings(m.Columns)
	r.Splitter = m.Splitter.ValueString()
	if m.Tables != nil {
		r.Tables = new(transfer.TablesFilter)
		diags.Append(m.Tables.convert(rqt, r.Tables)...)
	}

	return diags
}

func (m *transferTransformerTableSplitter) parse(t *transfer.TableSplitterTransformer) diag.Diagnostics {
	var diags diag.Diagnostics

	if tables := t.GetTables(); tables != nil {
		if m.Tables == nil {
			m.Tables = new(transferTransformerTablesFilter)
		}
		diags.Append(m.Tables.parse(tables)...)
	} else {
		m.Tables = nil
	}

	m.Columns = convertSliceToTFStrings(t.GetColumns())
	m.Splitter = types.StringValue(t.GetSplitter())

	return diags
}
