package main

import (
	"bytes"
	"os"
	"testing"

	"get.porter.sh/porter/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	pkg.Version = "v1.0.0"
	pkg.Commit = "abc123"

	t.Run("command", func(t *testing.T) {
		p := buildRootCommand()

		// Capture output
		var out bytes.Buffer
		p.SetOut(&out)

		// Set the command to run
		os.Args = []string{"porter", "version"}

		err := p.Execute()
		require.NoError(t, err)
		assert.Contains(t, out.String(), "porter v1.0.0 (abc123)")
	})

	t.Run("flag", func(t *testing.T) {
		p := buildRootCommand()

		// Capture output
		var out bytes.Buffer
		p.SetOut(&out)

		// Set the command to run
		os.Args = []string{"porter", "--version"}

		err := p.Execute()
		require.NoError(t, err)
		assert.Contains(t, out.String(), "porter v1.0.0 (abc123)")
	})
}
