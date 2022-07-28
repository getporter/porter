package main

import (
	"fmt"

	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildBuildCommand(m *exec.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Generate Dockerfile lines for the bundle invocation image",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(m.Config.Out, "# exec mixin has no buildtime dependencies")
		},
	}
	return cmd
}
