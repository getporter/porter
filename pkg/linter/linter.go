package linter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin/query"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
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
	buffer.WriteString(r.Location.String() + "\n")

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

func (l *Linter) Lint(ctx context.Context, m *manifest.Manifest) (Results, error) {
	// TODO: perform any porter level linting
	// e.g. metadata, credentials, properties, outputs, dependencies, etc

	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Running linters for each mixin used in the manifest...")
	q := query.New(l.Context, l.Mixins)
	responses, err := q.Execute(ctx, "lint", query.NewManifestGenerator(m))
	if err != nil {
		return nil, span.Error(err)
	}

	var results Results
	for mixin, response := range responses {
		var r Results
		err = json.Unmarshal([]byte(response), &r)
		if err != nil {
			return nil, span.Error(fmt.Errorf("unable to parse lint response from mixin %q: %w", mixin, err))
		}

		results = append(results, r...)
	}

	return results, nil
}
