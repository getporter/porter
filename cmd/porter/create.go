package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildCreateCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Generate porter configuration files",
		Long:  "Generates porter configuration files, porter.yaml and the CNAB run script, in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Create()
		},
	}
}
