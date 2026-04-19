package mcp

import (
	"context"
	"encoding/json"
	"errors"

	"get.porter.sh/porter/pkg/porter"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

var errSensitiveOutput = errors.New("output is sensitive and cannot be returned")

func (s *MCPServer) registerOutputTools() {
	s.server.AddTool(&sdkmcp.Tool{
		Name:        "list_outputs",
		Description: "List outputs from an installation run. Sensitive values are masked.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"installation": {"type": "string", "description": "Installation name."},
				"namespace":    {"type": "string", "description": "Namespace."},
				"run_id":       {"type": "string", "description": "Run ID. Omit for the latest run."}
			},
			"required": ["installation"]
		}`),
	}, s.listOutputs)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "get_output",
		Description: "Get a specific output value from an installation. Returns an error for sensitive outputs.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"installation": {"type": "string", "description": "Installation name."},
				"output_name":  {"type": "string", "description": "Output name."},
				"namespace":    {"type": "string", "description": "Namespace."}
			},
			"required": ["installation", "output_name"]
		}`),
	}, s.getOutput)
}

func (s *MCPServer) listOutputs(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Installation string `json:"installation"`
		Namespace    string `json:"namespace"`
		RunID        string `json:"run_id"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}
	if args.Installation == "" {
		return toolErr(errors.New("installation is required")), nil
	}

	opts := &porter.OutputListOptions{}
	opts.Name = args.Installation
	opts.Namespace = args.Namespace
	opts.RunID = args.RunID

	outputs, err := s.porter.ListBundleOutputs(s.ctx, opts)
	if err != nil {
		return toolErr(err), nil
	}
	return toolJSON(maskSensitiveOutputs(outputs))
}

func (s *MCPServer) getOutput(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Installation string `json:"installation"`
		OutputName   string `json:"output_name"`
		Namespace    string `json:"namespace"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	// Check sensitivity via list before returning the raw value.
	opts := &porter.OutputListOptions{}
	opts.Name = args.Installation
	opts.Namespace = args.Namespace
	outputs, err := s.porter.ListBundleOutputs(s.ctx, opts)
	if err != nil {
		return toolErr(err), nil
	}
	for _, o := range outputs {
		if o.Name == args.OutputName && o.Sensitive {
			return toolErr(errSensitiveOutput), nil
		}
	}

	val, err := s.porter.ReadBundleOutput(s.ctx, args.OutputName, args.Installation, args.Namespace)
	if err != nil {
		return toolErr(err), nil
	}
	return toolText(val), nil
}
