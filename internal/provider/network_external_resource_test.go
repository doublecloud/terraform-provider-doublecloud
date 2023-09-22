package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testAccExternalNetworkId string = fmt.Sprintf("doublecloud_external_network.%v", testAccNetworkName)
)

func TestAccExternalNetworkResource(t *testing.T) {
	t.Parallel()
	m := ExternalNetworkResourceModel{
		ProjectID: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccNetworkName),
		AWS: &awsExternalNetworkResourceModel{
			VPCID:      types.StringValue("vpc-074db6dce26ab6541"),
			RegionID:   types.StringValue("eu-central-1"),
			AccountID:  types.StringValue("118027436691"),
			IAMRoleARN: types.StringValue("arn:aws:iam::118027436691:role/DoubleCloud/import-vpc-074db6dce26ab6541"),
			//PrivateSubnets: types.Bool{},
		},
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config:      testAccExternalNetworkResourceWrongConfig(&m),
				ExpectError: regexp.MustCompile(`No attribute specified when one \(and only one\) of \[aws,google] is required`),
			},
			{
				Config: testAccExternalNetworkResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.vpc_id", m.AWS.VPCID.ValueString()),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.region_id", m.AWS.RegionID.ValueString()),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.account_id", m.AWS.AccountID.ValueString()),
					resource.TestCheckResourceAttr(testAccExternalNetworkId, "aws.iam_role_arn", m.AWS.IAMRoleARN.ValueString()),
					resource.TestCheckResourceAttrSet(testAccExternalNetworkId, "ipv4_cidr_block"),
					resource.TestCheckResourceAttrSet(testAccExternalNetworkId, "ipv6_cidr_block"),
				),
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
