## 0.6.0 - 2026-06-01

### Features
- Add `tide batch-add` for bulk feed subscription from JSON files
- Add `tide changes` for delta/cursor-based incremental entry tracking
- Add `tide discover <url>` to auto-discover RSS/Atom feeds from a website
- Add `tide health` for feed health status (healthy/stale/failing/dead)
- Add `tide mark <id> --state <state>` for agent processing state tracking
- Add `tide mcp` to expose Tide as MCP tools for AI agent integration
- Add `tide rule add/list/remove` for auto-routing rules
- Add `EntryStateRepo` with SetState, SetStateWithTags, SetStateFull, ListByState
- Add `RuleRepo` with Apply method for regex-based entry routing
- Add schema versions v4 (entry_states), v5 (change_log), v6 (rules)
- Add `tide export entries` for JSONL/Markdown RAG export
- Add `tide get --content-only`, `--max-chars`, `--token-budget`, `--text` for LLM context control
- Structured error codes in output envelope for all commands

### Performance
- `list`/`search` no longer reads the `content` column — 13-field light scan reduces I/O and token cost
- `BatchInsertEntries` uses single transaction + prepared statement for bulk entry insertion
- Prepared statement reuse for `InsertOrIgnore`, `UpdateMeta`, `UpdateFetchResult`, `UpdateFetchError` (lazy-init via `sync.Once`)
- All fetch paths (cmd/fetch and worker) use batch inserts instead of per-row `InsertOrIgnore`
- Output response helpers with JSONL streaming / CSV support

### Documentation
- Add `docs/performance-optimization.md` — performance optimization roadmap with Phase 1 completed
- Add `docs/agent-native-rss-tool-design.md` — Agent-native RSS design specification
- Add `docs/agent-rss-tool-advanced-roadmap.md` — Advanced feature roadmap (MCP, delta, token budgets)
- Add `tide/SKILL.md` — full AI agent skill definition for Claude Code/Codex/Cursor
- Web landing page: add Performance Optimized bento card with throughput and token cost widgets
- README/README.zh: add performance optimization feature bullet

## 0.5.1 - 2026-05-31

### Breaking Changes
- Remove human-oriented commands: `read`, `unread`, `star`, `info` — Tide is now purely agent-oriented
- Remove `--unread`, `--starred`, `--format table` flags from `list` and `search` commands
- Drop `is_read` and `is_starred` columns from entries table (schema migration v3)
- Remove `UnreadCount` from feed output; `GetEntryCount` now returns only total count
- Remove `ListUnread`, `ListStarred`, `MarkRead`, `MarkUnread`, `ToggleStar`, `MarkAllRead` repo methods
- Remove `EntryTable`, `FeedTable`, `CategoryTable`, `PrintTable` from output package (no more human-facing tables)

### Features
- `tide list` now defaults to compact CSV output for minimal AI context usage; `--json` flag for full JSON
- CSV columns: id, title, url, author, published_at, feed_id, feed_title, description, categories, guid
- `content` field excluded from list CSV — use `tide get <id>` for full content
- `description`, `author`, `published_at`, `categories` fields always present in entry output

### Documentation
- Update web components (QuickStart, Hero, AISkill, Features) to reflect agent-only command set
- Update i18n translations to remove human-oriented terminology
- Sync README, README.zh, and SKILL.md with CSV default and removed commands

## 0.5.0 - 2026-05-30

### Features
- Add OPML 2.0 import (`tide import <file>`) — recursively parses category hierarchies, preserves feed metadata (title, site URL), skips duplicates with best-effort error reporting
- Add OPML 2.0 export (`tide export [-o <file>]`) — groups feeds by category, defaults to stdout, file output returns JSON confirmation
- OPML round-trip: import then export preserves title, xmlUrl, htmlUrl, and category assignments

### Tests
- Add 9 unit tests for OPML parser and generator (flat, nested, deep-nested, empty, no-xmlUrl, invalid XML, generate, round-trip, default title)

### Documentation
- Add OPML import/export to README, README.zh, and SKILL.md
- Add migration and backup workflows to docs

## 0.4.0 - 2026-05-30

### Features
- Standardize JSON contract across all 13 commands — stable `{ok, data, error, meta}` envelope on stdout, progress and diagnostics on stderr, structured error codes with non-zero exits
- Upgrade search from `LIKE %keyword%` to SQLite FTS5 with `MATCH` — real full-text search with triggers for insert/update/delete sync
- Add `tide get <entry-id>` command for full entry retrieval including description and content
- Redirect fetch progress bar to stderr and add `--quiet` flag for clean JSON output
- Convert `tide schedule` and `tide upgrade` from plain text to structured JSON output
- Add content fields (description, content, categories, guid) to entry output and `entryToOutput`

### Fixes
- Fix User-Agent header from legacy "GoRSS/1.0" to "Tide/1.0"
- Remove legacy `gorss` binary and stale build artifacts
- Isolate `web/` directory from Go toolchain to prevent `go list` scanning `node_modules`

### Tests
- Add 21 core path tests covering fetcher conversion, repo operations, and CLI JSON contract
- Test fixtures for RSS 2.0 and Atom feed parsing

### Documentation
- Rewrite README and SKILL for AI-agent-first positioning
- Add "For AI Agents" / "面向 AI Agent" sections with JSON contract and error handling guidance
- Add recommended AI agent workflow to SKILL.md
