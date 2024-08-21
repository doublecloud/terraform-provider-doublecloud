package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/doublecloud/go-genproto/doublecloud/v1"
)

var (
	testAccNetworkName string = fmt.Sprintf("%v-network", testPrefix)
	testAccNetworkId   string = fmt.Sprintf("doublecloud_network.%v", testAccNetworkName)
)

func TestAccNetworkResource(t *testing.T) {
	t.Parallel()
	m := NetworkResourceModel{
		ProjectID:     types.StringValue(testProjectId),
		Name:          types.StringValue(testAccNetworkName),
		RegionID:      types.StringValue("eu-central-1"),
		Ipv4CidrBlock: types.StringValue("10.0.0.0/16"),
		CloudType:     types.StringValue("aws"),
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNetworkResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccNetworkId, "region_id", m.RegionID.ValueString()),
					resource.TestCheckResourceAttr(testAccNetworkId, "ipv4_cidr_block", m.Ipv4CidrBlock.ValueString()),
					resource.TestCheckResourceAttr(testAccNetworkId, "cloud_type", m.CloudType.ValueString()),
					resource.TestCheckResourceAttrSet(testAccNetworkId, "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(testAccNetworkId, "is_external", "false"),
				),
			},
			// Update not supported
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestBYOCNetworkResource(t *testing.T) {
	t.Parallel()
	const (
		vpcID                 = "vpcID"
		regionID              = "regionID"
		accountID             = "accountID"
		roleARN               = "roleARN"
		permissionBoundaryARN = "permissionBoundaryARN"
		networkID             = "networkID"
		ipv4                  = "IPv4"
		ipv6                  = "IPv6"

		netName     = "name"
		subnetName  = "snname"
		projectName = "project"
		saName      = "sa"
	)

	m := NetworkResourceModel{
		ProjectID: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccNetworkName),
		RegionID:  types.StringValue(regionID),
		CloudType: types.StringValue("aws"),
		AWS: &awsExternalNetworkResourceModel{
			VPCID:          types.StringValue(vpcID),
			AccountID:      types.StringValue(accountID),
			IAMRoleARN:     types.StringValue(roleARN),
			PrivateSubnets: types.BoolValue(true),
		},
		GCP: &googleExternalNetworkResourceModel{
			NetworkName:    types.StringValue(netName),
			SubnetworkName: types.StringValue(subnetName),
			ProjectName:    types.StringValue(projectName),
			SAEmail:        types.StringValue(saName),
		},
	}

	awsImportMock := func(ctx context.Context, req *network.ImportNetworkRequest) (*doublecloud.Operation, error) {
		require.Equal(t, testAccNetworkName, req.Name)
		require.Equal(t, testProjectId, req.ProjectId)
		awsParams, ok := req.Params.(*network.ImportNetworkRequest_Aws)
		require.True(t, ok)
		require.Equal(t, vpcID, awsParams.Aws.VpcId)
		require.Equal(t, regionID, awsParams.Aws.RegionId)
		require.Equal(t, accountID, awsParams.Aws.AccountId)
		require.Equal(t, roleARN, awsParams.Aws.IamRoleArn)
		require.True(t, awsParams.Aws.PrivateSubnets)
		return &doublecloud.Operation{
			Id:         uuid.NewString(),
			ProjectId:  testProjectId,
			Status:     doublecloud.Operation_STATUS_DONE,
			ResourceId: networkID,
		}, nil
	}
	gcpImportMock := func(ctx context.Context, req *network.ImportNetworkRequest) (*doublecloud.Operation, error) {
		require.Equal(t, testAccNetworkName, req.Name)
		require.Equal(t, testProjectId, req.ProjectId)
		gcpParams, ok := req.Params.(*network.ImportNetworkRequest_Google)
		require.True(t, ok)
		require.Equal(t, netName, gcpParams.Google.NetworkName)
		require.Equal(t, subnetName, gcpParams.Google.SubnetworkName)
		require.Equal(t, projectName, gcpParams.Google.ProjectName)
		require.Equal(t, saName, gcpParams.Google.ServiceAccountEmail)
		return &doublecloud.Operation{
			Id:         uuid.NewString(),
			ProjectId:  testProjectId,
			Status:     doublecloud.Operation_STATUS_DONE,
			ResourceId: networkID,
		}, nil
	}
	awsGetMock := func(ctx context.Context, req *network.GetNetworkRequest) (*network.Network, error) {
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
					VpcId:          vpcID,
					AccountId:      &wrappers.StringValue{Value: accountID},
					IamRoleArn:     &wrappers.StringValue{Value: roleARN},
					PrivateSubnets: &wrappers.BoolValue{Value: true},
				},
			},
			IsExternal: true,
		}, nil
	}
	gcpGetMock := func(ctx context.Context, req *network.GetNetworkRequest) (*network.Network, error) {
		require.Equal(t, networkID, req.NetworkId)
		return &network.Network{
			Id:            networkID,
			ProjectId:     testProjectId,
			CloudType:     "gcp",
			RegionId:      regionID,
			Name:          testAccNetworkName,
			Ipv4CidrBlock: ipv4,
			Ipv6CidrBlock: ipv6,
			Status:        network.Network_NETWORK_STATUS_ACTIVE,
			ExternalResources: &network.Network_Gcp{
				Gcp: &network.GcpExternalResources{
					NetworkName:         &wrappers.StringValue{Value: netName},
					SubnetworkName:      &wrappers.StringValue{Value: subnetName},
					ProjectName:         &wrappers.StringValue{Value: projectName},
					ServiceAccountEmail: &wrappers.StringValue{Value: saName},
				},
			},
			IsExternal: true,
		}, nil
	}
	f := &fakeNetworkServiceServer{
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

	for _, tc := range []testCaseErrorConfig{
		{
			name:   "bothExternalAndCidrNetworkConfig",
			config: bothExternalAndCidrNetworkConfig,
			err:    regexp.MustCompile(`Attribute "aws" cannot be specified when "ipv4_cidr_block" is specified`),
		},
		{
			name:   "bothExternalProvidersNetworkConfig",
			config: bothExternalProvidersNetworkConfig,
			err:    regexp.MustCompile(`Attribute "gcp" cannot be specified when "aws" is specified`),
		},
		{
			name:   "cloudTypeMismatchNetworkConfig",
			config: cloudTypeMismatchNetworkConfig,
			err:    regexp.MustCompile(`Provided BYOC AWS configuration, but "cloud_type" is set to "gcp".`),
		},
		{
			name:   "ipv4MissedNetworkConfig",
			config: ipv4MissedNetworkConfig,
			err:    regexp.MustCompile(`IPv4 CIDR block is required for non-BYOC networks.`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, unitTestCase(endpoint, tc))
		})
	}

	t.Run("create AWS network", func(t *testing.T) {
		f.importMock = awsImportMock
		f.getMock = awsGetMock
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					Config: testAWSNetworkResourceConfig(&m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(testAccNetworkId, "id", networkID),
						resource.TestCheckResourceAttr(testAccNetworkId, "name", testAccNetworkName),
						resource.TestCheckResourceAttr(testAccNetworkId, "project_id", testProjectId),
						resource.TestCheckResourceAttr(testAccNetworkId, "description", ""),
						resource.TestCheckResourceAttr(testAccNetworkId, "region_id", regionID),
						resource.TestCheckResourceAttr(testAccNetworkId, "cloud_type", "aws"),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.vpc_id", vpcID),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.account_id", accountID),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.iam_role_arn", roleARN),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.iam_policy_permission_boundary_arn", permissionBoundaryARN),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.private_subnets", "true"),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv4_cidr_block", ipv4),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv6_cidr_block", ipv6),
						resource.TestCheckResourceAttr(testAccNetworkId, "is_external", "true"),
					),
				},
			},
		})
	})

	t.Run("import AWS network", func(t *testing.T) {
		f.importMock = nil
		f.getMock = awsGetMock

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					ImportState:   true,
					ResourceName:  testAccNetworkId,
					ImportStateId: networkID,
					Config:        testAWSNetworkResourceConfig(&m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(testAccNetworkId, "id", networkID),
						resource.TestCheckResourceAttr(testAccNetworkId, "name", testAccNetworkName),
						resource.TestCheckResourceAttr(testAccNetworkId, "project_id", testProjectId),
						resource.TestCheckResourceAttr(testAccNetworkId, "description", ""),
						resource.TestCheckResourceAttr(testAccNetworkId, "region_id", regionID),
						resource.TestCheckResourceAttr(testAccNetworkId, "cloud_type", "aws"),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.vpc_id", vpcID),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.account_id", accountID),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.iam_role_arn", roleARN),
						resource.TestCheckResourceAttr(testAccNetworkId, "aws.private_subnets", "true"),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv4_cidr_block", ipv4),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv6_cidr_block", ipv6),
						resource.TestCheckResourceAttr(testAccNetworkId, "is_external", "true"),
					),
				},
			},
		})
	})

	t.Run("create GCP network", func(t *testing.T) {
		f.importMock = gcpImportMock
		f.getMock = gcpGetMock
		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					Config: testGCPNetworkResourceConfig(&m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(testAccNetworkId, "id", networkID),
						resource.TestCheckResourceAttr(testAccNetworkId, "name", testAccNetworkName),
						resource.TestCheckResourceAttr(testAccNetworkId, "project_id", testProjectId),
						resource.TestCheckResourceAttr(testAccNetworkId, "description", ""),
						resource.TestCheckResourceAttr(testAccNetworkId, "region_id", regionID),
						resource.TestCheckResourceAttr(testAccNetworkId, "cloud_type", "gcp"),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.network_name", netName),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.subnetwork_name", subnetName),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.project_name", projectName),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.service_account_email", saName),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv4_cidr_block", ipv4),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv6_cidr_block", ipv6),
						resource.TestCheckResourceAttr(testAccNetworkId, "is_external", "true"),
					),
				},
			},
		})
	})

	t.Run("import GCP network", func(t *testing.T) {
		f.importMock = nil
		f.getMock = gcpGetMock

		resource.UnitTest(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
			Steps: []resource.TestStep{
				{
					ImportState:   true,
					ResourceName:  testAccNetworkId,
					ImportStateId: networkID,
					Config:        testGCPNetworkResourceConfig(&m),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(testAccNetworkId, "id", networkID),
						resource.TestCheckResourceAttr(testAccNetworkId, "name", testAccNetworkName),
						resource.TestCheckResourceAttr(testAccNetworkId, "project_id", testProjectId),
						resource.TestCheckResourceAttr(testAccNetworkId, "description", ""),
						resource.TestCheckResourceAttr(testAccNetworkId, "region_id", regionID),
						resource.TestCheckResourceAttr(testAccNetworkId, "cloud_type", "aws"),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.network_name", netName),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.subnetwork_name", subnetName),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.project_name", projectName),
						resource.TestCheckResourceAttr(testAccNetworkId, "gcp.service_account_email", saName),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv4_cidr_block", ipv4),
						resource.TestCheckResourceAttr(testAccNetworkId, "ipv6_cidr_block", ipv6),
						resource.TestCheckResourceAttr(testAccNetworkId, "is_external", "true"),
					),
				},
			},
		})
	})
}

func testAWSNetworkResourceConfig(m *NetworkResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_network" %[2]q {
  project_id = %[1]q
  name = %[2]q
  region_id = %[6]q
  cloud_type = "aws"
  aws = {
    vpc_id = %[3]q
    account_id = %[4]q
    iam_role_arn = %[5]q
    iam_policy_permission_boundary_arn = %[8]q
	private_subnets = %[7]v
  }
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
		m.AWS.VPCID.ValueString(),
		m.AWS.AccountID.ValueString(),
		m.AWS.IAMRoleARN.ValueString(),
		m.RegionID.ValueString(),
		m.AWS.PrivateSubnets,
		m.AWS.IAMPolicyPermissionBoundaryARN.ValueString(),
	)
}

func testGCPNetworkResourceConfig(m *NetworkResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_network" %[2]q {
  project_id = %[1]q
  name = %[2]q
  region_id = %[6]q
  cloud_type = "gcp"
  gcp = {
    network_name = %[3]q
    subnetwork_name = %[4]q
    project_name = %[5]q
	service_account_email = %[7]v
  }
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
		m.GCP.NetworkName.ValueString(),
		m.GCP.SubnetworkName.ValueString(),
		m.GCP.ProjectName.ValueString(),
		m.RegionID.ValueString(),
		m.GCP.SAEmail,
	)
}

func testAccNetworkResourceConfig(m *NetworkResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_network" %[2]q {
  project_id = %[1]q
  name = %[2]q
  region_id = %[3]q
  ipv4_cidr_block = %[4]q
  cloud_type = %[5]q
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
		m.RegionID.ValueString(),
		m.Ipv4CidrBlock.ValueString(),
		m.CloudType.ValueString())
}

func init() {
	resource.AddTestSweepers("network", &resource.Sweeper{
		Name:         "network",
		F:            sweepNetworks,
		Dependencies: []string{},
	})
}

func sweepNetworks(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return err
	}

	var errs error
	rq := &network.ListNetworksRequest{ProjectId: conf.ProjectId}
	svc := conf.sdk.Network().Network()
	it := svc.NetworkIterator(conf.ctx, rq)

	for it.Next() {
		v := it.Value()
		if strings.HasPrefix(v.Name, testPrefix) {
			err := sweepNetwork(conf, v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to sweep %v: %v", v.Id, err))
			}
		}
	}
	return errs
}

func sweepNetwork(conf *Config, t *network.Network) error {
	op, err := conf.sdk.WrapOperation(conf.sdk.Network().Network().Delete(conf.ctx, &network.DeleteNetworkRequest{NetworkId: t.Id}))
	if err != nil {
		return err
	}
	return op.Wait(conf.ctx)
}

const (
	bothExternalAndCidrNetworkConfig = `
resource "doublecloud_network" "bothExternalAndCidrNetworkConfig" {
  project_id = "p"
  name = "n"
  region_id = "r"
  cloud_type = "aws"
  ipv4_cidr_block = "10.0.0.0/16"
  aws = {
    vpc_id = "v"
    account_id = "a"
    iam_role_arn = "i"
    iam_policy_permission_boundary_arn = "p"
  }
}
`
	bothExternalProvidersNetworkConfig = `
resource "doublecloud_network" "bothExternalAndCidrNetworkConfig" {
  project_id = "p"
  name = "n"
  region_id = "r"
  cloud_type = "aws"
  aws = {
    vpc_id = "v"
    account_id = "a"
    iam_role_arn = "i"
    iam_policy_permission_boundary_arn = "p"
  }
  gcp = {
    network_name = "n"
    subnetwork_name = "s"
    project_name = "p"
	service_account_email = "sa"
  }
}
`
	cloudTypeMismatchNetworkConfig = `
resource "doublecloud_network" "cloudTypeMismatchNetworkConfig" {
  project_id = "p"
  name = "n"
  region_id = "r"
  cloud_type = "gcp"
  aws = {
    vpc_id = "v"
    account_id = "a"
    iam_role_arn = "i"
    iam_policy_permission_boundary_arn = "p"
  }
}
`
	ipv4MissedNetworkConfig = `
resource "doublecloud_network" "ipv4MissedNetworkConfig" {
  project_id = "p"
  name = "n"
  region_id = "r"
  cloud_type = "aws"
}
`
)
