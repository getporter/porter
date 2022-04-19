package pluginstore

import (
	"context"
	"net/rpc"

	"get.porter.sh/porter/pkg/storage/plugins"
	"go.mongodb.org/mongo-driver/bson"
)

var _ plugins.StorageProtocol = &Client{}

type Client struct {
	client *rpc.Client
}

func (g *Client) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	var results []bson.Raw
	args := map[string]interface{}{
		"opts": opts,
	}
	err := g.client.Call("Plugin.Aggregate", args, &results)
	return results, err
}

func (g *Client) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	var count int64
	args := map[string]interface{}{
		"opts": opts,
	}
	err := g.client.Call("Plugin.Count", args, &count)
	return count, err
}

func (g *Client) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.EnsureIndex", args, &resp)
}

func (g *Client) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	var results []bson.Raw
	args := map[string]interface{}{
		"opts": opts,
	}
	err := g.client.Call("Plugin.Find", args, &results)
	return results, err
}

func (g *Client) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.Find", args, &resp)
}

func (g *Client) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.Patch", args, &resp)
}

func (g *Client) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	var resp interface{}
	args := map[string]interface{}{
		"opts": opts,
	}
	return g.client.Call("Plugin.Find", args, &resp)
}

func (g *Client) Update(ctx context.Context, opts plugins.UpdateOptions) error {
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
	*resp, err = s.Impl.Aggregate(context.Background(), args["opts"].(plugins.AggregateOptions))
	return err
}

func (s *Server) Count(args map[string]interface{}, resp *int64) error {
	var err error
	*resp, err = s.Impl.Count(context.Background(), args["opts"].(plugins.CountOptions))
	return err
}

func (s *Server) EnsureIndex(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.EnsureIndex(context.Background(), args["opts"].(plugins.EnsureIndexOptions))
}

func (s *Server) Find(args map[string]interface{}, resp *[]bson.Raw) error {
	var err error
	*resp, err = s.Impl.Find(context.Background(), args["opts"].(plugins.FindOptions))
	return err
}

func (s *Server) Insert(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Insert(context.Background(), args["opts"].(plugins.InsertOptions))
}

func (s *Server) Patch(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Patch(context.Background(), args["opts"].(plugins.PatchOptions))
}

func (s *Server) Remove(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Remove(context.Background(), args["opts"].(plugins.RemoveOptions))
}

func (s *Server) Update(args map[string]interface{}, _ *interface{}) error {
	return s.Impl.Update(context.Background(), args["opts"].(plugins.UpdateOptions))
}
