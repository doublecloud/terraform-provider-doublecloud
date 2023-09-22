package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/network/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
				),
			},
			// Update not supported
			// Delete testing automatically occurs in TestCase
		},
	})
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
