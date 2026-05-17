package main

import (
	"get.porter.sh/porter/pkg/mcp"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildMCPCommand(p *porter.Porter) *cobra.Command {
	opts := mcp.ServerOptions{}
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start an MCP server over stdio",
		Long: `Start a Model Context Protocol (MCP) server that exposes Porter
operations as tools for LLM agents.

The server communicates over stdin/stdout using the MCP protocol.
Read-only tools are always available. Mutating tools (install, upgrade,
uninstall, invoke) require the --allow-write flag.

Configure your MCP client (e.g. Claude Desktop) with:
  {"command": "porter", "args": ["mcp"]}
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := mcp.NewMCPServer(p, opts)
			srv.RegisterTools()
			return srv.ServeStdio(cmd.Context())
		},
	}
	f := cmd.Flags()
	f.BoolVar(&opts.AllowWrite, "allow-write", false,
		"Enable mutating tools: install, upgrade, uninstall, invoke")
	return cmd
}
