package main

import (
	"strings"

	"github.com/deislabs/porter/pkg/porter"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/spf13/cobra"
)

func buildBundlesCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bundles",
		Aliases: []string{"bundle"},
		Short:   "Bundle commands",
		Long:    "Commands for working with bundles. These all have shortcuts so that you can call these commands without the bundle resource prefix. For example, porter bundle install is available as porter install as well.",
	}
	cmd.Annotations = map[string]string{
		"group": "resource",
	}

	cmd.AddCommand(buildBundleCreateCommand(p))
	cmd.AddCommand(buildBundleBuildCommand(p))
	cmd.AddCommand(buildBundleListCommand(p))
	cmd.AddCommand(buildBundleInstallCommand(p))
	cmd.AddCommand(buildBundleUpgradeCommand(p))
	cmd.AddCommand(buildBundleUninstallCommand(p))
	cmd.AddCommand(buildBundleShowCommand(p))

	return cmd
}

func buildBundleAliasCommands(p *porter.Porter) []*cobra.Command {
	return []*cobra.Command{
		buildCreateCommand(p),
		buildBuildCommand(p),
		buildInstallCommand(p),
		buildUpgradeCommand(p),
		buildUninstallCommand(p),
		buildPublishCommand(p),
		buildShowCommand(p),
	}
}

func buildBundleCreateCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a bundle",
		Long:  "Create a bundle. This generates a porter bundle in the current directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Create()
		},
	}
}

func buildCreateCommand(p *porter.Porter) *cobra.Command {
	cmd := buildBundleCreateCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle create", "porter create", -1)
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildBundleBuildCommand(p *porter.Porter) *cobra.Command {
	opts := porter.BuildOptions{}

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a bundle",
		Long:  "Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Build(opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose logging")

	return cmd
}

func buildBuildCommand(p *porter.Porter) *cobra.Command {
	cmd := buildBundleBuildCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle build", "porter build", -1)
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildBundleListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list installed bundles",
		Long: `List all bundles installed by Porter.

A listing of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.

Optional output formats include json and yaml.`,
		Example: `  porter bundle list
  porter bundle list -o json`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			opts.Format, err = printer.ParseFormat(opts.RawFormat)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ListBundles(printer.PrintOptions{Format: opts.Format})
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}

func buildBundleInstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.InstallOptions{}
	cmd := &cobra.Command{
		Use:   "install [CLAIM]",
		Short: "Install a bundle",
		Long: `Install a bundle.

The first argument is the name of the claim to create for the installation. The claim name defaults to the name of the bundle. 

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For instance, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle install
  porter bundle install --insecure
  porter bundle install MyAppInDev --file myapp/bundle.json
  porter bundle install --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter bundle install --cred azure --cred kubernetes
  porter bundle install --driver debug
  porter bundle install MyAppFromTag --tag deislabs/porter-kube-bundle:v1.0
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.Insecure, "insecure", true,
		"Allow working with untrusted bundles")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVar(&opts.ParamFiles, "param-file", nil,
		"Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Install from a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")
	return cmd
}

func buildInstallCommand(p *porter.Porter) *cobra.Command {
	cmd := buildBundleInstallCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle install", "porter install", -1)
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildBundleUpgradeCommand(p *porter.Porter) *cobra.Command {
	opts := porter.UpgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade [CLAIM]",
		Short: "Upgrade a bundle",
		Long: `Upgrade a bundle.

The first argument is the name of the claim to upgrade. The claim name defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For instance, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle upgrade
  porter bundle upgrade --insecure
  porter bundle upgrade MyAppInDev --file myapp/bundle.json
  porter bundle upgrade --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter bundle upgrade --cred azure --cred kubernetes
  porter bundle upgrade --driver debug
  porter bundle upgrade MyAppFromTag --tag deislabs/porter-kube-bundle:v1.0
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UpgradeBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.Insecure, "insecure", true,
		"Allow working with untrusted bundles")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVar(&opts.ParamFiles, "param-file", nil,
		"Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Install from a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")

	return cmd
}

func buildUpgradeCommand(p *porter.Porter) *cobra.Command {
	cmd := buildBundleUpgradeCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle upgrade", "porter upgrade", -1)
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildBundleUninstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.UninstallOptions{}
	cmd := &cobra.Command{
		Use:   "uninstall [CLAIM]",
		Short: "Uninstall a bundle",
		Long: `Uninstall a bundle

The first argument is the name of the claim to uninstall. The claim name defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For instance, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle uninstall
  porter bundle uninstall --insecure
  porter bundle uninstall MyAppInDev --file myapp/bundle.json
  porter bundle uninstall --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter bundle uninstall --cred azure --cred kubernetes
  porter bundle uninstall --driver debug
  porter bundle uninstall MyAppFromTag --tag deislabs/porter-kube-bundle:v1.0
  
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UninstallBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.Insecure, "insecure", true,
		"Allow working with untrusted bundles")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory. Optional unless a newer version of the bundle should be used to uninstall the bundle.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVar(&opts.ParamFiles, "param-file", nil,
		"Path to a parameters definition file for the bundle, each line in the form of NAME=VALUE. May be specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters set with the same name using --param-file. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when uninstalling the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Install from a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")

	return cmd
}

func buildUninstallCommand(p *porter.Porter) *cobra.Command {
	cmd := buildBundleUninstallCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle uninstall", "porter uninstall", -1)
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

func buildBundlePublishCommand(p *porter.Porter) *cobra.Command {

	opts := porter.PublishOptions{}
	cmd := cobra.Command{
		Use:   "publish",
		Short: "Publish a bundle",
		Long:  "Publishes a bundle by pushing the invocation image and bundle to a registry.",
		Example: `  porter bundle publish
  porter bundle publish --file myapp/porter.yaml
  porter bundle publish --insecure
		`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Publish(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "", "Path to the Porter manifest. Defaults to `porter.yaml` in the current directory.")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false, "Don't require TLS for the registry.")
	return &cmd
}

func buildPublishCommand(p *porter.Porter) *cobra.Command {
	cmd := buildBundlePublishCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle publish", "porter publish", -1)
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}

// TODO: test!
func buildBundleShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ShowOptions{}

	cmd := cobra.Command{
		Use:   "show",
		Short: "Show a bundle",
		Long:  "Displays info relating to a bundle claim, including status and a listing of outputs.",
		Example: `  porter bundle show [CLAIM]

Optional output formats include json and yaml.
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowBundle(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return &cmd
}

func buildShowCommand(p *porter.Porter) *cobra.Command {
	cmd := buildBundleShowCommand(p)
	cmd.Example = strings.Replace(cmd.Example, "porter bundle show", "porter show", -1)
	cmd.Annotations = map[string]string{
		"group": "alias",
	}
	return cmd
}
