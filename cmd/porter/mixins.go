package main

import (
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/mixin/feed"
	"github.com/deislabs/porter/pkg/porter"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/spf13/cobra"
)

func buildMixinsCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mixins",
		Aliases: []string{"mixin"},
		Short:   "Mixin commands",
		Annotations: map[string]string{
			"group": "resource",
		},
	}

	cmd.AddCommand(buildMixinsListCommand(p))
	cmd.AddCommand(BuildMixinInstallCommand(p))
	cmd.AddCommand(buildMixinsFeedCommand(p))

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
		Use:   "install NAME",
		Short: "Install a mixin",
		Example: `  porter mixin install helm --url https://deislabs.blob.core.windows.net/porter/mixins/helm
  porter mixin install azure --version v0.4.0-ralpha.1+dubonnet --url https://deislabs.blob.core.windows.net/porter/mixins/azure
  porter mixin install kubernetes --version canary --url https://deislabs.blob.core.windows.net/porter/mixins/kubernetes`,
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

func buildMixinsFeedCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "feed",
		Aliases: []string{"feeds"},
		Short:   "Feed commands",
		Annotations: map[string]string{
			"group": "resource",
		},
	}

	cmd.AddCommand(BuildMixinFeedGenerateCommand(p))

	return cmd
}

func BuildMixinFeedGenerateCommand(p *porter.Porter) *cobra.Command {
	opts := feed.GenerateOptions{}
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate an atom feed from the mixins in a directory",
		Long: `Generate an atom feed from the mixins in a directory. 

A template is required, providing values for text properties such as the author name, base URLs and other values that cannot be inferred from the mixin file names. You can make a default template by running 'porter mixins feed template'.

The file names of the mixins must follow the naming conventions required of published mixins:

VERSION/MIXIN-GOOS-GOARCH[FILE_EXT]

More than one mixin may be present in the directory, and the directories may be nested a few levels dep, as long as the file path ends with the above naming convention, porter will find and match it. Below is an example directory structure that porter can list to generate a feed:

bin/
└── v1.2.3/
    ├── mymixin-darwin-amd64
    ├── mymixin-linux-amd64
    └── mymixin-windows-amd64.exe

See https://porter.sh/mixin-distribution more details.
`,
		Example: `  porter mixin feed generate
  porter mixin feed generate --dir bin --file bin/atom.xml --template porter-atom-template.xml`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.GenerateMixinFeed(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.SearchDirectory, "dir", "d", "",
		"The directory to search for mixin versions to publish in the feed. Defaults to the current directory.")
	cmd.Flags().StringVarP(&opts.AtomFile, "file", "f", "atom.xml",
		"The path of the atom feed output by this command.")
	cmd.Flags().StringVarP(&opts.TemplateFile, "template", "t", "atom-template.xml",
		"The template atom file used to populate the text fields in the generated feed.")

	return cmd
}
