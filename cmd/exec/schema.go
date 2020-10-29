package main

import (
	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildSchemaCommand(m *exec.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the json schema for the mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.PrintSchema()
		},
	}
	return cmd
}
