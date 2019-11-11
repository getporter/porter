package secrets

import (
	"encoding/gob"
	"net/rpc"

	"github.com/cnabio/cnab-go/credentials"
)

var _ Store = &Client{}

type Client struct {
	client *rpc.Client
}

func init() {
	gob.Register(credentials.Source{})
}

func (g *Client) Resolve(source credentials.Source) (string, error) {
	var resp string
	err := g.client.Call("Plugin.Resolve", source, &resp)
	return resp, err
}

type Server struct {
	Impl Store
}

func (s *Server) Resolve(source credentials.Source, resp *string) error {
	var err error
	*resp, err = s.Impl.Resolve(source)
	return err
}
