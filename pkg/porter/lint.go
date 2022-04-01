package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"github.com/pkg/errors"
)

type LintOptions struct {
	contextOptions
	printer.PrintOptions

	// File path to the porter manifest. Defaults to the bundle in the current directory.
	File string
}

var (
	LintAllowFormats   = printer.Formats{printer.FormatPlaintext, printer.FormatJson}
	LintDefaultFormats = printer.FormatPlaintext
)

func (o *LintOptions) Validate(cxt *portercontext.Context) error {
	err := o.PrintOptions.Validate(LintDefaultFormats, LintAllowFormats)
	if err != nil {
		return err
	}

	return o.validateFile(cxt)
}

func (o *LintOptions) validateFile(cxt *portercontext.Context) error {
	if o.File == "" {
		manifestExists, err := cxt.FileSystem.Exists(config.Name)
		if err != nil {
			return errors.Wrap(err, "could not check if porter manifest exists in current directory")
		}

		if manifestExists {
			o.File = config.Name
		}
	}

	// Verify the file can be accessed
	if _, err := cxt.FileSystem.Stat(o.File); err != nil {
		return errors.Wrapf(err, "unable to access --file %s", o.File)
	}

	return nil
}

// Lint porter.yaml for any problems and report the results.
// This calls the mixins to analyze their sections of the manifest.
func (p *Porter) Lint(ctx context.Context, opts LintOptions) (linter.Results, error) {
	opts.Apply(p.Context)

	manifest, err := manifest.LoadManifestFrom(ctx, p.Config, opts.File)
	if err != nil {
		return nil, err
	}

	l := linter.New(p.Context, p.Mixins)
	return l.Lint(manifest)
}

// PrintLintResults lints the manifest and prints the results to the attached output.
func (p *Porter) PrintLintResults(ctx context.Context, opts LintOptions) error {
	results, err := p.Lint(ctx, opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatPlaintext:
		fmt.Fprintln(p.Out, results.String())
		return nil
	case printer.FormatJson:
		return printer.PrintJson(p.Out, results)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
