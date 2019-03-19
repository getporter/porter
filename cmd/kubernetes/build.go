package main

import (
	"github.com/deislabs/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func buildBuildCommand(mixin *kubernetes.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Generate Dockerfile contribution for invocation image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mixin.Build()
		},
	}
	return cmd
}
