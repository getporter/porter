package syft

import (
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/sbom"
	"get.porter.sh/porter/pkg/sbom/plugins"
	"get.porter.sh/porter/pkg/sbom/pluginstore"
	"github.com/hashicorp/go-plugin"
	"github.com/mitchellh/mapstructure"
)

const PluginKey = plugins.PluginInterface + ".porter.syft"

var _ plugins.SBOMGeneratorProtocol = &Plugin{}

type PluginConfig struct {
}

// Plugin is the plugin wrapper for accessing secrets from a local filesystem.
type Plugin struct {
	sbom.SBOMGenerator
}

func NewPlugin(c *portercontext.Context, rawCfg interface{}) (plugin.Plugin, error) {
	cfg := PluginConfig{}
	if err := mapstructure.Decode(rawCfg, &cfg); err != nil {
		return nil, fmt.Errorf("error reading plugin configuration: %w", err)
	}

	impl := NewSBOMGenerator(c, cfg)
	return pluginstore.NewPlugin(c, impl), nil
}
