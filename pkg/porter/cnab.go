package porter

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg/build"
	cnabprovider "get.porter.sh/porter/pkg/cnab/provider"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
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

	// ReferenceSet indicates whether a bundle reference is present, to determine whether or not to default bundle files
	ReferenceSet bool

	// Dir represents the build context directory containing bundle assets
	Dir string
}

func (o *bundleFileOptions) Validate(cxt *portercontext.Context) error {
	var err error

	if o.ReferenceSet {
		return nil
	}

	// Resolve the proper build context directory
	if o.Dir != "" {
		_, err = cxt.FileSystem.IsDir(o.Dir)
		if err != nil {
			return fmt.Errorf("%q is not a valid directory: %w", o.Dir, err)
		}
		o.Dir = cxt.FileSystem.Abs(o.Dir)
	} else {
		// default to current working directory
		o.Dir = cxt.Getwd()
	}

	if o.File != "" {
		if !filepath.IsAbs(o.File) {
			o.File = cxt.FileSystem.Abs(filepath.Join(o.Dir, o.File))
		} else {
			o.File = cxt.FileSystem.Abs(o.File)
		}
	}

	err = o.validateBundleFiles(cxt)
	if err != nil {
		return err
	}

	err = o.defaultBundleFiles(cxt)
	if err != nil {
		return err
	}

	// Enter the resolved build context directory after all defaults
	// have been populated
	cxt.Chdir(o.Dir)
	return nil
}

// installationOptions are common options that apply to commands that use an installation
type installationOptions struct {
	bundleFileOptions

	// Namespace of the installation.
	Namespace string

	// Name of the installation. Defaults to the name of the bundle.
	Name string
}

// Validate prepares for an action and validates the options.
// For example, relative paths are converted to full paths and then checked that
// they exist and are accessible.
func (o *installationOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	err := o.validateInstallationName(args)
	if err != nil {
		return err
	}

	err = o.bundleFileOptions.Validate(p.Context)
	if err != nil {
		return err
	}

	err = p.applyDefaultOptions(ctx, o)
	if err != nil {
		return err
	}

	return nil
}

// validateInstallationName grabs the installation name from the first positional argument.
func (o *installationOptions) validateInstallationName(args []string) error {
	if len(args) == 1 {
		o.Name = args[0]
	} else if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, the installation name, but multiple were received: %s", args)
	}

	return nil
}

// defaultBundleFiles defaults the porter manifest and the bundle.json files.
func (o *bundleFileOptions) defaultBundleFiles(cxt *portercontext.Context) error {
	if o.File != "" { // --file
		o.defaultCNABFile()
	} else if o.CNABFile != "" { // --cnab-file
		// Nothing to default
	} else {
		defaultPath := filepath.Join(o.Dir, config.Name)
		manifestExists, err := cxt.FileSystem.Exists(defaultPath)
		if err != nil {
			return fmt.Errorf("could not find a porter manifest at %s: %w", defaultPath, err)
		} else if !manifestExists {
			return nil
		}

		o.File = defaultPath
		o.defaultCNABFile()
	}

	return nil
}

func (o *bundleFileOptions) defaultCNABFile() {
	o.CNABFile = filepath.Join(o.Dir, build.LOCAL_BUNDLE)
}

func (o *bundleFileOptions) validateBundleFiles(cxt *portercontext.Context) error {
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

func (o *bundleFileOptions) validateFile(cxt *portercontext.Context) error {
	if o.File == "" {
		return nil
	}

	// Verify the file can be accessed
	if _, err := cxt.FileSystem.Stat(o.File); err != nil {
		return fmt.Errorf("unable to access --file %s: %w", o.File, err)
	}

	return nil
}

// validateCNABFile converts the bundle file path to an absolute filepath and verifies that it exists.
func (o *bundleFileOptions) validateCNABFile(cxt *portercontext.Context) error {
	if o.CNABFile == "" {
		return nil
	}

	originalPath := o.CNABFile
	if !filepath.IsAbs(o.CNABFile) {
		// Convert to an absolute filepath because runtime needs it that way
		o.CNABFile = filepath.Join(cxt.Getwd(), o.CNABFile)
	}

	// Verify the file can be accessed
	if _, err := cxt.FileSystem.Stat(o.CNABFile); err != nil {
		// warn about the original relative path
		return fmt.Errorf("unable to access --cnab-file %s: %w", originalPath, err)
	}

	return nil
}
