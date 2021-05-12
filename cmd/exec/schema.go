package main

import (
	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildSchemaCommand(m *exec.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the json schema for the mixin",
		Run: func(cmd *cobra.Command, args []string) {
			m.PrintSchema()
		},
	}
	return cmd
}
