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
  porter bundle explain --reference getporter/porter-hello:v0.1.0
  porter bundle explain --reference localhost:5000/getporter/porter-hello:v0.1.0 --insecure-registry --force
  porter bundle explain --file another/porter.yaml
  porter bundle explain --cnab-file some/bundle.json
  porter bundle explain --action install
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
	f.StringVarP(&opts.RawFormat, "output", "o", "plaintext",
		"Specify an output format.  Allowed values: plaintext, json, yaml")
	f.StringVar(&opts.Action, "action", "", "Hide parameters and outputs that are not used by the specified action.")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	return &cmd
}
