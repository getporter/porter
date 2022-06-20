package main

import (
	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildInstallationCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "installations",
		Aliases: []string{"inst", "installation"},
		Short:   "Installation commands",
		Long:    "Commands for working with installations of a bundle",
	}
	cli.SetCommandGroup(cmd, "resource")

	cmd.AddCommand(buildInstallationsListCommand(p))
	cmd.AddCommand(buildInstallationShowCommand(p))
	cmd.AddCommand(buildInstallationApplyCommand(p))
	cmd.AddCommand(buildInstallationOutputsCommands(p))
	cmd.AddCommand(buildInstallationDeleteCommand(p))
	cmd.AddCommand(buildInstallationLogCommands(p))
	cmd.AddCommand(buildInstallationRunsCommands(p))

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
