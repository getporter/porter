package main

import (
	"github.com/deislabs/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildUpgradeCommand(m *exec.Mixin) *cobra.Command {
	opts := exec.UpgradeOptions{}

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Execute the upgrade functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Upgrade(opts.File)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.File, "file", "f", "", "Path to the script to execute")
	return cmd
}
