package porter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/pkg/errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	depsv1 "get.porter.sh/porter/pkg/cnab/dependencies/v1"
	"get.porter.sh/porter/pkg/editor"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/generator"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/olekukonko/tablewriter"
	"go.mongodb.org/mongo-driver/bson"
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
func (p *Porter) ListParameters(ctx context.Context, opts ListOptions) ([]storage.ParameterSet, error) {
	return p.Parameters.ListParameterSets(ctx, storage.ListOptions{
		Namespace: opts.GetNamespace(),
		Name:      opts.Name,
		Labels:    opts.ParseLabels(),
		Skip:      opts.Skip,
		Limit:     opts.Limit,
	})
}

// PrintParameters prints saved parameter sets.
func (p *Porter) PrintParameters(ctx context.Context, opts ListOptions) error {
	params, err := p.ListParameters(ctx, opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, params)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, params)
	case printer.FormatPlaintext:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		printParamRow :=
			func(v interface{}) []string {
				cr, ok := v.(storage.ParameterSet)
				if !ok {
					return nil
				}
				return []string{cr.Namespace, cr.Name, tp.Format(cr.Status.Modified)}
			}
		return printer.PrintTable(p.Out, params, printParamRow,
			"NAMESPACE", "NAME", "MODIFIED")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

// ParameterOptions represent generic/base options for a Porter parameters command
type ParameterOptions struct {
	BundleReferenceOptions
	Silent bool
	Labels []string
}

func (o ParameterOptions) ParseLabels() map[string]string {
	return parseLabels(o.Labels)
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (o *ParameterOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	err := o.validateParamName(args)
	if err != nil {
		return err
	}

	return o.BundleReferenceOptions.Validate(ctx, args, p)
}

func (o *ParameterOptions) validateParamName(args []string) error {
	if len(args) == 1 {
		o.Name = args[0]
	} else if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, the parameter set name, but multiple were received: %s", args)
	}
	return nil
}

// GenerateParameters builds a new parameter set based on the given options. This can be either
// a silent build, based on the opts.Silent flag, or interactive using a survey. Returns an
// error if unable to generate parameters
func (p *Porter) GenerateParameters(ctx context.Context, opts ParameterOptions) error {
	bundleRef, err := p.resolveBundleReference(ctx, &opts.BundleReferenceOptions)

	if err != nil {
		return err
	}
	name := opts.Name
	if name == "" {
		name = bundleRef.Definition.Name
	}
	genOpts := generator.GenerateParametersOptions{
		GenerateOptions: generator.GenerateOptions{
			Name:      name,
			Namespace: opts.Namespace,
			Labels:    opts.ParseLabels(),
			Silent:    opts.Silent,
		},
		Bundle: bundleRef.Definition,
	}
	fmt.Fprintf(p.Out, "Generating new parameter set %s from bundle %s\n", genOpts.Name, bundleRef.Definition.Name)
	numExternalParams := 0

	for name := range bundleRef.Definition.Parameters {
		if !bundleRef.Definition.IsInternalParameter(name) {
			numExternalParams += 1
		}
	}

	fmt.Fprintf(p.Out, "==> %d parameter(s) declared for bundle %s\n", numExternalParams, bundleRef.Definition.Name)

	pset, err := genOpts.GenerateParameters()
	if err != nil {
		return fmt.Errorf("unable to generate parameter set: %w", err)
	}

	pset.Status.Created = time.Now()
	pset.Status.Modified = pset.Status.Created

	err = p.Parameters.UpsertParameterSet(ctx, pset)
	if err != nil {
		return fmt.Errorf("unable to save parameter set: %w", err)
	}

	return nil
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
func (p *Porter) EditParameter(ctx context.Context, opts ParameterEditOptions) error {
	paramSet, err := p.Parameters.GetParameterSet(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	contents, err := encoding.MarshalYaml(paramSet)
	if err != nil {
		return fmt.Errorf("unable to load parameter set: %w", err)
	}

	editor := editor.New(p.Context, fmt.Sprintf("porter-%s.yaml", paramSet.Name), contents)
	output, err := editor.Run(ctx)
	if err != nil {
		return fmt.Errorf("unable to open editor to edit parameter set: %w", err)
	}

	err = encoding.UnmarshalYaml(output, &paramSet)
	if err != nil {
		return fmt.Errorf("unable to process parameter set: %w", err)
	}

	err = p.Parameters.Validate(ctx, paramSet)
	if err != nil {
		return fmt.Errorf("parameter set is invalid: %w", err)
	}

	paramSet.Status.Modified = time.Now()
	err = p.Parameters.UpdateParameterSet(ctx, paramSet)
	if err != nil {
		return fmt.Errorf("unable to save parameter set: %w", err)
	}

	return nil
}

type DisplayParameterSet struct {
	// SchemaType helps when we export the definition so editors can detect the type of document, it's not used by porter.
	SchemaType           string `json:"schemaType" yaml:"schemaType"`
	storage.ParameterSet `yaml:",inline"`
}

// ShowParameter shows the parameter set corresponding to the provided name, using
// the provided printer.PrintOptions for display.
func (p *Porter) ShowParameter(ctx context.Context, opts ParameterShowOptions) error {
	ps, err := p.Parameters.GetParameterSet(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	paramSet := DisplayParameterSet{
		SchemaType:   "ParameterSet",
		ParameterSet: ps,
	}
	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, paramSet)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, paramSet)
	case printer.FormatPlaintext:
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
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(paramSet.Status.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n\n", tp.Format(paramSet.Status.Modified))

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
func (p *Porter) DeleteParameter(ctx context.Context, opts ParameterDeleteOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	err := p.Parameters.RemoveParameterSet(ctx, opts.Namespace, opts.Name)
	if errors.Is(err, storage.ErrNotFound{}) {
		span.Debug("Cannot remove parameter set because it already doesn't exist")
		return nil
	}
	if err != nil {
		return span.Error(fmt.Errorf("unable to delete parameter set: %w", err))
	}

	return nil
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
		return fmt.Errorf("no parameter set name was specified")
	case 1:
		return nil
	default:
		return fmt.Errorf("only one positional argument may be specified, the parameter set name, but multiple were received: %s", args)
	}
}

// loadParameterSets loads parameter values per their parameter set strategies
func (p *Porter) loadParameterSets(ctx context.Context, bun cnab.ExtendedBundle, namespace string, params []string) (secrets.Set, error) {
	resolvedParameters := secrets.Set{}
	for _, name := range params {

		// Try to get the params in the local namespace first, fallback to the global creds
		query := storage.FindOptions{
			Sort: []string{"-namespace"},
			Filter: bson.M{
				"name": name,
				"$or": []bson.M{
					{"namespace": ""},
					{"namespace": namespace},
				},
			},
		}
		store := p.Parameters.GetDataStore()

		var pset storage.ParameterSet
		err := store.FindOne(ctx, storage.CollectionParameters, query, &pset)
		if err != nil {
			return nil, err
		}

		// A parameter may correspond to a Porter-specific parameter type of 'file'
		// If so, add value (filepath) directly to map and remove from pset
		for paramName, paramDef := range bun.Parameters {
			paramSchema, ok := bun.Definitions[paramDef.Definition]
			if !ok {
				return nil, fmt.Errorf("definition %s not defined in bundle", paramDef.Definition)
			}

			if bun.IsFileType(paramSchema) {
				for i, param := range pset.Parameters {
					if param.Name == paramName {
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

		rc, err := p.Parameters.ResolveAll(ctx, pset)
		if err != nil {
			return nil, err
		}

		for k, v := range rc {
			resolvedParameters[k] = v
		}
	}

	return resolvedParameters, nil
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

func NewDisplayValuesFromParameters(bun cnab.ExtendedBundle, params map[string]interface{}) DisplayValues {
	// Iterate through all Bundle Outputs, fetch their metadata
	// via their corresponding Definitions and add to rows
	displayParams := make(DisplayValues, 0, len(params))
	for name, value := range params {
		def, ok := bun.Parameters[name]
		if !ok || bun.IsInternalParameter(name) {
			continue
		}

		dp := &DisplayValue{Name: name}
		dp.SetValue(value)

		schema, ok := bun.Definitions[def.Definition]
		if ok {
			dp.Type = bun.GetParameterType(schema)
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

func (p *Porter) ParametersApply(ctx context.Context, o ApplyOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debugf("Reading input file %s...", o.File)
	namespace, err := p.getNamespaceFromFile(o)
	if err != nil {
		return span.Error(err)
	}

	var params storage.ParameterSet
	err = encoding.UnmarshalFile(p.FileSystem, o.File, &params)
	if err != nil {
		return span.Error(fmt.Errorf("could not load %s as a parameter set: %w", o.File, err))
	}

	if err = params.Validate(); err != nil {
		return span.Error(fmt.Errorf("invalid parameter set: %w", err))
	}

	params.Namespace = namespace
	params.Status.Modified = time.Now()

	err = p.Parameters.Validate(ctx, params)
	if err != nil {
		return span.Error(fmt.Errorf("parameter set is invalid: %w", err))
	}

	err = p.Parameters.UpsertParameterSet(ctx, params)
	if err != nil {
		return err
	}

	span.Infof("Applied %s parameter set", params)
	return nil
}

// resolveParameters accepts a set of parameter assignments and combines them
// with parameter sources and default parameter values to create a full set
// of parameters.
func (p *Porter) resolveParameters(ctx context.Context, installation storage.Installation, bun cnab.ExtendedBundle, action string, params map[string]string) (map[string]interface{}, error) {
	mergedParams := make(secrets.Set, len(params))
	paramSources, err := p.resolveParameterSources(ctx, bun, installation)
	if err != nil {
		return nil, err
	}

	for key, val := range paramSources {
		mergedParams[key] = val
	}

	// Apply user supplied parameter overrides last
	for key, rawValue := range params {
		param, ok := bun.Parameters[key]
		if !ok {
			return nil, fmt.Errorf("parameter %s not defined in bundle", key)
		}

		def, ok := bun.Definitions[param.Definition]
		if !ok {
			return nil, fmt.Errorf("definition %s not defined in bundle", param.Definition)
		}

		// Apply porter specific conversions, like retrieving file contents
		value, err := p.getUnconvertedValueFromRaw(bun, def, key, rawValue)
		if err != nil {
			return nil, err
		}

		mergedParams[key] = value
	}

	// Now convert all parameters which are currently strings into the
	// proper type for the parameter, e.g. "false" -> false
	typedParams := make(map[string]interface{}, len(mergedParams))
	for key, unconverted := range mergedParams {
		param, ok := bun.Parameters[key]
		if !ok {
			return nil, fmt.Errorf("parameter %s not defined in bundle", key)
		}

		def, ok := bun.Definitions[param.Definition]
		if !ok {
			return nil, fmt.Errorf("definition %s not defined in bundle", param.Definition)
		}

		if def.Type != nil {
			value, err := def.ConvertValue(unconverted)
			if err != nil {
				return nil, fmt.Errorf("unable to convert parameter's %s value %s to the destination parameter type %s: %w", key, unconverted, def.Type, err)
			}
			typedParams[key] = value
		} else {
			// bundle dependency parameters can be any type, not sure we have a solid way to do a typed conversion
			typedParams[key] = unconverted
		}

	}

	return bundle.ValuesOrDefaults(typedParams, &bun.Bundle, action)
}

func (p *Porter) getUnconvertedValueFromRaw(b cnab.ExtendedBundle, def *definition.Schema, key, rawValue string) (string, error) {
	// the parameter value (via rawValue) may represent a file on the local filesystem
	if b.IsFileType(def) {
		if _, err := p.FileSystem.Stat(rawValue); err == nil {
			bytes, err := p.FileSystem.ReadFile(rawValue)
			if err != nil {
				return "", fmt.Errorf("unable to read file parameter %s: %w", key, err)
			}
			return base64.StdEncoding.EncodeToString(bytes), nil
		}
	}
	return rawValue, nil
}

func (p *Porter) resolveParameterSources(ctx context.Context, bun cnab.ExtendedBundle, installation storage.Installation) (secrets.Set, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if !bun.HasParameterSources() {
		span.Debug("No parameter sources defined, skipping")
		return nil, nil
	}

	span.Debug("Resolving parameter sources...")
	parameterSources, err := bun.ReadParameterSources()
	if err != nil {
		return nil, span.Error(err)
	}

	values := secrets.Set{}
	for parameterName, parameterSource := range parameterSources {
		span.Debugf("Resolving parameter source %s", parameterName)
		for _, rawSource := range parameterSource.ListSourcesByPriority() {
			var installationName string
			var outputName string
			var mount *cnab.MountParameterSourceDefn
			switch source := rawSource.(type) {
			case cnab.OutputParameterSource:
				installationName = installation.Name
				outputName = source.OutputName
			case cnab.DependencyOutputParameterSource:
				// TODO(carolynvs): does this need to take namespace into account
				installationName = depsv1.BuildPrerequisiteInstallationName(installation.Name, source.Dependency)
				outputName = source.OutputName
			case cnab.MountParameterSourceDefn:
				installationName = installation.Name
				mount = &source
			}

			output, err := p.Installations.GetLastOutput(ctx, installation.Namespace, installationName, outputName)
			if err != nil {
				// When we can't find the output, it may be a directory parameter, or we may have to find it some other way
				if !errors.Is(err, storage.ErrNotFound{}) {
					// Otherwise, something else has happened, perhaps bad data or connectivity problems, we can't ignore it
					return nil, errors.Wrapf(err, "could not set parameter %s from output %s of %s", parameterName, outputName, installation)
				}
			}

			if output.Key != "" {
				resolved, err := p.Sanitizer.RestoreOutput(ctx, output)
				if err != nil {
					return nil, span.Error(fmt.Errorf("could not resolve %s's output %s: %w", installation, outputName, err))
				}
				output = resolved
			}

			param, ok := bun.Parameters[parameterName]
			if !ok {
				return nil, span.Error(fmt.Errorf("resolveParameterSources:  %s not defined in bundle", parameterName))
			}

			def, ok := bun.Definitions[param.Definition]
			if !ok {
				return nil, span.Error(fmt.Errorf("definition %s not defined in bundle", param.Definition))
			}

			if bun.IsFileType(def) {
				values[parameterName] = base64.StdEncoding.EncodeToString(output.Value)
			} else if bun.IsDirType(def) && mount != nil {
				p := bun.Parameters[parameterName]
				// p.Definition = config.CustomPorterKey + ".directory"
				bun.Parameters[parameterName] = p
				if bytes, err := json.Marshal(mount); err == nil {
					values[parameterName] = string(bytes)
				} else {
					return nil, fmt.Errorf("Could not marshal source for definition %s", param.Definition)
				}
			} else {
				values[parameterName] = string(output.Value)
			}

			span.Debugf("Injected installation %s output %s as parameter %s", installation, outputName, parameterName)
		}
	}

	return values, nil
}

// ParameterCreateOptions represent options for Porter's parameter create command
type ParameterCreateOptions struct {
	FileName   string
	OutputType string
}

func (o *ParameterCreateOptions) Validate(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, fileName, but multiple were received: %s", args)
	}

	if len(args) > 0 {
		o.FileName = args[0]
	}

	if o.OutputType == "" && o.FileName != "" && strings.Trim(filepath.Ext(o.FileName), ".") == "" {
		return errors.New("could not detect the file format from the file extension (.txt). Specify the format with --output")
	}

	return nil
}

func (p *Porter) CreateParameter(opts ParameterCreateOptions) error {
	if opts.OutputType == "" {
		opts.OutputType = strings.Trim(filepath.Ext(opts.FileName), ".")
	}

	if opts.FileName == "" {
		if opts.OutputType == "" {
			opts.OutputType = "yaml"
		}

		switch opts.OutputType {
		case "json":
			parameterSet, err := p.Templates.GetParameterSetJSON()
			if err != nil {
				return err
			}
			fmt.Fprintln(p.Out, string(parameterSet))

			return nil
		case "yaml", "yml":
			parameterSet, err := p.Templates.GetParameterSetYAML()
			if err != nil {
				return err
			}
			fmt.Fprintln(p.Out, string(parameterSet))

			return nil
		default:
			return newUnsupportedFormatError(opts.OutputType)
		}

	}

	fmt.Fprintln(p.Err, "creating porter parameter set in the current directory")

	switch opts.OutputType {
	case "json":
		return p.CopyTemplate(p.Templates.GetParameterSetJSON, opts.FileName)
	case "yaml", "yml":
		return p.CopyTemplate(p.Templates.GetParameterSetYAML, opts.FileName)
	default:
		return newUnsupportedFormatError(opts.OutputType)
	}
}
