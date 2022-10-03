package build

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/templates"
	"get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_buildDockerfile(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.Data.BuildDriver = config.BuildDriverBuildkit
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	gotlines, err := g.buildDockerfile(context.Background())
	require.NoError(t, err)
	gotDockerfile := strings.Join(gotlines, "\n")

	wantDockerfilePath := "testdata/buildkit.Dockerfile"
	test.CompareGoldenFile(t, wantDockerfilePath, gotDockerfile)
}

func TestPorter_buildCustomDockerfile(t *testing.T) {
	t.Parallel()

	t.Run("build from custom docker without supplying ARG BUNDLE_DIR", func(t *testing.T) {
		t.Parallel()

		c := config.NewTestConfig(t)
		tmpl := templates.NewTemplates(c.Config)
		configTpl, err := tmpl.GetManifest()
		require.Nil(t, err)
		c.TestContext.AddTestFileContents(configTpl, config.Name)

		m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
		require.NoError(t, err, "could not load manifest")

		// Use a custom dockerfile template
		m.Dockerfile = "Dockerfile.template"
		customFrom := `FROM ubuntu:latest
COPY mybin /cnab/app/

`
		c.TestContext.AddTestFileContents([]byte(customFrom), "Dockerfile.template")

		// ignore mixins in the unit tests
		m.Mixins = []manifest.MixinDeclaration{}
		mp := mixin.NewTestMixinProvider()
		g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
		gotlines, err := g.buildDockerfile(context.Background())

		// We should inject initialization lines even when they didn't include the token
		require.NoError(t, err)
		test.CompareGoldenFile(t, "testdata/missing-args-expected-output.Dockerfile", strings.Join(gotlines, "\n"))
	})

	t.Run("build from custom docker with PORTER_INIT supplied", func(t *testing.T) {
		t.Parallel()

		c := config.NewTestConfig(t)
		tmpl := templates.NewTemplates(c.Config)
		configTpl, err := tmpl.GetManifest()
		require.Nil(t, err)
		c.TestContext.AddTestFileContents(configTpl, config.Name)

		m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
		require.NoError(t, err, "could not load manifest")

		// Use a custom dockerfile template
		m.Dockerfile = "Dockerfile.template"
		customFrom := `FROM ubuntu:latest
# stuff
# PORTER_INIT
COPY mybin /cnab/app/

`
		c.TestContext.AddTestFileContents([]byte(customFrom), "Dockerfile.template")

		// ignore mixins in the unit tests
		m.Mixins = []manifest.MixinDeclaration{}
		mp := mixin.NewTestMixinProvider()
		g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
		gotlines, err := g.buildDockerfile(context.Background())

		require.NoError(t, err)
		test.CompareGoldenFile(t, "testdata/custom-dockerfile-expected-output.Dockerfile", strings.Join(gotlines, "\n"))
	})
}

func TestPorter_generateDockerfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := config.NewTestConfig(t)
	defer c.Close()

	// Start a span so we can capture the output
	ctx, log := c.StartRootSpan(ctx, t.Name())
	defer log.Close()

	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	err = g.GenerateDockerFile(ctx)
	require.NoError(t, err)

	wantDockerfilePath := ".cnab/Dockerfile"
	gotDockerfile, err := c.FileSystem.ReadFile(wantDockerfilePath)
	require.NoError(t, err)

	// Verify that we logged the dockerfile contents
	tests.RequireOutputContains(t, c.TestContext.GetError(), string(gotDockerfile), "expected the dockerfile to be printed to the logs")
	test.CompareGoldenFile(t, "testdata/buildkit.Dockerfile", string(gotDockerfile))

	// Verify that we didn't generate a Dockerfile at the root of the bundle dir
	oldDockerfilePathExists, _ := c.FileSystem.Exists("Dockerfile")
	assert.False(t, oldDockerfilePathExists, "expected the Dockerfile to be placed only at .cnab/Dockerfile, not at the root of the bundle directory")
}

func TestPorter_prepareDockerFilesystem(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	err = g.PrepareFilesystem()
	require.NoError(t, err)

	wantRunscript := LOCAL_RUN
	runscriptExists, err := c.FileSystem.Exists(wantRunscript)
	require.NoError(t, err)
	assert.True(t, runscriptExists, "The run script wasn't copied into %s", wantRunscript)

	wantPorterRuntime := filepath.Join(LOCAL_APP, "runtimes", "porter-runtime")
	porterMixinExists, err := c.FileSystem.Exists(wantPorterRuntime)
	require.NoError(t, err)
	assert.True(t, porterMixinExists, "The porter-runtime wasn't copied into %s", wantPorterRuntime)

	wantExecMixin := filepath.Join(LOCAL_APP, "mixins", "exec", "runtimes", "exec-runtime")
	execMixinExists, err := c.FileSystem.Exists(wantExecMixin)
	require.NoError(t, err)
	assert.True(t, execMixinExists, "The exec-runtime mixin wasn't copied into %s", wantExecMixin)
}

func TestPorter_appendBuildInstructionsIfMixinTokenIsNotPresent(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Use a custom dockerfile template
	m.Dockerfile = "Dockerfile.template"
	customFrom := `FROM ubuntu:light
ARG BUNDLE_DIR
COPY mybin /cnab/app/
`
	c.TestContext.AddTestFileContents([]byte(customFrom), "Dockerfile.template")

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)

	gotlines, err := g.buildDockerfile(context.Background())
	require.NoError(t, err)

	test.CompareGoldenFile(t, "testdata/missing-mixins-token-expected-output.Dockerfile", strings.Join(gotlines, "\n"))
}

func TestPorter_buildMixinsSection_mixinErr(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	m.Mixins = []manifest.MixinDeclaration{{Name: "exec"}}

	mp := mixin.NewTestMixinProvider()
	mp.ReturnBuildError = true
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	_, err = g.buildMixinsSection(ctx)
	require.EqualError(t, err, "1 error occurred:\n\t* error encountered from mixin \"exec\": encountered build error\n\n")
}
