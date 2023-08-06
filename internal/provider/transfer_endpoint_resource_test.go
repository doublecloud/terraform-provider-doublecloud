package provider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/doublecloud/go-genproto/doublecloud/transfer/v1"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("transfer_endpoint", &resource.Sweeper{
		Name:         "transfer_endpoint",
		F:            testSweepTransferEndpoints,
		Dependencies: []string{"transfer"},
	})
}

func testSweepTransferEndpoints(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return err
	}

	var errs error
	rq := &transfer.ListEndpointsRequest{ProjectId: conf.ProjectId}
	svc := conf.sdk.Transfer().Endpoint()
	it := svc.EndpointIterator(conf.ctx, rq)

	for it.Next() {
		v := it.Value()
		if strings.HasPrefix(v.Name, testPrefix) {
			err := sweepTransferEndpoint(conf, v)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to sweep endpoint %v: %v", v.Id, err))
			}
		}
	}
	return errs
}

func sweepTransferEndpoint(conf *Config, t *transfer.Endpoint) error {
	op, err := conf.sdk.WrapOperation(conf.sdk.Transfer().Endpoint().Delete(conf.ctx, &transfer.DeleteEndpointRequest{EndpointId: t.Id}))
	if err != nil {
		return err
	}
	return op.Wait(conf.ctx)
}
