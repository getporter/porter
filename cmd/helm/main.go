package main

import (
	"fmt"
	"os"

	"github.com/deislabs/porter/pkg/mixin/helm"

	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	m := &helm.Mixin{}
	cmd := &cobra.Command{
		Use:  "helm",
		Long: "A helm mixin for porter 👩🏽‍✈️",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			m.Out = cmd.OutOrStdout()
			m.Err = cmd.OutOrStderr()
		},
	}

	cmd.AddCommand(buildVersionCommand(m))
	cmd.AddCommand(buildBuildCommand(m))
	cmd.AddCommand(buildInstallCommand(m))
	return cmd
}
