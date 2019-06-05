package porter

import (
	"encoding/json"
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_buildDockerfile(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	err = p.LoadManifest()
	require.NoError(t, err)

	// ignore mixins in the unit tests
	p.Manifest.Mixins = []string{}

	gotlines, err := p.buildDockerfile()
	require.NoError(t, err)

	wantlines := []string{
		"FROM quay.io/deis/lightweight-docker-go:v0.2.0",
		"FROM debian:stretch",
		"COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt",
		"",
		"COPY . /cnab/app",
		"RUN mv /cnab/app/cnab/app/* /cnab/app && rm -r /cnab/app/cnab",
		"WORKDIR /cnab/app",
		`CMD ["/cnab/app/run"]`,
	}
	assert.Equal(t, wantlines, gotlines)
}

func TestPorter_buildCustomDockerfile(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	err = p.LoadManifest()
	require.NoError(t, err)

	// Use a custom dockerfile template
	p.Manifest.Dockerfile = "Dockerfile.template"
	customFrom := `FROM ubuntu:latest
COPY mybin /cnab/app/

`
	p.TestConfig.TestContext.AddTestFileContents([]byte(customFrom), "Dockerfile.template")

	// ignore mixins in the unit tests
	p.Manifest.Mixins = []string{}

	gotlines, err := p.buildDockerfile()
	require.NoError(t, err)

	wantLines := []string{
		"FROM ubuntu:latest",
		"COPY mybin /cnab/app/",
		"",
		"COPY cnab/ /cnab/",
		"COPY porter.yaml /cnab/app/porter.yaml",
		"WORKDIR /cnab/app",
		"CMD [\"/cnab/app/run\"]",
	}
	assert.Equal(t, wantLines, gotlines)
}

func TestPorter_buildDockerfile_output(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	err = p.LoadManifest()
	require.NoError(t, err)

	// ignore mixins in the unit tests
	p.Manifest.Mixins = []string{}

	_, err = p.buildDockerfile()
	require.NoError(t, err)

	wantlines := `
Generating Dockerfile =======>
FROM quay.io/deis/lightweight-docker-go:v0.2.0
FROM debian:stretch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY . /cnab/app
RUN mv /cnab/app/cnab/app/* /cnab/app && rm -r /cnab/app/cnab
WORKDIR /cnab/app
CMD ["/cnab/app/run"]
`
	assert.Equal(t, wantlines, p.TestConfig.TestContext.GetOutput())
}

func TestPorter_generateDockerfile(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	err = p.LoadManifest()
	require.NoError(t, err)

	// ignore mixins in the unit tests
	p.Manifest.Mixins = []string{}

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

func TestPorter_prepareDockerFilesystem(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	err = p.LoadManifest()
	require.NoError(t, err)

	err = p.prepareDockerFilesystem()
	require.NoError(t, err)

	wantRunscript := "cnab/app/run"
	runscriptExists, err := p.FileSystem.Exists(wantRunscript)
	require.NoError(t, err)
	assert.True(t, runscriptExists, "The run script wasn't copied into %s", wantRunscript)

	wantPorterRuntime := "cnab/app/porter-runtime"
	porterMixinExists, err := p.FileSystem.Exists(wantPorterRuntime)
	require.NoError(t, err)
	assert.True(t, porterMixinExists, "The porter-runtime wasn't copied into %s", wantPorterRuntime)

	wantExecMixin := "cnab/app/mixins/exec/exec-runtime"
	execMixinExists, err := p.FileSystem.Exists(wantExecMixin)
	require.NoError(t, err)
	assert.True(t, execMixinExists, "The exec-runtime mixin wasn't copied into %s", wantExecMixin)
}

func TestPorter_buildBundle(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	err = p.LoadManifest()
	require.NoError(t, err)

	err = p.buildBundle("foo", "digest")
	require.NoError(t, err)

	bundleJSONExists, err := p.FileSystem.Exists("cnab/bundle.json")
	require.NoError(t, err)
	require.True(t, bundleJSONExists, "cnab/bundle.json wasn't written")

	f, _ := p.FileSystem.Stat("cnab/bundle.json")
	if f.Size() == 0 {
		t.Fatalf("cnab/bundle.json is empty")
	}

	bundleBytes, err := p.FileSystem.ReadFile("cnab/bundle.json")
	require.NoError(t, err)

	var bundle bundle.Bundle
	err = json.Unmarshal(bundleBytes, &bundle)
	require.NoError(t, err)

	assert.Equal(t, bundle.Name, "HELLO")
	assert.Equal(t, bundle.Version, "0.1.0")
	assert.Equal(t, bundle.Description, "An example Porter configuration")

	stamp, err := p.LoadStamp(bundle)
	require.NoError(t, err)
	assert.Equal(t, "06a51d04297375bf111ab15e579b8a7ab72e2661018c4d08d1d3f38198028e49", stamp.ManifestDigest)

	debugParam, ok := bundle.Parameters["porter-debug"]
	require.True(t, ok)
	assert.Equal(t, "PORTER_DEBUG", debugParam.Destination.EnvironmentVariable)
	assert.Equal(t, "bool", debugParam.DataType)
	assert.Equal(t, false, debugParam.DefaultValue)
}

func TestPorter_paramRequired(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.TestConfig.TestContext.AddTestFile("./testdata/paramafest.yaml", config.Name)

	err := p.LoadManifest()
	require.NoError(t, err)

	err = p.buildBundle("foo", "digest")
	require.NoError(t, err)

	bundleBytes, err := p.FileSystem.ReadFile("cnab/bundle.json")
	require.NoError(t, err)

	var bundle bundle.Bundle
	err = json.Unmarshal(bundleBytes, &bundle)
	require.NoError(t, err)

	p1, ok := bundle.Parameters["command"]
	require.True(t, ok)
	require.False(t, p1.Required)

	p2, ok := bundle.Parameters["command2"]
	require.True(t, ok)
	require.True(t, p2.Required)
}
