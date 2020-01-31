package plugins

import "fmt"

// TestPluginProvider helps us test Porter.Mixins in our unit tests without actually hitting any real mixins on the file system.
type TestPluginProvider struct{}

func (p *TestPluginProvider) List() ([]string, error) {
	mixins := []string{"plugin1", "plugin2", "plugin3", "unknown"}
	return mixins, nil
}

func (p *TestPluginProvider) GetMetadata(pluginName string) (*PluginMetadata, error) {
	var impl []Implementaion
	if pluginName != "unknown" {
		impl = []Implementaion{
			{Type: "instance-storage", Name: "blob"},
			{Type: "instance-storage", Name: "mongo"},
		}
	}
	return &PluginMetadata{
		Name:            pluginName,
		ClientPath:      fmt.Sprintf("/home/porter/.porter/plugins/%s", pluginName),
		Implementations: impl,
		VersionInfo:     VersionInfo{Version: "v1.0", Commit: "abc123", Author: "Deis Labs"},
	}, nil
}
