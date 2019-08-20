// +build integration

package mixin

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner_Run(t *testing.T) {
	// Provide a way for tests to capture stdout
	output := &bytes.Buffer{}

	binDir := context.NewTestContext(t).FindBinDir()

	// I'm not using the TestRunner because I want to use the current filesystem, not an isolated one
	r := NewRunner("exec", filepath.Join(binDir, "mixins/exec"), false)
	r.Command = "install"
	r.File = "testdata/exec_input.yaml"

	// Capture the output
	r.Out = output
	r.Err = output

	err := r.Validate()
	require.NoError(t, err)

	err = r.Run()
	assert.NoError(t, err)
	assert.Contains(t, string(output.Bytes()), "Hello World")
}

func TestRunner_RunWithMaskedOutput(t *testing.T) {
	// Provide a way for tests to capture stdout
	output := &bytes.Buffer{}

	// Copy output to the test log simultaneously, use go test -v to see the output
	aggOutput := io.MultiWriter(output, test.Logger{T: t})

	// Supply an CensoredWriter with values needing masking
	censoredWriter := context.NewCensoredWriter(aggOutput)
	censoredWriter.SetSensitiveValues([]string{"World"})

	binDir := context.NewTestContext(t).FindBinDir()

	// I'm not using the TestRunner because I want to use the current filesystem, not an isolated one
	r := NewRunner("exec", filepath.Join(binDir, "mixins/exec"), false)
	r.Command = "install"
	r.File = "testdata/exec_input.yaml"

	// Capture the output
	r.Out = censoredWriter
	r.Err = censoredWriter

	err := r.Validate()
	require.NoError(t, err)

	err = r.Run()
	assert.NoError(t, err)
	assert.Contains(t, string(output.Bytes()), "Hello *******")
}
