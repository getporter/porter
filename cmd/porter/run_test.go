package main

import (
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Validate(t *testing.T) {
	defer os.Unsetenv(config.EnvACTION)

	p := porter.NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.TestConfig.TestContext.AddFile("../../templates/porter.yaml", config.Name)
	cmd := buildRunCommand(p.Porter)

	os.Setenv(config.EnvACTION, string(config.ActionInstall))

	err := cmd.PreRunE(cmd, []string{})
	require.Nil(t, err)
}

func TestRun_Validate_MissingFile(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	cmd := buildRunCommand(p.Porter)

	cmd.Flags().Set("action", string(config.ActionInstall))

	err := cmd.PreRunE(cmd, []string{})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid --file")
}

func TestRun_Validate_MissingAction(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	cmd := buildRunCommand(p.Porter)

	err := cmd.PreRunE(cmd, []string{})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

func TestRun_Validate_InvalidAction(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	cmd := buildRunCommand(p.Porter)

	cmd.Flags().Set("action", "phony")

	err := cmd.PreRunE(cmd, []string{})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), `invalid action "phony`)
}
