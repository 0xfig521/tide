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

- **Get full entry details** — `tide get <id>` (includes description + content)
- **Subscribe to RSS feeds** — `tide add <url>`
- **Import/Export OPML** — `tide import <file>` / `tide export [--output <file>]`
- **Fetch articles** — `tide fetch`
- **Browse or list articles** — `tide list` with filters
- **Search feed content** — `tide search <keyword>`
- **Mark articles as read** — `tide read <id>`
- **Star/bookmark articles** — `tide star <id>`
- **Manage categories** — `tide category`
- **View subscriptions** — `tide sources`
- **Manage the background daemon** — `tide schedule start|stop|status|logs`
- **Run RSS fetching daemon** — `tide fetch --daemon` (low-level, prefer `tide schedule` for lifecycle management)
- **Self-update tide** — `tide upgrade` or `tide upgrade --check`
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

If the user cannot run `tide`, offer to help with installation first.

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

**Output** (JSON):
```json
{"ok": true, "data": {"id": 1, "feed_url": "...", "title": "..."}, "error": null, "meta": null}
```

**Error**: duplicate feed returns `{"ok": false, "error": {"code": "feed_already_exists", "message": "already exists"}}` with non-zero exit code.

---

### `tide fetch`

Fetch articles from feeds with concurrent workers. By default fetches all due feeds.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--feed` | | `0` | Fetch specific feed by ID |
| `--category` | `-c` | | Fetch feeds in a category |
| `--concurrency` | `-n` | `5` | Number of concurrent workers |
| `--force` | `-f` | `false` | Force refresh (ignore cache interval) |
| `--daemon` | | `false` | Run as daemon (continuous scheduler) |
| `--interval` | | `30m` | Daemon fetch interval |

**Daemon mode**: `tide fetch --daemon --interval 15m` runs a background scheduler.

---

### `tide list`

List articles with filtering, pagination, and time range. Default output is JSON.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--search` | | | Search keyword (full-text on title, description, content) |
| `--category` | `-c` | | Filter by category name |
| `--feed` | | `0` | Filter by feed ID |
| `--unread` | `-u` | `false` | Only unread articles |
| `--starred` | | `false` | Only starred articles |
| `--since` | | | Time range: `1h`, `6h`, `12h`, `24h`, `3d`, `7d`, `14d`, `30d` |
| `--page` | `-p` | `1` | Page number |
| `--page-size` | | `20` | Articles per page |
| `--format` | | | Output format: `table` (default: JSON) |

**Examples**:
```bash
tide list                                      # All articles, page 1
tide list --unread --since 24h                 # Unread from last 24h
tide list --search kubernetes --category tech  # Search within category
tide list --page 2 --page-size 50              # Pagination
tide list --format table                       # Terminal table view
```

**JSON Output**:
```json
{
  "ok": true,
  "data": {
    "items": [{ "id": 1, "title": "...", "url": "...", "feed_title": "...", "is_read": false, "is_starred": false }],
    "total": 42,
    "page": 1,
    "page_size": 20
  },
  "error": null,
  "meta": null
}
```

---

### `tide search <keyword>`

Alias for `tide list --search <keyword>`. Full-text search (SQLite FTS5) across titles, descriptions, and content.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--category` | `-c` | | Filter by category |
| `--feed` | | `0` | Filter by feed ID |
| `--unread` | | `false` | Only unread entries |
| `--starred` | | `false` | Only starred entries |
| `--limit` | `-n` | `50` | Maximum results |

---

### `tide get <entry-id>`

Get full details of a single entry, including description and content.

**Output** (JSON envelope):
```json
{
  "ok": true,
  "data": {
    "id": 42,
    "title": "Introducing Tide",
    "url": "https://example.com/post",
    "author": "Jane Doe",
    "published_at": "2026-05-30 10:00:00",
    "feed_title": "Go Blog",
    "feed_id": 1,
    "description": "A new RSS reader for the terminal...",
    "content": "<p>Full article HTML content...</p>",
    "categories": "golang,cli",
    "guid": "https://example.com/post/42",
    "is_read": false,
    "is_starred": false
  },
  "error": null,
  "meta": null
}
```

**Error**: entry not found returns `{"ok": false, "error": {"code": "entry_not_found", "message": "entry not found"}}` with non-zero exit code.

---

### `tide unread`

Alias for `tide list --unread`. Quick shortcut for unread articles.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--category` | `-c` | | Filter by category |
| `--limit` | `-n` | `50` | Maximum entries |

---

### `tide read <id>`

Mark an article as read by ID.

**Output**: `{"ok": true, "data": {"id": 1, "read": true}, "error": null, "meta": null}`

---

### `tide star <id>`

Toggle star/bookmark on an article. Calling on a starred article un-stars it.

**Output**: `{"ok": true, "data": {"id": 1, "starred": true}, "error": null, "meta": null}`

---

### `tide import <file>`

Import RSS feed subscriptions from an OPML 2.0 file. Supports nested category groups. Duplicate feeds are skipped silently.

**Output** (JSON summary):
```json
{
  "ok": true,
  "data": {
    "imported": 12,
    "skipped": 3,
    "errors": []
  },
  "error": null,
  "meta": null
}
```

**Error codes**: `invalid_args` (file not found or invalid OPML).

---

### `tide export`

Export all RSS feed subscriptions as an OPML 2.0 file. By default writes OPML XML to stdout. Use `--output` to write to a file instead.

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output file path (default: stdout) |

When `--output` is specified, returns a JSON success envelope:
```json
{"ok": true, "data": {"file": "feeds.opml", "feeds": "15"}, "error": null, "meta": null}
```

When writing to stdout, the raw OPML XML is output directly (for piping to files or other tools).

---

### `tide sources`

List all RSS feed subscriptions. Alias: `tide feeds`.

| Flag | Short | Description |
|------|-------|-------------|
| `--category` | `-c` | Filter by category name |

**Output** (JSON envelope with array of feed objects):
```json
{
  "ok": true,
  "data": [
    {
      "id": 1,
      "title": "Go Blog",
      "feed_url": "https://blog.golang.org/feed.atom",
      "site_url": "...",
      "description": "...",
      "categories": ["Tech"],
      "entry_count": 150,
      "unread_count": 12,
      "last_fetched": "2026-05-30 15:00:00",
      "is_active": true
    }
  ],
  "error": null,
  "meta": null
}
```

---

### `tide info <id>`

Show detailed info for a specific feed by ID.

---

### `tide remove <id>`

Unsubscribe from a feed by ID. This also removes all associated articles.

---

### `tide category`

Manage feed categories. Subcommands:

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `create` | `tide category create <name> [--desc "..."]` | Create a new category |
| `list` | `tide category list` | List all categories |
| `assign` | `tide category assign <feed-id> <category-name>` | Assign a feed to a category |
| `remove` | `tide category remove <name>` | Remove a category |

---

### `tide schedule`

Manage the background daemon lifecycle. This is the recommended way to run continuous fetching.

| Subcommand | Usage | Description |
|------------|-------|-------------|
| `start` | `tide schedule start [--interval 30m] [--concurrency 5]` | Start the daemon as a detached background process |
| `stop` | `tide schedule stop` | Gracefully stop the running daemon |
| `status` | `tide schedule status` | Check if daemon is running (shows PID and uptime) |
| `logs` | `tide schedule logs [-n 50] [--clear]` | View or clear daemon log output |

**Start flags**:
| Flag | Default | Description |
|------|---------|-------------|
| `--interval` | `30m` | Fetch interval (e.g. `1h`, `5m`) |
| `--concurrency` | `5` | Number of concurrent workers |

**Logs flags**:
| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--lines` | `-n` | `0` (all) | Show last N lines |
| `--clear` | | `false` | Clear the log file |

**Behavior**:
- The daemon persists across terminal sessions (detached process with its own session)
- PID file stored at `~/.local/share/tide/daemon.pid`
- Logs stored at `~/.local/share/tide/logs/daemon.log`
- Starting while already running is rejected
- Stopping when not running reports cleanly
- Uses the same `--data-dir` or `--db` flags for database path

---

### `tide upgrade`

Self-update tide to the latest version from GitHub Releases.

| Flag | Short | Description |
|------|-------|-------------|
| `--check` | | Check if a new version is available (no install) |
| `--tag` | | Install a specific version tag (e.g. `v0.2.0`) |

**Examples**:
```bash
tide upgrade              # Upgrade to latest
tide upgrade --check      # Check for updates
tide upgrade --tag v0.1.2 # Downgrade to a specific version
```

**Behavior**:
- Downloads the appropriate prebuilt binary (`tide-{os}-{arch}.tar.gz`) from GitHub Releases
- Verifies the downloaded binary before installing
- Creates a backup (`.old`) before replacing, removes on success
- Reports "Already up to date" if current version matches latest

## Common Workflows

### Quick Setup (New User)
```bash
tide add "https://blog.golang.org/feed.atom" -c "Tech"
tide add "https://hnrss.org/frontpage" -c "News"
tide fetch --concurrency 10
tide list --unread --format table
```

### Migrate from Another Reader
```bash
# Import OPML from another RSS reader
tide import feeds.opml
tide fetch
```

### Backup Subscriptions
```bash
# Export feeds to OPML file for backup or migration
tide export -o my-feeds.opml
```

### Daily Reading
```bash
tide fetch && tide list --unread --since 24h --format table
```

### Search & Filter
```bash
tide search "kubernetes"           # Full-text search
tide list --unread --category Tech  # Filter by category
tide list --starred                 # Your bookmarks
```

### Scripting with jq
```bash
tide list --unread | jq '.items[] | {title, feed_title}'
tide sources | jq '.[] | select(.unread_count > 0) | {title, unread_count}'
```

### Set Up Scheduled Fetching
```bash
tide schedule start --interval 30m --concurrency 5
tide schedule status       # Check it's running
tide schedule logs -n 20   # Verify it's fetching
```

### Upgrade
```bash
tide upgrade --check       # Check for new version
tide upgrade               # Install latest
```

### Manual Daemon (Advanced)
```bash
tide fetch --daemon --interval 30m --concurrency 5
```

## Tips for AI Agents

1. **All output is JSON with a standard envelope**: `{"ok": true/false, "data": ..., "error": {"code": "...", "message": "..."}, "meta": null}`. Parse `.ok` first — if `false`, check `.error.code`.
2. **Exit codes signal success/failure**: exit code 0 = success, non-zero = failure. Always check exit code alongside `.ok`.
3. **Error codes** are stable machine-readable strings: `feed_not_found`, `entry_not_found`, `feed_already_exists`, `already_exists`, `invalid_args`, `internal_error`.
4. **Always check if `tide` is installed** before running commands. If `command not found`, guide the user to install it first.
5. **The `fetch` command is the first step** when the user wants fresh content. Use `--quiet` to suppress the progress bar for clean JSON output. For continuous fetching, use `tide schedule start` instead of `tide fetch --daemon`.
6. **Use `--format table`** when the user wants human-readable output instead of JSON.
7. **The `--since` flag** supports: `1h`, `6h`, `12h`, `24h`, `3d`, `7d`, `14d`, `30d`.
8. **Categories are auto-created** when using `--category` with `tide add` or during OPML import.
9. **Feed IDs** are integers. Use `tide sources` to find them.
10. **Schedule management**: Use `tide schedule start` to set up automatic fetching. Check `tide schedule status` to verify it's running, `tide schedule logs` for troubleshooting.
11. **Self-update**: `tide upgrade --check` before proposing commands to ensure the user has the latest features. `tide upgrade` handles cross-version updates safely.
12. **Get full content**: Use `tide get <id>` to retrieve description and content for summarization. The default list/search output is lightweight — explicit retrieval is needed for full text.
13. **OPML import**: Use `tide import <file>` to migrate subscriptions from other RSS readers. Duplicates are skipped automatically. The command returns a JSON summary.
14. **OPML export**: Use `tide export -o <file>` for backup. Without `-o`, OPML XML is written to stdout.

## Recommended AI Agent Workflow

```bash
# 1. Fetch fresh content (suppress progress bar for clean JSON)
tide fetch --quiet

# 2. Search for articles
tide search "rust async" --since 7d --limit 5

# 3. Get full content for summarization
tide get 42

# 4. Mark as read
tide read 42

# 5. Check for errors
# All commands: parse .ok, check .error.code, verify exit code
```
