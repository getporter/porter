package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// AnnotationSkipConfig indicates that config should not be loaded for this command.
	// This is used for commands like help and version which should never
	// fail, even with porter is mis-configured.
	AnnotationSkipConfig string = "skipConfig"

	// AnnotationIsPluginCommand indicates that a command is running a plugin.
	AnnotationIsPluginCommand string = "isPlugin"

	// AnnotationGroup specifies how the command should be grouped in the help text.
	AnnotationGroup string = "group"

	// ExitCodeSuccess indicates the program ran successfully.
	ExitCodeSuccess = 0

	// ExitCodeErr indicates the program encountered an error.
	ExitCodeErr = 1

	// ExitCodeInterrupt indicates the program was cancelled.
	ExitCodeInterrupt = 2
)

// Main implements your Porter application's entrypoint.
func Main(ctx context.Context, rootCmd *cobra.Command, app PorterApp) {
	ctx, cancel := HandleInterrupt(ctx)
	defer cancel()

	// Wrapping the main run logic in a function because os.Exit will not
	// execute defer statements
	os.Exit(RunCommand(ctx, rootCmd, app))
}

// RunCommand configures and runs the specified command against your app.
func RunCommand(ctx context.Context, rootCmd *cobra.Command, app PorterApp) int {
	config := app.GetConfig()

	// Configure Porter based on the command called
	calledCmd := ConfigureCommand(rootCmd, config)

	// Prepare the app to run only if the called command requires configuration/setup
	if !calledCmd.SkipConfig {
		if err := app.Connect(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(ExitCodeErr)
		}
	}

	// Trace the command
	ctx, log := config.Context.StartRootSpan(ctx, calledCmd.CommandPath, attribute.String("command", calledCmd.FormattedCommand))
	defer func() {
		// Capture panics and trace them
		if panicErr := recover(); panicErr != nil {
			log.Error(errors.New(fmt.Sprintf("%s", panicErr)),
				attribute.Bool("panic", true),
				attribute.String("stackTrace", string(debug.Stack())))
			log.EndSpan()
			app.Close()
			os.Exit(ExitCodeErr)
		} else {
			log.Close()
			app.Close()
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// Ideally we log all errors in the span that generated it,
		// but as a failsafe, always log the error at the root span as well
		log.Error(err)
		return 1
	}
	return 0
}

// HandleInterrupt tries to exit gracefully when the interrupt signal is sent (CTRL+C)
// Thanks to Mat Ryer, https://pace.dev/blog/2020/02/17/repond-to-ctrl-c-interrupt-signals-gracefully-with-context-in-golang-by-mat-ryer.html
func HandleInterrupt(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(ExitCodeInterrupt)
	}()

	return ctx, func() {
		signal.Stop(signalChan)
		cancel()
	}
}
