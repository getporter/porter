package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestPorter_buildDockerfile(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

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

	err := p.generateDockerFile()
	require.NoError(t, err)

	exists, err := p.FileSystem.Exists("Dockerfile")
	require.NoError(t, err)
	require.True(t, exists, "Dockerfile wasn't written")

	f, _ := p.FileSystem.Stat("Dockerfile")
	if f.Size() == 0 {
		t.Fatalf("Dockerfile is empty")
	}
}
