package main

import (
	"get.porter.sh/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func buildSchemaCommand(mixin *kubernetes.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the json schema for the mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mixin.PrintSchema()
		},
	}
	return cmd
}
