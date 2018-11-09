package mixin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/porter/pkg/context"
)

type TestRunner struct {
	*Runner
	TestContext *context.TestContext
}

// NewTestRunner initializes a mixin test runner, with the output buffered, and an in-memory file system.
func NewTestRunner(t *testing.T, mixin string) *TestRunner {
	c := context.NewTestContext(t)
	mixinDir := "/root/.porter/mixins/exec"
	r := &TestRunner{
		Runner:      NewRunner(mixin, mixinDir),
		TestContext: c,
	}
	r.Context = c.Context

	// Setup Mixin Home
	err := c.FileSystem.MkdirAll(mixinDir, os.ModePerm)
	require.NoError(t, err)
	c.AddFile("../../bin/mixins/exec/exec", filepath.Join(mixinDir, "exec"))

	return r
}
