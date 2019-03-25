package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/deislabs/duffle/pkg/action"
)

const uninstallUsage = `Uninstalls an installation of a CNAB bundle.

When using '--parameters' or '--set', the uninstall command will replace the old
parameters with the new ones supplied (even if the new set is an empty set). If neither
'--parameters' nor '--set' is passed, then the parameters used for 'duffle install' will
be re-used.
`

type uninstallCmd struct {
	out              io.Writer
	name             string
	valuesFile       string
	driver           string
	bundle           string
	bundleFile       string
	setParams        []string
	insecure         bool
	credentialsFiles []string
}

func newUninstallCmd(w io.Writer) *cobra.Command {
	uninstall := &uninstallCmd{out: w}

	cmd := &cobra.Command{
		Use:   "uninstall [NAME]",
		Short: "uninstall CNAB installation",
		Long:  uninstallUsage,
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return uninstall.setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			uninstall.name = args[0]
			return uninstall.run()
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&uninstall.driver, "driver", "d", "docker", "Specify a driver name")
	flags.StringArrayVarP(&uninstall.credentialsFiles, "credentials", "c", []string{}, "Specify credentials to use inside the CNAB bundle. This can be a credentialset name or a path to a file.")
	flags.StringVarP(&uninstall.valuesFile, "parameters", "p", "", "Specify file containing parameters. Formats: toml, MORE SOON")
	flags.StringVarP(&uninstall.bundle, "bundle", "b", "", "bundle to uninstall")
	flags.StringVar(&uninstall.bundleFile, "bundle-file", "", "path to a bundle file to uninstal")
	flags.StringArrayVarP(&uninstall.setParams, "set", "s", []string{}, "set individual parameters as NAME=VALUE pairs")
	flags.BoolVarP(&uninstall.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}

func (un *uninstallCmd) setup() error {
	bundleFile, err := prepareBundleFile(un.bundle, un.bundleFile, un.insecure)
	if err != nil {
		return err
	}

	un.bundleFile = bundleFile
	return nil
}

func (un *uninstallCmd) run() error {
	claim, err := claimStorage().Read(un.name)
	if err != nil {
		return fmt.Errorf("%v not found: %v", un.name, err)
	}

	if un.bundleFile != "" {
		b, err := loadBundle(un.bundleFile, un.insecure)
		if err != nil {
			return err
		}
		claim.Bundle = b
	}

	// If no params are specified, allow re-use. But if params are set -- even if empty --
	// replace the existing params.
	if len(un.setParams) > 0 || un.valuesFile != "" {
		if claim.Bundle == nil {
			return errors.New("parameters can only be set if a bundle is provided")
		}
		params, err := calculateParamValues(claim.Bundle, un.valuesFile, un.setParams, []string{})
		if err != nil {
			return err
		}
		claim.Parameters = params
	}

	driverImpl, err := prepareDriver(un.driver)
	if err != nil {
		return fmt.Errorf("could not prepare driver: %s", err)
	}

	creds, err := loadCredentials(un.credentialsFiles, claim.Bundle)
	if err != nil {
		return fmt.Errorf("could not load credentials: %s", err)
	}

	uninst := &action.Uninstall{
		Driver: driverImpl,
	}

	fmt.Fprintln(un.out, "Executing uninstall action...")
	if err := uninst.Run(&claim, creds, un.out); err != nil {
		return fmt.Errorf("could not uninstall %q: %s", un.name, err)
	}
	return claimStorage().Delete(un.name)
}
