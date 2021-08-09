package secrets

import (
	"net/rpc"

	"get.porter.sh/porter/pkg/secrets/plugins"
)

var _ plugins.SecretsProtocol = &Client{}

type Client struct {
	client *rpc.Client
}

func (g *Client) Resolve(keyName string, keyValue string) (string, error) {
	args := map[string]interface{}{
		"keyName":  keyName,
		"keyValue": keyValue,
	}
	var resp string
	err := g.client.Call("Plugin.Resolve", args, &resp)
	return resp, err
}

type Server struct {
	Impl plugins.SecretsProtocol
}

func (s *Server) Resolve(args map[string]interface{}, resp *string) error {
	var err error
	*resp, err = s.Impl.Resolve(args["keyName"].(string), args["keyValue"].(string))
	return err
}
