package main

import (
	"github.com/spf13/cobra"

	"github.com/deislabs/porter/pkg/porter"
)

func buildBundleCommands(p *porter.Porter) *cobra.Command {
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
	cmd.AddCommand(buildBundleInstallCommand(p))
	cmd.AddCommand(buildBundleUpgradeCommand(p))
	cmd.AddCommand(buildBundleInvokeCommand(p))
	cmd.AddCommand(buildBundleUninstallCommand(p))
	cmd.AddCommand(buildBundleArchiveCommand(p))
	cmd.AddCommand(buildBundleExplainCommand(p))

	return cmd
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

func buildBundleInstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.InstallOptions{}
	cmd := &cobra.Command{
		Use:   "install [INSTANCE]",
		Short: "Install a new instance of a bundle",
		Long: `Install a new instance of a bundle.

The first argument is the bundle instance name to create for the installation. This defaults to the name of the bundle. 

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
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
		"Use a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")
	f.BoolVar(&opts.Force, "force", false,
		"Force a fresh pull of the bundle and all dependencies")
	return cmd
}

func buildBundleUpgradeCommand(p *porter.Porter) *cobra.Command {
	opts := porter.UpgradeOptions{}
	cmd := &cobra.Command{
		Use:   "upgrade [INSTANCE]",
		Short: "Upgrade a bundle instance",
		Long: `Upgrade a bundle instance.

The first argument is the bundle instance name to upgrade. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
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
		"Use a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")
	f.BoolVar(&opts.Force, "force", false,
		"Force a fresh pull of the bundle and all dependencies")

	return cmd
}

func buildBundleInvokeCommand(p *porter.Porter) *cobra.Command {
	opts := porter.InvokeOptions{}
	cmd := &cobra.Command{
		Use:   "invoke [INSTANCE] --action ACTION",
		Short: "Invoke a custom action on a bundle instance",
		Long: `Invoke a custom action on a bundle instance.

The first argument is the bundle instance name upon which to invoke the action. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle invoke --action ACTION
  porter bundle invoke --action ACTION MyAppInDev --file myapp/bundle.json
  porter bundle invoke --action ACTION --param-file base-values.txt --param-file dev-values.txt --param test-mode=true --param header-color=blue
  porter bundle invoke --action ACTION --cred azure --cred kubernetes
  porter bundle invoke --action ACTION --driver debug
  porter bundle invoke --action ACTION MyAppFromTag --tag deislabs/porter-kube-bundle:v1.0
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InvokeBundle(opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.Action, "action", "",
		"Custom action name to invoke.")
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
		"Use a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")
	f.BoolVar(&opts.Force, "force", false,
		"Force a fresh pull of the bundle and all dependencies")

	return cmd
}

func buildBundleUninstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.UninstallOptions{}
	cmd := &cobra.Command{
		Use:   "uninstall [INSTANCE]",
		Short: "Uninstall a bundle instance",
		Long: `Uninstall a bundle instance

The first argument is the bundle instance name to uninstall. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
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
		"Use a bundle in an OCI registry specified by the given tag")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false,
		"Don't require TLS for the registry")
	f.BoolVar(&opts.Force, "force", false,
		"Force a fresh pull of the bundle and all dependencies")

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
	porter bundle publish --archive /tmp/mybuns.tgz --tag myrepo/my-buns:0.1.0
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
	f.StringVarP(&opts.ArchiveFile, "archive", "a", "", "Path to the bundle archive in .tgz format")
	f.StringVarP(&opts.Tag, "tag", "t", "", "Bundle tag for newly published bundle; required if --archive is supplied")

	return &cmd
}

func buildBundleArchiveCommand(p *porter.Porter) *cobra.Command {

	opts := porter.ArchiveOptions{}
	cmd := cobra.Command{
		Hidden: true,
		Use:    "archive",
		Short:  "Archive a bundle",
		Long:   "Archives a bundle by generating a gzipped tar archive containing the bundle, invocation image and any referenced images.",
		Example: `  porter bundle archive [FILENAME]
  porter bundle archive --file another/porter.yaml [FILENAME]
  porter bundle archive --cnab-file some/bundle.json [FILENAME]
  porter bundle archive --tag repo/bundle:tag [FILENAME]
		  `,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Archive(opts)
		},
	}
	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "", "Path to the Porter manifest. Defaults to `porter.yaml` in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "", "Path to the CNAB bundle.json file.")
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Use a bundle in an OCI registry specified by the given tag")
	return &cmd
}
