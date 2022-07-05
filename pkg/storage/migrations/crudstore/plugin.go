package crudstore

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

var _ plugin.Plugin = &Plugin{}

// Plugin is a generic type of plugin for working with any implementation of a crud store.
type Plugin struct {
	Impl Store
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &Server{Impl: p.Impl}, nil
}

func (Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &Client{client: c}, nil
}
