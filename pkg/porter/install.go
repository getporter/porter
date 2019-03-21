package porter

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/parameters"
	"github.com/pkg/errors"
)

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	// Name of the claim. Defaults to the name of the bundle.
	Name string

	// File path to the CNAB bundle. Defaults to the bundle in the current directory.
	File string

	// Insecure bundle installation allowed.
	Insecure bool

	// Params is the unparsed list of NAME=VALUE parameters set on the command line.
	Params []string

	// ParamFiles is a list of file paths containing lines of NAME=VALUE parameter definitions.
	ParamFiles []string

	// CredentialIdentifiers is a list of credential names or paths to make available to the bundle.
	CredentialIdentifiers []string

	// parsedParams is the parsed set of parameters from Params.
	parsedParams map[string]string

	// parsedParamFiles is the parsed set of parameters from Params.
	parsedParamFiles []map[string]string
}

// Validate prepares for an installation and validates the installation options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (o *InstallOptions) Validate() error {
	err := o.validateBundlePath()
	if err != nil {
		return err
	}

	err = o.validateParams()
	if err != nil {
		return err
	}

	return nil
}

// validateBundlePath gets the absolute path to the bundle file.
func (o *InstallOptions) validateBundlePath() error {
	err := o.defaultBundleFile()
	if err != nil {
		return err
	}

	err = o.prepareBundleFile()
	if err != nil {
		return err
	}

	// Verify the file can be accessed
	if _, err := os.Stat(o.File); err != nil {
		return errors.Wrapf(err, "unable to access bundle file %s", o.File)
	}

	return nil
}

// defaultBundleFile defaults the bundle file to the bundle in the current directory
// when none is set.
func (o *InstallOptions) defaultBundleFile() error {
	if o.File != "" {
		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "could not get current working directory")
	}

	files, err := ioutil.ReadDir(pwd)
	if err != nil {
		return errors.Wrapf(err, "could not list current directory %s", pwd)
	}

	// We are looking both for a bundle.json OR a porter manifest
	// If we can't find a bundle.json, but we found manifest, tell them to run porter build first
	foundManifest := false
	for _, f := range files {
		// TODO: handle defaulting to a signed bundle
		if !f.IsDir() && f.Name() == "bundle.json" {
			o.File = "bundle.json"
			break
		}

		if !f.IsDir() && f.Name() == config.Name {
			foundManifest = true
		}
	}

	if o.File == "" && foundManifest {
		return errors.New("first run 'porter build' to generate a bundle.json, then run 'porter install'")
	}

	return nil
}

// prepareBundleFile converts the bundle file path to an absolute filepath.
func (o *InstallOptions) prepareBundleFile() error {
	if filepath.IsAbs(o.File) {
		return nil
	}

	// Convert to an absolute filepath
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "could not get current working directory")
	}

	f := filepath.Join(pwd, o.File)
	f, err = filepath.Abs(f)
	if err != nil {
		return errors.Wrapf(err, "could not get absolute path for bundle file %s", f)
	}

	o.File = f
	return nil
}

func (o *InstallOptions) validateParams() error {
	err := o.parseParams()
	if err != nil {
		return err
	}

	err = o.parseParamFiles()
	if err != nil {
		return err
	}

	return nil
}

// parsedParams parses the variable assignments in Params.
func (o *InstallOptions) parseParams() error {
	p, err := parameters.ParseVariableAssignments(o.Params)
	if err != nil {
		return err
	}
	o.parsedParams = p
	return nil
}

// parseParamFiles parses the variable assignments in ParamFiles.
func (o *InstallOptions) parseParamFiles() error {
	o.parsedParamFiles = make([]map[string]string, 0, len(o.ParamFiles))

	for _, path := range o.ParamFiles {
		err := o.parseParamFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *InstallOptions) parseParamFile(path string) error {
	f, err := os.Open(path)
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
func (o *InstallOptions) combineParameters() map[string]string {
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

// ToDuffleArgs converts this instance of user-provided installation options
// to duffle installation arguments.
func (o *InstallOptions) ToDuffleArgs() cnabprovider.InstallArguments {
	return cnabprovider.InstallArguments{
		Claim:                 o.Name,
		BundleIdentifier:      o.File,
		BundleIsFile:          true,
		Insecure:              o.Insecure,
		Params:                o.combineParameters(),
		CredentialIdentifiers: o.CredentialIdentifiers,
	}
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(opts InstallOptions) error {
	err := p.Config.LoadManifest()
	if err != nil {
		return err
	}

	if opts.Name == "" {
		opts.Name = p.Manifest.Name
	}

	fmt.Fprintf(p.Out, "installing %s...\n", p.Manifest.Name)
	return p.Install(opts.ToDuffleArgs())
}
