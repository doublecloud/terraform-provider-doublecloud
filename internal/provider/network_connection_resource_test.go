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

const (
	ncName = "nc"
)

var (
	ncTerraformID = fmt.Sprintf("doublecloud_network_connection.%v", ncName)
)

func TestNetworkConnectionResource(t *testing.T) {
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

	f := &fakeNetworkConnectionServiceServer{
		deleteMock: func(ctx context.Context, req *network.DeleteNetworkConnectionRequest) (*doublecloud.Operation, error) {
			require.Equal(t, ncID, req.NetworkConnectionId)
			return networkOperationDone(ncID), nil
		},
	}
	endpoint, err := startNetworkConnectionServiceMock(f)
	require.NoError(t, err)

	for _, tc := range []testCaseErrorConfig{
		{
			name:   "missedProviderNCConfig",
			config: missedProviderNCConfig,
			err:    regexp.MustCompile(`No attribute specified when one \(and only one\) of \[aws,google] is required`),
		},
		{
			name:   "severalProvidersNCConfig",
			config: severalProvidersNCConfig,
			err:    regexp.MustCompile(`2 attributes specified when one \(and only one\) of \[aws,google] is required`),
		},
		{
			name:   "missedTypeAWSNCConfig",
			config: missedTypeAWSNCConfig,
			err:    regexp.MustCompile(`No attribute specified when one \(and only one\) of \[aws.peering] is required`),
		},
		{
			name:   "missedIPAWSPeeringNCConfig",
			config: missedIPAWSPeeringNCConfig,
			err:    regexp.MustCompile(`"ipv4_cidr_block" is required`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, unitTestCase(endpoint, tc))
		})
	}

	t.Run("create AWS Peering", func(t *testing.T) {
		f.createMock = func(ctx context.Context, req *network.CreateNetworkConnectionRequest) (*doublecloud.Operation, error) {
			require.Equal(t, netID, req.NetworkId)
			aws := req.GetAws()
			require.NotNil(t, aws)
			peering := aws.GetPeering()
			require.NotNil(t, peering)
			require.Equal(t, vpcID, peering.VpcId)
			require.Equal(t, accID, peering.AccountId)
			require.Equal(t, regID, peering.RegionId)
			require.Equal(t, v4, peering.Ipv4CidrBlock)
			require.Equal(t, v6, peering.Ipv6CidrBlock)

			return networkOperationDone(ncID), nil
		}
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
				Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_CREATING,
			}, nil
		}
		defer func() {
			f.createMock = nil
			f.getMock = nil
		}()

		m := &NetworkConnectionModel{
			NetworkID: types.StringValue(netID),
			AWS: &awsNetworkConnectionInfo{
				Peering: &awsNetworkConnectionPeeringInfo{
					VPCID:         types.StringValue(vpcID),
					AccountID:     types.StringValue(accID),
					RegionID:      types.StringValue(regID),
					IPv4CIDRBlock: types.StringValue(v4),
					IPv6CIDRBlock: types.StringValue(v6),
				},
			},
		}

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					Config: testNetworkConnectionAWSPeeringResourceConfig(m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncTerraformID, "id", ncID),
						resource.TestCheckResourceAttr(ncTerraformID, "network_id", netID),
						resource.TestCheckResourceAttr(ncTerraformID, "description", ""),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.vpc_id", vpcID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.account_id", accID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.region_id", regID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.ipv4_cidr_block", v4),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.ipv6_cidr_block", v6),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.peering_connection_id", peeringID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.managed_ipv4_cidr_block", managedV4),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.managed_ipv6_cidr_block", managedV6),
					),
				},
			},
		})
	})

	t.Run("create Google Peering", func(t *testing.T) {
		f.createMock = func(ctx context.Context, req *network.CreateNetworkConnectionRequest) (*doublecloud.Operation, error) {
			require.Equal(t, netID, req.NetworkId)
			peering := req.GetGoogle()
			require.NotNil(t, peering)
			require.Equal(t, name, peering.Name)
			require.Equal(t, peerURL, peering.PeerNetworkUrl)

			return networkOperationDone(ncID), nil
		}
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
				Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_CREATING,
			}, nil
		}
		defer func() {
			f.createMock = nil
			f.getMock = nil
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
					Config: testNetworkConnectionGooglePeeringResourceConfig(m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncTerraformID, "id", ncID),
						resource.TestCheckResourceAttr(ncTerraformID, "network_id", netID),
						resource.TestCheckResourceAttr(ncTerraformID, "description", ""),
						resource.TestCheckResourceAttr(ncTerraformID, "google.name", name),
						resource.TestCheckResourceAttr(ncTerraformID, "google.peer_network_url", peerURL),
						resource.TestCheckResourceAttr(ncTerraformID, "google.managed_network_url", managedURL),
					),
				},
			},
		})
	})

	t.Run("import AWS Peering", func(t *testing.T) {
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
				Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_CREATING,
			}, nil
		}
		defer func() {
			f.getMock = nil
		}()

		m := &NetworkConnectionModel{
			NetworkID: types.StringValue(netID),
			AWS: &awsNetworkConnectionInfo{
				Peering: &awsNetworkConnectionPeeringInfo{
					VPCID:         types.StringValue(vpcID),
					AccountID:     types.StringValue(accID),
					RegionID:      types.StringValue(regID),
					IPv4CIDRBlock: types.StringValue(v4),
					IPv6CIDRBlock: types.StringValue(v6),
				},
			},
		}

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					ImportState:   true,
					ResourceName:  ncTerraformID,
					ImportStateId: ncID,
					Config:        testNetworkConnectionAWSPeeringResourceConfig(m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncTerraformID, "id", ncID),
						resource.TestCheckResourceAttr(ncTerraformID, "network_id", netID),
						resource.TestCheckResourceAttr(ncTerraformID, "description", ""),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.vpc_id", vpcID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.account_id", accID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.region_id", regID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.ipv4_cidr_block", v4),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.ipv6_cidr_block", v6),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.peering_connection_id", peeringID),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.managed_ipv4_cidr_block", managedV4),
						resource.TestCheckResourceAttr(ncTerraformID, "aws.peering.managed_ipv6_cidr_block", managedV6),
					),
				},
			},
		})
	})

	t.Run("import Google Peering", func(t *testing.T) {
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
				Status: network.NetworkConnection_NETWORK_CONNECTION_STATUS_CREATING,
			}, nil
		}
		defer func() {
			f.getMock = nil
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
					ImportState:   true,
					ResourceName:  ncTerraformID,
					ImportStateId: ncID,
					Config:        testNetworkConnectionGooglePeeringResourceConfig(m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(ncTerraformID, "id", ncID),
						resource.TestCheckResourceAttr(ncTerraformID, "network_id", netID),
						resource.TestCheckResourceAttr(ncTerraformID, "description", ""),
						resource.TestCheckResourceAttr(ncTerraformID, "google.name", name),
						resource.TestCheckResourceAttr(ncTerraformID, "google.peer_network_url", peerURL),
						resource.TestCheckResourceAttr(ncTerraformID, "google.managed_network_url", managedURL),
					),
				},
			},
		})
	})
}

const (
	missedProviderNCConfig = `
resource "doublecloud_network_connection" "missedProviderNCConfig" {
  network_id = "n"
}
`
	severalProvidersNCConfig = `
resource "doublecloud_network_connection" "missedProviderNCConfig" {
  network_id = "n"
  aws = {
    peering = {
      vpc_id = "v"
      account_id = "a"
      region_id = "r"
      ipv4_cidr_block = "i"
    }
  }
  google = {
    name = "n"
    peer_network_url = "u"
  }
}
`
	missedTypeAWSNCConfig = `
resource "doublecloud_network_connection" "missedProviderNCConfig" {
  network_id = "n"
  aws = {}
}
`
	missedIPAWSPeeringNCConfig = `
resource "doublecloud_network_connection" "missedProviderNCConfig" {
  network_id = "n"
  aws = {
    peering = {
      vpc_id = "v"
      account_id = "a"
      region_id = "r"
    }
  }
}
`
)

func testNetworkConnectionAWSPeeringResourceConfig(m *NetworkConnectionModel) string {
	return fmt.Sprintf(`
resource "doublecloud_network_connection" %[1]q {
  network_id = %[2]q
  aws = {
    peering = {
      vpc_id = %[3]q
      account_id = %[4]q
      region_id = %[5]q
      ipv4_cidr_block = %[6]q
      ipv6_cidr_block = %[7]q
    }
  }
}

output "test_attr" {
   value = doublecloud_network_connection.%[1]s.aws.peering.peering_connection_id
}
`,
		ncName,
		m.NetworkID.ValueString(),
		m.AWS.Peering.VPCID.ValueString(),
		m.AWS.Peering.AccountID.ValueString(),
		m.AWS.Peering.RegionID.ValueString(),
		m.AWS.Peering.IPv4CIDRBlock.ValueString(),
		m.AWS.Peering.IPv6CIDRBlock.ValueString(),
	)
}

func testNetworkConnectionGooglePeeringResourceConfig(m *NetworkConnectionModel) string {
	return fmt.Sprintf(`
resource "doublecloud_network_connection" %[1]q {
  network_id = %[2]q
  google = {
    name = %[3]q
    peer_network_url = %[4]q
  }
}

output "test_attr" {
   value = doublecloud_network_connection.%[1]s.google.managed_network_url
}
`,
		ncName,
		m.NetworkID.ValueString(),
		m.Google.Name.ValueString(),
		m.Google.PeerNetworkURL.ValueString(),
	)
}
