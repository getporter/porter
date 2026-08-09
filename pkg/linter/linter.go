package linter

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
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
	fmt.Fprintf(&buffer, "%s(%s) - %s\n", r.Level, r.Code, r.Title)
	if r.Location.Mixin != "" {
		buffer.WriteString(r.Location.String() + "\n")
	}

	if r.Message != "" {
		buffer.WriteString(r.Message + "\n")
	}

	if r.URL != "" {
		fmt.Fprintf(&buffer, "See %s for more information\n", r.URL)
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

func (l *Linter) Lint(ctx context.Context, m *manifest.Manifest, config *config.Config, depBundles map[string]cnab.ExtendedBundle) (Results, error) {
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
					URL:     "https://porter.sh/docs/references/linter/#porter-100",
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

		if config.IsFeatureEnabled(experimental.FlagDependenciesV2) {
			depOutputResults, err := validateDependencyOutputsNotInActionSteps(m, action.steps, action.name)
			if err != nil {
				return nil, span.Error(fmt.Errorf("error validating action: %s: %w", action.name, err))
			}
			results = append(results, depOutputResults...)
		}
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
				URL:     "https://porter.sh/docs/references/linter/#porter-102",
			}
			results = append(results, res)
		} else {
			deps[dep.Name] = nil
		}

		if depBundle, hasDepBundle := depBundles[dep.Name]; hasDepBundle {
			for paramName := range dep.Parameters {
				if _, ok := depBundle.Parameters[paramName]; !ok {
					results = append(results, Result{
						Level:   LevelError,
						Code:    "porter-103",
						Title:   "Dependency error",
						Message: fmt.Sprintf("dependencies.%s.parameters.%s is not defined as a parameter on the dependency bundle", dep.Name, paramName),
						URL:     "https://porter.sh/docs/references/linter/#porter-103",
					})
				}
			}

			for credName := range dep.Credentials {
				if _, ok := depBundle.Credentials[credName]; !ok {
					results = append(results, Result{
						Level:   LevelError,
						Code:    "porter-104",
						Title:   "Dependency error",
						Message: fmt.Sprintf("dependencies.%s.credentials.%s is not defined as a credential on the dependency bundle", dep.Name, credName),
						URL:     "https://porter.sh/docs/references/linter/#porter-104",
					})
				}
			}

			if config.IsFeatureEnabled(experimental.FlagDependenciesV2) {
				var unmappedParams []string
				for paramName, paramDef := range depBundle.Parameters {
					if _, ok := dep.Parameters[paramName]; !ok && paramDef.Required {
						unmappedParams = append(unmappedParams, paramName)
					}
				}
				sort.Strings(unmappedParams)
				for _, paramName := range unmappedParams {
					results = append(results, Result{
						Level:   LevelWarning,
						Code:    "porter-110",
						Title:   "Dependency warning",
						Message: fmt.Sprintf("dependencies.%s.parameters.%s is required by the dependency bundle but is not mapped", dep.Name, paramName),
						URL:     "https://porter.sh/docs/references/linter/#porter-110",
					})
				}

				var unmappedCreds []string
				for credName, credDef := range depBundle.Credentials {
					if _, ok := dep.Credentials[credName]; !ok && credDef.Required {
						unmappedCreds = append(unmappedCreds, credName)
					}
				}
				sort.Strings(unmappedCreds)
				for _, credName := range unmappedCreds {
					results = append(results, Result{
						Level:   LevelWarning,
						Code:    "porter-111",
						Title:   "Dependency warning",
						Message: fmt.Sprintf("dependencies.%s.credentials.%s is required by the dependency bundle but is not mapped", dep.Name, credName),
						URL:     "https://porter.sh/docs/references/linter/#porter-111",
					})
				}
			}
		}

		if config.IsFeatureEnabled(experimental.FlagDependenciesV2) {
			depVarResults, err := validateDependencyTemplateVariables(m, dep, depBundles)
			if err != nil {
				return nil, span.Error(err)
			}
			results = append(results, depVarResults...)
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

// validateDependencyTemplateVariables checks that the template variables
// used in a dependency's parameters, credentials, and outputs mappings are
// supported (bundle.*, installation.*, and the outputs.* shorthand only
// within the dependency's own outputs mapping), and that the referenced
// parameter, credential, and dependency output names actually exist.
func validateDependencyTemplateVariables(m *manifest.Manifest, dep *manifest.Dependency, depBundles map[string]cnab.ExtendedBundle) (Results, error) {
	var results Results

	fields := []struct {
		name         string
		values       map[string]string
		allowOutputs bool
	}{
		{"parameters", dep.Parameters, false},
		{"credentials", dep.Credentials, false},
		{"outputs", dep.Outputs, true},
	}

	for _, field := range fields {
		keys := make([]string, 0, len(field.values))
		for key := range field.values {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			vars, err := m.GetTemplateVariables(field.values[key])
			if err != nil {
				return nil, fmt.Errorf("error parsing the templating used for dependencies.%s.%s.%s: %w", dep.Name, field.name, key, err)
			}

			varNames := make([]string, 0, len(vars))
			for v := range vars {
				varNames = append(varNames, v)
			}
			sort.Strings(varNames)

			for _, v := range varNames {
				switch {
				case strings.HasPrefix(v, "bundle."):
					if paramName, ok := m.GetTemplateParameterName(v); ok {
						if _, defined := m.Parameters[paramName]; !defined {
							results = append(results, Result{
								Level:   LevelError,
								Code:    "porter-107",
								Title:   "Dependency error",
								Message: fmt.Sprintf("dependencies.%s.%s.%s references %s, which is not defined as a parameter on the bundle", dep.Name, field.name, key, v),
								URL:     "https://porter.sh/docs/references/linter/#porter-107",
							})
						}
						continue
					}

					if credName, ok := m.GetTemplateCredentialName(v); ok {
						if _, defined := m.Credentials[credName]; !defined {
							results = append(results, Result{
								Level:   LevelError,
								Code:    "porter-108",
								Title:   "Dependency error",
								Message: fmt.Sprintf("dependencies.%s.%s.%s references %s, which is not defined as a credential on the bundle", dep.Name, field.name, key, v),
								URL:     "https://porter.sh/docs/references/linter/#porter-108",
							})
						}
						continue
					}

					if refDepName, outputName, ok := m.GetTemplateDependencyOutputName(v); ok {
						if res, flagged := checkDependencyOutput(m, depBundles, dep.Name, field.name, key, v, refDepName, outputName); flagged {
							results = append(results, res)
						}
						continue
					}

					// Other bundle.* variables, e.g. bundle.outputs.X or bundle.name, are allowed without further checking.

				case v == "installation.name" || v == "installation.namespace" || v == "installation.id":
					// Allowed.

				case strings.HasPrefix(v, "outputs."):
					if !field.allowOutputs {
						results = append(results, Result{
							Level:   LevelError,
							Code:    "porter-106",
							Title:   "Dependency error",
							Message: fmt.Sprintf("dependencies.%s.%s.%s references %s, but the outputs variable can only be used within a dependency's output mappings", dep.Name, field.name, key, v),
							URL:     "https://porter.sh/docs/references/linter/#porter-106",
						})
						continue
					}

					outputName := strings.TrimPrefix(v, "outputs.")
					if res, flagged := checkDependencyOutput(m, depBundles, dep.Name, field.name, key, v, dep.Name, outputName); flagged {
						results = append(results, res)
					}

				default:
					results = append(results, Result{
						Level:   LevelError,
						Code:    "porter-106",
						Title:   "Dependency error",
						Message: fmt.Sprintf("dependencies.%s.%s.%s references %s, which is not a supported template variable in the dependencies section (supported: bundle.*, installation.*, and outputs.* inside a dependency's output mappings)", dep.Name, field.name, key, v),
						URL:     "https://porter.sh/docs/references/linter/#porter-106",
					})
				}
			}
		}
	}

	return results, nil
}

// checkDependencyOutput validates that outputName is defined for refDepName,
// either as a key in refDepName's own outputs mapping (declared directly in
// porter.yaml) or as an output on refDepName's resolved bundle. sourceDep,
// fieldName, key, and v identify where the reference was made, for the
// resulting porter-109 message.
func checkDependencyOutput(m *manifest.Manifest, depBundles map[string]cnab.ExtendedBundle, sourceDep, fieldName, key, v, refDepName, outputName string) (Result, bool) {
	refDep := findDependency(m, refDepName)
	if refDep == nil {
		return Result{
			Level:   LevelError,
			Code:    "porter-109",
			Title:   "Dependency error",
			Message: fmt.Sprintf("dependencies.%s.%s.%s references %s, but %s is not declared under dependencies.requires", sourceDep, fieldName, key, v, refDepName),
			URL:     "https://porter.sh/docs/references/linter/#porter-109",
		}, true
	}

	if _, ok := refDep.Outputs[outputName]; ok {
		return Result{}, false
	}

	if depBundle, ok := depBundles[refDepName]; ok {
		if _, ok := depBundle.Outputs[outputName]; ok {
			return Result{}, false
		}

		return Result{
			Level:   LevelError,
			Code:    "porter-109",
			Title:   "Dependency error",
			Message: fmt.Sprintf("dependencies.%s.%s.%s references %s, which is not defined as an output on the %s dependency", sourceDep, fieldName, key, v, refDepName),
			URL:     "https://porter.sh/docs/references/linter/#porter-109",
		}, true
	}

	// refDepName's bundle wasn't resolved and it doesn't declare outputName
	// itself, so this can't be validated here; porter-105 already warns that
	// the dependency couldn't be resolved.
	return Result{}, false
}

// findDependency returns the dependency definition with the given name, or
// nil if it isn't declared under dependencies.requires.
func findDependency(m *manifest.Manifest, name string) *manifest.Dependency {
	for _, dep := range m.Dependencies.Requires {
		if dep.Name == name {
			return dep
		}
	}
	return nil
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

// validateDependencyOutputsNotInActionSteps checks that action steps don't
// reference bundle.dependencies.DEP.outputs.NAME directly; that form is only
// allowed inside dependencies.requires[] wiring (see
// validateDependencyTemplateVariables). Action steps must go through
// bundle.outputs.NAME after the value has been promoted.
func validateDependencyOutputsNotInActionSteps(m *manifest.Manifest, steps manifest.Steps, actionName string) (Results, error) {
	var results Results
	for stepNumber, step := range steps {
		data, err := yaml.Marshal(step.Data)
		if err != nil {
			return nil, fmt.Errorf("error during marshalling: %w", err)
		}

		// Use the raw variable parser here, not ScanManifestTemplating: that
		// helper injects every bundle.dependencies.*.outputs.* variable
		// implied by outputs.NAME shorthand usage anywhere in
		// dependencies.requires[].outputs into every scan's result, which
		// would produce false positives for steps that don't actually
		// reference a dependency output.
		vars, err := m.GetTemplateVariables(string(data))
		if err != nil {
			return nil, fmt.Errorf("error parsing templating: %w", err)
		}

		varNames := make([]string, 0, len(vars))
		for v := range vars {
			varNames = append(varNames, v)
		}
		sort.Strings(varNames)

		for _, v := range varNames {
			depName, outputName, ok := m.GetTemplateDependencyOutputName(v)
			if !ok {
				continue
			}

			description, err := step.GetDescription()
			if err != nil {
				return nil, fmt.Errorf("error getting step description: %w", err)
			}
			results = append(results, Result{
				Level: LevelError,
				Location: Location{
					Action:          actionName,
					Mixin:           step.GetMixinName(),
					StepNumber:      stepNumber + 1,
					StepDescription: description,
				},
				Code:    "porter-112",
				Title:   "Dependency output referenced directly in an action step",
				Message: fmt.Sprintf("references %s directly; promote dependencies.%s.outputs.%s via dependencies.requires[].outputs and reference bundle.outputs.%s instead", v, depName, outputName, outputName),
				URL:     "https://porter.sh/docs/references/linter/#porter-112",
			})
		}
	}

	return results, nil
}
