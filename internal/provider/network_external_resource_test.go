package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	testAccExternalNetworkId string = fmt.Sprintf("doublecloud_external_network.%v", testAccNetworkName)
)

func TestExternalNetworkResource(t *testing.T) {
	t.Parallel()
	const (
		vpcID     = "vpcID"
		regionID  = "regionID"
		accountID = "accountID"
		roleARN   = "roleARN"
		networkID = "networkID"
		ipv4      = "IPv4"
		ipv6      = "IPv6"
	)

	m := ExternalNetworkResourceModel{
		ProjectID: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccNetworkName),
		AWS: &awsExternalNetworkResourceModel{
			VPCID:      types.StringValue(vpcID),
			RegionID:   types.StringValue(regionID),
			AccountID:  types.StringValue(accountID),
			IAMRoleARN: types.StringValue("roleARN"),
			//PrivateSubnets: types.Bool{},
		},
	}

	f := &fakeNetworkServiceServer{
		importMock: func(ctx context.Context, req *network.ImportNetworkRequest) (*doublecloud.Operation, error) {
			require.Equal(t, testAccNetworkName, req.Name)
			require.Equal(t, testProjectId, req.ProjectId)
			awsParams, ok := req.Params.(*network.ImportNetworkRequest_Aws)
			require.True(t, ok)
			require.Equal(t, vpcID, awsParams.Aws.VpcId)
			require.Equal(t, regionID, awsParams.Aws.RegionId)
			require.Equal(t, accountID, awsParams.Aws.AccountId)
			require.Equal(t, roleARN, awsParams.Aws.IamRoleArn)
			return &doublecloud.Operation{
				Id:         uuid.NewString(),
				ProjectId:  testProjectId,
				Status:     doublecloud.Operation_STATUS_DONE,
				ResourceId: networkID,
			}, nil
		},
		getMock: func(ctx context.Context, req *network.GetNetworkRequest) (*network.Network, error) {
			require.Equal(t, networkID, req.NetworkId)
			return &network.Network{
				Id:            networkID,
				ProjectId:     testProjectId,
				CloudType:     "aws",
				RegionId:      regionID,
				Name:          testAccNetworkName,
				Ipv4CidrBlock: ipv4,
				Ipv6CidrBlock: ipv6,
				Status:        network.Network_NETWORK_STATUS_ACTIVE,
				ExternalResources: &network.Network_Aws{
					Aws: &network.AwsExternalResources{
						VpcId:      vpcID,
						AccountId:  &wrappers.StringValue{Value: accountID},
						IamRoleArn: &wrappers.StringValue{Value: roleARN},
					},
				},
				IsExternal: true,
			}, nil
		},
		deleteMock: func(ctx context.Context, req *network.DeleteNetworkRequest) (*doublecloud.Operation, error) {
			require.Equal(t, networkID, req.NetworkId)
			return &doublecloud.Operation{
				Id:         uuid.NewString(),
				ProjectId:  testProjectId,
				Status:     doublecloud.Operation_STATUS_DONE,
				ResourceId: networkID,
			}, nil
		},
	}
	endpoint, err := startNetworkServiceMock(f)
	require.NoError(t, err)

	resource.UnitTest(t, resource.TestCase{
		IsUnitTest:               true,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testAccExternalNetworkResourceWrongConfig(&m),
				ExpectError: regexp.MustCompile(`No attribute specified when one \(and only one\) of \[aws,google] is required`),
			},
			{
				Config: testAccExternalNetworkResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "id", networkID),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "name", testAccNetworkName),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "project_id", testProjectId),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.vpc_id", vpcID),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.region_id", regionID),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.account_id", accountID),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.iam_role_arn", roleARN),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "ipv4_cidr_block", ipv4),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "ipv6_cidr_block", ipv6),
				),
			},
			{
				PreConfig: func() {
					f.getMock = func(ctx context.Context, req *network.GetNetworkRequest) (*network.Network, error) {
						return &network.Network{
							Id:            networkID,
							ProjectId:     testProjectId,
							CloudType:     "aws",
							RegionId:      regionID,
							Name:          testAccNetworkName,
							Ipv4CidrBlock: ipv4,
							Ipv6CidrBlock: ipv6,
							Status:        network.Network_NETWORK_STATUS_ACTIVE,
						}, nil
					}
				},
				Config:      testAccExternalNetworkResourceConfig(&m),
				ExpectError: regexp.MustCompile(`Failed parse Network External Resource`),
			},
			// Update not supported
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccExternalNetworkResourceConfig(m *ExternalNetworkResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_external_network" %[2]q {
  project_id = %[1]q
  name = %[2]q
  aws = {
    vpc_id = %[3]q
    region_id = %[4]q
    account_id = %[5]q
    iam_role_arn = %[6]q
  }
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
		m.AWS.VPCID.ValueString(),
		m.AWS.RegionID.ValueString(),
		m.AWS.AccountID.ValueString(),
		m.AWS.IAMRoleARN.ValueString(),
	)
}

func testAccExternalNetworkResourceWrongConfig(m *ExternalNetworkResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_external_network" %[2]q {
  project_id = %[1]q
  name = %[2]q
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
	)
}
