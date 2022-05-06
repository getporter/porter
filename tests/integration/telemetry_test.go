//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Validate that we can configure a live connection to a telemetry endpoint
func TestTelemetrySetup(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err, "test setup failed")

	ctx := context.Background()
	_, _, err = test.RunPorter("install", "otel-jaeger", "-r=ghcr.io/getporter/examples/otel-jaeger:v0.1.0", "--allow-docker-host-access")
	require.NoError(t, err)
	defer test.RunPorter("uninstall", "otel-jaeger", "--allow-docker-host-access")

	// Wait until the collection should be up
	time.Sleep(10 * time.Second)

	// Try to run porter with telemetry enabled
	p := porter.New()
	defer p.Close()
	os.Setenv("PORTER_EXPERIMENTAL", "structured-logs")
	os.Setenv("PORTER_TELEMETRY_ENABLED", "true")
	os.Setenv("PORTER_TELEMETRY_PROTOCOL", "grpc")
	os.Setenv("PORTER_TELEMETRY_INSECURE", "true")
	defer func() {
		os.Unsetenv("PORTER_EXPERIMENTAL")
		os.Unsetenv("PORTER_TELEMETRY_ENABLED")
		os.Unsetenv("PORTER_TELEMETRY_PROTOCOL")
		os.Unsetenv("PORTER_TELEMETRY_INSECURE")
	}()
	err = p.Load(ctx, nil)
	require.NoError(t, err, "error loading porter configuration")

	ctx, log := p.StartRootSpan(ctx, t.Name())
	defer log.Close()
	assert.True(t, log.IsTracingEnabled())
}
