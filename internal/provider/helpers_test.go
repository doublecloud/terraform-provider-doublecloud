package provider

import (
	"regexp"

	"github.com/doublecloud/go-genproto/doublecloud/v1"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type testCaseErrorConfig struct {
	name   string
	config string
	err    *regexp.Regexp
}

func unitTestCase(endpoint string, tc testCaseErrorConfig) resource.TestCase {
	return resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: testFakeProtoV6ProviderFactories(endpoint),
		Steps: []resource.TestStep{
			{
				Config:      tc.config,
				ExpectError: tc.err,
			},
		},
	}
}

func networkOperationDone(resourceID string) *doublecloud.Operation {
	return &doublecloud.Operation{
		Id:         uuid.NewString(),
		ProjectId:  testProjectId,
		Status:     doublecloud.Operation_STATUS_DONE,
		ResourceId: resourceID,
	}
}
