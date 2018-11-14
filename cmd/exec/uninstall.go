package main

import (
	"github.com/deislabs/porter/pkg/mixin/exec"
	"github.com/spf13/cobra"
)

func buildUninstallCommand(e *exec.Exec) *cobra.Command {
	var opts struct {
		file string
	}
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Execute the uninstall functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return e.Uninstall(opts.file)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.file, "file", "f", "", "Path to the script to execute")
	return cmd
}
