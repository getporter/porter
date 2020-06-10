package pluggable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

// PluginLoader handles finding, configuring and loading porter plugins.
type PluginLoader struct {
	*config.Config

	SelectedPluginKey    *plugins.PluginKey
	SelectedPluginConfig interface{}
}

func NewPluginLoader(c *config.Config) *PluginLoader {
	return &PluginLoader{
		Config: c,
	}
}

// Load a plugin, returning the plugin's interface which the caller must then cast to
// the typed interface, a cleanup function to stop the plugin when finished communicating with it,
// and an error if the plugin could not be loaded.
func (l *PluginLoader) Load(pluginType PluginTypeConfig) (interface{}, func(), error) {
	err := l.selectPlugin(pluginType)
	if err != err {
		return nil, nil, err
	}

	l.SelectedPluginKey.Interface = pluginType.Interface

	var pluginCommand *exec.Cmd
	if l.SelectedPluginKey.IsInternal {
		porterPath, err := l.GetPorterPath()
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not determine the path to the porter client")
		}

		pluginCommand = l.NewCommand(porterPath, "plugin", "run", l.SelectedPluginKey.String())
	} else {
		pluginPath, err := l.GetPluginPath(l.SelectedPluginKey.Binary)
		if err != nil {
			return nil, nil, err
		}

		pluginCommand = l.NewCommand(pluginPath, "run", l.SelectedPluginKey.String())
	}
	configReader, err := l.readPluginConfig()
	if err != err {
		return nil, nil, err
	}

	pluginCommand.Stdin = configReader

	// Explicitly set PORTER_HOME for the plugin
	pluginCommand.Env = os.Environ()
	if _, homeSet := os.LookupEnv(config.EnvHOME); !homeSet {
		home, err := l.GetHomeDir()
		if err != nil {
			return nil, nil, err
		}
		pluginCommand.Env = append(pluginCommand.Env, home)
	}

	if l.Config.Debug {
		fmt.Fprintf(l.Err, "Resolved %s plugin to %s\n", pluginType.Interface, l.SelectedPluginKey)
		if l.SelectedPluginConfig != nil {
			fmt.Fprintf(l.Err, "Resolved plugin config: \n %#v\n", l.SelectedPluginConfig)
		}
		fmt.Fprintln(l.Err, strings.Join(pluginCommand.Args, " "))
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "porter",
		Output: l.Err,
		Level:  hclog.Error,
	})

	pluginTypes := map[string]plugin.Plugin{
		pluginType.Interface: pluginType.Plugin,
	}

	var errbuf bytes.Buffer
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugins.HandshakeConfig,
		Plugins:         pluginTypes,
		Cmd:             pluginCommand,
		Logger:          logger,
		Stderr:          &errbuf,
	})
	cleanup := func() {
		client.Kill()
	}

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		cleanup()
		if stderr := errbuf.String(); stderr != "" {
			err = errors.Wrap(errors.New(stderr), err.Error())
		}
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginType.Interface)
	if err != nil {
		cleanup()
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey)
	}

	return raw, cleanup, nil
}

// selectPlugin picks the plugin to use and loads its configuration.
func (l *PluginLoader) selectPlugin(cfg PluginTypeConfig) error {
	l.SelectedPluginKey = nil
	l.SelectedPluginConfig = nil

	var pluginKey string

	defaultStore := cfg.GetDefaultPluggable(l.Config.Data)
	if defaultStore != "" {
		is, err := cfg.GetPluggable(l.Config.Data, defaultStore)
		if err != nil {
			return err
		}
		pluginKey = is.GetPluginSubKey()
		l.SelectedPluginConfig = is.GetConfig()
	}

	// If there isn't a specific plugin configured for this plugin type, fall back to the default plugin for this type
	if pluginKey == "" {
		pluginKey = cfg.GetDefaultPlugin(l.Config.Data)
	}

	key, err := plugins.ParsePluginKey(pluginKey)
	if err != nil {
		return err
	}
	l.SelectedPluginKey = &key

	return nil
}

func (l *PluginLoader) readPluginConfig() (io.Reader, error) {
	if l.SelectedPluginConfig == nil {
		return &bytes.Buffer{}, nil
	}

	b, err := json.Marshal(l.SelectedPluginConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal plugin config %#v", l.SelectedPluginConfig)
	}

	return bytes.NewBuffer(b), nil
}
