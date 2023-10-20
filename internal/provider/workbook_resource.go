package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/doublecloud/go-genproto/doublecloud/visualization/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgenvis "github.com/doublecloud/go-sdk/gen/visualization"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ resource.Resource = &WorkbookResource{}
var _ resource.ResourceWithImportState = &WorkbookResource{}

func NewWorkbookResource() resource.Resource {
	return &WorkbookResource{}
}

type WorkbookResource struct {
	sdk *dcsdk.SDK
	svc *dcgenvis.WorkbookServiceClient
}

type WorkbookResourceModel struct {
	Id          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Title       types.String `tfsdk:"title"`
	Config      types.String `tfsdk:"config"`
	Connections types.Set    `tfsdk:"connect"`
}

func (r *WorkbookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbook"
}

func (r *WorkbookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Workbook resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Workbook identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project identifier",
			},
			"title": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Title of resource",
			},
			"config": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "Workbook configuration (json encoded)",
			},
		},
		Blocks: map[string]schema.Block{
			"connect": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Connection name",
						},
						"config": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "[Configuration of connection (json encoded)](https://double.cloud/docs/en/public-api/api-reference/visualization/configs/Connection)",
						},
						"secret": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Secret",
						},
					},
				},
			},
		},
	}
}

func (r *WorkbookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.svc = r.sdk.Visualization().Workbook()
}

func createWorkbookRequest(m *WorkbookResourceModel) (*visualization.CreateWorkbookRequest, diag.Diagnostics) {
	rq := &visualization.CreateWorkbookRequest{}
	rq.ProjectId = m.ProjectID.ValueString()
	rq.WorkbookTitle = m.Title.ValueString()
	return rq, nil
}

func deleteWorkbookRequest(m *WorkbookResourceModel) (*visualization.DeleteWorkbookRequest, diag.Diagnostics) {
	return &visualization.DeleteWorkbookRequest{WorkbookId: m.Id.ValueString()}, nil
}

func createWorkbookConnectionRequests(m *WorkbookResourceModel) ([]*visualization.CreateWorkbookConnectionRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if m.Connections.IsNull() {
		return nil, nil
	}
	requests := make([]*visualization.CreateWorkbookConnectionRequest, 0)
	for _, el := range m.Connections.Elements() {
		cnn := el.(types.Object).Attributes()

		t := &visualization.Connection{Config: &structpb.Value{}}
		err := t.Config.UnmarshalJSON([]byte(cnn["config"].(types.String).ValueString()))
		if err != nil {
			diags.AddError("failed to parse connection", err.Error())
		}

		ss := cnn["secret"].(types.String).ValueString()
		s := &visualization.Secret{Secret: &visualization.Secret_PlainSecret{PlainSecret: &visualization.PlainSecret{Secret: ss}}}
		requests = append(requests, &visualization.CreateWorkbookConnectionRequest{
			WorkbookId:     m.Id.ValueString(),
			ConnectionName: cnn["name"].(types.String).ValueString(),
			Connection:     t,
			Secret:         s,
		})
	}
	return requests, diags
}

func modifyWorkbookRequest(m *WorkbookResourceModel) (*visualization.UpdateWorkbookRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	workbook := &visualization.Workbook{Config: &structpb.Value{}}
	err := workbook.Config.UnmarshalJSON([]byte(m.Config.ValueString()))
	if err != nil {
		diags.AddError("failed to parse config", err.Error())
	}

	return &visualization.UpdateWorkbookRequest{
		WorkbookId: m.Id.ValueString(),
		Workbook:   workbook,
	}, diags
}

func (r *WorkbookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *WorkbookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: move to validation
	if !json.Valid([]byte(data.Config.ValueString())) {
		resp.Diagnostics.AddError("incorrect config format", data.Config.ValueString())
		return
	}

	rq, diag := createWorkbookRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.svc.Create(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}

	data.Id = types.StringValue(op.ResourceId())

	getRequest, diag := getWorkbookResourceRequest(data)
	if diag.HasError() {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	getResponse, err := r.svc.Get(ctx, getRequest)
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}
	// Temporary hack to align json formats between Lens and Terraform
	m := protojson.MarshalOptions{
		Indent:          "",
		EmitUnpopulated: true,
	}
	generatedConfig, err := m.Marshal(getResponse.Workbook.Config)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse config from server", err.Error())
		return
	}

	connections, diag := createWorkbookConnectionRequests(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	for _, c := range connections {
		rs, err := r.svc.CreateConnection(ctx, c)
		if err != nil {
			resp.Diagnostics.AddError("failed to create", err.Error())
			return
		}
		op, err := r.sdk.WrapOperation(rs, err)
		if err != nil {
			resp.Diagnostics.AddError("failed to create", err.Error())
			return
		}
		err = op.Wait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to create", err.Error())
			return
		}
	}

	if data.Config.ValueString() != "" {
		// Update workbook config
		modifyReq, diag := modifyWorkbookRequest(data)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		rs, err := r.svc.Update(ctx, modifyReq)
		if err != nil {
			resp.Diagnostics.AddError("failed to modify", err.Error())
			return
		}
		op, err := r.sdk.WrapOperation(rs, err)
		if err != nil {
			resp.Diagnostics.AddError("failed to modify", err.Error())
			return
		}
		err = op.Wait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to modify", err.Error())
			return
		}
	} else {
		// Use computed value as default
		data.Config = types.StringValue(string(generatedConfig))
	}

	tflog.Trace(ctx, fmt.Sprintf("doublecloud_workbook has been created: %s", op.ResourceId()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:unused
func compareJSONs(a string, b any) (bool, error) {
	var astruct interface{}
	err := json.Unmarshal([]byte(a), astruct)
	if err != nil {
		return false, errors.Join(errors.New("invalid json doc a1"), err)
	}
	aformatted, err := json.MarshalIndent(&astruct, "", "")
	if err != nil {
		return false, errors.New("invalid json document a2")
	}
	return reflect.DeepEqual(aformatted, b), nil
}

func getWorkbookResourceRequest(m *WorkbookResourceModel) (*visualization.GetWorkbookRequest, diag.Diagnostics) {
	if m.Id == types.StringNull() {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("unknown id", "missed one of required fields: id or name")}
	}
	return &visualization.GetWorkbookRequest{WorkbookId: m.Id.ValueString()}, nil
}

func (r *WorkbookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *WorkbookResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: support json comparison with null values

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func updateWorkbookRequest(m *WorkbookResourceModel) (*visualization.UpdateWorkbookRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	rq := &visualization.UpdateWorkbookRequest{WorkbookId: m.Id.ValueString(), Workbook: &visualization.Workbook{Config: &structpb.Value{}}}
	err := rq.Workbook.Config.UnmarshalJSON([]byte(m.Config.ValueString()))
	if err != nil {
		diags.AddError("failed to parse config", err.Error())
	}
	return rq, diags
}

func (r *WorkbookResource) updateConnections(ctx context.Context, resp *resource.UpdateResponse, data *WorkbookResourceModel) {
	connections, diag := createWorkbookConnectionRequests(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	toCreate := make([]*visualization.CreateWorkbookConnectionRequest, 0)
	toUpdate := make([]*visualization.UpdateWorkbookConnectionRequest, 0)
	for _, c := range connections {
		rs, _ := r.svc.GetConnection(ctx, &visualization.GetWorkbookConnectionRequest{WorkbookId: data.Id.ValueString(), ConnectionName: c.ConnectionName})
		if rs == nil {
			toCreate = append(toCreate, c)
			continue
		}
		if rs.Connection != c.Connection {
			toUpdate = append(toUpdate, &visualization.UpdateWorkbookConnectionRequest{
				WorkbookId:     c.WorkbookId,
				ConnectionName: c.ConnectionName,
				Connection:     c.Connection,
				Secret:         c.Secret,
			})
		}
	}

	for _, c := range toCreate {
		rs, err := r.svc.CreateConnection(ctx, c)
		if err != nil {
			resp.Diagnostics.AddError("failed to create", err.Error())
			return
		}
		op, err := r.sdk.WrapOperation(rs, err)
		if err != nil {
			resp.Diagnostics.AddError("failed to create", err.Error())
			return
		}
		err = op.Wait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to create", err.Error())
			return
		}
	}

	for _, c := range toUpdate {
		rs, err := r.svc.UpdateConnection(ctx, c)
		if err != nil {
			resp.Diagnostics.AddError("failed to update", err.Error())
			return
		}
		op, err := r.sdk.WrapOperation(rs, err)
		if err != nil {
			resp.Diagnostics.AddError("failed to update", err.Error())
			return
		}
		err = op.Wait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to update", err.Error())
			return
		}
	}

}

func (r *WorkbookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *WorkbookResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	r.updateConnections(ctx, resp, data)

	rq, diag := updateWorkbookRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	rs, err := r.svc.Update(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}

	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update", err.Error())
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WorkbookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *WorkbookResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rq, diag := deleteWorkbookRequest(data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	rs, err := r.svc.Delete(ctx, rq)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(rs, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("deleted workbook: %v", data.Id))
}

func (r *WorkbookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
