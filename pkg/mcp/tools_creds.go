package mcp

import (
	"context"
	"encoding/json"

	"get.porter.sh/porter/pkg/porter"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *MCPServer) registerCredentialTools() {
	s.server.AddTool(&sdkmcp.Tool{
		Name:        "list_credentials",
		Description: "List Porter credential sets.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"namespace": {"type": "string", "description": "Namespace to filter by."},
				"name":      {"type": "string", "description": "Filter by credential set name."}
			}
		}`),
	}, s.listCredentials)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "show_credential",
		Description: "Show details of a Porter credential set.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":      {"type": "string", "description": "Credential set name."},
				"namespace": {"type": "string", "description": "Namespace."}
			},
			"required": ["name"]
		}`),
	}, s.showCredential)
}

func (s *MCPServer) listCredentials(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	creds, err := s.porter.ListCredentials(s.ctx, porter.ListOptions{
		Namespace: args.Namespace,
		Name:      args.Name,
	})
	if err != nil {
		return toolErr(err), nil
	}
	return toolJSON(creds)
}

func (s *MCPServer) showCredential(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	cs, err := s.porter.Credentials.GetCredentialSet(s.ctx, args.Namespace, args.Name)
	if err != nil {
		return toolErr(err), nil
	}
	return toolJSON(porter.NewDisplayCredentialSet(cs))
}
