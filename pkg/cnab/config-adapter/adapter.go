package configadapter

import (
	"fmt"
	"path"
	"regexp"
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

func (c *ManifestConverter) ToBundle() (bundle.Bundle, error) {
	stamp, err := c.GenerateStamp()
	if err != nil {
		return bundle.Bundle{}, err
	}

	b := bundle.Bundle{
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
	b.Custom = c.generateCustomExtensions(&b)
	b.RequiredExtensions = c.generateRequiredExtensions(b)

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
				EnvironmentVariable: ParamToEnvVar(param.Name),
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
				ID:          "https://porter.sh/schema/bundle.json#porter-debug",
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

func (c *ManifestConverter) generateDependencies() extensions.Dependencies {
	deps := extensions.Dependencies{
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

func (c *ManifestConverter) generateParameterSources(b *bundle.Bundle) extensions.ParameterSources {
	ps := extensions.ParameterSources{}

	// Parameter sources come from two places, indirectly from our template wiring
	// and directly when they use `source` on a parameter

	for _, p := range c.Manifest.Parameters {
		if p.Source.Output == "" {
			continue
		}

		pso := c.generateParameterSource(p.Source.Output)
		ps[p.Name] = pso
	}

	for _, v := range c.Manifest.TemplateVariables {
		outputName, ok := c.getTemplateOutputName(v)
		if !ok {
			continue
		}

		// Check if a bundle level output is defined, could be step level
		if _, ok := b.Outputs[outputName]; !ok {
			continue
		}

		wiringName, p, def := c.generateOutputWiringParameter(*b, outputName)
		if b.Parameters == nil {
			b.Parameters = make(map[string]bundle.Parameter, 1)
		}
		b.Parameters[wiringName] = p
		b.Definitions[wiringName] = &def

		pso := c.generateParameterSource(outputName)
		ps[wiringName] = pso
	}

	return ps
}

func (c *ManifestConverter) generateOutputWiringParameter(b bundle.Bundle, outputName string) (string, bundle.Parameter, definition.Schema) {
	wiringName := fmt.Sprintf("porter-%s-output", outputName)

	wiringParam := bundle.Parameter{
		Definition:  wiringName,
		Description: fmt.Sprintf("Wires up the %s output for use as a parameter. Porter internal parameter that should not be set manually.", outputName),
		Required:    false,
		Destination: &bundle.Location{
			EnvironmentVariable: fmt.Sprintf("PORTER_%s_OUTPUT", ParamToEnvVar(outputName)),
		},
	}

	// Copy the output definition for use with the wiring parameter
	// and identify the definition as a porter internal structure
	outputDefName := b.Outputs[outputName].Definition
	outputDef := b.Definitions[outputDefName]
	var wiringDef definition.Schema
	wiringDef = *outputDef
	wiringDef.ID = "https://porter.sh/schema/bundle.json#porter-parameter-source-definition"

	return wiringName, wiringParam, wiringDef
}

func (c *ManifestConverter) generateParameterSource(outputName string) extensions.ParameterSource {
	return extensions.ParameterSource{
		Priority: []string{extensions.ParameterSourceTypeOutput},
		Sources: map[string]extensions.ParameterSourceDefinition{
			extensions.ParameterSourceTypeOutput: extensions.OutputParameterSource{
				OutputName: outputName,
			},
		},
	}
}

var outputNameRegex = regexp.MustCompile(`^bundle\.outputs\.(.+)$`)

func (c *ManifestConverter) getTemplateOutputName(value string) (string, bool) {
	matches := outputNameRegex.FindStringSubmatch(value)
	if len(matches) < 2 {
		return "", false
	}

	outputName := matches[1]
	return outputName, true
}

func toBool(value bool) *bool {
	return &value
}

func toInt(v int) *int {
	return &v
}

func (c *ManifestConverter) generateCustomExtensions(b *bundle.Bundle) map[string]interface{} {
	customExtensions := map[string]interface{}{}

	// Add custom metadata defined in the manifest
	for key, value := range c.Manifest.Custom {
		customExtensions[key] = value
	}

	// Add the dependency extension
	deps := c.generateDependencies()
	if len(deps.Requires) > 0 {
		customExtensions[extensions.DependenciesKey] = deps
	}

	// Add the parameter sources extension
	ps := c.generateParameterSources(b)
	if len(ps) > 0 {
		customExtensions[extensions.ParameterSourcesKey] = ps
	}

	// Add entries for user-specified required extensions, like docker
	for _, ext := range c.Manifest.Required {
		customExtensions[lookupExtensionKey(ext.Name)] = ext.Config
	}

	return customExtensions
}

func (c *ManifestConverter) generateRequiredExtensions(b bundle.Bundle) []string {
	var requiredExtensions []string

	// Add the appropriate dependencies key if applicable
	if extensions.HasDependencies(b) {
		requiredExtensions = append(requiredExtensions, extensions.DependenciesKey)
	}

	if extensions.HasParameterSources(b) {
		requiredExtensions = append(requiredExtensions, extensions.ParameterSourcesKey)
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
	supportedExt, err := extensions.GetSupportedExtension(name)
	if err != nil {
		// TODO: Issue linter warning
	} else {
		key = supportedExt.Key
	}
	return key
}

// Convert a parameter name to an environment variable.
// Anything more complicated should define the variable explicitly.
func ParamToEnvVar(name string) string {
	name = strings.ToUpper(name)
	fixer := strings.NewReplacer("-", "_", ".", "_")
	return fixer.Replace(name)
}
