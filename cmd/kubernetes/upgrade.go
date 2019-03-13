package main

import (
	"github.com/deislabs/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func buildUpgradeCommand(mixin *kubernetes.Mixin) *cobra.Command {
	return &cobra.Command{
		Use:   "Upgrade",
		Short: "Use kubectl to apply manifests to cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mixin.Upgrade()
		},
	}
}
