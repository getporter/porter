package main

import (
	"get.porter.sh/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func buildInstallCommand(mixin *kubernetes.Mixin) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Use kubectl to apply manifests to a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mixin.Install()
		},
	}
}
