package main

import (
	"get.porter.sh/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func buildInvokeCommand(mixin *kubernetes.Mixin) *cobra.Command {
	return &cobra.Command{
		Use:   "invoke",
		Short: "Use kubectl to apply manifests to a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mixin.Execute()
		},
	}
}
