package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	dc "github.com/doublecloud/go-sdk"
	"github.com/doublecloud/go-sdk/iamkey"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const envProjectId = "DC_PROJECT_ID"
const envAuthkey = "DC_AUTHKEY"
const envNetworkId = "DC_NETWORK_ID"

// These cluster's are used for data source tests
const envClickhouseName = "DC_CLICKHOUSE_NAME"
const envKafkaName = "DC_KAFKA_NAME"
const envNetworkName = "DC_NETWORK_NAME"
const envTransferName = "DC_TRANSFER_NAME"

const testPrefix = "tf-acc"

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"doublecloud": providerserver.NewProtocol6WithError(New("test")()),
	"dc":          providerserver.NewProtocol6WithError(New("test")()),
}

var testProjectId = os.Getenv(envProjectId)
var testNetworkId = os.Getenv(envNetworkId)
var testClickhouseName = os.Getenv(envClickhouseName)
var testKafkaName = os.Getenv(envKafkaName)
var testNetworkName = os.Getenv(envNetworkName)
var testDSTransferName = os.Getenv(envTransferName)

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	requiredEnvs := []string{envAuthkey, envProjectId, envNetworkId, envClickhouseName, envKafkaName, envNetworkName, envTransferName}
	for _, envName := range requiredEnvs {
		if os.Getenv(envName) == "" {
			t.Fatalf("%s must be set for acceptance tests", envName)
		}
	}
}

func configForSweepers() (*Config, error) {
	config := &Config{}

	if v := os.Getenv(envAuthkey); v != "" {
		key, err := iamkey.ReadFromJSONFile(v)
		if err != nil {
			return nil, err
		}
		credentials, err := dc.ServiceAccountKey(key)
		if err != nil {
			return nil, err
		}
		config.Credentials = &credentials
	} else {
		return nil, fmt.Errorf("%s must be set for sweep", envAuthkey)
	}
	config.ProjectId = os.Getenv(envProjectId)
	if config.ProjectId == "" {
		return nil, fmt.Errorf("%s must be set for sweep", envProjectId)
	}

	err := config.init(context.Background())
	return config, err
}
