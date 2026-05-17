---
title: "porter mcp"
slug: porter_mcp
url: /cli/porter_mcp/
---
## porter mcp

Start an MCP server over stdio

### Synopsis

Start a Model Context Protocol (MCP) server that exposes Porter
operations as tools for LLM agents.

The server communicates over stdin/stdout using the MCP protocol.
Read-only tools are always available. Mutating tools (install, upgrade,
uninstall, invoke) require the --allow-write flag.

Configure your MCP client (e.g. Claude Desktop) with:
  {"command": "porter", "args": ["mcp"]}


```
porter mcp [flags]
```

### Options

```
      --allow-write   Enable mutating tools: install, upgrade, uninstall, invoke
  -h, --help          help for mcp
```

### Options inherited from parent commands

```
      --context string         Name of the configuration context to use. When unset, Porter uses the current-context from the config file, falling back to the context named "default".
      --experimental strings   Comma separated list of experimental features to enable. See https://porter.sh/configuration/#experimental-feature-flags for available feature flags.
      --verbosity string       Threshold for printing messages to the console. Available values are: debug, info, warning, error. (default "info")
```

### SEE ALSO

* [porter](/cli/porter/)	 - With Porter you can package your application artifact, client tools, configuration and deployment logic together as a versioned bundle that you can distribute, and then install with a single command.

Most commands require a Docker daemon, either local or remote.

Try our QuickStart https://porter.sh/quickstart to learn how to use Porter.


