package configadapter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/porter/pkg/cnab/extensions"
	"github.com/deislabs/porter/pkg/config"
)

const SchemaVersion = "v1.0.0-WD"

// ManifestConverter converts from a porter manifest to a CNAB bundle definition.
type ManifestConverter struct {
	*config.Config
	ImageDigests map[string]string
}

func NewManifestConverter(cfg *config.Config, imageDigests map[string]string) *ManifestConverter {
	return &ManifestConverter{
		Config:       cfg,
		ImageDigests: imageDigests,
	}
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
			Digest:    c.ImageDigests[c.Manifest.Image],
		},
	}

	b.Actions = c.generateCustomActionDefinitions()
	b.Definitions = make(definition.Definitions, len(c.Manifest.Parameters)+len(c.Manifest.Outputs))
	b.InvocationImages = []bundle.InvocationImage{image}
	b.Parameters = c.generateBundleParameters(&b.Definitions)
	b.Outputs = c.generateBundleOutputs(&b.Definitions)
	b.Credentials = c.generateBundleCredentials()
	b.Images = c.generateBundleImages()
	b.Custom[config.CustomBundleKey] = c.GenerateStamp()

	b.Custom[extensions.DependenciesKey] = c.generateDependencies()
	if len(c.Manifest.Dependencies) > 0 {
		b.RequiredExtensions = []string{extensions.DependenciesKey}
	}

	return b
}

func (c *ManifestConverter) generateCustomActionDefinitions() map[string]bundle.Action {
	if len(c.Manifest.CustomActions) == 0 {
		return nil
	}

	defs := make(map[string]bundle.Action, len(c.Manifest.CustomActions))
	for action, def := range c.Manifest.CustomActionDefinitions {
		def := bundle.Action{
			Description: def.Description,
			Modifies:    def.ModifiesResources,
			Stateless:   def.Stateless,
		}
		defs[action] = def
	}

	// If they used a custom action but didn't define it, default it to a safe action definition
	for action := range c.Manifest.CustomActions {
		if _, ok := c.Manifest.CustomActionDefinitions[action]; !ok {
			defs[action] = c.generateDefaultAction(action)
		}
	}

	return defs
}

func (c *ManifestConverter) generateDefaultAction(action string) bundle.Action {
	// See https://github.com/deislabs/cnab-spec/blob/master/804-well-known-custom-actions.md
	switch action {
	case "dry-run", "io.cnab.dry-run":
		return bundle.Action{
			Description: "Execute the installation in a dry-run mode, allowing to see what would happen with the given set of parameter values",
			Modifies:    false,
			Stateless:   true,
		}
	case "help", "io.cnab.help":
		return bundle.Action{
			Description: "Print an help message to the standard output",
			Modifies:    false,
			Stateless:   true,
		}
	case "log", "io.cnab.log":
		return bundle.Action{
			Description: "Print logs of the installed system to the standard output",
			Modifies:    false,
			Stateless:   false,
		}
	case "status", "io.cnab.status":
		return bundle.Action{
			Description: "Print a human readable status message to the standard output",
			Modifies:    false,
			Stateless:   false,
		}
	case "status+json", "io.cnab.status+json":
		return bundle.Action{
			Description: "Print a json payload describing the detailed status with the following the CNAB status schema",
			Modifies:    false,
			Stateless:   false,
		}
	default:
		// By default assume that any custom action could modify state
		return bundle.Action{
			Description: action,
			Modifies:    true,
			Stateless:   false,
		}
	}
}

func (c *ManifestConverter) generateBundleParameters(defs *definition.Definitions) map[string]bundle.Parameter {
	params := make(map[string]bundle.Parameter, len(c.Manifest.Parameters))

	for _, param := range append(c.Manifest.Parameters, c.buildDefaultPorterParameters()...) {
		fmt.Fprintf(c.Out, "Generating parameter definition %s ====>\n", param.Name)
		p := bundle.Parameter{
			Definition:  param.Name,
			ApplyTo:     param.ApplyTo,
			Description: param.Description,
		}

		// If the default is empty, set required to true.
		if param.Default == nil {
			p.Required = true
		}

		if param.Sensitive {
			param.Schema.WriteOnly = toBool(true)
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

		// Only set definition if it doesn't already exist
		// (Both Params and Outputs may reference same Definition)
		if _, exists := (*defs)[param.Name]; !exists {
			def := param.Schema
			(*defs)[param.Name] = &def
		}
		params[param.Name] = p
	}
	return params
}

func (c *ManifestConverter) generateBundleOutputs(defs *definition.Definitions) map[string]bundle.Output {
	var outputs map[string]bundle.Output

	if len(c.Manifest.Outputs) > 0 {
		outputs = make(map[string]bundle.Output, len(c.Manifest.Outputs))

		for _, output := range c.Manifest.Outputs {
			fmt.Fprintf(c.Out, "Generating output definition %s ====>\n", output.Name)
			o := bundle.Output{
				Definition:  output.Name,
				Description: output.Description,
				ApplyTo:     output.ApplyTo,
				Path:        filepath.Join(config.BundleOutputsDir, output.Name),
			}

			if output.Sensitive {
				output.Schema.WriteOnly = toBool(true)
			}

			// Only set definition if it doesn't already exist
			// (Both Params and Outputs may reference same Definition)
			if _, exists := (*defs)[output.Name]; !exists {
				def := output.Schema
				(*defs)[output.Name] = &def
			}
			outputs[output.Name] = o
		}
	}
	return outputs
}

func (c *ManifestConverter) buildDefaultPorterParameters() []config.ParameterDefinition {
	return []config.ParameterDefinition{
		{
			Name: "porter-debug",
			Destination: &config.Location{
				EnvironmentVariable: "PORTER_DEBUG",
			},
			Schema: definition.Schema{
				Description: "Print debug information from Porter when executing the bundle",
				Type:        "boolean",
				Default:     false,
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
				Labels:    refImage.Labels,
			},
		}
		images[i] = img
	}

	return images
}

func (c *ManifestConverter) generateDependencies() *extensions.Dependencies {
	if len(c.Manifest.Dependencies) == 0 {
		return nil
	}

	deps := &extensions.Dependencies{
		Requires: make(map[string]extensions.Dependency, len(c.Manifest.Dependencies)),
	}

	for name, dep := range c.Manifest.Dependencies {
		r := extensions.Dependency{
			Bundle: dep.Tag,
		}
		if len(dep.Versions) > 0 || dep.AllowPrereleases {
			r.Version = &extensions.DependencyVersion{
				AllowPrereleases: dep.AllowPrereleases,
			}
			if len(dep.Versions) > 0 {
				r.Version.Ranges = make([]string, len(dep.Versions))
				copy(r.Version.Ranges, dep.Versions)
			}
		}
		deps.Requires[name] = r
	}

	return deps
}

func toBool(value bool) *bool {
	return &value
}

func toInt(v int) *int {
	return &v
}
