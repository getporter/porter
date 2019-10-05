package build

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/manifest"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_buildDockerfile(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates()
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := &mixin.TestMixinProvider{}
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	gotlines, err := g.buildDockerfile()
	require.NoError(t, err)

	wantlines := []string{
		"FROM debian:stretch",
		"",
		"ARG BUNDLE_DIR",
		"",
		"RUN apt-get update && apt-get install -y ca-certificates",
		"",
		"COPY .cnab /cnab",
		"COPY . $BUNDLE_DIR",
		"RUN rm -fr $BUNDLE_DIR/.cnab",
		"WORKDIR $BUNDLE_DIR",
		"CMD [\"/cnab/app/run\"]",
	}
	assert.Equal(t, wantlines, gotlines)
}

func TestPorter_buildCustomDockerfile(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates()
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Use a custom dockerfile template
	m.Dockerfile = "Dockerfile.template"
	customFrom := `FROM ubuntu:latest
COPY mybin /cnab/app/

`
	c.TestContext.AddTestFileContents([]byte(customFrom), "Dockerfile.template")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := &mixin.TestMixinProvider{}
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	gotlines, err := g.buildDockerfile()
	require.NoError(t, err)

	wantLines := []string{
		"FROM ubuntu:latest",
		"COPY mybin /cnab/app/",
		"",
		"COPY .cnab/ /cnab/",
		"COPY porter.yaml $BUNDLE_DIR/porter.yaml",
		"WORKDIR $BUNDLE_DIR",
		"CMD [\"/cnab/app/run\"]",
	}
	assert.Equal(t, wantLines, gotlines)
}

func TestPorter_buildDockerfile_output(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates()
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := &mixin.TestMixinProvider{}
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	_, err = g.buildDockerfile()
	require.NoError(t, err)

	wantlines := `
Generating Dockerfile =======>
FROM debian:stretch

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates

COPY .cnab /cnab
COPY . $BUNDLE_DIR
RUN rm -fr $BUNDLE_DIR/.cnab
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]
`
	assert.Equal(t, wantlines, c.TestContext.GetOutput())
}

func TestPorter_generateDockerfile(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates()
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := &mixin.TestMixinProvider{}
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	err = g.GenerateDockerFile()
	require.NoError(t, err)

	dockerfileExists, err := c.FileSystem.Exists("Dockerfile")
	require.NoError(t, err)
	require.True(t, dockerfileExists, "Dockerfile wasn't written")

	f, _ := c.FileSystem.Stat("Dockerfile")
	if f.Size() == 0 {
		t.Fatalf("Dockerfile is empty")
	}
}

func TestPorter_prepareDockerFilesystem(t *testing.T) {
	c := config.NewTestConfig(t)
	c.SetupPorterHome()
	tmpl := templates.NewTemplates()
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	mp := &mixin.TestMixinProvider{}
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	err = g.PrepareFilesystem()
	require.NoError(t, err)

	wantRunscript := LOCAL_RUN
	runscriptExists, err := c.FileSystem.Exists(wantRunscript)
	require.NoError(t, err)
	assert.True(t, runscriptExists, "The run script wasn't copied into %s", wantRunscript)

	wantPorterRuntime := filepath.Join(LOCAL_APP, "porter-runtime")
	porterMixinExists, err := c.FileSystem.Exists(wantPorterRuntime)
	require.NoError(t, err)
	assert.True(t, porterMixinExists, "The porter-runtime wasn't copied into %s", wantPorterRuntime)

	wantExecMixin := filepath.Join(LOCAL_APP, "mixins", "exec", "exec-runtime")
	execMixinExists, err := c.FileSystem.Exists(wantExecMixin)
	require.NoError(t, err)
	assert.True(t, execMixinExists, "The exec-runtime mixin wasn't copied into %s", wantExecMixin)
}

func TestDockerFileGenerator_getMixinBuildInput(t *testing.T) {
	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates()
	c.TestContext.AddTestFile("testdata/multiple-mixins.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	mp := &mixin.TestMixinProvider{}
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)

	input := g.getMixinBuildInput("exec")

	assert.Nil(t, input.Config, "exec mixin should have no config")
	assert.Len(t, input.Actions, 4, "expected 4 actions")

	require.Contains(t, input.Actions, "install")
	assert.Len(t, input.Actions["install"], 2, "expected 2 exec install steps")

	require.Contains(t, input.Actions, "upgrade")
	assert.Len(t, input.Actions["upgrade"], 1, "expected 1 exec upgrade steps")

	require.Contains(t, input.Actions, "uninstall")
	assert.Len(t, input.Actions["uninstall"], 1, "expected 1 exec uninstall steps")

	require.Contains(t, input.Actions, "status")
	assert.Len(t, input.Actions["status"], 1, "expected 1 exec status steps")

	input = g.getMixinBuildInput("az")
	assert.Equal(t, map[interface{}]interface{}{"extensions": []interface{}{"iot"}}, input.Config, "az mixin should have config")
}
