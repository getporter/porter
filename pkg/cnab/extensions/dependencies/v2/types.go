package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"get.porter.sh/porter/pkg/secrets"
)

// Dependencies describes the set of custom extension metadata associated with the dependencies spec
// https://github.com/cnabio/cnab-spec/blob/master/500-CNAB-dependencies.md
type Dependencies struct {
	// Requires is a list of bundles required by this bundle
	Requires map[string]Dependency `json:"requires,omitempty" mapstructure:"requires"`
}

/*
dependencies:
  requires: # dependencies are always created in the current namespace, never global though they can match globally?
    mysql:
      bundle:
        reference: getporter/mysql:v1.0.2
        version: 1.x
        interface: # Porter defaults the interface based on usage
          reference: getporter/generic-mysql-interface:v1.0.0 # point to an interface bundle to be more specific
          document: # add extra interface requirements
            outputs:
              - $id: "mysql-5.7-connection-string" # match on something other than name, so that outputs with different names can be reused
      installation:
        labels: # labels applied to the installation if created
          app: myapp
          installation: {{ installation.name }} # exclusive resource
        criteria: # criteria for reusing an existing installation, by default must be the same bundle, labels and allows global
          matchInterface: true # only match the interface, not the bundle too
          matchNamespace: true # must be in the same namespace, disallow global
          ignoreLabels: true # allow different labels
*/

// Dependency describes a dependency on another bundle
type Dependency struct {
	// Name of the dependency
	// This is used internally but isn't persisted to bundle.json
	Name string `json:"-" mapstructure:"-"`

	// Bundle is the location of the bundle in a registry, for example REGISTRY/NAME:TAG
	Bundle string `json:"bundle" mapstructure:"bundle"`

	// Version is a set of allowed versions
	Version string `json:"version,omitempty" mapstructure:"version"`

	Interface *DependencyInterface `json:"interface,omitempty" mapstructure:"interface,omitempty"`

	Installation *DependencyInstallation `json:"installation,omitempty" mapstructure:"installation,omitempty"`

	Parameters  map[string]DependencySource `json:"parameters,omitempty" mapstructure:"parameters,omitempty"`
	Credentials map[string]DependencySource `json:"credentials,omitempty" mapstructure:"credentials,omitempty"`
}

type DependencySource struct {
	Value      string `json:"value,omitempty" mapstructure:"value,omitempty"`
	Dependency string `json:"dependency,omitempty" mapstructure:"dependency,omitempty"`
	Credential string `json:"credential,omitempty" mapstructure:"credential,omitempty"`
	Parameter  string `json:"parameter,omitempty" mapstructure:"parameter,omitempty"`
	Output     string `json:"output,omitempty" mapstructure:"output,omitempty"`
}

// ignore template syntax, ${...}, if found
var dependencySourceWiringRegex = regexp.MustCompile(`(\s*\$\{\s*)?bundle(\.dependencies\.([^.]+))?\.([^.]+)\.([^\s\}]+)(\s*\}\s*)?`)

// ParseDependencySource identifies the components specified in a wiring string.
func ParseDependencySource(value string) (DependencySource, error) {
	matches := dependencySourceWiringRegex.FindStringSubmatch(value)

	// If it doesn't match our wiring syntax, assume that it is a hard coded value
	if matches == nil || len(matches) < 5 {
		return DependencySource{Value: value}, nil
	}

	dependencyName := matches[3] // bundle.dependencies.DEPENDENCY_NAME
	itemType := matches[4]       // bundle.dependencies.dependency_name.PARAMETERS.name or bundle.OUTPUTS.name
	itemName := matches[5]       // bundle.dependencies.dependency_name.parameters.NAME or bundle.outputs.NAME

	result := DependencySource{Dependency: dependencyName}
	switch itemType {
	case "parameters":
		result.Parameter = itemName
	case "credentials":
		result.Credential = itemName
	case "outputs":
		// Cannot pass the root bundle's output to a dependency
		// Check that we are attempting to pass another dependency's output
		if dependencyName == "" {
			return DependencySource{}, errors.New("cannot pass the root bundle output to a dependency")
		}
		result.Output = itemName
	}
	return result, nil
}

func (s DependencySource) AsWorkflowStrategy(name string, parentJob string) secrets.Strategy {
	strategy := secrets.Strategy{
		Name: name,
		Source: secrets.Source{
			// bundle.dependencies.DEP.outputs.OUTPUT -> workflow.jobs.JOB.outputs.OUTPUT
			// TODO(PEP003): Figure out if we need a job id, or if we can do okay with just a job key that we resolve to a run later
			Value: s.AsWorkflowWiring(parentJob),
		},
	}

	// TODO(PEP003): Are other strategies valid when talking about dependency wiring? Or can we only pass hard-coded values and data from a previous job?
	if s.Value != "" {
		strategy.Source.Key = "value"
	} else {
		strategy.Source.Key = "porter"
	}

	return strategy
}

// AsBundleWiring is the wiring string representation in the bundle definition.
// For example, bundle.parameters.PARAM or bundle.dependencies.DEP.outputs.OUTPUT
func (s DependencySource) AsBundleWiring() string {
	if s.Value != "" {
		return s.Value
	}

	suffix := s.WiringSuffix()
	if s.Dependency != "" {
		return fmt.Sprintf("bundle.dependencies.%s.%s", s.Dependency, suffix)
	}

	return fmt.Sprintf("bundle.%s", suffix)
}

// AsWorkflowWiring is the wiring string representation in a workflow definition.
// For example, workflow.jobs.JOB.outputs.OUTPUT
func (s DependencySource) AsWorkflowWiring(jobID string) string {
	if s.Value != "" {
		return s.Value
	}

	return fmt.Sprintf("workflow.jobs.%s.%s", jobID, s.WiringSuffix())
}

// WiringSuffix identifies the data to retrieve from the source.
// For example, parameters.PARAM or outputs.OUTPUT
func (s DependencySource) WiringSuffix() string {
	if s.Parameter != "" {
		return fmt.Sprintf("parameters.%s", s.Parameter)
	}

	if s.Credential != "" {
		return fmt.Sprintf("credentials.%s", s.Credential)
	}

	if s.Output != "" {
		return fmt.Sprintf("outputs.%s", s.Output)
	}

	return s.Value
}

type WorkflowWiring struct {
	WorkflowID string
	JobKey     string
	Parameter  string
	Credential string
	Output     string
}

var workflowWiringRegex = regexp.MustCompile(`workflow\.([^\.]+)\.jobs\.([^\.]+)\.([^\.]+)\.(.+)`)

func ParseWorkflowWiring(value string) (WorkflowWiring, error) {
	matches := workflowWiringRegex.FindStringSubmatch(value)
	if len(matches) < 5 {
		return WorkflowWiring{}, fmt.Errorf("invalid workflow wiring was passed to the porter strategy, %s", value)
	}

	// the first group is the entire match, we don't care about it
	workflowID := matches[1]
	jobKey := matches[2]
	dataType := matches[3] // e.g. parameters, credentials or outputs
	dataKey := matches[4]  // e.g. the name of the param/cred/output

	wiring := WorkflowWiring{
		WorkflowID: workflowID,
		JobKey:     jobKey,
	}

	switch dataType {
	case "parameters":
		wiring.Parameter = dataKey
	case "credentials":
		wiring.Credential = dataKey
	case "outputs":
		wiring.Output = dataKey
	default:
		return WorkflowWiring{}, fmt.Errorf("invalid workflow wiring was passed to the porter strategy, %s", value)
	}

	return wiring, nil
}

type DependencyInstallation struct {
	Labels   map[string]string     `json:"labels,omitempty" mapstructure:"labels,omitempty"`
	Criteria *InstallationCriteria `json:"criteria,omitempty" mapstructure:"criteria,omitempty"`
}

type InstallationCriteria struct {
	// MatchInterface specifies if the installation should use the same bundle or just needs to match the interface
	MatchInterface bool `json:"matchInterface,omitempty" mapstructure:"matchInterface,omitEmpty"`
	MatchNamespace bool `json:"matchNamespace,omitempty" mapstructure:"matchNamespace,omitEmpty"`
	IgnoreLabels   bool `json:"ignoreLabels,omitempty" mapstructure:"ignoreLabels,omitempty"`
}

type DependencyInterface struct {
	Reference string           `json:"reference,omitempty" mapstructure:"reference,omitempty"`
	Document  *json.RawMessage `json:"document,omitempty" mapstructure:"document,omitempty"`
}
