package crudstore

import (
	"net/rpc"

	"github.com/cnabio/cnab-go/utils/crud"
)

var _ crud.Store = &Client{}

type Client struct {
	client *rpc.Client
}

func (g *Client) Read(itemType string, name string) ([]byte, error) {
	var resp []byte
	args := map[string]interface{}{
		"itemType": itemType,
		"name":     name,
	}
	err := g.client.Call("Plugin.Read", args, &resp)
	return resp, err
}

func (g *Client) Count(itemType string, group string) (int, error) {
	var resp int
	args := map[string]interface{}{
		"itemType": itemType,
		"group":    group,
	}
	err := g.client.Call("Plugin.Count", args, &resp)
	return resp, err
}

func (g *Client) List(itemType string, group string) ([]string, error) {
	var resp []string
	args := map[string]interface{}{
		"itemType": itemType,
		"group":    group,
	}
	err := g.client.Call("Plugin.List", args, &resp)
	return resp, err
}

func (g *Client) Save(itemType string, group string, name string, data []byte) error {
	var resp interface{}
	args := map[string]interface{}{
		"itemType": itemType,
		"group":    group,
		"name":     name,
		"data":     data,
	}
	return g.client.Call("Plugin.Save", args, &resp)
}

func (g *Client) Delete(itemType string, name string) error {
	var resp interface{}
	args := map[string]interface{}{
		"itemType": itemType,
		"name":     name,
	}
	return g.client.Call("Plugin.Delete", args, &resp)
}

type Server struct {
	Impl crud.Store
}

func (s *Server) Read(args map[string]interface{}, resp *[]byte) error {
	var err error
	*resp, err = s.Impl.Read(args["itemType"].(string), args["name"].(string))
	return err
}

func (s *Server) Count(args map[string]interface{}, resp *int) error {
	var err error
	*resp, err = s.Impl.Count(args["itemType"].(string), args["group"].(string))
	return err
}

func (s *Server) List(args map[string]interface{}, resp *[]string) error {
	var err error
	*resp, err = s.Impl.List(args["itemType"].(string), args["group"].(string))
	return err
}

func (s *Server) Save(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.Save(args["itemType"].(string), args["group"].(string), args["name"].(string), args["data"].([]byte))
}

func (s *Server) Delete(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.Delete(args["itemType"].(string), args["name"].(string))
}
