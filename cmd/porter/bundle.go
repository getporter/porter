package main

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildBundleCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bundles",
		Aliases: []string{"bundle"},
		Short:   "Bundle commands",
		Long:    "Commands for working with bundles. These all have shortcuts so that you can call these commands without the bundle resource prefix. For example, porter bundle build is available as porter build as well.",
	}
	cmd.Annotations = map[string]string{
		"group": "resource",
	}

	cmd.AddCommand(buildBundleCreateCommand(p))
	cmd.AddCommand(buildBundleBuildCommand(p))
	cmd.AddCommand(buildBundleLintCommand(p))
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
			return opts.Validate(p.Config)
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
	f.BoolVar(&opts.Force, "force", false, "Force push the bundle to overwrite the previously published bundle")
	// Allow configuring the --force flag with "force-overwrite" in the configuration file
	cmd.Flag("force").Annotations = map[string][]string{
		"viper-key": {"force-overwrite"},
	}

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
