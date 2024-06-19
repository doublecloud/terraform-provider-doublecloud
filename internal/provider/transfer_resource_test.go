package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferResource(t *testing.T) {
	t.Parallel()

	const testTransferResource = "doublecloud_transfer.ttr-transfer"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: (testTransferResourceEndpointsConfig() +
					"\n\n" +
					fmt.Sprintf(`resource "doublecloud_transfer" "ttr-transfer" {
						project_id = %[1]q
						name = "ttr-transfer"
						description = "test description"
						source = doublecloud_transfer_endpoint.ttr-src-pg.id
						target = doublecloud_transfer_endpoint.ttr-dst-ch.id
						type = "SNAPSHOT_ONLY"
						activated = false
						transformation = {
							transformers = [
								{
									replace_primary_key = {
										tables = {
											include = ["t1"]
											exclude = ["t2"]
										}
										keys = [
											"pk_field_1",
											"pk_field_2"
										]
									}
								},
								{
									convert_to_string = {
										tables = {
											include = ["t1"]
											exclude = ["t2"]
										}
										columns = {
											include = ["c1"]
											exclude = ["c2"]
										}
									}
								},
								{
									lambda_function = {
										name = "cloud-function-example"
										name_space = "example-namespace"
										options = {
											cloud_function = "example-cloud-function"
											cloud_function_url = "https://example.com/function"
											number_of_retries = 3
											buffer_size = "10 MB"
											buffer_flush_interval = "1m0s"
											invocation_timeout = "30s"
											headers = [
												{
													key = "Authorization"
													value = "Bearer example-token"
												},
												{
													key = "Content-Type"
													value = "application/json"
												}
											]
										}
									}
								}
							]
						}
					}`, testProjectId)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTransferResource, "name", "ttr-transfer"),
					resource.TestCheckResourceAttr(testTransferResource, "description", "test description"),
					resource.TestCheckResourceAttrSet(testTransferResource, "source"),
					resource.TestCheckResourceAttrSet(testTransferResource, "target"),
					resource.TestCheckResourceAttr(testTransferResource, "type", "SNAPSHOT_ONLY"),
					resource.TestCheckResourceAttr(testTransferResource, "activated", "false"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.#", "3"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.replace_primary_key.tables.include.0", "t1"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.replace_primary_key.tables.exclude.0", "t2"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.replace_primary_key.keys.0", "pk_field_1"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.replace_primary_key.keys.1", "pk_field_2"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.1.convert_to_string.tables.include.0", "t1"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.1.convert_to_string.tables.exclude.0", "t2"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.1.convert_to_string.columns.include.0", "c1"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.1.convert_to_string.columns.exclude.0", "c2"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.name", "cloud-function-example"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.name_space", "example-namespace"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.cloud_function", "example-cloud-function"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.cloud_function_url", "https://example.com/function"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.number_of_retries", "3"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.buffer_size", "10 MB"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.buffer_flush_interval", "1m0s"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.invocation_timeout", "30s"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.headers.0.key", "Authorization"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.headers.0.value", "Bearer example-token"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.headers.1.key", "Content-Type"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.2.lambda_function.options.headers.1.value", "application/json"),
				),
			},
			{
				Config: (testTransferResourceEndpointsConfig() +
					"\n\n" +
					fmt.Sprintf(`resource "doublecloud_transfer" "ttr-transfer" {
						project_id = %[1]q
						name = "ttr-transfer"
						description = "test description"
						source = doublecloud_transfer_endpoint.ttr-src-pg.id
						target = doublecloud_transfer_endpoint.ttr-dst-ch.id
						type = "SNAPSHOT_ONLY"
						activated = false
						transformation = {
							transformers = [
								{
									convert_to_string = {
										tables = {
											include = ["t1"]
											exclude = ["t2"]
										}
										columns = {
											include = ["c1"]
											exclude = ["c2"]
										}
									}
								},
							]
						}
					}`, testProjectId)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.#", "1"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.convert_to_string.tables.include.0", "t1"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.convert_to_string.tables.exclude.0", "t2"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.convert_to_string.columns.include.0", "c1"),
					resource.TestCheckResourceAttr(testTransferResource, "transformation.transformers.0.convert_to_string.columns.exclude.0", "c2"),
				),
			},
			{
				Config: (testTransferResourceEndpointsConfig() +
					"\n\n" +
					fmt.Sprintf(`resource "doublecloud_transfer" "ttr-transfer" {
						project_id = %[1]q
						name = "ttr-transfer"
						description = "test description"
						source = doublecloud_transfer_endpoint.ttr-src-pg.id
						target = doublecloud_transfer_endpoint.ttr-dst-ch.id
						type = "SNAPSHOT_ONLY"
						activated = false
						runtime = {
							dedicated = {
								flavor = "TINY"
							}
						}
					}`, testProjectId)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTransferResource, "runtime.dedicated.flavor", "TINY"),
				),
			},
			{
				Config: (testTransferResourceEndpointsConfig() +
					"\n\n" +
					fmt.Sprintf(`resource "doublecloud_transfer" "ttr-transfer" {
						project_id = %[1]q
						name = "ttr-transfer"
						description = "test description"
						source = doublecloud_transfer_endpoint.ttr-src-pg.id
						target = doublecloud_transfer_endpoint.ttr-dst-ch.id
						type = "SNAPSHOT_ONLY"
						activated = false
						runtime = {
							dedicated = {
								flavor = "TINY"
							}
						}
						data_objects = ["foo.barovich", "bar.fooovich"]
					}`, testProjectId)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testTransferResource, "data_objects.0", "foo.barovich"),
					resource.TestCheckResourceAttr(testTransferResource, "data_objects.1", "bar.fooovich"),
				),
			},
			// Delete occurs automatically
		},
	})
}

func testTransferResourceEndpointsConfig() string {
	return fmt.Sprintf(
		`resource "doublecloud_transfer_endpoint" "ttr-src-pg" {
			project_id = %[1]q
			name = "ttr-src-pg"
			settings {
				postgres_source {
					connection {
						on_premise {
							hosts = ["leader-0.company.tech"]
							port = 5432
						}
					}
					database = "production"
					user = "dc-transfer"
					password = "foobar123"
				}
			}
		}

		resource "doublecloud_transfer_endpoint" "ttr-dst-ch" {
		project_id = %[1]q
		name = "ttr-dst-ch"
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
		}`,
		testProjectId,
	)
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
