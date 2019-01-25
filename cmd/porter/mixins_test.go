package main

import (
	"testing"

	"github.com/deislabs/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestBuildListMixinsCommand_DefaultFormat(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := buildMixinListCommand(p.Porter)
	cmd.ParseFlags([]string{"-o", "json"})

	err := cmd.PreRunE(cmd, []string{})

	require.Nil(t, err)
}

func TestBuildListMixinsCommand_AlternateFormat(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := buildMixinListCommand(p.Porter)

	err := cmd.PreRunE(cmd, []string{})

	require.Nil(t, err)
}

func TestBuildListMixinsCommand_BadFormat(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := buildMixinListCommand(p.Porter)
	cmd.ParseFlags([]string{"-o", "flarts"})

	err := cmd.PreRunE(cmd, []string{})

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid format: flarts")
}
