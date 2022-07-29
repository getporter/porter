package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func buildInstallationCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "installations",
		Aliases: []string{"inst", "installation"},
		Short:   "Installation commands",
		Long:    "Commands for working with installations of a bundle",
	}
	cmd.Annotations = map[string]string{
		"group": "resource",
	}

	cmd.AddCommand(buildInstallationsListCommand(p))
	cmd.AddCommand(buildInstallationShowCommand(p))
	cmd.AddCommand(buildInstallationApplyCommand(p))
	cmd.AddCommand(buildInstallationOutputsCommands(p))
	cmd.AddCommand(buildInstallationDeleteCommand(p))
	cmd.AddCommand(buildInstallationLogCommands(p))
	cmd.AddCommand(buildInstallationRunsCommands(p))
	cmd.AddCommand(buildInstallationInstallCommand(p))
	cmd.AddCommand(buildInstallationUpgradeCommand(p))
	cmd.AddCommand(buildInstallationInvokeCommand(p))
	cmd.AddCommand(buildInstallationUninstallCommand(p))

	return cmd
}

func buildInstallationsListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed bundles",
		Long: `List all bundles installed by Porter.

A listing of bundles currently installed by Porter will be provided, along with metadata such as creation time, last action, last status, etc.
Optionally filters the results name, which returns all results whose name contain the provided query.
The results may also be filtered by associated labels and the namespace in which the installation is defined. 

Optional output formats include json and yaml.`,
		Example: `  porter installations list
  porter installations list -o json
  porter installations list --all-namespaces,
  porter installations list --label owner=myname --namespace dev
  porter installations list --name myapp
  porter installations list --skip 2 --limit 2`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintInstallations(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Filter the installations by namespace. Defaults to the global namespace.")
	f.BoolVar(&opts.AllNamespaces, "all-namespaces", false,
		"Include all namespaces in the results.")
	f.StringVar(&opts.Name, "name", "",
		"Filter the installations where the name contains the specified substring.")
	f.StringSliceVarP(&opts.Labels, "label", "l", nil,
		"Filter the installations by a label formatted as: KEY=VALUE. May be specified multiple times.")
	f.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Specify an output format.  Allowed values: plaintext, json, yaml")
	f.Int64Var(&opts.Skip, "skip", 0,
		"Skip the number of installations by a certain amount. Defaults to 0.")
	f.Int64Var(&opts.Limit, "limit", 0,
		"Limit the number of installations by a certain amount. Defaults to 0.")

	return cmd
}

func buildInstallationShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ShowOptions{}

	cmd := cobra.Command{
		Use:   "show [INSTALLATION]",
		Short: "Show an installation of a bundle",
		Long:  "Displays info relating to an installation of a bundle, including status and a listing of outputs.",
		Example: `  porter installation show
  porter installation show another-bundle

Optional output formats include json and yaml.
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowInstallation(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the installation is defined. Defaults to the global namespace.")
	f.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Specify an output format.  Allowed values: plaintext, json, yaml")

	return &cmd
}

func buildInstallationApplyCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ApplyOptions{}

	cmd := cobra.Command{
		Use:   "apply FILE",
		Short: "Apply changes to an installation",
		Long: `Apply changes from the specified file to an installation. If the installation doesn't already exist, it is created.
The installation's bundle is automatically executed if changes are detected.

When the namespace is not set in the file, the current namespace is used.

You can use the show command to create the initial file:
  porter installation show mybuns --output yaml > mybuns.yaml
`,
		Example: `  porter installation apply myapp.yaml
  porter installation apply myapp.yaml --dry-run
  porter installation apply myapp.yaml --force`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallationApply(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the installation is defined. Defaults to the namespace defined in the file.")
	f.BoolVar(&opts.Force, "force", false,
		"Force the bundle to be executed when no changes are detected.")
	f.BoolVar(&opts.DryRun, "dry-run", false,
		"Evaluate if the bundle would be executed based on the changes in the file.")
	return &cmd
}

func buildInstallationDeleteCommand(p *porter.Porter) *cobra.Command {
	opts := porter.DeleteOptions{}

	cmd := cobra.Command{
		Use:   "delete [INSTALLATION]",
		Short: "Delete an installation",
		Long:  "Deletes all records and outputs associated with an installation",
		Example: `  porter installation delete
  porter installation delete wordpress
  porter installation delete --force
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.DeleteInstallation(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the installation is defined. Defaults to the global namespace.")
	f.BoolVar(&opts.Force, "force", false,
		"Force a delete the installation, regardless of last completed action")

	return &cmd
}

func buildInstallationRunsCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "runs",
		Aliases: []string{"run"},
		Short:   "Commands for working with runs of an Installation",
		Long:    "Commands for working with runs of an Installation",
	}

	cmd.AddCommand(buildInstallationRunsListCommand(p))

	return cmd
}

func buildInstallationRunsListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.RunListOptions{}

	cmd := cobra.Command{
		Use:   "list",
		Short: "List runs of an Installation",
		Long:  "List runs of an Installation",
		Example: `  porter installation runs list [NAME] [--namespace NAMESPACE] [--output FORMAT]

  porter installations runs list --name myapp --namespace dev

`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintInstallationRuns(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the installation is defined. Defaults to the global namespace.")
	f.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Specify an output format.  Allowed values: plaintext, json, yaml")

	return &cmd
}

func buildInstallationInstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewInstallOptions()
	cmd := &cobra.Command{
		Use:   "install [INSTALLATION]",
		Short: "Create a new installation of a bundle",
		Long: `Create a new installation of a bundle.

The first argument is the name of the installation to create. This defaults to the name of the bundle. 

Once a bundle has been successfully installed, the install action cannot be repeated. This is a precaution to avoid accidentally overwriting an existing installation. If you need to re-run install, which is common when authoring a bundle, you can use the --force flag to by-pass this check.

Porter uses the docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
`,
		Example: `  porter installation install
  porter installation install MyAppFromReference --reference ghcr.io/getporter/examples/kubernetes:v0.2.0 --namespace dev
  porter installation install --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter installation install MyAppInDev --file myapp/bundle.json
  porter installation install --parameter-set azure --param test-mode=true --param header-color=blue
  porter installation install --credential-set azure --credential-set kubernetes
  porter installation install --driver debug
  porter installation install --label env=dev --label owner=myuser
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Create the installation in the specified namespace. Defaults to the global namespace.")
	f.StringSliceVarP(&opts.Labels, "label", "l", nil,
		"Associate the specified labels with the installation. May be specified multiple times.")
	addBundleActionFlags(f, opts)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

func buildInstallationUpgradeCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewUpgradeOptions()
	cmd := &cobra.Command{
		Use:   "upgrade [INSTALLATION]",
		Short: "Upgrade an installation",
		Long: `Upgrade an installation.

The first argument is the installation name to upgrade. This defaults to the name of the bundle.

Porter uses the docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
`,
		Example: `  porter installation upgrade --version 0.2.0
  porter installation upgrade --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter installation upgrade --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter installation upgrade MyAppInDev --file myapp/bundle.json
  porter installation upgrade --parameter-set azure --param test-mode=true --param header-color=blue
  porter installation upgrade --credential-set azure --credential-set kubernetes
  porter installation upgrade --driver debug
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UpgradeBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace of the specified installation. Defaults to the global namespace.")
	f.StringVar(&opts.Version, "version", "",
		"Version to which the installation should be upgraded. This represents the version of the bundle, which assumes the convention of setting the bundle tag to its version.")
	addBundleActionFlags(f, opts)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

func buildInstallationInvokeCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewInvokeOptions()
	cmd := &cobra.Command{
		Use:   "invoke [INSTALLATION] --action ACTION",
		Short: "Invoke a custom action on an installation",
		Long: `Invoke a custom action on an installation.

The first argument is the installation name upon which to invoke the action. This defaults to the name of the bundle.

Porter uses the docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
`,
		Example: `  porter installation invoke --action ACTION
  porter installation invoke --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter installation invoke --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter installation invoke --action ACTION MyAppInDev --file myapp/bundle.json
  porter installation invoke --action ACTION  --parameter-set azure --param test-mode=true --param header-color=blue
  porter installation invoke --action ACTION --credential-set azure --credential-set kubernetes
  porter installation invoke --action ACTION --driver debug
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InvokeBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.Action, "action", "",
		"Custom action name to invoke.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace of the specified installation. Defaults to the global namespace.")
	addBundleActionFlags(f, opts)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

func buildInstallationUninstallCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewUninstallOptions()
	cmd := &cobra.Command{
		Use:   "uninstall [INSTALLATION]",
		Short: "Uninstall an installation",
		Long: `Uninstall an installation

The first argument is the installation name to uninstall. This defaults to the name of the bundle.

Porter uses the docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d'' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
`,
		Example: `  porter installation uninstall
  porter installation uninstall --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter installation uninstall --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter installation uninstall MyAppInDev --file myapp/bundle.json
  porter installation uninstall --parameter-set azure --param test-mode=true --param header-color=blue
  porter installation uninstall --credential-set azure --credential-set kubernetes
  porter installation uninstall --driver debug
  porter installation uninstall --delete
  porter installation uninstall --force-delete
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UninstallBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory. Optional unless a newer version of the bundle should be used to uninstall the bundle.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.BoolVar(&opts.Delete, "delete", false,
		"Delete all records associated with the installation, assuming the uninstall action succeeds")
	f.BoolVar(&opts.ForceDelete, "force-delete", false,
		"UNSAFE. Delete all records associated with the installation, even if uninstall fails. This is intended for cleaning up test data and is not recommended for production environments.")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace of the specified installation. Defaults to the global namespace.")
	addBundleActionFlags(f, opts)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

// Add flags for command that execute a bundle (install, upgrade, invoke and uninstall)
func addBundleActionFlags(f *pflag.FlagSet, actionOpts porter.BundleAction) {
	opts := actionOpts.GetOptions()
	addBundlePullFlags(f, &opts.BundlePullOptions)
	f.BoolVar(&opts.AllowDockerHostAccess, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://getporter.org/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.BoolVar(&opts.NoLogs, "no-logs", false,
		"Do not persist the bundle execution logs")
	f.StringArrayVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Parameter sets to use when running the bundle. It should be a named set of parameters and may be specified multiple times.")
	f.StringArrayVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringArrayVarP(&opts.CredentialIdentifiers, "credential-set", "c", nil,
		"Credential sets to use when running the bundle. It should be a named set of credentials and may be specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.BoolVar(&opts.DebugMode, "debug", false,
		"Run the bundle in debug mode.")

	// Gracefully support any renamed flags
	f.StringArrayVar(&opts.CredentialIdentifiers, "cred", nil, "DEPRECATED")
	f.MarkDeprecated("cred", "please use credential-set instead.")
}
