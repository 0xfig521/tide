# Tide — Agent Guide

## Quick start

```bash
go build .          # produces tide binary
go test ./...        # all tests
go vet ./...         # static checks
make lint            # fmt + vet
```

No CGO (`CGO_ENABLED=0` enforced in releases). Go 1.25.5.

## Single binary entrypoint

```
main.go → cmd/root.go → cobra subcommands in cmd/*.go
```

Commands in `cmd/` are self-contained (init + RunE per file). New commands: add file in `cmd/`, register via `rootCmd.AddCommand()`.

## Output contract (mandatory)

- **stdout** = JSON only, stable `{"ok":bool,"data":any,"error":{code,message},"meta":any}` envelope
- **stderr** = progress bars, logs, diagnostics, warnings
- **Exit code 0** = success, **non-zero** = failure (return `output.CmdError` from RunE)
- Error codes: `output.CodeFetchFailed`, `CodeFeedNotFound`, `CodeInvalidArgs`, `CodeInternalError`, etc.
- Every RunE must return either `output.PrintSuccess(data, meta)` or `output.PrintError(code, msg)`

3 output formats: `jsonl` (default for list-like), `json` (envelope), `csv` (via `output.PrintCSV`).

## Database

- SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- WAL mode, `busy_timeout=5000`, `foreign_keys=ON`
- `db.SetMaxOpenConns(1)` — single writer
- Default path: `~/.local/share/tide/tide.db` or `$XDG_DATA_HOME/tide/tide.db`
- Embedded migrations in `internal/db/migrate.go`: versioned integer keys, auto-applied on `db.Open()`
- Time format in SQL: `"2006-01-02 15:04:05"` (Go constant, used everywhere)

## Architecture

```
cmd/      → cobra commands (thin dispatch)
internal/
  db/         → SQLite open + migration runner
  models/     → data types (Feed, Entry, Category, FeedFailure, FailureType)
  repo/       → data access (feed_repo, entry_repo, category_repo, rule_repo, feed_failure_repo)
  fetcher/    → RSS fetch: parser, worker pool, scheduler
  output/     → JSON envelope, CSV, terminal styling (lipgloss)
  opml/       → OPML import/export
  schedule/   → daemon process (PID file, log management)
pkg/
  hash.go     → sha256(feedID + guid) for entry dedup
web/          → separate React+vite+tailwind+gsap site, not relevant to CLI
```

## Key constraints

- **Import cycle guard**: `fetcher` imports `repo` → `repo` must NOT import `fetcher`. Shared logic that both need (like error classification) goes in `models/`.
- **SQLite single writer**: Every connection pool uses `SetMaxOpenConns(1)`. Don't change this — SQLite cannot safely parallel-write.
- **Prepared statements**: Lazily initialized via `sync.Once` per statement in repo layer (pattern: `prepareXxx()` + `xxxOnce`).
- **Version injected at build**: `-X github.com/0xfig521/tide/cmd.version={{.Version}}` via goreleaser. The `version` variable lives in `cmd/root.go`.

## Testing

- `setupTestDB(t)` creates a temp-directory SQLite DB with all migrations applied — pattern in `internal/repo/entry_repo_test.go`
- Use `t.Cleanup(func() { database.Close() })` for cleanup
- Tests live in `_test.go` files alongside the production code; no separate test directory
- OPML tests in `internal/opml/opml_test.go`

## Notable features that exist

- FTS5 full-text search on entries (auto-synced via triggers in migration v2)
- Entry state tracking: `entry_states` table (new/seen/processed/ignored/failed)
- Change log: `change_log` table for event sourcing (`entry_created`, `feed_failed`, `feed_recovered`, `state_changed`)
- Auto-classification rules: `rules` table, applies on fetch with `--apply-rules`
- Failed-source management: `feed_failures` table (migration v7), `tide failures` command family
- Background daemon: `tide schedule start|stop|status|logs`
- MCP protocol server: `tide mcp` exposes tools for AI agents

## Common pitfalls

- Adding a new flag to a command: the var must be declared at package level (cmd/*.go pattern), not inside RunE
- `output.PrintError` returns a `*CmdError` for the RunE return — always `return output.PrintError(...)`
- JSON time formatting uses `Format("2006-01-02 15:04:05")` — this specific format string is non-negotiable in SQLite
- When adding a migration, bump the version integer and append to the `migrations` slice in `migrate()`
- `UpdateFetchError` signature is `(id int64, errMsg string, statusCode int)` — it auto-classifies the error and records a `feed_failure` row + `change_log` event
