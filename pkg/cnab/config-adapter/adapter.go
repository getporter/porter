package configadapter

import (
	"fmt"
	"path"
	"strings"

	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
)

const SchemaVersion = "v1.0.0"

// ManifestConverter converts from a porter manifest to a CNAB bundle definition.
type ManifestConverter struct {
	*context.Context
	Manifest     *manifest.Manifest
	ImageDigests map[string]string
	Mixins       []mixin.Metadata
}

func NewManifestConverter(cxt *context.Context, manifest *manifest.Manifest, imageDigests map[string]string, mixins []mixin.Metadata) *ManifestConverter {
	return &ManifestConverter{
		Context:      cxt,
		Manifest:     manifest,
		ImageDigests: imageDigests,
		Mixins:       mixins,
	}
}

func (c *ManifestConverter) ToBundle() (*bundle.Bundle, error) {
	stamp, err := c.GenerateStamp()
	if err != nil {
		return nil, err
	}

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
	b.RequiredExtensions = c.generateRequiredExtensions()
	b.Custom = c.generateCustomExtensions()

	b.Custom[config.CustomPorterKey] = stamp

	return b, nil
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
	// See https://github.com/cnabio/cnab-spec/blob/master/804-well-known-custom-actions.md
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

		if !param.Destination.IsEmpty() {
			p.Destination = &bundle.Location{
				EnvironmentVariable: param.Destination.EnvironmentVariable,
				Path:                param.Destination.Path,
			}
		} else {
			p.Destination = &bundle.Location{
				EnvironmentVariable: strings.ToUpper(param.Name),
			}
		}

		defName := c.addDefinition(param.Name, "parameter", param.Schema, defs)
		p.Definition = defName
		params[param.Name] = p
	}
	return params
}

func (c *ManifestConverter) generateBundleOutputs(defs *definition.Definitions) map[string]bundle.Output {
	var outputs map[string]bundle.Output

	if len(c.Manifest.Outputs) > 0 {
		outputs = make(map[string]bundle.Output, len(c.Manifest.Outputs))

		for _, output := range c.Manifest.Outputs {
			o := bundle.Output{
				Definition:  output.Name,
				Description: output.Description,
				ApplyTo:     output.ApplyTo,
				// must be a standard Unix path as this will be inside of the container
				// (only linux containers supported currently)
				Path: path.Join(config.BundleOutputsDir, output.Name),
			}

			if output.Sensitive {
				output.Schema.WriteOnly = toBool(true)
			}

			defName := c.addDefinition(output.Name, "output", output.Schema, defs)
			o.Definition = defName
			outputs[output.Name] = o
		}
	}
	return outputs
}

func (c *ManifestConverter) addDefinition(name string, kind string, def definition.Schema, defs *definition.Definitions) string {
	defName := name + "-" + kind

	// file is a porter specific type, swap it out for something CNAB understands
	if def.Type == "file" {
		def.Type = "string"
		def.ContentEncoding = "base64"
	}

	(*defs)[defName] = &def

	return defName
}

func (c *ManifestConverter) buildDefaultPorterParameters() []manifest.ParameterDefinition {
	return []manifest.ParameterDefinition{
		{
			Name: "porter-debug",
			Destination: manifest.Location{
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
		imgRefStr := refImage.Repository
		if refImage.Digest != "" {
			imgRefStr = fmt.Sprintf("%s@%s", imgRefStr, refImage.Digest)
		} else if refImage.Tag != "" {
			imgRefStr = fmt.Sprintf("%s:%s", imgRefStr, refImage.Tag)
		} else { // default to `latest` if no tag is provided
			imgRefStr = fmt.Sprintf("%s:latest", imgRefStr)
		}
		imgType := refImage.ImageType
		if imgType == "" {
			imgType = "docker"
		}
		img := bundle.Image{
			Description: refImage.Description,
			BaseImage: bundle.BaseImage{
				Image:     imgRefStr,
				Digest:    refImage.Digest,
				ImageType: imgType,
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

func (c *ManifestConverter) generateCustomExtensions() map[string]interface{} {
	customExtensions := map[string]interface{}{}

	for key, value := range c.Manifest.Custom {
		customExtensions[key] = value
	}

	deps := c.generateDependencies()
	if deps != nil && len(deps.Requires) > 0 {
		customExtensions[extensions.DependenciesKey] = deps
	}

	// Add entries for each required extension
	for _, ext := range c.Manifest.Required {
		customExtensions[lookupExtensionKey(ext.Name)] = ext.Config
	}

	return customExtensions
}

func (c *ManifestConverter) generateRequiredExtensions() []string {
	requiredExtensions := []string{}

	// Add the appropriate dependencies key if applicable
	if len(c.Manifest.Dependencies) > 0 {
		requiredExtensions = append(requiredExtensions, extensions.DependenciesKey)
	}

	// Add all under required section of manifest
	for _, ext := range c.Manifest.Required {
		requiredExtensions = append(requiredExtensions, lookupExtensionKey(ext.Name))
	}

	return requiredExtensions
}

// lookupExtensionKey is a helper method to return a full key matching a
// supported extension, if applicable
func lookupExtensionKey(name string) string {
	key := name
	// If an official supported extension, we grab the full key
	// TODO: Do we want to error out here if unsupported?
	if supportedExt, _ := extensions.GetSupportedExtension(name); supportedExt != nil {
		key = supportedExt.Key
	}
	return key
}
