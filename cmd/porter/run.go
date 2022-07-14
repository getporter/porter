package main

import (
	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildRunCommand(p *porter.Porter) *cobra.Command {
	opts := porter.NewRunOptions(p.Config)
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute runtime bundle instructions",
		Long:  "Execute the runtime bundle instructions contained in a porter configuration file",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Run(cmd.Context(), opts)
		},
		Hidden: true, // Hide runtime commands from the helptext
	}
	cli.SetCommandGroup(cmd, "runtime")

	cmd.Flags().StringVarP(&opts.File, "file", "f", "porter.yaml", "The porter configuration file (Defaults to porter.yaml)")
	cmd.Flags().StringVar(&opts.Action, "action", "", "The bundle action to execute (Defaults to CNAB_ACTION)")

	return cmd
}
