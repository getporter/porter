package porter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
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
func (p *Porter) ListParameters(ctx context.Context, opts ListOptions) ([]DisplayParameterSet, error) {
	listOpts := storage.ListOptions{
		Namespace: opts.GetNamespace(),
		Name:      opts.Name,
		Labels:    opts.ParseLabels(),
		Skip:      opts.Skip,
		Limit:     opts.Limit,
	}
	results, err := p.Parameters.ListParameterSets(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	displayResults := make([]DisplayParameterSet, len(results))
	for i, ps := range results {
		displayResults[i] = NewDisplayParameterSet(ps)
	}

	return displayResults, nil
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
		paramsSets := [][]string{}
		for _, ps := range params {
			for _, param := range ps.Parameters {
				list := []string{}
				list = append(list, ps.Namespace, param.Name, param.Source.Strategy, param.Source.Hint, tp.Format(ps.Status.Modified))
				paramsSets = append(paramsSets, list)
			}
		}
		return printer.PrintTableParameterSet(p.Out, paramsSets,
			"NAMESPACE", "NAME", "TYPE", "VALUE", "MODIFIED")
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
	bundleRef, err := opts.GetBundleReference(ctx, p)
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
	storage.ParameterSet `yaml:",inline"`
}

func NewDisplayParameterSet(ps storage.ParameterSet) DisplayParameterSet {
	ds := DisplayParameterSet{ParameterSet: ps}
	ds.SchemaType = storage.SchemaTypeParameterSet
	return ds
}

// ShowParameter shows the parameter set corresponding to the provided name, using
// the provided printer.PrintOptions for display.
func (p *Porter) ShowParameter(ctx context.Context, opts ParameterShowOptions) error {
	ps, err := p.Parameters.GetParameterSet(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	paramSet := NewDisplayParameterSet(ps)

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
			rows = append(rows, []string{pset.Name, pset.Source.Hint, pset.Source.Strategy})
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
						resolvedParameters[param.Name] = param.Source.Hint
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

func (v DisplayValues) Get(name string) (DisplayValue, bool) {
	for _, entry := range v {
		if entry.Name == name {
			return entry, true
		}
	}

	return DisplayValue{}, false
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

	var params DisplayParameterSet
	err = encoding.UnmarshalFile(p.FileSystem, o.File, &params)
	if err != nil {
		return span.Error(fmt.Errorf("could not load %s as a parameter set: %w", o.File, err))
	}

	checkStrategy := p.GetSchemaCheckStrategy(ctx)
	if err = params.Validate(ctx, checkStrategy); err != nil {
		return span.Error(fmt.Errorf("invalid parameter set: %w", err))
	}

	params.Namespace = namespace
	params.Status.Modified = time.Now()

	err = p.Parameters.Validate(ctx, params.ParameterSet)
	if err != nil {
		return span.Error(fmt.Errorf("parameter set is invalid: %w", err))
	}

	err = p.Parameters.UpsertParameterSet(ctx, params.ParameterSet)
	if err != nil {
		return err
	}

	span.Infof("Applied %s parameter set", params)
	return nil
}

// finalizeParameters accepts a set of resolved parameters and combines them
// with parameter sources and default parameter values to create a full set
// of parameters that are defined in proper Go types, and not strings.
func (p *Porter) finalizeParameters(ctx context.Context, installation storage.Installation, bun cnab.ExtendedBundle, action string, params map[string]string) (map[string]interface{}, error) {
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
				return "", fmt.Errorf("unable to read file parameter %s at %s: %w", key, rawValue, err)
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
			switch source := rawSource.(type) {
			case cnab.OutputParameterSource:
				installationName = installation.Name
				outputName = source.OutputName
			case cnab.DependencyOutputParameterSource:
				// TODO(carolynvs): does this need to take namespace into account
				installationName = bun.BuildPrerequisiteInstallationName(installation.Name, source.Dependency)
				outputName = source.OutputName
			}

			output, err := p.Installations.GetLastOutput(ctx, installation.Namespace, installationName, outputName)
			if err != nil {
				// When we can't find the output, skip it and let the parameter be set another way
				if errors.Is(err, storage.ErrNotFound{}) {
					span.Debugf("No previous output found for %s from %s/%s", outputName, installation.Namespace, installationName)
					continue
				}
				// Otherwise, something else has happened, perhaps bad data or connectivity problems, we can't ignore it
				return nil, span.Error(fmt.Errorf("could not set parameter %s from output %s of %s: %w", parameterName, outputName, installation, err))
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

// applyActionOptionsToInstallation applies the specified action (e.g. install/upgrade) to an installation record.
// This resolves the parameters to their final form to be passed to the CNAB runtime, and modifies the specified installation record.
// You must sanitize the parameters before saving the installation so that sensitive values are not saved to the database.
func (p *Porter) applyActionOptionsToInstallation(ctx context.Context, ba BundleAction, inst *storage.Installation) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	o := ba.GetOptions()

	bundleRef, err := o.GetBundleReference(ctx, p)
	if err != nil {
		return err
	}
	bun := bundleRef.Definition

	// Update the installation with metadata from the options
	inst.TrackBundle(bundleRef.Reference)
	inst.Status.Modified = time.Now()

	// Remove installation parameters no longer present in the bundle
	if inst.Parameters.Parameters != nil {
		updatedInstParams := make(secrets.StrategyList, 0, len(inst.Parameters.Parameters))
		for _, param := range inst.Parameters.Parameters {
			if _, ok := bun.Parameters[param.Name]; ok {
				updatedInstParams = append(updatedInstParams, param)
			}
		}
		inst.Parameters.Parameters = updatedInstParams
	}

	//
	// 1. Record the parameter and credential sets used on the installation
	// if none were specified, reuse the previous sets from the installation
	//
	span.SetAttributes(
		tracing.ObjectAttribute("override-parameter-sets", o.ParameterSets),
		tracing.ObjectAttribute("override-credential-sets", o.CredentialIdentifiers))
	if len(o.ParameterSets) > 0 {
		inst.ParameterSets = o.ParameterSets
	}
	if len(o.CredentialIdentifiers) > 0 {
		inst.CredentialSets = o.CredentialIdentifiers
	}

	//
	// 2. Parse parameter flags from the command line and apply to the installation as overrides
	//
	// This contains resolved sensitive values, so only trace it in special dev builds (nothing is traced for release builds)
	span.SetSensitiveAttributes(tracing.ObjectAttribute("override-parameters", o.Params))
	parsedOverrides, err := storage.ParseVariableAssignments(o.Params)
	if err != nil {
		return err
	}

	// Default the porter-debug param to --debug
	if o.DebugMode {
		parsedOverrides["porter-debug"] = "true"
	}

	// Apply overrides on to of any pre-existing parameters that were specified previously
	if len(parsedOverrides) > 0 {
		for name, value := range parsedOverrides {
			// Do not resolve parameters from dependencies
			if strings.Contains(name, "#") {
				continue
			}

			// Replace previous value if present
			replaced := false
			paramStrategy := storage.ValueStrategy(name, value)
			for i, existingParam := range inst.Parameters.Parameters {
				if existingParam.Name == name {
					inst.Parameters.Parameters[i] = paramStrategy
					replaced = true
				}
			}
			if !replaced {
				inst.Parameters.Parameters = append(inst.Parameters.Parameters, paramStrategy)
			}
		}

		// Keep the parameter overrides sorted, so that comparisons and general troubleshooting is easier
		sort.Sort(inst.Parameters.Parameters)
	}
	// This contains resolved sensitive values, so only trace it in special dev builds (nothing is traced for release builds)
	span.SetSensitiveAttributes(tracing.ObjectAttribute("merged-installation-parameters", inst.Parameters.Parameters))

	//
	// 3. Resolve named parameter sets
	//
	resolvedParams, err := p.loadParameterSets(ctx, bun, o.Namespace, inst.ParameterSets)
	if err != nil {
		return fmt.Errorf("unable to process provided parameter sets: %w", err)
	}

	// This contains resolved sensitive values, so only trace it in special dev builds (nothing is traced for release builds)
	span.SetSensitiveAttributes(tracing.ObjectAttribute("resolved-parameter-sets-keys", resolvedParams))

	//
	// 4. Resolve the installation's internal parameter set
	resolvedOverrides, err := p.Parameters.ResolveAll(ctx, inst.Parameters)
	if err != nil {
		return err
	}

	// This contains resolved sensitive values, so only trace it in special dev builds (nothing is traced for release builds)
	span.SetSensitiveAttributes(tracing.ObjectAttribute("resolved-installation-parameters", inst.Parameters.Parameters))

	//
	// 5. Apply the overrides on top of the parameter sets
	//
	for k, v := range resolvedOverrides {
		resolvedParams[k] = v
	}

	//
	// 6. Separate out params for the root bundle from the ones intended for dependencies
	//    This only applies to the dep v1 implementation, in dep v2 you can't specify rando params for deps
	//
	o.depParams = make(map[string]string)
	for k, v := range resolvedParams {
		if strings.Contains(k, "#") {
			o.depParams[k] = v
			delete(resolvedParams, k)
		}
	}

	// This contains resolved sensitive values, so only trace it in special dev builds (nothing is traced for release builds)
	span.SetSensitiveAttributes(tracing.ObjectAttribute("user-specified-parameters", resolvedParams))

	//
	// 7. When a parameter is not specified, fallback to a parameter source or default
	//
	finalParams, err := p.finalizeParameters(ctx, *inst, bun, ba.GetAction(), resolvedParams)
	if err != nil {
		return err
	}

	// This contains resolved sensitive values, so only trace it in special dev builds (nothing is traced for release builds)
	span.SetSensitiveAttributes(tracing.ObjectAttribute("final-parameters", finalParams))

	// Remember the final set of parameters so we don't have to resolve them more than once
	o.finalParams = finalParams

	// Ensure we aren't storing any secrets on the installation resource
	if err = p.sanitizeInstallation(ctx, inst, bundleRef.Definition); err != nil {
		return err
	}

	// re-validate the installation since we modified it here
	return inst.Validate(ctx, p.GetSchemaCheckStrategy(ctx))
}
