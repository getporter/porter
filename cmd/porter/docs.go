package main

import (
	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/docs"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildDocsCommand(p *porter.Porter) *cobra.Command {
	opts := &docs.DocsOptions{}

	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate markdown docs",
		Long:   "Generate markdown docs for https://getporter.org/cli",
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.RootCommand = cmd.Root()
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return docs.GenerateCliDocs(opts)
		},
	}

	cli.SkipConfigForCommand(cmd)
	cli.SetCommandGroup(cmd, "meta")

	flags := cmd.Flags()
	flags.StringVarP(&opts.Destination, "dest", "d", docs.DefaultDestination,
		"Destination directory")

	return cmd
}
