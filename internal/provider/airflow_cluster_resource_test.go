package provider

import (
	"errors"
	"fmt"
	"github.com/doublecloud/go-genproto/doublecloud/airflow/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"strings"
	"testing"
)

var (
	testAccAirflowName string = fmt.Sprintf("%v-airflow", testPrefix)
	testAccAirflowId   string = fmt.Sprintf("doublecloud_airflow_cluster.%v", testAccAirflowName)
)

func TestAccAirflowClusterResource(t *testing.T) {
	t.Parallel()

	// Initial configuration for the Airflow cluster resource
	a := AirflowClusterModel{
		ProjectID: types.StringValue(testProjectId),
		Name:      types.StringValue(testAccAirflowName),
		RegionID:  types.StringValue("eu-central-1"),
		CloudType: types.StringValue("aws"),
		NetworkId: types.StringValue(testNetworkId),

		Resources: &AirflowResourcesModel{
			Airflow: AirflowResourcesAirflowModel{
				MaxWorkerCount:    types.Int64Value(1),
				MinWorkerCount:    types.Int64Value(1),
				EnvironmentFlavor: types.StringValue("small"),
				WorkerConcurrency: types.Int64Value(16),
				WorkerDiskSize:    types.Int64Value(10),
				WorkerPreset:      types.StringValue("small"),
			},
		},

		Config: &AirflowClusterConfigModel{
			VersionId: types.StringValue("2.9.0"),
			syncConfig: &AirflowClusterSyncConfigModel{
				RepoUrl:  types.StringValue("https://github.com/apache/airflow"),
				Branch:   types.StringValue("main"),
				DagsPath: types.StringValue("airflow/example_dags"),
			},
		},
	}

	// Updated configuration for the Airflow cluster resource
	a2 := a
	a2.Name = types.StringValue("terraform-airflow-changed")
	r1 := *a.Resources
	r2 := r1
	a2.Resources = &r2
	a2.Resources.Airflow.EnvironmentFlavor = types.StringValue("medium")

	// Run the acceptance test
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Initial creation of the resource
				Config: testAccAirflowClusterResourceConfig(&a),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccAirflowId, "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr(testAccAirflowId, "name", testAccAirflowName),
					resource.TestCheckResourceAttr(testAccAirflowId, "cloud_type", "aws"),
					resource.TestCheckResourceAttr(testAccAirflowId, "access.data_services.0", "transfer"),
					resource.TestCheckResourceAttr(testAccAirflowId, "access.ipv4_cidr_blocks.0.value", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(testAccAirflowId, "access.ipv4_cidr_blocks.0.description", "Office in Berlin"),
					resource.TestCheckResourceAttr(testAccAirflowId, "resources.airflow.environment_flavor", "small"),
				),
			},
			{
				// Update the resource with new attributes
				Config: testAccAirflowClusterResourceConfig(&a2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccAirflowId, "name", "terraform-airflow-changed"),
					resource.TestCheckResourceAttr(testAccAirflowId, "resources.airflow.environment_flavor", "medium"),
					resource.TestCheckResourceAttr(testAccAirflowId, "region_id", "eu-central-1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Helper function to create Terraform configuration for Airflow Cluster Resource
func testAccAirflowClusterResourceConfig(a *AirflowClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_airflow_cluster" "test" {
  project_id = %[1]q
  name       = %[2]q
  region_id  = %[3]q
  cloud_type = %[4]q
  network_id = %[5]q

  resources {
    airflow {
      max_worker_count    = %[6]d
      min_worker_count    = %[7]d
      environment_flavor  = %[8]q
      worker_concurrency  = %[9]d
      worker_disk_size    = %[10]d
      worker_preset       = %[11]q
    }
  }

  config {
    version_id = %[12]q
    sync_config {
      repo_url  = %[13]q
      branch    = %[14]q
      dags_path = %[15]q
    }
  }

  access {
    data_services = ["transfer"]
    ipv4_cidr_blocks = [
      {
        value       = "10.0.0.0/8"
        description = "Office in Berlin"
      }
    ]
  }
}
`, a.ProjectID.ValueString(),
		a.Name.ValueString(),
		a.RegionID.ValueString(),
		a.CloudType.ValueString(),
		a.NetworkId.ValueString(),
		a.Resources.Airflow.MaxWorkerCount.ValueInt64(),
		a.Resources.Airflow.MinWorkerCount.ValueInt64(),
		a.Resources.Airflow.EnvironmentFlavor.ValueString(),
		a.Resources.Airflow.WorkerConcurrency.ValueInt64(),
		a.Resources.Airflow.WorkerDiskSize.ValueInt64(),
		a.Resources.Airflow.WorkerPreset.ValueString(),
		a.Config.VersionId.ValueString(),
		a.Config.syncConfig.RepoUrl.ValueString(),
		a.Config.syncConfig.Branch.ValueString(),
		a.Config.syncConfig.DagsPath.ValueString(),
	)
}

func init() {
	resource.AddTestSweepers("airflow", &resource.Sweeper{
		Name:         "airflow",
		F:            sweepAirflows,
		Dependencies: []string{},
	})
}

func sweepAirflows(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return err
	}

	var errs error
	rq := &airflow.ListClustersRequest{ProjectId: conf.ProjectId}
	svc := conf.sdk.Airflow().Cluster()
	it := svc.ClusterIterator(conf.ctx, rq)

	for it.Next() {
		v := it.Value()
		if strings.HasPrefix(v.Name, testPrefix) {
			err := sweepAirflow(conf, v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to sweep %v: %v", v.Id, err))
			}
		}
	}
	return errs
}

func sweepAirflow(conf *Config, t *airflow.Cluster) error {
	_, err := conf.sdk.Airflow().Cluster().Delete(conf.ctx, &airflow.DeleteClusterRequest{ClusterId: t.Id})
	return err
}
