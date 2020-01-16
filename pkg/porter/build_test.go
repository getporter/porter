package porter

import (
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/build"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	bundleJSONExists, err := p.FileSystem.Exists(build.LOCAL_BUNDLE)
	require.NoError(t, err)
	require.True(t, bundleJSONExists, "%s wasn't written", build.LOCAL_BUNDLE)

	f, _ := p.FileSystem.Stat(build.LOCAL_BUNDLE)
	if f.Size() == 0 {
		t.Fatalf("%s is empty", build.LOCAL_BUNDLE)
	}

	bundleBytes, err := p.FileSystem.ReadFile(build.LOCAL_BUNDLE)
	require.NoError(t, err)

	bun := &bundle.Bundle{}
	err = json.Unmarshal(bundleBytes, bun)
	require.NoError(t, err)

	assert.Equal(t, bun.Name, "HELLO")
	assert.Equal(t, bun.Version, "0.1.0")
	assert.Equal(t, bun.Description, "An example Porter configuration")

	stamp, err := configadapter.LoadStamp(bun)
	require.NoError(t, err)
	assert.Equal(t, "cf9472c4f17e8dc0c700ba1ac6a71a5d0759731cfa5f21aa92e125026c21c1b7", stamp.ManifestDigest)

	debugParam, ok := bun.Parameters["porter-debug"]
	require.True(t, ok, "porter-debug parameter was not defined")
	assert.Equal(t, "PORTER_DEBUG", debugParam.Destination.EnvironmentVariable)
	debugDef, ok := bun.Definitions[debugParam.Definition]
	require.True(t, ok, "porter-debug definition was not defined")
	assert.Equal(t, "boolean", debugDef.Type)
	assert.Equal(t, false, debugDef.Default)
}

func TestPorter_paramRequired(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.TestConfig.TestContext.AddTestFile("./testdata/paramafest.yaml", config.Name)

	err := p.LoadManifest()
	require.NoError(t, err)

	err = p.buildBundle("foo", "digest")
	require.NoError(t, err)

	bundleBytes, err := p.FileSystem.ReadFile(build.LOCAL_BUNDLE)
	require.NoError(t, err)

	var bundle bundle.Bundle
	err = json.Unmarshal(bundleBytes, &bundle)
	require.NoError(t, err)

	require.False(t, bundle.Parameters["command"].Required, "expected command param to not be required")
	require.True(t, bundle.Parameters["command2"].Required, "expected command2 param to be required")
}
