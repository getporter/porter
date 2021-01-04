// +build integration

package client

import (
	"bytes"
	"io"
	"path/filepath"
	"runtime"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Run(t *testing.T) {
	// Provide a way for tests to capture stdout
	output := &bytes.Buffer{}

	cxt := context.NewTestContext(t)
	cxt.UseFilesystem()
	binDir := cxt.FindBinDir()

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
	if runtime.GOOS == "windows" {
		cmd.File = "testdata/exec_input.windows.yaml"
	}

	err = r.Run(cmd)
	require.NoError(t, err, "Run failed: %s", output.String())
	assert.Contains(t, output.String(), "Hello")
}

func TestRunner_RunWithMaskedOutput(t *testing.T) {
	// Provide a way for tests to capture stdout
	output := &bytes.Buffer{}

	// Copy output to the test log simultaneously, use go test -v to see the output
	aggOutput := io.MultiWriter(output, test.Logger{T: t})

	// Supply an CensoredWriter with values needing masking
	censoredWriter := context.NewCensoredWriter(aggOutput)
	sensitiveValues := []string{"World"}
	// Add some whitespace values as well, to be sure writer does not replace
	sensitiveValues = append(sensitiveValues, " ", "", "\n", "\r", "\t")
	censoredWriter.SetSensitiveValues(sensitiveValues)

	cxt := context.NewTestContext(t)
	cxt.UseFilesystem()
	binDir := cxt.FindBinDir()

	// I'm not using the TestRunner because I want to use the current filesystem, not an isolated one
	r := NewRunner("exec", filepath.Join(binDir, "mixins/exec"), false)

	// Capture the output
	r.Out = censoredWriter
	r.Err = censoredWriter

	err := r.Validate()
	require.NoError(t, err)

	cmd := pkgmgmt.CommandOptions{
		Command: "install",
		File:    "testdata/exec_input.yaml",
	}
	if runtime.GOOS == "windows" {
		cmd.File = "testdata/exec_input.windows.yaml"
	}

	err = r.Run(cmd)
	require.NoError(t, err, "Run failed: %s", output.String())
	assert.Contains(t, output.String(), "Hello *******")
}
