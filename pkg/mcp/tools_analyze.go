package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *MCPServer) registerAnalyzeTools() {
	s.server.AddTool(&sdkmcp.Tool{
		Name: "analyze_failure",
		Description: "Aggregate failure context for an installation: finds the last failed run " +
			"(or a specific run by ID), then fetches its logs and outputs. " +
			"Returns an error if no failed run is found or the specified run did not fail.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"installation": {"type": "string", "description": "Installation name."},
				"namespace":    {"type": "string", "description": "Namespace."},
				"run_id":       {"type": "string", "description": "Specific run ID to analyze. Omit to use the last failed run."}
			},
			"required": ["installation"]
		}`),
	}, s.analyzeFailure)
}

func (s *MCPServer) analyzeFailure(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Installation string `json:"installation"`
		Namespace    string `json:"namespace"`
		RunID        string `json:"run_id"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	var targetRun porter.DisplayRun

	if args.RunID != "" {
		// Specific run requested — look it up and verify it failed.
		run, err := s.porter.Installations.GetRun(s.ctx, args.RunID)
		if err != nil {
			return toolErr(fmt.Errorf("run %s not found: %w", args.RunID, err)), nil
		}
		targetRun = porter.NewDisplayRun(run)
		// Populate status from results.
		results, err := s.porter.Installations.ListResults(s.ctx, args.RunID)
		if err == nil && len(results) > 0 {
			targetRun.Status = results[len(results)-1].Status
		}
		if targetRun.Status != cnab.StatusFailed {
			return toolErr(fmt.Errorf("run %s has status %q, not %q", args.RunID, targetRun.Status, cnab.StatusFailed)), nil
		}
		if args.Installation != "" && args.Installation != run.Installation {
			return toolErr(fmt.Errorf("run %s does not belong to installation %q", args.RunID, args.Installation)), nil
		}
		if args.Namespace != "" && args.Namespace != run.Namespace {
			return toolErr(fmt.Errorf("run %s does not belong to namespace %q", args.RunID, args.Namespace)), nil
		}
	} else {
		// Find the most recent failed run.
		listOpts := porter.RunListOptions{}
		listOpts.Name = args.Installation
		listOpts.Namespace = args.Namespace

		runs, err := s.porter.ListInstallationRuns(s.ctx, listOpts)
		if err != nil {
			return toolErr(err), nil
		}
		found := false
		for i := len(runs) - 1; i >= 0; i-- {
			if runs[i].Status == cnab.StatusFailed {
				targetRun = runs[i]
				found = true
				break
			}
		}
		if !found {
			return toolErr(errors.New("no failed run found")), nil
		}
	}

	// Fetch logs.
	logsOpts := &porter.LogsShowOptions{}
	logsOpts.RunID = targetRun.ID
	logs, ok, logsErr := s.porter.GetInstallationLogs(s.ctx, logsOpts)
	if !ok && logsErr == nil {
		logsErr = errors.New("no logs found")
	}

	// Fetch outputs.
	outputOpts := &porter.OutputListOptions{}
	outputOpts.RunID = targetRun.ID
	outputs, outputsErr := s.porter.ListBundleOutputs(s.ctx, outputOpts)
	if outputsErr == nil {
		outputs = maskSensitiveOutputs(outputs)
	}

	type analyzeResult struct {
		Run     porter.DisplayRun    `json:"run"`
		Logs    string               `json:"logs,omitempty"`
		Outputs porter.DisplayValues `json:"outputs,omitempty"`
		Errors  []string             `json:"errors,omitempty"`
	}
	result := analyzeResult{
		Run:     targetRun,
		Logs:    logs,
		Outputs: outputs,
	}
	if logsErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("logs: %s", logsErr))
	}
	if outputsErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("outputs: %s", outputsErr))
	}

	return toolJSON(result)
}
