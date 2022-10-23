package manifest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/Masterminds/semver/v3"
	"github.com/cbroglie/mustache"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/hashicorp/go-multierror"
	"github.com/opencontainers/go-digest"
)

const (
	invalidStepErrorFormat = "validation of action \"%s\" failed: %w"

	// TemplateDelimiterPrefix must be present at the beginning of any porter.yaml
	// that wants to use ${} as the template delimiter instead of the mustache
	// default of {{}}.
	TemplateDelimiterPrefix = "{{=${ }=}}\n"
)

var (
	// SupportedSchemaVersions is the Porter manifest (porter.yaml) schema
	// versions supported by this version of Porter, specified as a semver range.
	// When the Manifest structure is changed, this field should be incremented.
	SupportedSchemaVersions, _ = semver.NewConstraint("1.0.0-alpha.1 || 1.0.0")

	// DefaultSchemaVersion is the most recently supported schema version.
	// When the Manifest structure is changed, this field should be incremented.
	DefaultSchemaVersion = semver.MustParse("1.0.0")
)

type Manifest struct {
	// ManifestPath is location to the original, user-supplied manifest, such as the path on the filesystem or a url
	ManifestPath string `yaml:"-"`

	// TemplateVariables are the variables used in the templating, e.g. bundle.parameters.NAME, or bundle.outputs.NAME
	TemplateVariables []string `yaml:"-"`

	// SchemaVersion is a semver value that indicates which version of the porter.yaml schema is used in the file.
	SchemaVersion string `yaml:"schemaVersion"`
	Name          string `yaml:"name,omitempty"`
	Description   string `yaml:"description,omitempty"`
	Version       string `yaml:"version,omitempty"`

	Maintainers []MaintainerDefinition `yaml:"maintainers,omitempty"`

	// Registry is the OCI registry and org/subdomain for the bundle
	Registry string `yaml:"registry,omitempty"`

	// Reference is the optional, full bundle reference
	// in the format REGISTRY/NAME or REGISTRY/NAME:TAG
	Reference string `yaml:"reference,omitempty"`

	// DockerTag is the Docker tag portion of the published invocation
	// image and bundle.  It will only be set at time of publishing.
	DockerTag string `yaml:"-"`

	// Image is the name of the invocation image in the format REGISTRY/NAME:TAG
	// It doesn't map to any field in the manifest as it has been deprecated
	// and isn't meant to be user-specified
	Image string `yaml:"-"`

	// Dockerfile is the relative path to the Dockerfile template for the invocation image
	Dockerfile string `yaml:"dockerfile,omitempty"`

	Mixins []MixinDeclaration `yaml:"mixins,omitempty"`

	Install   Steps `yaml:"install"`
	Uninstall Steps `yaml:"uninstall"`
	Upgrade   Steps `yaml:"upgrade"`

	Custom                  CustomDefinitions                 `yaml:"custom,omitempty"`
	CustomActions           map[string]Steps                  `yaml:"-"`
	CustomActionDefinitions map[string]CustomActionDefinition `yaml:"customActions,omitempty"`

	StateBag     StateBag              `yaml:"state,omitempty"`
	Parameters   ParameterDefinitions  `yaml:"parameters,omitempty"`
	Credentials  CredentialDefinitions `yaml:"credentials,omitempty"`
	Dependencies Dependencies          `yaml:"dependencies,omitempty"`
	Outputs      OutputDefinitions     `yaml:"outputs,omitempty"`

	// ImageMap is a map of images referenced in the bundle. If an image relocation mapping is later provided, that
	// will be mounted at as a file at runtime to /cnab/app/relocation-mapping.json.
	ImageMap map[string]MappedImage `yaml:"images,omitempty"`

	Required []RequiredExtension `yaml:"required,omitempty"`
}

func (m *Manifest) Validate(cxt *portercontext.Context, strategy schema.CheckStrategy) error {
	var result error

	err := m.validateMetadata(cxt, strategy)
	if err != nil {
		return err
	}

	err = m.SetDefaults()
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
		result = multierror.Append(result, fmt.Errorf(invalidStepErrorFormat, "install", err))
	}

	if m.Uninstall == nil {
		result = multierror.Append(result, errors.New("no uninstall action defined"))
	}
	err = m.Uninstall.Validate(m)
	if err != nil {
		result = multierror.Append(result, fmt.Errorf(invalidStepErrorFormat, "uninstall", err))
	}

	for actionName, steps := range m.CustomActions {
		err := steps.Validate(m)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf(invalidStepErrorFormat, actionName, err))
		}
	}

	for _, dep := range m.Dependencies.Requires {
		err = dep.Validate(cxt)
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

func (m *Manifest) validateMetadata(cxt *portercontext.Context, strategy schema.CheckStrategy) error {
	if warnOnly, err := schema.ValidateSchemaVersion(strategy, SupportedSchemaVersions, m.SchemaVersion, DefaultSchemaVersion); err != nil {
		if warnOnly {
			fmt.Fprintln(cxt.Err, err)
		} else {
			return err
		}
	}

	if m.Name == "" {
		return errors.New("bundle name must be set")
	}

	if m.Registry == "" && m.Reference == "" {
		return errors.New("a registry or reference value must be provided")
	}

	if m.Reference != "" && m.Registry != "" {
		fmt.Fprintf(cxt.Out, "WARNING: both registry and reference were provided; "+
			"using the reference value of %s for the bundle reference\n", m.Reference)
	}

	// Allow for the user to have specified the version with a leading v prefix but save it as
	// proper semver
	if m.Version != "" {
		v, err := semver.NewVersion(m.Version)
		if err != nil {
			return fmt.Errorf("version %q is not a valid semver value: %w", m.Version, err)
		}
		m.Version = v.String()
	}
	return nil
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

type CustomDefinitions map[string]interface{}

func (cd *CustomDefinitions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	raw, err := yaml.UnmarshalMap(unmarshal)
	if err != nil {
		return err
	}
	*cd = raw
	return nil
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

var _ bundle.Scoped = &ParameterDefinition{}

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

	// IsState identifies if the parameter was generated from a state variable
	IsState bool `yaml:"-"`
}

func (pd *ParameterDefinition) GetApplyTo() []string {
	return pd.ApplyTo
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

	// Validate the Parameter Definition schema itself
	if _, err := pdCopy.Schema.ValidateSchema(); err != nil {
		return multierror.Append(result, fmt.Errorf("encountered an error while validating definition for parameter %q: %w", pdCopy.Name, err))
	}

	if pdCopy.Default != nil {
		schemaValidationErrs, err := pdCopy.Schema.Validate(pdCopy.Default)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("encountered error while validating parameter %s: %w", pdCopy.Name, err))
		}
		for _, schemaValidationErr := range schemaValidationErrs {
			result = multierror.Append(result, fmt.Errorf("encountered an error validating the default value %v for parameter %q: %s", pdCopy.Default, pdCopy.Name, schemaValidationErr.Error))
		}
	}

	return result.ErrorOrNil()
}

// DeepCopy copies a ParameterDefinition and returns the copy
func (pd *ParameterDefinition) DeepCopy() *ParameterDefinition {
	p2 := *pd
	p2.ApplyTo = make([]string, len(pd.ApplyTo))
	copy(p2.ApplyTo, pd.ApplyTo)
	return &p2
}

// AppliesTo returns a boolean value specifying whether or not
// the Parameter applies to the provided action
func (pd *ParameterDefinition) AppliesTo(action string) bool {
	return bundle.AppliesTo(pd, action)
}

// exemptFromInstall returns true if a parameter definition:
//   - has an output source (which will not exist prior to install)
//   - doesn't already have applyTo specified
//   - doesn't have a default value
func (pd *ParameterDefinition) exemptFromInstall() bool {
	return pd.Source.Output != "" && pd.ApplyTo == nil && pd.Default == nil
}

// UpdateApplyTo updates a parameter definition's applyTo section
// based on the provided manifest
func (pd *ParameterDefinition) UpdateApplyTo(m *Manifest) {
	if pd.exemptFromInstall() {
		applyTo := []string{cnab.ActionUninstall}
		// The core action "Upgrade" is technically still optional
		// so only add it if it is declared in the manifest
		if m.Upgrade != nil {
			applyTo = append(applyTo, cnab.ActionUpgrade)
		}
		// Add all custom actions
		for action := range m.CustomActions {
			applyTo = append(applyTo, action)
		}
		sort.Strings(applyTo)
		pd.ApplyTo = applyTo
	}
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

	// Required specifies if the credential must be specified for applicable actions. Defaults to true.
	Required bool `yaml:"required,omitempty"`

	// ApplyTo lists the actions to which the credential applies. When unset, defaults to all actions.
	ApplyTo []string `yaml:"applyTo,omitempty"`

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
// - helm3
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
		return fmt.Errorf("could not unmarshal raw yaml of mixin declarations: %w", err)
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
// - helm3
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
		if _, err := digest.Parse(mi.Digest); err != nil {
			return err
		}
	}

	if _, err := cnab.ParseOCIReference(mi.Repository); err != nil {
		return err
	}

	return nil
}

func (mi *MappedImage) ToOCIReference() (cnab.OCIReference, error) {
	ref, err := cnab.ParseOCIReference(mi.Repository)
	if err != nil {
		return cnab.OCIReference{}, err
	}

	if mi.Digest != "" {
		refWithDigest, err := ref.WithDigest(digest.Digest(mi.Digest))
		if err != nil {
			return cnab.OCIReference{}, fmt.Errorf("failed to create a new reference with digest for repository %s: %w", mi.Repository, err)
		}

		return refWithDigest, nil
	}

	if mi.Tag != "" {
		refWithTag, err := ref.WithTag(mi.Tag)
		if err != nil {
			return cnab.OCIReference{}, fmt.Errorf("failed to create a new reference with tag for repository %s: %w", mi.Repository, err)
		}

		return refWithTag, nil
	}

	return ref, nil
}

type Dependencies struct {
	Requires []*Dependency `yaml:"requires,omitempty"`
}

type Dependency struct {
	Name string `yaml:"name"`

	Bundle BundleCriteria `yaml:"bundle"`

	Parameters map[string]string `yaml:"parameters,omitempty"`
}

type BundleCriteria struct {
	// Reference is the full bundle reference for the dependency
	// in the format REGISTRY/NAME:TAG
	Reference string `yaml:"reference"`

	// "When constraint checking is used for checks or validation
	// it will follow a different set of rules that are common for ranges with tools like npm/js and Rust/Cargo.
	// This includes considering prereleases to be invalid if the ranges does not include one.
	// If you want to have it include pre-releases a simple solution is to include -0 in your range."
	// https://github.com/Masterminds/semver/blob/master/README.md#checking-version-constraints
	Version string `yaml:"version,omitempty"`
}

func (d *Dependency) Validate(cxt *portercontext.Context) error {
	if d.Name == "" {
		return errors.New("dependency name is required")
	}

	if d.Bundle.Reference == "" {
		return fmt.Errorf("reference is required for dependency %q", d.Name)
	}

	if strings.Contains(d.Bundle.Reference, ":") && len(d.Bundle.Version) > 0 {
		return fmt.Errorf("reference for dependency %q can only specify REGISTRY/NAME when version ranges are specified", d.Name)
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

	// IsState identifies if the output was generated from a state variable
	IsState bool `yaml:"-"`
}

// DeepCopy copies a ParameterDefinition and returns the copy
func (od *OutputDefinition) DeepCopy() *OutputDefinition {
	o2 := *od
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

	// Validate the Output Definition schema itself
	if _, err := odCopy.Schema.ValidateSchema(); err != nil {
		return multierror.Append(result, fmt.Errorf("encountered an error while validating definition for output %q: %w", odCopy.Name, err))
	}

	if odCopy.Default != nil {
		schemaValidationErrs, err := odCopy.Schema.Validate(odCopy.Default)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("encountered error while validating output %s: %w", odCopy.Name, err))
		}
		for _, schemaValidationErr := range schemaValidationErrs {
			result = multierror.Append(result, fmt.Errorf("encountered an error validating the default value %v for output %q: %s", odCopy.Default, odCopy.Name, schemaValidationErr.Error))
		}
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
		return fmt.Errorf("mixin (%s) was not declared", mixinType)
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
	d, ok := children.(map[string]interface{})["description"]
	if !ok {
		return "", nil
	}
	desc, ok := d.(string)
	if !ok {
		return "", fmt.Errorf("invalid description type (%T) for mixin step (%s)", desc, mixinName)
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

func UnmarshalManifest(cxt *portercontext.Context, manifestData []byte) (*Manifest, error) {
	// Unmarshal the manifest into the normal struct
	manifest := &Manifest{}
	err := yaml.Unmarshal(manifestData, &manifest)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling the typed manifest: %w", err)
	}

	// Do a second pass to identify custom actions, which don't have yaml tags since they are dynamic
	// 1. Marshal the manifest a second time into a plain map
	// 2. Remove keys for fields that are already mapped with yaml tags
	// 3. Anything left is a custom action

	// Marshal the manifest into an untyped map
	unmappedData := make(map[string]interface{})
	err = yaml.Unmarshal(manifestData, &unmappedData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling the untyped manifest: %w", err)
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
		// Delete known deprecated fields with no yaml tags
		if key == "invocationImage" || key == "tag" {
			fmt.Fprintf(cxt.Out, "WARNING: The %q field has been deprecated and can no longer be user-specified; ignoring.\n", key)
			delete(unmappedData, key)
		}
	}

	// Marshal the remaining keys in the unmappedData as custom actions and append them to the typed manifest
	manifest.CustomActions = make(map[string]Steps, len(unmappedData))
	for key, chunk := range unmappedData {
		chunkData, err := yaml.Marshal(chunk)
		if err != nil {
			return nil, fmt.Errorf("error remarshaling custom action %s: %w", key, err)
		}

		steps := Steps{}
		err = yaml.Unmarshal(chunkData, &steps)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling custom action %s: %w", key, err)
		}

		manifest.CustomActions[key] = steps
	}

	return manifest, nil
}

// SetDefaults updates the manifest with default values where not populated
func (m *Manifest) SetDefaults() error {
	return m.SetInvocationImageAndReference("")
}

// SetInvocationImageAndReference sets the invocation image name and the
// bundle reference on the manifest per the provided reference or via the
// registry or name values on the manifest.
func (m *Manifest) SetInvocationImageAndReference(ref string) error {
	if ref != "" {
		m.Reference = ref
	}

	if m.Reference == "" && m.Registry != "" {
		repo, err := cnab.ParseOCIReference(path.Join(m.Registry, m.Name))
		if err != nil {
			return fmt.Errorf("invalid bundle reference %s: %w", path.Join(m.Registry, m.Name), err)
		}
		m.Reference = repo.Repository()
	}

	bundleRef, err := cnab.ParseOCIReference(m.Reference)
	if err != nil {
		return fmt.Errorf("invalid bundle reference %s: %w", m.Reference, err)
	}

	dockerTag, err := m.getDockerTagFromBundleRef(bundleRef)
	if err != nil {
		return fmt.Errorf("unable to derive docker tag from bundle reference %q: %w", m.Reference, err)
	}

	// If the docker tag is initially missing from bundleTag, update with
	// returned dockerTag
	if !bundleRef.HasTag() {
		bundleRef, err = bundleRef.WithTag(dockerTag)
		if err != nil {
			return fmt.Errorf("could not set bundle tag to %q: %w", dockerTag, err)
		}
		m.Reference = bundleRef.String()
	}

	installerImage, err := cnab.CalculateTemporaryImageTag(bundleRef)
	if err != nil {
		return err
	}

	m.Image = installerImage.String()
	return nil
}

// getDockerTagFromBundleRef returns the Docker tag portion of the bundle tag,
// using the bundle version as a fallback
func (m *Manifest) getDockerTagFromBundleRef(bundleRef cnab.OCIReference) (string, error) {
	// If the manifest has a DockerTag override already set (e.g. on publish), use this
	if m.DockerTag != "" {
		return m.DockerTag, nil
	}

	if bundleRef.HasTag() {
		return bundleRef.Tag(), nil
	}

	if bundleRef.HasDigest() {
		return "", errors.New("invalid bundle tag format, must be an OCI image tag")
	}

	// Docker tag is missing from the provided bundle tag, so default it
	// to use the manifest version prefixed with v
	// Example: bundle version is 1.0.0, so the bundle tag is v1.0.0
	cleanTag := strings.ReplaceAll(m.Version, "+", "_") // Semver may include a + which is not allowed in a docker tag, e.g. v1.0.0-alpha.1+buildmetadata, change that to v1.0.0-alpha.1_buildmetadata
	return fmt.Sprintf("v%s", cleanTag), nil
}

// ResolvePath resolves a path specified in the Porter manifest into
// an absolute path, assuming the current directory is /cnab/app.
// Returns an empty string when the specified value is empty.
func ResolvePath(value string) string {
	if value == "" {
		return ""
	}

	if path.IsAbs(value) {
		return value
	}

	return path.Join("/cnab/app", value)
}

func readFromFile(cxt *portercontext.Context, path string) ([]byte, error) {
	if exists, _ := cxt.FileSystem.Exists(path); !exists {
		return nil, fmt.Errorf("the specified porter configuration file %s does not exist", path)
	}

	data, err := cxt.FileSystem.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read manifest at %q: %w", path, err)
	}
	return data, nil
}

func readFromURL(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, fmt.Errorf("could not reach url %s: %w", path, err)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read from url %s: %w", path, err)
	}
	return data, nil
}

func ReadManifestData(cxt *portercontext.Context, path string) ([]byte, error) {
	if strings.HasPrefix(path, "http") {
		return readFromURL(path)
	} else {
		return readFromFile(cxt, path)
	}
}

// ReadManifest determines if specified path is a URL or a filepath.
// After reading the data in the path it returns a Manifest and any errors
func ReadManifest(cxt *portercontext.Context, path string) (*Manifest, error) {
	data, err := ReadManifestData(cxt, path)
	if err != nil {
		return nil, err
	}

	m, err := UnmarshalManifest(cxt, data)
	if err != nil {
		return nil, fmt.Errorf("unsupported property set or a custom action is defined incorrectly: %w", err)
	}

	tmplResult, err := m.scanManifestTemplating(data)
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

func (m *Manifest) GetTemplatePrefix() string {
	if m.SchemaVersion == "" {
		// Super-old bundles use the mustache default
		return ""
	}

	// In 1.0.0-alpha.2+, the prefix is ${}. Beforehand it was {{}}
	v, err := semver.NewVersion(m.SchemaVersion)
	if err == nil {
		if v.GreaterThan(semver.MustParse("v1.0.0-alpha.1")) {
			// Change the delimiter
			return TemplateDelimiterPrefix
		}
	}

	// Fallback to the mustache default if we can't determine the schema version
	return ""
}

func (m *Manifest) scanManifestTemplating(data []byte) (templateScanResult, error) {
	const disableHtmlEscaping = true
	templateSrc := m.GetTemplatePrefix() + string(data)
	tmpl, err := mustache.ParseStringRaw(templateSrc, disableHtmlEscaping)
	if err != nil {
		return templateScanResult{}, fmt.Errorf("error parsing the templating used in the manifest: %w", err)
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

// LoadManifestFrom reads and validates the manifest at the specified location,
// and returns a populated Manifest structure.
func LoadManifestFrom(ctx context.Context, config *config.Config, file string) (*Manifest, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	m, err := ReadManifest(config.Context, file)
	if err != nil {
		return nil, err
	}

	strategy := config.GetSchemaCheckStrategy(ctx)
	if err = m.Validate(config.Context, strategy); err != nil {
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
		return fmt.Errorf("could not unmarshal raw yaml of required extensions: %w", err)
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

type MaintainerDefinition struct {
	Name  string `yaml:"name,omitempty"`
	Email string `yaml:"email,omitempty"`
	Url   string `yaml:"url,omitempty"`
}

// StateBag is the set of state files and variables that Porter should
// track between bundle executions.
type StateBag []StateVariable

type StateVariable struct {
	// Name of the state variable
	Name string `yaml:"name"`

	// Description of the state variable and how it's used by the bundle
	Description string `yaml:"description,omitempty"`

	// Mixin is the name of the mixin that manages the state variable.
	Mixin string `yaml:"mixin,omitempty"`

	// Location defines where the state variable is located in the bundle.
	Location `yaml:",inline"`
}
