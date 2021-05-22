package main

import (
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/feed"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildMixinCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mixins",
		Aliases: []string{"mixin"},
		Short:   "Mixin commands. Mixins assist with authoring bundles.",
		Annotations: map[string]string{
			"group": "resource",
		},
	}

	cmd.AddCommand(buildMixinsListCommand(p))
	cmd.AddCommand(buildMixinsSearchCommand(p))
	cmd.AddCommand(BuildMixinInstallCommand(p))
	cmd.AddCommand(BuildMixinUninstallCommand(p))
	cmd.AddCommand(buildMixinsFeedCommand(p))
	cmd.AddCommand(buildMixinsCreateCommand(p))

	return cmd
}

func buildMixinsListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.PrintMixinsOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed mixins",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.ParseFormat()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintMixins(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.RawFormat, "output", "o", "table",
		"Output format, allowed values are: table, json, yaml")

	return cmd
}

func buildMixinsSearchCommand(p *porter.Porter) *cobra.Command {
	opts := porter.SearchOptions{
		Type: "mixin",
	}

	cmd := &cobra.Command{
		Use:   "search [QUERY]",
		Short: "Search available mixins",
		Long: `Search available mixins. You can specify an optional mixin name query, where the results are filtered by mixins whose name contains the query term.

By default the community mixin index at https://cdn.porter.sh/mixins/index.json is searched. To search from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.`,
		Example: `  porter mixin search
  porter mixin search helm
  porter mixin search -o json`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.SearchPackages(opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Output format, allowed values are: table, json, yaml")
	flags.StringVar(&opts.Mirror, "mirror", pkgmgmt.DefaultPackageMirror,
		"Mirror of official Porter assets")

	return cmd
}

func BuildMixinInstallCommand(p *porter.Porter) *cobra.Command {
	opts := mixin.InstallOptions{}
	cmd := &cobra.Command{
		Use:   "install NAME",
		Short: "Install a mixin",
		Long: `Install a mixin.

By default mixins are downloaded from the official Porter mixin feed at https://cdn.porter.sh/mixins/atom.xml. To download from a mirror, set the environment variable PORTER_MIRROR, or mirror in the Porter config file, with the value to replace https://cdn.porter.sh with.`,
		Example: `  porter mixin install helm --url https://cdn.porter.sh/mixins/helm
  porter mixin install helm --feed-url https://cdn.porter.sh/mixins/atom.xml
  porter mixin install azure --version v0.4.0-ralpha.1+dubonnet --url https://cdn.porter.sh/mixins/azure
  porter mixin install kubernetes --version canary --url https://cdn.porter.sh/mixins/kubernetes`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.InstallMixin(opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Version, "version", "v", "latest",
		"The mixin version. This can either be a version number, or a tagged release like 'latest' or 'canary'")
	flags.StringVar(&opts.URL, "url", "",
		"URL from where the mixin can be downloaded, for example https://github.com/org/proj/releases/downloads")
	flags.StringVar(&opts.FeedURL, "feed-url", "",
		"URL of an atom feed where the mixin can be downloaded. Defaults to the official Porter mixin feed.")
	flags.StringVar(&opts.Mirror, "mirror", pkgmgmt.DefaultPackageMirror,
		"Mirror of official Porter assets")
	return cmd
}

func BuildMixinUninstallCommand(p *porter.Porter) *cobra.Command {
	opts := pkgmgmt.UninstallOptions{}
	cmd := &cobra.Command{
		Use:     "uninstall NAME",
		Short:   "Uninstall a mixin",
		Example: `  porter mixin uninstall helm`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.UninstallMixin(opts)
		},
	}

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
	cmd.AddCommand(BuildMixinFeedTemplateCommand(p))

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

More than one mixin may be present in the directory, and the directories may be nested a few levels deep, as long as the file path ends with the above naming convention, porter will find and match it. Below is an example directory structure that porter can list to generate a feed:

bin/
└── v1.2.3/
    ├── mymixin-darwin-amd64
    ├── mymixin-linux-amd64
    └── mymixin-windows-amd64.exe

See https://porter.sh/mixin-dev-guide/distribution more details.
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

func BuildMixinFeedTemplateCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Create an atom feed template",
		Long:  "Create an atom feed template in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.CreateMixinFeedTemplate()
		},
	}
	return cmd
}

func buildMixinsCreateCommand(p *porter.Porter) *cobra.Command {
	opts := porter.MixinsCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create NAME --author YOURNAME [--dir /path/to/mixin/dir]",
		Short: "Create a new mixin project based on the getporter/skeletor repository",
		Long: `Create a new mixin project based on the getporter/skeletor repository.

The first argument is the name of the mixin to create and is required.
A flag of --author to declare the author of the mixin is also a required input.
You can also specify where to put mixin directory. It will default to the current directory.`,
		Example: ` porter mixin create MyMixin --author MyName
		porter mixin create MyMixin --author MyName --dir path/to/mymixin
		`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.CreateMixin(opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.AuthorName, "author", "", "Name of the mixin's author.")
	f.StringVar(&opts.DirPath, "dir", "", "Path to the designated location of the mixin's directory.")

	return cmd
}
