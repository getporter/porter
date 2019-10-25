package instancestorageprovider

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"

	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/deislabs/porter/pkg/config"
	instancestorage "github.com/deislabs/porter/pkg/instance-storage"
	"github.com/deislabs/porter/pkg/instance-storage/claimstore"
	"github.com/deislabs/porter/pkg/plugins"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

var _ instancestorage.Provider = &PluginDelegator{}

// A sad hack because claim.Store has a method called Store which prevents us from embedding it as a field
type ClaimStore = claim.Store

// TODO: Move this into pkg/plugins so that every provider backed by plugins can use it
// PluginDelegator provides access to instance storage (claims) by instantiating plugins that
// implement claim (CRUD) storage.
type PluginDelegator struct {
	*config.Config
	ClaimStore
}

func NewPluginDelegator(c *config.Config) *PluginDelegator {
	l := &PluginDelegator{
		Config: c,
	}

	crud := instancestorage.NewDynamicCrudStore(l.connect)
	// this is silly, we can't embed Store because it has a method named Store...
	l.ClaimStore = claim.NewClaimStore(crud)

	return l
}

func (d *PluginDelegator) connect() (crud.Store, func(), error) {
	pluginKey, config, err := d.selectInstanceStoragePlugin()
	pluginKey.Interface = claimstore.PluginKey

	var pluginCommand *exec.Cmd
	if pluginKey.IsInternal {
		porterPath, err := d.GetPorterPath()
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not determine the path to the porter client")
		}

		pluginCommand = d.NewCommand(porterPath, "plugin", "run", pluginKey.String())
	} else {
		pluginPath, err := d.GetPluginPath(pluginKey.Binary)
		if err != nil {
			return nil, nil, err
		}

		pluginCommand = d.NewCommand(pluginPath, "run", pluginKey.String())
	}
	pluginCommand.Stdin = config

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "porter",
		Output: d.Err,
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
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", pluginKey)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(claimstore.PluginKey)
	if err != nil {
		cleanup()
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", pluginKey)
	}

	store, ok := raw.(crud.Store)
	if !ok {
		cleanup()
		return nil, nil, errors.Errorf("the interface exposed by the %s plugin was not instancestorage.ClaimStore", pluginKey)
	}

	return store, cleanup, nil
}

// selectInstanceStoragePlugin picks the plugin to use and loads its configuration.
func (d *PluginDelegator) selectInstanceStoragePlugin() (plugins.PluginKey, io.Reader, error) {
	var pluginId string
	var config interface{}

	defaultStore := d.Config.Data.GetDefaultInstanceStore()
	if defaultStore != "" {
		is, err := d.Config.Data.GetInstanceStore(defaultStore)
		if err != nil {
			return plugins.PluginKey{}, nil, err
		}
		pluginId = is.PluginSubkey
		config = is.Config
	}

	if pluginId == "" {
		pluginId = d.Config.Data.GetInstanceStoragePlugin()
	}

	key, err := plugins.ParsePluginKey(pluginId)
	if err != nil {
		return plugins.PluginKey{}, nil, err
	}

	configInput, err := d.writePluginConfig(config)
	return key, configInput, err
}

func (d *PluginDelegator) writePluginConfig(config interface{}) (io.Reader, error) {
	if config == nil {
		return &bytes.Buffer{}, nil
	}

	b, err := json.Marshal(config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal plugin config %#v", config)
	}

	return bytes.NewBuffer(b), nil
}
