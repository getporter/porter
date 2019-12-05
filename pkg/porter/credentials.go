package porter

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/credentialsgenerator"
	"get.porter.sh/porter/pkg/printer"

	dtprinter "github.com/carolynvs/datetime-printer"
	credentials "github.com/deislabs/cnab-go/credentials"
	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// CredentialShowOptions represent options for Porter's credential show command
type CredentialShowOptions struct {
	printer.PrintOptions
	Name string
}

// CredentialsFile represents a CNAB credentials file and corresponding metadata
type CredentialsFile struct {
	Name     string
	Modified time.Time
}

// CredentialsFileList is a slice of CredentialsFiles
type CredentialsFileList []CredentialsFile

func (l CredentialsFileList) Len() int {
	return len(l)
}
func (l CredentialsFileList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l CredentialsFileList) Less(i, j int) bool {
	return l[i].Modified.Before(l[j].Modified)
}

// fetchCredential returns a *credentials.CredentialsSet according to the supplied
// credential name, or an error if encountered
func (p *Porter) fetchCredential(name string) (*credentials.CredentialSet, error) {
	credsDir, err := p.Config.GetCredentialsDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine credentials directory")
	}

	path := filepath.Join(credsDir, fmt.Sprintf("%s.yaml", name))
	return p.readCredential(name, path)
}

// fetchCredentials fetches all credentials in the form of a CredentialsFileList
// from the designated credentials dir, or an error if encountered
func (p *Porter) fetchCredentials() (*CredentialsFileList, error) {
	credsDir, err := p.Config.GetCredentialsDir()
	if err != nil {
		return &CredentialsFileList{}, errors.Wrap(err, "unable to determine credentials directory")
	}

	credentialsFiles := CredentialsFileList{}
	if ok, _ := p.Context.FileSystem.DirExists(credsDir); ok {
		p.Context.FileSystem.Walk(credsDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				credName := strings.Split(info.Name(), ".")[0]
				credSet, err := p.readCredential(credName, path)
				if err != nil {
					// If an error is encountered while reading, log and move on to the next
					if p.Debug {
						fmt.Fprint(p.Err, err.Error())
					}
					return nil
				}
				credentialsFiles = append(credentialsFiles,
					CredentialsFile{Name: credSet.Name, Modified: info.ModTime()})
			}
			return nil
		})
		sort.Sort(sort.Reverse(credentialsFiles))
	}
	return &credentialsFiles, nil
}

// readCredential reads a credential with the given name via the provided path
// and returns a CredentialSet or an error, if encountered
func (p *Porter) readCredential(name, path string) (*credentials.CredentialSet, error) {
	credSet := &credentials.CredentialSet{}

	data, err := p.Context.FileSystem.ReadFile(path)
	if err != nil {
		return credSet, errors.Wrapf(err, "unable to load credential %s", name)
	}

	if err = yaml.Unmarshal(data, credSet); err != nil {
		return credSet, errors.Wrapf(err, "unable to unmarshal credential %s", name)
	}

	return credSet, nil
}

// ListCredentials lists saved credential sets.
func (p *Porter) ListCredentials(opts ListOptions) error {
	credentialsFiles, err := p.fetchCredentials()
	if err != nil {
		return errors.Wrap(err, "unable to fetch credentials")
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, *credentialsFiles)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, *credentialsFiles)
	case printer.FormatTable:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		printCredRow :=
			func(v interface{}) []interface{} {
				cr, ok := v.(CredentialsFile)
				if !ok {
					return nil
				}
				return []interface{}{cr.Name, tp.Format(cr.Modified)}
			}
		return printer.PrintTable(p.Out, *credentialsFiles, printCredRow,
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
	bundle, err := p.CNAB.LoadBundle(opts.CNABFile, opts.Insecure)

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

	//write the credential out to PORTER_HOME with Porter's Context
	data, err := yaml.Marshal(cs)
	if err != nil {
		return errors.Wrap(err, "unable to generate credentials YAML")
	}
	if opts.DryRun {
		fmt.Fprintf(p.Out, "%v", string(data))
		return nil
	}
	credentialsDir, err := p.Config.GetCredentialsDir()
	if err != nil {
		return errors.Wrap(err, "unable to get credentials directory")
	}
	// Make the credentials path if it doesn't exist. MkdirAll does nothing if it already exists
	// Readable, writable only by the user
	err = p.Config.FileSystem.MkdirAll(credentialsDir, 0700)
	if err != nil {
		return errors.Wrap(err, "unable to create credentials directory")
	}
	dest, err := p.Config.GetCredentialPath(genOpts.Name)
	if err != nil {
		return errors.Wrap(err, "unable to determine credentials path")
	}

	fmt.Fprintf(p.Out, "Saving credential to %s\n", dest)

	err = p.Context.FileSystem.WriteFile(dest, data, 0600)
	if err != nil {
		return errors.Wrapf(err, "couldn't write credential file %s", dest)
	}
	return nil
}

// Validate validates the args provided Porter's credential show command
func (o *CredentialShowOptions) Validate(args []string) error {
	switch len(args) {
	case 0:
		return errors.Errorf("no credential name was specified")
	case 1:
		o.Name = strings.ToLower(args[0])
	default:
		return errors.Errorf("only one positional argument may be specified, the credential name, but multiple were received: %s", args)
	}

	return o.ParseFormat()
}

// ShowCredential shows the credential set corresponding to the provided name, using
// the provided printer.PrintOptions for display.
func (p *Porter) ShowCredential(opts CredentialShowOptions) error {
	credSet, err := p.fetchCredential(opts.Name)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, credSet)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, credSet)
	case printer.FormatTable:
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

		// First, print the CredentialSet name
		fmt.Fprintf(p.Out, "Name: %s\n\n", credSet.Name)

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

type reflectedStruct struct {
	Value reflect.Value
	Type  reflect.Type
}

// GetCredentialSourceValueAndType takes a given credentials.Source struct and
// returns the source value itself as well as source type, e.g., 'Path', 'EnvVar', etc.,
// both in their string forms
func GetCredentialSourceValueAndType(cs credentials.Source) (string, string) {
	var sourceVal, sourceType string

	// Build a reflected credentials.Source struct
	reflectedSource := reflectedStruct{
		Value: reflect.ValueOf(cs),
		Type:  reflect.TypeOf(cs),
	}

	// Iterate through all of the fields of a credentials.Source struct
	for i := 0; i < reflectedSource.Type.NumField(); i++ {
		// A Field name would be 'Path', 'EnvVar', etc.
		fieldName := reflectedSource.Type.Field(i).Name
		// Get the value for said Field
		fieldValue := reflect.Indirect(reflectedSource.Value).FieldByName(fieldName).String()
		// If value non-empty, this field value and name represent our source value
		// and source type, respectively
		if fieldValue != "" {
			sourceVal, sourceType = fieldValue, fieldName
		}
	}
	return sourceVal, sourceType
}
