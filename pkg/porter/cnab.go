package porter

import (
	"fmt"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/build"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/driver/command"
	"github.com/pkg/errors"
)

const (
	// DockerDriver is the name of the Docker driver.
	DockerDriver = cnabprovider.DriverNameDocker

	// DebugDriver is the name of the Debug driver.
	DebugDriver = cnabprovider.DriverNameDebug

	// DefaultDriver is the name of the default driver (Docker).
	DefaultDriver = DockerDriver
)

type bundleFileOptions struct {
	// File path to the porter manifest. Defaults to the bundle in the current directory.
	File string

	// CNABFile is the path to the bundle.json file. Cannot be specified at the same time as the porter manifest or a tag.
	CNABFile string

	// RelocationMapping is the path to the relocation-mapping.json file, if one exists. Populated only for published bundles
	RelocationMapping string

	// TagSet indicates whether a bundle tag is present, to determine whether or not to default bundle files
	TagSet bool
}

func (o *bundleFileOptions) Validate(cxt *context.Context) error {
	err := o.validateBundleFiles(cxt)
	if err != nil {
		return err
	}

	if !o.TagSet {
		err = o.defaultBundleFiles(cxt)
		if err != nil {
			return err
		}
	}

	return err
}

// sharedOptions are common options that apply to multiple CNAB actions.
type sharedOptions struct {
	bundleFileOptions

	// Name of the installation. Defaults to the name of the bundle.
	Name string

	// Params is the unparsed list of NAME=VALUE parameters set on the command line.
	Params []string

	// ParameterSets is a list of parameter sets containing parameter sources
	ParameterSets []string

	// CredentialIdentifiers is a list of credential names or paths to make available to the bundle.
	CredentialIdentifiers []string

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string

	// parsedParams is the parsed set of parameters from Params.
	parsedParams map[string]string

	// parsedParamSets is the parsed set of parameter from ParameterSets
	parsedParamSets map[string]string

	// combinedParameters is parsedParams merged on top of parsedParamSets.
	combinedParameters map[string]string
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (o *sharedOptions) Validate(args []string, p *Porter) error {
	err := o.validateInstallationName(args)
	if err != nil {
		return err
	}

	err = o.bundleFileOptions.Validate(p.Context)
	if err != nil {
		return err
	}

	err = p.applyDefaultOptions(o)
	if err != nil {
		return err
	}

	// Only validate the syntax of the --param flags
	// We will validate the parameter sets later once we have the bundle loaded.
	err = o.parseParams()
	if err != nil {
		return err
	}

	o.defaultDriver()
	err = o.validateDriver()
	if err != nil {
		return err
	}

	return nil
}

// validateInstallationName grabs the installation name from the first positional argument.
func (o *sharedOptions) validateInstallationName(args []string) error {
	if len(args) == 1 {
		o.Name = args[0]
	} else if len(args) > 1 {
		return errors.Errorf("only one positional argument may be specified, the installation name, but multiple were received: %s", args)
	}

	return nil
}

// defaultBundleFiles defaults the porter manifest and the bundle.json files.
func (o *bundleFileOptions) defaultBundleFiles(cxt *context.Context) error {
	if o.File != "" { // --file
		bundleDir := filepath.Dir(o.File)
		o.CNABFile = filepath.Join(bundleDir, build.LOCAL_BUNDLE)
	} else if o.CNABFile != "" { // --cnab-file
		// Nothing to default
	} else { // no flags passed (--tag is handled elsewhere)
		pwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "could not get current working directory")
		}

		manifestExists, err := cxt.FileSystem.Exists(filepath.Join(pwd, config.Name))
		if err != nil {
			return errors.Wrap(err, "could not check if porter manifest exists in current directory")
		}

		if manifestExists {
			o.File = config.Name
			o.CNABFile = build.LOCAL_BUNDLE
		}
	}

	return nil
}

func (o *bundleFileOptions) validateBundleFiles(cxt *context.Context) error {
	if o.File != "" && o.CNABFile != "" {
		return errors.New("cannot specify both --file and --cnab-file")
	}

	err := o.validateFile(cxt)
	if err != nil {
		return err
	}

	err = o.validateCNABFile(cxt)
	if err != nil {
		return err
	}

	return nil
}

func (o *bundleFileOptions) validateFile(cxt *context.Context) error {
	if o.File == "" {
		return nil
	}

	// Verify the file can be accessed
	if _, err := cxt.FileSystem.Stat(o.File); err != nil {
		return errors.Wrapf(err, "unable to access --file %s", o.File)
	}

	return nil
}

// validateCNABFile converts the bundle file path to an absolute filepath and verifies that it exists.
func (o *bundleFileOptions) validateCNABFile(cxt *context.Context) error {
	if o.CNABFile == "" {
		return nil
	}

	originalPath := o.CNABFile
	if !filepath.IsAbs(o.CNABFile) {
		// Convert to an absolute filepath because runtime needs it that way
		pwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "could not get current working directory")
		}

		f := filepath.Join(pwd, o.CNABFile)
		f, err = filepath.Abs(f)
		if err != nil {
			return errors.Wrapf(err, "could not get absolute path for --cnab-file %s", o.CNABFile)
		}

		o.CNABFile = f
	}

	// Verify the file can be accessed
	if _, err := cxt.FileSystem.Stat(o.CNABFile); err != nil {
		// warn about the original relative path
		return errors.Wrapf(err, "unable to access --cnab-file %s", originalPath)
	}

	return nil
}

// LoadParameters validates and resolves the parameters and sets. It must be
// called after porter has loaded the bundle definition.
func (o *sharedOptions) LoadParameters(p *Porter) error {
	err := o.parseParams()
	if err != nil {
		return err
	}

	err = o.parseParamSets(p)
	if err != nil {
		return err
	}

	o.combinedParameters = o.combineParameters(p.Context)

	return nil
}

// parsedParams parses the variable assignments in Params.
func (o *sharedOptions) parseParams() error {
	p, err := parameters.ParseVariableAssignments(o.Params)
	if err != nil {
		return err
	}
	o.parsedParams = p
	return nil
}

// parseParamSets parses the variable assignments in ParameterSets.
func (o *sharedOptions) parseParamSets(p *Porter) error {
	if len(o.ParameterSets) > 0 {
		parsed, err := p.loadParameterSets(o.ParameterSets)
		if err != nil {
			return errors.Wrapf(err, "unable to process provided parameter sets: %v", o.ParameterSets)
		}
		o.parsedParamSets = parsed
	}
	return nil
}

// Combine the parameters into a single map
// The params set on the command line take precedence over the params set in
// parameter set files
// Anything set multiple times, is decided by "last one set wins"
func (o *sharedOptions) combineParameters(c *context.Context) map[string]string {
	final := make(map[string]string)

	for k, v := range o.parsedParamSets {
		final[k] = v
	}

	for k, v := range o.parsedParams {
		final[k] = v
	}

	//
	// Default the porter-debug param to --debug
	//
	if c.Debug {
		final["porter-debug"] = "true"
	}

	return final
}

// defaultDriver supplies the default driver if none is specified
func (o *sharedOptions) defaultDriver() {
	if o.Driver == "" {
		o.Driver = DefaultDriver
	}
}

// validateDriver validates that the provided driver is supported by Porter
func (o *sharedOptions) validateDriver() error {
	switch o.Driver {
	case DockerDriver, DebugDriver:
		return nil
	default:
		cmddriver := &command.Driver{Name: o.Driver}
		if cmddriver.CheckDriverExists() {
			return nil
		}

		return fmt.Errorf("unsupported driver or driver not found in PATH: %s", o.Driver)
	}
}
