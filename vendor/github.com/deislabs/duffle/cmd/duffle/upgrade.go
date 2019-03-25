package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/deislabs/duffle/pkg/action"
)

const upgradeUsage = `perform the upgrade action in the CNAB bundle`
const upgradeLong = `Upgrades an existing application.

An upgrade can do the following:

	- Upgrade a current release to a newer bundle (optionally with parameters)
	- Upgrade a current release using the same bundle but different parameters

Credentials must be supplied when applicable, though they need not be the same credentials that were used
to do the install.

If no parameters are passed, the parameters from the previous release will be used. If '--set' or '--parameters'
are specified, the parameters there will be used (even if the resolved set is empty).
`

var ErrBundleAndBundleFile = errors.New("Both --bundle and --bundle-file flags cannot be set")

type upgradeCmd struct {
	out              io.Writer
	name             string
	driver           string
	valuesFile       string
	bundle           string
	bundleFile       string
	setParams        []string
	insecure         bool
	setFiles         []string
	credentialsFiles []string
}

func newUpgradeCmd(w io.Writer) *cobra.Command {
	upgrade := &upgradeCmd{out: w}

	cmd := &cobra.Command{
		Use:   "upgrade [NAME]",
		Short: upgradeUsage,
		Long:  upgradeLong,
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return upgrade.setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			upgrade.name = args[0]

			return upgrade.run()
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&upgrade.driver, "driver", "d", "docker", "Specify a driver name")
	flags.StringVarP(&upgrade.bundle, "bundle", "b", "", "bundle to use for upgrading")
	flags.StringVar(&upgrade.bundleFile, "bundle-file", "", "path of the bundle file to use for upgrading")
	flags.StringArrayVarP(&upgrade.credentialsFiles, "credentials", "c", []string{}, "Specify credentials to use inside the CNAB bundle. This can be a credentialset name or a path to a file.")
	flags.StringVarP(&upgrade.valuesFile, "parameters", "p", "", "Specify file containing parameters. Formats: toml, MORE SOON")
	flags.StringArrayVarP(&upgrade.setParams, "set", "s", []string{}, "Set individual parameters as NAME=VALUE pairs")
	flags.BoolVarP(&upgrade.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")
	flags.StringArrayVarP(&upgrade.setFiles, "set-file", "i", []string{}, "Set individual parameters from file content as NAME=SOURCE-PATH pairs")

	return cmd
}

func (up *upgradeCmd) setup() error {
	bundleFile, err := prepareBundleFile(up.bundle, up.bundleFile, up.insecure)
	if err != nil {
		return err
	}

	up.bundleFile = bundleFile
	return nil
}

func (up *upgradeCmd) run() error {

	claim, err := claimStorage().Read(up.name)
	if err != nil {
		return fmt.Errorf("%v not found: %v", up.name, err)
	}

	// If the user specifies a bundle file, override the existing one.
	if up.bundleFile != "" {
		bun, err := loadBundle(up.bundleFile, up.insecure)
		if err != nil {
			return err
		}
		claim.Bundle = bun
	}

	driverImpl, err := prepareDriver(up.driver)
	if err != nil {
		return err
	}

	creds, err := loadCredentials(up.credentialsFiles, claim.Bundle)
	if err != nil {
		return err
	}

	// Override parameters only if some are set.
	if up.valuesFile != "" || len(up.setParams) > 0 {
		claim.Parameters, err = calculateParamValues(claim.Bundle, up.valuesFile, up.setParams, up.setFiles)
		if err != nil {
			return err
		}
	}

	upgr := &action.Upgrade{
		Driver: driverImpl,
	}
	err = upgr.Run(&claim, creds, up.out)

	// persist the claim, regardless of the success of the upgrade action
	persistErr := claimStorage().Store(claim)

	if err != nil {
		return fmt.Errorf("could not upgrade %q: %s", up.name, err)
	}
	return persistErr
}

func prepareBundleFile(bundle, bundleFile string, insecure bool) (string, error) {
	if bundle != "" && bundleFile != "" {
		return "", ErrBundleAndBundleFile
	}

	if bundle != "" {
		bundleFile, err := getBundleFilepath(bundle, homePath(), insecure)
		if err != nil {
			return bundleFile, err
		}
	}

	return bundleFile, nil
}
