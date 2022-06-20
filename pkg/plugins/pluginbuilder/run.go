package pluginbuilder

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/tracing"
)

// RunOptions are the arguments passed to the run command.
type RunOptions struct {
	// Key is the fully-qualified 3-part plugin key.
	Key string
}

// ApplyArgs applies the arguments from the command-line to the run command options.
func (o *RunOptions) ApplyArgs(args []string) error {
	if len(args) == 0 {
		return errors.New("the positional argument PLUGIN_KEY was not specified")
	}
	if len(args) > 1 {
		return errors.New("multiple positional arguments were specified but only one, PLUGIN_KEY, is expected")
	}

	o.Key = args[0]
	return nil
}

// Run executes the plugin.
func (p PorterPlugin) Run(ctx context.Context, opts RunOptions) error {
	err := p.validateRunOptions(opts)
	if err != nil {
		return err
	}

	// Read the plugin configuration from the porter pluginConfig file from stdin
	if err := p.loadConfig(); err != nil {
		return err
	}

	// Create an instance of the plugin
	selectedPlugin := p.opts.RegisteredPlugins[opts.Key]
	impl, err := selectedPlugin.Create(p.porterConfig, p.pluginConfig)
	if err != nil {
		return fmt.Errorf("could not create an instance of the requested internal plugin %s: %w", opts.Key, err)
	}

	// Clean up after the plugin when it is done
	defer func() {
		if panicErr := recover(); err != nil {
			err = fmt.Errorf("%v", panicErr)
		}

		if closer, ok := impl.(plugins.PluginCloser); ok {
			if err = closer.Close(ctx); err != nil {
				log := tracing.LoggerFromContext(ctx)
				log.Error(fmt.Errorf("error stopping the %s plugin: %w", opts.Key, err))
			}
		}
	}()

	// Run the plugin
	plugins.Serve(p.porterConfig.Context, selectedPlugin.Interface, impl, selectedPlugin.ProtocolVersion)

	// Return the error that may have been set during recover above
	return err
}

// validateRunOptions validates the arguments and flags passed to the run command.
func (p PorterPlugin) validateRunOptions(opts RunOptions) error {
	if opts.Key == "" {
		return fmt.Errorf("no plugin key was specified")
	}

	if _, ok := p.opts.RegisteredPlugins[opts.Key]; !ok {
		return fmt.Errorf("unsupported plugin key specified: %s: the plugin supports the following keys: %s", opts.Key, p.opts.ListSupportedKeys())
	}

	return nil
}
