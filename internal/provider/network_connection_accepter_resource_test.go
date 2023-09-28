package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestNetworkConnectionAccepterResource(t *testing.T) {
	const (
		ncID = "ncID"
	)

	var mocks []getNetworkConnectionMockFunc
	f := &fakeNetworkConnectionServiceServer{
		getMock: func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
			// mock several different responses

			var m getNetworkConnectionMockFunc
			switch len(mocks) {
			case 0:
				return nil, fmt.Errorf("there is no more mocked responses")
			case 1:
				m, mocks = mocks[0], nil
				return m(ctx, req)
			default:
				m, mocks = mocks[0], mocks[1:]
				return m(ctx, req)
			}
		},
	}
	endpoint, err := startNetworkConnectionServiceMock(f)
	require.NoError(t, err)

	t.Run("accept peering connection", func(t *testing.T) {
		mocks = []getNetworkConnectionMockFunc{
			func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
				require.Equal(t, ncID, req.NetworkConnectionId)
				return &network.NetworkConnection{
					Id:     ncID,
					Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_CREATING,
					ConnectionInfo: &network.NetworkConnection_Aws{
						Aws: &network.AWSNetworkConnectionInfo{
							Type: &network.AWSNetworkConnectionInfo_Peering{
								Peering: &network.AWSNetworkConnectionPeeringInfo{},
							},
						},
					},
				}, nil
			},
			func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
				require.Equal(t, ncID, req.NetworkConnectionId)
				return &network.NetworkConnection{
					Id:     ncID,
					Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_ACTIVE,
					ConnectionInfo: &network.NetworkConnection_Aws{
						Aws: &network.AWSNetworkConnectionInfo{
							Type: &network.AWSNetworkConnectionInfo_Peering{
								Peering: &network.AWSNetworkConnectionPeeringInfo{},
							},
						},
					},
				}, nil
			},
		}
		expectedCalls := len(mocks)

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					Config: networkConnectionAccepterResourceConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("doublecloud_network_connection_accepter.accept", "id", ncID),
					),
				},
			},
		})
		require.Nil(t, mocks, "expected %d calls, but %d mocks left", expectedCalls, len(mocks))
	})

	t.Run("failed to accept peering connection", func(t *testing.T) {
		mocks = []getNetworkConnectionMockFunc{
			func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
				require.Equal(t, ncID, req.NetworkConnectionId)
				return &network.NetworkConnection{
					Id:     ncID,
					Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_CREATING,
					ConnectionInfo: &network.NetworkConnection_Aws{
						Aws: &network.AWSNetworkConnectionInfo{
							Type: &network.AWSNetworkConnectionInfo_Peering{
								Peering: &network.AWSNetworkConnectionPeeringInfo{},
							},
						},
					},
				}, nil
			},
			func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
				require.Equal(t, ncID, req.NetworkConnectionId)
				return &network.NetworkConnection{
					Id:           ncID,
					Status:       network.NetworkConnection_NETWORK_CONNECTION_STATUS_ERROR,
					StatusReason: "THE reason",
					ConnectionInfo: &network.NetworkConnection_Aws{
						Aws: &network.AWSNetworkConnectionInfo{
							Type: &network.AWSNetworkConnectionInfo_Peering{
								Peering: &network.AWSNetworkConnectionPeeringInfo{},
							},
						},
					},
				}, nil
			},
		}
		expectedCalls := len(mocks)

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					Config:      networkConnectionAccepterResourceConfig,
					ExpectError: regexp.MustCompile("(?s)can not accept network connection.*error: network connection \"ncID\" is in ERROR state, reason: THE reason"),
				},
			},
		})
		require.Nil(t, mocks, "expected %d calls, but %d mocks left", expectedCalls, len(mocks))
	})

	t.Run("import accepter", func(t *testing.T) {
		require.Nil(t, mocks)
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					ImportState:   true,
					ResourceName:  "doublecloud_network_connection_accepter.accept",
					ImportStateId: ncID,
					Config:        networkConnectionAccepterResourceConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncTerraformID, "id", ncID),
					),
				},
			},
		})
	})

	t.Run("create Google peering and accept it", func(t *testing.T) {
		const (
			netID = "netID"
			ncID  = "ncID"

			name    = "name"
			peerURL = "peerURL"

			managedURL = "managedURL"
		)
		mocks = []getNetworkConnectionMockFunc{
			func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
				require.Equal(t, ncID, req.NetworkConnectionId)
				return &network.NetworkConnection{
					Id:        ncID,
					NetworkId: netID,
					ConnectionInfo: &network.NetworkConnection_Google{
						Google: &network.GoogleNetworkConnectionInfo{
							Name:              name,
							PeerNetworkUrl:    peerURL,
							ManagedNetworkUrl: managedURL,
						},
					},
					Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_CREATING,
				}, nil
			},
			func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
				require.Equal(t, ncID, req.NetworkConnectionId)
				return &network.NetworkConnection{
					Id:        ncID,
					NetworkId: netID,
					ConnectionInfo: &network.NetworkConnection_Google{
						Google: &network.GoogleNetworkConnectionInfo{
							Name:              name,
							PeerNetworkUrl:    peerURL,
							ManagedNetworkUrl: managedURL,
						},
					},
					Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_ACTIVE,
				}, nil
			},
		}
		expectedCalls := len(mocks)

		f.createMock = func(ctx context.Context, req *network.CreateNetworkConnectionRequest) (*doublecloud.Operation, error) {
			require.Equal(t, netID, req.NetworkId)
			peering := req.GetGoogle()
			require.NotNil(t, peering)
			require.Equal(t, name, peering.Name)
			require.Equal(t, peerURL, peering.PeerNetworkUrl)

			return networkOperationDone(ncID), nil
		}
		f.deleteMock = func(ctx context.Context, req *network.DeleteNetworkConnectionRequest) (*doublecloud.Operation, error) {
			require.Equal(t, ncID, req.NetworkConnectionId)
			return networkOperationDone(ncID), nil
		}
		defer func() {
			f.createMock = nil
			f.deleteMock = nil
		}()

		m := &NetworkConnectionModel{
			NetworkID: types.StringValue(netID),
			Google: &googleNetworkConnectionInfo{
				Name:           types.StringValue(name),
				PeerNetworkURL: types.StringValue(peerURL),
			},
		}

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					Config: testNetworkConnectionGooglePeeringResourceConfig(m) + "\n" + networkConnectionAndAccepterResourceConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncTerraformID, "id", ncID),
						resource.TestCheckResourceAttr("doublecloud_network_connection_accepter.accept", "id", ncID),
					),
				},
			},
		})
		require.Nil(t, mocks, "expected %d calls, but %d mocks left", expectedCalls, len(mocks))
	})
}

const (
	networkConnectionAccepterResourceConfig = `
resource "doublecloud_network_connection_accepter" "accept" {
  id = "ncID"
}
`
	networkConnectionAndAccepterResourceConfig = `
resource "doublecloud_network_connection_accepter" "accept" {
  id = doublecloud_network_connection.nc.id
}
`
)
