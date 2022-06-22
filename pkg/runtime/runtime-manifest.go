package runtime

import (
	"archive/tar"
	"compress/gzip"
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
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cbroglie/mustache"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/hashicorp/go-multierror"
)

const (
	// Path to the archive of the state from the previous run
	statePath = "/porter/state.tgz"
)

type RuntimeManifest struct {
	*portercontext.Context
	*manifest.Manifest

	Action string

	// bundle is the executing bundle definition
	bundle cnab.ExtendedBundle

	// bundles is map of the dependencies bundle definitions, keyed by the alias used in the root manifest
	bundles map[string]cnab.ExtendedBundle

	steps           manifest.Steps
	outputs         map[string]string
	sensitiveValues []string
}

func NewRuntimeManifest(cxt *portercontext.Context, action string, manifest *manifest.Manifest) *RuntimeManifest {
	return &RuntimeManifest{
		Context:  cxt,
		Action:   action,
		Manifest: manifest,
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
	b, err := cnab.LoadBundle(m.Context, "/cnab/bundle.json")
	if err != nil {
		return err
	}

	m.bundle = b
	return nil
}

func (m *RuntimeManifest) GetInstallationNamespace() string {
	return m.Getenv(config.EnvPorterInstallationNamespace)
}

func (m *RuntimeManifest) GetInstallationName() string {
	return m.Getenv(config.EnvPorterInstallationName)
}

func (m *RuntimeManifest) loadDependencyDefinitions() error {
	m.bundles = make(map[string]cnab.ExtendedBundle, len(m.Dependencies.RequiredDependencies))
	for _, dep := range m.Dependencies.RequiredDependencies {
		bunD, err := GetDependencyDefinition(m.Context, dep.Name)
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
		return m.Getenv(pd.Destination.EnvironmentVariable)
	}
	if pd.Destination.Path != "" {
		return pd.Destination.Path
	}
	envVar := manifest.ParamToEnvVar(pd.Name)
	return m.Getenv(envVar)
}

func (m *RuntimeManifest) resolveCredential(cd manifest.CredentialDefinition) (string, error) {
	if cd.EnvironmentVariable != "" {
		return m.Getenv(cd.EnvironmentVariable), nil
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
	outputValue, ok := m.LookupEnv(psParamEnv)
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
	bun["invocationImage"] = m.Image
	bun["custom"] = m.Custom

	// Make environment variable accessible
	env := m.EnvironMap()
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
func (m *RuntimeManifest) ResolveStep(step *manifest.Step) error {
	mustache.AllowMissingVariables = false
	sourceData, err := m.buildSourceData()
	if err != nil {
		return fmt.Errorf("unable to build step template data: %w", err)
	}

	if m.Debug {
		fmt.Fprintf(m.Err, "=== Step Data ===\n%v\n", sourceData)
	}

	payload, err := yaml.Marshal(step)
	if err != nil {
		return fmt.Errorf("invalid step data %v: %w", step, err)
	}

	if m.Debug {
		fmt.Fprintf(m.Err, "=== Step Template ===\n%v\n", string(payload))
	}

	rendered, err := mustache.RenderRaw(string(payload), true, sourceData)
	if err != nil {
		return fmt.Errorf("unable to render step template %s: %w", string(payload), err)
	}

	if m.Debug {
		fmt.Fprintf(m.Err, "=== Rendered Step ===\n%s\n", rendered)
	}

	err = yaml.Unmarshal([]byte(rendered), step)
	if err != nil {
		return fmt.Errorf("invalid step yaml\n%s: %w", rendered, err)
	}

	return nil
}

// Init prepares the runtime environment prior to step execution
func (m *RuntimeManifest) Initialize() error {
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
			bytes, err := m.FileSystem.ReadFile(param.Destination.Path)
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
				m.FileSystem.Remove(param.Destination.Path)
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(string(bytes))
			if err != nil {
				return fmt.Errorf("unable to decode parameter %s: %w", paramName, err)
			}

			err = m.FileSystem.WriteFile(param.Destination.Path, decoded, pkg.FileModeWritable)
			if err != nil {
				return fmt.Errorf("unable to write decoded parameter %s: %w", paramName, err)
			}
		}
	}

	return m.unpackStateBag()
}

func (m *RuntimeManifest) createOutputsDir() error {
	// Ensure outputs directory exists
	if err := m.FileSystem.MkdirAll(config.BundleOutputsDir, pkg.FileModeDirectory); err != nil {
		return fmt.Errorf("unable to ensure CNAB outputs directory exists: %w", err)
	}
	return nil
}

// Unpack each state variable from /porter/state.tgz and copy it to its
// declared location in the bundle.
func (m *RuntimeManifest) unpackStateBag() error {
	_, err := m.FileSystem.Open(statePath)
	if os.IsNotExist(err) || len(m.StateBag) == 0 {
		if m.Debug {
			fmt.Fprintln(m.Err, "No existing bundle state to unpack")
		}
		return nil
	}

	if m.Debug {
		fmt.Fprintln(m.Err, "Unpacking bundle state...")
	}
	// Unpack the state file and copy its contents to where the bundle expects them
	// state var name -> path in bundle
	stateFiles := make(map[string]string, len(m.StateBag))
	for _, s := range m.StateBag {
		stateFiles[s.Name] = s.Path
	}

	unpackStateFile := func(tr *tar.Reader, header *tar.Header) error {
		name := strings.TrimPrefix(header.Name, "porter-state/")
		dest := stateFiles[name]
		if m.Debug {
			fmt.Fprintln(m.Err, "  -", name, "->", dest)
		}

		f, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("error creating state file %s: %w", dest, err)
		}
		defer f.Close()

		_, err = io.Copy(f, tr)
		if err != nil {
			return fmt.Errorf("error unpacking state file %s: %w", dest, err)
		}

		return nil
	}

	stateArchive, err := m.FileSystem.Open(statePath)
	if err != nil {
		return err
	}

	gzr, err := gzip.NewReader(stateArchive)
	if err != nil {
		return err
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
func (m *RuntimeManifest) Finalize() error {
	var bigErr *multierror.Error

	if err := m.applyUnboundBundleOutputs(); err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	// Always try to persist state, even when errors occur
	if err := m.packStateBag(); err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	return bigErr.ErrorOrNil()
}

// Pack each state variable into /porter/state/tgz.
func (m *RuntimeManifest) packStateBag() error {
	if m.Debug {
		fmt.Fprintln(m.Err, "Packing bundle state...")
	}

	packStateFile := func(tw *tar.Writer, s manifest.StateVariable) error {
		fi, err := m.FileSystem.Stat(s.Path)
		if os.IsNotExist(err) {
			return nil
		}

		if m.Debug {
			fmt.Fprintln(m.Err, "  -", s.Path)
		}
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return fmt.Errorf("error creating tar header for state variable %s from path %s: %w", s.Name, s.Path, err)
		}
		header.Name = filepath.Join("porter-state", s.Name)

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing tar header for state variable %s: %w", s.Name, err)
		}

		f, err := os.Open(s.Path)
		if err != nil {
			return fmt.Errorf("error reading state file %s for variable %s: %w", s.Path, s.Name, err)
		}

		_, err = io.Copy(tw, f)
		if err != nil {
			return fmt.Errorf("error archiving state file %s for variable %s: %w", s.Path, s.Name, err)
		}

		return nil
	}

	// Save directly to the final output location since we've already collected outputs at this point
	stateArchive, err := m.FileSystem.Create("/cnab/app/outputs/porter-state")
	if err != nil {
		return err
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

	return bigErr.ErrorOrNil()
}

// applyUnboundBundleOutputs finds outputs that haven't been bound yet by a step,
// and if they can be bound, i.e. they grab a file from the bundle's filesystem,
// apply the output.
func (m *RuntimeManifest) applyUnboundBundleOutputs() error {
	if len(m.bundle.Outputs) > 0 {
		if m.Debug {
			fmt.Fprintln(m.Err, "Collecting bundle outputs...")
		}
	}

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
		if m.Debug {
			fmt.Fprintln(m.Err, "  -", name)
		}
		if _, hasOutput := outputs[name]; !hasOutput {
			// Use the path as originally defined in the manifest
			// TODO(carolynvs): When we switch to driving everything completely
			// from the bundle.json, we need to find a better way to get the original path value that the user specified.
			// We don't force people to output files immediately to /cnab/app/outputs and so that original location should
			// be persisted somewhere in the bundle.json (probably in custom)
			srcPath := manifest.ResolvePath(outputSrcPath)
			if _, err := m.FileSystem.Stat(srcPath); err != nil {
				continue
			}
			dstPath := filepath.Join(config.BundleOutputsDir, name)
			if dstExists, _ := m.FileSystem.Exists(dstPath); dstExists {
				continue
			}

			err := m.CopyFile(srcPath, dstPath)
			if err != nil {
				bigErr = multierror.Append(bigErr, fmt.Errorf("unable to copy output file from %s to %s: %w", srcPath, dstPath, err))
				continue
			}
		}
	}

	return bigErr.ErrorOrNil()
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
