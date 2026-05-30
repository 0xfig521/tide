<pre align="center">
  ████████╗██╗██████╗ ███████╗
  ╚══██╔══╝██║██╔══██╗██╔════╝
     ██║   ██║██║  ██║█████╗
     ██║   ██║██║  ██║██╔══╝
     ██║   ██║██████╔╝███████╗
     ╚═╝   ╚═╝╚═════╝ ╚══════╝
</pre>

<p align="center"><em>A fast, concurrent RSS reader for the terminal.</em></p>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/0xfig521/tide/releases"><img src="https://img.shields.io/github/v/release/0xfig521/tide?style=flat" alt="Release"></a>
</p>

<p align="center">English | <a href="./README.zh.md">中文</a></p>

A fast, concurrent RSS reader for the terminal. `tide` keeps your feeds in a SQLite database, fetches in parallel, and returns everything as JSON — easy to pipe, script, or browse.

## Features

- **⚡ Concurrent** — pulls dozens of feeds in parallel, progress bar included
- **📦 Zero deps** — single binary, SQLite embedded, no runtime dependencies
- **🗃️ Categories** — organize feeds, filter by category everywhere
- **🔍 Full-text search** — across titles, descriptions, and content
- **📡 Smart caching** — ETag / Last-Modified conditional requests, no wasted bandwidth
- **⏱️ Time filters** — `--since 24h`, `--since 7d`
- **📄 Pagination** — `--page`, `--page-size`
- **🤖 Daemon mode** — `tide schedule start` runs the fetcher in the background on a schedule

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

# Browse unread from last 24h
tide list --unread --since 24h

# Full-text search
tide search "kubernetes"

# Mark read, star
tide read 3
tide star 7

# Pipe to jq
tide list --unread | jq '.items[] | {title, feed_title}'
```

## Commands

| Command | Description |
|---|---|
| `add <url> [-c <cat>]` | Subscribe to a feed |
| `remove <id>` | Unsubscribe |
| `sources` | List subscriptions |
| `list` | Browse articles (filters, pagination, time range) |
| `search <kw>` | Full-text search |
| `unread` | Unread articles |
| `fetch [--force]` | Pull latest from feeds |
| `schedule` | Manage background daemon (start/stop/status/logs) |
| `read <id>` | Mark as read |
| `star <id>` | Bookmark / unbookmark |
| `category` | Manage categories (create/list/assign/remove) |
| `info <id>` | Feed details |

All commands output JSON by default. Use `--format table` on `list` for a terminal view.

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

## AI Skill

Tide ships with a [skill](https://skills.sh/) for AI coding agents (Claude Code, Codex, Cursor, etc.). Install it once and your agent can manage RSS feeds for you:

```bash
npx skills add 0xfig521/tide
```

The skill gives AI agents full knowledge of every tide command, flag, and workflow — so you can say "find me the top 5 unread articles about Rust this week" and it just works.

See [`tide/SKILL.md`](./tide/SKILL.md) for the full skill definition.

## Powered by

- [gofeed](https://github.com/mmcdole/gofeed) • RSS/Atom/JSON Feed parser
- [SQLite](https://sqlite.org) • embedded database
- [cobra](https://github.com/spf13/cobra) • CLI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) • terminal styling

## License

[MIT](./LICENSE)
