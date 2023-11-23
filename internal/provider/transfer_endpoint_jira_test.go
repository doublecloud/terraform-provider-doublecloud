package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointJiraSource(t *testing.T) {
	t.Parallel()

	testEndpointName := fmt.Sprintf("%s-jira-source", testPrefix)
	testEndpointID := fmt.Sprintf("doublecloud_transfer_endpoint.%s", testEndpointName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							jira_source {
								api_token = "test"
								domain = "test-domain@jira.com"
								email = "test@example.com"
								projects = ["Test-1", "Test-2"]
								start_date = "2017-01-25T23:59:59Z"
								issues_stream_expand_with = ["transitions", "rendered_fields"]
								enable_experimental_streams	= false
							}
						}
					}`,
					testEndpointName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointID, "name", testEndpointName),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.domain", "test-domain@jira.com"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.email", "test@example.com"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.start_date", "2017-01-25T23:59:59Z"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.projects.0", "Test-1"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.projects.1", "Test-2"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.issues_stream_expand_with.0", "transitions"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.issues_stream_expand_with.1", "rendered_fields"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.enable_experimental_streams", "false"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							jira_source {
								api_token = "test"
								domain = "test-domain-new@jira.com"
								email = "test-2@example.com"
								projects = ["Test-3"]
								start_date = "2017-01-25T23:59:59Z"
								issues_stream_expand_with = ["rendered_fields"]
								enable_experimental_streams	= false
							}
						}
					}`,
					testEndpointName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointID, "name", testEndpointName),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.domain", "test-domain-new@jira.com"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.email", "test-2@example.com"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.projects.0", "Test-3"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.jira_source.issues_stream_expand_with.0", "rendered_fields"),
				),
			},
			// Delete occurs automatically in TestCase
		},
	})
}
