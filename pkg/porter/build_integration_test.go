// +build integration

package porter

import (
	"encoding/json"
	"io/fs"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/build"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_Build(t *testing.T) {
	p := NewTestPorter(t)
	p.SetupIntegrationTest() // Build on the filesystem so we can test file permissions
	defer p.CleanupIntegrationTest()

	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	// Create some junk in the previous .cnab directory, build should clean it up and not copy it into the bundle
	junkDir := ".cnab/test/junk"
	require.NoError(t, p.FileSystem.MkdirAll(junkDir, 0700), "could not create test junk files")
	junkExists, _ := p.FileSystem.DirExists(junkDir)
	assert.True(t, junkExists, "failed to create junk files for the test")

	err = p.LoadManifest()
	require.NoError(t, err)

	opts := BuildOptions{}
	require.NoError(t, opts.Validate(p.Context), "Validate failed")

	err = p.Build(opts)
	require.NoError(t, err)

	// Check file permissions on .cnab contents
	bundleJSONStats, err := p.FileSystem.Stat(build.LOCAL_BUNDLE)
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, build.LOCAL_BUNDLE, os.FileMode(0600), bundleJSONStats.Mode())

	runStats, err := p.FileSystem.Stat(build.LOCAL_RUN)
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, build.LOCAL_RUN, os.FileMode(0700), runStats.Mode())

	manifestStats, err := p.FileSystem.Stat(build.LOCAL_MANIFEST)
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, build.LOCAL_MANIFEST, os.FileMode(0600), manifestStats.Mode())

	err = p.FileSystem.Walk(build.LOCAL_MIXINS, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		tests.AssertFilePermissionsEqual(t, path, os.FileMode(0700), runStats.Mode())
		return nil
	})
	require.NoError(t, err)

	// Check that the junk files were cleaned up
	junkExists, _ = p.FileSystem.DirExists(junkDir)
	assert.False(t, junkExists, "junk files were not cleaned up before building")

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
	assert.NotEmpty(t, stamp.ManifestDigest)

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
		err = opts.Validate(p.Context)
		require.NoError(t, err)

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
		err = opts.Validate(p.Context)
		require.NoError(t, err)

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
	p := NewTestPorter(t)

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
		name:      "valid version - v prefix",
		opts:      BuildOptions{metadataOpts: metadataOpts{Version: "v1.0.0"}},
		wantError: "",
	}, {
		name:      "valid version - with hash",
		opts:      BuildOptions{metadataOpts: metadataOpts{Version: "v0.1.7+58d98af56c3a4c40c69535654216bd4a1fa701e7"}},
		wantError: "",
	}, {
		name:      "valid name and value set",
		opts:      BuildOptions{metadataOpts: metadataOpts{Name: "newname", Version: "1.0.0"}},
		wantError: "",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate(p.Context)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
