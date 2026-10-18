package porter

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"get.porter.sh/porter/pkg/cnab"
)

// setEntry is a credential or parameter listed in a generated set file.
type setEntry struct {
	Name string

	// Description is split into lines so that each one can be a yaml comment.
	Description []string
}

func newSetEntry(name string, description string) setEntry {
	entry := setEntry{Name: name}
	if description != "" {
		entry.Description = strings.Split(description, "\n")
	}
	return entry
}

// setTemplateData is the data passed to the *-from-bundle templates.
type setTemplateData struct {
	Reference string
	Name      string
	Items     []setEntry
}

// createSetFromReference writes a credential or parameter set file, populated
// from the bundle at the specified reference, to stdout or fileName.
// render builds the file contents in the requested output type.
func (p *Porter) createSetFromReference(ctx context.Context, kind string, pullOpts BundlePullOptions, fileName string, outputType string, render func(bun cnab.ExtendedBundle, outputType string) ([]byte, error)) error {
	if outputType == "" {
		outputType = "yaml"
	}
	switch outputType {
	case "json", "yaml", "yml":
	default:
		return newUnsupportedFormatError(outputType)
	}

	bundleOpts := BundleReferenceOptions{BundlePullOptions: pullOpts}
	bundleRef, err := bundleOpts.GetBundleReference(ctx, p)
	if err != nil {
		return err
	}

	contents, err := render(bundleRef.Definition, outputType)
	if err != nil {
		return err
	}

	if fileName == "" {
		fmt.Fprintln(p.Out, string(contents))
		return nil
	}

	fmt.Fprintf(p.Err, "creating porter %s set in the current directory\n", kind)
	return p.CopyTemplate(func() ([]byte, error) { return contents, nil }, fileName)
}

// renderSetTemplate executes a *-from-bundle template.
func renderSetTemplate(tmplSrc []byte, data setTemplateData) ([]byte, error) {
	tmpl, err := template.New("set").Parse(string(tmplSrc))
	if err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("could not render template: %w", err)
	}
	return buf.Bytes(), nil
}
