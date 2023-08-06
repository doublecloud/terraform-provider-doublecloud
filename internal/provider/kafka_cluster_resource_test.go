package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/kafka/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testAccKafkaName string = fmt.Sprintf("%v-kafka", testPrefix)
)

func TestAccKafkaClusterResource(t *testing.T) {
	return
	m := KafkaClusterModel{
		ProjectID: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccKafkaName),
		RegionID:  types.StringValue("eu-central-1"),
		CloudType: types.StringValue("aws"),
		NetworkId: types.StringValue(testNetworkId),

		Resources: KafkaResourcesModel{
			Kafka: KafkaResourcesKafkaModel{
				ResourcePresetId: types.StringValue("s1-c2-m4"),
				DiskSize:         types.Int64Value(34359738368),
				BrokerCount:      types.Int64Value(1),
				ZoneCount:        types.Int64Value(1),
			},
		},


		SchemaRegistry: &schemaRegistryModel{
			Enabled: types.BoolValue(false),
		},
	}

	m2 := KafkaClusterModel(m)
	m2.Name = types.StringValue("terraform-kafka-changed")
	m2.Resources.Kafka.DiskSize = types.Int64Value(51539607552)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKafkaClusterResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("doublecloud_kafka_cluster.test", "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr("doublecloud_kafka_cluster.test", "name", testAccKafkaName),
					resource.TestCheckResourceAttr("doublecloud_kafka_cluster.test", "resources.kafka.disk_size", "34359738368"),
					// resource.TestCheckResourceAttr("doublecloud_kafka_cluster.test", "encryption.enabled", "true"),
					resource.TestCheckResourceAttr("doublecloud_kafka_cluster.test", "schema_registry.enabled", "false"),
				),
			},
			// Update and Read testing
			// {
			// 	Config: testAccKafkaClusterResourceConfigUpdated(&m2),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("doublecloud_kafka_cluster.test", "name", "terraform-kafka-changed"),
			// 		resource.TestCheckResourceAttr("doublecloud_kafka_cluster.test", "resources.kafka.disk_size", "51539607552"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

// func testAccKafkaUsersResourceConfig(m *KafkaClusterModel) string {
// 	if m.Users.IsNull() {
// 		return "empty"
// 	}
// 	var users string

// 	for _, u := range m.Users.Elements() {
// 		user := u.(types.Object)
// 		name := user.Attributes()["name"]
// 		password := user.Attributes()["password"]
// 		// permissions := user.Attributes()["permissions"]

// 		users = fmt.Sprintf(`
//   user {
// 	name = %[1]q
// 	password = %[2]q
//   }
// 		`, name, password)
// 	}
// 	return users
// }

func testAccKafkaClusterResourceConfig(m *KafkaClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_kafka_cluster" "test" {
  project_id = %[1]q
  name = %[2]q
  region_id = %[3]q
  cloud_type = %[4]q
  network_id = %[5]q

  resources {
	kafka {
		resource_preset_id = %[6]q
		disk_size =  %[7]q
		broker_count = %[8]q
		zone_count =  %[9]q
	}
  }

  schema_registry {
	enabled = false
  }

  user {
	name = "alice"
	password = "foobar123"
	permission {
	  topic = "events"
	  role = "ACCESS_ROLE_PRODUCER"
	}
  }

  user {
	name = "bob"
	password = "foobar124"

	permission {
		topic = "transactions"
		role = "ACCESS_ROLE_PRODUCER"
	  }
  }
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
		m.RegionID.ValueString(),
		m.CloudType.ValueString(),
		m.NetworkId.ValueString(),
		m.Resources.Kafka.ResourcePresetId.ValueString(),
		m.Resources.Kafka.DiskSize.String(),
		m.Resources.Kafka.BrokerCount.String(),
		m.Resources.Kafka.ZoneCount.String(),
	)
}

func testAccKafkaClusterResourceConfigUpdated(m *KafkaClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_kafka_cluster" "test" {
  project_id = %[1]q
  name = %[2]q
  region_id = %[3]q
  cloud_type = %[4]q
  network_id = %[5]q

  resources {
	kafka {
		resource_preset_id = %[6]q
		disk_size =  %[7]q
		broker_count = %[8]q
		zone_count =  %[9]q
	}
  }

  schema_registry {
	enabled = true
  }

  user {
	name = "alice"
	password = "foobar123"
	permission {
	  topic = "events"
	  role = "producer"
	}
	permission {
	  topic = "transactions"
	  role = "consumer"
	}
  }

  user {
	name = "bob"
	password = "foobar125"

	permission {
		topic = "transactions"
		role = "producer"
	  }
  }
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
		m.RegionID.ValueString(),
		m.CloudType.ValueString(),
		m.NetworkId.ValueString(),
		m.Resources.Kafka.ResourcePresetId.ValueString(),
		m.Resources.Kafka.DiskSize.String(),
		m.Resources.Kafka.BrokerCount.String(),
		m.Resources.Kafka.ZoneCount.String(),
	)
}

func init() {
	resource.AddTestSweepers("kafka", &resource.Sweeper{
		Name:         "kafka",
		F:            sweepClickhouses,
		Dependencies: []string{},
	})
}

func sweepKafkas(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return err
	}

	var errs error
	rq := &kafka.ListClustersRequest{ProjectId: conf.ProjectId}
	svc := conf.sdk.Kafka().Cluster()
	it := svc.ClusterIterator(conf.ctx, rq)

	for it.Next() {
		v := it.Value()
		if strings.HasPrefix(v.Name, testPrefix) {
			err := sweepKafka(conf, v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to sweep %v: %v", v.Id, err))
			}
		}
	}
	return errs
}

func sweepKafka(conf *Config, t *kafka.Cluster) error {
	op, err := conf.sdk.WrapOperation(conf.sdk.Kafka().Cluster().Delete(conf.ctx, &kafka.DeleteClusterRequest{ClusterId: t.Id}))
	if err != nil {
		return err
	}
	return op.Wait(conf.ctx)
}
