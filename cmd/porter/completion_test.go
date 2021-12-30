package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {

	t.Run("completion", func(t *testing.T) {
		p := buildRootCommand()

		// Capture the output of the command.
		var out bytes.Buffer
		p.SetOut(&out)

		// Run the initial completion command with a bash example.
		os.Args = []string{"porter", "completion", "bash"}

		err := p.Execute()
		require.NoError(t, err)
	})
}
