package tester

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryOptions controls how a test registry is run.
type TestRegistryOptions struct {
	Port   *int
	UseTLS bool
}

// TestRegistry is a temporary registry that is stopped when the test completes.
type TestRegistry struct {
	T           *testing.T
	ContainerID string
	Port        string

	// closed tracks if Close has been called so that we only try to close once.
	closed bool
}

// String prints the registry URI.
func (t *TestRegistry) String() string {
	return fmt.Sprintf("localhost:%s", t.Port)
}

// Close stops the registry.
func (t *TestRegistry) Close() {
	// Only call close once
	if t.closed {
		return
	}

	err := shx.RunE("docker", "rm", "-vf", t.ContainerID)
	t.closed = true
	assert.NoError(t.T, err)
}

// StartTestRegistry runs an OCI registry in a container,
// returning details about the registry.
// The registry is cleaned up by default when the test completes.
func (t Tester) StartTestRegistry(opts TestRegistryOptions) *TestRegistry {
	cmd := shx.Command("docker", "run", "-d", "--restart=always")

	if opts.Port == nil {
		cmd.Args("-P")
	}

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
	reg := &TestRegistry{T: t.T}
	reg.ContainerID, err = cmd.OutputE()
	require.NoError(t.T, err, "Could not start a temporary registry")

	// Automatically close the registry when the test is done
	t.T.Cleanup(reg.Close)

	if opts.Port == nil {
		// Get the port that it is running on
		reg.Port, err = shx.OutputE("docker", "inspect", reg.ContainerID, "--format", `{{ (index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort }}`)
		require.NoError(t.T, err, "Could not get the published port of the temporary registry")

	}

	return reg
}
