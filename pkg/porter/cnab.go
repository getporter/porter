package porter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/build"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/parameters"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/driver/command"
	"github.com/pkg/errors"
)

// CNABProvider
type CNABProvider interface {
	LoadBundle(bundleFile string, insecure bool) (*bundle.Bundle, error)
	Install(arguments cnabprovider.ActionArguments) error
	Upgrade(arguments cnabprovider.ActionArguments) error
	Invoke(action string, arguments cnabprovider.ActionArguments) error
	Uninstall(arguments cnabprovider.ActionArguments) error
}

const DefaultDriver = "docker"

type bundleFileOptions struct {
	// File path to the porter manifest. Defaults to the bundle in the current directory.
	File string

	// CNABFile is the path to the bundle.json file. Cannot be specified at the same time as the porter manifest or a tag.
	CNABFile string

	// RelocationMapping is the path to the relocation-mapping.json file, if one exists. Populated only for published bundles
	RelocationMapping string
}

func (o *bundleFileOptions) Validate(cxt *context.Context) error {
	err := o.validateBundleFiles(cxt)
	if err != nil {
		return err
	}

	err = o.defaultBundleFiles(cxt)
	if err != nil {
		return err
	}

	return err
}

// sharedOptions are common options that apply to multiple CNAB actions.
type sharedOptions struct {
	bundleFileOptions

	// Name of the instance. Defaults to the name of the bundle.
	Name string

	// Insecure bundles allowed.
	Insecure bool

	// Params is the unparsed list of NAME=VALUE parameters set on the command line.
	Params []string

	// ParamFiles is a list of file paths containing lines of NAME=VALUE parameter definitions.
	ParamFiles []string

	// CredentialIdentifiers is a list of credential names or paths to make available to the bundle.
	CredentialIdentifiers []string

	// Driver is the CNAB-compliant driver used to run bundle actions.
	Driver string

	// parsedParams is the parsed set of parameters from Params.
	parsedParams map[string]string

	// parsedParamFiles is the parsed set of parameters from Params.
	parsedParamFiles []map[string]string

	// combinedParameters is parsedParams merged on top of parsedParamFiles.
	combinedParameters map[string]string
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (o *sharedOptions) Validate(args []string, cxt *context.Context) error {
	o.Insecure = true

	err := o.validateInstanceName(args)
	if err != nil {
		return err
	}

	err = o.bundleFileOptions.Validate(cxt)
	if err != nil {
		return err
	}

	err = o.validateParams(cxt)
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

// validateInstanceName grabs the claim name from the first positional argument.
func (o *sharedOptions) validateInstanceName(args []string) error {
	if len(args) == 1 {
		o.Name = args[0]
	} else if len(args) > 1 {
		return errors.Errorf("only one positional argument may be specified, the bundle instance name, but multiple were received: %s", args)
	}

	return nil
}

// defaultBundleFiles defaults the porter manifest and the bundle.json files.
func (o *bundleFileOptions) defaultBundleFiles(cxt *context.Context) error {
	if o.File == "" {
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
	} else {
		bundleDir := filepath.Dir(o.File)
		o.CNABFile = filepath.Join(bundleDir, build.LOCAL_BUNDLE)
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

func (o *sharedOptions) validateParams(cxt *context.Context) error {
	err := o.parseParams()
	if err != nil {
		return err
	}

	err = o.parseParamFiles(cxt)
	if err != nil {
		return err
	}

	o.combinedParameters = o.combineParameters()

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

// parseParamFiles parses the variable assignments in ParamFiles.
func (o *sharedOptions) parseParamFiles(cxt *context.Context) error {
	o.parsedParamFiles = make([]map[string]string, 0, len(o.ParamFiles))

	for _, path := range o.ParamFiles {
		err := o.parseParamFile(path, cxt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *sharedOptions) parseParamFile(path string, cxt *context.Context) error {
	f, err := cxt.FileSystem.Open(path)
	if err != nil {
		return errors.Wrapf(err, "could not read param file %s", path)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrapf(err, "unable to read contents of param file %s", path)
	}

	p, err := parameters.ParseVariableAssignments(lines)
	if err != nil {
		return err
	}

	o.parsedParamFiles = append(o.parsedParamFiles, p)
	return nil
}

// Combine the parameters into a single map
// The params set on the command line take precedence over the params set in files
// Anything set multiple times, is decided by "last one set wins"
func (o *sharedOptions) combineParameters() map[string]string {
	final := make(map[string]string)

	for _, pf := range o.parsedParamFiles {
		for k, v := range pf {
			final[k] = v
		}
	}

	for k, v := range o.parsedParams {
		final[k] = v
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
	case "docker", "debug":
		return nil
	default:
		cmddriver := &command.Driver{Name: o.Driver}
		if cmddriver.CheckDriverExists() {
			return nil
		}

		return fmt.Errorf("unsupported driver or driver not found in PATH: %s", o.Driver)
	}
}
