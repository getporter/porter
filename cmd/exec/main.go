package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"get.porter.sh/porter/pkg/cli"
	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	run := func() int {
		ctx := context.Background()
		m := exec.New()
		if err := m.Config.ConfigureLogging(ctx); err != nil {
			fmt.Println(err)
			os.Exit(cli.ExitCodeErr)
		}
		cmd := buildRootCommand(m, os.Stdin)

		// We don't have tracing working inside a bundle working currently.
		// We are using StartRootSpan anyway because it creates a TraceLogger and sets it
		// on the context, so we can grab it later
		ctx, log := m.Config.StartRootSpan(ctx, "exec")
		defer func() {
			// Capture panics and trace them
			if panicErr := recover(); panicErr != nil {
				log.Error(fmt.Errorf("%s", panicErr),
					attribute.Bool("panic", true),
					attribute.String("stackTrace", string(debug.Stack())))
				log.EndSpan()
				m.Close()
				os.Exit(cli.ExitCodeErr)
			} else {
				log.Close()
				m.Close()
			}
		}()

		if err := cmd.ExecuteContext(ctx); err != nil {
			return cli.ExitCodeErr
		}
		return cli.ExitCodeSuccess
	}
	os.Exit(run())
}

func buildRootCommand(m *exec.Mixin, in io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "exec",
		Long: "exec is a porter 👩🏽‍✈️ mixin that you can you can use to execute arbitrary commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr/stdin for testing
			m.Config.In = in
			m.Config.Out = cmd.OutOrStdout()
			m.Config.Err = cmd.OutOrStderr()
		},
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(&m.Debug, "debug", false, "Enable debug mode")

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
