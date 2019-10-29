package pluggable

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"

	"github.com/deislabs/porter/pkg/plugins"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/instance-storage/claimstore"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

type PluginLoader struct {
	*config.Config

	SelectedPluginKey    *plugins.PluginKey
	SelectedPluginConfig io.Reader
}

func NewPluginLoader(c *config.Config) *PluginLoader {
	return &PluginLoader{
		Config: c,
	}
}

func (l *PluginLoader) Load(pluginType PluginTypeConfig) (interface{}, func(), error) {
	err := l.selectPlugin(pluginType)
	l.SelectedPluginKey.Interface = pluginType.Name

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
	pluginCommand.Stdin = l.SelectedPluginConfig

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "porter",
		Output: l.Err,
		Level:  hclog.Error,
	})

	pluginTypes := map[string]plugin.Plugin{
		claimstore.PluginKey: &claimstore.Plugin{},
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugins.HandshakeConfig,
		Plugins:         pluginTypes,
		Cmd:             pluginCommand,
		Logger:          logger,
	})
	cleanup := func() {
		client.Kill()
	}

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		cleanup()
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginType.Name)
	if err != nil {
		cleanup()
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", l.SelectedPluginKey)
	}

	store, ok := raw.(crud.Store)
	if !ok {
		cleanup()
		return nil, nil, errors.Errorf("the interface exposed by the %s plugin was not instancestorage.ClaimStore", l.SelectedPluginKey)
	}

	return store, cleanup, nil
}

// selectPlugin picks the plugin to use and loads its configuration.
func (l *PluginLoader) selectPlugin(cfg PluginTypeConfig) error {
	l.SelectedPluginKey = nil
	l.SelectedPluginConfig = nil

	var pluginId string
	var config interface{}

	defaultStore := cfg.GetDefaultPluggable(l.Config.Data)
	if defaultStore != "" {
		is, err := cfg.GetPluggable(l.Config.Data, defaultStore)
		if err != nil {
			return err
		}
		pluginId = is.GetPluginSubKey()
		config = is.GetConfig()
	}

	if pluginId == "" {
		pluginId = cfg.GetDefaultPlugin(l.Config.Data)
	}

	key, err := plugins.ParsePluginKey(pluginId)
	if err != nil {
		return err
	}
	l.SelectedPluginKey = &key

	configInput, err := l.writePluginConfig(config)
	if err != nil {
		return err
	}

	l.SelectedPluginConfig = configInput
	return nil
}

func (l *PluginLoader) writePluginConfig(config interface{}) (io.Reader, error) {
	if config == nil {
		return &bytes.Buffer{}, nil
	}

	b, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal plugin config %#v", config)
	}

	return bytes.NewBuffer(b), nil
}
