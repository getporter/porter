package porter

import (
	"fmt"

	"github.com/cnabio/cnab-go/claim"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// OutputShowOptions represent options for a bundle output show command
type OutputShowOptions struct {
	sharedOptions
	Output string
}

// OutputListOptions represent options for a bundle output list command
type OutputListOptions struct {
	sharedOptions
	printer.PrintOptions
}

// Validate validates the provided args, using the provided context,
// setting attributes of OutputShowOptions as applicable
func (o *OutputShowOptions) Validate(args []string, cxt *context.Context) error {
	switch len(args) {
	case 0:
		return errors.New("an output name must be provided")
	case 1:
		o.Output = args[0]
	default:
		return errors.Errorf("only one positional argument may be specified, the output name, but multiple were received: %s", args)
	}

	// If not provided, attempt to derive installation name from context
	if o.sharedOptions.Name == "" {
		err := o.sharedOptions.defaultBundleFiles(cxt)
		if err != nil {
			return errors.New("installation name must be provided via [--installation|-i INSTALLATION]")
		}
	}

	return nil
}

// Validate validates the provided args, using the provided context,
// setting attributes of OutputListOptions as applicable
func (o *OutputListOptions) Validate(args []string, cxt *context.Context) error {
	// Ensure only one argument exists (installation name) if args length non-zero
	err := o.sharedOptions.validateInstallationName(args)
	if err != nil {
		return err
	}

	// Attempt to derive installation name from context
	err = o.sharedOptions.defaultBundleFiles(cxt)
	if err != nil {
		return errors.Wrap(err, "installation name must be provided")
	}

	return o.ParseFormat()
}

// ShowBundleOutput shows a bundle output value, according to the provided options
func (p *Porter) ShowBundleOutput(opts *OutputShowOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}
	name := opts.sharedOptions.Name

	output, err := p.ReadBundleOutput(opts.Output, name)
	if err != nil {
		return errors.Wrapf(err, "unable to read output '%s' for installation '%s'", opts.Output, name)
	}

	fmt.Fprintln(p.Out, output)
	return nil
}

type DisplayOutput struct {
	Name  string
	Value string
	Type  string
}

type DisplayOutputs []DisplayOutput

func NewDisplayOutputs(outputs claim.Outputs, format printer.Format) DisplayOutputs {
	// Iterate through all Bundle Outputs, fetch their metadata
	// via their corresponding Definitions and add to rows
	displayOutputs := make(DisplayOutputs, outputs.Len())
	for i := 0; i < outputs.Len(); i++ {
		output, _ := outputs.GetByIndex(i)
		do := DisplayOutput{
			Name:  output.Name,
			Value: string(output.Value),
			Type:  "unknown",
		}

		schema, exists := output.GetSchema()
		if !exists {
			continue
		}

		outputType, _, err := schema.GetType()
		if err != nil {
			continue
		}
		do.Type = outputType

		// Try to figure out if this was originally a file output. Long term, we should find a way to find
		// and crack open the invocation image to get the porter.yaml
		if do.Type == "string" && schema.ContentEncoding == "base64" {
			do.Type = "file"
		}

		// If table output is desired, truncate the value to a reasonable length
		if format == printer.FormatTable {
			do.Value = truncateString(do.Value, 60)
		}

		displayOutputs[i] = do
	}

	return displayOutputs
}

// ListBundleOutputs lists the outputs for a given bundle according to the
// provided display format
func (p *Porter) ListBundleOutputs(opts *OutputListOptions) (DisplayOutputs, error) {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return nil, err
	}

	outputs, err := p.Claims.ReadLastOutputs(opts.Name)
	if err != nil {
		return nil, err
	}

	displayOutputs := NewDisplayOutputs(outputs, opts.Format)
	if err != nil {
		return nil, err
	}

	return displayOutputs, nil
}

func (p *Porter) PrintBundleOutputs(opts OutputListOptions) error {
	outputs, err := p.ListBundleOutputs(&opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, outputs)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, outputs)
	case printer.FormatTable:
		return p.printOutputsTable(outputs)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

// ReadBundleOutput reads a bundle output from an installation
func (p *Porter) ReadBundleOutput(outputName, installation string) (string, error) {
	o, err := p.Claims.ReadLastOutput(installation, outputName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", string(o.Value)), nil
}

func (p *Porter) printOutputsTable(outputs []DisplayOutput) error {
	// Build and configure our tablewriter for the outputs
	table := tablewriter.NewWriter(p.Out)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
	table.SetAutoFormatHeaders(false)

	// Print the outputs table
	table.SetHeader([]string{"Name", "Type", "Value"})
	for _, output := range outputs {
		table.Append([]string{output.Name, output.Type, output.Value})
	}
	table.Render()

	return nil
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
