package main

import (
	"github.com/deislabs/porter/pkg/mixin"
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
	cmd.AddCommand(BuildMixinInstallCommand(p))

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

	cmd.Flags().StringVarP(&opts.rawFormat, "output", "o", "table",
		"Output format, allowed values are: table, json")

	return cmd
}

func BuildMixinInstallCommand(p *porter.Porter) *cobra.Command {
	opts := mixin.InstallOptions{}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a mixin",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallMixin(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Version, "version", "v", "latest",
		"The mixin version. This can either be a version number, or a tagged release like 'latest' or 'canary'")
	cmd.Flags().StringVar(&opts.URL, "url", "",
		"URL from where the mixin can be downloaded, for example https://github.com/org/proj/releases/downloads")

	return cmd
}
