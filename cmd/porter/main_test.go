package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandWiring(t *testing.T) {
	testcases := []string{
		"build",
		"create",
		"install",
		"run",
		"schema",
		"bundle install",
		"list mixins",
		"version",
	}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			osargs := strings.Split(tc, " ")

			rootCmd := buildRootCommand()
			_, _, err := rootCmd.Find(osargs)
			assert.NoError(t, err)
		})
	}
}
