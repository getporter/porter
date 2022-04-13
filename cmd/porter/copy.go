package main

import (
	"github.com/spf13/cobra"

	"get.porter.sh/porter/pkg/porter"
)

func buildBundleCopyCommand(p *porter.Porter) *cobra.Command {

	opts := &porter.CopyOpts{}

	cmd := cobra.Command{
		Use:   "copy",
		Short: "Copy a bundle",
		Long: `Copy a published bundle from one registry to another.		
Source bundle can be either a tagged reference or a digest reference.
Destination can be either a registry, a registry/repository, or a fully tagged bundle reference. 
If the source bundle is a digest reference, destination must be a tagged reference.
`,
		Example: `  porter bundle copy
  porter bundle copy --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination portersh
  porter bundle copy --source ghcr.io/getporter/examples/porter-hello:v0.2.0 --destination portersh --insecure-registry
		  `,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.CopyBundle(opts)
		},
	}
	f := cmd.Flags()
	f.StringVarP(&opts.Source, "source", "", "", " The fully qualified source bundle, including tag or digest.")
	f.StringVarP(&opts.Destination, "destination", "", "", "The registry to copy the bundle to. Can be registry name, registry plus a repo prefix, or a new tagged reference. All images and the bundle will be prefixed with registry.")
	f.BoolVar(&opts.InsecureRegistry, "insecure-registry", false, "Don't require TLS for registries")
	return &cmd
}
