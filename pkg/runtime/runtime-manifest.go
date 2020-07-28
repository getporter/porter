package runtime

import (
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/cbroglie/mustache"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type RuntimeManifest struct {
	*context.Context
	*manifest.Manifest

	Action string

	// bundles is map of the dependencies bundle definitions, keyed by the alias used in the root manifest
	bundles map[string]bundle.Bundle

	steps           manifest.Steps
	outputs         map[string]string
	sensitiveValues []string
}

func NewRuntimeManifest(cxt *context.Context, action string, manifest *manifest.Manifest) *RuntimeManifest {
	return &RuntimeManifest{
		Context:  cxt,
		Action:   action,
		Manifest: manifest,
	}
}

func (m *RuntimeManifest) Validate() error {
	err := m.loadDependencyDefinitions()
	if err != nil {
		return err
	}

	err = m.setStepsByAction()
	if err != nil {
		return err
	}

	err = m.steps.Validate(m.Manifest)
	if err != nil {
		return errors.Wrap(err, "invalid action configuration")
	}

	return nil
}

func (m *RuntimeManifest) loadDependencyDefinitions() error {
	m.bundles = make(map[string]bundle.Bundle, len(m.Dependencies.Elements))
	for alias := range m.Dependencies.Elements {
		bunD, err := GetDependencyDefinition(m.Context, alias)
		if err != nil {
			return err
		}

		bun, err := bundle.Unmarshal(bunD)
		if err != nil {
			return errors.Wrapf(err, "error unmarshaling bundle definition for dependency %s", alias)
		}

		m.bundles[alias] = *bun
	}

	return nil
}

func resolveParameter(pd manifest.ParameterDefinition) string {
	if pd.Destination.EnvironmentVariable != "" {
		return os.Getenv(pd.Destination.EnvironmentVariable)
	}
	if pd.Destination.Path != "" {
		return pd.Destination.Path
	}
	envVar := manifest.ParamToEnvVar(pd.Name)
	return os.Getenv(envVar)
}

func resolveCredential(cd manifest.CredentialDefinition) (string, error) {
	if cd.EnvironmentVariable != "" {
		return os.Getenv(cd.EnvironmentVariable), nil
	} else if cd.Path != "" {
		return cd.Path, nil
	} else {
		return "", fmt.Errorf("credential: %s is malformed", cd.Name)
	}
}

func (m *RuntimeManifest) resolveBundleOutput(def manifest.OutputDefinition) (string, error) {
	// Get the output's value from the injected parameter source
	psParamEnv := manifest.GetParameterSourceEnvVar(def.Name)
	outputValue, ok := os.LookupEnv(psParamEnv)
	if !ok {
		return "", errors.Errorf("No parameter source was injected for output %s", def.Name)
	}
	return outputValue, nil
}

func (m *RuntimeManifest) GetSensitiveValues() []string {
	if m.sensitiveValues == nil {
		return []string{}
	}
	return m.sensitiveValues
}

func (m *RuntimeManifest) setSensitiveValue(val string) {
	exists := false
	for _, item := range m.sensitiveValues {
		if item == val {
			exists = true
		}
	}

	if !exists {
		m.sensitiveValues = append(m.sensitiveValues, val)
	}
}

func (m *RuntimeManifest) GetSteps() manifest.Steps {
	return m.steps
}

func (m *RuntimeManifest) GetOutputs() map[string]string {
	outputs := make(map[string]string, len(m.outputs))

	for k, v := range m.outputs {
		outputs[k] = v
	}

	return outputs
}

func (m *RuntimeManifest) setStepsByAction() error {
	switch m.Action {
	case claim.ActionInstall:
		m.steps = m.Install
	case claim.ActionUninstall:
		m.steps = m.Uninstall
	case claim.ActionUpgrade:
		m.steps = m.Upgrade
	default:
		customAction, ok := m.CustomActions[string(m.Action)]
		if !ok {
			actions := make([]string, 0, len(m.CustomActions))
			for a := range m.CustomActions {
				actions = append(actions, a)
			}
			errors.Errorf("unsupported action %q, custom actions are defined for: %s", m.Action, strings.Join(actions, ", "))
		}
		m.steps = customAction
	}

	return nil
}

func (m *RuntimeManifest) ApplyStepOutputs(assignments map[string]string) error {
	if m.outputs == nil {
		m.outputs = map[string]string{}
	}

	for outvar, outval := range assignments {
		m.outputs[outvar] = outval
	}
	return nil
}

type StepOutput struct {
	// The final value of the output returned by the mixin after executing
	value string

	Name string                 `yaml:"name"`
	Data map[string]interface{} `yaml:",inline"`
}

func (m *RuntimeManifest) buildSourceData() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	m.sensitiveValues = []string{}
	bun := make(map[string]interface{})
	data["bundle"] = bun

	// Enable interpolation of manifest/bundle name via bundle.name
	bun["name"] = m.Name
	bun["version"] = m.Version
	bun["description"] = m.Description
	bun["invocationImage"] = m.Image

	params := make(map[string]interface{})
	bun["parameters"] = params
	for _, param := range m.Parameters {
		if !param.AppliesTo(m.Action) {
			continue
		}

		pe := param.Name
		val := resolveParameter(param)
		if param.Sensitive {
			m.setSensitiveValue(val)
		}
		params[pe] = val
	}

	creds := make(map[string]interface{})
	bun["credentials"] = creds
	for _, cred := range m.Credentials {
		pe := cred.Name
		val, err := resolveCredential(cred)
		if err != nil {
			return nil, err
		}
		m.setSensitiveValue(val)
		creds[pe] = val
	}

	bun["outputs"] = m.outputs
	// Iterate through the runtime manifest's step outputs and mask by default
	for _, stepOutput := range m.outputs {
		// TODO: check if output declared sensitive or not; currently we default always sensitive
		// See https://github.com/deislabs/porter/issues/855
		m.setSensitiveValue(stepOutput)
	}

	// Iterate through the bundle-level manifests and resolve for interpolation
	for _, outputDef := range m.GetTemplatedOutputs() {
		// TODO: ApplyTo can impact if the output is available
		// See https://github.com/deislabs/porter/issues/1159

		val, err := m.resolveBundleOutput(outputDef)
		if err != nil {
			return nil, err
		}
		if outputDef.Sensitive {
			m.setSensitiveValue(val)
		}

		if m.outputs == nil {
			m.outputs = map[string]string{}
		}
		m.outputs[outputDef.Name] = val
		bun["outputs"] = m.outputs
	}

	deps := make(map[string]interface{})
	bun["dependencies"] = deps
	for alias, bun := range m.bundles {
		// TODO: Support bundle.dependencies.ALIAS.parameters.NAME

		// bundle.dependencies.ALIAS.outputs.NAME
		depBundle := make(map[string]interface{})
		deps[alias] = depBundle

		depBundle["name"] = bun.Name
		depBundle["version"] = bun.Version
		depBundle["description"] = bun.Description

		depOutputs := make(map[string]interface{})
		depBundle["outputs"] = depOutputs

		if bun.Outputs == nil || m.Action == claim.ActionUninstall {
			// uninstalls are done backwards, so we don't have outputs available from dependencies
			// TODO: validate that they weren't trying to use them at build time so they don't find out at uninstall time
			continue
		}
		for name, output := range bun.Outputs {
			if !output.AppliesTo(string(m.Action)) {
				continue
			}

			value, err := ReadDependencyOutputValue(m.Context, alias, name)
			if err != nil {
				return nil, err
			}

			depOutputs[name] = value

			def := bun.Definitions[output.Definition]
			if def.WriteOnly != nil && *def.WriteOnly == true {
				m.setSensitiveValue(value)
			}
		}
	}
	images := make(map[string]interface{})
	bun["images"] = images
	for alias, image := range m.ImageMap {
		// just assigning the struct here results in uppercase keys, which would give us
		// strange things like {{ bundle.images.something.Repository }}
		// So reflect and walk through the struct (this way we don't need to update this later)
		val := reflect.ValueOf(image)
		img := make(map[string]string)
		typeOfT := val.Type()
		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			name := strings.ToLower(typeOfT.Field(i).Name)
			img[name] = f.String()
		}
		images[alias] = img
	}
	return data, nil
}

// ResolveStep will walk through the Step's data and resolve any placeholder
// data using the definitions in the manifest, like parameters or credentials.
func (m *RuntimeManifest) ResolveStep(step *manifest.Step) error {
	mustache.AllowMissingVariables = false
	sourceData, err := m.buildSourceData()
	if err != nil {
		return errors.Wrap(err, "unable to build step template data")
	}

	if m.Debug {
		fmt.Fprintf(m.Err, "=== Step Data ===\n%v\n", sourceData)
	}

	payload, err := yaml.Marshal(step)
	if err != nil {
		return errors.Wrapf(err, "invalid step data %v", step)
	}

	if m.Debug {
		fmt.Fprintf(m.Err, "=== Step Template ===\n%v\n", string(payload))
	}

	rendered, err := mustache.RenderRaw(string(payload), true, sourceData)
	if err != nil {
		return errors.Wrapf(err, "unable to render step template %s", string(payload))
	}

	if m.Debug {
		fmt.Fprintf(m.Err, "=== Rendered Step ===\n%s\n", rendered)
	}

	err = yaml.Unmarshal([]byte(rendered), step)
	if err != nil {
		return errors.Wrapf(err, "invalid step yaml\n%s", rendered)
	}

	return nil
}

// Prepare prepares the runtime environment prior to step execution
func (m *RuntimeManifest) Prepare() error {
	// For parameters of type "file", we may need to decode files on the filesystem
	// before execution of the step/action
	for _, param := range m.Parameters {
		if !param.AppliesTo(string(m.Action)) {
			continue
		}

		if param.Type == "file" {
			if param.Destination.Path == "" {
				return fmt.Errorf("destination path is not supplied for parameter %s", param.Name)
			}

			// Porter by default places parameter value into file determined by Destination.Path
			bytes, err := m.FileSystem.ReadFile(param.Destination.Path)
			if err != nil {
				return fmt.Errorf("unable to acquire value for parameter %s", param.Name)
			}

			decoded, err := base64.StdEncoding.DecodeString(string(bytes))
			if err != nil {
				return errors.Wrapf(err, "unable to decode parameter %s", param.Name)
			}

			err = m.FileSystem.WriteFile(param.Destination.Path, decoded, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "unable to write decoded parameter %s", param.Name)
			}
		}
	}
	return nil
}

// ResolveImages updates the RuntimeManifest to properly reflect the image map passed to the bundle via the
// mounted bundle.json and relocation mapping
func (m *RuntimeManifest) ResolveImages(bun *bundle.Bundle, reloMap relocation.ImageRelocationMap) error {
	// It only makes sense to process this if the runtime manifest has images defined. If none are defined
	// return early
	if len(m.ImageMap) == 0 {
		return nil
	}
	reverseLookup := make(map[string]string)
	for alias, image := range bun.Images {
		manifestImage, ok := m.ImageMap[alias]
		if !ok {
			return fmt.Errorf("unable to find image in porter manifest: %s", alias)
		}
		manifestImage.Digest = image.Digest
		err := resolveImage(&manifestImage, image.Image)
		if err != nil {
			return errors.Wrap(err, "unable to update image map from bundle.json")
		}
		m.ImageMap[alias] = manifestImage
		reverseLookup[image.Image] = alias
	}
	for oldRef, reloRef := range reloMap {
		alias := reverseLookup[oldRef]
		if manifestImage, ok := m.ImageMap[alias]; ok { //note, there might be other images in the relocation mapping, like the invocation image
			err := resolveImage(&manifestImage, reloRef)
			if err != nil {
				return errors.Wrap(err, "unable to update image map from relocation mapping")
			}
			m.ImageMap[alias] = manifestImage
		}
	}
	return nil
}

func resolveImage(image *manifest.MappedImage, refString string) error {
	//figure out what type of Reference it is so we can extract useful things for our image map
	ref, err := reference.Parse(refString)
	if err != nil {
		return errors.Wrapf(err, "unable to parse docker image %s", refString)
	}
	switch v := ref.(type) {
	case reference.Canonical:
		if tagged, ok := ref.(reference.NamedTagged); ok {
			image.Tag = tagged.Tag()
		}
		image.Repository = v.Name()
		image.Digest = v.Digest().String()

	case reference.NamedTagged:
		image.Tag = v.Tag()
		image.Repository = v.Name()
	case reference.Named:
		image.Repository = v.Name()
		image.Tag = "latest" //Populate this with latest so that the {{ can reference something }}
	}
	return nil
}
