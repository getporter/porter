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
		Short:       "Parameter set commands",
	}

	cmd.AddCommand(buildParametersApplyCommand(p))
	cmd.AddCommand(buildParametersEditCommand(p))
	cmd.AddCommand(buildParametersGenerateCommand(p))
	cmd.AddCommand(buildParametersListCommand(p))
	cmd.AddCommand(buildParametersDeleteCommand(p))
	cmd.AddCommand(buildParametersShowCommand(p))

	return cmd
}

func buildParametersApplyCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ApplyOptions{}

	cmd := &cobra.Command{
		Use:   "apply FILE",
		Short: "Apply changes to a parameter set",
		Long: `Apply changes from the specified file to a parameter set. If the parameter set doesn't already exist, it is created.

Supported file extensions: json and yaml.

You can use the generate and show commands to create the initial file:
  porter parameters generate myparams --reference SOME_BUNDLE
  porter parameters show myparams --output yaml > myparams.yaml
`,
		Example: `  porter parameters apply --file myparams.yaml`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ParametersApply(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the parameter set is defined. The namespace in the file, if set, takes precedence.")

	return cmd
}

func buildParametersEditCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ParameterEditOptions{}

	cmd := &cobra.Command{
		Use:     "edit",
		Short:   "Edit Parameter Set",
		Long:    `Edit a named parameter set.`,
		Example: `  porter parameter edit debug-tweaks --namespace dev`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.EditParameter(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the parameter set is defined. Defaults to the global namespace.")

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
  porter parameter generate myparamset --reference getporter/hello-llama:v0.1.1 --namespace dev
  porter parameter generate myparamset --label owner=myname --reference getporter/hello-llama:v0.1.1
  porter parameter generate myparamset --reference localhost:5000/getporter/hello-llama:v0.1.1 --insecure-registry --force
  porter parameter generate myparamset --file myapp/porter.yaml
  porter parameter generate myparamset --cnab-file myapp/bundle.json
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.GenerateParameters(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the parameter set is defined. Defaults to the global namespace.")
	f.StringSliceVarP(&opts.Labels, "label", "l", nil,
		"Associate the specified labels with the parameter set. May be specified multiple times.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	return cmd
}

func buildParametersListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List parameter sets",
		Long: `List named sets of parameters defined by the user.

Optionally filters the results name, which returns all results whose name contain the provided query.
The results may also be filtered by associated labels and the namespace in which the parameter set is defined.`,
		Example: `  porter parameters list
  porter parameters list --namespace prod -o json
  porter parameters list --all-namespaces,
  porter parameters list --name myapp
  porter parameters list --label env=dev`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintParameters(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the parameter set is defined. Defaults to the global namespace. Use * to list across all namespaces.")
	f.BoolVar(&opts.AllNamespaces, "all-namespaces", false,
		"Include all namespaces in the results.")
	f.StringVar(&opts.Name, "name", "",
		"Filter the parameter sets where the name contains the specified substring.")
	f.StringSliceVarP(&opts.Labels, "label", "l", nil,
		"Filter the parameter sets by a label formatted as: KEY=VALUE. May be specified multiple times.")
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}

func buildParametersDeleteCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ParameterDeleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a Parameter Set",
		Long:  `Delete a named parameter set.`,
		PreRunE: func(_ *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return p.DeleteParameter(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the parameter set is defined. Defaults to the global namespace.")

	return cmd
}

func buildParametersShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ParameterShowOptions{}

	cmd := &cobra.Command{
		Use:     "show",
		Short:   "Show a Parameter Set",
		Long:    `Show a named parameter set, including all named parameters and their corresponding mappings.`,
		Example: `  porter parameter show NAME [-o table|json|yaml]`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowParameter(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the parameter set is defined. Defaults to the global namespace.")
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}
