package porter

import (
	"fmt"
	"sort"

	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/printer"
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

	// If not provided, attempt to derive claim name from context
	if o.sharedOptions.Name == "" {
		err := o.sharedOptions.defaultBundleFiles(cxt)
		if err != nil {
			return errors.New("bundle instance name must be provided via [--instance|-i INSTANCE]")
		}
	}

	return nil
}

// Validate validates the provided args, using the provided context,
// setting attributes of OutputListOptions as applicable
func (o *OutputListOptions) Validate(args []string, cxt *context.Context) error {
	// Ensure only one argument exists (claim name) if args length non-zero
	err := o.sharedOptions.validateInstanceName(args)
	if err != nil {
		return err
	}

	// Attempt to derive claim name from context
	err = o.sharedOptions.defaultBundleFiles(cxt)
	if err != nil {
		return errors.Wrap(err, "bundle instance name must be provided")
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
		return errors.Wrapf(err, "unable to read output '%s' for bundle instance '%s'", opts.Output, name)
	}

	fmt.Fprintln(p.Out, output)
	return nil
}

// ListBundleOutputs lists the outputs for a given bundle,
// according to the provided options
func (p *Porter) ListBundleOutputs(opts *OutputListOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	c, err := p.InstanceStorage.Read(opts.sharedOptions.Name)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, c.Outputs)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, c.Outputs)
	case printer.FormatTable:
		return p.printOutputsTable(c)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

// ReadBundleOutput reads a bundle output from a claim
func (p *Porter) ReadBundleOutput(name, claim string) (string, error) {
	c, err := p.InstanceStorage.Read(claim)
	if err != nil {
		return "", err
	}

	if output, exists := c.Outputs[name]; exists {
		return fmt.Sprintf("%v", output), nil
	}
	return "", fmt.Errorf("unable to read output %q for bundle instance %q", name, claim)
}

func (p *Porter) printOutputsTable(c claim.Claim) error {
	var rows [][]string

	// Get sorted keys for ordered printing
	keys := make([]string, 0, len(c.Outputs))
	for k := range c.Outputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Iterate through all Bundle Outputs, fetch their metadata
	// via their corresponding Definitions and add to rows
	for _, name := range keys {
		var outputType string
		valueStr := fmt.Sprintf("%v", c.Outputs[name])

		if c.Bundle == nil {
			continue
		}

		output, exists := c.Bundle.Outputs[name]
		if !exists {
			continue
		}

		def, exists := c.Bundle.Definitions[output.Definition]
		if !exists {
			continue
		}

		if def.WriteOnly != nil && *def.WriteOnly {
			valueStr = output.Path
		}

		outputType, _, err := def.GetType()
		if err != nil {
			return errors.Wrapf(err, "unable to get output type for %s", name)
		}

		truncatedValue := truncateString(valueStr, 60)
		rows = append(rows, []string{name, outputType, truncatedValue})
	}

	// Build and configure our tablewriter for the outputs
	table := tablewriter.NewWriter(p.Out)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
	table.SetAutoFormatHeaders(false)

	// Print the outputs table
	table.SetHeader([]string{"Name", "Type", "Value (Path if sensitive)"})
	for _, row := range rows {
		table.Append(row)
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
