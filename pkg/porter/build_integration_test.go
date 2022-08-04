//go:build integration
// +build integration

package porter

import (
	"context"
	"encoding/json"
	"io/fs"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/yaml"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_Build(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	configTpl, err := p.Templates.GetManifest()
	require.Nil(t, err)
	p.TestConfig.TestContext.AddTestFileContents(configTpl, config.Name)

	// Create some junk in the previous .cnab directory, build should clean it up and not copy it into the bundle
	junkDir := ".cnab/test/junk"
	require.NoError(t, p.FileSystem.MkdirAll(junkDir, pkg.FileModeDirectory), "could not create test junk files")
	junkExists, _ := p.FileSystem.DirExists(junkDir)
	assert.True(t, junkExists, "failed to create junk files for the test")

	opts := BuildOptions{}
	require.NoError(t, opts.Validate(p.Porter), "Validate failed")

	err = p.Build(context.Background(), opts)
	require.NoError(t, err)

	// Check file permissions on .cnab contents
	bundleJSONStats, err := p.FileSystem.Stat(build.LOCAL_BUNDLE)
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, build.LOCAL_BUNDLE, pkg.FileModeWritable, bundleJSONStats.Mode())

	runStats, err := p.FileSystem.Stat(build.LOCAL_RUN)
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, build.LOCAL_RUN, pkg.FileModeExecutable, runStats.Mode())

	manifestStats, err := p.FileSystem.Stat(build.LOCAL_MANIFEST)
	require.NoError(t, err)
	tests.AssertFilePermissionsEqual(t, build.LOCAL_MANIFEST, pkg.FileModeWritable, manifestStats.Mode())

	err = p.FileSystem.Walk(build.LOCAL_MIXINS, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		tests.AssertFilePermissionsEqual(t, path, pkg.FileModeExecutable, runStats.Mode())
		return nil
	})
	require.NoError(t, err)

	// Check that the junk files were cleaned up
	junkExists, _ = p.FileSystem.DirExists(junkDir)
	assert.False(t, junkExists, "junk files were not cleaned up before building")

	bun, err := p.CNAB.LoadBundle(build.LOCAL_BUNDLE)
	require.NoError(t, err)

	assert.Equal(t, "porter-hello", bun.Name)
	assert.Equal(t, "1.2.0", string(bun.SchemaVersion))
	assert.Equal(t, "0.1.0", bun.Version)
	assert.Equal(t, "An example Porter configuration", bun.Description)

	stamp, err := configadapter.LoadStamp(bun)
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

func TestPorter_Build_ChecksManifestSchemaVersion(t *testing.T) {
	testcases := []struct {
		name          string
		schemaVersion string
		wantErr       string
	}{
		{name: "current version", schemaVersion: manifest.DefaultSchemaVersion.String()},
		{name: "its an older code but it checks out", schemaVersion: "1.0.0-alpha.1"},
		{name: "invalid version", schemaVersion: "", wantErr: schema.ErrInvalidSchemaVersion.Error()},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			// Make a bundle with the specified schemaVersion
			p.TestConfig.TestContext.AddTestDirectoryFromRoot("tests/testdata/mybuns", "/")
			e := yaml.NewEditor(p.Context)
			require.NoError(t, e.ReadFile("porter.yaml"))
			require.NoError(t, e.SetValue("schemaVersion", tc.schemaVersion))
			require.NoError(t, e.WriteFile("porter.yaml"))

			opts := BuildOptions{}
			opts.File = "porter.yaml"
			err := p.Build(context.Background(), opts)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, schema.ErrInvalidSchemaVersion)
				tests.RequireErrorContains(t, err, tc.wantErr)
			}
		})
	}
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
		defer p.Close()

		testMixins := p.Mixins.(*mixin.TestMixinProvider)
		testMixins.LintResults = lintResults

		err := p.Create()
		require.NoError(t, err, "Create failed")

		opts := BuildOptions{NoLint: false}
		err = opts.Validate(p.Porter)
		require.NoError(t, err)

		err = p.Build(context.Background(), opts)
		require.Errorf(t, err, "Build should have been aborted with lint errors")
		assert.Contains(t, err.Error(), "lint errors were detected")
	})

	t.Run("ignores lint error with --no-lint", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		testMixins := p.Mixins.(*mixin.TestMixinProvider)
		testMixins.LintResults = lintResults

		err := p.Create()
		require.NoError(t, err, "Create failed")

		opts := BuildOptions{NoLint: true}
		err = opts.Validate(p.Porter)
		require.NoError(t, err)

		err = p.Build(context.Background(), opts)
		require.NoError(t, err, "Build failed but should have not run lint")
	})

}

func TestPorter_paramRequired(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("./testdata/paramafest.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, p.Config, config.Name)
	require.NoError(t, err)

	err = p.buildBundle(ctx, m, "digest")
	require.NoError(t, err)

	bundleBytes, err := p.FileSystem.ReadFile(build.LOCAL_BUNDLE)
	require.NoError(t, err)

	var bundle bundle.Bundle
	err = json.Unmarshal(bundleBytes, &bundle)
	require.NoError(t, err)

	require.False(t, bundle.Parameters["command"].Required, "expected command param to not be required")
	require.True(t, bundle.Parameters["command2"].Required, "expected command2 param to be required")
}

func TestBuildOptions_Validate(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	testcases := []struct {
		name       string
		opts       BuildOptions
		wantDriver string
		wantError  string
	}{{
		name:       "no opts",
		opts:       BuildOptions{},
		wantDriver: config.BuildDriverBuildkit,
	}, {
		name:      "invalid version set - latest",
		opts:      BuildOptions{metadataOpts: metadataOpts{Version: "latest"}},
		wantError: `invalid bundle version: "latest" is not a valid semantic version`,
	}, {
		name: "valid version - v prefix",
		opts: BuildOptions{metadataOpts: metadataOpts{Version: "v1.0.0"}},
	}, {
		name: "valid version - with hash",
		opts: BuildOptions{metadataOpts: metadataOpts{Version: "v0.1.7+58d98af56c3a4c40c69535654216bd4a1fa701e7"}},
	}, {
		name: "valid name and value set",
		opts: BuildOptions{metadataOpts: metadataOpts{Name: "newname", Version: "1.0.0"}},
	}, {
		name:      "deprecated driver: docker",
		opts:      BuildOptions{Driver: config.BuildDriverDocker},
		wantError: `invalid --driver value docker`,
	}, {
		name:       "valid driver: buildkit",
		opts:       BuildOptions{Driver: config.BuildDriverBuildkit},
		wantDriver: config.BuildDriverBuildkit,
	}, {
		name:      "invalid driver",
		opts:      BuildOptions{Driver: "missing-driver"},
		wantError: `invalid --driver value missing-driver`,
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate(p.Porter)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)

				if tc.wantDriver != "" {
					assert.Equal(t, tc.wantDriver, p.Data.BuildDriver)
				}
			}
		})
	}
}

func TestBuildOptions_Defaults(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	t.Run("default driver", func(t *testing.T) {
		opts := BuildOptions{}
		err := opts.Validate(p.Porter)
		require.NoError(t, err, "Validate failed")
		assert.Equal(t, config.BuildDriverBuildkit, opts.Driver)
	})
}

func TestPorter_BuildWithCustomValues(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("./testdata/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, p.Config, config.Name)
	require.NoError(t, err)

	err = p.buildBundle(ctx, m, "digest")
	require.NoError(t, err)

	opts := BuildOptions{Customs: []string{"customKey1=editedCustomValue1"}}
	require.NoError(t, opts.Validate(p.Porter), "Validate failed")

	err = p.Build(ctx, opts)
	require.NoError(t, err)

	bun, err := p.CNAB.LoadBundle(build.LOCAL_BUNDLE)
	require.NoError(t, err)

	assert.Equal(t, bun.Custom["customKey1"], "editedCustomValue1")
}
