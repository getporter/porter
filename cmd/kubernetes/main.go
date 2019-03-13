package main

import (
	"fmt"
	"io"
	"os"

	"github.com/deislabs/porter/pkg/kubernetes"
	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand(os.Stdin)
	if err := cmd.Execute(); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}

func buildRootCommand(in io.Reader) *cobra.Command {
	mixin := kubernetes.New()
	mixin.In = in
	cmd := &cobra.Command{
		Use:  "kubernetes",
		Long: "kuberetes is a porter 👩🏽‍✈️ mixin that you can you can use to leverage kubernetes manifests",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			mixin.Out = cmd.OutOrStdout()
			mixin.Err = cmd.OutOrStderr()
		},
	}

	cmd.PersistentFlags().BoolVar(&mixin.Debug, "debug", false, "Enable debug logging")
	cmd.AddCommand(buildVersionCommand(mixin))
	cmd.AddCommand(buildBuildCommand(mixin))
	cmd.AddCommand(buildInstallCommand(mixin))
	cmd.AddCommand(buildUpgradeCommand(mixin))
	cmd.AddCommand(buildUnInstallCommand(mixin))
	cmd.AddCommand(buildSchemaCommand(mixin))
	return cmd
}
