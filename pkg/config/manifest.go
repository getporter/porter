package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/deislabs/porter/pkg/mixin"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	// path where the manifest was loaded, used to resolve local bundle references
	path            string
	outputs         map[string]string
	sensitiveValues []string

	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version,omitempty"`

	// Image is the name of the invocation image in the format REGISTRY/NAME:TAG
	Image string `yaml:"invocationImage,omitempty"`

	// BundleTag is the name of the bundle in the format REGISTRY/NAME:TAG
	BundleTag string `yaml:"tag"`

	// Dockerfile is the relative path to the Dockerfile template for the invocation image
	Dockerfile string `yaml:"dockerfile,omitempty"`

	Mixins       []string               `yaml:"mixins,omitempty"`
	Install      Steps                  `yaml:"install"`
	Uninstall    Steps                  `yaml:"uninstall"`
	Upgrade      Steps                  `yaml:"upgrade"`
	Parameters   []ParameterDefinition  `yaml:"parameters,omitempty"`
	Credentials  []CredentialDefinition `yaml:"credentials,omitempty"`
	Dependencies []*Dependency          `yaml:"dependencies,omitempty"`

	// ImageMap is a map of images referenced in the bundle. The mappings are mounted as a file at runtime to
	// /cnab/app/image-map.json. This data is not used by porter or any of the deislabs mixins, so only populate when you
	// plan on manually using this data in your own scripts.
	ImageMap map[string]MappedImage `yaml:"imageMap,omitempty"`
}

// ParameterDefinition defines a single parameter for a CNAB bundle
type ParameterDefinition struct {
	Name          string            `yaml:"name"`
	DataType      string            `yaml:"type"`
	DefaultValue  interface{}       `yaml:"default,omitempty"`
	AllowedValues []interface{}     `yaml:"allowed,omitempty"`
	Required      bool              `yaml:"required"`
	MinValue      *int              `yaml:"minValue,omitempty"`
	MaxValue      *int              `yaml:"maxValue,omitempty"`
	MinLength     *int              `yaml:"minLength,omitempty"`
	MaxLength     *int              `yaml:"maxLength,omitempty"`
	Metadata      ParameterMetadata `yaml:"metadata,omitempty"`
	Destination   *Location         `yaml:"destination,omitempty"`
	Sensitive     bool              `yaml:"sensitive"`
}

type CredentialDefinition struct {
	Name                string `yaml:"name"`
	Path                string `yaml:"path,omitempty"`
	EnvironmentVariable string `yaml:"env,omitempty"`
}

type Location struct {
	Path                string `yaml:"path,omitempty"`
	EnvironmentVariable string `yaml:"env,omitempty"`
}

// ParameterMetadata contains metadata for a parameter definition.
type ParameterMetadata struct {
	Description string `yaml:"description,omitempty"`
}

type MappedImage struct {
	Description   string         `yaml:"description"`
	ImageType     string         `yaml:"imageType"`
	Image         string         `yaml:"image"`
	OriginalImage string         `yaml:"originalImage,omitempty"`
	Digest        string         `yaml:"digest,omitempty"`
	Size          uint64         `yaml:"size,omitempty"`
	MediaType     string         `yaml:"mediaType,omitempty"`
	Platform      *ImagePlatform `yaml:"platform,omitempty"`
}

type ImagePlatform struct {
	Architecture string `yaml:"architecture,omitempty"`
	OS           string `yaml:"os,omitempty"`
}

type Dependency struct {
	// The manifest for the dependency
	m *Manifest

	Name        string             `yaml:"name"`
	Parameters  map[string]string  `yaml:"parameters,omitempty"`
	Connections []BundleConnection `yaml:"connections",omitempty`
}

func (d *Dependency) Validate() error {
	if d.Name == "" {
		return errors.New("dependency name is required")
	}
	return nil
}

func resolveParameter(pd ParameterDefinition) (string, error) {
	pe := pd.Name
	if pd.Destination == nil {
		// Porter by default sets CNAB params to name.ToUpper()
		return os.Getenv(strings.ToUpper(pe)), nil
	} else if pd.Destination.EnvironmentVariable != "" {
		return os.Getenv(pd.Destination.EnvironmentVariable), nil
	} else if pd.Destination == nil && pd.Destination.Path != "" {
		return pd.Destination.Path, nil
	}
	return "", fmt.Errorf("parameter: %s is malformed", pd.Name)

}

func resolveCredential(cd CredentialDefinition) (string, error) {
	if cd.EnvironmentVariable != "" {
		return os.Getenv(cd.EnvironmentVariable), nil
	} else if cd.Path != "" {
		return cd.Path, nil
	} else {
		return "", fmt.Errorf("credential: %s is malformed", cd.Name)
	}
}

func (d *Dependency) resolve() (map[string]interface{}, []string, error) {
	sensitiveStuff := []string{}
	depVals := make(map[string]interface{})
	params := make(map[string]interface{})
	depVals["parameters"] = params
	for _, param := range d.m.Parameters {
		val, err := resolveParameter(param)
		if err != nil {
			return nil, nil, errors.Wrap(err, fmt.Sprintf("could not handle bundle dependency %s", d.Name))
		}
		if param.Sensitive {
			sensitiveStuff = append(sensitiveStuff, val)
		}
		params[param.Name] = val
	}

	creds := make(map[string]interface{})
	depVals["credentials"] = creds
	for _, cred := range d.m.Credentials {
		val, err := resolveCredential(cred)
		if err != nil {
			return nil, nil, errors.Wrap(err, fmt.Sprintf("could not handle bundle dependency %s", d.Name))
		}
		sensitiveStuff = append(sensitiveStuff, val)
		params[cred.Name] = val
	}
	depVals["outputs"] = d.m.outputs
	for _, output := range d.m.outputs {
		sensitiveStuff = append(sensitiveStuff, output)
	}
	return depVals, sensitiveStuff, nil
}

type BundleOutput struct {
	Name                string `yaml:"name"`
	Path                string `yaml:"path"`
	EnvironmentVariable string `yaml:"env"`
}

type BundleConnection struct {
	Source      string `yaml:source`
	Destination string `yaml:destination`
	// TODO: Need to add type once it's completed in #20
}

func (c *Config) readFromFile(path string) (*Manifest, error) {
	if exists, _ := c.FileSystem.Exists(path); !exists {
		return nil, errors.Errorf("the specified porter configuration file %s does not exist", path)
	}

	data, err := c.FileSystem.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read manifest at %q", path)
	}

	m := &Manifest{}
	err = yaml.Unmarshal(data, m)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse manifest yaml in %q", path)
	}
	m.path = path

	return m, nil
}

func (c *Config) readFromURL(path string) (*Manifest, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not reach url %s", path)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read from url %s", path)
	}

	m := &Manifest{}
	err = yaml.Unmarshal(body, m)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse manifest yaml in %q", path)
	}
	m.path = path

	return m, nil
}

// ReadManifest determines if specified path is a URL or a filepath.
// After reading the data in the path it returns a Manifest and any errors
func (c *Config) ReadManifest(path string) (*Manifest, error) {
	if strings.HasPrefix(path, "http") {
		return c.readFromURL(path)
	}

	return c.readFromFile(path)
}

func (c *Config) LoadManifest() error {
	return c.LoadManifestFrom(Name)
}

func (c *Config) LoadManifestFrom(file string) error {
	m, err := c.ReadManifest(file)
	if err != nil {
		return err
	}

	c.Manifest = m

	err = c.Manifest.Validate()
	if err != nil {
		return err
	}

	return c.LoadDependencies()
}

// GetManifestDir returns the path to the directory that contains the manifest.
func (m *Manifest) GetManifestDir() string {
	return filepath.Dir(m.path)
}

func (c *Config) LoadDependencies() error {
	for _, dep := range c.Manifest.Dependencies {
		path, err := c.GetBundleManifestPath(dep.Name)
		if err != nil {
			return err
		}

		dep.m, err = c.ReadManifest(path)
		if err != nil {
			return err
		}

		err = dep.m.Validate()
		if err != nil {
			return err
		}

		err = c.Manifest.MergeDependency(dep)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manifest) Validate() error {
	var result error

	if strings.ToLower(m.Dockerfile) == "dockerfile" {
		return errors.New("Dockerfile template cannot be named 'Dockerfile' because that is the filename generated during porter build")
	}

	if len(m.Mixins) == 0 {
		result = multierror.Append(result, errors.New("no mixins declared"))
	}

	if m.Install == nil {
		result = multierror.Append(result, errors.New("no install action defined"))
	}
	err := m.Install.Validate(m)
	if err != nil {
		result = multierror.Append(result, err)
	}

	if m.Uninstall == nil {
		result = multierror.Append(result, errors.New("no uninstall action defined"))
	}
	err = m.Uninstall.Validate(m)
	if err != nil {
		result = multierror.Append(result, err)
	}

	for _, dep := range m.Dependencies {
		err = dep.Validate()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (m *Manifest) GetSensitiveValues() []string {
	if m.sensitiveValues == nil {
		return []string{}
	}
	return m.sensitiveValues
}

func (m *Manifest) GetSteps(action Action) (Steps, error) {
	var steps Steps
	switch action {
	case ActionInstall:
		steps = m.Install
	case ActionUninstall:
		steps = m.Uninstall
	case ActionUpgrade:
		steps = m.Upgrade
	}

	if len(steps) == 0 {
		return nil, errors.Errorf("unsupported action: %q", action)
	}

	return steps, nil
}

func (m *Manifest) ApplyOutputs(step *Step, assignments []string) error {
	scope := m
	if step.dep != nil {
		scope = step.dep.m
	}
	if scope.outputs == nil {
		scope.outputs = map[string]string{}
	}

	for _, assignment := range assignments {
		parts := strings.SplitN(assignment, "=", 2)
		if len(parts) != 2 {
			return errors.Errorf("invalid output assignment %v", assignment)
		}
		outvar := parts[0]
		outval := parts[1]
		if _, exists := scope.outputs[outvar]; exists {
			return fmt.Errorf("output already set: %s", outvar)
		}
		scope.outputs[outvar] = outval
	}

	return nil
}

func (m *Manifest) MergeDependency(dep *Dependency) error {
	// include any unique credentials from the dependency
	for i, cred := range dep.m.Credentials {
		dupe := false
		for _, x := range m.Credentials {
			if cred.Name == x.Name {
				result, err := mergeCredentials(x, cred)
				if err != nil {
					return err
				}

				// Allow for having the same credential populated both as an env var and a file
				dep.m.Credentials[i].EnvironmentVariable = result.EnvironmentVariable
				dep.m.Credentials[i].Path = result.Path
				dupe = true
				break
			}
		}
		if !dupe {
			m.Credentials = append(m.Credentials, cred)
		}
	}

	err := m.MergeParameters(dep)
	if err != nil {
		return err
	}

	// prepend the dependency's mixins
	m.Mixins = prependMixins(dep.m.Mixins, m.Mixins)

	// prepend dependency's install steps
	m.MergeInstall(dep)

	// append uninstall steps so that we unroll it in dependency order (i.e. uninstall wordpress before we delete the database)
	m.MergeUninstall(dep)

	// prepend dependency's upgrade steps
	m.MergeUpgrade(dep)

	return nil
}

func (m *Manifest) MergeInstall(dep *Dependency) {
	dep.m.Install.setDependency(dep)

	m.Install = prependSteps(dep.m.Install, m.Install)
}

func (m *Manifest) MergeUpgrade(dep *Dependency) {
	dep.m.Upgrade.setDependency(dep)

	m.Upgrade = prependSteps(dep.m.Upgrade, m.Upgrade)
}

func (m *Manifest) MergeUninstall(dep *Dependency) {
	dep.m.Uninstall.setDependency(dep)

	m.Uninstall = append(m.Uninstall, dep.m.Uninstall...)
}

func prependSteps(s1, s2 Steps) Steps {
	result := make(Steps, len(s2)+len(s1))
	copy(result[:len(s2)], s1)
	copy(result[len(s2):], s2)

	return result
}

func prependMixins(m1, m2 []string) []string {
	mixins := make([]string, len(m1), len(m1)+len(m2))
	copy(mixins, m1)
	for _, m := range m2 {
		dupe := false
		for _, x := range m1 {
			if m == x {
				dupe = true
				break
			}
		}
		if !dupe {
			mixins = append(mixins, m)
		}
	}
	return mixins
}

func mergeCredentials(c1, c2 CredentialDefinition) (CredentialDefinition, error) {
	result := CredentialDefinition{Name: c1.Name}

	if c1.Name != c2.Name {
		return result, fmt.Errorf("cannot merge credentials that don't have the same name: %s and %s", c1.Name, c2.Name)
	}

	if c1.Path != "" && c2.Path != "" && c1.Path != c2.Path {
		return result, fmt.Errorf("cannot merge credential %s: conflict on path", c1.Name)
	}
	result.Path = c1.Path
	if result.Path == "" {
		result.Path = c2.Path
	}

	if c1.EnvironmentVariable != "" && c2.EnvironmentVariable != "" && c1.EnvironmentVariable != c2.EnvironmentVariable {
		return result, fmt.Errorf("cannot merge credential %s: conflict on environment variable", c1.Name)
	}
	result.EnvironmentVariable = c1.EnvironmentVariable
	if result.EnvironmentVariable == "" {
		result.EnvironmentVariable = c2.EnvironmentVariable
	}

	return result, nil
}

func (m *Manifest) MergeParameters(dep *Dependency) error {
	// include any unique parameters from the dependency
	for _, param := range dep.m.Parameters {
		dupe := false
		for _, x := range m.Parameters {
			if param.Name == x.Name {
				dupe = true
				break
			}
		}
		if !dupe {
			m.Parameters = append(m.Parameters, param)
		}
	}

	// Default the bundle parameters from any hard-coded values set in the dependencies
	for depP, defaultValue := range dep.Parameters {
		for i, param := range m.Parameters {
			if param.Name == depP {
				m.Parameters[i].DefaultValue = defaultValue
			}
		}
	}

	return nil
}

type Steps []*Step

func (s Steps) Validate(m *Manifest) error {
	for _, step := range s {
		if step != nil {
			err := step.Validate(m)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// setDependency remembers the dependency that generated the step
func (s Steps) setDependency(dep *Dependency) {
	for _, s := range s {
		s.dep = dep
	}
}

type Step struct {
	runner *mixin.Runner
	dep    *Dependency // The dependency that owns this step

	Data map[string]interface{} `yaml:",inline"`
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

type StepOutput struct {
	// The final value of the output returned by the mixin after executing
	value string

	Name string                 `yaml:"name"`
	Data map[string]interface{} `yaml:",inline"`
}

func (s *Step) Validate(m *Manifest) error {
	if len(s.Data) == 0 {
		return errors.New("no mixin specified")
	}
	if len(s.Data) > 1 {
		return errors.New("more than one mixin specified")
	}

	mixinDeclared := false
	mixinType := s.GetMixinName()
	for _, mixin := range m.Mixins {
		if mixin == mixinType {
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

func (s *Step) GetMixinName() string {
	var mixinName string
	for k := range s.Data {
		mixinName = k
	}
	return mixinName
}

func (m *Manifest) buildSourceData() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	m.sensitiveValues = []string{}
	bundle := make(map[string]interface{})
	data["bundle"] = bundle

	// Enable interpolation of manifest/bundle name via bundle.name
	bundle["name"] = m.Name

	params := make(map[string]interface{})
	bundle["parameters"] = params
	for _, param := range m.Parameters {
		//pe := strings.ToUpper(param.Name)
		pe := param.Name
		var val string
		val, err := resolveParameter(param)
		if err != nil {
			return nil, err
		}
		if param.Sensitive {
			m.sensitiveValues = append(m.sensitiveValues, val)
		}
		params[pe] = val
	}

	creds := make(map[string]interface{})
	bundle["credentials"] = creds
	for _, cred := range m.Credentials {
		pe := cred.Name
		val, err := resolveCredential(cred)
		if err != nil {
			return nil, err
		}
		m.sensitiveValues = append(m.sensitiveValues, val)
		creds[pe] = val
	}
	bundle["outputs"] = m.outputs
	for _, output := range m.outputs {
		m.sensitiveValues = append(m.sensitiveValues, output)
	}
	deps := make(map[string]interface{})
	bundle["dependencies"] = deps
	for _, dependency := range m.Dependencies {
		dep, sensitives, err := dependency.resolve()
		if err != nil {
			return nil, err
		}
		deps[dependency.Name] = dep
		m.sensitiveValues = append(m.sensitiveValues, sensitives...)
	}
	return data, nil
}

// ResolveStep will walk through the Step's data and resolve any placeholder
// data using the definitions in the manifest, like parameters or credentials.
func (m *Manifest) ResolveStep(step *Step) error {

	mustache.AllowMissingVariables = false
	sourceData, err := m.buildSourceData()
	if err != nil {
		return errors.Wrap(err, "unable to resolve step: unable to populate source data")
	}
	payload, err := yaml.Marshal(step)
	if err != nil {
		return err
	}
	rendered, err := mustache.Render(string(payload), sourceData)
	if err != nil {
		return errors.Wrap(err, "unable to resolve step: unable to render template values")
	}
	err = yaml.Unmarshal([]byte(rendered), step)
	if err != nil {
		return errors.Wrap(err, "unable to resolve step: invalid step yaml")
	}
	return nil
}
