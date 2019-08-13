package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type RuntimeManifest struct {
	*Manifest

	// path where the manifest was loaded, used to resolve local bundle references
	path            string
	outputs         map[string]string
	sensitiveValues []string
}

func NewRuntimeManifest(m *Manifest, path string) *RuntimeManifest {
	return &RuntimeManifest{
		Manifest: m,
		path:     path,
	}
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

// GetManifestDir returns the path to the directory that contains the manifest.
func (m *RuntimeManifest) GetManifestDir() string {
	return filepath.Dir(m.path)
}

// GetManifestPath returns the path where the manifest was loaded. May be a URL.
func (m *RuntimeManifest) GetManifestPath() string {
	return m.path
}

func (m *RuntimeManifest) GetSensitiveValues() []string {
	if m.sensitiveValues == nil {
		return []string{}
	}
	return m.sensitiveValues
}

func (m *RuntimeManifest) GetSteps(action Action) (Steps, error) {
	switch action {
	case ActionInstall:
		return m.Install, nil
	case ActionUninstall:
		return m.Uninstall, nil
	case ActionUpgrade:
		return m.Upgrade, nil
	default:
		customAction, ok := m.CustomActions[string(action)]
		if !ok {
			actions := make([]string, 0, len(m.CustomActions))
			for a := range m.CustomActions {
				actions = append(actions, a)
			}
			return nil, errors.Errorf("unsupported action %q, custom actions are defined for: %s", action, strings.Join(actions, ", "))
		}
		return customAction, nil
	}
}

func (m *RuntimeManifest) ApplyStepOutputs(step *Step, assignments map[string]string) error {
	if m.outputs == nil {
		m.outputs = map[string]string{}
	}

	for outvar, outval := range assignments {
		if _, exists := m.outputs[outvar]; exists {
			return fmt.Errorf("output already set: %s", outvar)
		}
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
	for name, dependency := range m.Dependencies {
		dep, sensitives, err := dependency.resolve()
		if err != nil {
			return nil, err
		}
		deps[name] = dep
		m.sensitiveValues = append(m.sensitiveValues, sensitives...)
	}
	return data, nil
}

// ResolveStep will walk through the Step's data and resolve any placeholder
// data using the definitions in the manifest, like parameters or credentials.
func (m *RuntimeManifest) ResolveStep(step *Step) error {

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
