package main

import (
	"io"
	"os"

	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand(os.Stdin)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCommand(in io.Reader) *cobra.Command {
	m := exec.New()
	m.In = in
	cmd := &cobra.Command{
		Use:  "exec",
		Long: "exec is a porter 👩🏽‍✈️ mixin that you can you can use to execute arbitrary commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			m.Out = cmd.OutOrStdout()
			m.Err = cmd.OutOrStderr()
		},
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(&m.Debug, "debug", false, "Enable debug logging")

	cmd.AddCommand(buildVersionCommand(m))
	cmd.AddCommand(buildSchemaCommand(m))
	cmd.AddCommand(buildBuildCommand(m))
	cmd.AddCommand(buildLintCommand(m))
	cmd.AddCommand(buildInstallCommand(m))
	cmd.AddCommand(buildUpgradeCommand(m))
	cmd.AddCommand(buildInvokeCommand(m))
	cmd.AddCommand(buildUninstallCommand(m))

	return cmd
}
