package provider

import (
	"context"
	"fmt"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	dcsdk "github.com/doublecloud/go-sdk"
	dcgennet "github.com/doublecloud/go-sdk/gen/network"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewNetworkConnectionResource() resource.Resource {
	return &NetworkConnectionResource{}
}

type NetworkConnectionResource struct {
	sdk                      *dcsdk.SDK
	networkConnectionService *dcgennet.NetworkConnectionServiceClient
}

func (r *NetworkConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_connection"
}

func (r *NetworkConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = networkConnectionResourceSchema
}

func (r *NetworkConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.networkConnectionService = r.sdk.Network().NetworkConnection()
}

func (r *NetworkConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NetworkConnectionModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &network.CreateNetworkConnectionRequest{
		NetworkId:   data.NetworkID.ValueString(),
		Params:      nil,
		Description: data.Description.ValueString(),
	}

	switch {
	case data.AWS != nil:
		aws := &network.CreateAWSNetworkConnectionRequest{}
		createReq.Params = &network.CreateNetworkConnectionRequest_Aws{
			Aws: aws,
		}
		switch {
		case data.AWS.Peering != nil:
			aws.Type = &network.CreateAWSNetworkConnectionRequest_Peering{
				Peering: &network.CreateAWSNetworkConnectionPeeringRequest{
					VpcId:         data.AWS.Peering.VPCID.ValueString(),
					AccountId:     data.AWS.Peering.AccountID.ValueString(),
					RegionId:      data.AWS.Peering.RegionID.ValueString(),
					Ipv4CidrBlock: data.AWS.Peering.IPv4CIDRBlock.ValueString(),
					Ipv6CidrBlock: data.AWS.Peering.IPv6CIDRBlock.ValueString(),
				},
			}
		default:
			resp.Diagnostics.AddError("misconfiguration", "\"aws.peering\" must be specified")
			return
		}
	case data.Google != nil:
		createReq.Params = &network.CreateNetworkConnectionRequest_Google{
			Google: &network.CreateGoogleNetworkConnectionRequest{
				Name:           data.Google.Name.ValueString(),
				PeerNetworkUrl: data.Google.PeerNetworkURL.ValueString(),
			},
		}
	default:
		resp.Diagnostics.AddError("misconfiguration", "at least one of \"aws\" or \"google\" must be specified")
		return
	}

	opObj, err := r.networkConnectionService.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
		return
	}

	op, err := r.sdk.WrapOperation(opObj, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to create", err.Error())
	}

	data.ID = types.StringValue(op.ResourceId())

	if !getNetworkConnection(ctx, r.networkConnectionService, op.ResourceId(), data, resp.Diagnostics) {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *NetworkConnectionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !getNetworkConnection(ctx, r.networkConnectionService, data.ID.ValueString(), data, resp.Diagnostics) {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Failed to update network connection", "network connections don't support updates")
}

func (r *NetworkConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NetworkConnectionModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	net, err := r.networkConnectionService.Delete(ctx, &network.DeleteNetworkConnectionRequest{NetworkConnectionId: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
		return
	}
	op, err := r.sdk.WrapOperation(net, err)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete", err.Error())
	}
}

func (r *NetworkConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
