package main

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/require"
)

func TestRun_Validate(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()

	configTpl, err := p.Templates.GetManifest()
	require.NoError(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)
	cmd := buildRunCommand(p.Porter)

	p.Setenv(config.EnvACTION, cnab.ActionInstall)

	err = cmd.PreRunE(cmd, []string{})
	require.Nil(t, err)
}

func TestRun_ValidateCustomAction(t *testing.T) {
	p := porter.NewTestPorter(t)
	defer p.Close()

	configTpl, err := p.Templates.GetManifest()
	require.NoError(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)
	cmd := buildRunCommand(p.Porter)

	p.Setenv(config.EnvACTION, "status")

	err = cmd.PreRunE(cmd, []string{})
	require.Nil(t, err)
}
