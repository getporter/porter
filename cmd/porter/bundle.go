package main

import (
	"github.com/spf13/cobra"

	"get.porter.sh/porter/pkg/porter"
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
	cmd.AddCommand(buildBundleLintCommand(p))
	cmd.AddCommand(buildBundleInstallCommand(p))
	cmd.AddCommand(buildBundleUpgradeCommand(p))
	cmd.AddCommand(buildBundleInvokeCommand(p))
	cmd.AddCommand(buildBundleUninstallCommand(p))
	cmd.AddCommand(buildBundleArchiveCommand(p))
	cmd.AddCommand(buildBundleExplainCommand(p))
	cmd.AddCommand(buildBundleCopyCommand(p))
	cmd.AddCommand(buildBundleInspectCommand(p))

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

	f := cmd.Flags()
	f.BoolVar(&opts.NoLint, "no-lint", false, "Do not run the linter")
	f.BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose logging")

	return cmd
}

func buildBundleLintCommand(p *porter.Porter) *cobra.Command {
	var opts porter.LintOptions
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint a bundle",
		Long: `Check the bundle for problems and adherence to best practices by running linters for porter and the mixins used in the bundle.

The lint command is run automatically when you build a bundle. The command is available separately so that you can just lint your bundle without also building it.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintLintResults(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Specify an output format.  Allowed values: "+porter.AllowedLintFormats.String())
	f.BoolVarP(&opts.Verbose, "verbose", "v", false,
		"Enable verbose logging")

	return cmd
}

func buildBundleInstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewInstallOptions()
	cmd := &cobra.Command{
		Use:   "install [INSTALLATION]",
		Short: "Create a new installation of a bundle",
		Long: `Create a new installation of a bundle.

The first argument is the name of the installation to create. This defaults to the name of the bundle. 

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle install
  porter bundle install MyAppFromTag --tag getporter/kubernetes:v0.1.0
  porter bundle install --tag localhost:5000/getporter/kubernetes:v0.1.0 --insecure-registry --force
  porter bundle install MyAppInDev --file myapp/bundle.json
  porter bundle install --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle install --cred azure --cred kubernetes
  porter bundle install --driver debug
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowAccessToDockerHost, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	addBundlePullFlags(f, &opts.BundlePullOptions)
	return cmd
}

func buildBundleUpgradeCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewUpgradeOptions()
	cmd := &cobra.Command{
		Use:   "upgrade [INSTALLATION]",
		Short: "Upgrade an installation",
		Long: `Upgrade an installation.

The first argument is the installation name to upgrade. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle upgrade
  porter bundle upgrade --tag getporter/kubernetes:v0.1.0
  porter bundle upgrade --tag localhost:5000/getporter/kubernetes:v0.1.0 --insecure-registry --force
  porter bundle upgrade MyAppInDev --file myapp/bundle.json
  porter bundle upgrade --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle upgrade --cred azure --cred kubernetes
  porter bundle upgrade --driver debug
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UpgradeBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowAccessToDockerHost, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	return cmd
}

func buildBundleInvokeCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewInvokeOptions()
	cmd := &cobra.Command{
		Use:   "invoke [INSTALLATION] --action ACTION",
		Short: "Invoke a custom action on an installation",
		Long: `Invoke a custom action on an installation.

The first argument is the installation name upon which to invoke the action. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle invoke --action ACTION
  porter bundle invoke --tag getporter/kubernetes:v0.1.0
  porter bundle invoke --tag localhost:5000/getporter/kubernetes:v0.1.0 --insecure-registry --force
  porter bundle invoke --action ACTION MyAppInDev --file myapp/bundle.json
  porter bundle invoke --action ACTION  --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle invoke --action ACTION --cred azure --cred kubernetes
  porter bundle invoke --action ACTION --driver debug
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InvokeBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowAccessToDockerHost, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVar(&opts.Action, "action", "",
		"Custom action name to invoke.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	return cmd
}

func buildBundleUninstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewUninstallOptions()
	cmd := &cobra.Command{
		Use:   "uninstall [INSTALLATION]",
		Short: "Uninstall an installation",
		Long: `Uninstall an installation

The first argument is the installation name to uninstall. This defaults to the name of the bundle.

Porter uses the Docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.`,
		Example: `  porter bundle uninstall
  porter bundle uninstall --tag getporter/kubernetes:v0.1.0
  porter bundle uninstall --tag localhost:5000/getporter/kubernetes:v0.1.0 --insecure-registry --force
  porter bundle uninstall MyAppInDev --file myapp/bundle.json
  porter bundle uninstall --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle uninstall --cred azure --cred kubernetes
  porter bundle uninstall --driver debug
  porter bundle uninstall --delete
  porter bundle uninstall --force-delete
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UninstallBundle(opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowAccessToDockerHost, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://porter.sh/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory. Optional unless a newer version of the bundle should be used to uninstall the bundle.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringSliceVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.")
	f.StringSliceVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringSliceVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when uninstalling the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.BoolVar(&opts.Delete, "delete", false,
		"Delete all records associated with the installation, assuming the uninstall action succeeds")
	f.BoolVar(&opts.ForceDelete, "force-delete", false,
		"UNSAFE. Delete all records associated with the installation, even if uninstall fails. This is intended for cleaning up test data and is not recommended for production environments.")
	addBundlePullFlags(f, &opts.BundlePullOptions)

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
	f.StringVarP(&opts.ArchiveFile, "archive", "a", "", "Path to the bundle archive in .tgz format")
	addTagFlag(f, &opts.BundlePullOptions)
	addInsecureRegistryFlag(f, &opts.BundlePullOptions)
	// We aren't using addBundlePullFlags because we don't use --force since we are pushing, and that flag isn't needed

	return &cmd
}

func buildBundleArchiveCommand(p *porter.Porter) *cobra.Command {

	opts := porter.ArchiveOptions{}
	cmd := cobra.Command{
		Use:   "archive FILENAME --tag PUBLISHED_BUNDLE",
		Short: "Archive a bundle from a tag",
		Long:  "Archives a bundle by generating a gzipped tar archive containing the bundle, invocation image and any referenced images.",
		Example: `  porter bundle archive mybun.tgz --tag getporter/porter-hello:v0.1.0
  porter bundle archive mybun.tgz --tag localhost:5000/getporter/porter-hello:v0.1.0 --force
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Archive(opts)
		},
	}

	addBundlePullFlags(cmd.Flags(), &opts.BundlePullOptions)

	return &cmd
}
