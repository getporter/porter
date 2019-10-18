package instancestorageprovider

import (
	"fmt"
	"os/exec"
	"strings"

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
	pluginId := d.Config.Data.GetInstanceStoragePlugin()
	parts := strings.Split(pluginId, ".")
	isInternal := false
	if len(parts) == 1 {
		isInternal = true
	} else if len(parts) > 2 {
		return nil, nil, errors.New("invalid config value for instance-storage-plugin, can only have two parts PLUGIN_BINARY.IMPLEMENTATION_KEY")
	}

	var pluginCommand *exec.Cmd
	if isInternal {
		pluginImpl := parts[0]
		pluginKey := fmt.Sprintf("%s.porter.%s", claimstore.PluginKey, pluginImpl)
		porterPath, err := d.GetPorterPath()
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not determine the path to the porter client")
		}

		pluginCommand = d.NewCommand(porterPath, "plugin", "run", pluginKey)
	} else {
		pluginBinary := parts[0]
		pluginImpl := parts[1]
		pluginKey := fmt.Sprintf("%s.%s.%s", claimstore.PluginKey, pluginBinary, pluginImpl)
		pluginPath, err := d.GetPluginPath(pluginBinary)
		if err != nil {
			return nil, nil, err
		}

		pluginCommand = d.NewCommand(pluginPath, "run", pluginKey)
	}

	// Create an hclog.Logger
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
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", pluginId)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(claimstore.PluginKey)
	if err != nil {
		cleanup()
		return nil, nil, errors.Wrapf(err, "could not connect to the %s plugin", pluginId)
	}

	store, ok := raw.(crud.Store)
	if !ok {
		cleanup()
		return nil, nil, errors.Errorf("the interface exposed by the %s plugin was not instancestorage.ClaimStore", pluginId)
	}

	return store, cleanup, nil
}
