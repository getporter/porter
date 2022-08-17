package tester

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryAlias is the environment variable that contains a pre-configured
// hostname alias that can be used to access localhost. This environment variable
// is only set in on the linux and macos CI agents so that we can test a variant
// of communicating with a registry that is unsecured but is not obviously "localhost" or 127.0.0.1.
const TestRegistryAlias = "PORTER_TEST_REGISTRY_ALIAS"

// TestRegistryOptions controls how a test registry is run.
type TestRegistryOptions struct {
	// UseTLS indicates the registry should use http, secured with a self-signed certificate.
	UseTLS bool

	// UseAlias indicates that when the TestRegistryAlias environment variable is set,
	// the registry address use the hostname alias, and not localhost.
	UseAlias bool
}

// TestRegistry is a temporary registry that is stopped when the test completes.
type TestRegistry struct {
	t           *testing.T
	containerID string

	// port that the registry is running on
	port string

	// closed tracks if Close has been called so that we only try to close once.
	closed bool

	// hostname is the address or name used to reference the registry
	hostname string
}

// String prints the registry URI.
func (t *TestRegistry) String() string {
	return fmt.Sprintf("%s:%s", t.hostname, t.port)
}

// Close stops the registry.
func (t *TestRegistry) Close() {
	// Only call close once
	if t.closed {
		return
	}

	err := shx.RunE("docker", "rm", "-vf", t.containerID)
	t.closed = true
	assert.NoError(t.t, err)
}

// StartTestRegistry runs an OCI registry in a container,
// returning details about the registry.
// The registry is cleaned up by default when the test completes.
func (t Tester) StartTestRegistry(opts TestRegistryOptions) *TestRegistry {
	cmd := shx.Command("docker", "run", "-d", "-P", "--restart=always")

	if opts.UseTLS {
		certDir := filepath.Join(t.RepoRoot, "tests/integration/testdata/certs")
		cmd.Args(
			"-v", fmt.Sprintf("%s:/certs", certDir),
			"-e", "REGISTRY_HTTP_TLS_CERTIFICATE=/certs/registry_auth.crt",
			"-e", "REGISTRY_HTTP_TLS_KEY=/certs/registry_auth.key",
		)
	}

	// The docker image name must go last
	cmd.Args("registry:2")

	var err error
	reg := &TestRegistry{t: t.T}
	reg.containerID, err = cmd.OutputE()
	require.NoError(t.T, err, "Could not start a temporary registry")

	// Automatically close the registry when the test is done
	t.T.Cleanup(reg.Close)

	// Get the port that it is running on
	reg.port, err = shx.OutputE("docker", "inspect", reg.containerID, "--format", `{{ (index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort }}`)
	require.NoError(t.T, err, "Could not get the published port of the temporary registry")

	// Determine if we have a hostname alias set up for the registry
	var hostname string
	if opts.UseAlias {
		hostname = os.Getenv(TestRegistryAlias)
	}
	if hostname == "" {
		hostname = "localhost"
	}
	reg.hostname = hostname

	return reg
}
