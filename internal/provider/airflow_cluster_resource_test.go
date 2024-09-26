package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/airflow/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
			Airflow: &AirflowResourcesAirflowModel{
				MaxWorkerCount:    types.Int64Value(1),
				MinWorkerCount:    types.Int64Value(1),
				EnvironmentFlavor: types.StringValue("dev_test"),
				WorkerConcurrency: types.Int64Value(16),
				WorkerDiskSize:    types.Int64Value(10),
				WorkerPreset:      types.StringValue("small"),
			},
		},

		Config: &AirflowClusterConfigModel{
			VersionId: types.StringValue("2.9.0"),
			SyncConfig: &AirflowClusterSyncConfigModel{
				RepoUrl:  types.StringValue("https://github.com/apache/airflow"),
				Branch:   types.StringValue("main"),
				DagsPath: types.StringValue("airflow/example_dags"),
			},
		},
	}
	// Updated configuration for the Airflow cluster resource
	a2 := a
	r2 := *a.Resources.Airflow
	a2.Resources = &AirflowResourcesModel{
		Airflow: &r2,
	}
	a2.Resources.Airflow.MaxWorkerCount = types.Int64Value(3)
	a2.Resources.Airflow.EnvironmentFlavor = types.StringValue("prod")
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

					resource.TestCheckResourceAttr(testAccAirflowId, "resources.airflow.environment_flavor", "dev_test"),
					resource.TestCheckResourceAttr(testAccAirflowId, "resources.airflow.worker_preset", "small"),

					resource.TestCheckResourceAttrSet(testAccAirflowId, "connection_info.host"),
					resource.TestCheckResourceAttr(testAccAirflowId, "connection_info.user", "admin"),
					resource.TestCheckResourceAttrSet(testAccAirflowId, "connection_info.password"),

					resource.TestCheckResourceAttrSet(testAccAirflowId, "cr_connection_info.host"),
					resource.TestCheckResourceAttrSet(testAccAirflowId, "cr_connection_info.password"),
				),
			},
			{
				// Update the resource with new attributes
				Config: testAccAirflowClusterResourceConfig(&a2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccAirflowId, "resources.airflow.max_worker_count", "3"),
					resource.TestCheckResourceAttr(testAccAirflowId, "resources.airflow.environment_flavor", "prod"),
				),
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}

// Helper function to create Terraform configuration for Airflow Cluster Resource
func testAccAirflowClusterResourceConfig(a *AirflowClusterModel) string {
	return fmt.Sprintf(`
resource "doublecloud_airflow_cluster" "tf-acc-airflow" {
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
	  credentials {
	  	api_credentials {
		  username = "test-username"
 		  password = "test-password"
	  	}
	  }
    }
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
		a.Config.SyncConfig.RepoUrl.ValueString(),
		a.Config.SyncConfig.Branch.ValueString(),
		a.Config.SyncConfig.DagsPath.ValueString(),
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
