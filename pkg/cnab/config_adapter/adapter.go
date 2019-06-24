package configadapter

import (
	"fmt"
	"strings"

	"github.com/deislabs/cnab-go/bundle/definition"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/context"
)

const SchemaVersion = "v1.0.0-WD"

// ManifestConverter converts from a porter manifest to a CNAB bundle definition.
type ManifestConverter struct {
	*context.Context
	Manifest *config.Manifest
}

func (c *ManifestConverter) ToBundle() *bundle.Bundle {
	fmt.Fprintf(c.Out, "\nGenerating Bundle File with Invocation Image %s =======> \n", c.Manifest.Image)
	b := &bundle.Bundle{
		SchemaVersion: SchemaVersion,
		Name:          c.Manifest.Name,
		Description:   c.Manifest.Description,
		Version:       c.Manifest.Version,
		Custom:        make(map[string]interface{}, 1),
	}
	image := bundle.InvocationImage{
		BaseImage: bundle.BaseImage{
			Image:     c.Manifest.Image,
			ImageType: "docker",
		},
	}

	b.InvocationImages = []bundle.InvocationImage{image}
	b.Definitions, b.Parameters = c.generateBundleParameters()
	b.Credentials = c.generateBundleCredentials()
	b.Images = c.generateBundleImages()
	b.Custom[config.CustomBundleKey] = c.GenerateStamp()

	return b
}

func (c *ManifestConverter) generateBundleParameters() (definition.Definitions, *bundle.ParametersDefinition) {
	defs := make(definition.Definitions, len(c.Manifest.Parameters))
	params := &bundle.ParametersDefinition{
		Fields: make(map[string]bundle.ParameterDefinition, len(c.Manifest.Parameters)),
	}
	for _, param := range append(c.Manifest.Parameters, c.buildDefaultPorterParameters()...) {
		fmt.Fprintf(c.Out, "Generating parameter definition %s ====>\n", param.Name)
		d := &definition.Schema{
			Type:      param.DataType,
			Default:   param.Default,
			Enum:      param.AllowedValues,
			Minimum:   param.MinValue,
			Maximum:   param.MaxValue,
			MinLength: param.MinLength,
			MaxLength: param.MaxLength,
		}
		p := bundle.ParameterDefinition{
			Definition:  param.Name,
			Description: param.Description,
		}

		// If the default is empty, set required to true.
		if param.Default == nil {
			params.Required = append(params.Required, param.Name)
		}

		if param.Destination != nil {
			p.Destination = &bundle.Location{
				EnvironmentVariable: param.Destination.EnvironmentVariable,
				Path:                param.Destination.Path,
			}
		} else {
			p.Destination = &bundle.Location{
				EnvironmentVariable: strings.ToUpper(param.Name),
			}
		}

		defs[param.Name] = d
		params.Fields[param.Name] = p
	}
	return defs, params
}

func (c *ManifestConverter) buildDefaultPorterParameters() []config.ParameterDefinition {
	return []config.ParameterDefinition{
		{
			Name:        "porter-debug",
			Description: "Print debug information from Porter when executing the bundle",
			Destination: &config.Location{
				EnvironmentVariable: "PORTER_DEBUG",
			},
			Schema: config.Schema{
				DataType: "boolean",
				Default:  false,
			},
		},
	}
}

func (c *ManifestConverter) generateBundleCredentials() map[string]bundle.Credential {
	params := map[string]bundle.Credential{}
	for _, cred := range c.Manifest.Credentials {
		fmt.Fprintf(c.Out, "Generating credential %s ====>\n", cred.Name)
		l := bundle.Credential{
			Description: cred.Description,
			Required:    cred.Required,
			Location: bundle.Location{
				Path:                cred.Path,
				EnvironmentVariable: cred.EnvironmentVariable,
			},
		}
		params[cred.Name] = l
	}
	return params
}

func (c *ManifestConverter) generateBundleImages() map[string]bundle.Image {
	images := make(map[string]bundle.Image, len(c.Manifest.ImageMap))

	for i, refImage := range c.Manifest.ImageMap {
		img := bundle.Image{
			Description: refImage.Description,
			BaseImage: bundle.BaseImage{
				Image:     refImage.Image,
				Digest:    refImage.Digest,
				ImageType: refImage.ImageType,
				MediaType: refImage.MediaType,
				Size:      refImage.Size,
			},
		}
		if refImage.Platform != nil {
			img.Platform = &bundle.ImagePlatform{
				Architecture: refImage.Platform.Architecture,
				OS:           refImage.Platform.OS,
			}
		}
		images[i] = img
	}

	return images
}
