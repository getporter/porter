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

	pluginPath := config.New().GetPluginPath(r.pluginName)
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
		fmt.Fprintln(r.Err, "DEBUG Plugin Name: ", r.pluginName)
		fmt.Fprintln(r.Err, "DEBUG Plugin Command: ", commandOpts.Command)
	}

	pluginPath := config.New().GetPluginPath(r.pluginName)
	if r.Debug {
		fmt.Fprintln(r.Err, "DEBUG Plugin Path: ", pluginPath)
	}

	cmdArgs := strings.Split(commandOpts.Command, " ")
	cmd := r.NewCommand(pluginPath, cmdArgs...)

	// Pipe the output from the plugin to porter
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err

	prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
	if r.Debug {
		fmt.Fprintln(r.Err, "DEBUG Plugin Full Command: ", prettyCmd)
	}

	err := cmd.Start()
	if err != nil {
		return errors.Wrapf(err, "could not run plugin command %s", prettyCmd)
	}

	return cmd.Wait()
}
