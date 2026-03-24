package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	pkg "get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/porter"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

var errNoLogs = errors.New("no logs found")

// ServerOptions configures the MCP server.
type ServerOptions struct {
	// AllowWrite enables mutating tools: install, upgrade, uninstall, invoke.
	AllowWrite bool
}

// MCPServer wraps a porter.Porter and exposes its operations as MCP tools.
type MCPServer struct {
	porter *porter.Porter
	opts   ServerOptions
	server *sdkmcp.Server
	// ctx is the server lifetime context. All Porter API calls use this
	// instead of the per-request context, because plugin subprocesses are
	// started with exec.CommandContext and would be killed when the
	// per-request context is cancelled after each tool call completes.
	ctx context.Context
	mu  sync.Mutex // protects p.Out during output capture
}

// NewMCPServer creates an MCPServer. p must already be connected.
// p.Out is redirected to stderr to prevent writes from corrupting the MCP stdio stream.
func NewMCPServer(p *porter.Porter, opts ServerOptions) *MCPServer {
	p.Out = os.Stderr

	s := sdkmcp.NewServer(
		&sdkmcp.Implementation{Name: "porter", Version: pkg.Version},
		nil,
	)
	return &MCPServer{porter: p, opts: opts, server: s}
}

// RegisterTools registers all MCP tools on the server.
func (s *MCPServer) RegisterTools() {
	s.registerBundleTools()
	s.registerRunTools()
	s.registerOutputTools()
	s.registerCredentialTools()
	s.registerParameterTools()
	s.registerAnalyzeTools()
}

// ServeStdio starts the MCP server and blocks until ctx is cancelled or stdin is closed.
func (s *MCPServer) ServeStdio(ctx context.Context) error {
	s.ctx = ctx
	return s.server.Run(ctx, &sdkmcp.StdioTransport{})
}

// captureOutput temporarily redirects p.Out to a buffer for the duration of fn,
// then restores stderr. The mutex serializes concurrent captures.
func (s *MCPServer) captureOutput(fn func() error) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var buf bytes.Buffer
	s.porter.Out = &buf
	defer func() { s.porter.Out = os.Stderr }()

	err := fn()
	return buf.String(), err
}

// toolJSON marshals data to a JSON text tool result.
func toolJSON(data any) (*sdkmcp.CallToolResult, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return toolErr(err), nil
	}
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: string(b)}},
	}, nil
}

// toolText returns a plain-text tool result.
func toolText(text string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: text}},
	}
}

// toolErr returns an IsError tool result.
func toolErr(err error) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: err.Error()}},
		IsError: true,
	}
}

// unmarshalArgs deserializes the raw JSON arguments from a tool request.
func unmarshalArgs(req *sdkmcp.CallToolRequest, out any) error {
	if len(req.Params.Arguments) == 0 {
		return nil
	}
	return json.Unmarshal(req.Params.Arguments, out)
}

// maskSensitiveOutputs replaces sensitive output values with "***".
func maskSensitiveOutputs(vals porter.DisplayValues) porter.DisplayValues {
	result := make(porter.DisplayValues, len(vals))
	for i, v := range vals {
		if v.Sensitive {
			v.Value = "***"
		}
		result[i] = v
	}
	return result
}

// paramsToSlice converts a map of parameter key-values to Porter's NAME=VALUE slice format.
func paramsToSlice(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	params := make([]string, 0, len(m))
	for k, v := range m {
		params = append(params, k+"="+v)
	}
	return params
}
