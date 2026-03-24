package mcp

import (
	"context"
	"encoding/json"

	"get.porter.sh/porter/pkg/porter"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *MCPServer) registerParameterTools() {
	s.server.AddTool(&sdkmcp.Tool{
		Name:        "list_parameters",
		Description: "List Porter parameter sets.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"namespace": {"type": "string", "description": "Namespace to filter by."},
				"name":      {"type": "string", "description": "Filter by parameter set name."}
			}
		}`),
	}, s.listParameters)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "show_parameter",
		Description: "Show details of a Porter parameter set.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":      {"type": "string", "description": "Parameter set name."},
				"namespace": {"type": "string", "description": "Namespace."}
			},
			"required": ["name"]
		}`),
	}, s.showParameter)
}

func (s *MCPServer) listParameters(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	params, err := s.porter.ListParameters(s.ctx, porter.ListOptions{
		Namespace: args.Namespace,
		Name:      args.Name,
	})
	if err != nil {
		return toolErr(err), nil
	}
	return toolJSON(params)
}

func (s *MCPServer) showParameter(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	ps, err := s.porter.Parameters.GetParameterSet(s.ctx, args.Namespace, args.Name)
	if err != nil {
		return toolErr(err), nil
	}
	return toolJSON(porter.NewDisplayParameterSet(ps))
}
