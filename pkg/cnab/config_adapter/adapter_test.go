package configadapter

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle/definition"

	"github.com/deislabs/porter/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_ToBundle(t *testing.T) {
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

	defs, _ := a.generateBundleParameters()

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

	defs, params := a.generateBundleParameters()

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
