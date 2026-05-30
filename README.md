# tide

> A sleek, high-speed RSS reader for the terminal.

**tide** brings your information flow into the command line. Add feeds, organize by category, search, and read — all with clean JSON output that plays nice with pipes and scripts.

---

## Why tide?

- **All JSON, all the time** — every command outputs structured JSON. Pipe to `jq`, `fzf`, or your own scripts.
- **Concurrent fetching** — pulls dozens of feeds in parallel. Progress bar included.
- **Smart caching** — ETag-aware conditional requests. No wasted bandwidth on unchanged feeds.
- **Categories** — organize feeds your way. Filter by category when listing or fetching.
- **Pagination + time filters** — `--page`, `--page-size`, `--since 24h`. Find what you need fast.
- **Daemon mode** — `tide fetch --daemon` runs in the background, pulling feeds on a schedule.

---

## Install

```bash
# One-liner (recommended)
curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash

# Homebrew
brew install ./Formula/tide.rb

# Go toolchain
go install github.com/0xfig521/tide@latest
```

The shell script auto-detects OS/arch, downloads the latest binary, installs to `/usr/local/bin` (or `~/.local/bin` via `TIDE_INSTALL_DIR`).

---

## Quick Start

```bash
# Add your first feed
tide add "https://blog.golang.org/feed.atom" --category "Tech"

# Pull articles (10 concurrent workers)
tide fetch --concurrency 10

# List unread articles from the last 24 hours
tide list --unread --since 24h

# Search
tide search "kubernetes"

# Mark as read, star a favorite
tide read 3
tide star 7

# Pipe anywhere
tide list --unread | jq '.items[] | {title, feed_title}'
```

---

## Commands

| Command | What it does |
|---------|-------------|
| `add <url> --category <name>` | Subscribe to a feed |
| `remove <id>` | Unsubscribe |
| `sources` | List all subscriptions |
| `list` | Browse articles with filters, pagination, time range |
| `search <keyword>` | Full-text search across all articles |
| `unread` | List unread articles |
| `fetch` | Pull latest articles from feeds |
| `fetch --daemon` | Run as a background scheduler |
| `read <id>` | Mark as read |
| `star <id>` | Bookmark / unbookmark |
| `category create/list/assign/remove` | Organize feeds |
| `info <id>` | Feed details |

---

## Powered by

- [gofeed](https://github.com/mmcdole/gofeed) — RSS/Atom/JSON Feed parsing
- [SQLite](https://sqlite.org) — embedded database, zero setup
- [cobra](https://github.com/spf13/cobra) — clean CLI
