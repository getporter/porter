package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/printer"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *MCPServer) registerBundleTools() {
	s.server.AddTool(&sdkmcp.Tool{
		Name:        "explain_bundle",
		Description: "Show parameters, credentials, outputs, and actions of a published bundle.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"reference": {"type": "string", "description": "OCI reference, e.g. ghcr.io/org/bundle:v1.0.0"}
			},
			"required": ["reference"]
		}`),
	}, s.explainBundle)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "list_installations",
		Description: "List all Porter installations.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"namespace":      {"type": "string",  "description": "Namespace to filter by."},
				"all_namespaces": {"type": "boolean", "description": "List across all namespaces."},
				"name":           {"type": "string",  "description": "Filter by installation name."}
			}
		}`),
	}, s.listInstallations)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "show_installation",
		Description: "Show details and current status of a Porter installation.",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":      {"type": "string", "description": "Installation name."},
				"namespace": {"type": "string", "description": "Namespace."}
			},
			"required": ["name"]
		}`),
	}, s.showInstallation)

	if !s.opts.AllowWrite {
		return
	}

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "install_bundle",
		Description: "Install a bundle (requires --allow-write).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":            {"type": "string", "description": "Installation name."},
				"namespace":       {"type": "string", "description": "Namespace."},
				"reference":       {"type": "string", "description": "OCI reference to the bundle."},
				"params":          {"type": "object", "description": "Parameter overrides.", "additionalProperties": {"type": "string"}},
				"credential_sets": {"type": "array",  "description": "Credential set names.", "items": {"type": "string"}},
				"parameter_sets":  {"type": "array",  "description": "Parameter set names.",  "items": {"type": "string"}}
			},
			"required": ["reference"]
		}`),
	}, s.installBundle)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "upgrade_bundle",
		Description: "Upgrade an existing installation (requires --allow-write).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":            {"type": "string", "description": "Installation name."},
				"namespace":       {"type": "string", "description": "Namespace."},
				"reference":       {"type": "string", "description": "New bundle OCI reference (optional)."},
				"params":          {"type": "object", "description": "Parameter overrides.", "additionalProperties": {"type": "string"}},
				"credential_sets": {"type": "array",  "description": "Credential set names.", "items": {"type": "string"}}
			},
			"required": ["name"]
		}`),
	}, s.upgradeBundle)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "uninstall_bundle",
		Description: "Uninstall an installation (requires --allow-write).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":      {"type": "string", "description": "Installation name."},
				"namespace": {"type": "string", "description": "Namespace."}
			},
			"required": ["name"]
		}`),
	}, s.uninstallBundle)

	s.server.AddTool(&sdkmcp.Tool{
		Name:        "invoke_bundle",
		Description: "Invoke a custom action on an installation (requires --allow-write).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"name":      {"type": "string", "description": "Installation name."},
				"action":    {"type": "string", "description": "Custom action name."},
				"namespace": {"type": "string", "description": "Namespace."},
				"params":    {"type": "object", "description": "Parameter overrides.", "additionalProperties": {"type": "string"}}
			},
			"required": ["name", "action"]
		}`),
	}, s.invokeBundle)
}

func (s *MCPServer) explainBundle(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Reference string `json:"reference"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := porter.ExplainOpts{}
	opts.Reference = args.Reference
	opts.RawFormat = string(printer.FormatJson)

	out, err := s.captureOutput(func() error {
		return s.porter.Explain(s.ctx, opts)
	})
	if err != nil {
		return toolErr(err), nil
	}
	return toolText(out), nil
}

func (s *MCPServer) listInstallations(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Namespace     string `json:"namespace"`
		AllNamespaces bool   `json:"all_namespaces"`
		Name          string `json:"name"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	installs, err := s.porter.ListInstallations(s.ctx, porter.ListOptions{
		Namespace:     args.Namespace,
		AllNamespaces: args.AllNamespaces,
		Name:          args.Name,
	})
	if err != nil {
		return toolErr(err), nil
	}
	return toolJSON(installs)
}

func (s *MCPServer) showInstallation(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := porter.ShowOptions{}
	opts.Name = args.Name
	opts.Namespace = args.Namespace

	inst, lastRun, err := s.porter.GetInstallation(s.ctx, opts)
	if err != nil {
		return toolErr(err), nil
	}

	type showResponse struct {
		Installation porter.DisplayInstallation `json:"installation"`
		LastRun      any                        `json:"lastRun,omitempty"`
	}
	resp := showResponse{Installation: porter.NewDisplayInstallation(inst)}
	if lastRun != nil {
		resp.LastRun = porter.NewDisplayRun(*lastRun)
	}
	return toolJSON(resp)
}

func (s *MCPServer) installBundle(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Name           string            `json:"name"`
		Namespace      string            `json:"namespace"`
		Reference      string            `json:"reference"`
		Params         map[string]string `json:"params"`
		CredentialSets []string          `json:"credential_sets"`
		ParameterSets  []string          `json:"parameter_sets"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := porter.NewInstallOptions()
	opts.Name = args.Name
	opts.Namespace = args.Namespace
	opts.Reference = args.Reference
	opts.Params = paramsToSlice(args.Params)
	opts.CredentialIdentifiers = args.CredentialSets
	opts.ParameterSets = args.ParameterSets

	if err := opts.Validate(s.ctx, nil, s.porter); err != nil {
		return toolErr(err), nil
	}

	out, execErr := s.captureOutput(func() error {
		return s.porter.InstallBundle(s.ctx, opts)
	})
	if execErr != nil {
		return toolErr(fmt.Errorf("install failed: %w\n%s", execErr, out)), nil
	}
	return toolText("installed successfully\n" + out), nil
}

func (s *MCPServer) upgradeBundle(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Name           string            `json:"name"`
		Namespace      string            `json:"namespace"`
		Reference      string            `json:"reference"`
		Params         map[string]string `json:"params"`
		CredentialSets []string          `json:"credential_sets"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := porter.NewUpgradeOptions()
	opts.Name = args.Name
	opts.Namespace = args.Namespace
	opts.Reference = args.Reference
	opts.Params = paramsToSlice(args.Params)
	opts.CredentialIdentifiers = args.CredentialSets

	if err := opts.Validate(s.ctx, nil, s.porter); err != nil {
		return toolErr(err), nil
	}

	out, execErr := s.captureOutput(func() error {
		return s.porter.UpgradeBundle(s.ctx, opts)
	})
	if execErr != nil {
		return toolErr(fmt.Errorf("upgrade failed: %w\n%s", execErr, out)), nil
	}
	return toolText("upgraded successfully\n" + out), nil
}

func (s *MCPServer) uninstallBundle(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := porter.NewUninstallOptions()
	opts.Name = args.Name
	opts.Namespace = args.Namespace

	if err := opts.Validate(s.ctx, nil, s.porter); err != nil {
		return toolErr(err), nil
	}

	out, execErr := s.captureOutput(func() error {
		return s.porter.UninstallBundle(s.ctx, opts)
	})
	if execErr != nil {
		return toolErr(fmt.Errorf("uninstall failed: %w\n%s", execErr, out)), nil
	}
	return toolText("uninstalled successfully\n" + out), nil
}

func (s *MCPServer) invokeBundle(_ context.Context, req *sdkmcp.CallToolRequest) (*sdkmcp.CallToolResult, error) {
	var args struct {
		Name      string            `json:"name"`
		Action    string            `json:"action"`
		Namespace string            `json:"namespace"`
		Params    map[string]string `json:"params"`
	}
	if err := unmarshalArgs(req, &args); err != nil {
		return toolErr(err), nil
	}

	opts := porter.NewInvokeOptions()
	opts.Name = args.Name
	opts.Namespace = args.Namespace
	opts.Action = args.Action
	opts.Params = paramsToSlice(args.Params)

	if err := opts.Validate(s.ctx, nil, s.porter); err != nil {
		return toolErr(err), nil
	}

	out, execErr := s.captureOutput(func() error {
		return s.porter.InvokeBundle(s.ctx, opts)
	})
	if execErr != nil {
		return toolErr(fmt.Errorf("invoke failed: %w\n%s", execErr, out)), nil
	}
	return toolText("invoked successfully\n" + out), nil
}
