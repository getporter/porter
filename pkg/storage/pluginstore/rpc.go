package pluginstore

import (
	"net/rpc"

	"get.porter.sh/porter/pkg/storage/plugins"
	"github.com/globalsign/mgo/bson"
)

var _ plugins.StorageProtocol = &Client{}

type Client struct {
	client *rpc.Client
}

func (g *Client) Aggregate(opts plugins.AggregateOptions) ([]bson.Raw, error) {
	var results []bson.Raw
	args := map[string]interface{}{
		"opts": opts,
	}
	err := g.client.Call("Plugin.Aggregate", args, &results)
	return results, err
}

func (g *Client) Count(opts plugins.CountOptions) (int, error) {
	var count int
	args := map[string]interface{}{
		"opts": opts,
	}
	err := g.client.Call("Plugin.Count", args, &count)
	return count, err
}

func (g *Client) EnsureIndex(opts plugins.EnsureIndexOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.EnsureIndex", args, &resp)
}

func (g *Client) Find(opts plugins.FindOptions) ([]bson.Raw, error) {
	var results []bson.Raw
	args := map[string]interface{}{
		"opts": opts,
	}
	err := g.client.Call("Plugin.Find", args, &results)
	return results, err
}

func (g *Client) Insert(opts plugins.InsertOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.Find", args, &resp)
}

func (g *Client) Patch(opts plugins.PatchOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.Patch", args, &resp)
}

func (g *Client) Remove(opts plugins.RemoveOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.Find", args, &resp)
}

func (g *Client) Update(opts plugins.UpdateOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.Find", args, &resp)
}

type Server struct {
	Impl plugins.StorageProtocol
}

func (s *Server) Aggregate(args map[string]interface{}, resp *[]bson.Raw) error {
	var err error
	*resp, err = s.Impl.Aggregate(args["opts"].(plugins.AggregateOptions))
	return err
}

func (s *Server) Count(args map[string]interface{}, resp *int) error {
	var err error
	*resp, err = s.Impl.Count(args["opts"].(plugins.CountOptions))
	return err
}

func (s *Server) EnsureIndex(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.EnsureIndex(args["opts"].(plugins.EnsureIndexOptions))
}

func (s *Server) Find(args map[string]interface{}, resp *[]bson.Raw) error {
	var err error
	*resp, err = s.Impl.Find(args["opts"].(plugins.FindOptions))
	return err
}

func (s *Server) Insert(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Insert(args["opts"].(plugins.InsertOptions))
}

func (s *Server) Patch(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Patch(args["opts"].(plugins.PatchOptions))
}

func (s *Server) Remove(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Remove(args["opts"].(plugins.RemoveOptions))
}

func (s *Server) Update(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Update(args["opts"].(plugins.UpdateOptions))
}
