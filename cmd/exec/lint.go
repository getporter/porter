package main

import (
	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildLintCommand(m *exec.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Check sections of the bundle associated with this mixin for problems and adherence to best practices",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.PrintLintResults(cmd.Context())
		},
	}
	return cmd
}
