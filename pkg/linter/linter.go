package linter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin/query"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/Masterminds/semver/v3"
	"github.com/dustin/go-humanize"
)

// Level of severity for a lint result.
type Level int

func (l Level) String() string {
	switch l {
	case LevelError:
		return "error"
	case LevelWarning:
		return "warning"
	}
	return ""
}

// Code representing the problem identified by the linter
// Recommended to use the pattern MIXIN-NUMBER so that you don't collide with
// codes from another mixin or with Porter's codes.
// Example:
// - exec-105
// - helm-410
type Code string

const (
	// LevelError indicates a lint result is an error that will prevent the bundle from building properly.
	LevelError Level = 0

	// LevelWarning indicates a lint result is a warning about a best practice or identifies a problem that is not
	// guaranteed to break the build.
	LevelWarning Level = 2
)

// Result is a single item identified by the linter.
type Result struct {
	// Level of severity
	Level Level

	// Location of the problem in the manifest.
	Location Location

	// Code uniquely identifying the type of problem.
	Code Code

	// Title to display (80 chars).
	Title string

	// Message explaining the problem.
	Message string

	// URL that provides additional assistance with this problem.
	URL string
}

func (r Result) String() string {
	var buffer strings.Builder
	buffer.WriteString(fmt.Sprintf("%s(%s) - %s\n", r.Level, r.Code, r.Title))
	if r.Location.Mixin != "" {
		buffer.WriteString(r.Location.String() + "\n")
	}

	if r.Message != "" {
		buffer.WriteString(r.Message + "\n")
	}

	if r.URL != "" {
		buffer.WriteString(fmt.Sprintf("See %s for more information\n", r.URL))
	}

	buffer.WriteString("---\n")
	return buffer.String()
}

// Location identifies the offending mixin step within a manifest.
type Location struct {
	// Action containing the step, e.g. Install.
	Action string

	// Mixin name, e.g. exec.
	Mixin string

	// StepNumber is the position of the step, starting from 1, within the action.
	// Example
	// install:
	//  - exec: (1)
	//     ...
	//  - helm3: (2)
	//     ...
	//  - exec: (3)
	//     ...
	StepNumber int

	// StepDescription is the description of the step provided in the manifest.
	// Example
	// install:
	//  - exec:
	//      description: THIS IS THE STEP DESCRIPTION
	//      command: ./helper.sh
	StepDescription string
}

func (l Location) String() string {
	return fmt.Sprintf("%s: %s step in the %s mixin (%s)",
		l.Action, humanize.Ordinal(l.StepNumber), l.Mixin, l.StepDescription)
}

// Results is a set of items identified by the linter.
type Results []Result

func (r Results) String() string {
	var buffer strings.Builder
	// TODO: Sort, display errors first
	for _, result := range r {
		buffer.WriteString(result.String())
	}

	return buffer.String()
}

// HasError checks if any of the results is an error.
func (r Results) HasError() bool {
	for _, result := range r {
		if result.Level == LevelError {
			return true
		}
	}
	return false
}

// Linter manages executing the lint command for all affected mixins and reporting
// the results.
type Linter struct {
	*portercontext.Context
	Mixins pkgmgmt.PackageManager
}

func New(cxt *portercontext.Context, mixins pkgmgmt.PackageManager) *Linter {
	return &Linter{
		Context: cxt,
		Mixins:  mixins,
	}
}

type action struct {
	name  string
	steps manifest.Steps
}

func (l *Linter) Lint(ctx context.Context, m *manifest.Manifest, config *config.Config) (Results, error) {
	// Check for reserved porter prefix on parameter names
	reservedPrefixes := []string{"porter-", "porter_"}
	params := m.Parameters

	var results Results

	for _, param := range params {
		paramName := strings.ToLower(param.Name)
		for _, reservedPrefix := range reservedPrefixes {
			if strings.HasPrefix(paramName, reservedPrefix) {

				res := Result{
					Level: LevelError,
					Location: Location{
						Action:          "",
						Mixin:           "",
						StepNumber:      0,
						StepDescription: "",
					},
					Code:    "porter-100",
					Title:   "Reserved name error",
					Message: param.Name + " has a reserved prefix. Parameters cannot start with porter- or porter_",
					URL:     "https://porter.sh/reference/linter/#porter-100",
				}
				results = append(results, res)
			}
		}
	}

	// Check if parameters apply to the steps
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Validating that parameters applies to the actions...")
	tmplParams := m.GetTemplatedParameters()
	actions := []action{
		{"install", m.Install},
		{"upgrade", m.Upgrade},
		{"uninstall", m.Uninstall},
	}
	for actionName, steps := range m.CustomActions {
		actions = append(actions, action{actionName, steps})
	}
	for _, action := range actions {
		res, err := validateParamsAppliesToAction(m, action.steps, tmplParams, action.name, config)
		if err != nil {
			return nil, span.Error(fmt.Errorf("error validating action: %s", action.name))
		}
		results = append(results, res...)
	}

	deps := make(map[string]interface{}, len(m.Dependencies.Requires))
	for _, dep := range m.Dependencies.Requires {
		if _, exists := deps[dep.Name]; exists {
			res := Result{
				Level: LevelError,
				Location: Location{
					Action:          "",
					Mixin:           "",
					StepNumber:      0,
					StepDescription: "",
				},
				Code:    "porter-102",
				Title:   "Dependency error",
				Message: fmt.Sprintf("The dependency %s is defined multiple times", dep.Name),
				URL:     "https://porter.sh/reference/linter/#porter-102",
			}
			results = append(results, res)
		} else {
			deps[dep.Name] = nil
		}
	}

	span.Debug("Running linters for each mixin used in the manifest...")
	q := query.New(l.Context, l.Mixins)
	responses, err := q.Execute(ctx, "lint", query.NewManifestGenerator(m))
	if err != nil {
		return nil, span.Error(err)
	}

	for _, response := range responses {
		if response.Error != nil {
			// Ignore mixins that do not support the lint command
			if strings.Contains(response.Error.Error(), "unknown command") {
				continue
			}
			// put a helpful error when the mixin is not installed
			if strings.Contains(response.Error.Error(), "not installed") {
				return nil, span.Error(fmt.Errorf("mixin %[1]s is not currently installed. To find view more details you can run: porter mixin search %[1]s. To install you can run porter mixin install %[1]s", response.Name))
			}
			return nil, span.Error(fmt.Errorf("lint command failed for mixin %s: %s", response.Name, response.Stdout))
		}

		var r Results
		err = json.Unmarshal([]byte(response.Stdout), &r)
		if err != nil {
			return nil, span.Error(fmt.Errorf("unable to parse lint response from mixin %s: %w", response.Name, err))
		}

		results = append(results, r...)
	}

	span.Debug("Getting versions for each mixin used in the manifest...")
	err = l.validateVersionNumberConstraints(ctx, m)
	if err != nil {
		return nil, span.Error(err)
	}

	return results, nil
}

func (l *Linter) validateVersionNumberConstraints(ctx context.Context, m *manifest.Manifest) error {
	for _, mixin := range m.Mixins {
		if mixin.Version != nil {
			installedMeta, err := l.Mixins.GetMetadata(ctx, mixin.Name)
			if err != nil {
				return fmt.Errorf("unable to get metadata from mixin %s: %w", mixin.Name, err)
			}
			installedVersion := installedMeta.GetVersionInfo().Version

			err = validateSemverConstraint(mixin.Name, installedVersion, mixin.Version)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validateSemverConstraint(name string, installedVersion string, versionConstraint *semver.Constraints) error {
	v, err := semver.NewVersion(installedVersion)
	if err != nil {
		return fmt.Errorf("invalid version number from mixin %s: %s. %w", name, installedVersion, err)
	}

	if !versionConstraint.Check(v) {
		return fmt.Errorf("mixin %s is installed at version %s but your bundle requires version %s", name, installedVersion, versionConstraint)
	}
	return nil
}

func validateParamsAppliesToAction(m *manifest.Manifest, steps manifest.Steps, tmplParams manifest.ParameterDefinitions, actionName string, config *config.Config) (Results, error) {
	var results Results
	for stepNumber, step := range steps {
		data, err := yaml.Marshal(step.Data)
		if err != nil {
			return nil, fmt.Errorf("error during marshalling: %w", err)
		}

		tmplResult, err := m.ScanManifestTemplating(data, config)
		if err != nil {
			return nil, fmt.Errorf("error parsing templating: %w", err)
		}

		for _, variable := range tmplResult.Variables {
			paramName, ok := m.GetTemplateParameterName(variable)
			if !ok {
				continue
			}

			for _, tmplParam := range tmplParams {
				if tmplParam.Name != paramName {
					continue
				}
				if !tmplParam.AppliesTo(actionName) {
					description, err := step.GetDescription()
					if err != nil {
						return nil, fmt.Errorf("error getting step description: %w", err)
					}
					res := Result{
						Level: LevelError,
						Location: Location{
							Action:          actionName,
							Mixin:           step.GetMixinName(),
							StepNumber:      stepNumber + 1,
							StepDescription: description,
						},
						Code:    "porter-101",
						Title:   "Parameter does not apply to action",
						Message: fmt.Sprintf("Parameter %s does not apply to %s action", paramName, actionName),
						URL:     "https://porter.sh/docs/references/linter/#porter-101",
					}
					results = append(results, res)
				}
			}
		}
	}

	return results, nil
}
