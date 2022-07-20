package main

import (
	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildUninstallCommand(m *exec.Mixin) *cobra.Command {
	opts := exec.ExecuteOptions{}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Execute the uninstall functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Execute(cmd.Context(), opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.File, "file", "f", "", "Path to the script to execute")
	return cmd
}
