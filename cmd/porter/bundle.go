package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildBundleCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "bundle commands",
	}

	cmd.AddCommand(buildBundleInstallCommand(p))

	return cmd
}

func buildBundleInstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.InstallOptions{}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a bundle",
		Example: `  porter install
  porter install --insecure
  porter install --file myapp/bundle.json
  porter install --name MyAppInDev
  porter install --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter install --cred azure --cred kubernetes
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.Insecure, "insecure", false,
		"Allow installing untrusted bundles")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the CNAB definition to install. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.Name, "name", "",
		"Name of the claim, defaults to the name of the bundle")
	f.StringSliceVar(&opts.ParamFiles, "param-file", nil,
		"Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")

	return cmd
}

func buildInstallCommand(p *porter.Porter) *cobra.Command {
	return buildBundleInstallCommand(p)
}
