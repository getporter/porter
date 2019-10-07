package main

import (
	"github.com/deislabs/porter/pkg/docs"
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildDocsCommand(p *porter.Porter) *cobra.Command {
	opts := &docs.DocsOptions{}

	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate markdown docs",
		Long:   "Generate markdown docs for https://porter.sh/cli",
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			opts.RootCommand = cmd.Root()
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return docs.GenerateCliDocs(opts)
		},
	}

	cmd.Annotations = map[string]string{
		"group": "meta",
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Destination, "dest", "d", docs.DefaultDestination,
		"Destination directory")

	return cmd
}
