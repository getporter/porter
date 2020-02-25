package porter

import (
	"encoding/json"
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/credentialsgenerator"
	"get.porter.sh/porter/pkg/editor"
	"get.porter.sh/porter/pkg/printer"
	"gopkg.in/yaml.v2"

	dtprinter "github.com/carolynvs/datetime-printer"
	credentials "github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/utils/crud"
	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// CredentialShowOptions represent options for Porter's credential show command
type CredentialShowOptions struct {
	printer.PrintOptions
	Name string
}

type CredentialEditOptions struct {
	Name string
}

// ListCredentials lists saved credential sets.
func (p *Porter) ListCredentials(opts ListOptions) error {
	creds, err := p.Credentials.ReadAll()
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, creds)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, creds)
	case printer.FormatTable:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		printCredRow :=
			func(v interface{}) []interface{} {
				cr, ok := v.(credentials.CredentialSet)
				if !ok {
					return nil
				}
				return []interface{}{cr.Name, tp.Format(cr.Modified)}
			}
		return printer.PrintTable(p.Out, creds, printCredRow,
			"NAME", "MODIFIED")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

type CredentialOptions struct {
	BundleLifecycleOpts
	DryRun bool
	Silent bool
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (g *CredentialOptions) Validate(args []string, cxt *context.Context) error {
	err := g.validateCredName(args)
	if err != nil {
		return err
	}

	return g.bundleFileOptions.Validate(cxt)
}

func (g *CredentialOptions) validateCredName(args []string) error {
	if len(args) == 1 {
		g.Name = args[0]
	} else if len(args) > 1 {
		return errors.Errorf("only one positional argument may be specified, the credential name, but multiple were received: %s", args)
	}
	return nil
}

// GenerateCredentials builds a new credential set based on the given options. This can be either
// a silent build, based on the opts.Silent flag, or interactive using a survey. Returns an
// error if unable to generate credentials
func (p *Porter) GenerateCredentials(opts CredentialOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before invoking credentials generate")
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
	genOpts := credentialsgenerator.GenerateOptions{
		Name:        name,
		Credentials: bundle.Credentials,
		Silent:      opts.Silent,
	}
	fmt.Fprintf(p.Out, "Generating new credential %s from bundle %s\n", genOpts.Name, bundle.Name)
	fmt.Fprintf(p.Out, "==> %d credentials required for bundle %s\n", len(genOpts.Credentials), bundle.Name)

	cs, err := credentialsgenerator.GenerateCredentials(genOpts)
	if err != nil {
		return errors.Wrap(err, "unable to generate credentials")
	}

	cs.Created = time.Now()
	cs.Modified = cs.Created

	if opts.DryRun {
		data, err := json.Marshal(cs)
		if err != nil {
			return errors.Wrap(err, "unable to generate credentials JSON")
		}
		fmt.Fprintf(p.Out, "%v", string(data))
		return nil
	}
	err = p.Credentials.Save(*cs)
	return errors.Wrapf(err, "unable to save credentials")
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
func (p *Porter) EditCredential(opts CredentialEditOptions) error {
	credSet, err := p.Credentials.Read(opts.Name)
	if err != nil {
		return err
	}

	contents, err := yaml.Marshal(credSet)
	if err != nil {
		return errors.Wrap(err, "unable to load credentials")
	}

	editor := editor.New(p.Context, fmt.Sprintf("porter-%s.yaml", credSet.Name), contents)
	output, err := editor.Run()
	if err != nil {
		return errors.Wrap(err, "unable to open editor to edit credentials")
	}

	err = yaml.Unmarshal(output, &credSet)
	if err != nil {
		return errors.Wrap(err, "unable to load credentials")
	}

	credSet.Modified = time.Now()
	err = p.Credentials.Save(credSet)
	if err != nil {
		return errors.Wrap(err, "unable to save credentials")
	}

	return nil
}

// ShowCredential shows the credential set corresponding to the provided name, using
// the provided printer.PrintOptions for display.
func (p *Porter) ShowCredential(opts CredentialShowOptions) error {
	credSet, err := p.Credentials.Read(opts.Name)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, credSet)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, credSet)
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

		// Iterate through all CredentialStrategies and add to rows
		for _, cs := range credSet.Credentials {
			sourceVal, sourceType := GetCredentialSourceValueAndType(cs.Source)
			rows = append(rows, []string{cs.Name, sourceVal, sourceType})
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
		fmt.Fprintf(p.Out, "Name: %s\n", credSet.Name)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(credSet.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n\n", tp.Format(credSet.Modified))

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

// GetCredentialSourceValueAndType takes a given credentials.Source struct and
// returns the source value itself as well as source type, e.g., 'path', 'env', etc.,
// both in their string forms
func GetCredentialSourceValueAndType(cs credentials.Source) (value string, key string) {
	return cs.Value, cs.Key
}

// CredentialDeleteOptions represent options for Porter's credential delete command
type CredentialDeleteOptions struct {
	Name string
}

// DeleteCredential deletes the credential set corresponding to the provided
// names.
func (p *Porter) DeleteCredential(opts CredentialDeleteOptions) error {
	err := p.Credentials.Delete(opts.Name)
	if err == crud.ErrRecordDoesNotExist {
		if p.Debug {
			fmt.Fprintln(p.Err, "credential set does not exist")
		}
		return nil
	}
	return errors.Wrapf(err, "unable to delete credential")
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
		return errors.Errorf("no credential name was specified")
	case 1:
		return nil
	default:
		return errors.Errorf("only one positional argument may be specified, the credential name, but multiple were received: %s", args)
	}
}
