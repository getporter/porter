package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/deislabs/duffle/pkg/action"
	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/claim"
	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/repo"
)

const installUsage = `Installs a CNAB bundle.

This command installs a CNAB bundle with a specific installation name.
A claim (a record about the application installed) is created during
the install process and can be referenced by the installation name.
Example:
	$ duffle install my_release example:0.1.0
	$ duffle status my_release

Different drivers are available for executing the duffle invocation
image. The following drivers are built-in:

	- docker: run the Docker client. Works for OCI and Docker images
	- debug: fake a run of the invocation image, and print out what
        would have been sent

Some drivers have additional configuration that can be passed via
environment variable. When using the Docker driver, the VERBOSE
environment variable can be set to "true" to turn on extra output.

UNIX Example:
	$ VERBOSE=true duffle install -d docker my_release example:0.1.0

Windows Example:
	$ $env:VERBOSE = true
	$ duffle install -d docker my_release example:0.1.0

You can also load the bundle.json file directly:

	$ duffle install dev_bundle path/to/bundle.json --bundle-is-file

Verifying and --insecure:
When the --insecure flag is passed, verification steps will not be
performed. This means that Duffle will accept both unsigned
(bundle.json) and signed (bundle.cnab) files, but will not perform
any validation. The following table illustrates this:

	Bundle     Key known?    Flag            Result
	------     ----------    -----------     ------
	Signed     Known         None            Okay
	Signed     Known         --insecure      Okay
	Signed     Unknown       None            Verification error
	Signed     Unknown       --insecure      Okay
	Unsigned   N/A           None            Verification error
	Unsigned   N/A           --insecure      Okay
`

type installCmd struct {
	bundle string
	home   home.Home
	out    io.Writer

	driver           string
	credentialsFiles []string
	valuesFile       string
	setParams        []string
	insecure         bool
	setFiles         []string
	bundleIsFile     bool
	name             string
}

func newInstallCmd(w io.Writer) *cobra.Command {
	install := &installCmd{out: w}

	cmd := &cobra.Command{
		Use:   "install [NAME] [BUNDLE]",
		Short: "install a CNAB bundle",
		Long:  installUsage,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			install.bundle = args[1]
			install.name = args[0]
			install.home = home.Home(homePath())
			return install.run()
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&install.bundleIsFile, "bundle-is-file", "f", false, "Indicates that the bundle source is a file path")
	f.BoolVarP(&install.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")
	f.StringVarP(&install.driver, "driver", "d", "docker", "Specify a driver name")
	f.StringVarP(&install.valuesFile, "parameters", "p", "", "Specify file containing parameters. Formats: toml, MORE SOON")
	f.StringArrayVarP(&install.credentialsFiles, "credentials", "c", []string{}, "Specify credentials to use inside the CNAB bundle. This can be a credentialset name or a path to a file.")
	f.StringArrayVarP(&install.setParams, "set", "s", []string{}, "Set individual parameters as NAME=VALUE pairs")
	f.StringArrayVarP(&install.setFiles, "set-file", "i", []string{}, "Set individual parameters from file content as NAME=SOURCE-PATH pairs")

	return cmd
}

func (i *installCmd) run() error {
	bundleFile, err := resolveBundleFilePath(i.bundle, i.home.String(), i.bundleIsFile, i.insecure)
	if err != nil {
		return err
	}
	// look in claims store for another claim with the same name
	_, err = claimStorage().Read(i.name)
	if err != claim.ErrClaimNotFound {
		return fmt.Errorf("a claim with the name %v already exists", i.name)
	}

	bun, err := loadBundle(bundleFile, i.insecure)
	if err != nil {
		return err
	}

	if err = bun.Validate(); err != nil {
		return err
	}

	driverImpl, err := prepareDriver(i.driver)
	if err != nil {
		return err
	}

	creds, err := loadCredentials(i.credentialsFiles, bun)
	if err != nil {
		return err
	}

	// Because this is an install, we create a new claim. For upgrades, we'd
	// load the claim based on install name
	c, err := claim.New(i.name)
	if err != nil {
		return err
	}

	c.Bundle = bun
	c.Parameters, err = calculateParamValues(bun, i.valuesFile, i.setParams, i.setFiles)
	if err != nil {
		return err
	}

	inst := &action.Install{
		Driver: driverImpl,
	}
	fmt.Fprintf(i.out, "Executing install action...\n")
	err = inst.Run(c, creds, i.out)

	// Even if the action fails, we want to store a claim. This is because
	// we cannot know, based on a failure, whether or not any resources were
	// created. So we want to suggest that the user take investigative action.
	err2 := claimStorage().Store(*c)
	if err != nil {
		return fmt.Errorf("Install step failed: %v", err)
	}
	return err2
}

func getBundleFilepath(bun, homePath string, insecure bool) (string, error) {
	home := home.Home(homePath)
	ref, err := getReference(bun)
	if err != nil {
		return "", fmt.Errorf("could not parse reference for %s: %v", bun, err)
	}

	// read the bundle reference from repositories.json
	index, err := repo.LoadIndex(home.Repositories())
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %v", home.Repositories(), err)
	}

	tag := ref.Tag()
	if ref.Tag() == "latest" {
		tag = ""
	}

	digest, err := index.Get(ref.Name(), tag)
	if err != nil {
		return "", fmt.Errorf("could not find %s:%s in %s: %v", ref.Name(), ref.Tag(), home.Repositories(), err)
	}
	return filepath.Join(home.Bundles(), digest), nil
}

// overrides parses the --set data and returns values that should override other params.
func overrides(overrides []string, paramDefs map[string]bundle.ParameterDefinition) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for _, p := range overrides {
		pair := strings.SplitN(p, "=", 2)
		if len(pair) != 2 {
			// For now, I guess we skip cases where someone does --set foo or --set foo=
			// We could set this to an explicit nil and then use it as a trigger to unset
			// a parameter in the file.
			continue
		}
		def, ok := paramDefs[pair[0]]
		if !ok {
			return res, fmt.Errorf("parameter %s not defined in bundle", pair[0])
		}

		if _, ok := res[pair[0]]; ok {
			return res, fmt.Errorf("parameter %q specified multiple times", pair[0])
		}

		var err error
		res[pair[0]], err = def.ConvertValue(pair[1])
		if err != nil {
			return res, fmt.Errorf("cannot use %s as value of %s: %s", pair[1], pair[0], err)
		}
	}
	return res, nil
}

func parseValues(file string) (map[string]interface{}, error) {
	v := viper.New()
	v.SetConfigFile(file)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return v.AllSettings(), nil
}

func calculateParamValues(bun *bundle.Bundle, valuesFile string, setParams, setFilePaths []string) (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	if valuesFile != "" {
		var err error
		vals, err = parseValues(valuesFile)
		if err != nil {
			return vals, err
		}

	}
	overridden, err := overrides(setParams, bun.Parameters)
	if err != nil {
		return vals, err
	}
	for k, v := range overridden {
		vals[k] = v
	}

	// Now add files.
	for _, p := range setFilePaths {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return vals, fmt.Errorf("malformed set-file parameter: %q (must be NAME=PATH)", p)
		}

		// Check that this is a known param
		if _, ok := bun.Parameters[parts[0]]; !ok {
			return vals, fmt.Errorf("bundle does not have a parameter named %q", parts[0])
		}

		if _, ok := overridden[parts[0]]; ok {
			return vals, fmt.Errorf("parameter %q specified multiple times", parts[0])
		}
		content, err := ioutil.ReadFile(parts[1])
		if err != nil {
			return vals, fmt.Errorf("could not read file %q: %s", parts[1], err)
		}
		vals[parts[0]] = string(content)
	}

	return bundle.ValuesOrDefaults(vals, bun)
}
