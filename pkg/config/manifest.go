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
	Outputs      []OutputDefinition     `yaml:"outputs,omitempty"`

	// ImageMap is a map of images referenced in the bundle. The mappings are mounted as a file at runtime to
	// /cnab/app/image-map.json. This data is not used by porter or any of the deislabs mixins, so only populate when you
	// plan on manually using this data in your own scripts.
	ImageMap map[string]MappedImage `yaml:"imageMap,omitempty"`
}

// ParameterDefinition defines a single parameter for a CNAB bundle
type ParameterDefinition struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description,omitempty"`
	Sensitive   bool      `yaml:"sensitive"`
	Destination *Location `yaml:"destination,omitempty"`

	Schema `yaml:",inline"`
}

type Schema struct {
	Type             string        `yaml:"type"`
	Default          interface{}   `yaml:"default,omitempty"`
	Enum             []interface{} `yaml:"enum,omitempty"`
	Required         bool          `yaml:"required"`
	Minimum          *float64      `yaml:"minimum,omitempty"`
	ExclusiveMinimum *float64      `yaml:"exclusiveMinimum,omitempty"`
	Maximum          *float64      `yaml:"maximum,omitempty"`
	ExclusiveMaximum *float64      `yaml:"exclusiveMaximum,omitempty"`
	MinLength        *float64      `yaml:"minLength,omitempty"`
	MaxLength        *float64      `yaml:"maxLength,omitempty"`
}

type CredentialDefinition struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Required    bool   `yaml:"required,omitempty"`

	Location `yaml:",inline"`
}

type Location struct {
	Path                string `yaml:"path,omitempty"`
	EnvironmentVariable string `yaml:"env,omitempty"`
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
	Name       string            `yaml:"name"`
	Parameters map[string]string `yaml:"parameters,omitempty"`
}

// OutputDefinition defines a single output for a CNAB
type OutputDefinition struct {
	Name        string   `yaml:"name"`
	ApplyTo     []string `yaml:"applyTo,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Sensitive   bool     `yaml:"sensitive"`

	Schema `yaml:",inline"`
}

func (od *OutputDefinition) Validate() error {
	if od.Name == "" {
		return errors.New("output name is required")
	}

	// TODO: Validate inline Schema

	return nil
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
	// TODO: Populate dependency parameters lookup

	creds := make(map[string]interface{})
	depVals["credentials"] = creds
	// TODO: Resolve dependency credentials lookup, or remove it from the template language if it shouldn't be accessible

	outputs := make(map[string]interface{})
	depVals["outputs"] = outputs
	// TODO: Populate dependency output lookups
	// TODO: Add outputs onto sensitive stuff

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

	// TODO: Temporarily disable loading dependencies while we rewrite the dependency feature
	//return c.LoadDependencies()
	return nil
}

// GetManifestDir returns the path to the directory that contains the manifest.
func (m *Manifest) GetManifestDir() string {
	return filepath.Dir(m.path)
}

// GetManifestPath returns the path where the manifest was loaded. May be a URL.
func (m *Manifest) GetManifestPath() string {
	return m.path
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

	for _, output := range m.Outputs {
		err = output.Validate()
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

func (m *Manifest) ApplyStepOutputs(step *Step, assignments []string) error {
	if m.outputs == nil {
		m.outputs = map[string]string{}
	}

	for _, assignment := range assignments {
		parts := strings.SplitN(assignment, "=", 2)
		if len(parts) != 2 {
			return errors.Errorf("invalid output assignment %v", assignment)
		}
		outvar := parts[0]
		outval := parts[1]
		if _, exists := m.outputs[outvar]; exists {
			return fmt.Errorf("output already set: %s", outvar)
		}
		m.outputs[outvar] = outval
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

type Step struct {
	runner *mixin.Runner

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
