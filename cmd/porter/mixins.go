package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildMixinCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mixins",
		Short: "Interact with the porter mixins",
	}

	cmd.AddCommand(buildMixinListCommand(p))
	return cmd
}

func buildMixinListCommand(p *porter.Porter) *cobra.Command {
	opts := struct {
		format string
	}{}
	cmd := &cobra.Command{
		Use: "list",
		Short: "List installed mixins",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintMixins(opts.format)
		},
	}

	cmd.Flags().StringVarP(&opts.format, "output", "o", "human", "Output format: human or json")
	return cmd
}