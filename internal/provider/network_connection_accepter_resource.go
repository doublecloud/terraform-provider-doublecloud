package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewNetworkConnectionAccepterResource() resource.Resource {
	return &NetworkConnectionAccepterResource{}
}

// NetworkConnectionAccepterResource is meta resource to provide async work
// with NetworkConnectionResource
type NetworkConnectionAccepterResource struct {
	NetworkConnectionResource
}

type NetworkConnectionAccepterModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *NetworkConnectionAccepterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_connection_accepter"
}

func (r *NetworkConnectionAccepterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network Connection Accepter resource",

		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *NetworkConnectionAccepterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NetworkConnectionAccepterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ncData := &NetworkConnectionModel{}
	if !getNetworkConnection(ctx, r.networkConnectionService, data.ID.ValueString(), ncData, resp.Diagnostics) {
		return
	}

	for !ncData.IsReady() {
		ncData.Poll(ctx, r.networkConnectionService, resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		timer := time.NewTimer(time.Second)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			resp.Diagnostics.AddError("poll network connection context done", fmt.Sprintf("err: %s", ctx.Err()))
			return
		}
	}

	if ok, reason := ncData.IsOK(); !ok {
		resp.Diagnostics.AddError("can not accept network connection", fmt.Sprintf("error: %s", reason))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkConnectionAccepterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *NetworkConnectionAccepterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkConnectionAccepterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// nothing to update
}

func (r *NetworkConnectionAccepterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// nothing to delete
}
