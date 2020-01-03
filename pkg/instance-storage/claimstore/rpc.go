package claimstore

import (
	"net/rpc"

	"github.com/cnabio/cnab-go/utils/crud"
)

var _ crud.Store = &Client{}

type Client struct {
	client *rpc.Client
}

func (g *Client) Read(name string) ([]byte, error) {
	var resp []byte
	err := g.client.Call("Plugin.Read", name, &resp)
	return resp, err
}

func (g *Client) List() ([]string, error) {
	var resp []string
	err := g.client.Call("Plugin.List", new(interface{}), &resp)
	return resp, err
}

func (g *Client) Store(name string, data []byte) error {
	var resp interface{}
	args := map[string]interface{}{
		"name": name,
		"data": data,
	}
	return g.client.Call("Plugin.Store", args, &resp)
}

func (g *Client) Delete(name string) error {
	var resp interface{}
	return g.client.Call("Plugin.Delete", name, &resp)
}

type Server struct {
	Impl crud.Store
}

func (s *Server) Read(name string, resp *[]byte) error {
	var err error
	*resp, err = s.Impl.Read(name)
	return err
}

func (s *Server) List(args interface{}, resp *[]string) error {
	var err error
	*resp, err = s.Impl.List()
	return err
}

func (s *Server) Store(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.Store(args["name"].(string), args["data"].([]byte))
}

func (s *Server) Delete(name string, resp *interface{}) error {
	return s.Impl.Delete(name)
}
