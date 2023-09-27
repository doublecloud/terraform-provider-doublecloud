package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	ncDataTerraformID = fmt.Sprintf("data.doublecloud_network_connection.%v", ncName)
)

func TestNewNetworkConnectionDataSource(t *testing.T) {
	const (
		netID = "netID"
		ncID  = "ncID"

		vpcID = "vpcID"
		accID = "accID"
		regID = "regID"
		v4    = "v4"
		v6    = "v6"

		peeringID = "peeringID"
		managedV4 = "managedV4"
		managedV6 = "managedV6"

		name    = "name"
		peerURL = "peerURL"

		managedURL = "managedURL"
	)

	f := &fakeNetworkConnectionServiceServer{}
	endpoint, err := startNetworkConnectionServiceMock(f)
	require.NoError(t, err)

	t.Run("AWS Peering", func(t *testing.T) {
		f.getMock = func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
			require.Equal(t, ncID, req.NetworkConnectionId)
			return &network.NetworkConnection{
				Id:        ncID,
				NetworkId: netID,
				ConnectionInfo: &network.NetworkConnection_Aws{
					Aws: &network.AWSNetworkConnectionInfo{
						Type: &network.AWSNetworkConnectionInfo_Peering{
							Peering: &network.AWSNetworkConnectionPeeringInfo{
								VpcId:                vpcID,
								AccountId:            accID,
								RegionId:             regID,
								Ipv4CidrBlock:        v4,
								Ipv6CidrBlock:        v6,
								PeeringConnectionId:  peeringID,
								ManagedIpv4CidrBlock: managedV4,
								ManagedIpv6CidrBlock: managedV6,
							},
						},
					},
				},
			}, nil
		}
		defer func() {
			f.getMock = nil
		}()

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{

					Config: testNetworkConnectionDatasourceConfig(ncID),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncDataTerraformID, "id", ncID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "network_id", netID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "description", ""),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.vpc_id", vpcID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.account_id", accID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.region_id", regID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.ipv4_cidr_block", v4),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.ipv6_cidr_block", v6),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.peering_connection_id", peeringID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.managed_ipv4_cidr_block", managedV4),
						resource.TestCheckResourceAttr(ncDataTerraformID, "aws.peering.managed_ipv6_cidr_block", managedV6),
					),
				},
			},
		})
	})

	t.Run("Google Peering", func(t *testing.T) {
		f.getMock = func(ctx context.Context, req *network.GetNetworkConnectionRequest) (*network.NetworkConnection, error) {
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
			}, nil
		}
		defer func() {
			f.getMock = nil
		}()

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{

					Config: testNetworkConnectionDatasourceConfig(ncID),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncDataTerraformID, "id", ncID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "network_id", netID),
						resource.TestCheckResourceAttr(ncDataTerraformID, "description", ""),
						resource.TestCheckResourceAttr(ncDataTerraformID, "google.name", name),
						resource.TestCheckResourceAttr(ncDataTerraformID, "google.peer_network_url", peerURL),
						resource.TestCheckResourceAttr(ncDataTerraformID, "google.managed_network_url", managedURL),
					),
				},
			},
		})
	})
}

func testNetworkConnectionDatasourceConfig(id string) string {
	return fmt.Sprintf(`
data "doublecloud_network_connection" %[1]q {
  id = %[2]q
}
`,
		ncName,
		id,
	)
}
