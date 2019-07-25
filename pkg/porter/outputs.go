package porter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/outputs"
	"github.com/deislabs/porter/pkg/printer"
	tablewriter "github.com/olekukonko/tablewriter"
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
			return errors.New("claim name must be provided via [--claim|-c CLAIM]")
		}
	}

	return nil
}

// Validate validates the provided args, using the provided context,
// setting attributes of OutputListOptions as applicable
func (o *OutputListOptions) Validate(args []string, cxt *context.Context) error {
	// Ensure only one argument exists (claim name) if args length non-zero
	err := o.sharedOptions.validateClaimName(args)
	if err != nil {
		return err
	}

	// Attempt to derive claim name from context
	err = o.sharedOptions.defaultBundleFiles(cxt)
	if err != nil {
		return errors.Wrap(err, "claim name must be provided")
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

	output, err := outputs.ReadBundleOutput(p.Config, opts.Output, name)
	if err != nil {
		return errors.Wrapf(err, "unable to read output '%s' for claim '%s'", opts.Output, name)
	}

	fmt.Fprintln(p.Out, output.Value)
	return nil
}

// ListBundleOutputs lists the outputs for a given bundle,
// according to the provided options
func (p *Porter) ListBundleOutputs(opts *OutputListOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}
	claim := opts.sharedOptions.Name

	outputs, err := p.fetchBundleOutputs(claim)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, outputs)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, outputs)
	case printer.FormatTable:
		return p.printOutputsTable(outputs, claim)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) fetchBundleOutputs(claim string) (*outputs.Outputs, error) {
	outputsDir, err := p.Config.GetOutputsDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get outputs directory")
	}
	bundleOutputsDir := filepath.Join(outputsDir, claim)

	var outputList outputs.Outputs
	// Walk through bundleOutputsDir, if exists, and read all output filenames.
	// We truncate actual output values, intending for the full values to be
	// retrieved by another command.
	if ok, _ := p.Context.FileSystem.DirExists(bundleOutputsDir); ok {
		err := p.Context.FileSystem.Walk(bundleOutputsDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				output, err := outputs.ReadBundleOutput(p.Config, info.Name(), claim)
				if err != nil {
					return errors.Wrapf(err, "unable to read output '%s' for claim '%s'", info.Name(), claim)
				}

				outputList = append(outputList, *output)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		sort.Sort(sort.Reverse(outputList))
	}
	return &outputList, nil
}

func (p *Porter) printOutputsTable(outputs *outputs.Outputs, claim string) error {
	// Get local output directory for this claim
	outputsDir, err := p.Config.GetOutputsDir()
	if err != nil {
		return errors.Wrap(err, "unable to get outputs directory")
	}
	claimOutputsDir := filepath.Join(outputsDir, claim)

	var rows [][]string

	// Iterate through all Bundle Outputs and add to rows
	for _, o := range *outputs {
		value := o.Value
		// If output is sensitive, substitute local path
		if o.Sensitive {
			value = filepath.Join(claimOutputsDir, o.Name)
		}
		truncatedValue := truncateString(value, 60)
		rows = append(rows, []string{o.Name, o.Type, truncatedValue})
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

// TODO: refactor to truncate in the middle?  (Handy if paths are long)
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
