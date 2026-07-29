---
title: "Porter 1.6.0: Porter Now Speaks MCP"
date: 2026-07-30
tags: [release, mcp, dependencies]
---

# Porter 1.6.0: Porter Now Speaks MCP

Porter 1.6.0 is out, and the headline feature is one we're genuinely excited about: **Porter now speaks MCP.**

If you've been packaging your apps, tools, and deploy logic into Porter bundles, you already know the value of having one distributable, versioned unit instead of a pile of scripts. This release makes that unit easier to work with than ever — including, for the first time, by an AI assistant sitting right next to your terminal.

## `porter mcp`: an AI assistant that actually knows your bundles

Running `porter mcp` starts an MCP (Model Context Protocol) server over stdio. If you're using an AI coding assistant that supports MCP — Claude Code, Cursor, and VS Code Copilot all do — that assistant can now talk directly to Porter instead of guessing from your `porter.yaml`.

In practice, that means you can ask things like:

- "What does this bundle actually install?"
- "Show me the installations running in the `prod` namespace."
- "Walk me through what would change if I upgrade this bundle."

...and get answers grounded in what Porter actually knows, not an educated guess based on static file contents. For anyone who's ever had an assistant hallucinate a plausible-sounding but wrong answer about a deploy script, this is the fix: the assistant is querying Porter's real state, the same way you would from the CLI.

By default, the server only exposes read-only tools — safe to point an assistant at without worrying it'll change anything. If you want your assistant to actually install, upgrade, uninstall, or invoke bundles on your behalf, start the server with `--allow-write` to opt in to those mutating tools explicitly.

We think this is a meaningful shift in how bundles get explored and operated day-to-day, and it's just the first step — expect more MCP-native workflows in future releases.

**Try it now:**

```
# read-only: explore and inspect bundles and installations
porter mcp

# read-write: also allow install, upgrade, uninstall, and invoke
porter mcp --allow-write
```

Point an MCP-compatible client at it — for Claude Desktop, that's:

```json
{"command": "porter", "args": ["mcp"]}
```

Then ask it about a bundle you're already working with — that's the fastest way to see what's changed.

## Also new in 1.6.0

MCP is the headline, but this release carries a handful of other improvements worth knowing about:

**Dependency version ranges.** When a dependency declares a version range instead of a pinned tag, `dependencies.version-strategy` now controls how Porter resolves it: `exact` (the default — refuse to resolve ranges, and error if no pinned tag is given), `max-patch`, `max-minor`, or `min`. Set it in the config file, override it with `PORTER_DEPENDENCIES_VERSION_STRATEGY`, or per-command with `--dependencies-version-strategy`. Full details in the [Configuration docs](https://porter.sh/docs/configuration/configuration/#dependency-version-strategy).

**Experimental file sources.** `porter.yaml` can now declare files that Porter downloads over HTTP/HTTPS at build time, rather than requiring everything to be pre-placed in the bundle directory. This requires `schemaVersion: 1.3.0` and the `file-sources` experimental flag — and, since downloading arbitrary URLs into a bundle build has real security implications, it's blocked by default even then. You also need to explicitly set `allow-file-downloads: true` (or pass `--allow-file-downloads` to `porter build`), and should only do so for bundles whose declared URLs you trust. Still experimental, and we'd love feedback from anyone who tries it against a real bundle.

**Graceful shutdown.** Hitting Ctrl+C during a Porter operation now forwards a proper SIGTERM with a grace period, instead of killing things abruptly. If your bundle has cleanup steps or outputs that need to finish writing, this means far less risk of losing them to an interrupted run.

**Reliability fixes**, including:
- Correct parameter-set precedence when running an upgrade
- `porter explain` now shows parameters that were injected automatically, not just the ones you specified
- A fix for `porter-state` corruption that could occur with custom actions marked `modifies: false`

Full details, as always, are in the [GitHub release notes](https://github.com/getporter/porter/releases/tag/v1.6.0).

## Where dependencies are headed next

If you work with Porter's dependency system, you've probably felt its current limits: every dependency creates a brand-new installation, even when you'd rather point at something that already exists — a shared database, a cluster-wide operator, infrastructure someone stood up by hand outside Porter entirely.

We're actively working on **Dependencies v2**, a more flexible model that supports sharing an existing installation across multiple bundles, scoping that sharing by group or namespace, and depending on *any* bundle that satisfies a contract (think "any bundle providing MySQL 5.7") rather than a hardcoded reference. The design is laid out in [PEP003 - Advanced Dependencies](https://github.com/getporter/proposals/blob/main/pep/003-advanced-dependencies.md).

The `dependencies-v2` experimental flag exists today, but the functionality behind it is still being built out incrementally — so turning it on gets you a work-in-progress, not the complete feature set described in the PEP just yet. If you want to follow along closely or try what's there so far, you can enable it using any of Porter's standard ways to set an [experimental feature flag](https://porter.sh/docs/configuration/configuration/#experimental-feature-flags) — a CLI flag, an environment variable, or the config file. Expect rough edges and gaps, and consider anything you see subject to change.

One concrete way to see it in action already: `porter inspect --show-dependencies` resolves and displays a bundle's full *transitive* dependency graph, not just its direct dependencies — shared dependencies get deduplicated to a single node, and for Dependencies v2 (shared) bundles specifically, it detects and shows the output-wiring relationships between sibling dependencies. It's one of the clearest windows into what the shared model actually does today. Worth noting: execution (`install`/`upgrade`/`uninstall`) currently only resolves a bundle's *direct* dependencies — transitive dependencies aren't automatically installed yet, though it's on our backlog.

The version-range work in this release is a small but real piece of that larger direction — we'll follow up with a full write-up once Dependencies v2 is further along.

## Feedback welcome

Found a bug? Please open an [issue on GitHub](https://github.com/getporter/porter/issues). For questions, ideas, and "here's what I wish this did" feedback, [GitHub Discussions](https://github.com/getporter/porter/discussions) or `#porter` on the CNCF Slack are the best places to reach us.
