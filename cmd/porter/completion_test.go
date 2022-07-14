package main

import (
	"bytes"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {
	p := buildRootCommand()

	// Capture the output of the command.
	var out bytes.Buffer
	p.SetOut(&out)

	// Run the initial completion command with a bash example.
	os.Args = []string{"porter", "completion", "bash"}

	err := p.Execute()
	require.NoError(t, err)
	// Test the output of the command contains a specific string for bash.
	assert.Contains(t, out.String(), "bash completion for porter")
}

func TestCompletion_SkipConfig(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := buildCompletionCommand(p.Porter)
	shouldSkip := cli.ShouldSkipConfig(cmd)
	require.True(t, shouldSkip, "expected that we skip loading configuration for the completion command")
}
