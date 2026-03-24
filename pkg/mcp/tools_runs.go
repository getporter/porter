package mcp

import (
	"context"
	"encoding/json"

	"get.porter.sh/porter/pkg/porter"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *MCPServer) registerRunTools() {
	s.server.AddTool(&sdkmcp.Tool{
		Name:        "list_runs",
		Description: "List execution runs of a Porter installation.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"installation": {"type": "string", "description": "Installation name."},
				"namespace":    {"type": "string", "description": "Namespace."}
			},
			"required": ["installation"]
		}`),
	}, s.listRuns)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "get_logs",
		Description: "Get execution logs for an installation run. Provide either installation (latest run) or run_id.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"installation": {"type": "string", "description": "Installation name (returns logs from the latest run)."},
				"run_id":       {"type": "string", "description": "Specific run ID."},
				"namespace":    {"type": "string", "description": "Namespace."}
			}
		}`),
	}, s.getLogs)
}

func (s *MCPServer) listRuns(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Installation string `json:"installation"`
		Namespace    string `json:"namespace"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := porter.RunListOptions{}
	opts.Name = args.Installation
	opts.Namespace = args.Namespace

	runs, err := s.porter.ListInstallationRuns(s.ctx, opts)
	if err != nil {
		return toolErr(err), nil
	}
	return toolJSON(runs)
}

func (s *MCPServer) getLogs(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Installation string `json:"installation"`
		RunID        string `json:"run_id"`
		Namespace    string `json:"namespace"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := &porter.LogsShowOptions{}
	opts.Name = args.Installation
	opts.RunID = args.RunID
	opts.Namespace = args.Namespace

	logs, ok, err := s.porter.GetInstallationLogs(s.ctx, opts)
	if err != nil {
		return toolErr(err), nil
	}
	if !ok {
		return toolErr(errNoLogs), nil
	}
	return toolText(logs), nil
}
