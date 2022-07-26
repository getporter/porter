package porter

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
)

// OutputShowOptions represent options for a bundle output show command
type OutputShowOptions struct {
	installationOptions
	Output string
}

// OutputListOptions represent options for a bundle output list command
type OutputListOptions struct {
	installationOptions
	printer.PrintOptions
}

// Validate validates the provided args, using the provided context,
// setting attributes of OutputShowOptions as applicable
func (o *OutputShowOptions) Validate(args []string, cxt *portercontext.Context) error {
	switch len(args) {
	case 0:
		return errors.New("an output name must be provided")
	case 1:
		o.Output = args[0]
	default:
		return fmt.Errorf("only one positional argument may be specified, the output name, but multiple were received: %s", args)
	}

	// If not provided, attempt to derive installation name from context
	if o.installationOptions.Name == "" {
		err := o.installationOptions.defaultBundleFiles(cxt)
		if err != nil {
			return errors.New("installation name must be provided via [--installation|-i INSTALLATION]")
		}
	}

	return nil
}

// Validate validates the provided args, using the provided context,
// setting attributes of OutputListOptions as applicable
func (o *OutputListOptions) Validate(args []string, cxt *portercontext.Context) error {
	// Ensure only one argument exists (installation name) if args length non-zero
	err := o.installationOptions.validateInstallationName(args)
	if err != nil {
		return err
	}

	// Attempt to derive installation name from context
	err = o.installationOptions.defaultBundleFiles(cxt)
	if err != nil {
		return fmt.Errorf("installation name must be provided: %w", err)
	}

	return o.ParseFormat()
}

// ShowBundleOutput shows a bundle output value, according to the provided options
func (p *Porter) ShowBundleOutput(ctx context.Context, opts *OutputShowOptions) error {
	err := p.applyDefaultOptions(ctx, &opts.installationOptions)
	if err != nil {
		return err
	}

	output, err := p.ReadBundleOutput(ctx, opts.Output, opts.Name, opts.Namespace)
	if err != nil {
		return fmt.Errorf("unable to read output '%s' for installation '%s/%s': %w", opts.Output, opts.Namespace, opts.Name, err)
	}

	fmt.Fprintln(p.Out, output)
	return nil
}

func NewDisplayValuesFromOutputs(bun cnab.ExtendedBundle, outputs storage.Outputs) DisplayValues {
	// Iterate through all Bundle Outputs, fetch their metadata
	// via their corresponding Definitions and add to rows
	displayOutputs := make(DisplayValues, 0, outputs.Len())
	for i := 0; i < outputs.Len(); i++ {
		output, _ := outputs.GetByIndex(i)

		if bun.IsInternalOutput(output.Name) {
			continue
		}

		do := &DisplayValue{Name: output.Name}
		do.SetValue(output.Value)
		schema, ok := output.GetSchema(bun)
		if ok {
			do.Type = bun.GetParameterType(&schema)
			if schema.WriteOnly != nil && *schema.WriteOnly {
				do.Sensitive = true
			}
		} else {
			// Skip outputs not defined in the bundle, e.g. io.cnab.outputs.invocationImageLogs
			continue
		}

		displayOutputs = append(displayOutputs, *do)
	}

	sort.Sort(displayOutputs)
	return displayOutputs
}

// ListBundleOutputs lists the outputs for a given bundle according to the
// provided display format
func (p *Porter) ListBundleOutputs(ctx context.Context, opts *OutputListOptions) (DisplayValues, error) {
	err := p.applyDefaultOptions(ctx, &opts.installationOptions)
	if err != nil {
		return nil, err
	}

	outputs, err := p.Installations.GetLastOutputs(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return nil, err
	}

	resolved, err := p.Sanitizer.RestoreOutputs(ctx, outputs)
	if err != nil {
		return nil, err
	}

	c, err := p.Installations.GetLastRun(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return nil, err
	}

	bun := cnab.NewBundle(c.Bundle)

	displayOutputs := NewDisplayValuesFromOutputs(bun, resolved)
	if err != nil {
		return nil, err
	}

	return displayOutputs, nil
}

func (p *Porter) PrintBundleOutputs(ctx context.Context, opts OutputListOptions) error {
	outputs, err := p.ListBundleOutputs(ctx, &opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, outputs)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, outputs)
	case printer.FormatPlaintext:
		return p.printDisplayValuesTable(outputs)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

// ReadBundleOutput reads a bundle output from an installation
func (p *Porter) ReadBundleOutput(ctx context.Context, outputName, installation, namespace string) (string, error) {
	o, err := p.Installations.GetLastOutput(ctx, namespace, installation, outputName)
	if err != nil {
		return "", err
	}

	o, err = p.Sanitizer.RestoreOutput(ctx, o)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", string(o.Value)), nil
}

func truncateString(str string, num int) string {
	truncated := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		truncated = str[0:num] + "..."
	}
	return truncated
}
