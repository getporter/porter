//go:build integration

package client

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Run(t *testing.T) {
	ctx := context.Background()
	// Provide a way for tests to capture stdout
	output := &bytes.Buffer{}

	c := portercontext.NewTestContext(t)
	c.UseFilesystem()
	binDir := c.FindBinDir()

	// I'm not using the TestRunner because I want to use the current filesystem, not an isolated one
	r := NewRunner("exec", filepath.Join(binDir, "mixins/exec"), false)

	// Capture the output
	r.Out = output
	r.Err = output

	err := r.Validate()
	require.NoError(t, err)

	cmd := pkgmgmt.CommandOptions{
		Command: "install",
		File:    "testdata/exec_input.yaml",
	}
	err = r.Run(ctx, cmd)
	assert.NoError(t, err)
	assert.Contains(t, string(output.Bytes()), "Hello World")
}

func TestRunner_RunWithMaskedOutput(t *testing.T) {
	ctx := context.Background()

	// Provide a way for tests to capture stdout
	output := &bytes.Buffer{}

	// Copy output to the test log simultaneously, use go test -v to see the output
	aggOutput := io.MultiWriter(output, test.Logger{T: t})

	// Supply an CensoredWriter with values needing masking
	censoredWriter := portercontext.NewCensoredWriter(aggOutput)
	sensitiveValues := []string{"World"}
	// Add some whitespace values as well, to be sure writer does not replace
	sensitiveValues = append(sensitiveValues, " ", "", "\n", "\r", "\t")
	censoredWriter.SetSensitiveValues(sensitiveValues)

	c := portercontext.NewTestContext(t)
	c.UseFilesystem()
	binDir := c.FindBinDir()

	// I'm not using the TestRunner because I want to use the current filesystem, not an isolated one
	r := NewRunner("exec", filepath.Join(binDir, "mixins/exec"), false)

	// Capture the output
	r.Out = censoredWriter
	r.Err = censoredWriter

	err := r.Validate()
	require.NoError(t, err)

	cmd := pkgmgmt.CommandOptions{
		Command: "install",
		File:    "testdata/exec_input_with_whitespace.yaml",
	}

	err = r.Run(ctx, cmd)
	assert.NoError(t, err)
	tests.RequireOutputContains(t, output.String(), "Hello *******")
}
