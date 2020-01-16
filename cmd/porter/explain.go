package main

import (
	"github.com/spf13/cobra"

	"get.porter.sh/porter/pkg/porter"
)

func buildBundleExplainCommand(p *porter.Porter) *cobra.Command {

	opts := porter.ExplainOpts{}
	cmd := cobra.Command{
		Use:   "explain",
		Short: "Explain a bundle",
		Long:  "Explain how to use a bundle by printing the parameters, credentials, outputs, actions.",
		Example: `  porter bundle explain
  porter bundle explain --file another/porter.yaml
  porter bundle explain --cnab-file some/bundle.json
  porter bundle explain --tag getporter/porter-hello:v0.1.0
		  `,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Explain(opts)
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
