package main

import (
	"github.com/deislabs/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildUninstallCommand(m *exec.Mixin) *cobra.Command {
	opts := exec.UninstallOptions{}

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Execute the uninstall functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Uninstall(opts.File)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.File, "file", "f", "", "Path to the script to execute")
	return cmd
}
