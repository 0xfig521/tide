<pre align="center">
  ████████╗██╗██████╗ ███████╗
  ╚══██╔══╝██║██╔══██╗██╔════╝
     ██║   ██║██║  ██║█████╗
     ██║   ██║██║  ██║██╔══╝
     ██║   ██║██████╔╝███████╗
     ╚═╝   ╚═╝╚═════╝ ╚══════╝
</pre>

<p align="center"><em>RSS, reimagined for AI agents and the command line.</em></p>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/0xfig521/tide/releases"><img src="https://img.shields.io/github/v/release/0xfig521/tide?style=flat" alt="Release"></a>
</p>

<p align="center">English | <a href="./README.zh.md">中文</a></p>

**Tide** is an RSS data adapter built for AI agents and terminal users alike. It stores feeds in SQLite, fetches in parallel, and speaks one language: **JSON**. Every command returns a stable `{ok, data, error, meta}` envelope — clean stdout, progress on stderr, machine-readable error codes with non-zero exits. No parsing gymnastics, no grep hacks, no guesswork.

## Features

- **🧠 AI-native** — stable JSON envelope, structured error codes, clean stdout/stderr separation
- **⚡ Concurrent** — pulls dozens of feeds in parallel, progress bar writes to stderr
- **📦 Zero deps** — single binary, SQLite embedded, no runtime dependencies
- **🗃️ Categories** — organize feeds, filter by category everywhere
- **🔍 FTS5 search** — real full-text search with `MATCH`, not `LIKE %keyword%`
- **📡 Smart caching** — ETag / Last-Modified conditional requests
- **⏱️ Time filters** — `--since 24h`, `--since 7d`
- **📄 Pagination** — `--page`, `--page-size`
- **🤖 Daemon mode** — `tide schedule start` for scheduled background fetching
- **↔️ OPML** — `tide import` / `tide export` for migrating subscriptions between RSS readers
- **🔄 Delta tracking** — `tide changes` cursor mode, avoid reprocessing
- **📏 Token budgets** — `tide get --max-chars` / `--token-budget`, control LLM context costs
- **🔗 MCP integration** — `tide mcp` exposes Tide tools for AI agents via MCP protocol
- **🏥 Feed health** — `tide health` auto-detect dead/stale feeds
- **📤 RAG export** — `tide export entries` JSONL/Markdown for knowledge base ingestion
- **📐 Auto-routing** — `tide rule add` + `tide fetch --apply-rules` auto-classify entries
- **⚡ Performance-optimized** — batch transactional inserts, prepared statement reuse, lightweight list/search without content bloat

## Install

```bash
# macOS / Linux — one line
curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash

# Homebrew
brew install 0xfig521/tap/tide

# Go
go install github.com/0xfig521/tide@latest
```

## Quick Start

```bash
# Subscribe
tide add "https://blog.golang.org/feed.atom" --category "Tech"

# Fetch (10 concurrent workers)
tide fetch --concurrency 10

# Full-text search
tide search "kubernetes" --since 7d

# Get full content
tide get 42

# Pipe to jq
tide list --json | jq '.data.items[] | {title, feed_title}'
```

## For AI Agents

Tide speaks JSON — always. Every command returns a stable envelope:

```json
{"ok": true, "data": {...}, "error": null, "meta": null}
```

No parsing hacks required:

- **stdout** = clean JSON, always. **stderr** = progress bars, logs, diagnostics.
- **Exit code 0** = success, **non-zero** = failure. Check both `.ok` and exit code.
- **Error codes** are stable strings: `feed_not_found`, `entry_not_found`, `feed_already_exists`, `invalid_args`, `internal_error`.
- **`--quiet`** on `tide fetch` suppresses the progress bar for pristine output.

```bash
tide fetch --quiet                     # silent fetch, JSON result on stdout
tide search "rust async" --since 7d    # FTS5 search, last 7 days
tide get 42                            # full entry with description + content
```

Install the skill once and your agent knows every command:

```bash
npx skills add 0xfig521/tide
```

Full skill at [`tide/SKILL.md`](./tide/SKILL.md).

## Commands

| Command | Description |
|---|---|---|
| `add <url> [-c <cat>]` | Subscribe to a feed |
| `remove <id>` | Unsubscribe |
| `discover <url>` | Discover feeds from a website |
| `mark <id> --state <s>` | Set processing state (new/seen/processed/ignored/failed) |
| `sources` | List subscriptions |
| `import <file>` | Import feeds from OPML file |
| `export [--output <f>]` | Export feeds to OPML |
| `export entries` | Export entries (JSONL or Markdown) |
| `list` | Browse articles (JSONL default) |
| `search <kw>` | Full-text search (FTS5) |
| `get <id>` | Get entry details (--text, --max-chars, --token-budget) |
| `fetch [--force]` | Pull latest from feeds (--apply-rules) |
| `changes` | Get incremental new/changed entries |
| `health` | Show feed health status |
| `rule` | Manage auto-routing rules |
| `schedule` | Manage background daemon (start/stop/status/logs) |
| `category` | Manage categories (create/list/assign/remove) |
| `mcp` | Start MCP server (AI agent tool integration) |
| `upgrade` | Self-update to the latest version |

All commands output JSON by default (stable `{ok, data, error, meta}` envelope). Errors return non-zero exit codes with structured error codes.

## AI Agent Self-Setup Prompt

Paste the prompt below into your AI assistant (Claude, Codex, Cursor, etc.) to have it install and configure Tide automatically:

> You have access to the Shell tool. Please set up Tide on this machine following these steps:
>
> **Step 1 — Install Tide**
> ```bash
> curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash
> ```
> Verify the installation: `tide --version`
>
> **Step 2 — Install the Agent Skill** (if your platform supports skills)
> ```bash
> npx skills add 0xfig521/tide
> ```
> This gives AI assistants structured knowledge of all Tide commands and output formats.
>
> **Step 3 — Add Some Feeds**
> ```bash
> tide add "https://hnrss.org/frontpage" -c News
> tide add "https://blog.golang.org/feed.atom" -c Tech
> tide fetch --quiet
> ```
>
> **Step 4 — (Optional) Start MCP Server**
> If you are an MCP-compatible agent (Claude Desktop, Cursor, VS Code), run Tide in MCP mode to call tools directly:
> ```bash
> tide mcp &
> # This exposes these tools: discover_feeds, add_feed, fetch_feeds,
> # search_entries, list_entries, get_entry, mark_entry, get_feed_health
> ```
>
> **Step 5 — Verify**
> ```bash
> tide list --since 24h --limit 5
> tide health
> ```
>
> Once confirmed, I'll have Tide ready as my RSS data pipeline. I can fetch, search, summarize, and track content without writing ad-hoc RSS parsers.

## Scheduled Fetching

Tide can run as a background daemon that automatically fetches feeds on a schedule.

```bash
# Start the daemon (default: every 30 minutes, 5 workers)
tide schedule start

# Custom interval and concurrency
tide schedule start --interval 1h --concurrency 10

# Check daemon status
tide schedule status

# View recent logs
tide schedule logs -n 20

# Stop the daemon
tide schedule stop
```

The daemon persists across terminal sessions. It writes a PID file and logs to `~/.local/share/tide/logs/`.

## OPML Import & Export

Migrate subscriptions between RSS readers using standard OPML 2.0 format.

```bash
# Import feeds from another RSS reader
tide import feeds.opml

# Export all subscriptions for backup or migration
tide export -o tide-backup.opml

# Export to stdout (for piping)
tide export
```

`tide import` preserves category structure and feed metadata (title, site URL). Duplicate feeds are skipped automatically.

## Upgrading

```bash
# Check for new version
tide upgrade --check

# Upgrade to latest
tide upgrade

# Install a specific version
tide upgrade --tag v0.2.0
```

Tide downloads prebuilt binaries from GitHub Releases and self-replaces.

## Powered by

- [gofeed](https://github.com/mmcdole/gofeed) • RSS/Atom/JSON Feed parser
- [SQLite](https://sqlite.org) • embedded database
- [cobra](https://github.com/spf13/cobra) • CLI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) • terminal styling

## License

[MIT](./LICENSE)
