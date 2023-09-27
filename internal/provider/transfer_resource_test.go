package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testTChSourceName string = fmt.Sprintf("%v-transfer-ch-source", testPrefix)
	testTChTargetName string = fmt.Sprintf("%v-transfer-ch-target", testPrefix)
	testTransferName  string = fmt.Sprintf("%v-transfer", testPrefix)

	testTChSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testTChSourceName)
	testTChTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testTChTargetName)
	testTransferId  string = fmt.Sprintf("doublecloud_transfer.%v", testTransferName)
)

func TestAccTransferResource(t *testing.T) {
	t.Parallel()

	m := TransferResourceModel{
		ProjectID:   types.StringValue(testProjectId),
		Name:        types.StringValue(testTransferName),
		Description: types.StringValue("transfer description"),
		Type:        types.StringValue("SNAPSHOT_ONLY"),
		Activated:   types.BoolValue(false),
	}

	m2 := TransferResourceModel{
		ProjectID:   m.ProjectID,
		Name:        types.StringValue(fmt.Sprintf("%v-updated", testTransferName)),
		Description: m.Description,
		Type:        m.Type,
		Activated:   m.Activated,
	}

	m3 := TransferResourceModel{
		ProjectID:   m.ProjectID,
		Name:        m2.Name,
		Description: m.Description,
		Type:        m.Type,
		Activated:   types.BoolValue(true),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTChSourceId, "name", testTChSourceName),
					resource.TestCheckResourceAttr(testTChTargetId, "name", testTChTargetName),

					resource.TestCheckResourceAttr(testTransferId, "name", m.Name.ValueString()),
					resource.TestCheckResourceAttr(testTransferId, "description", m.Description.ValueString()),
					resource.TestCheckResourceAttr(testTransferId, "type", m.Type.ValueString()),
					resource.TestCheckResourceAttr(testTransferId, "activated", m.Activated.String()),

					resource.TestCheckResourceAttrSet(testTransferId, "source"),
					resource.TestCheckResourceAttrSet(testTransferId, "target"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferResourceConfig(&m2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTransferId, "name", m2.Name.ValueString()),
					resource.TestCheckResourceAttr(testTransferId, "description", m2.Description.ValueString()),

					resource.TestCheckResourceAttr(testTransferId, "type", m2.Type.ValueString()),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferResourceConfig(&m3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTransferId, "name", m2.Name.ValueString()),

					resource.TestCheckResourceAttr(testTransferId, "activated", m3.Activated.String()),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferResourceConfig(m *TransferResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_transfer" "tf-acc-transfer" {
	project_id = %[1]q
	name = %[2]q
	description = %[3]q
	source = %[8]s.id
	target = %[9]s.id
	type = %[4]q
	activated = %[5]q
}

resource "doublecloud_transfer_endpoint" %[6]q {
	project_id = %[1]q
	name = %[6]q
	settings {
		clickhouse_source {
			connection {
				address {
					cluster_id = "cluster-foo-id"
				}
				database = "default"
				user = "admin"
				password = "foobar123"	
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[7]q {
  project_id = %[1]q
  name = %[7]q
  settings {
	clickhouse_target {
			connection {
				address {
					cluster_id = "cluster-foo-id2"
				}
				database = "default"
				user = "admin"
				password = "foobar123"	
			}
		}
	}
}
`, m.ProjectID.ValueString(), m.Name.ValueString(), m.Description.ValueString(), m.Type.ValueString(), m.Activated.String(),
		testTChSourceName, testTChTargetName, testTChSourceId, testTChTargetId)
}

func init() {
	resource.AddTestSweepers("transfer", &resource.Sweeper{
		Name:         "transfer",
		F:            sweepTransfers,
		Dependencies: []string{},
	})
}

func sweepTransfers(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return err
	}

	var errs error
	rq := &transfer.ListTransfersRequest{ProjectId: conf.ProjectId}
	svc := conf.sdk.Transfer().Transfer()
	it := svc.TransferIterator(conf.ctx, rq)

	for it.Next() {
		v := it.Value()
		if strings.HasPrefix(v.Name, testPrefix) {
			err := sweepTransfer(conf, v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to sweep %v: %v", v.Id, err))
			}
		}
	}
	return errs
}

func sweepTransfer(conf *Config, t *transfer.Transfer) error {
	op, err := conf.sdk.WrapOperation(conf.sdk.Transfer().Transfer().Delete(conf.ctx, &transfer.DeleteTransferRequest{TransferId: t.Id}))
	if err != nil {
		return err
	}
	return op.Wait(conf.ctx)
}
