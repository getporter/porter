package plugins

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

type CommandOptions struct {
	Command string
}

type PluginRunner struct {
	*portercontext.Context
	pluginName string
}

func NewRunner(pluginName string) *PluginRunner {
	return &PluginRunner{
		Context:    portercontext.New(),
		pluginName: pluginName,
	}
}

func (r *PluginRunner) Validate() error {
	if r.pluginName == "" {
		return errors.New("Plugin not specified")
	}

	pluginPath, err := config.New().GetPluginPath(r.pluginName)
	if err != nil {
		return fmt.Errorf("Failed to get plugin path for %s: %w", r.pluginName, err)
	}

	exists, err := r.FileSystem.Exists(pluginPath)
	if err != nil {
		return fmt.Errorf("Failed to stat path %s: %w", pluginPath, err)
	}
	if !exists {
		return fmt.Errorf("Plugin %s doesn't exist in filesystem with path %s", r.pluginName, pluginPath)
	}

	return nil
}

func (r *PluginRunner) Run(ctx context.Context, commandOpts CommandOptions) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("name", r.pluginName),
		attribute.String("partial-command", commandOpts.Command),
	)
	defer span.EndSpan()

	pluginPath, err := config.New().GetPluginPath(r.pluginName)
	if err != nil {
		return span.Error(fmt.Errorf("Failed to get plugin path for %s: %w", r.pluginName, err))
	}
	span.SetAttributes(attribute.String("plugin-path", pluginPath))

	cmdArgs := strings.Split(commandOpts.Command, " ")
	cmd := r.NewCommand(ctx, pluginPath, cmdArgs...)

	// Pipe the output from the plugin to porter
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err

	prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
	span.SetAttributes(attribute.String("full-command", prettyCmd))

	err = cmd.Start()
	if err != nil {
		return span.Error(fmt.Errorf("could not run plugin command %s: %w", prettyCmd, err))
	}

	return span.Error(cmd.Wait())
}
