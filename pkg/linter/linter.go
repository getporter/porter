package linter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin/query"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/pkg/errors"
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

	// Key locates the problem in the manifest.
	Key string

	// Location is the location of the problem in the manifest.
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

type Location struct {
	Line   int
	Column int
}

func (l Location) String() string {
	return fmt.Sprintf("Location in manifest: Line: %d, Column: %d",
		l.Line, l.Column)
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
	*context.Context
	Mixins pkgmgmt.PackageManager
}

func New(cxt *context.Context, mixins pkgmgmt.PackageManager) *Linter {
	return &Linter{
		Context: cxt,
		Mixins:  mixins,
	}
}

func (l *Linter) Lint(m *manifest.Manifest) (Results, error) {

	// TODO: perform any porter level linting
	// e.g. metadata, credentials, properties, outputs, dependencies, etc

	if l.Debug {
		fmt.Fprintln(l.Err, "Running linters for each mixin used in the manifest...")
	}

	q := query.New(l.Context, l.Mixins)
	responses, err := q.Execute("lint", query.NewManifestGenerator(m))
	if err != nil {
		return nil, err
	}

	var results Results
	// Read manifest data.  This will be used to determine locations of results
	manifestData, err := manifest.ReadManifestData(l.Context, m.ManifestPath)
	if err != nil {
		return results, errors.New("unable to read manifest data")
	}

	for mixin, response := range responses {
		var r Results
		err = json.Unmarshal([]byte(response), &r)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to parse lint response from mixin %q", mixin)
		}

		// Derive location of result
		for i, result := range r {
			location, err := getLocation(manifestData, result.Key)
			if err != nil {
				return results, errors.Wrap(err, "unable to resolve location of result in manifest")
			}
			result.Location = location
			r[i] = result
		}

		results = append(results, r...)
	}

	return results, nil
}

func getLocation(contents []byte, key string) (Location, error) {
	r := bytes.NewReader(contents)
	// Splits on newlines by default.
	scanner := bufio.NewScanner(r)

	line := 1
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, key) {
			col := strings.Index(text, key) + 1
			return Location{Line: line, Column: col}, nil
		}

		line++
	}

	if err := scanner.Err(); err != nil {
		return Location{}, err
	}
	// TODO: should we actively error out here?
	// Currently, this supports unit tests with no real manifest data, etc.
	// return Location{}, fmt.Errorf("unable to determine line and column coordinates for string %q", key)
	return Location{}, nil
}
