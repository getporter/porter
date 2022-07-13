package porter

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/editor"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/generator"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/olekukonko/tablewriter"
	"go.opentelemetry.io/otel/attribute"
)

// CredentialShowOptions represent options for Porter's credential show command
type CredentialShowOptions struct {
	printer.PrintOptions
	Name      string
	Namespace string
}

type CredentialEditOptions struct {
	Name      string
	Namespace string
}

// ListCredentials lists saved credential sets.
func (p *Porter) ListCredentials(ctx context.Context, opts ListOptions) ([]storage.CredentialSet, error) {
	return p.Credentials.ListCredentialSets(ctx, storage.ListOptions{
		Namespace: opts.GetNamespace(),
		Name:      opts.Name,
		Labels:    opts.ParseLabels(),
		Skip:      opts.Skip,
		Limit:     opts.Limit,
	})
}

// PrintCredentials prints saved credential sets.
func (p *Porter) PrintCredentials(ctx context.Context, opts ListOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	creds, err := p.ListCredentials(ctx, opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, creds)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, creds)
	case printer.FormatPlaintext:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		printCredRow :=
			func(v interface{}) []string {
				cr, ok := v.(storage.CredentialSet)
				if !ok {
					return nil
				}
				return []string{cr.Namespace, cr.Name, tp.Format(cr.Status.Modified)}
			}
		return printer.PrintTable(p.Out, creds, printCredRow,
			"NAMESPACE", "NAME", "MODIFIED")
	default:
		return span.Error(fmt.Errorf("invalid format: %s", opts.Format))
	}
}

// CredentialsOptions are the set of options available to Porter.GenerateCredentials
type CredentialOptions struct {
	BundleReferenceOptions
	Silent bool
	Labels []string
}

func (o CredentialOptions) ParseLabels() map[string]string {
	return parseLabels(o.Labels)
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (o *CredentialOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	err := o.validateCredName(args)
	if err != nil {
		return err
	}

	return o.BundleReferenceOptions.Validate(ctx, args, p)
}

func (o *CredentialOptions) validateCredName(args []string) error {
	if len(args) == 1 {
		o.Name = args[0]
	} else if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, the credential name, but multiple were received: %s", args)
	}
	return nil
}

// GenerateCredentials builds a new credential set based on the given options. This can be either
// a silent build, based on the opts.Silent flag, or interactive using a survey. Returns an
// error if unable to generate credentials
func (p *Porter) GenerateCredentials(ctx context.Context, opts CredentialOptions) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("reference", opts.Reference))
	defer span.EndSpan()

	bundleRef, err := p.resolveBundleReference(ctx, &opts.BundleReferenceOptions)
	if err != nil {
		return err
	}

	name := opts.Name
	if name == "" {
		name = bundleRef.Definition.Name
	}
	genOpts := generator.GenerateCredentialsOptions{
		GenerateOptions: generator.GenerateOptions{
			Name:      name,
			Namespace: opts.Namespace,
			Labels:    opts.ParseLabels(),
			Silent:    opts.Silent,
		},
		Credentials: bundleRef.Definition.Credentials,
	}
	span.Infof("Generating new credential %s from bundle %s\n", genOpts.Name, bundleRef.Definition.Name)
	span.Infof("==> %d credentials required for bundle %s\n", len(genOpts.Credentials), bundleRef.Definition.Name)

	cs, err := generator.GenerateCredentials(genOpts)
	if err != nil {
		return span.Error(fmt.Errorf("unable to generate credentials: %w", err))
	}

	cs.Status.Created = time.Now()
	cs.Status.Modified = cs.Status.Created

	err = p.Credentials.UpsertCredentialSet(ctx, cs)
	if err != nil {
		return span.Error(fmt.Errorf("unable to save credentials: %w", err))
	}

	return nil
}

// Validate validates the args provided to Porter's credential show command
func (o *CredentialShowOptions) Validate(args []string) error {
	if err := validateCredentialName(args); err != nil {
		return err
	}
	o.Name = args[0]
	return o.ParseFormat()
}

// Validate validates the args provided to Porter's credential edit command
func (o *CredentialEditOptions) Validate(args []string) error {
	if err := validateCredentialName(args); err != nil {
		return err
	}
	o.Name = args[0]
	return nil
}

// EditCredential edits the credentials of the provided name.
func (p *Porter) EditCredential(ctx context.Context, opts CredentialEditOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	credSet, err := p.Credentials.GetCredentialSet(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	// TODO(carolynvs): support editing in yaml, json or toml
	contents, err := encoding.MarshalYaml(credSet)
	if err != nil {
		return span.Error(fmt.Errorf("unable to load credentials: %w", err))
	}

	editor := editor.New(p.Context, fmt.Sprintf("porter-%s.yaml", credSet.Name), contents)
	output, err := editor.Run(ctx)
	if err != nil {
		return span.Error(fmt.Errorf("unable to open editor to edit credentials: %w", err))
	}

	err = encoding.UnmarshalYaml(output, &credSet)
	if err != nil {
		return span.Error(fmt.Errorf("unable to process credentials: %w", err))
	}

	err = p.Credentials.Validate(ctx, credSet)
	if err != nil {
		return span.Error(fmt.Errorf("credentials are invalid: %w", err))
	}

	credSet.Status.Modified = time.Now()
	err = p.Credentials.UpdateCredentialSet(ctx, credSet)
	if err != nil {
		return span.Error(fmt.Errorf("unable to save credentials: %w", err))
	}

	return nil
}

type DisplayCredentialSet struct {
	// SchemaType helps when we export the definition so editors can detect the type of document, it's not used by porter.
	SchemaType            string `json:"schemaType" yaml:"schemaType"`
	storage.CredentialSet `yaml:",inline"`
}

// ShowCredential shows the credential set corresponding to the provided name, using
// the provided printer.PrintOptions for display.
func (p *Porter) ShowCredential(ctx context.Context, opts CredentialShowOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	cs, err := p.Credentials.GetCredentialSet(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	credSet := DisplayCredentialSet{
		SchemaType:    "CredentialSet",
		CredentialSet: cs,
	}

	switch opts.Format {
	case printer.FormatJson, printer.FormatYaml:
		result, err := encoding.Marshal(string(opts.Format), credSet)
		if err != nil {
			return err
		}

		// Note that we are not using span.Info because the command's output must go to standard out
		fmt.Fprintln(p.Out, string(result))
		return nil
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

		// Iterate through all CredentialStrategies and add to rows
		for _, cs := range credSet.Credentials {
			rows = append(rows, []string{cs.Name, cs.Source.Value, cs.Source.Key})
		}

		// Build and configure our tablewriter
		table := tablewriter.NewWriter(p.Out)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
		table.SetAutoFormatHeaders(false)

		// First, print the CredentialSet metadata
		// Note that we are not using span.Info because the command's output must go to standard out
		fmt.Fprintf(p.Out, "Name: %s\n", credSet.Name)
		fmt.Fprintf(p.Out, "Namespace: %s\n", credSet.Namespace)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(credSet.Status.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n\n", tp.Format(credSet.Status.Modified))

		// Print labels, if any
		if len(credSet.Labels) > 0 {
			fmt.Fprintln(p.Out, "Labels:")

			for k, v := range credSet.Labels {
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
		return span.Error(fmt.Errorf("invalid format: %s", opts.Format))
	}
}

// CredentialDeleteOptions represent options for Porter's credential delete command
type CredentialDeleteOptions struct {
	Name      string
	Namespace string
}

// DeleteCredential deletes the credential set corresponding to the provided
// names.
func (p *Porter) DeleteCredential(ctx context.Context, opts CredentialDeleteOptions) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("namespace", opts.Namespace),
		attribute.String("name", opts.Name),
	)
	defer span.EndSpan()

	err := p.Credentials.RemoveCredentialSet(ctx, opts.Namespace, opts.Name)
	if errors.Is(err, storage.ErrNotFound{}) {
		span.Debug("nothing to remove, credential already does not exist")
		return nil
	}
	if err != nil {
		return span.Error(fmt.Errorf("unable to delete credential set: %w", err))
	}

	return nil
}

// Validate validates the args provided Porter's credential delete command
func (o *CredentialDeleteOptions) Validate(args []string) error {
	if err := validateCredentialName(args); err != nil {
		return err
	}
	o.Name = args[0]
	return nil
}

func validateCredentialName(args []string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("no credential name was specified")
	case 1:
		return nil
	default:
		return fmt.Errorf("only one positional argument may be specified, the credential name, but multiple were received: %s", args)
	}
}

func (p *Porter) CredentialsApply(ctx context.Context, o ApplyOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debugf("Reading input file %s...\n", o.File)
	namespace, err := p.getNamespaceFromFile(o)
	if err != nil {
		return span.Error(err)
	}

	var creds storage.CredentialSet
	err = encoding.UnmarshalFile(p.FileSystem, o.File, &creds)
	if err != nil {
		return span.Error(fmt.Errorf("could not load %s as a credential set: %w", o.File, err))
	}

	if err = creds.Validate(); err != nil {
		return span.Error(fmt.Errorf("invalid credential set: %w", err))
	}

	creds.Namespace = namespace
	creds.Status.Modified = time.Now()

	err = p.Credentials.Validate(ctx, creds)
	if err != nil {
		return span.Error(fmt.Errorf("credential set is invalid: %w", err))
	}

	err = p.Credentials.UpsertCredentialSet(ctx, creds)
	if err != nil {
		return err
	}

	span.Infof("Applied %s credential set", creds)
	return nil
}

func (p *Porter) getNamespaceFromFile(o ApplyOptions) (string, error) {
	// Check if the namespace was set in the file, if not, use the namespace set on the command
	var raw map[string]interface{}
	err := encoding.UnmarshalFile(p.FileSystem, o.File, &raw)
	if err != nil {
		return "", fmt.Errorf("invalid file '%s': %w", o.File, err)
	}

	if rawNamespace, ok := raw["namespace"]; ok {
		if ns, ok := rawNamespace.(string); ok {
			return ns, nil
		} else {
			return "", errors.New("invalid namespace specified in file, must be a string")
		}
	}

	return o.Namespace, nil
}

// CredentialCreateOptions represent options for Porter's credential create command
type CredentialCreateOptions struct {
	FileName   string
	OutputType string
}

func (o *CredentialCreateOptions) Validate(args []string) error {
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

func (p *Porter) CreateCredential(ctx context.Context, opts CredentialCreateOptions) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if opts.OutputType == "" {
		opts.OutputType = strings.Trim(filepath.Ext(opts.FileName), ".")
	}

	if opts.FileName == "" {
		if opts.OutputType == "" {
			opts.OutputType = "yaml"
		}

		switch opts.OutputType {
		case "json":
			credentialSet, err := p.Templates.GetCredentialSetJSON()
			if err != nil {
				return err
			}

			// Note that we are not using span.Info because this must be printed to stdout
			fmt.Fprintln(p.Out, string(credentialSet))

			return nil
		case "yaml", "yml":
			credentialSet, err := p.Templates.GetCredentialSetYAML()
			if err != nil {
				return err
			}

			// Note that we are not using span.Info because this must be printed to stdout
			fmt.Fprintln(p.Out, string(credentialSet))

			return nil
		default:
			return span.Error(newUnsupportedFormatError(opts.OutputType))
		}

	}

	span.Info("creating porter credential set in the current directory")

	switch opts.OutputType {
	case "json":
		err := p.CopyTemplate(p.Templates.GetCredentialSetJSON, opts.FileName)
		return span.Error(err)
	case "yaml", "yml":
		err := p.CopyTemplate(p.Templates.GetCredentialSetYAML, opts.FileName)
		return span.Error(err)
	default:
		return span.Error(newUnsupportedFormatError(opts.OutputType))
	}
}

func newUnsupportedFormatError(format string) error {
	return fmt.Errorf("unsupported format %s. Supported formats are: yaml and json", format)
}
