package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildParametersCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "parameters",
		Aliases:     []string{"parameter", "param", "params"},
		Annotations: map[string]string{"group": "resource"},
		Short:       "Parameters commands",
	}

	cmd.AddCommand(buildParametersEditCommand(p))
	cmd.AddCommand(buildParametersGenerateCommand(p))
	cmd.AddCommand(buildParametersListCommand(p))
	cmd.AddCommand(buildParametersDeleteCommand(p))
	cmd.AddCommand(buildParametersShowCommand(p))

	return cmd
}

func buildParametersEditCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ParameterEditOptions{}

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit Parameter",
		Long:  `Edit a named parameter set.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.EditParameter(opts)
		},
	}
	return cmd
}

func buildParametersGenerateCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ParameterOptions{}
	cmd := &cobra.Command{
		Use:   "generate [NAME]",
		Short: "Generate Parameter Set",
		Long: `Generate a named set of parameters.

The first argument is the name of parameter set you wish to generate. If not
provided, this will default to the bundle name. By default, Porter will
generate a parameter set for the bundle in the current directory. You may also
specify a bundle with --file.

Bundles define 1 or more parameter(s) that are required to interact with a
bundle. The bundle definition defines where the parameter should be delivered
to the bundle, i.e. via DB_USERNAME. A parameter set, on the other hand,
represents the source data that you wish to use when interacting with the
bundle. These will typically be environment variables or files on your local
file system.

When you wish to install, upgrade or delete a bundle, Porter will use the
parameter set to determine where to read the necessary information from and
will then provide it to the bundle in the correct location. `,
		Example: `  porter parameter generate
  porter parameter generate myparamset --file myapp/porter.yaml
  porter parameter generate myparamset --tag getporter/porter-hello:v0.1.0
  porter parameter generate myparamset --cnab-file myapp/bundle.json --dry-run
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.GenerateParameters(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	f.BoolVar(&opts.DryRun, "dry-run", false,
		"Generate parameter but do not save it.")
	f.StringVar(&opts.Tag, "tag", "",
		"Use a bundle in an OCI registry specified by the given tag.")
	f.BoolVar(&opts.Force, "force", false,
		"Force a fresh pull of the bundle")

	return cmd
}

func buildParametersListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List parameters",
		Long:    `List named sets of parameters defined by the user.`,
		Example: `  porter parameters list [-o table|json|yaml]`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.ParseFormat()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ListParameters(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}

func buildParametersDeleteCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ParameterDeleteOptions{}

	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a Parameter",
		Long:  `Delete a named parameter set.`,
		PreRunE: func(_ *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return p.DeleteParameter(opts)
		},
	}
}

func buildParametersShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ParameterShowOptions{}

	cmd := &cobra.Command{
		Use:     "show",
		Short:   "Show a Parameter",
		Long:    `Show a particular parameter set, including all named parameters and their corresponding mappings.`,
		Example: `  porter parameter show NAME [-o table|json|yaml]`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowParameter(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}
