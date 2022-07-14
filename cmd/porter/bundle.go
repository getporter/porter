package main

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildBundleCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bundles",
		Aliases: []string{"bundle"},
		Short:   "Bundle commands",
		Long:    "Commands for working with bundles. These all have shortcuts so that you can call these commands without the bundle resource prefix. For example, porter bundle install is available as porter install as well.",
	}
	cli.SetCommandGroup(cmd, "resource")

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
		Long: `Builds the bundle in the current directory by generating a Dockerfile and a CNAB bundle.json, and then building the invocation image.

The docker driver builds the bundle image using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
'
`,
		Example: `  porter build
  porter build --name newbuns
  porter build --version 0.1.0
  porter build --file path/to/porter.yaml
  porter build --dir path/to/build/context
  porter build --custom version=0.2.0 --custom myapp.version=0.1.2
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Build(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.NoLint, "no-lint", false, "Do not run the linter")
	f.BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose logging")
	f.StringVar(&opts.Name, "name", "", "Override the bundle name")
	f.StringVar(&opts.Version, "version", "", "Override the bundle version")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the Porter manifest. The path is relative to the build context directory. Defaults to porter.yaml in the current directory.")
	f.StringVarP(&opts.Dir, "dir", "d", "",
		"Path to the build context directory where all bundle assets are located. Defaults to the current directory.")
	f.StringVar(&opts.Driver, "driver", porter.BuildDriverDefault,
		fmt.Sprintf("Driver for building the invocation image. Allowed values are: %s", strings.Join(porter.BuildDriverAllowedValues, ", ")))
	f.MarkHidden("driver") // Hide the driver flag since there aren't any choices to make right now
	f.StringArrayVar(&opts.BuildArgs, "build-arg", nil,
		"Set build arguments in the template Dockerfile (format: NAME=VALUE). May be specified multiple times.")
	f.StringArrayVar(&opts.SSH, "ssh", nil,
		"SSH agent socket or keys to expose to the build (format: default|<id>[=<socket>|<key>[,<key>]]). May be specified multiple times.")
	f.StringArrayVar(&opts.Secrets, "secret", nil,
		"Secret file to expose to the build (format: id=mysecret,src=/local/secret). Custom values are assessible as build arguments in the template Dockerfile and in the manifest using template variables. May be specified multiple times.")
	f.BoolVar(&opts.NoCache, "no-cache", false,
		"Do not use the Docker cache when building the bundle's invocation image.")
	f.StringArrayVar(&opts.Customs, "custom", nil,
		"Define an individual key-value pair for the custom section in the form of NAME=VALUE. Use dot notation to specify a nested custom field. May be specified multiple times.")

	// Allow configuring the --driver flag with build-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"build-driver"},
	}

	return cmd
}

func buildBundleLintCommand(p *porter.Porter) *cobra.Command {
	var opts porter.LintOptions
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint a bundle",
		Long: `Check the bundle for problems and adherence to best practices by running linters for porter and the mixins used in the bundle.

The lint command is run automatically when you build a bundle. The command is available separately so that you can just lint your bundle without also building it.`,
		Example: `  porter lint
  porter lint --file path/to/porter.yaml
  porter lint --output plaintext
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintLintResults(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVarP(&opts.RawFormat, "output", "o", string(porter.LintDefaultFormats),
		"Specify an output format.  Allowed values: "+porter.LintAllowFormats.String())
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

Once a bundle has been successfully installed, the install action cannot be repeated. This is a precaution to avoid accidentally overwriting an existing installation. If you need to re-run install, which is common when authoring a bundle, you can use the --force flag to by-pass this check.

Porter uses the docker driver as the default runtime for executing a bundle's invocation image, but an alternate driver may be supplied via '--driver/-d' or the PORTER_RUNTIME_DRIVER environment variable.
For example, the 'debug' driver may be specified, which simply logs the info given to it and then exits.

The docker driver runs the bundle container using the local Docker host. To use a remote Docker host, set the following environment variables:
  DOCKER_HOST (required)
  DOCKER_TLS_VERIFY (optional)
  DOCKER_CERT_PATH (optional)
`,
		Example: `  porter bundle install
  porter bundle install MyAppFromReference --reference ghcr.io/getporter/examples/kubernetes:v0.2.0 --namespace dev
  porter bundle install --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter bundle install MyAppInDev --file myapp/bundle.json
  porter bundle install --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle install --cred azure --cred kubernetes
  porter bundle install --driver debug
  porter bundle install --label env=dev --label owner=myuser
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowDockerHostAccess, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://getporter.org/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringArrayVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set for the bundle. It should be a named set of parameters and may be specified multiple times.")
	f.StringArrayVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringArrayVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. It should be a named set of credentials and may be specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Create the installation in the specified namespace. Defaults to the global namespace.")
	f.StringSliceVarP(&opts.Labels, "label", "l", nil,
		"Associate the specified labels with the installation. May be specified multiple times.")
	f.BoolVar(&opts.NoLogs, "no-logs", false,
		"Do not persist the bundle execution logs")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

func buildBundleUpgradeCommand(p *porter.Porter) *cobra.Command {
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
		Example: `  porter bundle upgrade --version 0.2.0
  porter bundle upgrade --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter bundle upgrade --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter bundle upgrade MyAppInDev --file myapp/bundle.json
  porter bundle upgrade --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle upgrade --cred azure --cred kubernetes
  porter bundle upgrade --driver debug
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UpgradeBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowDockerHostAccess, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://getporter.org/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringArrayVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.")
	f.StringArrayVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringArrayVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. It should be a named set of credentials and may be specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace of the specified installation. Defaults to the global namespace.")
	f.StringVar(&opts.Version, "version", "",
		"Version to which the installation should be upgraded. This represents the version of the bundle, which assumes the convention of setting the bundle tag to its version.")
	f.BoolVar(&opts.NoLogs, "no-logs", false,
		"Do not persist the bundle execution logs")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

func buildBundleInvokeCommand(p *porter.Porter) *cobra.Command {
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
		Example: `  porter bundle invoke --action ACTION
  porter bundle invoke --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter bundle invoke --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter bundle invoke --action ACTION MyAppInDev --file myapp/bundle.json
  porter bundle invoke --action ACTION  --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle invoke --action ACTION --cred azure --cred kubernetes
  porter bundle invoke --action ACTION --driver debug
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InvokeBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowDockerHostAccess, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://getporter.org/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVar(&opts.Action, "action", "",
		"Custom action name to invoke.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringArrayVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.")
	f.StringArrayVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringArrayVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when installing the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace of the specified installation. Defaults to the global namespace.")
	f.BoolVar(&opts.NoLogs, "no-logs", false,
		"Do not persist the bundle execution logs")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

func buildBundleUninstallCommand(p *porter.Porter) *cobra.Command {
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
		Example: `  porter bundle uninstall
  porter bundle uninstall --reference ghcr.io/getporter/examples/kubernetes:v0.2.0
  porter bundle uninstall --reference localhost:5000/ghcr.io/getporter/examples/kubernetes:v0.2.0 --insecure-registry --force
  porter bundle uninstall MyAppInDev --file myapp/bundle.json
  porter bundle uninstall --parameter-set azure --param test-mode=true --param header-color=blue
  porter bundle uninstall --cred azure --cred kubernetes
  porter bundle uninstall --driver debug
  porter bundle uninstall --delete
  porter bundle uninstall --force-delete
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UninstallBundle(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.BoolVar(&opts.AllowDockerHostAccess, "allow-docker-host-access", false,
		"Controls if the bundle should have access to the host's Docker daemon with elevated privileges. See https://getporter.org/configuration/#allow-docker-host-access for the full implications of this flag.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory. Optional unless a newer version of the bundle should be used to uninstall the bundle.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.StringArrayVarP(&opts.ParameterSets, "parameter-set", "p", nil,
		"Name of a parameter set file for the bundle. May be either a named set of parameters or a filepath, and specified multiple times.")
	f.StringArrayVar(&opts.Params, "param", nil,
		"Define an individual parameter in the form NAME=VALUE. Overrides parameters otherwise set via --parameter-set. May be specified multiple times.")
	f.StringArrayVarP(&opts.CredentialIdentifiers, "cred", "c", nil,
		"Credential to use when uninstalling the bundle. May be either a named set of credentials or a filepath, and specified multiple times.")
	f.StringVarP(&opts.Driver, "driver", "d", porter.DefaultDriver,
		"Specify a driver to use. Allowed values: docker, debug")
	f.BoolVar(&opts.Delete, "delete", false,
		"Delete all records associated with the installation, assuming the uninstall action succeeds")
	f.BoolVar(&opts.ForceDelete, "force-delete", false,
		"UNSAFE. Delete all records associated with the installation, even if uninstall fails. This is intended for cleaning up test data and is not recommended for production environments.")
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace of the specified installation. Defaults to the global namespace.")
	f.BoolVar(&opts.NoLogs, "no-logs", false,
		"Do not persist the bundle execution logs")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	// Allow configuring the --driver flag with runtime-driver, to avoid conflicts with other commands
	cmd.Flag("driver").Annotations = map[string][]string{
		"viper-key": {"runtime-driver"},
	}
	return cmd
}

func buildBundlePublishCommand(p *porter.Porter) *cobra.Command {

	opts := porter.PublishOptions{}
	cmd := cobra.Command{
		Use:   "publish",
		Short: "Publish a bundle",
		Long: `Publishes a bundle by pushing the invocation image and bundle to a registry.

Note: if overrides for registry/tag/reference are provided, this command only re-tags the invocation image and bundle; it does not re-build the bundle.`,
		Example: `  porter bundle publish
  porter bundle publish --file myapp/porter.yaml
  porter bundle publish --dir myapp
  porter bundle publish --archive /tmp/mybuns.tgz --reference myrepo/my-buns:0.1.0
  porter bundle publish --tag latest
  porter bundle publish --registry myregistry.com/myorg
		`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Publish(cmd.Context(), opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "", "Path to the Porter manifest. Defaults to `porter.yaml` in the current directory.")
	f.StringVarP(&opts.Dir, "dir", "d", "",
		"Path to the build context directory where all bundle assets are located.")
	f.StringVarP(&opts.ArchiveFile, "archive", "a", "", "Path to the bundle archive in .tgz format")
	f.StringVar(&opts.Tag, "tag", "", "Override the Docker tag portion of the bundle reference, e.g. latest, v0.1.1")
	f.StringVar(&opts.Registry, "registry", "", "Override the registry portion of the bundle reference, e.g. docker.io, myregistry.com/myorg")
	addReferenceFlag(f, &opts.BundlePullOptions)
	addInsecureRegistryFlag(f, &opts.BundlePullOptions)
	// We aren't using addBundlePullFlags because we don't use --force since we are pushing, and that flag isn't needed

	return &cmd
}

func buildBundleArchiveCommand(p *porter.Porter) *cobra.Command {

	opts := porter.ArchiveOptions{}
	cmd := cobra.Command{
		Use:   "archive FILENAME --reference PUBLISHED_BUNDLE",
		Short: "Archive a bundle from a reference",
		Long:  "Archives a bundle by generating a gzipped tar archive containing the bundle, invocation image and any referenced images.",
		Example: `  porter bundle archive mybun.tgz --reference ghcr.io/getporter/examples/porter-hello:v0.2.0
  porter bundle archive mybun.tgz --reference localhost:5000/ghcr.io/getporter/examples/porter-hello:v0.2.0 --force
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(cmd.Context(), args, p)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Archive(cmd.Context(), opts)
		},
	}

	addBundlePullFlags(cmd.Flags(), &opts.BundlePullOptions)

	return &cmd
}
