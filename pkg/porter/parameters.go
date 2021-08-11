package porter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab/extensions"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/editor"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/generator"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"

	dtprinter "github.com/carolynvs/datetime-printer"
	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// ParameterShowOptions represent options for Porter's parameter show command
type ParameterShowOptions struct {
	printer.PrintOptions
	Name      string
	Namespace string
}

// ParameterEditOptions represent iptions for Porter's parameter edit command
type ParameterEditOptions struct {
	Name      string
	Namespace string
}

// ListParameters lists saved parameter sets.
func (p *Porter) ListParameters(opts ListOptions) ([]parameters.ParameterSet, error) {
	return p.Parameters.ListParameterSets(opts.GetNamespace(), opts.Name, opts.ParseLabels())
}

// PrintParameters prints saved parameter sets.
func (p *Porter) PrintParameters(opts ListOptions) error {
	params, err := p.ListParameters(opts)
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
				return []interface{}{cr.Namespace, cr.Name, tp.Format(cr.Modified)}
			}
		return printer.PrintTable(p.Out, params, printParamRow,
			"NAMESPACE", "NAME", "MODIFIED")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

// ParameterOptions represent generic/base options for a Porter parameters command
type ParameterOptions struct {
	BundleActionOptions
	Silent bool
	Labels []string
}

func (o ParameterOptions) ParseLabels() map[string]string {
	return parseLabels(o.Labels)
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (g *ParameterOptions) Validate(args []string, cxt *context.Context) error {
	g.checkForDeprecatedTagValue()

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
	err := p.prepullBundleByReference(&opts.BundleActionOptions)
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
			Name:      name,
			Namespace: opts.Namespace,
			Labels:    opts.ParseLabels(),
			Silent:    opts.Silent,
		},
		Bundle: bundle,
	}
	fmt.Fprintf(p.Out, "Generating new parameter set %s from bundle %s\n", genOpts.Name, bundle.Name)
	numExternalParams := 0

	for name := range bundle.Parameters {
		if !parameters.IsInternal(name, bundle) {
			numExternalParams += 1
		}
	}

	fmt.Fprintf(p.Out, "==> %d parameter(s) declared for bundle %s\n", numExternalParams, bundle.Name)

	pset, err := genOpts.GenerateParameters()
	if err != nil {
		return errors.Wrap(err, "unable to generate parameter set")
	}

	pset.Created = time.Now()
	pset.Modified = pset.Created

	err = p.Parameters.UpsertParameterSet(pset)
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
	paramSet, err := p.Parameters.GetParameterSet(opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	contents, err := encoding.MarshalYaml(paramSet)
	if err != nil {
		return errors.Wrap(err, "unable to load parameter set")
	}

	editor := editor.New(p.Context, fmt.Sprintf("porter-%s.yaml", paramSet.Name), contents)
	output, err := editor.Run()
	if err != nil {
		return errors.Wrap(err, "unable to open editor to edit parameter set")
	}

	err = encoding.UnmarshalYaml(output, &paramSet)
	if err != nil {
		return errors.Wrap(err, "unable to process parameter set")
	}

	err = p.Parameters.Validate(paramSet)
	if err != nil {
		return errors.Wrap(err, "parameter set is invalid")
	}

	paramSet.Modified = time.Now()
	err = p.Parameters.UpdateParameterSet(paramSet)
	if err != nil {
		return errors.Wrap(err, "unable to save parameter set")
	}

	return nil
}

// ShowParameter shows the parameter set corresponding to the provided name, using
// the provided printer.PrintOptions for display.
func (p *Porter) ShowParameter(opts ParameterShowOptions) error {
	paramSet, err := p.Parameters.GetParameterSet(opts.Namespace, opts.Name)
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

		// Print labels, if any
		if len(paramSet.Labels) > 0 {
			fmt.Fprintln(p.Out, "Labels:")

			for k, v := range paramSet.Labels {
				fmt.Fprintf(p.Out, "  %s: %s\n", k, v)
			}
			fmt.Fprintln(p.Out)
		}

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
	Name      string
	Namespace string
}

// DeleteParameter deletes the parameter set corresponding to the provided
// names.
func (p *Porter) DeleteParameter(opts ParameterDeleteOptions) error {
	err := p.Parameters.RemoveParameterSet(opts.Namespace, opts.Name)
	if errors.Is(err, storage.ErrNotFound{}) {
		if p.Debug {
			fmt.Fprintln(p.Err, err)
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
func (p *Porter) loadParameterSets(namespace string, params []string) (secrets.Set, error) {
	resolvedParameters := secrets.Set{}
	for _, name := range params {
		var pset parameters.ParameterSet
		var err error
		// If name looks pathy, attempt to load from a file
		// Else, read from Porter's parameters store
		if strings.Contains(name, string(filepath.Separator)) {
			pset, err = p.loadParameterFromFile(name)
		} else {
			pset, err = p.Parameters.GetParameterSet(namespace, name)
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

type DisplayValue struct {
	Name      string      `json:"name" yaml:"name"`
	Type      string      `json:"type" yaml:"type"`
	Sensitive bool        `json:"sensitive" yaml:"sensitive"`
	Value     interface{} `json:"value" yaml:"value"`
}

func (v *DisplayValue) SetValue(value interface{}) {
	switch val := value.(type) {
	case []byte:
		v.Value = string(val)
	default:
		v.Value = val
	}
}

func (v DisplayValue) PrintValue() string {
	if v.Sensitive {
		return "******"
	}

	var printedValue string
	switch val := v.Value.(type) {
	case string:
		printedValue = val
	default:
		b, err := json.Marshal(v.Value)
		if err != nil {
			return "error rendering value"
		}
		printedValue = string(b)
	}
	return truncateString(printedValue, 60)
}

type DisplayValues []DisplayValue

func (v DisplayValues) Len() int {
	return len(v)
}

func (v DisplayValues) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v DisplayValues) Less(i, j int) bool {
	return v[i].Name < v[j].Name
}

func NewDisplayValuesFromParameters(bun bundle.Bundle, params map[string]interface{}) DisplayValues {
	// Iterate through all Bundle Outputs, fetch their metadata
	// via their corresponding Definitions and add to rows
	displayParams := make(DisplayValues, 0, len(params))
	for name, value := range params {
		def, ok := bun.Parameters[name]
		if !ok || parameters.IsInternal(name, bun) {
			continue
		}

		dp := &DisplayValue{Name: name}
		dp.SetValue(value)

		schema, ok := bun.Definitions[def.Definition]
		if ok {
			dp.Type = extensions.GetParameterType(bun, schema)
			if schema.WriteOnly != nil && *schema.WriteOnly {
				dp.Sensitive = true
			}
		} else {
			dp.Type = "unknown"
		}

		displayParams = append(displayParams, *dp)
	}

	sort.Sort(displayParams)
	return displayParams
}

func (p *Porter) printDisplayValuesTable(values []DisplayValue) error {
	// Build and configure our tablewriter for the outputs
	table := tablewriter.NewWriter(p.Out)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
	table.SetAutoFormatHeaders(false)

	table.SetHeader([]string{"Name", "Type", "Value"})
	for _, param := range values {
		table.Append([]string{param.Name, param.Type, param.PrintValue()})
	}
	table.Render()

	return nil
}
