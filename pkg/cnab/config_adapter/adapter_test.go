package configadapter

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/porter/pkg/cnab/extensions"
	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_ToBundle(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}

	bun := a.ToBundle()

	assert.Equal(t, SchemaVersion, bun.SchemaVersion)
	assert.Equal(t, "hello", bun.Name)
	assert.Equal(t, "0.1.0", bun.Version)
	assert.Equal(t, "An example Porter configuration", bun.Description)

	stamp, err := LoadStamp(bun)
	assert.NoError(t, err, "could not load porter's stamp")
	assert.NotNil(t, stamp)

	assert.Contains(t, bun.Parameters.Fields, "porter-debug", "porter-debug parameter was not defined")
	assert.Contains(t, bun.Definitions, "porter-debug", "porter-debug definition was not defined")

	assert.Contains(t, bun.Custom, config.CustomBundleKey, "Dependencies was not populated")
	assert.Contains(t, bun.Custom, extensions.DependenciesKey, "Dependencies was not populated")
}

func makefloat64(v float64) *float64 {
	return &v
}

func TestPorter_generateBundleParametersSchema(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter-with-parameters.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}

	defs := make(definition.Definitions, len(c.Manifest.Parameters))
	_ = a.generateBundleParameters(&defs)

	testcases := []struct {
		propname string
		wantDef  definition.Schema
	}{
		{"ainteger",
			definition.Schema{
				Type:    "integer",
				Default: 1,
				Minimum: makefloat64(0),
				Maximum: makefloat64(10),
			},
		},
		{"anumber",
			definition.Schema{
				Type:             "number",
				Default:          0.5,
				ExclusiveMinimum: makefloat64(0),
				ExclusiveMaximum: makefloat64(1),
			},
		},
		{
			"astringenum",
			definition.Schema{
				Type:    "string",
				Default: "blue",
				Enum:    []interface{}{"blue", "red", "purple", "pink"},
			},
		},
		{
			"astring",
			definition.Schema{
				Type:      "string",
				MinLength: makefloat64(1),
				MaxLength: makefloat64(10),
			},
		},
		{
			"aboolean",
			definition.Schema{
				Type:    "boolean",
				Default: true,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.propname, func(t *testing.T) {
			def, ok := defs[tc.propname]
			require.True(t, ok, "property definition was not generated")

			assert.Equal(t, tc.wantDef, *def)
		})
	}
}

func TestPorter_buildDefaultPorterParameters(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("../../config/testdata/simple.porter.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}

	defs := make(definition.Definitions, len(c.Manifest.Parameters))
	params := a.generateBundleParameters(&defs)

	debugParam, ok := params.Fields["porter-debug"]
	assert.True(t, ok, "porter-debug parameter was not defined")
	assert.Equal(t, "porter-debug", debugParam.Definition)
	assert.Equal(t, "PORTER_DEBUG", debugParam.Destination.EnvironmentVariable)

	debugDef, ok := defs["porter-debug"]
	require.True(t, ok, "porter-debug definition was not defined")
	assert.Equal(t, "boolean", debugDef.Type)
	assert.Equal(t, false, debugDef.Default)
}

func TestPorter_generateImages(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("../../config/testdata/simple.porter.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}

	mappedImage := config.MappedImage{
		Description:   "un petite server",
		Image:         "deislabs/myserver:1.0.0",
		ImageType:     "docker",
		Digest:        "abc123",
		Size:          12,
		MediaType:     "download",
		OriginalImage: "deis/myserver:1.0.0",
		Platform: &config.ImagePlatform{
			OS:           "linux",
			Architecture: "amd64",
		},
	}
	a.Manifest.ImageMap = map[string]config.MappedImage{
		"server": mappedImage,
	}

	images := a.generateBundleImages()

	require.Len(t, images, 1)
	img := images["server"]
	assert.Equal(t, mappedImage.Description, img.Description)
	assert.Equal(t, mappedImage.Image, img.Image)
	assert.Equal(t, mappedImage.ImageType, img.ImageType)
	assert.Equal(t, mappedImage.Digest, img.Digest)
	assert.Equal(t, mappedImage.Size, img.Size)
	assert.Equal(t, mappedImage.MediaType, img.MediaType)
	assert.Equal(t, mappedImage.Platform.OS, img.Platform.OS)
	assert.Equal(t, mappedImage.Platform.Architecture, img.Platform.Architecture)
}

func TestPorter_generateBundleImages_EmptyPlatform(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("../../config/testdata/simple.porter.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}

	mappedImage := config.MappedImage{
		Description: "un petite server",
		Image:       "deislabs/myserver:1.0.0",
		ImageType:   "docker",
		Platform:    nil,
	}
	a.Manifest.ImageMap = map[string]config.MappedImage{
		"server": mappedImage,
	}

	images := a.generateBundleImages()
	require.Len(t, images, 1)
	img := images["server"]
	assert.Nil(t, img.Platform)
}

func TestPorter_generateBundleOutputs(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("../../config/testdata/simple.porter.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}

	outputDefinitions := []config.OutputDefinition{
		{
			Name:        "output1",
			Description: "Description of output1",
			ApplyTo: []string{
				"install",
				"upgrade",
			},
			Schema: config.Schema{
				Type: "string",
			},
		},
		{
			Name:        "output2",
			Description: "Description of output2",
			Schema: config.Schema{
				Type: "boolean",
			},
		},
	}

	a.Manifest.Outputs = outputDefinitions

	defs := make(definition.Definitions, len(a.Manifest.Outputs))
	outputs := a.generateBundleOutputs(&defs)
	require.Len(t, defs, 2)

	wantOutputDefinitions := bundle.OutputsDefinition{
		Fields: map[string]bundle.OutputDefinition{
			"output1": {
				Definition:  "output1",
				Description: "Description of output1",
				ApplyTo: []string{
					"install",
					"upgrade",
				},
				Path: "/cnab/app/outputs/output1",
			},
			"output2": {
				Definition:  "output2",
				Description: "Description of output2",
				Path:        "/cnab/app/outputs/output2",
			},
		},
	}

	require.Equal(t, &wantOutputDefinitions, outputs)

	wantDefinitions := definition.Definitions{
		"output1": &definition.Schema{
			Type: "string",
		},
		"output2": &definition.Schema{
			Type: "boolean",
		},
	}

	require.Equal(t, wantDefinitions, defs)
}

func TestPorter_generateDependencies(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter-with-deps.yaml", config.Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	a := ManifestConverter{
		Context:  c.Context,
		Manifest: c.Manifest,
	}

	deps := a.generateDependencies()

	require.NotNil(t, deps, "Dependencies should not be nil")
	require.Len(t, deps.Requires, 3, "incorrect number of dependencies were generated")

	testcases := []struct {
		name    string
		wantDep extensions.Dependency
	}{
		{"no-version", extensions.Dependency{
			Bundle: "deislabs/azure-mysql:5.7",
		}},
		{"no-ranges", extensions.Dependency{
			Bundle: "deislabs/azure-active-directory",
			Version: &extensions.DependencyVersion{
				AllowPrereleases: true,
			},
		}},
		{"with-ranges", extensions.Dependency{
			Bundle: "deislabs/azure-blob-storage",
			Version: &extensions.DependencyVersion{
				Ranges: []string{
					"1.x - 2",
					"2.1 - 3.x",
				},
			},
		}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var dep *extensions.Dependency
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
