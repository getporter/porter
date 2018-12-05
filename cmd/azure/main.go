package main

import (
	"fmt"
	"io"
	"os"

	"github.com/deislabs/porter/pkg/mixin/azure"

	"github.com/spf13/cobra"
)

func main() {
	cmd, err := buildRootCommand(os.Stdin)
	if err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}

func buildRootCommand(in io.Reader) (*cobra.Command, error) {
	m, err := azure.New()
	if err != nil {
		return nil, err
	}
	m.In = in
	cmd := &cobra.Command{
		Use:  "azure",
		Long: "An Azure mixin for porter 👩🏽‍✈️",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			m.Out = cmd.OutOrStdout()
			m.Err = cmd.OutOrStderr()
		},
	}

	cmd.PersistentFlags().BoolVar(&m.Debug, "debug", false, "Enable debug logging")

	cmd.AddCommand(buildVersionCommand(m))
	cmd.AddCommand(buildBuildCommand(m))
	cmd.AddCommand(buildInstallCommand(m))
	cmd.AddCommand(buildInstallCommand(m))

	return cmd, nil
}
