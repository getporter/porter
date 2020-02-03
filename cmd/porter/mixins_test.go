package main

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildListMixinsCommand_DefaultFormat(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := buildMixinsListCommand(p.Porter)

	err := cmd.PreRunE(cmd, []string{})

	require.Nil(t, err)
	assert.Equal(t, "table", cmd.Flag("output").Value.String())
}

func TestBuildListMixinsCommand_AlternateFormat(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := buildMixinsListCommand(p.Porter)
	cmd.ParseFlags([]string{"-o", "json"})

	err := cmd.PreRunE(cmd, []string{})

	require.Nil(t, err)
	assert.Equal(t, "json", cmd.Flag("output").Value.String())
}

func TestBuildListMixinsCommand_BadFormat(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := buildMixinsListCommand(p.Porter)
	cmd.ParseFlags([]string{"-o", "flarts"})

	err := cmd.PreRunE(cmd, []string{})

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid format: flarts")
}

func TestBuildMixinInstallCommand(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := BuildMixinInstallCommand(p.Porter)
	err := cmd.ParseFlags([]string{"--url", "https://example.com/mixins/helm"})
	require.NoError(t, err)

	err = cmd.PreRunE(cmd, []string{"helm"})
	require.NoError(t, err)
}

func TestBuildMixinInstallCommand_NoMixinName(t *testing.T) {
	p := porter.NewTestPorter(t)
	cmd := BuildMixinInstallCommand(p.Porter)

	err := cmd.PreRunE(cmd, []string{})
	require.EqualError(t, err, "no name was specified")
}
