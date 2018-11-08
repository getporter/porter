package main

import (
	"github.com/deislabs/porter/pkg/mixin/exec"
	"github.com/spf13/cobra"
)

var (
	commandFile string
)

func buildInstallCommand(e *exec.Exec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Execute the install functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return e.Install(commandFile)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&commandFile, "file", "f", "", "file to install")
	return cmd
}
