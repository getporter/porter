package porter

import (
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/build"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/mixin"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_buildBundle(t *testing.T) {
	p := NewTestPorter(t)

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

	assert.Equal(t, bun.Name, "porter-hello")
	assert.Equal(t, bun.Version, "0.1.0")
	assert.Equal(t, bun.Description, "An example Porter configuration")

	stamp, err := configadapter.LoadStamp(*bun)
	require.NoError(t, err)
	assert.Equal(t, "d421a6249dfbdba79e26e866da7533d59590565708dfdb32423cf989f588d0ea", stamp.ManifestDigest)

	debugParam, ok := bun.Parameters["porter-debug"]
	require.True(t, ok, "porter-debug parameter was not defined")
	assert.Equal(t, "PORTER_DEBUG", debugParam.Destination.EnvironmentVariable)
	debugDef, ok := bun.Definitions[debugParam.Definition]
	require.True(t, ok, "porter-debug definition was not defined")
	assert.Equal(t, "boolean", debugDef.Type)
	assert.Equal(t, false, debugDef.Default)
}

func TestPorter_LintDuringBuild(t *testing.T) {
	lintResults := linter.Results{
		{
			Level: linter.LevelError,
			Code:  "exec-100",
		},
	}

	t.Run("failing lint should stop build", func(t *testing.T) {
		p := NewTestPorter(t)
		testMixins := p.Mixins.(*mixin.TestMixinProvider)
		testMixins.LintResults = lintResults

		err := p.Create()
		require.NoError(t, err, "Create failed")

		opts := BuildOptions{NoLint: false}
		err = p.Build(opts)
		require.Errorf(t, err, "Build should have been aborted with lint errors")
		assert.Contains(t, err.Error(), "Lint errors were detected")
	})

	t.Run("ignores lint error with --no-lint", func(t *testing.T) {
		p := NewTestPorter(t)
		testMixins := p.Mixins.(*mixin.TestMixinProvider)
		testMixins.LintResults = lintResults

		err := p.Create()
		require.NoError(t, err, "Create failed")

		opts := BuildOptions{NoLint: true}
		err = p.Build(opts)
		require.NoError(t, err, "Build failed but should have not run lint")
	})

}

func TestPorter_paramRequired(t *testing.T) {
	p := NewTestPorter(t)
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

func TestValidateBuildOpts(t *testing.T) {
	testcases := []struct {
		name      string
		opts      BuildOptions
		wantError string
	}{{
		name:      "no opts",
		opts:      BuildOptions{},
		wantError: "",
	}, {
		name:      "invalid version set - latest",
		opts:      BuildOptions{metadataOpts: metadataOpts{Version: "latest"}},
		wantError: `invalid bundle version: "latest" is not a valid semantic version`,
	}, {
		name:      "invalid version set - v prefix",
		opts:      BuildOptions{metadataOpts: metadataOpts{Version: "v1.0.0"}},
		wantError: `invalid bundle version: "v1.0.0" is not a valid semantic version`,
	}, {
		name:      "valid name and value set",
		opts:      BuildOptions{metadataOpts: metadataOpts{Name: "newname", Version: "1.0.0"}},
		wantError: "",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
