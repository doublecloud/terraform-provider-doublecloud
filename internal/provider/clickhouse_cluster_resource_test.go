package provider

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
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

	testAccClickhouseTLSCert string = `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEcKT/wmDt+qLwEVOfU0UbJO5f77+0
nuYermx15MOZh4jg4H/r98b/tD2dNxdLAW/VJ4VTF3vD0AGY2+xN7J8aTA==
-----END PUBLIC KEY-----
`

	testAccClickhouseTLSKey string = `
-----BEGIN CERTIFICATE-----
MIICoTCCAkegAwIBAgIUWdVSBHIWp+w6Gtmt4Ps+RNgky00wCgYIKoZIzj0EAwIw
gacxCzAJBgNVBAYTAkRFMRIwEAYDVQQIDAlGcmFua2Z1cnQxEjAQBgNVBAcMCUZy
YW5rZnVydDEVMBMGA1UECgwMZG91YmxlLmNsb3VkMSAwHgYDVQQLDBdUZXJyYWZv
cm0gcHJvdmlkZXIgdGVzdDEVMBMGA1UEAwwMZG91YmxlLmNsb3VkMSAwHgYJKoZI
hvcNAQkBFhFpbmZvQGRvdWJsZS5jbG91ZDAeFw0yNDA5MTkxNjE5MDNaFw0yNTA5
MTkxNjE5MDNaMIG0MQswCQYDVQQGEwJERTESMBAGA1UECAwJRnJhbmtmdXJ0MRIw
EAYDVQQHDAlGcmFua2Z1cnQxFTATBgNVBAoMDGRvdWJsZS5jbG91ZDElMCMGA1UE
CwwcVGVycmFmb3JtIHByb3ZpZGVyIHRlc3QgaW1wbDEdMBsGA1UEAwwUdGVzdC5h
dC5kb3VibGUuY2xvdWQxIDAeBgkqhkiG9w0BCQEWEWluZm9AZG91YmxlLmNsb3Vk
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEcKT/wmDt+qLwEVOfU0UbJO5f77+0
nuYermx15MOZh4jg4H/r98b/tD2dNxdLAW/VJ4VTF3vD0AGY2+xN7J8aTKNCMEAw
HQYDVR0OBBYEFElk8x4Sw1IYKahZDqAKrbPrMQvaMB8GA1UdIwQYMBaAFC/+xZgT
4U3lxhcG2wdT5/NlGB7cMAoGCCqGSM49BAMCA0gAMEUCIBWS0StXMJCfOHU6UqKK
PB+UYxG5mwIw4IP/T7sLa3XlAiEAyS8vLtbgrh8mLXwacAe/SFRS3L/DhOJQa+0e
VQBbsVs=
-----END CERTIFICATE-----
`

	testAccClickhouseTLSRootCA string = `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE2fZnlTyuGtgATXh0FmgvgsqTI/aB
Wy2sRShP40UqdTQ4pxLkpkskb7RWssyrXZEiieGSIUY33setFOOMV6b4RA==
-----END PUBLIC KEY-----
`
)

func TestAccClickhouseClusterResource(t *testing.T) {
	t.Parallel()
	m := clickhouseClusterModel{
		ProjectId: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccClickhouseName),
		RegionId:  types.StringValue("eu-central-1"),
		CloudType: types.StringValue("aws"),
		NetworkId: types.StringValue(testNetworkId),
		Version:   types.StringValue("24.8"),
		Resources: &clickhouseClusterResources{
			Clickhouse: &clickhouseClusterResourcesClickhouse{
				ResourcePresetId: types.StringValue("g2-c2-m8"),
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
			ResourcePresetId: types.StringValue("g2-c2-m8"),
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
			SessionTimeoutMs: types.StringValue("1m"),
		},
	}

	m3 := m2
	m3.Resources = &clickhouseClusterResources{
		Clickhouse: &clickhouseClusterResourcesClickhouse{
			MinResourcePresetId: types.StringValue("g2-c2-m8"),
			MaxResourcePresetId: types.StringValue("g2-c4-m16"),
			DiskSize:            types.Int64Value(51539607552),
			MaxDiskSize:         types.Int64Value(68719476736),
			ReplicaCount:        types.Int64Value(1),
		},
	}

	m4 := m3
	m4.CustomCertificate = &clickhouseCustomCertificate{
		Certificate: types.StringValue(testAccClickhouseTLSCert),
		Key:         types.StringValue(testAccClickhouseTLSKey),
		RootCA:      types.StringValue(testAccClickhouseTLSRootCA),
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
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.resource_preset_id", "g2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.disk_size", "34359738368"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.log_level", "LOG_LEVEL_INFORMATION"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.security_protocol", "PLAINTEXT"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.session_timeout_ms", "15s"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.data_services.0", "transfer"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.0.value", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.0.description", "Office in Berlin"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "connection_info.user", "admin"),
					resource.TestMatchResourceAttr(testAccClickhouseId, "connection_info.password", regexp.MustCompile(`\S+`)),
					resource.TestCheckResourceAttr(testAccClickhouseId, "connection_info.https_port", "8443"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "connection_info.tcp_port_secure", "9440"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "private_connection_info.user", "admin"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "private_connection_info.https_port", "8443"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "private_connection_info.tcp_port_secure", "9440"),
				),
			},
			// Update and Read testing
			{
				Config: convertClickHouseModelToHCL(&m2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccClickhouseId, "name", m2.Name.ValueString()),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.resource_preset_id", "g2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.disk_size", "51539607552"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.log_level", "LOG_LEVEL_TRACE"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.max_connections", "120"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.security_protocol", "SASL_SSL"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.sasl_mechanism", "SCRAM_SHA_512"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.sasl_username", "admin"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.sasl_password", "Traffic3-Mushiness-Chariot"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "config.kafka.session_timeout_ms", "1m"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.data_services.0", "transfer"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.1.value", "11.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "access.ipv4_cidr_blocks.1.description", "Office in Cupertino"),
				),
			},
			// Enable autoscaling
			{
				Config: convertClickHouseModelToHCL(&m3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccClickhouseId, "resources.clickhouse.resource_preset_id"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.min_resource_preset_id", "g2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.max_resource_preset_id", "g2-c4-m16"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.clickhouse.max_disk_size", "68719476736"),
				),
			},
			// Check custom TLS certificate
			{
				Config: convertClickHouseModelToHCL(&m4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccClickhouseId, "custom_certificate.certificate", testAccClickhouseTLSCert),
					resource.TestCheckResourceAttr(testAccClickhouseId, "custom_certificate.key", testAccClickhouseTLSKey),
					resource.TestCheckResourceAttr(testAccClickhouseId, "custom_certificate.root_ca", testAccClickhouseTLSRootCA),
					resource.TestCheckResourceAttr(testAccClickhouseId, "connection_info.https_port_ctls", "8444"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "connection_info.tcp_port_secure_ctls", "9444"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "private_connection_info.https_port_ctls", "8444"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "private_connection_info.tcp_port_secure_ctls", "9444"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccClickhouseDedicatedKeeperClusterResource(t *testing.T) {
	t.Parallel()
	m := clickhouseClusterModel{
		ProjectId: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccClickhouseName + "-keeper"),
		RegionId:  types.StringValue("eu-central-1"),
		CloudType: types.StringValue("aws"),
		NetworkId: types.StringValue(testNetworkId),
		Resources: &clickhouseClusterResources{
			Clickhouse: &clickhouseClusterResourcesClickhouse{
				ResourcePresetId: types.StringValue("g2-c2-m8"),
				DiskSize:         types.Int64Value(34359738368),
				ReplicaCount:     types.Int64Value(1),
			},
			Keeper: &clickhouseClusterResourcesKeeper{
				ResourcePresetId: types.StringValue("g2-c2-m8"),
				DiskSize:         types.Int64Value(34359738368),
				ReplicaCount:     types.Int64Value(1),
			},
		},
	}

	m2 := m
	m2.Resources = &clickhouseClusterResources{
		Clickhouse: m.Resources.Clickhouse,
		Keeper: &clickhouseClusterResourcesKeeper{
			MinResourcePresetId: types.StringValue("g2-c2-m8"),
			MaxResourcePresetId: types.StringValue("g2-c4-m16"),
			DiskSize:            types.Int64Value(34359738368),
			MaxDiskSize:         types.Int64Value(51539607552),
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
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.dedicated_keeper.resource_preset_id", "g2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.dedicated_keeper.disk_size", "34359738368"),
				),
			},
			// Enable autoscaling
			{
				Config: convertClickHouseModelToHCL(&m2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(testAccClickhouseId, "resources.dedicated_keeper.resource_preset_id"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.dedicated_keeper.min_resource_preset_id", "g2-c2-m8"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.dedicated_keeper.max_resource_preset_id", "g2-c4-m16"),
					resource.TestCheckResourceAttr(testAccClickhouseId, "resources.dedicated_keeper.max_disk_size", "51539607552"),
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
	{{- if not .Version.IsNull }}
    version    = "{{ .Version.ValueString }}"{{end}}

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
      {{- if ne .Resources.Keeper nil }}
      dedicated_keeper {
          {{- if not .Resources.Keeper.ResourcePresetId.IsNull }}
          resource_preset_id     = "{{ .Resources.Keeper.ResourcePresetId.ValueString }}"{{end}}
          {{- if not .Resources.Keeper.MinResourcePresetId.IsNull }}
          min_resource_preset_id = "{{ .Resources.Keeper.MinResourcePresetId.ValueString }}"{{end}}
          {{- if not .Resources.Keeper.MaxResourcePresetId.IsNull }}
          max_resource_preset_id = "{{ .Resources.Keeper.MaxResourcePresetId.ValueString }}"{{end}}
          disk_size              = {{ .Resources.Keeper.DiskSize.ValueInt64 }}
          {{- if not .Resources.Keeper.MaxDiskSize.IsNull }}
          max_disk_size          = "{{ .Resources.Keeper.MaxDiskSize.ValueInt64 }}"{{end}}
          replica_count          = {{ .Resources.Keeper.ReplicaCount.ValueInt64 }}
      }
      {{- end }}
    }

    config {
    {{- if ne .Config nil }}
      {{- if not .Config.LogLevel.IsNull }}
      log_level = "{{ .Config.LogLevel.ValueString }}"{{end}}
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
     {{- end}}
    }
    access {
      data_services = ["transfer"]

      {{- if ne .Access nil }}
      ipv4_cidr_blocks = [
        {{- $length := len .Access.Ipv4CIDRBlocks }}
        {{- range $i, $block := .Access.Ipv4CIDRBlocks }}
        {
            value = "{{ $block.Value.ValueString }}"
            description = "{{ $block.Description.ValueString }}"
        },
        {{- end}}
      ]
      {{- end}}
    }
	{{- if ne .CustomCertificate nil }}
	custom_certificate {
	  certificate = {{ .CustomCertificate.Certificate }}
	  key = {{ .CustomCertificate.Key }}
	  root_ca = {{ .CustomCertificate.RootCA }}
	}
    {{- end}}
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
