<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://img.shields.io/badge/tide-111?style=flat&logo=terminal&logoColor=white&labelColor=111">
  <img alt="tide" src="https://img.shields.io/badge/tide-111?style=flat&logo=terminal&logoColor=white&labelColor=111">
</picture>

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/0xfig521/tide?style=flat)](https://github.com/0xfig521/tide/releases)

English | [中文](./README.zh.md)

---

A fast, concurrent RSS reader for the terminal. `tide` keeps your feeds in a SQLite database, fetches in parallel, and returns everything as JSON — easy to pipe, script, or browse.

## Features

- **⚡ Concurrent** — pulls dozens of feeds in parallel, progress bar included
- **📦 Zero deps** — single binary, SQLite embedded, no runtime dependencies
- **🗃️ Categories** — organize feeds, filter by category everywhere
- **🔍 Full-text search** — across titles, descriptions, and content
- **📡 Smart caching** — ETag / Last-Modified conditional requests, no wasted bandwidth
- **⏱️ Time filters** — `--since 24h`, `--since 7d`
- **📄 Pagination** — `--page`, `--page-size`
- **🤖 Daemon mode** — `tide fetch --daemon` runs in the background on a schedule

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
| `fetch --daemon` | Background scheduler |
| `read <id>` | Mark as read |
| `star <id>` | Bookmark / unbookmark |
| `category` | Manage categories (create/list/assign/remove) |
| `info <id>` | Feed details |

All commands output JSON by default. Use `--format table` on `list` for a terminal view.

## Powered by

- [gofeed](https://github.com/mmcdole/gofeed) • RSS/Atom/JSON Feed parser
- [SQLite](https://sqlite.org) • embedded database
- [cobra](https://github.com/spf13/cobra) • CLI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) • terminal styling

## License

[MIT](./LICENSE)
