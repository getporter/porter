package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildBuildCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Generate CNAB invocation image and bundle file",
		Long:  "Generates a Dockerfile and a bundle.json in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Build()
		},
	}
}
