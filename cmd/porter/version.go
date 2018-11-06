package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildVersionCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Run: func(cmd *cobra.Command, args []string) {
			p.PrintVersion()
		},
	}
}
