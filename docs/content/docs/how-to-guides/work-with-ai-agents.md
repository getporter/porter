---
title: Working with AI Agents
description: Connect AI coding assistants and LLM agents to Porter using the Model Context Protocol.
weight: 3
aliases:
  - /how-to-guides/work-with-ai-agents/
---

Porter includes a built-in [Model Context Protocol (MCP)][mcp] server that
exposes bundle operations as tools for LLM agents and AI coding assistants such
as Claude Code, Cursor, and VS Code Copilot. This lets an agent install,
inspect, and troubleshoot bundles on your behalf using natural language.

- [Prerequisites](#prerequisites)
- [Configure your MCP client](#configure-your-mcp-client)
  - [Claude Code](#claude-code)
  - [VS Code (GitHub Copilot)](#vs-code-github-copilot)
  - [Cursor](#cursor)
- [Available tools](#available-tools)
- [Enabling write operations](#enabling-write-operations)
- [Security considerations](#security-considerations)

## Prerequisites

- Porter [installed and configured](/install/)
- An MCP-compatible AI client (Claude Code, Cursor, VS Code Copilot, etc.)

## Configure your MCP client

Porter's MCP server communicates over stdin/stdout. All clients need the same
two values:

| Setting | Value |
|---------|-------|
| **command** | `porter` |
| **args** | `["mcp"]` |

### Claude Code

Add the server to your project:

```console
claude mcp add porter -- porter mcp
```

Or add it to your user-level configuration so it is available in every project:

```console
claude mcp add --scope user porter -- porter mcp
```

Verify the server is registered:

```console
claude mcp list
```

### VS Code (GitHub Copilot)

Add to `.vscode/mcp.json` in your workspace:

```json
{
  "servers": {
    "porter": {
      "type": "stdio",
      "command": "porter",
      "args": ["mcp"]
    }
  }
}
```

### Cursor

Open **Cursor Settings → MCP** and add:

```json
{
  "porter": {
    "command": "porter",
    "args": ["mcp"]
  }
}
```

## Available tools

### Read-only tools (always available)

| Tool | What the agent can do |
|------|-----------------------|
| `explain_bundle` | Show parameters, credentials, outputs, and actions of a published bundle by OCI reference |
| `list_installations` | List Porter installations, optionally filtered by namespace or name |
| `show_installation` | Show the current status and details of a specific installation |
| `list_runs` | List execution history for an installation |
| `get_logs` | Retrieve logs from a run (latest run or a specific run ID) |
| `list_outputs` | List output values from an installation run (sensitive values are masked) |
| `get_output` | Read a specific output value (blocked for sensitive outputs) |
| `list_credentials` | List available credential sets |
| `show_credential` | Show the definition of a credential set |
| `list_parameters` | List available parameter sets |
| `show_parameter` | Show the definition of a parameter set |
| `analyze_failure` | Aggregate failure context: finds the last failed run (or a specific run by ID) and returns its logs and outputs in one call |

### Write tools (require `--allow-write`)

| Tool | What the agent can do |
|------|-----------------------|
| `install_bundle` | Install a bundle from an OCI reference |
| `upgrade_bundle` | Upgrade an existing installation |
| `uninstall_bundle` | Uninstall an installation |
| `invoke_bundle` | Run a custom action on an installation |

## Enabling write operations

Write tools are disabled by default so that an agent browsing installations
cannot accidentally modify them. Pass `--allow-write` to expose the install,
upgrade, uninstall, and invoke tools:

```console
# Claude Code
claude mcp add porter -- porter mcp --allow-write

# VS Code / Cursor — add "--allow-write" to the args array
"args": ["mcp", "--allow-write"]
```

When `--allow-write` is not set, the write tools are not registered at all and
will not appear in the agent's tool list.

## Security considerations

- **Read-only by default.** Without `--allow-write`, no bundle lifecycle
  operations are exposed.
- **Sensitive outputs are masked.** Values marked sensitive in the bundle are
  returned as `***` in `list_outputs` and are blocked entirely in `get_output`.
- **Credentials are not exposed.** The server can show the structure of a
  credential set (which sources are configured) but never returns the resolved
  secret values.
- **Local access only.** The MCP server uses stdio transport; there is no
  network listener and no authentication is required or possible.
- **Porter's own access controls apply.** The server runs as the current OS
  user with the existing Porter configuration. It can only access namespaces
  and plugins that the user can access directly.

[mcp]: https://modelcontextprotocol.io
