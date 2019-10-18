package main

import (
	"github.com/spf13/cobra"

	"github.com/deislabs/porter/pkg/porter"
)

func buildBundleInspectCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ExplainOpts{}
	cmd := cobra.Command{
		Use:   "inspect",
		Short: "Inspect a bundle",
		Long:  "Inspect a bundle by printing the parameters, credentials, outputs, actions and images.",
		Example: `  porter bundle inspect
  porter bundle inspect --file another/porter.yaml
  porter bundle inspect --cnab-file some/bundle.json
  porter bundle inspect --tag deislabs/porter-bundle:v0.1.0
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
	f.StringVarP(&opts.Tag, "tag", "t", "",
		"Use a bundle in an OCI registry specified by the given tag")
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")
	return &cmd
}
