package porter

import (
	"encoding/json"
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/credentialsgenerator"
	"get.porter.sh/porter/pkg/printer"
	"gopkg.in/AlecAivazis/survey.v1"

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

	_, err = p.generateCredentials(opts.CNABFile, opts.Name, opts.Silent, opts.DryRun)
	return err
}

func (p *Porter) generateCredentials(CNABFile string, credIdentifierName string, silent bool, dryRun bool) (string, error) {
	var credSetNames []string
	credSets, err := p.Credentials.ReadAll()
	if err != nil {
		return "", errors.Wrap(err, "failed to read exisiting credential sets")
	}

	shouldGenerateCred := false
	var selectedOption string
	if len(credSets) > 0 {
		for _, credSet := range credSets {
			credSetNames = append(credSetNames, credSet.Name)
		}

		selectCredPrompt := &survey.Select{
			Message: "Choose a set of credentials to use while installing this bundle",
			Options: append(credSetNames, generateCredCode, quitGenerateCode),
			Default: credSetNames[0],
		}
		survey.AskOne(selectCredPrompt, &selectedOption, nil)

		switch selectedOption {
		case generateCredCode:
			shouldGenerateCred = true
		case quitGenerateCode:
			return "", errors.New("Credentials are mandatory to install this bundle")
		default:
			return selectedOption, nil
		}
	} else {
		shouldGenerateCredPrompt := &survey.Confirm{
			Message: "No credential identifier given. Generate one ?",
		}
		survey.AskOne(shouldGenerateCredPrompt, &shouldGenerateCred, nil)
	}

	if !shouldGenerateCred {
		return "", errors.New("Credentials are mandatory to install this bundle")
	}

	bundle, err := p.CNAB.LoadBundle(CNABFile)
	if err != nil {
		return "", errors.Wrap(err, "failed to load bundle while generating credentials")
	}

	if credIdentifierName == "" {
		inputCredNamePrompt := &survey.Input{
			Message: "Enter credential identifier name",
			Default: bundle.Name,
		}
		survey.AskOne(inputCredNamePrompt, &credIdentifierName, nil)
	}

	genOpts := credentialsgenerator.GenerateOptions{
		Name:        credIdentifierName,
		Credentials: bundle.Credentials,
		Silent:      silent,
		DryRun:      dryRun,
	}

	fmt.Fprintf(p.Out, "  Generating new credential %s from bundle %s\n", genOpts.Name, bundle.Name)
	fmt.Fprintf(p.Out, "  %d credentials required for bundle %s\n", len(genOpts.Credentials), bundle.Name)
	err = p.generateAndSaveCredentialSetForCNABFile(CNABFile, genOpts)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate credentials")
	}
	fmt.Fprintln(p.Out, "Credentials generated and saved successfully for future use")

	return credIdentifierName, nil

}

func (p *Porter) generateAndSaveCredentialSetForCNABFile(CNABFile string, genOpts credentialsgenerator.GenerateOptions) error {

	cs, err := credentialsgenerator.GenerateCredentials(genOpts)
	if err != nil {
		return errors.Wrap(err, "unable to generate credentials")
	}

	cs.Created = time.Now()
	cs.Modified = cs.Created

	if genOpts.DryRun {
		data, err := json.Marshal(cs)
		if err != nil {
			return errors.Wrap(err, "unable to generate credentials JSON")
		}
		fmt.Fprintf(p.Out, "%v", string(data))
		return nil
	}
	err = p.Credentials.Save(*cs)
	if err != nil {
		return errors.Wrapf(err, "unable to save credentials")
	}

	return nil
}

// Validate validates the args provided Porter's credential show command
func (o *CredentialShowOptions) Validate(args []string) error {
	if err := validateCredentialName(args); err != nil {
		return err
	}
	o.Name = args[0]
	return o.ParseFormat()
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
