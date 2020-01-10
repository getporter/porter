package porter

import (
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
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

type DisplayOutput struct {
	Name       string
	Definition definition.Schema
	Value      string
	Type       string
}

// ListBundleOutputs lists the outputs for a given bundle,
// according to the provided claim and display format
func (p *Porter) ListBundleOutputs(c claim.Claim, format printer.Format) []DisplayOutput {
	// Get sorted keys for ordered printing
	keys := make([]string, 0, len(c.Outputs))
	for k := range c.Outputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	outputs := make([]DisplayOutput, 0, len(c.Outputs))
	// Iterate through all Bundle Outputs, fetch their metadata
	// via their corresponding Definitions and add to rows
	for _, name := range keys {
		do := DisplayOutput{Name: name}

		var outputType string
		valueStr := fmt.Sprintf("%v", c.Outputs[name])
		do.Value = valueStr

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
		do.Definition = *def

		if def.WriteOnly != nil && *def.WriteOnly {
			valueStr = output.Path
		}

		outputType, _, err := def.GetType()
		if err != nil {
			// Do not have the entire listing fail because of one output type error
			if p.Debug {
				fmt.Fprintf(p.Err, "unable to get output type for %s\n", name)
			}
			outputType = "unknown"
		}
		do.Type = outputType

		// Try to figure out if this was originally a file output. Long term, we should find a way to find
		// and crack open the invocation image to get the porter.yaml
		if do.Type == "string" && do.Definition.ContentEncoding == "base64" {
			do.Type = "file"
		}

		// If table output is desired, truncate the value to a reasonable length
		if format == printer.FormatTable {
			do.Value = truncateString(valueStr, 60)
		}

		outputs = append(outputs, do)
	}

	return outputs
}

func (p *Porter) PrintBundleOutputs(opts *OutputListOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	c, err := p.InstanceStorage.Read(opts.Name)
	if err != nil {
		return err
	}

	outputs := p.ListBundleOutputs(c, opts.Format)
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
