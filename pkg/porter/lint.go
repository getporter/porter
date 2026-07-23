package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
)

type LintOptions struct {
	printer.PrintOptions

	// File path to the porter manifest. Defaults to the bundle in the current directory.
	File string

	// InsecureRegistry allows connecting to an unsecured registry or one without verifiable certificates,
	// when resolving dependency bundles to validate their parameter/credential mappings.
	InsecureRegistry bool
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
			return fmt.Errorf("could not check if porter manifest exists in current directory: %w", err)
		}

		if manifestExists {
			o.File = config.Name
		}
	}

	// Verify the file can be accessed
	if _, err := cxt.FileSystem.Stat(o.File); err != nil {
		return fmt.Errorf("unable to access --file %s: %w", o.File, err)
	}

	return nil
}

// Lint porter.yaml for any problems and report the results.
// This calls the mixins to analyze their sections of the manifest.
func (p *Porter) Lint(ctx context.Context, opts LintOptions) (linter.Results, error) {
	manifest, err := manifest.LoadManifestFrom(ctx, p.Config, opts.File)
	if err != nil {
		return nil, err
	}

	depBundles, depResults := p.resolveDependencyBundlesForLint(ctx, manifest, opts)

	l := linter.New(p.Context, p.Mixins)
	results, err := l.Lint(ctx, manifest, p.Config, depBundles)
	if err != nil {
		return nil, err
	}

	return append(depResults, results...), nil
}

// resolveDependencyBundlesForLint pulls (from cache, or the registry) the bundle definition for
// each dependency so that the linter can validate that mapped parameters and credentials are
// actually defined on the dependency. Dependencies whose bundle cannot be resolved are reported
// as a warning and omitted from the returned map, so the linter simply skips validating them.
func (p *Porter) resolveDependencyBundlesForLint(ctx context.Context, m *manifest.Manifest, opts LintOptions) (map[string]cnab.ExtendedBundle, linter.Results) {
	if len(m.Dependencies.Requires) == 0 {
		return nil, nil
	}

	depBundles := make(map[string]cnab.ExtendedBundle, len(m.Dependencies.Requires))
	var results linter.Results
	for _, dep := range m.Dependencies.Requires {
		pullOpts := BundlePullOptions{
			Reference:        dep.Bundle.Reference,
			InsecureRegistry: opts.InsecureRegistry,
		}
		if err := pullOpts.Validate(); err == nil {
			if cachedBundle, err := p.PullBundle(ctx, pullOpts); err == nil {
				depBundles[dep.Name] = cachedBundle.Definition
				continue
			}
		}

		results = append(results, linter.Result{
			Level:   linter.LevelWarning,
			Code:    "porter-105",
			Title:   "Dependency error",
			Message: fmt.Sprintf("unable to resolve dependency %s (%s), so its parameter and credential mappings could not be validated", dep.Name, dep.Bundle.Reference),
			URL:     "https://porter.sh/reference/linter/#porter-105",
		})
	}

	return depBundles, results
}

// PrintLintResults lints the manifest and prints the results to the attached output.
func (p *Porter) PrintLintResults(ctx context.Context, opts LintOptions) error {
	results, err := p.Lint(ctx, opts)
	if err != nil {
		return err
	}

	if results.String() != "" {
		switch opts.Format {
		case printer.FormatPlaintext:
			fmt.Fprintln(p.Out, results.String())
		case printer.FormatJson:
			err := printer.PrintJson(p.Out, results)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid format: %s", opts.Format)
		}
	}

	if !results.HasError() && opts.Format == printer.FormatPlaintext {
		fmt.Fprintln(p.Out, "✨ Bundle validation was successful!")
	}

	return nil
}
