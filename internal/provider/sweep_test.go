package provider

import (
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func debugLog(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}
