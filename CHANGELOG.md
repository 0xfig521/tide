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
