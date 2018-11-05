package main

import (
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the application version",
		Run: func(cmd *cobra.Command, args []string) {
			porter.PrintVersion(cmd.OutOrStdout())
		},
	}
}
