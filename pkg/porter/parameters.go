package porter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/editor"
	"get.porter.sh/porter/pkg/generator"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/printer"
	"gopkg.in/yaml.v2"

	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/cnabio/cnab-go/valuesource"
	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// ParameterShowOptions represent options for Porter's parameter show command
type ParameterShowOptions struct {
	printer.PrintOptions
	Name string
}

// ParameterEditOptions represent iptions for Porter's parameter edit command
type ParameterEditOptions struct {
	Name string
}

// ListParameters lists saved parameter sets.
func (p *Porter) ListParameters(opts ListOptions) error {
	params, err := p.Parameters.ReadAll()
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, params)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, params)
	case printer.FormatTable:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		printParamRow :=
			func(v interface{}) []interface{} {
				cr, ok := v.(parameters.ParameterSet)
				if !ok {
					return nil
				}
				return []interface{}{cr.Name, tp.Format(cr.Modified)}
			}
		return printer.PrintTable(p.Out, params, printParamRow,
			"NAME", "MODIFIED")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

// ParameterOptions represent generic/base options for a Porter parameters command
type ParameterOptions struct {
	BundleActionOptions
	Silent bool
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (g *ParameterOptions) Validate(args []string, cxt *context.Context) error {
	err := g.validateParamName(args)
	if err != nil {
		return err
	}

	return g.bundleFileOptions.Validate(cxt)
}

func (g *ParameterOptions) validateParamName(args []string) error {
	if len(args) == 1 {
		g.Name = args[0]
	} else if len(args) > 1 {
		return errors.Errorf("only one positional argument may be specified, the parameter set name, but multiple were received: %s", args)
	}
	return nil
}

// GenerateParameters builds a new parameter set based on the given options. This can be either
// a silent build, based on the opts.Silent flag, or interactive using a survey. Returns an
// error if unable to generate parameters
func (p *Porter) GenerateParameters(opts ParameterOptions) error {
	err := p.prepullBundleByTag(&opts.BundleActionOptions)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before invoking parameters generate")
	}

	err = p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}
	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}
	bundle, err := p.CNAB.LoadBundle(opts.CNABFile)

	if err != nil {
		return err
	}
	name := opts.Name
	if name == "" {
		name = bundle.Name
	}
	genOpts := generator.GenerateParametersOptions{
		GenerateOptions: generator.GenerateOptions{
			Name:   name,
			Silent: opts.Silent,
		},
		Bundle: bundle,
	}
	fmt.Fprintf(p.Out, "Generating new parameter set %s from bundle %s\n", genOpts.Name, bundle.Name)
	fmt.Fprintf(p.Out, "==> %d parameters declared for bundle %s\n", len(bundle.Parameters), bundle.Name)

	pset, err := genOpts.GenerateParameters()
	if err != nil {
		return errors.Wrap(err, "unable to generate parameter set")
	}

	pset.Created = time.Now()
	pset.Modified = pset.Created

	err = p.Parameters.Save(pset)
	return errors.Wrapf(err, "unable to save parameter set")
}

// Validate validates the args provided to Porter's parameter show command
func (o *ParameterShowOptions) Validate(args []string) error {
	if err := validateParameterName(args); err != nil {
		return err
	}
	o.Name = args[0]
	return o.ParseFormat()
}

// Validate validates the args provided to Porter's parameter edit command
func (o *ParameterEditOptions) Validate(args []string) error {
	if err := validateParameterName(args); err != nil {
		return err
	}
	o.Name = args[0]
	return nil
}

// EditParameter edits the parameters of the provided name.
func (p *Porter) EditParameter(opts ParameterEditOptions) error {
	paramSet, err := p.Parameters.Read(opts.Name)
	if err != nil {
		return err
	}

	contents, err := yaml.Marshal(paramSet)
	if err != nil {
		return errors.Wrap(err, "unable to load parameter set")
	}

	editor := editor.New(p.Context, fmt.Sprintf("porter-%s.yaml", paramSet.Name), contents)
	output, err := editor.Run()
	if err != nil {
		return errors.Wrap(err, "unable to open editor to edit parameter set")
	}

	err = yaml.Unmarshal(output, &paramSet)
	if err != nil {
		return errors.Wrap(err, "unable to process parameter set")
	}

	err = p.Parameters.Validate(paramSet)
	if err != nil {
		return errors.Wrap(err, "parameter set is invalid")
	}

	paramSet.Modified = time.Now()
	err = p.Parameters.Save(paramSet)
	if err != nil {
		return errors.Wrap(err, "unable to save parameter set")
	}

	return nil
}

// ShowParameter shows the parameter set corresponding to the provided name, using
// the provided printer.PrintOptions for display.
func (p *Porter) ShowParameter(opts ParameterShowOptions) error {
	paramSet, err := p.Parameters.Read(opts.Name)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, paramSet)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, paramSet)
	case printer.FormatTable:
		// Set up human friendly time formatter
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		// Here we use an instance of olekukonko/tablewriter as our table,
		// rather than using the printer pkg variant, as we wish to decorate
		// the table a bit differently from the default
		var rows [][]string

		// Iterate through all ParameterStrategies and add to rows
		for _, pset := range paramSet.Parameters {
			rows = append(rows, []string{pset.Name, pset.Source.Value, pset.Source.Key})
		}

		// Build and configure our tablewriter
		table := tablewriter.NewWriter(p.Out)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
		table.SetAutoFormatHeaders(false)

		// First, print the ParameterSet metadata
		fmt.Fprintf(p.Out, "Name: %s\n", paramSet.Name)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(paramSet.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n\n", tp.Format(paramSet.Modified))

		// Now print the table
		table.SetHeader([]string{"Name", "Local Source", "Source Type"})
		for _, row := range rows {
			table.Append(row)
		}
		table.Render()
		return nil
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

// ParameterDeleteOptions represent options for Porter's parameter delete command
type ParameterDeleteOptions struct {
	Name string
}

// DeleteParameter deletes the parameter set corresponding to the provided
// names.
func (p *Porter) DeleteParameter(opts ParameterDeleteOptions) error {
	err := p.Parameters.Delete(opts.Name)
	if err != nil && strings.Contains(err.Error(), crud.ErrRecordDoesNotExist.Error()) {
		if p.Debug {
			fmt.Fprintln(p.Err, "parameter set does not exist")
		}
		return nil
	}
	return errors.Wrapf(err, "unable to delete parameter set")
}

// Validate the args provided to the delete parameter command
func (o *ParameterDeleteOptions) Validate(args []string) error {
	if err := validateParameterName(args); err != nil {
		return err
	}
	o.Name = args[0]
	return nil
}

func validateParameterName(args []string) error {
	switch len(args) {
	case 0:
		return errors.Errorf("no parameter set name was specified")
	case 1:
		return nil
	default:
		return errors.Errorf("only one positional argument may be specified, the parameter set name, but multiple were received: %s", args)
	}
}

// loadParameterSets loads parameter values per their parameter set strategies
func (p *Porter) loadParameterSets(params []string) (valuesource.Set, error) {
	resolvedParameters := valuesource.Set{}
	for _, name := range params {
		var pset parameters.ParameterSet
		var err error
		// If name looks pathy, attempt to load from a file
		// Else, read from Porter's parameters store
		if strings.Contains(name, string(filepath.Separator)) {
			pset, err = p.loadParameterFromFile(name)
		} else {
			pset, err = p.Parameters.Read(name)
		}
		if err != nil {
			return nil, err
		}

		// A parameter may correspond to a Porter-specific parameter type of 'file'
		// If so, add value (filepath) directly to map and remove from pset
		for _, paramDef := range p.Manifest.Parameters {
			if paramDef.Type == "file" {
				for i, param := range pset.Parameters {
					if param.Name == paramDef.Name {
						// Pass through value (filepath) directly to resolvedParameters
						resolvedParameters[param.Name] = param.Source.Value
						// Eliminate this param from pset to prevent its resolution by
						// the cnab-go library, which doesn't support this parameter type
						pset.Parameters[i] = pset.Parameters[len(pset.Parameters)-1]
						pset.Parameters = pset.Parameters[:len(pset.Parameters)-1]
					}
				}
			}
		}

		rc, err := p.Parameters.ResolveAll(pset)
		if err != nil {
			return nil, err
		}

		for k, v := range rc {
			resolvedParameters[k] = v
		}
	}

	return resolvedParameters, nil
}

func (p *Porter) loadParameterFromFile(path string) (parameters.ParameterSet, error) {
	data, err := p.FileSystem.ReadFile(path)
	if err != nil {
		return parameters.ParameterSet{}, errors.Wrapf(err, "could not read file %s", path)
	}

	var cs parameters.ParameterSet
	err = json.Unmarshal(data, &cs)
	return cs, errors.Wrapf(err, "error loading parameter set in %s", path)
}
