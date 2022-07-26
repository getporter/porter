package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildBundleInspectCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ExplainOpts{}
	cmd := cobra.Command{
		Use:   "inspect REFERENCE",
		Short: "Inspect a bundle",
		Long: `Inspect a bundle by printing the invocation images and any related images images.

If you would like more information about the bundle, the porter explain command will provide additional information,
like parameters, credentials, outputs and custom actions available.
`,
		Example: `  porter bundle inspect
  porter bundle inspect ghcr.io/getporter/examples/porter-hello:v0.2.0
  porter bundle inspect localhost:5000/ghcr.io/getporter/examples/porter-hello:v0.2.0 --insecure-registry --force
  porter bundle inspect --file another/porter.yaml
  porter bundle inspect --cnab-file some/bundle.json
		  `,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Inspect(cmd.Context(), opts)
		},
	}
	f := cmd.Flags()
	f.StringVarP(&opts.File, "file", "f", "", "Path to the Porter manifest. Defaults to `porter.yaml` in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "", "Path to the CNAB bundle.json file.")
	f.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Specify an output format.  Allowed values: plaintext, json, yaml")
	addBundlePullFlags(f, &opts.BundlePullOptions)
	return &cmd
}
