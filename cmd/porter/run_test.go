package main

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestRun_Validate(t *testing.T) {
	defer os.Unsetenv(config.EnvACTION)

	p := porter.NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	configTpl, err := p.Templates.GetManifest()
	require.NoError(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)
	cmd := buildRunCommand(p.Porter)

	os.Setenv(config.EnvACTION, string(claim.ActionInstall))

	err = cmd.PreRunE(cmd, []string{})
	require.Nil(t, err)
}

func TestRun_ValidateCustomAction(t *testing.T) {
	defer os.Unsetenv(config.EnvACTION)

	p := porter.NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	configTpl, err := p.Templates.GetManifest()
	require.NoError(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)
	cmd := buildRunCommand(p.Porter)

	os.Setenv(config.EnvACTION, "status")

	err = cmd.PreRunE(cmd, []string{})
	require.Nil(t, err)
}
