package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/spf13/cobra"
)

func buildMixinsCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mixins",
		Aliases: []string{"mixin"},
		Short:   "Mixin commands",
	}
	cmd.Annotations = map[string]string{
		"group": "resource",
	}

	cmd.AddCommand(buildMixinsListCommand(p))

	return cmd
}

func buildMixinsListCommand(p *porter.Porter) *cobra.Command {
	opts := struct {
		rawFormat string
		format    printer.Format
	}{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed mixins",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			opts.format, err = printer.ParseFormat(opts.rawFormat)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintMixins(printer.PrintOptions{Format: opts.format})
		},
	}

	cmd.Flags().StringVarP(&opts.rawFormat, "output", "o", "table", "Output format, allowed values are: table, json")
	return cmd
}
