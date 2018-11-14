package main

import (
	"fmt"

	"github.com/deislabs/porter/pkg/mixin/exec"
	"github.com/spf13/cobra"
)

func buildBuildCommand(e *exec.Exec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Generate Dockerfile lines for the bundle invocation image",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStderr(), "exec mixin doesn't need any help, thank you very much 😉")
		},
	}
	return cmd
}
