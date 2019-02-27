package main

import (
	"github.com/deislabs/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildInstallCommand(m *exec.Mixin) *cobra.Command {
	var opts struct {
		file string
	}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Execute the install functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Install(opts.file)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.file, "file", "f", "", "Path to the script to execute")
	return cmd
}
