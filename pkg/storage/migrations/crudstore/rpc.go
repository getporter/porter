package crudstore

import (
	"net/rpc"
)

var _ Store = &Client{}

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

func (g *Client) List(itemType string, group string) ([]string, error) {
	var resp []string
	args := map[string]interface{}{
		"itemType": itemType,
		"group":    group,
	}
	err := g.client.Call("Plugin.List", args, &resp)
	return resp, err
}

type Server struct {
	Impl Store
}

func (s *Server) Read(args map[string]interface{}, resp *[]byte) error {
	var err error
	*resp, err = s.Impl.Read(args["itemType"].(string), args["name"].(string))
	return err
}

func (s *Server) List(args map[string]interface{}, resp *[]string) error {
	var err error
	*resp, err = s.Impl.List(args["itemType"].(string), args["group"].(string))
	return err
}
