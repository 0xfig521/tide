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
- **Manage categories** — `tide category`
- **View subscriptions** — `tide sources`
- **Manage the background daemon** — `tide schedule start|stop|status|logs`
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

### `tide fetch`

Fetch articles from feeds with concurrent workers.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--feed` | | `0` | Fetch specific feed by ID |
| `--category` | `-c` | | Fetch feeds in a category |
| `--concurrency` | `-n` | `5` | Number of concurrent workers |
| `--force` | `-f` | `false` | Force refresh (ignore cache interval) |

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

---

### `tide import <file>`

Import RSS feed subscriptions from an OPML 2.0 file. Supports nested category groups. Duplicate feeds are skipped silently.

---

### `tide export`

Export all RSS subscriptions as an OPML 2.0 file.

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output file path (default: stdout) |

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

### `tide upgrade`

Self-update tide to the latest version from GitHub Releases.

| Flag | Description |
|------|-------------|
| `--check` | Check if a new version is available |
| `--tag` | Install a specific version tag |

## Common Workflows

### Quick Setup
```bash
tide add "https://blog.golang.org/feed.atom" -c "Tech"
tide add "https://hnrss.org/frontpage" -c "News"
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
4. **Get full content**: Use `tide get <id>` for description and content. Default list/search output is lightweight.
5. **OPML import/export**: Use `tide import <file>` to migrate and `tide export -o <file>` for backup.
6. **Categories auto-create** when using `--category` with `tide add` or during OPML import.
7. **Feed IDs** are integers — use `tide sources` to find them.
8. **Schedule**: Use `tide schedule start` for automatic background fetching.

## Recommended AI Agent Workflow

```bash
tide fetch --quiet
tide search "rust async" --since 7d --limit 5
tide get 42
```
