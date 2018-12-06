package main

import (
	"github.com/deislabs/porter/pkg/mixin/azure"
	"github.com/spf13/cobra"
)

func buildUninstallCommand(m *azure.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Execute the uninstall functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Uninstall()
		},
	}
	return cmd
}
