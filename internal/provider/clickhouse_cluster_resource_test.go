package provider

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/doublecloud/go-genproto/doublecloud/clickhouse/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testAccClickhouseName string = fmt.Sprintf("%v-clickhouse", testPrefix)
	testAccClickhouseId   string = fmt.Sprintf("doublecloud_clickhouse_cluster.%v", testAccClickhouseName)
)

func TestAccClickhouseClusterResource(t *testing.T) {
	t.Parallel()
	m := clickhouseClusterModel{
		ProjectId: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccClickhouseName),
		RegionId:  types.StringValue("eu-central-1"),
		CloudType: types.StringValue("aws"),
		NetworkId: types.StringValue(testNetworkId),
		Resources: &clickhouseClusterResources{
			Clickhouse: &clickhouseClusterResourcesClickhouse{
				ResourcePresetId: types.StringValue("s2-c2-m8"),
				DiskSize:         types.Int64Value(34359738368),
				ReplicaCount:     types.Int64Value(1),
			},
		},
		Access: &AccessModel{
			Ipv4CIDRBlocks: []*CIDRBlock{{
				Value:       types.StringValue("10.0.0.0/8"),
				Description: types.StringValue("Office in Berlin"),
			}},
		},
		Config: &clickhouseConfig{
			LogLevel: types.StringValue("LOG_LEVEL_INFORMATION"),
			Kafka: &clickhouseConfigKafka{
				SecurityProtocol: types.StringValue("PLAINTEXT"),
				SessionTimeoutMs: types.StringValue("15s"),
			},
		},
	}

	m2 := m
	m2.Name = types.StringValue(fmt.Sprintf("%v-changed", testAccClickhouseName))
	m2.Resources = &clickhouseClusterResources{
		Clickhouse: &clickhouseClusterResourcesClickhouse{
			ResourcePresetId: types.StringValue("s2-c2-m8"),
			DiskSize:         types.Int64Value(51539607552),
			ReplicaCount:     types.Int64Value(1),
		},
	}
	m2.Access.Ipv4CIDRBlocks = append(m2.Access.Ipv4CIDRBlocks, &CIDRBlock{
		Value:       types.StringValue("11.0.0.0/8"),
		Description: types.StringValue("Office in Cupertino"),
	})
	m2.Config = &clickhouseConfig{
		LogLevel: types.StringValue("LOG_LEVEL_TRACE"),
		Kafka: &clickhouseConfigKafka{
			SecurityProtocol: types.StringValue("SASL_SSL"),
			SaslMechanism:    types.StringValue("SCRAM_SHA_512"),
			SaslUsername:     types.StringValue("admin"),
			SaslPassword:     types.StringValue("Traffic3-Mushiness-Chariot"),
			SessionTimeoutMs: types.StringValue("1m0s"),
		},
	}

	m3 := m2
	m3.Resources = &clickhouseClusterResources{
		Clickhouse: &clickhouseClusterResourcesClickhouse{
			MinResourcePresetId: types.StringValue("s2-c2-m8"),
			MaxResourcePresetId: types.StringValue("s2-c4-m16"),
			DiskSize:            types.Int64Value(51539607552),
			MaxDiskSize:         types.Int64Value(68719476736),
			ReplicaCount:        types.Int64Value(1),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: convertClickHouseModelToHCL(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccClickhouseId, "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "name", m.Name.ValueString()),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.resource_preset_id", "s2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.disk_size", "34359738368"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.log_level", "LOG_LEVEL_INFORMATION"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.security_protocol", "PLAINTEXT"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.session_timeout_ms", "15s"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.data_services.0", "transfer"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.0.value", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.0.description", "Office in Berlin"),
				),
			},
			// Update and Read testing
			{
				Config: convertClickHouseModelToHCL(&m2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccClickhouseId, "name", m2.Name.ValueString()),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.resource_preset_id", "s2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.disk_size", "51539607552"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.log_level", "LOG_LEVEL_TRACE"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.max_connections", "120"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.security_protocol", "SASL_SSL"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.sasl_mechanism", "SCRAM_SHA_512"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.sasl_username", "admin"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.sasl_password", "Traffic3-Mushiness-Chariot"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.session_timeout_ms", "1m0s"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.data_services.0", "transfer"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.1.value", "11.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.1.description", "Office in Cupertino"),
				),
			},
			// Enable autoscaling
			{
				Config: convertClickHouseModelToHCL(&m3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.min_resource_preset_id", "s2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.max_resource_preset_id", "s2-c4-m16"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.max_disk_size", "68719476736"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

const clickhouseHCLTemplateRaw = `
resource "doublecloud_clickhouse_cluster" "tf-acc-clickhouse" {
    project_id = "{{ .ProjectId.ValueString }}"
    name =       "{{ .Name.ValueString }}"
    region_id =  "{{ .RegionId.ValueString }}"
    cloud_type = "{{ .CloudType.ValueString }}"
    network_id = "{{ .NetworkId.ValueString }}"

    resources {
      clickhouse {
          {{- if not .Resources.Clickhouse.ResourcePresetId.IsNull }}
          resource_preset_id     = "{{ .Resources.Clickhouse.ResourcePresetId.ValueString }}"{{end}}
          {{- if not .Resources.Clickhouse.MinResourcePresetId.IsNull }}
          min_resource_preset_id = "{{ .Resources.Clickhouse.MinResourcePresetId.ValueString }}"{{end}}
          {{- if not .Resources.Clickhouse.MaxResourcePresetId.IsNull }}
          max_resource_preset_id = "{{ .Resources.Clickhouse.MaxResourcePresetId.ValueString }}"{{end}}
          disk_size              = {{ .Resources.Clickhouse.DiskSize.ValueInt64 }}
          {{- if not .Resources.Clickhouse.MaxDiskSize.IsNull }}
          max_disk_size          = "{{ .Resources.Clickhouse.MaxDiskSize.ValueInt64 }}"{{end}}
          replica_count          = {{ .Resources.Clickhouse.ReplicaCount.ValueInt64 }}
      }
    }

    config {
      log_level = "{{ .Config.LogLevel.ValueString }}"
      max_connections = 120

      kafka {
          security_protocol  = "{{.Config.Kafka.SecurityProtocol.ValueString}}"
          {{- if not .Config.Kafka.SaslMechanism.IsNull }}
          sasl_mechanism     = "{{ .Config.Kafka.SaslMechanism.ValueString }}"{{ end }}
          {{- if not .Config.Kafka.SaslUsername.IsNull }}
          sasl_username      = "{{ .Config.Kafka.SaslUsername.ValueString }}"{{ end }}
          {{- if not .Config.Kafka.SaslPassword.IsNull }}
          sasl_password      = "{{ .Config.Kafka.SaslPassword.ValueString }}"{{ end }}
          {{- if not .Config.Kafka.SessionTimeoutMs.IsNull }}
          session_timeout_ms = "{{ .Config.Kafka.SessionTimeoutMs.ValueString }}"{{ end }}
      }
    }

    access {
      data_services = ["transfer"]

      ipv4_cidr_blocks = [
{{- $length := len .Access.Ipv4CIDRBlocks }}
{{- range $i, $block := .Access.Ipv4CIDRBlocks }}
        {
            value = "{{ $block.Value.ValueString }}"
            description = "{{ $block.Description.ValueString }}"
        },
{{- end}}
      ]
    }
  }`

var clickhouseHCLTemplate *template.Template

func convertClickHouseModelToHCL(m *clickhouseClusterModel) string {
	var res bytes.Buffer
	if err := clickhouseHCLTemplate.Execute(&res, m); err != nil {
		panic(err)
	}

	return res.String()
}

func init() {
	resource.AddTestSweepers("clickhouse", &resource.Sweeper{
		Name:         "clickhouse",
		F:            sweepClickhouses,
		Dependencies: []string{},
	})

	var err error
	clickhouseHCLTemplate, err = template.New("clickhouse").Funcs(template.FuncMap{
		"inc": func(n int) int {
			return n + 1
		},
	}).Parse(clickhouseHCLTemplateRaw)
	if err != nil {
		panic(err)
	}
}

func sweepClickhouses(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return err
	}

	var errs error
	rq := &clickhouse.ListClustersRequest{ProjectId: conf.ProjectId}
	svc := conf.sdk.ClickHouse().Cluster()
	it := svc.ClusterIterator(conf.ctx, rq)

	for it.Next() {
		v := it.Value()
		if strings.HasPrefix(v.Name, testPrefix) {
			err := sweepClickhouse(conf, v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to sweep %v: %v", v.Id, err))
			}
		}
	}
	return errs
}

func sweepClickhouse(conf *Config, t *clickhouse.Cluster) error {
	_, err := conf.sdk.ClickHouse().Cluster().Delete(conf.ctx, &clickhouse.DeleteClusterRequest{ClusterId: t.Id})
	return err
}
