package porter

import (
	"testing"

	"github.com/deislabs/porter/pkg/config"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestPorter_buildDockerfile(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	p.TestConfig.TestContext.AddTestFile("../../templates/porter.yaml", config.Name)
	err := p.LoadManifest(config.Name)
	require.NoError(t, err)

	gotlines, err := p.buildDockerFile()
	require.NoError(t, err)

	wantlines := []string{
		"FROM ubuntu:latest",
		"COPY cnab/ /cnab/",
		"COPY porter.yaml /cnab/app/porter.yaml",
		"CMD [/cnab/app/run]",
	}
	assert.Equal(t, wantlines, gotlines)
}

func TestPorter_generateDockerfile(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	p.TestConfig.TestContext.AddTestFile("../../templates/porter.yaml", config.Name)
	err := p.LoadManifest(config.Name)
	require.NoError(t, err)

	err = p.generateDockerFile()
	require.NoError(t, err)

	dockerfileExists, err := p.FileSystem.Exists("Dockerfile")
	require.NoError(t, err)
	require.True(t, dockerfileExists, "Dockerfile wasn't written")

	f, _ := p.FileSystem.Stat("Dockerfile")
	if f.Size() == 0 {
		t.Fatalf("Dockerfile is empty")
	}
}

func TestPorter_copyMixins(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	p.TestConfig.TestContext.AddTestFile("../../templates/porter.yaml", config.Name)
	err := p.LoadManifest(config.Name)
	require.NoError(t, err)

	err = p.copyMixins()
	require.NoError(t, err)

	wantPorterMixin := "cnab/app/mixins/porter/porter-runtime"
	porterMixinExists, err := p.FileSystem.Exists(wantPorterMixin)
	require.NoError(t, err)
	assert.True(t, porterMixinExists, "The porter-runtime mixin wasn't copied into %s", wantPorterMixin)

	wantExecMixin := "cnab/app/mixins/exec/exec-runtime"
	execMixinExists, err := p.FileSystem.Exists(wantExecMixin)
	require.NoError(t, err)
	assert.True(t, execMixinExists, "The exec-runtime mixin wasn't copied into %s", wantExecMixin)
}
