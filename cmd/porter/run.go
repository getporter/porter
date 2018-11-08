package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildRunCommand(p *porter.Porter) *cobra.Command {
	var opts struct {
		file string
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute runtime bundle instructions",
		Long:  "Execute the runtime bundle instructions contained in a porter configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Run(opts.file)
		},
	}

	cmd.Flags().StringVarP(&opts.file, "file", "f", "porter.yaml", "The porter configuration file (Defaults to porter.yaml)")

	return cmd
}
