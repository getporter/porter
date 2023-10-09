package configadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	depsv1ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	depsv2ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestConverter(t *testing.T) {

	testcases := []struct {
		name          string
		configHandler func(c *config.Config)
		manifestPath  string
		goldenFile    string
	}{
		{name: "mybuns depsv1",
			configHandler: func(c *config.Config) {},
			manifestPath:  "tests/testdata/mybuns/porter.yaml",
			goldenFile:    "testdata/mybuns-depsv1.bundle.json"},
		{name: "mybuns depsv2",
			configHandler: func(c *config.Config) {
				c.SetExperimentalFlags(experimental.FlagDependenciesV2)
			},
			manifestPath: "tests/testdata/mybuns/porter.yaml",
			goldenFile:   "testdata/mybuns-depsv2.bundle.json"},
		{name: "myenv depsv2",
			configHandler: func(c *config.Config) {
				c.SetExperimentalFlags(experimental.FlagDependenciesV2)
			},
			manifestPath: "tests/testdata/myenv/porter.yaml",
			goldenFile:   "testdata/myenv-depsv2.bundle.json"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := config.NewTestConfig(t)
			tc.configHandler(c.Config)
			c.TestContext.AddTestFileFromRoot(tc.manifestPath, config.Name)

			ctx := context.Background()
			m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")

			installedMixins := []mixin.Metadata{
				{Name: "exec", VersionInfo: pkgmgmt.VersionInfo{Version: "v1.2.3"}},
			}

			a := NewManifestConverter(c.Config, m, nil, installedMixins)

			bun, err := a.ToBundle(ctx)
			require.NoError(t, err, "ToBundle failed")

			// Compare the regular json, not the canonical, because that's hard to diff
			prepBundleForDiff(&bun.Bundle)
			bunD, err := json.MarshalIndent(bun, "", "  ")
			require.NoError(t, err)
			c.TestContext.CompareGoldenFile(tc.goldenFile, string(bunD))
		})
	}
}

func prepBundleForDiff(b *bundle.Bundle) {
	// Unset the digest when we are comparing test bundle files because
	// otherwise the digest changes based on the version of the porter binary +
	// mixins that generated it, which makes the file change a lot
	// unnecessarily.
	stamp := b.Custom[cnab.PorterExtension].(Stamp)
	stamp.ManifestDigest = ""
	b.Custom[cnab.PorterExtension] = stamp
}

func TestManifestConverter_ToBundle(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	bun, err := a.ToBundle(ctx)
	require.NoError(t, err, "ToBundle failed")

	assert.Equal(t, cnab.BundleSchemaVersion(), bun.SchemaVersion)
	assert.Equal(t, "mybuns", bun.Name)
	assert.Equal(t, "0.1.2", bun.Version)
	assert.Equal(t, "A very thorough test bundle", bun.Description)

	stamp, err := LoadStamp(bun)
	require.NoError(t, err, "could not load porter's stamp")
	assert.NotNil(t, stamp)

	assert.Contains(t, bun.Actions, "status", "custom action 'status' was not populated")

	require.Len(t, bun.Credentials, 2, "expected two credentials")
	assert.Contains(t, bun.Parameters, "porter-debug", "porter-debug parameter was not defined")
	assert.Contains(t, bun.Definitions, "porter-debug-parameter", "porter-debug definition was not defined")

	assert.True(t, bun.HasDependenciesV1(), "DependenciesV1 was not populated")
	assert.Contains(t, bun.RequiredExtensions, "io.cnab.dependencies")

	assert.NotEmpty(t, bun.Outputs, "expected multiple outputs generated")
}

func TestManifestConverter_generateBundleCredentials(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	bun, err := a.ToBundle(ctx)
	require.NoError(t, err, "ToBundle failed")

	assert.Contains(t, bun.Credentials, "username", "credential 'username' was not populated")
	username := bun.Credentials["username"]
	assert.Equal(t, "The name you want on the audit log", username.Description, "credential.Description was not populated")
	assert.False(t, username.Required, "credential.Required was not populated correctly")
	assert.Equal(t, "ROOT_USERNAME", username.EnvironmentVariable, "credential.EnvironmentVariable was not populated")

	assert.Contains(t, bun.Credentials, "password", "credential 'password' was not populated")
	password := bun.Credentials["password"]
	assert.True(t, password.Required, "credential.Required was not populated correctly")
	assert.Equal(t, []string{"boom"}, password.ApplyTo, "credential.ApplyTo was not populated")
	assert.Equal(t, "/tmp/password", password.Path, "credential.Path was not populated")
}

func TestManifestConverter_generateBundleParametersSchema(t *testing.T) {
	testcases := []struct {
		propname  string
		wantParam bundle.Parameter
		wantDef   definition.Schema
	}{
		{"ainteger",
			bundle.Parameter{
				Definition: "ainteger-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "AINTEGER",
				},
			},
			definition.Schema{
				Type:    "integer",
				Default: 1,
				Minimum: toFloat(0),
				Maximum: toFloat(10),
			},
		},
		{"anumber",
			bundle.Parameter{
				Definition: "anumber-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "ANUMBER",
				},
			},
			definition.Schema{
				Type:             "number",
				Default:          0.5,
				ExclusiveMinimum: toFloat(0),
				ExclusiveMaximum: toFloat(1),
			},
		},
		{
			"astringenum",
			bundle.Parameter{
				Definition: "astringenum-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "ASTRINGENUM",
				},
			},
			definition.Schema{
				Type:    "string",
				Default: "blue",
				Enum:    []interface{}{"blue", "red", "purple", "pink"},
			},
		},
		{
			"astring",
			bundle.Parameter{
				Definition: "astring-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "ASTRING",
				},
				Required: true,
			},
			definition.Schema{
				Type:      "string",
				MinLength: toInt(1),
				MaxLength: toInt(10),
			},
		},
		{
			"aboolean",
			bundle.Parameter{
				Definition: "aboolean-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "ABOOLEAN",
				},
			},
			definition.Schema{
				Type:    "boolean",
				Default: true,
			},
		},
		{
			"installonly",
			bundle.Parameter{
				Definition: "installonly-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "INSTALLONLY",
				},
				ApplyTo: []string{
					"install",
				},
				Required: true,
			},
			definition.Schema{
				Type: "boolean",
			},
		},
		{
			"sensitive",
			bundle.Parameter{
				Definition: "sensitive-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "SENSITIVE",
				},
				Required: true,
			},
			definition.Schema{
				Type:      "string",
				WriteOnly: toBool(true),
			},
		},
		{
			"jsonobject",
			bundle.Parameter{
				Definition: "jsonobject-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "JSONOBJECT",
				},
			},
			definition.Schema{
				Type:    "string",
				Default: `"myobject": { "foo": "true", "bar": [ 1, 2, 3 ] }`,
			},
		},
		{
			"afile",
			bundle.Parameter{
				Definition: "afile-parameter",
				Destination: &bundle.Location{
					Path: "/home/nonroot/.kube/config",
				},
				Required: true,
			},
			definition.Schema{
				Type:            "string",
				ContentEncoding: "base64",
			},
		},
		{
			"notype-file",
			bundle.Parameter{
				Definition: "notype-file-parameter",
				Destination: &bundle.Location{
					Path: "/cnab/app/config.toml",
				},
				Required: true,
			},
			definition.Schema{
				Type:            "string",
				ContentEncoding: "base64",
			},
		},
		{
			"notype-string",
			bundle.Parameter{
				Definition: "notype-string-parameter",
				Destination: &bundle.Location{
					EnvironmentVariable: "NOTYPE_STRING",
				},
				Required: true,
			},
			definition.Schema{
				Type: "string",
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.propname, func(t *testing.T) {
			t.Parallel()
			c := config.NewTestConfig(t)
			c.TestContext.AddTestFile("testdata/porter-with-parameters.yaml", config.Name)

			ctx := context.Background()
			m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")

			a := NewManifestConverter(c.Config, m, nil, nil)

			defs := make(definition.Definitions, len(m.Parameters))
			params := a.generateBundleParameters(ctx, &defs)

			param, ok := params[tc.propname]
			require.True(t, ok, "parameter definition was not generated")

			def, ok := defs[param.Definition]
			require.True(t, ok, "property definition was not generated")

			assert.Equal(t, tc.wantParam, param)
			assert.Equal(t, tc.wantDef, *def)
		})
	}
}

func TestManifestConverter_buildDefaultPorterParameters(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	defs := make(definition.Definitions, len(m.Parameters))
	params := a.generateBundleParameters(ctx, &defs)

	debugParam, ok := params["porter-debug"]
	assert.True(t, ok, "porter-debug parameter was not defined")
	assert.Equal(t, "porter-debug-parameter", debugParam.Definition)
	assert.Equal(t, "PORTER_DEBUG", debugParam.Destination.EnvironmentVariable)

	debugDef, ok := defs["porter-debug-parameter"]
	require.True(t, ok, "porter-debug definition was not defined")
	assert.Equal(t, "boolean", debugDef.Type)
	assert.Equal(t, false, debugDef.Default)
}

func TestManifestConverter_generateImages(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	mappedImage := manifest.MappedImage{
		Description: "un petite server",
		Repository:  "getporter/myserver",
		ImageType:   "docker",
		Digest:      "abc123",
		Size:        12,
		MediaType:   "download",
		Labels: map[string]string{
			"OS":           "linux",
			"Architecture": "amd64",
		},
	}
	a.Manifest.ImageMap = map[string]manifest.MappedImage{
		"server": mappedImage,
	}

	images := a.generateBundleImages()

	require.Len(t, images, 1)
	img := images["server"]
	assert.Equal(t, mappedImage.Description, img.Description)
	assert.Equal(t, fmt.Sprintf("%s@%s", mappedImage.Repository, mappedImage.Digest), img.Image)
	assert.Equal(t, mappedImage.ImageType, img.ImageType)
	assert.Equal(t, mappedImage.Digest, img.Digest)
	assert.Equal(t, mappedImage.Size, img.Size)
	assert.Equal(t, mappedImage.MediaType, img.MediaType)
	assert.Equal(t, mappedImage.Labels["OS"], img.Labels["OS"])
	assert.Equal(t, mappedImage.Labels["Architecture"], img.Labels["Architecture"])
}

func TestManifestConverter_generateBundleImages_EmptyLabels(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	mappedImage := manifest.MappedImage{
		Description: "un petite server",
		Repository:  "getporter/myserver",
		Tag:         "1.0.0",
		ImageType:   "docker",
		Labels:      nil,
	}
	a.Manifest.ImageMap = map[string]manifest.MappedImage{
		"server": mappedImage,
	}

	images := a.generateBundleImages()
	require.Len(t, images, 1)
	img := images["server"]
	assert.Nil(t, img.Labels)
	assert.Equal(t, fmt.Sprintf("%s:%s", mappedImage.Repository, mappedImage.Tag), img.Image)
}

func TestManifestConverter_generateBundleOutputs(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/simple.porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	outputDefinitions := manifest.OutputDefinitions{
		"output1": {
			Name: "output1",
			ApplyTo: []string{
				"install",
				"upgrade",
			},
			Schema: definition.Schema{
				Type:        "string",
				Description: "Description of output1",
			},
			Sensitive: true,
		},
		"output2": {
			Name: "output2",
			Schema: definition.Schema{
				Type:        "boolean",
				Description: "Description of output2",
			},
		},
		"kubeconfig": {
			Name: "kubeconfig",
			Path: "/home/nonroot/.kube/config",
			Schema: definition.Schema{
				Type:        "file",
				Description: "Description of kubeconfig",
			},
		},
		"notype-string": {
			Name: "notype-string",
		},
		"notype-file": {
			Name: "notype-file",
			Path: "/home/nonroot/.kube/config",
		},
	}

	a.Manifest.Outputs = outputDefinitions

	defs := make(definition.Definitions, len(a.Manifest.Outputs))
	outputs := a.generateBundleOutputs(ctx, &defs)
	require.Len(t, defs, 6)

	wantOutputDefinitions := map[string]bundle.Output{
		"output1": {
			Definition:  "output1-output",
			Description: "Description of output1",
			ApplyTo: []string{
				"install",
				"upgrade",
			},
			Path: "/cnab/app/outputs/output1",
		},
		"output2": {
			Definition:  "output2-output",
			Description: "Description of output2",
			Path:        "/cnab/app/outputs/output2",
		},
		"kubeconfig": {
			Definition:  "kubeconfig-output",
			Description: "Description of kubeconfig",
			Path:        "/cnab/app/outputs/kubeconfig",
		},
		"notype-string": {
			Definition: "notype-string-output",
			Path:       "/cnab/app/outputs/notype-string",
		},
		"notype-file": {
			Definition: "notype-file-output",
			Path:       "/cnab/app/outputs/notype-file",
		},
		"porter-state": {
			Description: "Supports persisting state for bundles. Porter internal parameter that should not be set manually.",
			Definition:  "porter-state",
			Path:        "/cnab/app/outputs/porter-state",
		},
	}

	require.Equal(t, wantOutputDefinitions, outputs)

	wantDefinitions := definition.Definitions{
		"output1-output": &definition.Schema{
			Type:        "string",
			Description: "Description of output1",
			WriteOnly:   toBool(true),
		},
		"output2-output": &definition.Schema{
			Type:        "boolean",
			Description: "Description of output2",
		},
		"kubeconfig-output": &definition.Schema{
			Type:            "string",
			ContentEncoding: "base64",
			Description:     "Description of kubeconfig",
		},
		"notype-string-output": &definition.Schema{
			Type: "string",
		},
		"notype-file-output": &definition.Schema{
			Type:            "string",
			ContentEncoding: "base64",
		},
		"porter-state": &definition.Schema{
			ID:              "https://getporter.org/generated-bundle/#porter-state",
			Comment:         "porter-internal",
			Description:     "Supports persisting state for bundles. Porter internal parameter that should not be set manually.",
			Type:            "string",
			ContentEncoding: "base64",
		},
	}

	require.Equal(t, wantDefinitions, defs)
}

func TestManifestConverter_generateDependenciesv1(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		wantDep depsv1ext.Dependency
	}{
		{"no-version", depsv1ext.Dependency{
			Name:   "mysql",
			Bundle: "getporter/azure-mysql:5.7",
		}},
		{"no-ranges, uses prerelease", depsv1ext.Dependency{
			Name:   "ad",
			Bundle: "getporter/azure-active-directory",
			Version: &depsv1ext.DependencyVersion{
				AllowPrereleases: true,
				Ranges:           []string{"1.0.0-0"},
			},
		}},
		{"with-ranges", depsv1ext.Dependency{
			Name:   "storage",
			Bundle: "getporter/azure-blob-storage",
			Version: &depsv1ext.DependencyVersion{
				Ranges: []string{
					"1.x - 2,2.1 - 3.x",
				},
			},
		}},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := config.NewTestConfig(t)
			c.TestContext.AddTestFile("testdata/porter-with-deps.yaml", config.Name)

			ctx := context.Background()
			m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")

			a := NewManifestConverter(c.Config, m, nil, nil)
			defs := make(definition.Definitions, len(m.Parameters))

			depsExt, depsExtKey, err := a.generateDependencies(ctx, &defs)
			require.NoError(t, err)
			require.Equal(t, cnab.DependenciesV1ExtensionKey, depsExtKey, "expected the v1 dependencies extension key")
			require.IsType(t, &depsv1ext.Dependencies{}, depsExt, "expected a v1 dependencies extension section")
			deps := depsExt.(*depsv1ext.Dependencies)
			require.Len(t, deps.Requires, 3, "incorrect number of dependencies were generated")
			require.Equal(t, []string{"mysql", "ad", "storage"}, deps.Sequence, "incorrect sequence was generated")

			var dep *depsv1ext.Dependency
			for _, d := range deps.Requires {
				if d.Bundle == tc.wantDep.Bundle {
					dep = &d
					break
				}
			}

			require.NotNil(t, dep, "could not find bundle %s", tc.wantDep.Bundle)
			assert.Equal(t, &tc.wantDep, dep)
		})
	}
}

func TestManifestConverter_generateDependenciesv2(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		wantDep  depsv2ext.Dependency
		wantDefs definition.Definitions
	}{
		{name: "all fields", wantDep: depsv2ext.Dependency{
			Name:    "mysql",
			Bundle:  "getporter/azure-mysql:5.7",
			Version: "5.7.x",
			Interface: &depsv2ext.DependencyInterface{
				ID:        "https://getporter.org/interfaces/#mysql",
				Reference: "getporter/mysql-spec:5.7",
				Document: depsv2ext.DependencyInterfaceDocument{
					Outputs: map[string]bundle.Output{
						"myoutput": {
							Definition:  "myoutput-output",
							Description: "worlds smallest output",
							Path:        "/cnab/app/outputs/myoutput",
						},
					},
					Parameters: map[string]bundle.Parameter{
						"myparam": {
							Definition:  "myparam-parameter",
							Description: "worlds biggest param",
							Required:    false,
							Destination: &bundle.Location{
								Path:                "",
								EnvironmentVariable: "MYPARAM",
							},
						},
					},
					Credentials: map[string]bundle.Credential{
						"mycred": {
							Description: "credential",
							Required:    true,
						},
					},
				},
			},
			Sharing: depsv2ext.SharingCriteria{
				Mode:  true,
				Group: depsv2ext.SharingGroup{Name: "myapp"},
			},
			Parameters: map[string]string{
				"database":  "wordpress",
				"collation": "${bundle.parameters.db_collation}",
			},
			Credentials: map[string]string{
				"user": "${bundle.credentials.username}",
			},
		},
			wantDefs: map[string]*definition.Schema{
				"myoutput-output": {
					Type:        "string",
					Description: "worlds smallest output",
				},
				"myparam-parameter": {
					Type:        "string",
					Default:     false,
					Description: "worlds biggest param",
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := config.NewTestConfig(t)
			c.SetExperimentalFlags(experimental.FlagDependenciesV2)
			c.TestContext.AddTestFile("testdata/porter-with-depsv2.yaml", config.Name)

			ctx := context.Background()
			m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")

			a := NewManifestConverter(c.Config, m, nil, nil)
			defs := make(definition.Definitions, len(m.Parameters))

			depsExt, depsExtKey, err := a.generateDependencies(ctx, &defs)
			require.NoError(t, err)
			require.Equal(t, cnab.DependenciesV2ExtensionKey, depsExtKey, "expected the v1 dependencies extension key")
			require.IsType(t, &depsv2ext.Dependencies{}, depsExt, "expected a v1 dependencies extension section")
			deps := depsExt.(*depsv2ext.Dependencies)
			require.Len(t, deps.Requires, 3, "incorrect number of dependencies were generated")

			var dep *depsv2ext.Dependency
			for _, d := range deps.Requires {
				if d.Bundle == tc.wantDep.Bundle {
					dep = &d
					break
				}
			}

			require.NotNil(t, dep, "could not find bundle %s", tc.wantDep.Bundle)
			assert.Equal(t, &tc.wantDep, dep)
			assert.Equal(t, tc.wantDefs, defs)
		})
	}
}

func TestManifestConverter_generateParameterSources(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	b, err := a.ToBundle(ctx)
	require.NoError(t, err, "ToBundle failed")
	sources, err := b.ReadParameterSources()
	require.NoError(t, err, "ReadParameterSources failed")

	want := cnab.ParameterSources{}
	want.SetParameterFromOutput("porter-msg-output", "msg")
	want.SetParameterFromOutput("tfstate", "tfstate")
	want.SetParameterFromOutput("porter-state", "porter-state")
	want.SetParameterFromDependencyOutput("mysql-connstr", "db", "connstr")

	assert.Equal(t, want, sources)
}

func TestNewManifestConverter_generateOutputWiringParameter(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	outputDef := definition.Schema{
		Type: "string",
	}
	b := cnab.NewBundle(bundle.Bundle{
		Outputs: map[string]bundle.Output{
			"msg": {
				Definition: "stringDef",
			},
			"some-thing": {
				Definition: "stringDef",
			},
		},
		Definitions: map[string]*definition.Schema{
			"stringDef": &outputDef,
		},
	})

	t.Run("generate parameter", func(t *testing.T) {
		t.Parallel()

		name, param, paramDef := a.generateOutputWiringParameter(b, "msg")

		assert.Equal(t, "porter-msg-output", name, "unexpected parameter name")
		assert.False(t, param.Required, "wiring parameters should NOT be required")
		require.NotNil(t, param.Destination, "wiring parameters should have a destination set")
		assert.Equal(t, "PORTER_MSG_OUTPUT", param.Destination.EnvironmentVariable, "unexpected destination environment variable set")

		assert.Equal(t, "https://getporter.org/generated-bundle/#porter-parameter-source-definition", paramDef.ID, "wiring parameter should have a schema id set")
		assert.NotSame(t, outputDef, paramDef, "wiring parameter definition should be a copy")
		assert.Equal(t, outputDef.Type, paramDef.Type, "output def and param def should have the same type")
		assert.Equal(t, cnab.PorterInternal, paramDef.Comment, "wiring parameter should be flagged as internal")
	})

	t.Run("param with hyphen", func(t *testing.T) {
		t.Parallel()

		name, param, _ := a.generateOutputWiringParameter(b, "some-thing")

		assert.Equal(t, "porter-some-thing-output", name, "unexpected parameter name")
		require.NotNil(t, param.Destination, "wiring parameters should have a destination set")
		assert.Equal(t, "PORTER_SOME_THING_OUTPUT", param.Destination.EnvironmentVariable, "unexpected destination environment variable set")
	})
}

func TestNewManifestConverter_generateDependencyOutputWiringParameter(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	ref := manifest.DependencyOutputReference{Dependency: "mysql", Output: "mysql-password"}
	name, param, paramDef := a.generateDependencyOutputWiringParameter(ref)

	assert.Equal(t, "porter-mysql-mysql-password-dep-output", name, "unexpected parameter name")
	assert.False(t, param.Required, "wiring parameters should NOT be required")
	require.NotNil(t, param.Destination, "wiring parameters should have a destination set")
	assert.Equal(t, "PORTER_MYSQL_MYSQL_PASSWORD_DEP_OUTPUT", param.Destination.EnvironmentVariable, "unexpected destination environment variable set")

	assert.Equal(t, "https://getporter.org/generated-bundle/#porter-parameter-source-definition", paramDef.ID, "wiring parameter should have a schema id set")
	assert.Equal(t, cnab.PorterInternal, paramDef.Comment, "wiring parameter should be flagged as internal")
	assert.Empty(t, paramDef.Type, "dependency output types are of unknown types and should not be defined")
}

func TestManifestConverter_generateRequiredExtensions_ParameterSources(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	bun, err := a.ToBundle(ctx)
	require.NoError(t, err, "ToBundle failed")
	assert.Contains(t, bun.RequiredExtensions, "io.cnab.parameter-sources")
}

func TestManifestConverter_generateRequiredExtensions(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	bun, err := a.ToBundle(ctx)
	require.NoError(t, err, "ToBundle failed")

	expected := []string{"sh.porter.file-parameters", "io.cnab.dependencies", "io.cnab.parameter-sources", "io.cnab.docker"}
	assert.Equal(t, expected, bun.RequiredExtensions)
}

func TestManifestConverter_generateCustomExtensions_withRequired(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter-with-required-extensions.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	bun, err := a.ToBundle(ctx)
	require.NoError(t, err, "ToBundle failed")
	assert.Contains(t, bun.Custom, cnab.FileParameterExtensionKey)
	assert.Contains(t, bun.Custom, "requiredExtension1")
	assert.Contains(t, bun.Custom, "requiredExtension2")
	assert.Equal(t, map[string]interface{}{"config": true}, bun.Custom["requiredExtension2"])
}

func TestManifestConverter_GenerateCustomActionDefinitions(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	defs := a.generateCustomActionDefinitions()
	require.Len(t, defs, 3, "expected 3 custom action definitions to be generated")

	require.Contains(t, defs, "status")
	statusDef := defs["status"]
	assert.Equal(t, "Print the installation status", statusDef.Description)
	assert.False(t, statusDef.Stateless, "expected the status custom action to not be stateless")
	assert.False(t, statusDef.Modifies, "expected the status custom action to not modify resources")

	require.Contains(t, defs, "boom")
	boomDef := defs["boom"]
	assert.False(t, boomDef.Stateless, "expected the dry-run custom action to default to not stateless")
	assert.True(t, boomDef.Modifies, "expected the dry-run custom action to default to modifying resources")
}

func TestManifestConverter_generateDefaultAction(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		action     string
		wantAction bundle.Action
	}{
		{"dry-run", bundle.Action{
			Description: "Execute the installation in a dry-run mode, allowing to see what would happen with the given set of parameter values",
			Modifies:    false,
			Stateless:   true,
		}},
		{
			"help", bundle.Action{
				Description: "Print a help message to the standard output",
				Modifies:    false,
				Stateless:   true,
			}},
		{"log", bundle.Action{
			Description: "Print logs of the installed system to the standard output",
			Modifies:    false,
			Stateless:   false,
		}},
		{"status", bundle.Action{
			Description: "Print a human readable status message to the standard output",
			Modifies:    false,
			Stateless:   false,
		}},
		{"zombies", bundle.Action{
			Description: "zombies",
			Modifies:    true,
			Stateless:   false,
		}},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.action, func(t *testing.T) {
			t.Parallel()

			a := ManifestConverter{}
			gotAction := a.generateDefaultAction(tc.action)
			assert.Equal(t, tc.wantAction, gotAction)
		})
	}
}

func TestManifestConverter_generateCustomMetadata(t *testing.T) {
	t.Parallel()

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	bun, err := a.ToBundle(ctx)
	require.NoError(t, err, "ToBundle failed")
	assert.Len(t, bun.Custom, 7)

	f, err := os.CreateTemp("", "")
	require.NoError(t, err, "Failed to create bundle file")
	defer os.Remove(f.Name())

	_, err = bun.WriteTo(f)
	require.NoError(t, err, "Failed to write bundle file")

	expectedCustomMetaData := `"app":{"version":"1.2.3"},"foo":{"test1":true,"test2":1,"test3":"value","test4":["one","two","three"],"test5":{"1":"one","two":"two"}}`
	bundleData, err := os.ReadFile(f.Name())
	require.NoError(t, err, "Failed to read bundle file")

	assert.Contains(t, string(bundleData), expectedCustomMetaData, "Created bundle should be equal to expected bundle ")
}

func TestManifestConverter_generatedMaintainers(t *testing.T) {
	want := []bundle.Maintainer{
		{Name: "John Doe", Email: "john.doe@example.com", URL: "https://example.com/a"},
		{Name: "Jane Doe", Email: "", URL: "https://example.com/b"},
		{Name: "Janine Doe", Email: "janine.doe@example.com", URL: ""},
		{Name: "", Email: "mike.doe@example.com", URL: "https://example.com/c"},
	}

	c := config.NewTestConfig(t)
	c.TestContext.AddTestFileFromRoot("tests/testdata/mybuns/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	a := NewManifestConverter(c.Config, m, nil, nil)

	got := a.generateBundleMaintainers()
	assert.Len(t, got, len(want), "Created bundle should contain desired maintainers")

	for _, wanted := range want {
		gm, err := getMaintainerByName(got, wanted.Name)
		if err != nil {
			t.Errorf("Created bundle should container maintainer '%s'", wanted.Name)
		}
		assert.Equal(t, wanted.Email, gm.Email, "Created bundle should specify email '%s' for maintainer '%s'", wanted.Email, wanted.Name)
		assert.Equal(t, wanted.URL, gm.URL, "Created bundle should specify url '%s' for maintainer '%s'", wanted.URL, wanted.Name)
	}
}

func getMaintainerByName(source []bundle.Maintainer, name string) (bundle.Maintainer, error) {
	for _, m := range source {
		if m.Name == name {
			return m, nil
		}
	}
	return bundle.Maintainer{}, fmt.Errorf("Could not find maintainer with name '%s'", name)
}
