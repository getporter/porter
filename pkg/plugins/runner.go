package plugins

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"github.com/pkg/errors"
)

type CommandOptions struct {
	Command string
}

type PluginRunner struct {
	*context.Context
	pluginName string
}

func NewRunner(pluginName string) *PluginRunner {
	return &PluginRunner{
		Context:    context.New(),
		pluginName: pluginName,
	}
}

func (r *PluginRunner) Validate() error {
	if r.pluginName == "" {
		return errors.New("Plugin not specified")
	}

	pluginPath, err := config.New().GetPluginPath(r.pluginName)
	if err != nil {
		return errors.Wrapf(err, "Failed to get plugin path for %s", r.pluginName)
	}

	exists, err := r.FileSystem.Exists(pluginPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to stat path %s", pluginPath)
	}
	if !exists {
		return errors.Errorf("Plugin %s doesn't exist in filesystem with path %s", r.pluginName, pluginPath)
	}

	return nil
}

func (r *PluginRunner) Run(commandOpts CommandOptions) error {
	if r.Debug {
		fmt.Fprintf(r.Err, "DEBUG Plugin:    %s\n", r.pluginName)
		fmt.Fprintf(r.Err, "DEBUG Command:     %s\n", commandOpts.Command)
	}

	pluginPath, err := config.New().GetPluginPath(r.pluginName)
	if r.Debug {
		fmt.Fprintf(r.Err, "DEBUG PluginPath:    %s\n", pluginPath)
	}
	if err != nil {
		return errors.Wrapf(err, "Failed to get plugin path for %s", r.pluginName)
	}

	cmdArgs := strings.Split(commandOpts.Command, " ")
	cmd := r.NewCommand(pluginPath, cmdArgs...)

	// Pipe the output from the mixin to porter
	cmd.Stdout = r.Context.Out
	cmd.Stderr = r.Context.Err

	prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
	if r.Debug {
		fmt.Fprintln(r.Err, prettyCmd)
	}

	err = cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "could not run plugin command %s", prettyCmd)
	}

	return cmd.Wait()
}
