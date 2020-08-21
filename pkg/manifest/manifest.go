package manifest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"github.com/Masterminds/semver"
	"github.com/cbroglie/mustache"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/docker/distribution/reference"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const invalidStepErrorFormat = "validation of action \"%s\" failed"

type Manifest struct {
	// ManifestPath is the location from which the manifest was loaded, such as the path on the filesystem or a url.
	ManifestPath string `yaml:"-"`

	// TemplateVariables are the variables used in the templating, e.g. bundle.parameters.NAME, or bundle.outputs.NAME
	TemplateVariables []string `yaml:"-"`

	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version,omitempty"`

	// BundleTag is the name of the bundle in the format REGISTRY/NAME:TAG
	BundleTag string `yaml:"tag"`

	// Image is the name of the invocation image in the format REGISTRY/NAME:TAG
	// It doesn't map to any field in the manifest as it isn't meant to be user-specified
	Image string `yaml:"-"`

	// Dockerfile is the relative path to the Dockerfile template for the invocation image
	Dockerfile string `yaml:"dockerfile,omitempty"`

	Mixins []MixinDeclaration `yaml:"mixins,omitempty"`

	Install   Steps `yaml:"install"`
	Uninstall Steps `yaml:"uninstall"`
	Upgrade   Steps `yaml:"upgrade"`

	Custom                  map[string]interface{}            `yaml:"custom,omitempty"`
	CustomActions           map[string]Steps                  `yaml:"-"`
	CustomActionDefinitions map[string]CustomActionDefinition `yaml:"customActions,omitempty"`

	Parameters   []ParameterDefinition  `yaml:"parameters,omitempty"`
	Credentials  []CredentialDefinition `yaml:"credentials,omitempty"`
	Dependencies []Dependency           `yaml:"dependencies,omitempty"`
	Outputs      []OutputDefinition     `yaml:"outputs,omitempty"`

	// ImageMap is a map of images referenced in the bundle. If an image relocation mapping is later provided, that
	// will be mounted at as a file at runtime to /cnab/app/relocation-mapping.json.
	ImageMap map[string]MappedImage `yaml:"images,omitempty"`

	Required []RequiredExtension `yaml:"required,omitempty"`
}

func (m *Manifest) Validate() error {
	var result error

	err := m.SetDefaults()
	if err != nil {
		return err
	}

	if strings.ToLower(m.Dockerfile) == "dockerfile" {
		return errors.New("Dockerfile template cannot be named 'Dockerfile' because that is the filename generated during porter build")
	}

	if len(m.Mixins) == 0 {
		result = multierror.Append(result, errors.New("no mixins declared"))
	}

	if m.Install == nil {
		result = multierror.Append(result, errors.New("no install action defined"))
	}
	err = m.Install.Validate(m)
	if err != nil {
		result = multierror.Append(result, errors.Wrapf(err, fmt.Sprintf(invalidStepErrorFormat, "install")))
	}

	if m.Uninstall == nil {
		result = multierror.Append(result, errors.New("no uninstall action defined"))
	}
	err = m.Uninstall.Validate(m)
	if err != nil {
		result = multierror.Append(result, errors.Wrapf(err, fmt.Sprintf(invalidStepErrorFormat, "uninstall")))
	}

	for actionName, steps := range m.CustomActions {
		err := steps.Validate(m)
		if err != nil {
			result = multierror.Append(result, errors.Wrapf(err, fmt.Sprintf(invalidStepErrorFormat, actionName)))
		}
	}

	for _, dep := range m.Dependencies {
		err = dep.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	for _, output := range m.Outputs {
		err = output.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	for _, parameter := range m.Parameters {
		err = parameter.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	for _, image := range m.ImageMap {
		err = image.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

var templatedOutputRegex = regexp.MustCompile(`^bundle\.outputs\.(.+)$`)

// getTemplateOutputName returns the output name from the template variable.
func (m *Manifest) getTemplateOutputName(value string) (string, bool) {
	matches := templatedOutputRegex.FindStringSubmatch(value)
	if len(matches) < 2 {
		return "", false
	}

	outputName := matches[1]
	return outputName, true
}

var templatedDependencyOutputRegex = regexp.MustCompile(`^bundle\.dependencies\.(.+).outputs.(.+)$`)

// getTemplateDependencyOutputName returns the dependency and output name from the
// template variable.
func (m *Manifest) getTemplateDependencyOutputName(value string) (string, string, bool) {
	matches := templatedDependencyOutputRegex.FindStringSubmatch(value)
	if len(matches) < 3 {
		return "", "", false
	}

	dependencyName := matches[1]
	outputName := matches[2]
	return dependencyName, outputName, true
}

// GetTemplatedOutputs returns the output definitions for any bundle level outputs
// that have been templated, keyed by the output name.
func (m *Manifest) GetTemplatedOutputs() OutputDefinitions {
	outputs := make(OutputDefinitions, len(m.TemplateVariables))
	for _, tmplVar := range m.TemplateVariables {
		if name, ok := m.getTemplateOutputName(tmplVar); ok {
			outputDef, ok := m.Outputs[name]
			if !ok {
				// Only return bundle level definitions
				continue
			}
			outputs[name] = outputDef
		}
	}
	return outputs
}

// GetTemplatedOutputs returns the output definitions for any bundle level outputs
// that have been templated, keyed by "DEPENDENCY.OUTPUT".
func (m *Manifest) GetTemplatedDependencyOutputs() DependencyOutputReferences {
	outputs := make(DependencyOutputReferences, len(m.TemplateVariables))
	for _, tmplVar := range m.TemplateVariables {
		if dep, output, ok := m.getTemplateDependencyOutputName(tmplVar); ok {
			ref := DependencyOutputReference{
				Dependency: dep,
				Output:     output,
			}
			outputs[ref.String()] = ref
		}
	}
	return outputs
}

type DependencyOutputReference struct {
	Dependency string
	Output     string
}

func (r DependencyOutputReference) String() string {
	return fmt.Sprintf("%s.%s", r.Dependency, r.Output)
}

type DependencyOutputReferences map[string]DependencyOutputReference

// ParameterDefinitions allows us to represent parameters as a list in the YAML
// and work with them as a map internally
type ParameterDefinitions map[string]ParameterDefinition

func (pd ParameterDefinitions) MarshalYAML() (interface{}, error) {
	raw := make([]ParameterDefinition, 0, len(pd))

	for _, param := range pd {
		raw = append(raw, param)
	}

	return raw, nil
}

func (pd *ParameterDefinitions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw []ParameterDefinition
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	if *pd == nil {
		*pd = make(map[string]ParameterDefinition, len(raw))
	}

	for _, item := range raw {
		(*pd)[item.Name] = item
	}

	return nil
}

// ParameterDefinition defines a single parameter for a CNAB bundle
type ParameterDefinition struct {
	Name      string          `yaml:"name"`
	Sensitive bool            `yaml:"sensitive"`
	Source    ParameterSource `yaml:"source,omitempty"`

	// These fields represent a subset of bundle.Parameter as defined in cnabio/cnab-go,
	// minus the 'Description' field (definition.Schema's will be used) and `Definition` field
	ApplyTo     []string `yaml:"applyTo,omitempty"`
	Destination Location `yaml:",inline,omitempty"`

	definition.Schema `yaml:",inline"`
}

func (pd *ParameterDefinition) Validate() error {
	var result *multierror.Error

	if pd.Name == "" {
		result = multierror.Append(result, errors.New("parameter name is required"))
	}

	// Porter supports declaring a parameter of type: "file",
	// which we will convert to the appropriate bundle.Parameter type in adapter.go
	// Here, we copy the ParameterDefinition and make the same modification before validation
	pdCopy := pd.DeepCopy()
	if pdCopy.Type == "file" {
		if pd.Destination.Path == "" {
			result = multierror.Append(result, fmt.Errorf("no destination path supplied for parameter %s", pd.Name))
		}
		pdCopy.Type = "string"
		pdCopy.ContentEncoding = "base64"
	}

	schemaValidationErrs, err := pdCopy.Schema.Validate(pdCopy)
	if err != nil {
		result = multierror.Append(result, errors.Wrapf(err, "encountered error while validating parameter %s", pdCopy.Name))
	}
	for _, schemaValidationErr := range schemaValidationErrs {
		result = multierror.Append(result, errors.Wrapf(err, "encountered validation error(s) for parameter %s: %v", pdCopy.Name, schemaValidationErr))
	}

	return result.ErrorOrNil()
}

// DeepCopy copies a ParameterDefinition and returns the copy
func (pd *ParameterDefinition) DeepCopy() *ParameterDefinition {
	var p2 ParameterDefinition
	p2 = *pd
	p2.ApplyTo = make([]string, len(pd.ApplyTo))
	copy(p2.ApplyTo, pd.ApplyTo)
	return &p2
}

// AppliesTo returns a boolean value specifying whether or not
// the Parameter applies to the provided action
func (pd *ParameterDefinition) AppliesTo(action string) bool {
	if len(pd.ApplyTo) == 0 {
		return true
	}
	for _, act := range pd.ApplyTo {
		if action == act {
			return true
		}
	}
	return false
}

type ParameterSource struct {
	Dependency string `yaml:"dependency,omitempty"`
	Output     string `yaml:"output"`
}

// CredentialDefinitions allows us to represent credentials as a list in the YAML
// and work with them as a map internally
type CredentialDefinitions map[string]CredentialDefinition

func (cd CredentialDefinitions) MarshalYAML() (interface{}, error) {
	raw := make([]CredentialDefinition, 0, len(cd))

	for _, cred := range cd {
		raw = append(raw, cred)
	}

	return raw, nil
}

func (cd *CredentialDefinitions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw []CredentialDefinition
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	if *cd == nil {
		*cd = make(map[string]CredentialDefinition, len(raw))
	}

	for _, item := range raw {
		(*cd)[item.Name] = item
	}

	return nil
}

// CredentialDefinition represents the structure or fields of a credential parameter
type CredentialDefinition struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty"`

	Location `yaml:",inline"`
}

func (cd *CredentialDefinition) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawCreds CredentialDefinition
	rawCred := rawCreds{
		Name:        cd.Name,
		Description: cd.Description,
		Required:    true,
		Location:    cd.Location,
	}

	if err := unmarshal(&rawCred); err != nil {
		return err
	}

	*cd = CredentialDefinition(rawCred)

	return nil
}

// TODO: use cnab-go's bundle.Location instead, once yaml tags have been added
// Location represents a Parameter or Credential location in an InvocationImage
type Location struct {
	Path                string `yaml:"path,omitempty"`
	EnvironmentVariable string `yaml:"env,omitempty"`
}

func (l Location) IsEmpty() bool {
	var empty Location
	return l == empty
}

type MixinDeclaration struct {
	Name   string
	Config interface{}
}

// UnmarshalYAML allows mixin declarations to either be a normal list of strings
// mixins:
// - exec
// - helm
// or allow some entries to have config data defined
// - az:
//     extensions:
//       - iot
func (m *MixinDeclaration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First try to just read the mixin name
	var mixinNameOnly string
	err := unmarshal(&mixinNameOnly)
	if err == nil {
		m.Name = mixinNameOnly
		m.Config = nil
		return nil
	}

	// Next try to read a mixin name with config defined
	mixinWithConfig := map[string]interface{}{}
	err = unmarshal(&mixinWithConfig)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal raw yaml of mixin declarations")
	}

	if len(mixinWithConfig) == 0 {
		return errors.New("mixin declaration was empty")
	} else if len(mixinWithConfig) > 1 {
		return errors.New("mixin declaration contained more than one mixin")
	}

	for mixinName, config := range mixinWithConfig {
		m.Name = mixinName
		m.Config = config
		break // There is only one mixin anyway but break for clarity
	}
	return nil
}

// MarshalYAML allows mixin declarations to either be a normal list of strings
// mixins:
// - exec
// - helm
// or allow some entries to have config data defined
// - az:
//     extensions:
//       - iot
func (m MixinDeclaration) MarshalYAML() (interface{}, error) {
	if m.Config == nil {
		return m.Name, nil
	}

	raw := map[string]interface{}{
		m.Name: m.Config,
	}
	return raw, nil
}

type MappedImage struct {
	Description string            `yaml:"description"`
	ImageType   string            `yaml:"imageType"`
	Repository  string            `yaml:"repository"`
	Digest      string            `yaml:"digest,omitempty"`
	Size        uint64            `yaml:"size,omitempty"`
	MediaType   string            `yaml:"mediaType,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Tag         string            `yaml:"tag,omitempty"`
}

func (mi *MappedImage) Validate() error {
	if mi.Digest != "" {
		anchoredDigestRegex := regexp.MustCompile(`^` + reference.DigestRegexp.String() + `$`)
		if !anchoredDigestRegex.MatchString(mi.Digest) {
			return reference.ErrDigestInvalidFormat
		}
	}

	if _, err := reference.Parse(mi.Repository); err != nil {
		return err
	}

	return nil
}

type Dependency struct {
	Name             string            `yaml:"name"`
	Tag              string            `yaml:"tag"`
	Versions         []string          `yaml:"versions"`
	AllowPrereleases bool              `yaml:"prereleases"`
	Parameters       map[string]string `yaml:"parameters,omitempty"`
}

func (d *Dependency) Validate() error {
	if d.Name == "" {
		return errors.New("dependency name is required")
	}

	if d.Tag == "" {
		return errors.New("dependency tag is required")
	}

	if strings.Contains(d.Tag, ":") && len(d.Versions) > 0 {
		return errors.New("dependency tag can only specify REGISTRY/NAME when version ranges are specified")
	}

	return nil
}

type CustomActionDefinition struct {
	Description       string `yaml:"description,omitempty"`
	ModifiesResources bool   `yaml:"modifies,omitempty"`
	Stateless         bool   `yaml:"stateless,omitempty"`
}

// OutputDefinitions allows us to represent parameters as a list in the YAML
// and work with them as a map internally
type OutputDefinitions map[string]OutputDefinition

func (od OutputDefinitions) MarshalYAML() (interface{}, error) {
	raw := make([]OutputDefinition, 0, len(od))

	for _, output := range od {
		raw = append(raw, output)
	}

	return raw, nil
}

func (od *OutputDefinitions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw []OutputDefinition
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	if *od == nil {
		*od = make(map[string]OutputDefinition, len(raw))
	}

	for _, item := range raw {
		(*od)[item.Name] = item
	}

	return nil
}

// OutputDefinition defines a single output for a CNAB
type OutputDefinition struct {
	Name      string   `yaml:"name"`
	ApplyTo   []string `yaml:"applyTo,omitempty"`
	Sensitive bool     `yaml:"sensitive"`

	// This is not in the CNAB spec, but it allows a mixin to create a file
	// and porter will take care of making it a proper output.
	Path string `yaml:"path,omitempty"`

	definition.Schema `yaml:",inline"`
}

// DeepCopy copies a ParameterDefinition and returns the copy
func (od *OutputDefinition) DeepCopy() *OutputDefinition {
	var o2 OutputDefinition
	o2 = *od
	o2.ApplyTo = make([]string, len(od.ApplyTo))
	copy(o2.ApplyTo, od.ApplyTo)
	return &o2
}

func (od *OutputDefinition) Validate() error {
	var result *multierror.Error

	if od.Name == "" {
		return errors.New("output name is required")
	}

	// Porter supports declaring an output of type: "file",
	// which we will convert to the appropriate type in adapter.go
	// Here, we copy the definition and make the same modification before validation
	odCopy := od.DeepCopy()
	if odCopy.Type == "file" {
		if od.Path == "" {
			result = multierror.Append(result, fmt.Errorf("no path supplied for output %s", od.Name))
		}
		odCopy.Type = "string"
		odCopy.ContentEncoding = "base64"
	}

	schemaValidationErrs, err := odCopy.Schema.Validate(od)
	if err != nil {
		result = multierror.Append(result, errors.Wrapf(err, "encountered error while validating output %s", odCopy.Name))
	}
	for _, schemaValidationErr := range schemaValidationErrs {
		result = multierror.Append(result, errors.Wrapf(err, "encountered validation error(s) for output %s: %v", odCopy.Name, schemaValidationErr))
	}

	return result.ErrorOrNil()
}

type BundleOutput struct {
	Name                string `yaml:"name"`
	Path                string `yaml:"path"`
	EnvironmentVariable string `yaml:"env"`
}

type Steps []*Step

func (s Steps) Validate(m *Manifest) error {
	for _, step := range s {
		err := step.Validate(m)
		if err != nil {
			return err
		}
	}
	return nil
}

type Step struct {
	Data map[string]interface{} `yaml:",inline"`
}

func (s *Step) Validate(m *Manifest) error {
	if s == nil {
		return errors.New("found an empty step")
	}
	if len(s.Data) == 0 {
		return errors.New("no mixin specified")
	}
	if len(s.Data) > 1 {
		return errors.New("more than one mixin specified")
	}

	mixinDeclared := false
	mixinType := s.GetMixinName()
	for _, mixin := range m.Mixins {
		if mixin.Name == mixinType {
			mixinDeclared = true
			break
		}
	}
	if !mixinDeclared {
		return errors.Errorf("mixin (%s) was not declared", mixinType)
	}

	if _, err := s.GetDescription(); err != nil {
		return err
	}

	return nil
}

// GetDescription returns a description of the step.
// Every step must have this property.
func (s *Step) GetDescription() (string, error) {
	if s.Data == nil {
		return "", errors.New("empty step data")
	}

	mixinName := s.GetMixinName()
	children := s.Data[mixinName]
	d, ok := children.(map[interface{}]interface{})["description"]
	if !ok {
		return "", errors.Errorf("mixin step (%s) missing description", mixinName)
	}
	desc, ok := d.(string)
	if !ok {
		return "", errors.Errorf("invalid description type (%T) for mixin step (%s)", desc, mixinName)
	}

	return desc, nil
}

func (s *Step) GetMixinName() string {
	var mixinName string
	for k := range s.Data {
		mixinName = k
	}
	return mixinName
}

func UnmarshalManifest(cxt *context.Context, manifestData []byte) (*Manifest, error) {
	// Unmarshal the manifest into the normal struct
	manifest := &Manifest{}
	err := yaml.Unmarshal(manifestData, &manifest)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling the typed manifest")
	}

	// Do a second pass to identify custom actions, which don't have yaml tags since they are dynamic
	// 1. Marshal the manifest a second time into a plain map
	// 2. Remove keys for fields that are already mapped with yaml tags
	// 3. Anything left is a custom action

	// Marshal the manifest into an untyped map
	unmappedData := make(map[string]interface{})
	err = yaml.Unmarshal(manifestData, &unmappedData)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling the untyped manifest")
	}

	// Use reflection to figure out which fields are on the manifest and have yaml tags
	objValue := reflect.ValueOf(manifest).Elem()
	knownFields := map[string]reflect.Value{}
	for i := 0; i != objValue.NumField(); i++ {
		tagName := strings.Split(objValue.Type().Field(i).Tag.Get("yaml"), ",")[0]
		knownFields[tagName] = objValue.Field(i)
	}

	// Remove any fields that have yaml tags
	for key := range unmappedData {
		if _, found := knownFields[key]; found {
			delete(unmappedData, key)
		}
		// Print deprecation notice for this field
		if key == "invocationImage" {
			fmt.Fprintln(cxt.Out, "WARNING: The invocationImage field has been deprecated and can no longer be user-specified; ignoring.")
			delete(unmappedData, key)
		}
	}

	// Marshal the remaining keys in the unmappedData as custom actions and append them to the typed manifest
	manifest.CustomActions = make(map[string]Steps, len(unmappedData))
	for key, chunk := range unmappedData {
		chunkData, err := yaml.Marshal(chunk)
		if err != nil {
			return nil, errors.Wrapf(err, "error remarshaling custom action %s", key)
		}

		steps := Steps{}
		err = yaml.Unmarshal(chunkData, &steps)
		if err != nil {
			return nil, errors.Wrapf(err, "error unmarshaling custom action %s", key)
		}

		manifest.CustomActions[key] = steps
	}

	return manifest, nil
}

// SetDefaults updates the manifest with default values where not populated
func (m *Manifest) SetDefaults() error {
	return m.SetInvocationImageFromBundleTag(m.BundleTag)
}

// SetInvocationImageFromBundleTag sets the invocation image name on the manifest
// per the provided bundle tag, also updating the manifest BundleTag value
// if it initially lacks a Docker tag
func (m *Manifest) SetInvocationImageFromBundleTag(bundleTag string) error {
	bundleRef, err := reference.ParseNormalizedNamed(bundleTag)
	if err != nil {
		return errors.Wrapf(err, "invalid tag %s", bundleTag)
	}

	dockerTag, err := m.getDockerTagFromBundleRef(bundleRef)
	if err != nil {
		return errors.Wrapf(err, "unable to derive docker tag from bundle tag %q", bundleTag)
	}

	// If the docker tag is initially missing from bundleTag, update with
	// returned dockerTag
	switch v := bundleRef.(type) {
	case reference.Named:
		bundleRef, err = reference.WithTag(v, dockerTag)
		if err != nil {
			return errors.Wrapf(err, "could not set bundle tag to %q", dockerTag)
		}
		m.BundleTag = reference.FamiliarString(bundleRef)
	}

	imageName, err := reference.ParseNormalizedNamed(bundleRef.Name() + "-installer")
	if err != err {
		return errors.Wrapf(err, "could not set invocation image to %q", bundleRef.Name()+"-installer")
	}
	imageRef, err := reference.WithTag(imageName, dockerTag)
	if err != nil {
		return errors.Wrapf(err, "could not set invocation image tag to %q", dockerTag)
	}
	m.Image = reference.FamiliarString(imageRef)

	return nil
}

// getDockerTagFromBundleRef returns the Docker tag portion of the bundle tag,
// using the bundle version as a fallback
func (m *Manifest) getDockerTagFromBundleRef(bundleRef reference.Named) (string, error) {
	var dockerTag string
	switch v := bundleRef.(type) {
	case reference.Tagged:
		dockerTag = v.Tag()
	case reference.Named:
		ver, err := semver.NewVersion(m.Version)
		if err != nil {
			return "", errors.Wrapf(err, "could not parse the bundle version %q as a semantic version", m.Version)
		}
		// Docker tag is missing from the provided bundle tag, so default it
		// to use the manifest version prefixed with v
		// Example: bundle version is 1.0.0, so the bundle tag is v1.0.0
		dockerTag = fmt.Sprintf("v%s", ver.String())
	case reference.Digested:
		return "", errors.New("invalid bundle tag format, must be an OCI image tag")
	}

	return dockerTag, nil
}

func readFromFile(cxt *context.Context, path string) ([]byte, error) {
	if exists, _ := cxt.FileSystem.Exists(path); !exists {
		return nil, errors.Errorf("the specified porter configuration file %s does not exist", path)
	}

	data, err := cxt.FileSystem.ReadFile(path)
	return data, errors.Wrapf(err, "could not read manifest at %q", path)
}

func readFromURL(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not reach url %s", path)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return data, errors.Wrapf(err, "could not read from url %s", path)
}

func ReadManifestData(cxt *context.Context, path string) ([]byte, error) {
	if strings.HasPrefix(path, "http") {
		return readFromURL(path)
	} else {
		return readFromFile(cxt, path)
	}
}

// ReadManifest determines if specified path is a URL or a filepath.
// After reading the data in the path it returns a Manifest and any errors
func ReadManifest(cxt *context.Context, path string) (*Manifest, error) {
	data, err := ReadManifestData(cxt, path)
	if err != nil {
		return nil, err
	}

	tmplResult, err := scanManifestTemplating(data)
	if err != nil {
		return nil, err
	}

	m, err := UnmarshalManifest(cxt, data)
	if err != nil {
		return nil, err
	}

	m.ManifestPath = path
	m.TemplateVariables = tmplResult.Variables

	return m, nil
}

// templateScanResult is the result of parsing the mustache templating used in the manifest.
type templateScanResult struct {
	// Variables used in the template, e.g.  {{ bundle.parameters.NAME }}
	Variables []string
}

func scanManifestTemplating(data []byte) (templateScanResult, error) {
	const disableHtmlEscaping = true
	tmpl, err := mustache.ParseStringRaw(string(data), disableHtmlEscaping)
	if err != nil {
		return templateScanResult{}, errors.Wrap(err, "error parsing the templating used in the manifest")
	}

	tags := tmpl.Tags()
	vars := map[string]struct{}{} // Keep track of unique variable names
	for _, tag := range tags {
		if tag.Type() != mustache.Variable {
			continue
		}

		vars[tag.Name()] = struct{}{}
	}

	result := templateScanResult{
		Variables: make([]string, 0, len(vars)),
	}
	for v := range vars {
		result.Variables = append(result.Variables, v)
	}

	sort.Strings(result.Variables)
	return result, nil
}

func LoadManifestFrom(cxt *context.Context, file string) (*Manifest, error) {
	m, err := ReadManifest(cxt, file)
	if err != nil {
		return nil, err
	}

	err = m.Validate()
	if err != nil {
		return nil, err
	}

	return m, nil
}

// RequiredExtension represents a custom extension that is required
// in order for a bundle to work correctly
type RequiredExtension struct {
	Name   string
	Config map[string]interface{}
}

// UnmarshalYAML allows required extensions to either be a normal list of strings
// required:
// - docker
// or allow some entries to have config data defined
// - vpn:
//     name: mytrustednetwork
func (r *RequiredExtension) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First try to just read the mixin name
	var extNameOnly string
	err := unmarshal(&extNameOnly)
	if err == nil {
		r.Name = extNameOnly
		r.Config = nil
		return nil
	}

	// Next try to read a required extension with config defined
	extWithConfig := map[string]map[string]interface{}{}
	err = unmarshal(&extWithConfig)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal raw yaml of required extensions")
	}

	if len(extWithConfig) == 0 {
		return errors.New("required extension was empty")
	} else if len(extWithConfig) > 1 {
		return errors.New("required extension contained more than one extension")
	}

	for extName, config := range extWithConfig {
		r.Name = extName
		r.Config = config
		break // There is only one extension anyway but break for clarity
	}
	return nil
}

// Convert a parameter name to an environment variable.
// Anything more complicated should define the variable explicitly.
func ParamToEnvVar(name string) string {
	name = strings.ToUpper(name)
	fixer := strings.NewReplacer("-", "_", ".", "_")
	return fixer.Replace(name)
}

// GetParameterSourceForOutput builds the parameter source name used by Porter
// internally for wiring up an output to a parameter.
func GetParameterSourceForOutput(outputName string) string {
	return fmt.Sprintf("porter-%s-output", outputName)
}

// GetParameterSourceForDependency builds the parameter source name used by Porter
// internally for wiring up an dependency's output to a parameter.
func GetParameterSourceForDependency(ref DependencyOutputReference) string {
	return fmt.Sprintf("porter-%s-%s-dep-output", ref.Dependency, ref.Output)
}
