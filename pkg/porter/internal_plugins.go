package porter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/storage/plugins/noop"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/portercontext"
	secretsplugins "get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/filesystem"
	"get.porter.sh/porter/pkg/secrets/plugins/host"
	storageplugins "get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/hashicorp/go-plugin"
)

type RunInternalPluginOpts struct {
	Key string
}

func (o *RunInternalPluginOpts) ApplyArgs(args []string) error {
	if len(args) == 0 {
		return errors.New("The positional argument KEY was not specified")
	}
	if len(args) > 1 {
		return errors.New("Multiple positional arguments were specified but only one KEY is expected")
	}

	o.Key = args[0]
	return nil
}

func (o *RunInternalPluginOpts) Validate() error {
	if o.Key == "" {
		return fmt.Errorf("no plugin key was specified")
	}

	if _, ok := internalPlugins[o.Key]; !ok {
		return fmt.Errorf("invalid plugin key specified: %s", o.Key)
	}

	return nil
}

func (o *RunInternalPluginOpts) parsePluginConfig(c *portercontext.Context) (interface{}, error) {
	pluginCfg := map[string]interface{}{}
	if err := json.NewDecoder(c.In).Decode(&pluginCfg); err != nil {
		if err == io.EOF {
			// No plugin config was specified
			return pluginCfg, nil
		}
		return nil, fmt.Errorf("error parsing plugin configuration from stdin as json: %w", err)
	}
	return pluginCfg, nil
}

func (p *Porter) RunInternalPlugins(ctx context.Context, opts RunInternalPluginOpts) (err error) {
	err = opts.Validate()
	if err != nil {
		return err
	}

	// Read the plugin configuration from the porter config file from STDIN
	pluginCfg, err := opts.parsePluginConfig(p.Context)
	if err != nil {
		return err
	}

	// Create an instance of the plugin
	selectedPlugin := internalPlugins[opts.Key]
	impl, err := selectedPlugin.Create(p.Config, pluginCfg)
	if err != nil {
		return fmt.Errorf("could not create an instance of the requested internal plugin %s: %w", opts.Key, err)
	}

	defer func() {
		if panicErr := recover(); err != nil {
			err = fmt.Errorf("%v", panicErr)
		}

		if closer, ok := impl.(closablePlugin); ok {
			if err = closer.Close(ctx); err != nil {
				log := tracing.LoggerFromContext(ctx)
				log.Error(fmt.Errorf("error stopping the %s plugin: %w", opts.Key, err))
			}
		}
	}()

	plugins.Serve(p.Context, selectedPlugin.Interface, impl, selectedPlugin.ProtocolVersion)
	return err // Return the error that may have been set during recover above
}

// A list of available plugins that we can serve directly from the porter binary
var internalPlugins map[string]InternalPlugin = getInternalPlugins()

// InternalPlugin represents the information needed to run one of the plugins
// defined in porter's repository.
type InternalPlugin struct {
	Interface       string
	ProtocolVersion int
	Create          func(c *config.Config, pluginCfg interface{}) (plugin.Plugin, error)
}

// A long running plugin needs to setup a connection or other resources first,
// and will hold those resources until porter is done with the plugin.
type closablePlugin interface {
	Close(ctx context.Context) error
}

func getInternalPlugins() map[string]InternalPlugin {
	return map[string]InternalPlugin{
		host.PluginKey: {
			Interface:       secretsplugins.PluginInterface,
			ProtocolVersion: secretsplugins.PluginProtocolVersion,
			Create: func(c *config.Config, pluginCfg interface{}) (plugin.Plugin, error) {
				return host.NewPlugin(c.Context), nil
			},
		},
		filesystem.PluginKey: {
			Interface:       secretsplugins.PluginInterface,
			ProtocolVersion: secretsplugins.PluginProtocolVersion,
			Create: func(c *config.Config, pluginCfg interface{}) (plugin.Plugin, error) {
				return filesystem.NewPlugin(c, pluginCfg), nil
			},
		},
		noop.PluginKey: {
			Interface:       storageplugins.PluginInterface,
			ProtocolVersion: storageplugins.PluginProtocolVersion,
			Create: func(c *config.Config, pluginCfg interface{}) (plugin.Plugin, error) {
				return noop.NewPlugin(c.Context)
			},
		},
		mongodb.PluginKey: {
			Interface:       storageplugins.PluginInterface,
			ProtocolVersion: storageplugins.PluginProtocolVersion,
			Create: func(c *config.Config, pluginCfg interface{}) (plugin.Plugin, error) {
				return mongodb.NewPlugin(c.Context, pluginCfg)
			},
		},
		mongodb_docker.PluginKey: {
			Interface:       storageplugins.PluginInterface,
			ProtocolVersion: storageplugins.PluginProtocolVersion,
			Create: func(c *config.Config, pluginCfg interface{}) (plugin.Plugin, error) {
				return mongodb_docker.NewPlugin(c.Context, pluginCfg)
			},
		},
	}
}
