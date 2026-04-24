//go:build integration
// +build integration

package wanderlog

import (
	"os"
	"testing"
)

const prodIntegrationOptInEnv = "WANDERLOG_RUN_PROD_INTEGRATION"

func requireProductionIntegrationOptIn(t *testing.T) {
	t.Helper()
	if os.Getenv(prodIntegrationOptInEnv) != "1" {
		t.Skipf("skipping production integration test; set %s=1 to run against wanderlog.com", prodIntegrationOptInEnv)
	}
}
