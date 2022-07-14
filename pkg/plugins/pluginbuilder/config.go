package pluginbuilder

import (
	"encoding/json"
	"fmt"
	"io"
)

// LoadConfig reads the plugin's configuration from stdin.
func (p *PorterPlugin) loadConfig() error {
	// Use the configuration data structure provided by the plugin, if specified.
	if p.opts.DefaultConfig != nil {
		p.pluginConfig = p.opts.DefaultConfig
	} else {
		p.pluginConfig = make(map[string]interface{})
	}

	if err := json.NewDecoder(p.porterConfig.In).Decode(&p.pluginConfig); err != nil {
		if err == io.EOF {
			// No plugin pluginConfig was specified
			return nil
		}
		return fmt.Errorf("error unmarshaling the plugins configuration data from stdin into %T: %w", p.pluginConfig, err)
	}
	return nil
}
