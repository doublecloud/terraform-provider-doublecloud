package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgen "github.com/doublecloud/go-sdk/gen/transfer"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TransferDataSource{}

func NewTransferDataSource() datasource.DataSource {
	return &TransferDataSource{}
}

type TransferDataSource struct {
	sdk *dcsdk.SDK
	svc *dcgen.TransferServiceClient
}

type TransferDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	Type        types.String `tfsdk:"type"`
}

func (d *TransferDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transfer"
}

func (d *TransferDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Transfer data source",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project ID",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Transfer ID",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Transfer name",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Transfer description",
			},
			"status": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Transfer status",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Transfer type",
			},
		},
	}
}

func (d *TransferDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.sdk = sdk
	d.svc = d.sdk.Transfer().Transfer()
}

func (d *TransferDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TransferDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Id == types.StringNull() && data.Name == types.StringNull() {
		resp.Diagnostics.AddError("Missing attribute", "Specify either `id` or `name`")
		return
	}

	if data.Id == types.StringNull() {
		it := d.svc.TransferIterator(ctx, &transfer.ListTransfersRequest{ProjectId: data.ProjectID.ValueString()})
		for it.Next() {
			c := it.Value()
			if c.Name == data.Name.ValueString() {
				data.Id = types.StringValue(c.Id)
				break
			}
		}
		if data.Id == types.StringNull() {
			resp.Diagnostics.AddError("Transfer not found", fmt.Sprintf("Transfer `%v` hasn't been found", data.Name.ValueString()))
			return
		}
	}

	response, err := d.svc.Get(ctx, &transfer.GetTransferRequest{TransferId: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to get", err.Error())
		return
	}

	data.Description = types.StringValue(response.Description)
	data.Status = types.StringValue(response.Status.String())
	data.Type = types.StringValue(response.Type.String())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
