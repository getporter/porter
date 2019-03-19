package main

import (
	"github.com/deislabs/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func buildVersionCommand(mixin *kubernetes.Mixin) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the mixin verison",
		Run: func(cmd *cobra.Command, args []string) {
			mixin.PrintVersion()
		},
	}
}
