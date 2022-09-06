package configadapter

import (
	"context"
	"fmt"
	"path"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	depsv1 "get.porter.sh/porter/pkg/cnab/dependencies/v1"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
)

// ManifestConverter converts from a porter manifest to a CNAB bundle definition.
type ManifestConverter struct {
	config          *config.Config
	Manifest        *manifest.Manifest
	ImageDigests    map[string]string
	InstalledMixins []mixin.Metadata
}

func NewManifestConverter(
	config *config.Config,
	manifest *manifest.Manifest,
	imageDigests map[string]string,
	mixins []mixin.Metadata,
) *ManifestConverter {
	return &ManifestConverter{
		config:          config,
		Manifest:        manifest,
		ImageDigests:    imageDigests,
		InstalledMixins: mixins,
	}
}

func (c *ManifestConverter) ToBundle(ctx context.Context) (cnab.ExtendedBundle, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	stamp, err := c.GenerateStamp(ctx)
	if err != nil {
		return cnab.ExtendedBundle{}, span.Error(err)
	}

	b := cnab.NewBundle(bundle.Bundle{
		SchemaVersion: cnab.BundleSchemaVersion(),
		Name:          c.Manifest.Name,
		Description:   c.Manifest.Description,
		Version:       c.Manifest.Version,
		Maintainers:   c.generateBundleMaintainers(),
		Custom:        make(map[string]interface{}, 1),
	})
	image := bundle.InvocationImage{
		BaseImage: bundle.BaseImage{
			Image:     c.Manifest.Image,
			ImageType: "docker",
			Digest:    c.ImageDigests[c.Manifest.Image],
		},
	}

	b.Actions = c.generateCustomActionDefinitions()
	b.Definitions = make(definition.Definitions, len(c.Manifest.Parameters)+len(c.Manifest.Outputs)+len(c.Manifest.StateBag))
	b.InvocationImages = []bundle.InvocationImage{image}
	b.Parameters = c.generateBundleParameters(ctx, &b.Definitions)
	b.Outputs = c.generateBundleOutputs(ctx, &b.Definitions)
	b.Credentials = c.generateBundleCredentials()
	b.Images = c.generateBundleImages()
	custom, err := c.generateCustomExtensions(&b)
	if err != nil {
		return cnab.ExtendedBundle{}, err
	}
	b.Custom = custom
	b.RequiredExtensions = c.generateRequiredExtensions(b)

	b.Custom[config.CustomPorterKey] = stamp

	return b, nil
}

func (c *ManifestConverter) generateBundleMaintainers() []bundle.Maintainer {
	m := make([]bundle.Maintainer, len(c.Manifest.Maintainers))
	for i, item := range c.Manifest.Maintainers {
		m[i] = bundle.Maintainer{
			Name:  item.Name,
			Email: item.Email,
			URL:   item.Url,
		}
	}
	return m
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

func (c *ManifestConverter) generateBundleParameters(ctx context.Context, defs *definition.Definitions) map[string]bundle.Parameter {
	log := tracing.LoggerFromContext(ctx)

	params := make(map[string]bundle.Parameter, len(c.Manifest.Parameters))

	addParam := func(param manifest.ParameterDefinition) {
		// Update ApplyTo per parameter definition and manifest
		param.UpdateApplyTo(c.Manifest)

		p := bundle.Parameter{
			Definition:  param.Name,
			ApplyTo:     param.ApplyTo,
			Description: param.Description,
		}

		// If the default is empty, set required to true.
		// State parameters are always optional, and don't have a default
		if param.Default == nil && !param.IsState {
			p.Required = true
		}

		if param.Sensitive {
			param.Schema.WriteOnly = toBool(true)
		}

		if !param.Destination.IsEmpty() {
			p.Destination = &bundle.Location{
				EnvironmentVariable: param.Destination.EnvironmentVariable,
				Path:                manifest.ResolvePath(param.Destination.Path),
			}
		} else {
			p.Destination = &bundle.Location{
				EnvironmentVariable: manifest.ParamToEnvVar(param.Name),
			}
		}

		if param.Type == nil {
			// Default to a file type if the param is stored in a file
			if param.Destination.Path != "" {
				param.Type = "file"
			} else {
				// Assume it's a string otherwise
				param.Type = "string"
			}

			log.Debugf("Defaulting the type of parameter %s to %s", param.Name, param.Type)
		}

		// Create a definition that matches the parameter if one isn't already defined
		if _, ok := (*defs)[param.Name]; !ok {
			kind := "parameter"
			if param.IsState {
				kind = "state"
			}

			defName := c.addDefinition(param.Name, kind, param.Schema, defs)
			p.Definition = defName
		}
		params[param.Name] = p
	}

	for _, p := range c.Manifest.Parameters {
		addParam(p)
	}

	for _, p := range c.buildDefaultPorterParameters() {
		addParam(p)
	}

	return params
}

func (c *ManifestConverter) generateBundleOutputs(ctx context.Context, defs *definition.Definitions) map[string]bundle.Output {
	log := tracing.LoggerFromContext(ctx)

	outputs := make(map[string]bundle.Output, len(c.Manifest.Outputs))

	addOutput := func(output manifest.OutputDefinition) {
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

		if output.Type == nil {
			// Default to a file type if the param is stored in a file
			if output.Path != "" {
				output.Type = "file"
			} else {
				// Assume it's a string otherwise
				output.Type = "string"
			}
			log.Debugf("Defaulting the type of output %s to %s", output.Name, output.Type)
		}

		// Create a definition that matches the output if one isn't already defined
		if _, ok := (*defs)[output.Name]; !ok {
			kind := "output"
			if output.IsState {
				kind = "state"
			}

			defName := c.addDefinition(output.Name, kind, output.Schema, defs)
			o.Definition = defName
		}
		outputs[output.Name] = o
	}

	for _, o := range c.Manifest.Outputs {
		addOutput(o)
	}

	for _, o := range c.buildDefaultPorterOutputs() {
		addOutput(o)
	}

	return outputs
}

func (c *ManifestConverter) addDefinition(name string, kind string, def definition.Schema, defs *definition.Definitions) string {
	defName := name
	if !strings.HasSuffix(name, kind) {
		defName = name + "-" + kind
	}

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
				ID:          "https://getporter.org/generated-bundle/#porter-debug",
				Description: "Print debug information from Porter when executing the bundle",
				Type:        "boolean",
				Default:     false,
				Comment:     cnab.PorterInternal,
			},
		},
		{
			Name:    "porter-state",
			IsState: true,
			Destination: manifest.Location{
				Path: "/porter/state.tgz",
			},
			Schema: definition.Schema{
				ID:              "https://getporter.org/generated-bundle/#porter-state",
				Description:     "Supports persisting state for bundles. Porter internal parameter that should not be set manually.",
				Type:            "string",
				ContentEncoding: "base64",
				Comment:         cnab.PorterInternal,
			},
		},
	}
}

func (c *ManifestConverter) buildDefaultPorterOutputs() []manifest.OutputDefinition {
	return []manifest.OutputDefinition{
		{
			Name:    "porter-state",
			IsState: true,
			Path:    "/cnab/app/outputs/porter-state.tgz",
			Schema: definition.Schema{
				ID:              "https://getporter.org/generated-bundle/#porter-state",
				Description:     "Supports persisting state for bundles. Porter internal parameter that should not be set manually.",
				Type:            "string",
				ContentEncoding: "base64",
				Comment:         cnab.PorterInternal,
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
				Path:                manifest.ResolvePath(cred.Path),
				EnvironmentVariable: cred.EnvironmentVariable,
			},
			ApplyTo: cred.ApplyTo,
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

func (c *ManifestConverter) generateDependencies() (interface{}, string, error) {
	if len(c.Manifest.Dependencies.Requires) == 0 {
		return nil, "", nil
	}

	// Check if they are using v1 of the dependencies spec or v2
	if c.config.IsFeatureEnabled(experimental.FlagDependenciesV2) {
		panic("the dependencies-v2 experimental flag was specified but is not yet implemented")
	}

	deps, err := c.generateDependenciesV1()
	if err != nil {
		return nil, "", err
	}
	return deps, cnab.DependenciesV1ExtensionKey, nil
}

func (c *ManifestConverter) generateDependenciesV1() (*depsv1.Dependencies, error) {
	if len(c.Manifest.Dependencies.Requires) == 0 {
		return nil, nil
	}

	deps := &depsv1.Dependencies{
		Sequence: make([]string, 0, len(c.Manifest.Dependencies.Requires)),
		Requires: make(map[string]depsv1.Dependency, len(c.Manifest.Dependencies.Requires)),
	}

	for _, dep := range c.Manifest.Dependencies.Requires {
		dependencyRef := depsv1.Dependency{
			Name:   dep.Name,
			Bundle: dep.Bundle.Reference,
		}
		if len(dep.Bundle.Version) > 0 {
			dependencyRef.Version = &depsv1.DependencyVersion{
				Ranges: []string{dep.Bundle.Version},
			}

			// If we can detect that prereleases are used in the version, then set AllowPrereleases to true
			v, err := semver.NewVersion(dep.Bundle.Version)
			if err == nil {
				dependencyRef.Version.AllowPrereleases = v.Prerelease() != ""
			}
		}
		deps.Sequence = append(deps.Sequence, dep.Name)
		deps.Requires[dep.Name] = dependencyRef
	}

	return deps, nil
}

func (c *ManifestConverter) generateParameterSources(b *cnab.ExtendedBundle) cnab.ParameterSources {
	ps := cnab.ParameterSources{}

	// Parameter sources come from three places
	// 1. indirectly from our template wiring
	// 2. indirectly from state variables
	// 3. directly when they use `source` on a parameter

	// Directly wired outputs to parameters
	for _, p := range c.Manifest.Parameters {
		// Skip parameters that aren't set from an output
		if p.Source.Output == "" {
			continue
		}

		var pso cnab.ParameterSource
		if p.Source.Dependency == "" {
			pso = c.generateOutputParameterSource(p.Source.Output)
		} else {
			ref := manifest.DependencyOutputReference{
				Dependency: p.Source.Dependency,
				Output:     p.Source.Output,
			}
			pso = c.generateDependencyOutputParameterSource(ref)
		}
		ps[p.Name] = pso
	}

	// Directly wired state variables
	// All state variables are persisted in a single file, porter-state.tgz
	ps["porter-state"] = c.generateOutputParameterSource("porter-state")

	// bundle.outputs.OUTPUT
	for _, outputDef := range c.Manifest.GetTemplatedOutputs() {
		wiringName, p, def := c.generateOutputWiringParameter(*b, outputDef.Name)
		if b.Parameters == nil {
			b.Parameters = make(map[string]bundle.Parameter, 1)
		}
		b.Parameters[wiringName] = p
		b.Definitions[wiringName] = &def

		pso := c.generateOutputParameterSource(outputDef.Name)
		ps[wiringName] = pso
	}

	// bundle.dependencies.DEP.outputs.OUTPUT
	for _, ref := range c.Manifest.GetTemplatedDependencyOutputs() {
		wiringName, p, def := c.generateDependencyOutputWiringParameter(ref)
		if b.Parameters == nil {
			b.Parameters = make(map[string]bundle.Parameter, 1)
		}
		b.Parameters[wiringName] = p
		b.Definitions[wiringName] = &def

		pso := c.generateDependencyOutputParameterSource(ref)
		ps[wiringName] = pso
	}

	return ps
}

// generateOutputWiringParameter creates an internal parameter used only by porter, it won't be visible to the user.
// The parameter exists solely so that Porter can inject an output back into the bundle, using a parameter source.
// The parameter's definition is a copy of the output's definition, with the ID set so we know that it was generated by porter.
func (c *ManifestConverter) generateOutputWiringParameter(b cnab.ExtendedBundle, outputName string) (string, bundle.Parameter, definition.Schema) {
	wiringName := manifest.GetParameterSourceForOutput(outputName)

	paramDesc := fmt.Sprintf("Wires up the %s output for use as a parameter. Porter internal parameter that should not be set manually.", outputName)
	wiringParam := c.generateWiringParameter(wiringName, paramDesc)

	// Copy the output definition for use with the wiring parameter
	// and identify the definition as a porter internal structure
	outputDefName := b.Outputs[outputName].Definition
	outputDef := b.Definitions[outputDefName]
	wiringDef := *outputDef
	wiringDef.ID = "https://getporter.org/generated-bundle/#porter-parameter-source-definition"
	wiringDef.Comment = cnab.PorterInternal

	return wiringName, wiringParam, wiringDef
}

// generateDependencyOutputWiringParameter creates an internal parameter used only by porter, it won't be visible to
// the user. The parameter exists solely so that Porter can inject a dependency output into the bundle.
func (c *ManifestConverter) generateDependencyOutputWiringParameter(reference manifest.DependencyOutputReference) (string, bundle.Parameter, definition.Schema) {
	wiringName := manifest.GetParameterSourceForDependency(reference)

	paramDesc := fmt.Sprintf("Wires up the %s dependency %s output for use as a parameter. Porter internal parameter that should not be set manually.", reference.Dependency, reference.Output)
	wiringParam := c.generateWiringParameter(wiringName, paramDesc)

	wiringDef := definition.Schema{
		ID:      "https://getporter.org/generated-bundle/#porter-parameter-source-definition",
		Comment: cnab.PorterInternal,
		// any type, the dependency's bundle definition is not available at buildtime
	}

	return wiringName, wiringParam, wiringDef
}

// generateWiringParameter builds an internal Porter-only parameter for connecting a parameter source to a parameter.
func (g *ManifestConverter) generateWiringParameter(wiringName string, description string) bundle.Parameter {
	return bundle.Parameter{
		Definition:  wiringName,
		Description: description,
		Required:    false,
		Destination: &bundle.Location{
			EnvironmentVariable: manifest.ParamToEnvVar(wiringName),
		},
	}
}

// generateOutputParameterSource builds a parameter source that connects a bundle output to a parameter.
func (c *ManifestConverter) generateOutputParameterSource(outputName string) cnab.ParameterSource {
	return cnab.ParameterSource{
		Priority: []string{cnab.ParameterSourceTypeOutput},
		Sources: map[string]cnab.ParameterSourceDefinition{
			cnab.ParameterSourceTypeOutput: cnab.OutputParameterSource{
				OutputName: outputName,
			},
		},
	}
}

// generateDependencyOutputParameterSource builds a parameter source that connects a dependency output to a parameter.
func (c *ManifestConverter) generateDependencyOutputParameterSource(ref manifest.DependencyOutputReference) cnab.ParameterSource {
	return cnab.ParameterSource{
		Priority: []string{cnab.ParameterSourceTypeDependencyOutput},
		Sources: map[string]cnab.ParameterSourceDefinition{
			cnab.ParameterSourceTypeDependencyOutput: cnab.DependencyOutputParameterSource{
				Dependency: ref.Dependency,
				OutputName: ref.Output,
			},
		},
	}
}

func toBool(value bool) *bool {
	return &value
}

func toInt(v int) *int {
	return &v
}

func toFloat(v float64) *float64 {
	return &v
}

func (c *ManifestConverter) generateCustomExtensions(b *cnab.ExtendedBundle) (map[string]interface{}, error) {
	customExtensions := map[string]interface{}{
		cnab.FileParameterExtensionKey: struct{}{},
	}

	// Add custom metadata defined in the manifest
	for key, value := range c.Manifest.Custom {
		customExtensions[key] = value
	}

	// Add the dependency extension
	deps, depsExtKey, err := c.generateDependencies()
	if err != nil {
		return nil, err
	}
	if depsExtKey != "" {
		customExtensions[depsExtKey] = deps
	}

	// Add the parameter sources extension
	ps := c.generateParameterSources(b)
	if len(ps) > 0 {
		customExtensions[cnab.ParameterSourcesExtensionKey] = ps
	}

	// Add entries for user-specified required extensions, like docker
	for _, ext := range c.Manifest.Required {
		customExtensions[lookupExtensionKey(ext.Name)] = ext.Config
	}

	return customExtensions, nil
}

func (c *ManifestConverter) generateRequiredExtensions(b cnab.ExtendedBundle) []string {
	requiredExtensions := []string{cnab.FileParameterExtensionKey}

	// Add the appropriate dependencies key if applicable
	if b.HasDependenciesV1() {
		requiredExtensions = append(requiredExtensions, cnab.DependenciesV1ExtensionKey)
	}

	// Add the appropriate parameter sources key if applicable
	if b.HasParameterSources() {
		requiredExtensions = append(requiredExtensions, cnab.ParameterSourcesExtensionKey)
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

	supportedExt, err := cnab.GetSupportedExtension(name)
	if err != nil {
		// TODO: Issue linter warning
	} else {
		key = supportedExt.Key
	}
	return key
}
