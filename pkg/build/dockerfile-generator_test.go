package build

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/templates"
	"github.com/carolynvs/aferox"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_buildDockerfile(t *testing.T) {
	t.Parallel()

	drivers := []string{config.BuildDriverDocker, config.BuildDriverBuildkit}
	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {

			c := config.NewTestConfig(t)
			c.Data.BuildDriver = driver
			c.SetExperimentalFlags(experimental.FlagBuildDrivers)
			tmpl := templates.NewTemplates(c.Config)
			configTpl, err := tmpl.GetManifest()
			require.Nil(t, err)
			c.TestContext.AddTestFileContents(configTpl, config.Name)

			m, err := manifest.LoadManifestFrom(c.Context, config.Name)
			require.NoError(t, err, "could not load manifest")

			// ignore mixins in the unit tests
			m.Mixins = []manifest.MixinDeclaration{}

			mp := mixin.NewTestMixinProvider()
			g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
			gotlines, err := g.buildDockerfile()
			require.NoError(t, err)
			gotDockerfile := strings.Join(gotlines, "\n")

			wantDockerfilePath := fmt.Sprintf("testdata/%s.Dockerfile", driver)
			wantDockerfile, err := ioutil.ReadFile(wantDockerfilePath)
			require.NoError(t, err, "could not read %s", wantDockerfilePath)
			assert.Equal(t, string(wantDockerfile), gotDockerfile)
		})
	}
}

func TestPorter_buildDockerfile_alternateManifestLocation(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)

	manifestPath := "alternate/porter.yaml"
	c.TestContext.AddTestFileContents(configTpl, manifestPath)

	m, err := manifest.LoadManifestFrom(c.Context, manifestPath)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	gotlines, err := g.buildDockerfile()
	require.NoError(t, err)

	wantlines := []string{
		"FROM debian:stretch-slim",
		"",
		"ARG BUNDLE_DIR",
		"",
		"RUN apt-get update && apt-get install -y ca-certificates",
		"",
		"",
		"COPY . $BUNDLE_DIR",
		"RUN rm $BUNDLE_DIR/alternate/porter.yaml",
		"RUN rm -fr $BUNDLE_DIR/.cnab",
		"COPY .cnab /cnab",
		"WORKDIR $BUNDLE_DIR",
		"CMD [\"/cnab/app/run\"]",
	}
	assert.Equal(t, wantlines, gotlines)
}

func TestPorter_buildDockerfile_separateManifestLocation(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)

	// The standard test context uses an in-memory Filesystem with root dir of "/".
	// For this test, we update the context to be a subdirectory
	// so that the manifest path is suitably separate.
	c.FileSystem = aferox.NewAferox("/porter", afero.NewMemMapFs())

	manifestPath := "/tmp/separate/porter.yaml"
	c.TestContext.AddTestFileContents(configTpl, manifestPath)

	m, err := manifest.LoadManifestFrom(c.Context, manifestPath)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	gotlines, err := g.buildDockerfile()
	require.NoError(t, err)

	// Verify that the line to remove the manifest from BUNDLE_DIR is *not* added
	wantlines := []string{
		"FROM debian:stretch-slim",
		"",
		"ARG BUNDLE_DIR",
		"",
		"RUN apt-get update && apt-get install -y ca-certificates",
		"",
		"",
		"COPY . $BUNDLE_DIR",
		"RUN rm -fr $BUNDLE_DIR/.cnab",
		"COPY .cnab /cnab",
		"WORKDIR $BUNDLE_DIR",
		"CMD [\"/cnab/app/run\"]",
	}
	assert.Equal(t, wantlines, gotlines)
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
		mp := mixin.NewTestMixinProvider()
		g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
		gotlines, err := g.buildDockerfile()

		// We expect an error when ARG BUNDLE_DIR is not in Dockerfile
		require.EqualError(t, err, ErrorMessage)

		wantLines := []string(nil)
		assert.Equal(t, wantLines, gotlines)

	})

	t.Run("build from custom docker with ARG BUNDLE_DIR supplied", func(t *testing.T) {
		t.Parallel()

		c := config.NewTestConfig(t)
		tmpl := templates.NewTemplates(c.Config)
		configTpl, err := tmpl.GetManifest()
		require.Nil(t, err)
		c.TestContext.AddTestFileContents(configTpl, config.Name)

		m, err := manifest.LoadManifestFrom(c.Context, config.Name)
		require.NoError(t, err, "could not load manifest")

		// Use a custom dockerfile template
		m.Dockerfile = "Dockerfile.template"
		customFrom := `FROM ubuntu:latest
ARG BUNDLE_DIR
COPY mybin /cnab/app/

`
		c.TestContext.AddTestFileContents([]byte(customFrom), "Dockerfile.template")

		// ignore mixins in the unit tests
		m.Mixins = []manifest.MixinDeclaration{}
		mp := mixin.NewTestMixinProvider()
		g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
		gotlines, err := g.buildDockerfile()

		// We expect no error when ARG BUNDLE_DIR is in Dockerfile
		require.NoError(t, err)

		wantLines := []string{
			"FROM ubuntu:latest",
			"ARG BUNDLE_DIR",
			"COPY mybin /cnab/app/",
			"",
			"RUN rm $BUNDLE_DIR/porter.yaml",
			"RUN rm -fr $BUNDLE_DIR/.cnab",
			"COPY .cnab /cnab",
			"WORKDIR $BUNDLE_DIR",
			"CMD [\"/cnab/app/run\"]",
		}
		assert.Equal(t, wantLines, gotlines)
	})
}

func TestPorter_buildDockerfile_output(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	_, err = g.buildDockerfile()
	require.NoError(t, err)

	wantlines := `
Generating Dockerfile =======>
FROM debian:stretch-slim

ARG BUNDLE_DIR

RUN apt-get update && apt-get install -y ca-certificates


COPY . $BUNDLE_DIR
RUN rm $BUNDLE_DIR/porter.yaml
RUN rm -fr $BUNDLE_DIR/.cnab
COPY .cnab /cnab
WORKDIR $BUNDLE_DIR
CMD ["/cnab/app/run"]
`
	assert.Equal(t, wantlines, c.TestContext.GetOutput())
}

func TestPorter_generateDockerfile(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// ignore mixins in the unit tests
	m.Mixins = []manifest.MixinDeclaration{}

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	err = g.GenerateDockerFile()
	require.NoError(t, err)

	wantDockerfilePath := ".cnab/Dockerfile"
	dockerfileExists, err := c.FileSystem.Exists(wantDockerfilePath)
	require.NoError(t, err)
	require.True(t, dockerfileExists, "Dockerfile wasn't written")

	f, _ := c.FileSystem.Stat(wantDockerfilePath)
	if f.Size() == 0 {
		t.Fatalf("Dockerfile is empty")
	}

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

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
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

func TestPorter_replacePorterMixinTokenWithBuildInstructions(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Use a custom dockerfile template
	m.Dockerfile = "Dockerfile.template"
	customFrom := `FROM ubuntu:light
# PORTER_MIXINS
ARG BUNDLE_DIR
COPY mybin /cnab/app/
`
	c.TestContext.AddTestFileContents([]byte(customFrom), "Dockerfile.template")

	mp := mixin.NewTestMixinProvider()
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)

	gotlines, err := g.buildDockerfile()
	require.NoError(t, err)

	wantLines := []string{
		"FROM ubuntu:light",
		"# exec mixin has no buildtime dependencies",
		"",
		"ARG BUNDLE_DIR",
		"COPY mybin /cnab/app/",
		"RUN rm $BUNDLE_DIR/porter.yaml",
		"RUN rm -fr $BUNDLE_DIR/.cnab",
		"COPY .cnab /cnab",
		"WORKDIR $BUNDLE_DIR",
		"CMD [\"/cnab/app/run\"]",
	}
	// testcase did not specify build instructions
	// however, the # PORTER_MIXINS token should be removed
	assert.Equal(t, wantLines, gotlines)
}

func TestPorter_appendBuildInstructionsIfMixinTokenIsNotPresent(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
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

	gotlines, err := g.buildDockerfile()
	require.NoError(t, err)

	wantLines := []string{
		"FROM ubuntu:light",
		"ARG BUNDLE_DIR",
		"COPY mybin /cnab/app/",
		"# exec mixin has no buildtime dependencies",
		"",
		"RUN rm $BUNDLE_DIR/porter.yaml",
		"RUN rm -fr $BUNDLE_DIR/.cnab",
		"COPY .cnab /cnab",
		"WORKDIR $BUNDLE_DIR",

		"CMD [\"/cnab/app/run\"]",
	}
	// testcase did not specify build instructions
	// however, the # PORTER_MIXINS token should be removed
	assert.Equal(t, wantLines, gotlines)
}

func TestPorter_buildMixinsSection_mixinErr(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	tmpl := templates.NewTemplates(c.Config)
	configTpl, err := tmpl.GetManifest()
	require.Nil(t, err)
	c.TestContext.AddTestFileContents(configTpl, config.Name)

	m, err := manifest.LoadManifestFrom(c.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	m.Mixins = []manifest.MixinDeclaration{{Name: "exec"}}

	mp := mixin.NewTestMixinProvider()
	mp.ReturnBuildError = true
	g := NewDockerfileGenerator(c.Config, m, tmpl, mp)
	_, err = g.buildMixinsSection()
	require.EqualError(t, err, "1 error occurred:\n\t* error encountered from mixin \"exec\": encountered build error\n\n")
}
