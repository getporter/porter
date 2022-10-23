package runtime

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cbroglie/mustache"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/hashicorp/go-multierror"
	"github.com/opencontainers/go-digest"
	yaml3 "gopkg.in/yaml.v3"
)

const (
	// Path to the archive of the state from the previous run
	statePath = "/porter/state.tgz"
)

type RuntimeManifest struct {
	config RuntimeConfig
	*manifest.Manifest

	Action string

	// bundle is the executing bundle definition
	bundle cnab.ExtendedBundle

	// editor is the porter.yaml loaded into YQ so that we can
	// do advanced stuff with the manifest, like just read out the yaml for a particular step.
	editor *yaml.Editor

	// bundles is map of the dependencies bundle definitions, keyed by the alias used in the root manifest
	bundles map[string]cnab.ExtendedBundle

	steps           manifest.Steps
	outputs         map[string]string
	sensitiveValues []string
}

func NewRuntimeManifest(cfg RuntimeConfig, action string, manifest *manifest.Manifest) *RuntimeManifest {
	return &RuntimeManifest{
		config:   cfg,
		Action:   action,
		Manifest: manifest,
	}
}

// this is a temporary function to help write debug logs until we have PORTER_VERBOSITY passed into the bundle properly
func (m *RuntimeManifest) debugf(log tracing.TraceLogger, msg string, args ...interface{}) {
	if m.config.DebugMode {
		log.Infof(msg, args...)
	}
}

func (m *RuntimeManifest) Validate() error {
	err := m.loadBundle()
	if err != nil {
		return err
	}

	err = m.loadDependencyDefinitions()
	if err != nil {
		return err
	}

	err = m.setStepsByAction()
	if err != nil {
		return err
	}

	err = m.steps.Validate(m.Manifest)
	if err != nil {
		return fmt.Errorf("invalid action configuration: %w", err)
	}

	return nil
}

func (m *RuntimeManifest) loadBundle() error {
	// Load the CNAB representation of the bundle
	b, err := cnab.LoadBundle(m.config.Context, "/cnab/bundle.json")
	if err != nil {
		return err
	}

	m.bundle = b
	return nil
}

func (m *RuntimeManifest) GetInstallationNamespace() string {
	return m.config.Getenv(config.EnvPorterInstallationNamespace)
}

func (m *RuntimeManifest) GetInstallationName() string {
	return m.config.Getenv(config.EnvPorterInstallationName)
}

func (m *RuntimeManifest) loadDependencyDefinitions() error {
	m.bundles = make(map[string]cnab.ExtendedBundle, len(m.Dependencies.Requires))
	for _, dep := range m.Dependencies.Requires {
		bunD, err := GetDependencyDefinition(m.config.Context, dep.Name)
		if err != nil {
			return err
		}

		bun, err := bundle.Unmarshal(bunD)
		if err != nil {
			return fmt.Errorf("error unmarshaling bundle definition for dependency %s: %w", dep.Name, err)
		}

		m.bundles[dep.Name] = cnab.NewBundle(*bun)
	}

	return nil
}

func (m *RuntimeManifest) resolveParameter(pd manifest.ParameterDefinition) string {
	if pd.Destination.EnvironmentVariable != "" {
		return m.config.Getenv(pd.Destination.EnvironmentVariable)
	}
	if pd.Destination.Path != "" {
		return pd.Destination.Path
	}
	envVar := manifest.ParamToEnvVar(pd.Name)
	return m.config.Getenv(envVar)
}

func (m *RuntimeManifest) resolveCredential(cd manifest.CredentialDefinition) (string, error) {
	if cd.EnvironmentVariable != "" {
		return m.config.Getenv(cd.EnvironmentVariable), nil
	} else if cd.Path != "" {
		return cd.Path, nil
	} else {
		return "", fmt.Errorf("credential: %s is malformed", cd.Name)
	}
}

func (m *RuntimeManifest) resolveBundleOutput(outputName string) (string, error) {
	// Get the output's value from the injected parameter source
	ps := manifest.GetParameterSourceForOutput(outputName)
	psParamEnv := manifest.ParamToEnvVar(ps)
	outputValue, ok := m.config.LookupEnv(psParamEnv)
	if !ok {
		return "", fmt.Errorf("no parameter source was injected for output %s", outputName)
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
	case cnab.ActionInstall:
		m.steps = m.Install
	case cnab.ActionUninstall:
		m.steps = m.Uninstall
	case cnab.ActionUpgrade:
		m.steps = m.Upgrade
	default:
		customAction, ok := m.CustomActions[m.Action]
		if !ok {
			actions := make([]string, 0, len(m.CustomActions))
			for a := range m.CustomActions {
				actions = append(actions, a)
			}
			return fmt.Errorf("unsupported action %q, custom actions are defined for: %s", m.Action, strings.Join(actions, ", "))
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
	//lint:ignore U1000 ignore unused warning
	value string

	Name string                 `yaml:"name"`
	Data map[string]interface{} `yaml:",inline"`
}

func (m *RuntimeManifest) buildSourceData() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	m.sensitiveValues = []string{}

	inst := make(map[string]interface{})
	data["installation"] = inst
	inst["namespace"] = m.GetInstallationNamespace()
	inst["name"] = m.GetInstallationName()

	bun := make(map[string]interface{})
	data["bundle"] = bun

	// Enable interpolation of manifest/bundle name via bundle.name
	bun["name"] = m.Name
	bun["version"] = m.Version
	bun["description"] = m.Description
	bun["installerImage"] = m.Image
	bun["custom"] = m.Custom

	// Make environment variable accessible
	env := m.config.EnvironMap()
	data["env"] = env

	params := make(map[string]interface{})
	bun["parameters"] = params
	for _, param := range m.Parameters {
		if !param.AppliesTo(m.Action) {
			continue
		}

		pe := param.Name
		val := m.resolveParameter(param)
		if param.Sensitive {
			m.setSensitiveValue(val)
		}
		params[pe] = val
	}

	creds := make(map[string]interface{})
	bun["credentials"] = creds
	for _, cred := range m.Credentials {
		pe := cred.Name
		val, err := m.resolveCredential(cred)
		if err != nil {
			return nil, err
		}
		m.setSensitiveValue(val)
		creds[pe] = val
	}

	deps := make(map[string]interface{})
	bun["dependencies"] = deps
	for alias, depB := range m.bundles {
		// bundle.dependencies.ALIAS.outputs.NAME
		depBun := make(map[string]interface{})
		deps[alias] = depBun

		depBun["name"] = depB.Name
		depBun["version"] = depB.Version
		depBun["description"] = depB.Description
	}

	bun["outputs"] = m.outputs

	// Iterate through the runtime manifest's step outputs and determine if we should mask
	for name, val := range m.outputs {
		// TODO: support configuring sensitivity for step outputs that aren't also bundle-level outputs
		// See https://github.com/getporter/porter/issues/855

		// If step output is also a bundle-level output, defer to bundle-level output sensitivity
		if outputDef, ok := m.Outputs[name]; ok && !outputDef.Sensitive {
			continue
		}
		m.setSensitiveValue(val)
	}

	// Externally injected outputs (bundle level outputs and dependency outputs) are
	// injected through parameter sources.
	bunExt, err := m.bundle.ProcessRequiredExtensions()
	if err != nil {
		return nil, err
	}

	paramSources, _, err := bunExt.GetParameterSources()
	if err != nil {
		return nil, err
	}

	templatedOutputs := m.GetTemplatedOutputs()
	templatedDependencyOutputs := m.GetTemplatedDependencyOutputs()
	for paramName, sources := range paramSources {
		param := m.bundle.Parameters[paramName]
		if !param.AppliesTo(m.Action) {
			continue
		}

		for _, s := range sources.ListSourcesByPriority() {
			switch ps := s.(type) {
			case cnab.DependencyOutputParameterSource:
				outRef := manifest.DependencyOutputReference{Dependency: ps.Dependency, Output: ps.OutputName}

				// Ignore anything that isn't templated, because that's what we are building the source data for
				if _, isTemplated := templatedDependencyOutputs[outRef.String()]; !isTemplated {
					continue
				}

				depBun := deps[ps.Dependency].(map[string]interface{})
				var depOutputs map[string]interface{}
				if depBun["outputs"] == nil {
					depOutputs = make(map[string]interface{})
					depBun["outputs"] = depOutputs
				} else {
					depOutputs = depBun["outputs"].(map[string]interface{})
				}

				value, err := m.ReadDependencyOutputValue(outRef)
				if err != nil {
					return nil, err
				}

				depOutputs[ps.OutputName] = value

				// Determine if the dependency's output is defined as sensitive
				depB := m.bundles[ps.Dependency]
				if ok, _ := depB.IsOutputSensitive(ps.OutputName); ok {
					m.setSensitiveValue(value)
				}

			case cnab.OutputParameterSource:
				// Ignore anything that isn't templated, because that's what we are building the source data for
				if _, isTemplated := templatedOutputs[ps.OutputName]; !isTemplated {
					continue
				}

				// A bundle-level output may also be a step-level output
				// If already set, do not override
				if val, exists := m.outputs[ps.OutputName]; exists && val != "" {
					continue
				}

				val, err := m.resolveBundleOutput(ps.OutputName)
				if err != nil {
					return nil, err
				}

				if m.outputs == nil {
					m.outputs = map[string]string{}
				}
				m.outputs[ps.OutputName] = val
				bun["outputs"] = m.outputs

				outputDef := m.Manifest.Outputs[ps.OutputName]
				if outputDef.Sensitive {
					m.setSensitiveValue(val)
				}
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
			name := toCamelCase(typeOfT.Field(i).Name)
			img[name] = f.String()
		}
		images[alias] = img
	}
	return data, nil
}

// ResolveStep will walk through the Step's data and resolve any placeholder
// data using the definitions in the manifest, like parameters or credentials.
func (m *RuntimeManifest) ResolveStep(ctx context.Context, stepIndex int, step *manifest.Step) error {
	log := tracing.LoggerFromContext(ctx)

	// Refresh our template data
	sourceData, err := m.buildSourceData()
	if err != nil {
		return log.Error(fmt.Errorf("unable to build step template data: %w", err))
	}

	// Get the original yaml for the current step
	stepPath := fmt.Sprintf("%s[%d]", m.Action, stepIndex)
	stepTemplate, err := m.getStepTemplate(stepPath)
	if err != nil {
		return log.Error(fmt.Errorf("unable to retrieve original yaml for step %s: %w", stepPath, err))
	}

	// TODO: add back logging step data after we have a solid way to censor it in https://github.com/getporter/porter/issues/2256
	//fmt.Fprintf(m.Err, "=== Step Data ===\n%v\n", sourceData)
	m.debugf(log, "=== Step Template ===\n%v\n", stepTemplate)

	mustache.AllowMissingVariables = false
	rendered, err := mustache.RenderRaw(stepTemplate, true, sourceData)
	if err != nil {
		return log.Errorf("unable to render step template %s: %w", stepTemplate, err)
	}

	// TODO: add back logging step data after we have a solid way to censor it in https://github.com/getporter/porter/issues/2256
	//fmt.Fprintf(m.Err, "=== Rendered Step ===\n%s\n", rendered)

	// Update the step parameter with the result of rendering the template
	err = yaml.Unmarshal([]byte(rendered), step)
	if err != nil {
		return log.Error(fmt.Errorf("invalid step yaml after rendering template\n%s: %w", stepTemplate, err))
	}

	return nil
}

// Initialize prepares the runtime environment prior to step execution
func (m *RuntimeManifest) Initialize(ctx context.Context) error {
	if err := m.createOutputsDir(); err != nil {
		return err
	}

	// For parameters of type "file", we may need to decode files on the filesystem
	// before execution of the step/action
	for paramName, param := range m.bundle.Parameters {
		if !param.AppliesTo(m.Action) {
			continue
		}

		def, hasDef := m.bundle.Definitions[param.Definition]
		if !hasDef {
			continue
		}

		if m.bundle.IsFileType(def) {
			if param.Destination.Path == "" {
				return fmt.Errorf("destination path is not supplied for parameter %s", paramName)
			}

			// Porter by default places parameter value into file determined by Destination.Path
			bytes, err := m.config.FileSystem.ReadFile(param.Destination.Path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return fmt.Errorf("unable to acquire value for parameter %s: %w", paramName, err)
			}

			// TODO(carolynvs): hack around parameters ALWAYS being injected even when empty files mess things up
			// I'm not sure yet why it's injecting as null instead of "" (as required by the spec)
			// We want to get the null -> "" fixed, and also not write files into the bundle when unset.
			// that's a cnab change somewhere probably
			// the problem is in injectParameters in cnab-go
			if string(bytes) == "null" {
				m.config.FileSystem.Remove(param.Destination.Path)
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(string(bytes))
			if err != nil {
				return fmt.Errorf("unable to decode parameter %s: %w", paramName, err)
			}

			err = m.config.FileSystem.WriteFile(param.Destination.Path, decoded, pkg.FileModeWritable)
			if err != nil {
				return fmt.Errorf("unable to write decoded parameter %s: %w", paramName, err)
			}
		}
	}

	return m.unpackStateBag(ctx)
}

func (m *RuntimeManifest) createOutputsDir() error {
	// Ensure outputs directory exists
	if err := m.config.FileSystem.MkdirAll(config.BundleOutputsDir, pkg.FileModeDirectory); err != nil {
		return fmt.Errorf("unable to ensure CNAB outputs directory exists: %w", err)
	}
	return nil
}

// Unpack each state variable from /porter/state.tgz and copy it to its
// declared location in the bundle.
func (m *RuntimeManifest) unpackStateBag(ctx context.Context) error {
	log := tracing.LoggerFromContext(ctx)
	_, err := m.config.FileSystem.Open(statePath)
	if os.IsNotExist(err) || len(m.StateBag) == 0 {
		m.debugf(log, "No existing bundle state to unpack")
		return nil
	}
	bytes, err := m.config.FileSystem.ReadFile(statePath)
	if err != nil {
		m.debugf(log, "Unable to read bundle state file")
		return err
	}
	// TODO(sgettys): hack around state.tgz ALWAYS being injected even when empty files mess things up
	// I'm not sure yet why it's injecting as null instead of "" (as required by the spec)
	// We want to get the null -> "" fixed, and also not write files into the bundle when unset.
	// that's a cnab change somewhere probably
	// the problem is in injectParameters in cnab-go
	if string(bytes) == "null" {
		m.debugf(log, "Bundle state file has null content")
		m.config.FileSystem.Remove(statePath)
		return nil
	}
	// Unpack the state file and copy its contents to where the bundle expects them
	// state var name -> path in bundle
	log.Debug("Unpacking bundle state...")
	stateFiles := make(map[string]string, len(m.StateBag))
	for _, s := range m.StateBag {
		stateFiles[s.Name] = s.Path
	}

	unpackStateFile := func(tr *tar.Reader, header *tar.Header) error {
		name := strings.TrimPrefix(header.Name, "porter-state/")
		dest := stateFiles[name]
		m.debugf(log, "  - %s -> %s", name, dest)

		f, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return log.Error(fmt.Errorf("error creating state file %s: %w", dest, err))
		}
		defer f.Close()

		_, err = io.Copy(f, tr)
		if err != nil {
			return log.Error(fmt.Errorf("error unpacking state file %s: %w", dest, err))
		}

		return nil
	}

	stateArchive, err := m.config.FileSystem.Open(statePath)
	if err != nil {
		return log.Error(fmt.Errorf("could not open statefile at %s: %w", statePath, err))
	}

	gzr, err := gzip.NewReader(stateArchive)
	if err != nil {
		if err == io.EOF {
			log.Debug("statefile exists but is empty")
			return nil
		}
		return log.Error(fmt.Errorf("could not create a new gzip reader for the statefile: %w", err))
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
		} else if header.Typeflag != tar.TypeReg {
			continue
		}

		unpackStateFile(tr, header)
	}

	return nil
}

// Finalize cleans up the bundle before its completion.
func (m *RuntimeManifest) Finalize(ctx context.Context) error {
	var bigErr *multierror.Error

	if err := m.applyUnboundBundleOutputs(ctx); err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	// Always try to persist state, even when errors occur
	if err := m.packStateBag(ctx); err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	return bigErr.ErrorOrNil()
}

// Pack each state variable into /porter/state/tgz.
func (m *RuntimeManifest) packStateBag(ctx context.Context) error {
	log := tracing.LoggerFromContext(ctx)

	m.debugf(log, "Packing bundle state...")
	packStateFile := func(tw *tar.Writer, s manifest.StateVariable) error {
		fi, err := m.config.FileSystem.Stat(s.Path)
		if os.IsNotExist(err) {
			return nil
		}

		m.debugf(log, "  - %s", s.Path)
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return log.Error(fmt.Errorf("error creating tar header for state variable %s from path %s: %w", s.Name, s.Path, err))
		}
		header.Name = filepath.Join("porter-state", s.Name)

		if err := tw.WriteHeader(header); err != nil {
			return log.Error(fmt.Errorf("error writing tar header for state variable %s: %w", s.Name, err))
		}

		f, err := os.Open(s.Path)
		if err != nil {
			return log.Error(fmt.Errorf("error reading state file %s for variable %s: %w", s.Path, s.Name, err))
		}

		_, err = io.Copy(tw, f)
		if err != nil {
			return log.Error(fmt.Errorf("error archiving state file %s for variable %s: %w", s.Path, s.Name, err))
		}

		return nil
	}

	// Save directly to the final output location since we've already collected outputs at this point
	stateArchive, err := m.config.FileSystem.Create("/cnab/app/outputs/porter-state")
	if err != nil {
		return log.Error(fmt.Errorf("error creating porter statefile: %w", err))
	}

	gzw := gzip.NewWriter(stateArchive)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Persist as many of the state vars as possible, even if one fails
	var bigErr *multierror.Error
	for _, s := range m.StateBag {
		err := packStateFile(tw, s)
		if err != nil {
			bigErr = multierror.Append(bigErr, err)
		}
	}

	return log.Error(bigErr.ErrorOrNil())
}

// applyUnboundBundleOutputs finds outputs that haven't been bound yet by a step,
// and if they can be bound, i.e. they grab a file from the bundle's filesystem,
// apply the output.
func (m *RuntimeManifest) applyUnboundBundleOutputs(ctx context.Context) error {
	if len(m.bundle.Outputs) == 0 {
		return nil
	}

	log := tracing.LoggerFromContext(ctx)

	log.Debug("Collecting bundle outputs...")
	var bigErr *multierror.Error
	outputs := m.GetOutputs()
	for name, outputDef := range m.bundle.Outputs {
		outputSrcPath := m.Outputs[name].Path

		// We can only deal with outputs that are based on a file right now
		if outputSrcPath == "" {
			continue
		}

		if !outputDef.AppliesTo(m.Action) {
			continue
		}

		// Print the output that we've collected
		m.debugf(log, "  - %s", name)
		if _, hasOutput := outputs[name]; !hasOutput {
			// Use the path as originally defined in the manifest
			// TODO(carolynvs): When we switch to driving everything completely
			// from the bundle.json, we need to find a better way to get the original path value that the user specified.
			// We don't force people to output files immediately to /cnab/app/outputs and so that original location should
			// be persisted somewhere in the bundle.json (probably in custom)
			srcPath := manifest.ResolvePath(outputSrcPath)
			if _, err := m.config.FileSystem.Stat(srcPath); err != nil {
				continue
			}
			dstPath := filepath.Join(config.BundleOutputsDir, name)
			if dstExists, _ := m.config.FileSystem.Exists(dstPath); dstExists {
				continue
			}

			err := m.config.CopyFile(srcPath, dstPath)
			if err != nil {
				bigErr = multierror.Append(bigErr, fmt.Errorf("unable to copy output file from %s to %s: %w", srcPath, dstPath, err))
				continue
			}
		}
	}

	return log.Error(bigErr.ErrorOrNil())
}

// ResolveInvocationImage updates the RuntimeManifest to properly reflect the invocation image passed to the bundle via the
// mounted bundle.json and relocation mapping
func (m *RuntimeManifest) ResolveInvocationImage(bun cnab.ExtendedBundle, reloMap relocation.ImageRelocationMap) error {
	for _, image := range bun.InvocationImages {
		if image.Digest == "" || image.ImageType != "docker" {
			continue
		}

		ref, err := cnab.ParseOCIReference(image.Image)
		if err != nil {
			return fmt.Errorf("unable to parse invocation image reference: %w", err)
		}
		refWithDigest, err := ref.WithDigest(digest.Digest(image.Digest))
		if err != nil {
			return fmt.Errorf("unable to get invocation image reference with digest: %w", err)
		}

		m.Image = refWithDigest.String()
		break
	}
	relocated, ok := reloMap[m.Image]
	if ok {
		m.Image = relocated
	}
	return nil
}

// ResolveImages updates the RuntimeManifest to properly reflect the image map passed to the bundle via the
// mounted bundle.json and relocation mapping
func (m *RuntimeManifest) ResolveImages(bun cnab.ExtendedBundle, reloMap relocation.ImageRelocationMap) error {
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
			return fmt.Errorf("unable to update image map from bundle.json: %w", err)
		}
		m.ImageMap[alias] = manifestImage
		reverseLookup[image.Image] = alias
	}
	for oldRef, reloRef := range reloMap {
		alias := reverseLookup[oldRef]
		if manifestImage, ok := m.ImageMap[alias]; ok { //note, there might be other images in the relocation mapping, like the invocation image
			err := resolveImage(&manifestImage, reloRef)
			if err != nil {
				return fmt.Errorf("unable to update image map from relocation mapping: %w", err)
			}
			m.ImageMap[alias] = manifestImage
		}
	}
	return nil
}

func (m *RuntimeManifest) getEditor() (*yaml.Editor, error) {
	if m.editor != nil {
		return m.editor, nil
	}

	// Get the original porter.yaml with additional yaml metadata so that we can look at just the current step's yaml
	yq := yaml.NewEditor(m.config.Context)
	if err := yq.ReadFile(m.ManifestPath); err != nil {
		return nil, fmt.Errorf("error loading yaml editor from %s", m.ManifestPath)
	}

	return yq, nil
}

func (m *RuntimeManifest) getStepTemplate(stepPath string) (string, error) {
	yq, err := m.getEditor()
	if err != nil {
		return "", err
	}

	stepNode, err := yq.GetNode(stepPath)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve original yaml for step %s: %w", stepPath, err)
	}

	var stepYAML bytes.Buffer
	enc := yaml3.NewEncoder(&stepYAML)
	defer enc.Close()
	if err := enc.Encode(stepNode); err != nil {
		return "", fmt.Errorf("error re-encoding porter.yaml for templating: %w", err)
	}

	stepTemplate := m.GetTemplatePrefix() + stepYAML.String()
	return stepTemplate, nil
}

func resolveImage(image *manifest.MappedImage, refString string) error {
	//figure out what type of Reference it is so we can extract useful things for our image map
	ref, err := cnab.ParseOCIReference(refString)
	if err != nil {
		return err
	}

	image.Repository = ref.Repository()
	if ref.HasDigest() {
		image.Digest = ref.Digest().String()
	}

	if ref.HasTag() {
		image.Tag = ref.Tag()
	}

	if ref.IsRepositoryOnly() {
		image.Tag = "latest" // Populate this with latest so that the {{ can reference something }}
	}

	return nil
}

// toCamelCase returns a camel-cased variant of the provided string
func toCamelCase(str string) string {
	var b strings.Builder

	b.WriteString(strings.ToLower(string(str[0])))
	b.WriteString(str[1:])

	return b.String()
}
