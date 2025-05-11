package v2

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/cnabio/cnab-go/bundle"
)

const (
	// SharingModeGroup specifies that a dependency may be shared with other bundles in the same sharing group.
	// Group is the default sharing mode.
	SharingModeGroup = "group"

	// SharingModeNone specifies that a dependency must not be shared with other bundles.
	SharingModeNone = "none"
)

// Dependencies describes the set of custom extension metadata associated with v2 of Porter's implementation of dependencies.
type Dependencies struct {
	// Requires is a list of bundles required by this bundle
	Requires map[string]Dependency `json:"requires,omitempty" mapstructure:"requires,omitempty"`

	// Provides specifies how the bundle can satisfy a dependency.
	// This declares that the bundle can provide a dependency that another bundle requires.
	Provides *DependencyProvider `json:"provides,omitempty" mapstructure:"provides,omitempty"`
}

// Dependency describes a dependency on another bundle
type Dependency struct {
	// Name of the dependency
	// This is used internally but isn't persisted to bundle.json
	Name string `json:"-" mapstructure:"-"`

	// Bundle is the location of the bundle in a registry, for example REGISTRY/NAME:TAG
	Bundle string `json:"bundle" mapstructure:"bundle"`

	// Version is a set of allowed versions defined according to the https://github.com/Masterminds/semver constraint syntax.
	Version string `json:"version,omitempty" mapstructure:"version"`

	// Interface defines how the dependency is used by this bundle.
	Interface *DependencyInterface `json:"interface,omitempty" mapstructure:"interface,omitempty"`

	// Sharing is a set of rules for sharing a dependency with other bundles.
	Sharing SharingCriteria `json:"sharing,omitempty" mapstructure:"sharing,omitempty"`

	// Parameters to pass from the bundle to the dependency.
	// The value may use templates with the bundle and installation variables
	Parameters map[string]string `json:"parameters,omitempty" mapstructure:"parameters,omitempty"`

	// Credentials to pass from the bundle to the dependency.
	// The value may use templates with the bundle and installation variables
	Credentials map[string]string `json:"credentials,omitempty" mapstructure:"credentials,omitempty"`

	// Outputs to promote from the dependency to the parent bundle.
	// The value may use templates with the bundle, installation, and outputs variables.
	// The outputs variable is a copy of bundle.dependencies.CURRENT_DEP.outputs.
	Outputs map[string]string `json:"outputs,omitempty" mapstructure:"outputs,omitempty"`
}

// DependencySource represents how to pass data between dependencies.
// For example, passing the output of a dependency to the parameter of another.
//
// Exactly one of `value`, `parameter`, `credential` or `output` must be
// specified. A parent bundle can pass a hard-coded value, the value of a
// parameter or credential, or the output from another child dependency to the
// credential of a child dependency.
type DependencySource struct {
	// Value species a hard-coded value to pass to the dependency.
	Value string `json:"value,omitempty" mapstructure:"value,omitempty"`

	// Dependency is the name of another dependency from which the Output is defined.
	Dependency string `json:"dependency,omitempty" mapstructure:"dependency,omitempty"`

	// Credential is the name of a credential defined on the parent bundle.
	Credential string `json:"credential,omitempty" mapstructure:"credential,omitempty"`

	// Parameter is the name of a parameter defined on the parent bundle.
	Parameter string `json:"parameter,omitempty" mapstructure:"parameter,omitempty"`

	// Output is the name of an output defined on `dependency`. Used to pass an
	// output from a dependency to another dependency. MUST be specified with
	// `dependency`.
	Output string `json:"output,omitempty" mapstructure:"output,omitempty"`
}

// ignore template syntax, ${...}, if found
var dependencySourceWiringRegex = regexp.MustCompile(`(\s*\$\{\s*)?bundle(\.dependencies\.([^.]+))?\.([^.]+)\.([^\s\}]+)(\s*\}\s*)?`)

// ParseDependencySource identifies the components specified in a template variable.
func ParseDependencySource(templateVariable string) (DependencySource, error) {
	matches := dependencySourceWiringRegex.FindStringSubmatch(templateVariable)

	// If it doesn't match our wiring syntax, assume that it is a hard coded value
	if len(matches) < 5 {
		return DependencySource{Value: templateVariable}, nil
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

// SharingCriteria is a set of rules for sharing a dependency with other bundles.
type SharingCriteria struct {
	// Mode defines how a dependency can be shared.
	// * false: The dependency cannot be shared, even within the same dependency graph.
	// * true: The dependency is shared with other bundles who defined the dependency with the same sharing group.
	Mode bool `json:"mode,omitempty" mapstructure:"mode,omitempty"`

	// Group defines matching criteria for determining if two dependencies are in the same sharing group.
	Group SharingGroup `json:"group,omitempty" mapstructure:"group,omitempty"`
}

// SharingGroup defines a set of characteristics for sharing a dependency with
// other bundles.
// Reserved for future use: We may add more characteristics to the sharing group (such as labels) if it seems useful.
type SharingGroup struct {
	// Name of the sharing group. The name of the group must match for two bundles to share the same dependency.
	Name string `json:"name,omitempty" mapstructure:"name,omitempty"`
}

// DependencyInterface defines how the dependency is used by this bundle.
type DependencyInterface struct {
	// ID is the identifier or name of the bundle interface. It should be matched
	// against the Dependencies.Provides.Interface.ID to determine if two interfaces
	// are equivalent.
	ID string `json:"id,omitempty" mapstructure:"id,omitempty"`

	// Reference is an OCI reference to a bundle. The bundle.json is used to define
	// the interface using the bundle's credentials, parameters and outputs.
	Reference string `json:"reference,omitempty" mapstructure:"reference,omitempty"`

	// Document is an embedded subset of a bundle.json document, defining relevant
	// portions of a bundle's interface, such as credentials, parameters and outputs.
	Document DependencyInterfaceDocument `json:"document,omitempty" mapstructure:"document,omitempty"`
}

// DependencyProvider specifies how the current bundle can be used to satisfy a dependency.
type DependencyProvider struct {
	// Interface declares the bundle interface that the current bundle provides.
	Interface InterfaceDeclaration `json:"interface,omitempty" mapstructure:"interface,omitempty"`
}

// InterfaceDeclaration declares that the current bundle supports the specified bundle interface
// Reserved for future use. Right now we only use an interface id, but could support other fields later.
type InterfaceDeclaration struct {
	// ID is the URI of the interface that this bundle provides. Usually a well-known name defined by Porter or CNAB.
	ID string `json:"id,omitempty" mapstructure:"id,omitempty"`
}

// DependencyInterfaceDocument declares an inline bundle.json that defines the bundle interface
type DependencyInterfaceDocument struct {
	// Outputs defined on the bundle interface
	Outputs map[string]bundle.Output `json:"outputs,omitempty" mapstructure:"outputs,omitempty"`
	// Parameters defined on the bundle interface
	Parameters map[string]bundle.Parameter `json:"parameters,omitempty" mapstructure:"parameters,omitempty"`
	// Credentials defined on the bundle interface
	Credentials map[string]bundle.Credential `json:"credentials,omitempty" mapstructure:"credentials,omitempty"`
}
