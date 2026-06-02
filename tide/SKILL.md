---
name: tide
description: RSS data adapter for AI agents. Use when managing RSS feeds, fetching articles, searching feed content, or reading RSS articles. All output is JSON with a stable {ok, data, error, meta} envelope — stdout is always clean, errors are structured, exit codes are non-zero on failure.
metadata:
  author: 0xfig521
  version: "1.0.0"
  source: https://github.com/0xfig521/tide
---

# Tide

Tide is an RSS data adapter for AI agents and terminal users. It stores feeds in SQLite, fetches in parallel, and returns everything as a stable JSON envelope: `{ok, data, error, meta}`. Progress and diagnostics go to stderr. Errors are structured with machine-readable codes and non-zero exit codes.

## Trigger

Use this skill when the user asks to:

- **Get full entry details** — `tide get <id>` (supports --text, --max-chars, --token-budget)
- **Subscribe to RSS feeds** — `tide add <url>`
- **Batch subscribe** — `tide batch-add [file]` (JSON array via file or stdin)
- **Discover feeds from a website** — `tide discover <url>`
- **Import/Export OPML** — `tide import <file>` / `tide export [--output <file>]`
- **Export entries** — `tide export entries` (JSONL or Markdown, with filters)
- **Fetch articles** — `tide fetch` (supports --apply-rules for auto-classification)
- **Browse or list articles** — `tide list` with filters
- **Search feed content** — `tide search <keyword>`
- **Get incremental changes** — `tide changes --after <cursor>`
- **Mark entries as processed** — `tide mark <id> --state processed`
- **Check feed health** — `tide health`
- **Inspect failed sources** — `tide failures list` / `tide failures inspect <id>`
- **Clear or retry failed sources** — `tide failures clear [id] --yes` / `tide failures retry <id>`
- **Manage auto-routing rules** — `tide rule add|list|remove|apply`
- **Manage categories** — `tide category`
- **View subscriptions** — `tide sources`
- **Manage the background daemon** — `tide schedule start|stop|status|logs`
- **Self-update tide** — `tide upgrade` or `tide upgrade --check`
- **Start MCP server** — `tide mcp` (for AI agent tool integration)
- **Pipe RSS data to jq or other tools** — all output is JSON

## Installation

The `tide` binary must be installed before use. If not found, guide the user:

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash

# Homebrew
brew install 0xfig521/tap/tide

# Go
go install github.com/0xfig521/tide@latest
```

## Global Flags

All commands accept these persistent flags:

| Flag | Short | Description |
|------|-------|-------------|
| `--db` | `-d` | Database path (default: `~/.tide/tide.db`) |
| `--data-dir` | | Data directory (overrides `--db`) |
| `--version` | | Print version |

## Commands Reference

### `tide add <url>`

Subscribe to an RSS feed.

| Flag | Short | Description |
|------|-------|-------------|
| `--category` | `-c` | Assign feed to a category (auto-creates if needed) |

---

### `tide batch-add [file]`

Subscribe to multiple RSS feeds from a JSON array (file or stdin). Each element can be a plain URL string or an object with `"url"` and optional `"category"`. Categories auto-create if needed. Duplicates are skipped.

**Examples:**
```bash
# From file
tide batch-add feeds.json

# From stdin (ideal for AI agents)
echo '[{"url":"https://example.com/feed.xml","category":"tech"},{"url":"https://other.com/rss"}]' | tide batch-add

# Plain URL strings also work
echo '["https://blog.golang.org/feed.atom","https://hnrss.org/frontpage"]' | tide batch-add
```

**JSON input format:**
```json
[
  "https://blog.golang.org/feed.atom",
  {"url": "https://example.com/feed.xml", "category": "tech"},
  {"url": "https://hnrss.org/frontpage", "category": "news"}
]
```

**Output** returns a summary with per-feed results:
```json
{
  "ok": true,
  "data": {
    "total": 3,
    "imported": 2,
    "skipped": 1,
    "errored": 0,
    "results": [
      {"url": "...", "status": "imported", "id": 1, "title": "...", "category": "tech"},
      {"url": "...", "status": "imported", "id": 2, "title": "...", "category": "news"},
      {"url": "...", "status": "skipped", "reason": "already_exists", "id": 3}
    ]
  }
}
```

---

### `tide discover <url>`

Discover RSS/Atom feeds from a website URL. Scans the page HTML and common feed paths.

**Output** returns a list of discovered feeds with URL, type, and title:
```json
{"ok":true,"data":{"site_url":"...","feeds":[{"url":"...","type":"rss","title":"..."}]}}
```

**Examples:**
```bash
tide discover https://example.com
tide discover https://blog.golang.org
```

---

### `tide fetch`

Fetch articles from feeds with concurrent workers.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--feed` | | `0` | Fetch specific feed by ID |
| `--category` | `-c` | | Fetch feeds in a category |
| `--concurrency` | `-n` | `5` | Number of concurrent workers |
| `--force` | `-f` | `false` | Force refresh (ignore cache interval) |
| `--quiet` | | `false` | Suppress progress bar (pristine stdout) |
| `--fail-fast` | | `false` | Stop immediately on first fetch error |
| `--apply-rules` | | `false` | Apply routing rules to newly fetched entries |

---

### `tide changes`

Get new/changed entries since the last call (incremental cursor mode). Ideal for agent workflows that need to track what's new.

| Flag | Default | Description |
|------|---------|-------------|
| `--after` | `""` | Cursor to fetch changes after (default: auto-detect from history) |
| `--limit` | `50` | Maximum entries to return |

**Output** returns a new cursor for the next call:
```json
{"ok":true,"data":{"cursor":"2026-06-01 10:00:00:entry_42","items":[...],"count":5}}
```

Pass the `cursor` value from the response to `--after` on the next call to resume.

**Examples:**
```bash
tide changes
tide changes --after "2026-06-01 10:00:00:entry_42" --limit 20
```

---

### `tide mark <entry-id>`

Set processing state on an entry for agent workflow tracking. Valid states: `new`, `seen`, `processed`, `ignored`, `failed`.

| Flag | Default | Description |
|------|---------|-------------|
| `--state` | (required) | Processing state: `new`, `seen`, `processed`, `ignored`, `failed` |
| `--tag` | | Optional comma-separated tags (e.g., `summarized,rust`) |
| `--note` | | Optional note string |

**Examples:**
```bash
tide mark 42 --state processed
tide mark 42 --state processed --tag summarized --note "Used in weekly digest"
tide mark 42 --state ignored
```

---

### `tide list`

List articles with filtering, pagination, and time range. **Default output is CSV** for compact AI context. Use `--json` for the JSON envelope.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--search` | | | Search keyword (FTS5) |
| `--category` | `-c` | | Filter by category name |
| `--feed` | | `0` | Filter by feed ID |
| `--since` | | | Time range: `1h`, `6h`, `12h`, `24h`, `3d`, `7d`, `14d`, `30d` |
| `--page` | `-p` | `1` | Page number |
| `--page-size` | | `20` | Articles per page |
| `--json` | | `false` | Output as JSON envelope instead of CSV |

**CSV columns** (default): `id,title,url,author,published_at,feed_id,feed_title,description,categories,guid`

**Note**: `content` field is NOT in CSV output — use `tide get <id>` for full content.

---

### `tide search <keyword>`

Full-text search (SQLite FTS5) across titles, descriptions, and content.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--category` | `-c` | | Filter by category |
| `--feed` | | `0` | Filter by feed ID |
| `--limit` | `-n` | `50` | Maximum results |

---

### `tide get <entry-id>`

Get full details of a single entry, including description and content.

| Flag | Default | Description |
|------|---------|-------------|
| `--full` | `false` | Include full content in output |
| `--content-only` | `false` | Output only content-related fields (title, url, description, content, author) |
| `--text` | `false` | Strip HTML tags from content before output |
| `--max-chars` | `0` | Truncate content to N characters (0 = disabled) |
| `--token-budget` | `0` | Truncate to fit ~N tokens (rough: N*4 chars; 0 = disabled) |

When `--text` is active, HTML tags are stripped before output. When truncation flags are set (`--max-chars` or `--token-budget`), the response includes `truncated`, `char_count`, and `estimated_tokens` fields.

**Examples:**
```bash
tide get 42
tide get 42 --full
tide get 42 --text --max-chars 4000
tide get 42 --token-budget 2000
```

---

### `tide health`

Show feed health status including staleness, failure rate, and entry activity.

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `jsonl` | Output format: `jsonl` (default) or `json` (envelope) |

Default output is JSONL with one feed per line:
```jsonl
{"feed_id":1,"title":"Go Blog","status":"healthy","success_rate_7d":1.0,"entries_7d":3}
{"feed_id":2,"title":"Old Feed","status":"stale","stale_days":120,"entries_7d":0}
```

**Status classification:**
- `healthy` — recent success, low error rate
- `stale` — no fetch in 7+ days
- `failing` — 3+ consecutive failures
- `dead` — 10+ consecutive failures, 30+ days stale
- `unknown` — never fetched

**Examples:**
```bash
tide health
tide health --format json
```

---

### `tide failures`

Manage RSS feed sources that have crossed the failure threshold (default: 3 consecutive fetch errors). Every failure is classified into a machine-readable reason type so AI agents can decide how to respond without re-parsing raw error strings.

| Subcommand | Description |
|---|---|
| `list` | Show currently failing feeds with last error |
| `inspect <id>` | Show full failure history for one feed |
| `clear [id]` | Hard-delete failing feeds (destructive) |
| `retry <id>` | Reset error count and retry immediately |

| Flag | Applies to | Description |
|---|---|---|
| `--threshold` | list, clear | Minimum `parsing_error_count` (default 3) |
| `--type` | list | Filter by last-failure type: `http_4xx`, `http_5xx`, `timeout`, `dns`, `tls`, `parse`, `unknown` |
| `--yes` / `-y` | clear (bulk) | Confirm bulk clear (required for destructive operation) |
| `--limit` | inspect | Max failure history rows (default 20) |
| `--format` | list, inspect, clear | Output format: `jsonl` (default), `json` |

**Failure types:**
- `http_4xx` — Client error (404 Not Found, 410 Gone — likely permanent)
- `http_5xx` — Server error (503 Busy — may be transient)
- `timeout` — Connection or read timeout (often transient)
- `dns` — DNS resolution failed (may resolve itself)
- `tls` — TLS/certificate error (usually needs human intervention)
- `parse` — Feed content was unparseable (XML/JSON/RSS/Atom parse failure)
- `unknown` — Unrecognized error (inspect the raw message)

**Examples:**
```bash
tide failures list
tide failures list --type http_4xx
tide failures list --threshold 5 --format json
tide failures inspect 3 --limit 10
tide failures clear 42                        # removes single feed
tide failures clear --yes                     # clears all failing feeds
tide failures clear --threshold 5 --yes       # clears with custom threshold
tide failures retry 3                         # resets error count
```

---

### `tide import <file>`

Import RSS feed subscriptions from an OPML 2.0 file. Supports nested category groups. Duplicate feeds are skipped silently.

---

### `tide export`

Export RSS data. Parent command with two modes:

**`tide export`** — Export all subscriptions as OPML 2.0.

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output file path (default: stdout) |

**`tide export entries`** — Export entries in machine-readable format.

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `jsonl` | Output format: `jsonl` (default) or `markdown` |
| `--since` | | Time range: `1h`, `6h`, `12h`, `24h`, `3d`, `7d`, `14d`, `30d` |
| `--state` | | Filter by processing state: `new`, `seen`, `processed`, `ignored`, `failed` |
| `--category` | | Filter by category |
| `--limit` | `100` | Maximum entries |
| `--output` | | Output file path (default: stdout) |

**JSONL output** includes full provenance (id, title, url, feed_id, feed_title, feed_url, published_at, hash, description, content).

**Markdown output** uses YAML frontmatter, suitable for weekly digests or knowledge base ingestion.

**Examples:**
```bash
tide export entries --since 7d --state unprocessed
tide export entries --format markdown --output weekly-report.md --since 7d --category AI
tide export entries --format jsonl --limit 200
```

---

### `tide sources`

List all RSS feed subscriptions. Alias: `tide feeds`.

| Flag | Short | Description |
|------|-------|-------------|
| `--category` | `-c` | Filter by category name |

---

### `tide remove <id>`

Unsubscribe from a feed by ID. This also removes all associated articles.

---

### `tide rule`

Manage automatic routing rules for entry classification. Rules match entries by field content (title, description, content, author, category) and apply actions (tag, set state, ignore).

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `add --match <regex> [--field title] [--action tag] [--value <v>] [--priority 0]` | Create a new rule |
| `list` | List all rules |
| `remove <id>` | Delete a rule |
| `apply <entry-id>` | Test rules against a specific entry |

**Match fields:** `title` (default), `description`, `content`, `author`, `category`

**Actions:** `tag` (default), `state`, `ignore`, `priority`, `category`

**Examples:**
```bash
# Auto-tag AI/agent related entries
tide rule add --match "AI|agent|MCP" --action tag --value ai

# Auto-ignore sponsored content
tide rule add --match "sponsored|advertisement" --action ignore

# High-priority feed
tide rule add --match ".*" --field title --action priority --value high --feed 12

# List and apply
tide rule list
tide rule apply 42

# Apply rules during fetch
tide fetch --apply-rules
```

---

### `tide category`

Manage feed categories. Subcommands:

| Subcommand | Description |
|------------|-------------|
| `create <name>` | Create a new category |
| `list` | List all categories |
| `assign <feed-id> <category-name>` | Assign a feed to a category |
| `remove <name>` | Remove a category |

---

### `tide schedule`

Manage the background daemon lifecycle.

| Subcommand | Description |
|------------|-------------|
| `start` | Start the daemon as a detached background process |
| `stop` | Gracefully stop the running daemon |
| `status` | Check if daemon is running |
| `logs [-n N]` | View recent log output |

---

### `tide mcp`

Start an MCP (Model Context Protocol) server over stdio. Enables AI agents to call Tide tools directly without shell commands.

**Registered MCP tools:**
- `discover_feeds` — Discover RSS feeds from a URL
- `add_feed` — Subscribe to a feed
- `fetch_feeds` — Fetch articles from feeds
- `search_entries` — Full-text search entries
- `list_entries` — List entries with filters
- `get_entry` — Get entry details
- `mark_entry` — Set processing state
- `get_feed_health` — Check feed health
- `list_failed_feeds` — List feeds that are persistently failing (with classified reason)
- `clear_failed_feeds` — Hard-delete failing feeds (bulk with confirm=true)

Use this to connect Tide with MCP-compatible AI clients (Claude, Codex, Cursor, etc.).

**Examples:**
```bash
tide mcp
```

---

### `tide upgrade`

Self-update tide to the latest version from GitHub Releases.

| Flag | Description |
|------|-------------|
| `--check` | Check if a new version is available |
| `--tag` | Install a specific version tag |

## Common Workflows

### Quick Setup
```bash
# Single feed
tide add "https://blog.golang.org/feed.atom" -c "Tech"

# Batch subscribe (AI-friendly)
echo '[
  {"url":"https://blog.golang.org/feed.atom","category":"Tech"},
  {"url":"https://hnrss.org/frontpage","category":"News"},
  {"url":"https://lwn.net/headlines/rss","category":"Tech"}
]' | tide batch-add

tide fetch --concurrency 10
tide list --since 24h
```

### Migrate from Another Reader
```bash
tide import feeds.opml
tide fetch
```

### Backup Subscriptions
```bash
tide export -o tide-backup.opml
```

### Search & Filter
```bash
tide search "kubernetes"
tide list --category Tech
tide get 42
```

### Scheduled Fetching
```bash
tide schedule start --interval 30m --concurrency 5
tide schedule status
tide schedule logs -n 20
```

## Tips for AI Agents

1. **All output is JSON with a standard envelope**: `{"ok": true/false, "data": ..., "error": {...}, "meta": null}`. Parse `.ok` first.
2. **Exit codes**: 0 = success, non-zero = failure.
3. **Error codes** are stable strings: `feed_not_found`, `entry_not_found`, `feed_already_exists`, `invalid_args`, `internal_error`.
4. **Get full content**: Use `tide get <id>`. For HTML-free output, add `--text`. For token budgets, add `--max-chars` or `--token-budget`.
5. **OPML import/export**: Use `tide import <file>` to migrate and `tide export -o <file>` for backup.
6. **Batch subscribe**: Use `tide batch-add` with a JSON array (via file or pipe) to add multiple feeds at once.
7. **Categories auto-create** when using `--category` with `tide add`, `tide batch-add`, or during OPML import.
8. **Feed IDs** are integers — use `tide sources` to find them.
9. **Schedule**: Use `tide schedule start` for automatic background fetching.
10. **Incremental changes**: Use `tide changes` for delta tracking. Pass the returned cursor to `--after` on the next call.
11. **Feed health**: Use `tide health` (JSONL default) to detect stale or failing feeds before processing.
12. **Export entries**: Use `tide export entries --format jsonl` for RAG pipelines or `--format markdown` for weekly digests.
13. **Auto-classify**: Use `tide rule add` to create routing rules, then `tide fetch --apply-rules` to apply them during fetch.
14. **MCP integration**: Run `tide mcp` to expose Tide tools to MCP-compatible AI agents (Claude, Codex, Cursor).
15. **Discover first**: Not sure about a feed URL? Try `tide discover <website-url>` to find feeds automatically.

## Recommended AI Agent Workflows

### Quick Research Session
```bash
tide fetch --quiet
tide search "rust async" --since 7d --limit 5
tide get 42 --text --max-chars 4000
```

### Incremental Monitoring
```bash
tide changes --limit 10
tide get 1 --text --token-budget 2000
tide search "AI agent" --since 24h
tide mark 1 --state processed --tag summarized
```

### Feed Quality Check
```bash
tide health | grep -E "stale|failing|dead"
tide sources --health
tide remove 12  # remove dead feed
```

### Weekly Export to Knowledge Base
```bash
tide export entries --since 7d --state unprocessed --format jsonl
tide mark 42 --state processed
```

### MCP Integration (for MCP-compatible agents)
```bash
# Run MCP server in background
tide mcp
# AI agent now has direct tool access to:
# discover_feeds, add_feed, fetch_feeds, search_entries,
# list_entries, get_entry, mark_entry, get_feed_health
```
