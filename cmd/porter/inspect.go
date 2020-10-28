package main

import (
	"github.com/spf13/cobra"

	"get.porter.sh/porter/pkg/porter"
)

func buildBundleInspectCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ExplainOpts{}
	cmd := cobra.Command{
		Use:   "inspect",
		Short: "Inspect a bundle",
		Long: `Inspect a bundle by printing the invocation images and any related images images.

If you would like more information about the bundle, the porter explain command will provide additional information,
like parameters, credentials, outputs and custom actions available.
`,
		Example: `  porter bundle inspect
  porter bundle inspect --tag getporter/porter-hello:v0.1.0
  porter bundle inspect --tag localhost:5000/getporter/porter-hello:v0.1.0 --insecure-registry --force
  porter bundle inspect --file another/porter.yaml
  porter bundle inspect --cnab-file some/bundle.json
		  `,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Inspect(opts)
		},
	}
	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "", "Path to the Porter manifest. Defaults to `porter.yaml` in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "", "Path to the CNAB bundle.json file.")
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")
	addBundlePullFlags(f, &opts.BundlePullOptions)
	return &cmd
}
