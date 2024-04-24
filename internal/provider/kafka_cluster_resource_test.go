package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/doublecloud/go-genproto/doublecloud/kafka/v1"
)

var (
	testAccKafkaName string = fmt.Sprintf("%v-kafka", testPrefix)
	testAccKafkaId   string = fmt.Sprintf("doublecloud_kafka_cluster.%v", testAccKafkaName)
)

func TestAccKafkaClusterResource(t *testing.T) {
	t.Parallel()
	m := KafkaClusterModel{
		ProjectID: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccKafkaName),
		RegionID:  types.StringValue("eu-central-1"),
		CloudType: types.StringValue("aws"),
		NetworkId: types.StringValue(testNetworkId),

		Resources: &KafkaResourcesModel{
			Kafka: KafkaResourcesKafkaModel{
				ResourcePresetId: types.StringValue("g2-c2-m4"),
				DiskSize:         types.Int64Value(34359738368),
				BrokerCount:      types.Int64Value(1),
				ZoneCount:        types.Int64Value(1),
			},
		},

		SchemaRegistry: &schemaRegistryModel{
			Enabled: types.BoolValue(false),
		},

		Config: &KafkaClusterConfigModel{
			MessageMaxBytes:   types.Int64Value(1024),
			LogRetentionHours: types.Int64Value(168),
		},
	}

	m2 := m
	m2.Name = types.StringValue("terraform-kafka-changed")
	r1 := *m.Resources
	r2 := r1
	m2.Resources = &r2
	m2.Resources.Kafka.DiskSize = types.Int64Value(51539607552)
	m2.Config.MessageMaxBytes = types.Int64Value(2048)
	m2.Config.LogRetentionHours = types.Int64Value(336)

	m3 := m2
	m3.Resources = &KafkaResourcesModel{
		Kafka: KafkaResourcesKafkaModel{
			MinResourcePresetId: types.StringValue("g2-c2-m8"),
			MaxResourcePresetId: types.StringValue("g2-c4-m16"),
			DiskSize:            types.Int64Value(51539607552),
			MaxDiskSize:         types.Int64Value(137438953472),
			BrokerCount:         types.Int64Value(1),
			ZoneCount:           types.Int64Value(1),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccKafkaClusterResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccKafkaId, "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr(testAccKafkaId, "name", testAccKafkaName),
					resource.TestCheckResourceAttr(testAccKafkaId, "resources.kafka.disk_size", "34359738368"),
					resource.TestCheckResourceAttr(testAccKafkaId, "schema_registry.enabled", "false"),

					resource.TestCheckResourceAttr(testAccKafkaId, "access.data_services.0", "transfer"),
					resource.TestCheckResourceAttr(testAccKafkaId, "access.ipv4_cidr_blocks.0.value", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccKafkaId, "access.ipv4_cidr_blocks.0.description", "Office in Berlin"),

					resource.TestCheckResourceAttrSet(testAccKafkaId, "connection_info.connection_string"),
					resource.TestCheckResourceAttr(testAccKafkaId, "connection_info.user", "admin"),
					resource.TestCheckResourceAttrSet(testAccKafkaId, "connection_info.password"),

					resource.TestCheckResourceAttrSet(testAccKafkaId, "private_connection_info.connection_string"),
					resource.TestCheckResourceAttr(testAccKafkaId, "private_connection_info.user", "admin"),
					resource.TestCheckResourceAttrSet(testAccKafkaId, "private_connection_info.password"),

					resource.TestCheckResourceAttr(testAccKafkaId, "config.message_max_bytes", "1024"),
					resource.TestCheckResourceAttr(testAccKafkaId, "config.log_retention_hours", "168"),
				),
			},
			// Update and Read testing
			{
				Config: testAccKafkaClusterResourceConfigUpdated(&m2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccKafkaId, "name", "terraform-kafka-changed"),
					resource.TestCheckResourceAttr(testAccKafkaId, "resources.kafka.disk_size", "51539607552"),

					resource.TestCheckResourceAttr(testAccKafkaId, "config.message_max_bytes", "2048"),
					resource.TestCheckResourceAttr(testAccKafkaId, "config.log_retention_hours", "336"),
				),
			},
			// Enable autoscaling
			{
				Config: testAccKafkaClusterResourceConfigAutoscalingEnabled(&m3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccKafkaId, "resources.kafka.resource_preset_id"),
					resource.TestCheckResourceAttr(testAccKafkaId, "resources.kafka.min_resource_preset_id", "g2-c2-m8"),
					resource.TestCheckResourceAttr(testAccKafkaId, "resources.kafka.max_resource_preset_id", "g2-c4-m16"),
					resource.TestCheckResourceAttr(testAccKafkaId, "resources.kafka.max_disk_size", "137438953472"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

//nolint:unused
func testAccKafkaClusterResourceConfig(m *KafkaClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_kafka_cluster" "tf-acc-kafka" {
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

  config {
	message_max_bytes = 1024
	log_retention_hours = 168
  }

  schema_registry {
	enabled = false
  }

  access {
	data_services = ["transfer"]

	ipv4_cidr_blocks = [
		{
			value = "10.0.0.0/8"
			description = "Office in Berlin"
		}
	]
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

//nolint:unused
func testAccKafkaClusterResourceConfigUpdated(m *KafkaClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_kafka_cluster" "tf-acc-kafka" {
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

  config {
	message_max_bytes = 2048
	log_retention_hours = 336
  }

  schema_registry {
	enabled = true
  }

  access {
	data_services = ["transfer"]

	ipv4_cidr_blocks = [
		{
			value = "10.0.0.0/8"
			description = "Office in Berlin"
		}
	]
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

//nolint:unused
func testAccKafkaClusterResourceConfigAutoscalingEnabled(m *KafkaClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_kafka_cluster" "tf-acc-kafka" {
  project_id = %[1]q
  name = %[2]q
  region_id = %[3]q
  cloud_type = %[4]q
  network_id = %[5]q

  resources {
    kafka {
      min_resource_preset_id = %[6]q
      max_resource_preset_id = %[7]q
      disk_size =  %[8]q
      max_disk_size  = %[9]q
      broker_count = %[10]q
      zone_count =  %[11]q
    }
  }

  config {
	message_max_bytes = 2048
	log_retention_hours = 336
  }

  schema_registry {
	enabled = true
  }

  access {
	data_services = ["transfer"]

	ipv4_cidr_blocks = [
		{
			value = "10.0.0.0/8"
			description = "Office in Berlin"
		}
	]
  }
}
`, m.ProjectID.ValueString(),
		m.Name.ValueString(),
		m.RegionID.ValueString(),
		m.CloudType.ValueString(),
		m.NetworkId.ValueString(),
		m.Resources.Kafka.MinResourcePresetId.ValueString(),
		m.Resources.Kafka.MaxResourcePresetId.ValueString(),
		m.Resources.Kafka.DiskSize.String(),
		m.Resources.Kafka.MaxDiskSize.String(),
		m.Resources.Kafka.BrokerCount.String(),
		m.Resources.Kafka.ZoneCount.String(),
	)
}

func init() {
	resource.AddTestSweepers("kafka", &resource.Sweeper{
		Name:         "kafka",
		F:            sweepKafkas,
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
	_, err := conf.sdk.Kafka().Cluster().Delete(conf.ctx, &kafka.DeleteClusterRequest{ClusterId: t.Id})
	return err
}
