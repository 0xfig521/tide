package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/spf13/cobra"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/0xfig521/tide/internal/fetcher"
	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/repo"
)

// ensure gofeed import is retained for Parser.Fetch return type usage
var _ = (*gofeed.Feed)(nil)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server over stdio",
	Long: `Start a Model Context Protocol (MCP) server over stdio that exposes
Tide's RSS commands as structured tools for AI agents.

Registered tools:
  discover_feeds, add_feed, fetch_feeds, search_entries,
  list_entries, get_entry, mark_entry, get_feed_health,
  list_failed_feeds, clear_failed_feeds

All tool outputs are JSON. The server runs until the client disconnects
or the process receives SIGTERM/SIGINT.`,
	RunE: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) error {
	s := server.NewMCPServer("tide", version, server.WithToolCapabilities(true))

	registerDiscoverFeeds(s)
	registerAddFeed(s)
	registerFetchFeeds(s)
	registerSearchEntries(s)
	registerListEntries(s)
	registerGetEntry(s)
	registerMarkEntry(s)
	registerGetFeedHealth(s)
	registerListFailedFeeds(s)
	registerClearFailedFeeds(s)

	fmt.Fprintln(cmd.ErrOrStderr(), "Tide MCP server starting (stdio)...")
	return server.ServeStdio(s)
}

// ---------------------------------------------------------------------------
// discover_feeds — try parsing a URL as a feed and return metadata
// ---------------------------------------------------------------------------

func registerDiscoverFeeds(s *server.MCPServer) {
	tool := mcp.NewTool("discover_feeds",
		mcp.WithDescription("Discover RSS/Atom/JSON Feeds from a URL. Attempts to parse the given URL as a feed and returns feed metadata (title, description, link, feed type)."),
		mcp.WithString("url", mcp.Required(), mcp.Description("Website or feed URL to discover")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url, err := req.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError("url is required"), nil
		}

		parser := fetcher.DefaultConfig().NewParser()
		feed, _, _, statusCode, fetchErr := parser.Fetch(url, "", "")

		if fetchErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot fetch feed: %v (status %d)", fetchErr, statusCode)), nil
		}
		if feed == nil {
			return mcp.NewToolResultError(fmt.Sprintf("no feed data found at %s (status %d)", url, statusCode)), nil
		}

		data := map[string]any{
			"feed_url":    url,
			"title":       feed.Title,
			"description": feed.Description,
			"link":        feed.Link,
			"feed_type":   feed.FeedType,
			"language":    feed.Language,
			"item_count":  len(feed.Items),
		}

		b, _ := json.Marshal(data)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// add_feed — subscribe to a feed
// ---------------------------------------------------------------------------

func registerAddFeed(s *server.MCPServer) {
	tool := mcp.NewTool("add_feed",
		mcp.WithDescription("Subscribe to an RSS/Atom/JSON Feed URL. Optionally assign to a category (auto-created if it doesn't exist)."),
		mcp.WithString("feed_url", mcp.Required(), mcp.Description("Feed URL to subscribe to")),
		mcp.WithString("category", mcp.Description("Optional category name to assign the feed to")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		feedURL, err := req.RequireString("feed_url")
		if err != nil {
			return mcp.NewToolResultError("feed_url is required"), nil
		}

		existing, _ := feedRepo().GetByURL(feedURL)
		if existing != nil {
			b, _ := json.Marshal(map[string]any{
				"status":   "already_exists",
				"id":       existing.ID,
				"title":    existing.Title,
				"feed_url": existing.FeedURL,
			})
			return mcp.NewToolResultText(string(b)), nil
		}

		f, createErr := feedRepo().Create(feedURL)
		if createErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to add feed: %v", createErr)), nil
		}

		category := req.GetString("category", "")
		if category != "" {
			cat, catErr := categoryRepo().GetByName(category)
			if catErr != nil {
				cat, _ = categoryRepo().Create(category, "")
			}
			if cat != nil {
				_ = feedRepo().AssignCategory(f.ID, cat.ID)
			}
		}

		b, _ := json.Marshal(map[string]any{
			"status":   "imported",
			"id":       f.ID,
			"feed_url": f.FeedURL,
			"title":    f.Title,
			"category": category,
		})
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// fetch_feeds — run the fetcher
// ---------------------------------------------------------------------------

func registerFetchFeeds(s *server.MCPServer) {
	tool := mcp.NewTool("fetch_feeds",
		mcp.WithDescription("Fetch articles from RSS feeds. Fetches all due feeds by default, or a specific feed by ID."),
		mcp.WithNumber("feed_id", mcp.Description("Optional: fetch a specific feed by ID")),
		mcp.WithBoolean("quiet", mcp.Description("Suppress progress output (default true for MCP)")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		feedIDFloat, _ := req.RequireFloat("feed_id")
		feedID := int64(feedIDFloat)

		cfg := fetcher.DefaultConfig()
		parser := cfg.NewParser()

		fr := feedRepo()
		er := entryRepo()

		var jobs []fetcher.FetchJob
		if feedID > 0 {
			f, fErr := fr.GetByID(feedID)
			if fErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("feed not found: %v", fErr)), nil
			}
			lastFetchedAt := ""
			if f.LastFetchedAt != nil {
				lastFetchedAt = f.LastFetchedAt.Format("2006-01-02 15:04:05")
			}
			jobs = append(jobs, fetcher.FetchJob{
				FeedID: f.ID, FeedURL: f.FeedURL,
				ETag: f.ETagHeader, LastModified: f.LastModifiedHeader,
				LastFetchedAt: lastFetchedAt,
			})
		} else {
			feeds, _ := fr.GetDueFeeds(100)
			for _, f := range feeds {
				lastFetchedAt := ""
				if f.LastFetchedAt != nil {
					lastFetchedAt = f.LastFetchedAt.Format("2006-01-02 15:04:05")
				}
				jobs = append(jobs, fetcher.FetchJob{
					FeedID: f.ID, FeedURL: f.FeedURL,
					ETag: f.ETagHeader, LastModified: f.LastModifiedHeader,
					LastFetchedAt: lastFetchedAt,
				})
			}
		}

		if len(jobs) == 0 {
			b, _ := json.Marshal(map[string]any{
				"message":       "no feeds to fetch",
				"feeds_fetched": 0,
				"new_entries":   0,
				"unchanged":     0,
				"failed":        0,
			})
			return mcp.NewToolResultText(string(b)), nil
		}

		var newEntries, unchanged, failed int
		for _, job := range jobs {
			feed, etag, lastModified, statusCode, fetchErr := parser.Fetch(job.FeedURL, job.ETag, job.LastModified)
			if fetchErr != nil {
				_ = fr.UpdateFetchError(job.FeedID, fetchErr.Error(), statusCode)
				failed++
				continue
			}

			now := time.Now()
			nextCheck := now.Add(cfg.CheckInterval)

			if statusCode == 304 {
				_ = fr.UpdateFetchResult(job.FeedID, etag, lastModified, statusCode, now, nextCheck)
				unchanged++
				continue
			}

			if feed != nil {
				_ = fr.UpdateMeta(job.FeedID, feed.Title, feed.Description, feed.Link,
					fetcher.ImageURL(feed), feed.Language, feed.FeedType)
				_ = fr.UpdateFetchResult(job.FeedID, etag, lastModified, statusCode, now, nextCheck)

				for _, item := range feed.Items {
					entry := fetcher.ConvertEntry(job.FeedID, item)
					if err := er.InsertOrIgnore(entry); err == nil {
						newEntries++
					}
				}
			}
		}

		b, _ := json.Marshal(map[string]any{
			"feeds_fetched": len(jobs),
			"new_entries":   newEntries,
			"unchanged":     unchanged,
			"failed":        failed,
		})
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// search_entries — full-text search over entries
// ---------------------------------------------------------------------------

func registerSearchEntries(s *server.MCPServer) {
	tool := mcp.NewTool("search_entries",
		mcp.WithDescription("Full-text search across RSS entries using SQLite FTS5. Returns lightweight entry metadata (no full content by default)."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (FTS5 syntax supported)")),
		mcp.WithString("since", mcp.Description("Time range filter: 1h, 6h, 12h, 24h, 3d, 7d, 14d, 30d")),
		mcp.WithNumber("limit", mcp.Description("Maximum results to return (default 20)")),
		mcp.WithString("state", mcp.Description("Filter by processing state: new, seen, processed, ignored, failed")),
		mcp.WithString("sort", mcp.Description("Sort order: relevance (default with keyword) or published")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("query is required"), nil
		}

		since := req.GetString("since", "")
		limit, _ := req.RequireFloat("limit")
		state := req.GetString("state", "")
		sort := req.GetString("sort", "published")

		pageSize := 20
		if limit > 0 {
			pageSize = int(limit)
		}

		q := repo.EntryQuery{
			Keyword:  query,
			Since:    sinceExpr(since),
			SortBy:   sort,
			State:    state,
			Page:     1,
			PageSize: pageSize,
		}

		entries, listErr := entryRepo().ListEntries(q)
		if listErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", listErr)), nil
		}

		total, _ := entryRepo().CountEntries(q)

		outputs := make([]models.EntryOutput, 0, len(entries))
		for _, e := range entries {
			out := entryToFullOutput(e)
			out.Content = "" // lightweight — no full content
			outputs = append(outputs, out)
		}

		result := map[string]any{
			"items":     outputs,
			"total":     total,
			"page":      1,
			"page_size": pageSize,
		}

		b, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// list_entries — browse entries with filters (no keyword required)
// ---------------------------------------------------------------------------

func registerListEntries(s *server.MCPServer) {
	tool := mcp.NewTool("list_entries",
		mcp.WithDescription("List RSS entries with filtering by time, state, category, and feed. Returns lightweight entry metadata (no full content)."),
		mcp.WithString("since", mcp.Description("Time range filter: 1h, 6h, 12h, 24h, 3d, 7d, 14d, 30d")),
		mcp.WithString("state", mcp.Description("Filter by processing state: new, seen, processed, ignored, failed")),
		mcp.WithString("category", mcp.Description("Filter by category name")),
		mcp.WithNumber("feed_id", mcp.Description("Filter by feed ID")),
		mcp.WithNumber("limit", mcp.Description("Maximum results to return (default 20)")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		since := req.GetString("since", "")
		state := req.GetString("state", "")
		category := req.GetString("category", "")
		feedIDFloat, _ := req.RequireFloat("feed_id")
		limit, _ := req.RequireFloat("limit")

		pageSize := 20
		if limit > 0 {
			pageSize = int(limit)
		}

		q := repo.EntryQuery{
			CategoryName: category,
			FeedID:       int64(feedIDFloat),
			Since:        sinceExpr(since),
			State:        state,
			Page:         1,
			PageSize:     pageSize,
		}

		entries, listErr := entryRepo().ListEntries(q)
		if listErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list failed: %v", listErr)), nil
		}

		total, _ := entryRepo().CountEntries(q)

		outputs := make([]models.EntryOutput, 0, len(entries))
		for _, e := range entries {
			out := entryToFullOutput(e)
			out.Content = "" // lightweight — no full content
			outputs = append(outputs, out)
		}

		result := map[string]any{
			"items":     outputs,
			"total":     total,
			"page":      1,
			"page_size": pageSize,
		}

		b, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// get_entry — fetch a single entry with optional content control
// ---------------------------------------------------------------------------

func registerGetEntry(s *server.MCPServer) {
	tool := mcp.NewTool("get_entry",
		mcp.WithDescription("Get full details of a single RSS entry by ID. Supports content truncation for token budget control."),
		mcp.WithNumber("entry_id", mcp.Required(), mcp.Description("Entry ID to retrieve")),
		mcp.WithBoolean("full_content", mcp.Description("Include full article content (default true)")),
		mcp.WithNumber("token_budget", mcp.Description("Rough token budget for content (chars/4). Truncates content if exceeded.")),
		mcp.WithNumber("max_chars", mcp.Description("Maximum characters for content field. Truncates if exceeded.")),
		mcp.WithBoolean("as_text", mcp.Description("Return content as plain text instead of JSON")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		entryIDFloat, err := req.RequireFloat("entry_id")
		if err != nil {
			return mcp.NewToolResultError("entry_id is required"), nil
		}
		entryID := int64(entryIDFloat)

		entry, getErr := entryRepo().GetByID(entryID)
		if getErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("entry not found: %v", getErr)), nil
		}

		fullContent := req.GetBool("full_content", true)
		maxChars, _ := req.RequireFloat("max_chars")
		tokenBudget, _ := req.RequireFloat("token_budget")
		asText := req.GetBool("as_text", false)

		pubDate := ""
		if entry.PublishedAt != nil {
			pubDate = entry.PublishedAt.Format("2006-01-02 15:04:05")
		}

		// Apply token budget truncation (chars = tokens * 4)
		if tokenBudget > 0 && maxChars == 0 {
			maxChars = tokenBudget * 4
		}

		content := entry.Content
		description := entry.Description
		truncated := false
		contentLen := len(content)

		if maxChars > 0 {
			maxC := int(maxChars)
			if len(content) > maxC {
				content = content[:maxC]
				truncated = true
			}
			// Also truncate description if it's the only content source
			if len(description) > maxC {
				description = description[:maxC]
			}
		}

		if !fullContent {
			content = ""
		}

		result := map[string]any{
			"id":               entry.ID,
			"title":            entry.Title,
			"url":              entry.URL,
			"feed_id":          entry.FeedID,
			"feed_title":       entry.FeedTitle,
			"author":           entry.AuthorName,
			"published_at":     pubDate,
			"description":      description,
			"content":          content,
			"categories":       entry.Categories,
			"guid":             entry.GUID,
			"truncated":        truncated,
			"content_length":   contentLen,
			"estimated_tokens": contentLen / 4,
		}

		// Optionally return as plain text
		if asText {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Title: %s\n", entry.Title))
			sb.WriteString(fmt.Sprintf("URL: %s\n", entry.URL))
			sb.WriteString(fmt.Sprintf("Author: %s\n", entry.AuthorName))
			sb.WriteString(fmt.Sprintf("Published: %s\n", pubDate))
			sb.WriteString(fmt.Sprintf("Feed: %s\n", entry.FeedTitle))
			sb.WriteString(fmt.Sprintf("Description: %s\n", description))
			if content != "" {
				sb.WriteString(fmt.Sprintf("Content: %s\n", content))
			}
			if truncated {
				sb.WriteString("[Content was truncated]\n")
			}
			return mcp.NewToolResultText(sb.String()), nil
		}

		b, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// mark_entry — set processing state on an entry
// ---------------------------------------------------------------------------

func registerMarkEntry(s *server.MCPServer) {
	tool := mcp.NewTool("mark_entry",
		mcp.WithDescription("Set processing state on an entry for agent workflow tracking. Valid states: new, seen, processed, ignored, failed."),
		mcp.WithNumber("entry_id", mcp.Required(), mcp.Description("Entry ID to mark")),
		mcp.WithString("state", mcp.Required(), mcp.Description("Processing state: new, seen, processed, ignored, failed")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags (e.g., 'summarized,rust')")),
		mcp.WithString("note", mcp.Description("Optional note about the entry")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		entryIDFloat, err := req.RequireFloat("entry_id")
		if err != nil {
			return mcp.NewToolResultError("entry_id is required"), nil
		}
		entryID := int64(entryIDFloat)

		state, err := req.RequireString("state")
		if err != nil {
			return mcp.NewToolResultError("state is required"), nil
		}

		if !validStates[state] {
			return mcp.NewToolResultError(
				fmt.Sprintf("invalid state: %s. Must be one of: new, seen, processed, ignored, failed", state),
			), nil
		}

		entry, getErr := entryRepo().GetByID(entryID)
		if getErr != nil || entry == nil {
			return mcp.NewToolResultError("entry not found"), nil
		}

		tags := req.GetString("tags", "")
		note := req.GetString("note", "")

		var markErr error
		switch {
		case tags != "" && note != "":
			markErr = stateRepo().SetStateFull(entryID, state, tags, note)
		case tags != "":
			markErr = stateRepo().SetStateWithTags(entryID, state, tags)
		case note != "":
			markErr = stateRepo().SetStateFull(entryID, state, "", note)
		default:
			markErr = stateRepo().SetState(entryID, state)
		}

		if markErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to mark entry: %v", markErr)), nil
		}

		result := map[string]any{
			"entry_id": entryID,
			"state":    state,
		}
		if tags != "" {
			result["tags"] = tags
		}
		if note != "" {
			result["note"] = note
		}

		b, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// get_feed_health — health stats for feeds
// ---------------------------------------------------------------------------

func registerGetFeedHealth(s *server.MCPServer) {
	tool := mcp.NewTool("get_feed_health",
		mcp.WithDescription("Get health statistics for RSS feeds. Returns status, success rate, entry counts, and staleness info."),
		mcp.WithNumber("feed_id", mcp.Description("Optional: get health for a specific feed. Omit for all feeds.")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		feedIDFloat, _ := req.RequireFloat("feed_id")
		feedID := int64(feedIDFloat)

		query := `
			SELECT
				f.id,
				f.title,
				f.feed_url,
				COALESCE(f.last_fetched_at, '') as last_fetched_at,
				f.parsing_error_count,
				COALESCE(f.parsing_error_msg, '') as last_error,
				(SELECT COUNT(*) FROM entries WHERE feed_id = f.id
				 AND published_at >= datetime('now', '-7 days')) as entries_7d,
				(SELECT COUNT(*) FROM entries WHERE feed_id = f.id
				 AND published_at >= datetime('now', '-30 days')) as entries_30d,
				(SELECT COUNT(*) FROM entries WHERE feed_id = f.id) as total_entries,
				f.is_active
			FROM feeds f
		`
		var args []any
		if feedID > 0 {
			query += " WHERE f.id = ?"
			args = append(args, feedID)
		}
		query += " ORDER BY f.title"

		rows, qErr := dbConn.Conn.Query(query, args...)
		if qErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("health query failed: %v", qErr)), nil
		}
		defer rows.Close()

		type healthEntry struct {
			FeedID            int64  `json:"feed_id"`
			Title             string `json:"title"`
			FeedURL           string `json:"feed_url"`
			LastFetchedAt     string `json:"last_fetched_at"`
			ConsecutiveErrors int    `json:"consecutive_errors"`
			LastError         string `json:"last_error"`
			Entries7d         int    `json:"entries_7d"`
			Entries30d        int    `json:"entries_30d"`
			TotalEntries      int    `json:"total_entries"`
			IsActive          bool   `json:"is_active"`
			Status            string `json:"status"`
			StaleDays         int    `json:"stale_days"`
		}

		var results []healthEntry
		for rows.Next() {
			var h healthEntry
			if err := rows.Scan(&h.FeedID, &h.Title, &h.FeedURL,
				&h.LastFetchedAt, &h.ConsecutiveErrors, &h.LastError,
				&h.Entries7d, &h.Entries30d, &h.TotalEntries, &h.IsActive,
			); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("scan error: %v", err)), nil
			}

			// Compute status
			if !h.IsActive {
				h.Status = "dead"
			} else if h.ConsecutiveErrors >= 5 {
				h.Status = "failing"
			} else if h.LastFetchedAt != "" {
				lastFetched, parseErr := time.Parse("2006-01-02 15:04:05", h.LastFetchedAt)
				if parseErr == nil {
					h.StaleDays = int(time.Since(lastFetched).Hours() / 24)
					if h.StaleDays > 30 {
						h.Status = "stale"
					} else if h.ConsecutiveErrors > 0 {
						h.Status = "failing"
					} else {
						h.Status = "healthy"
					}
				} else {
					h.Status = "unknown"
				}
			} else {
				h.Status = "unknown"
			}

			results = append(results, h)
		}

		if err := rows.Err(); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("rows iteration error: %v", err)), nil
		}

		b, _ := json.Marshal(results)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// list_failed_feeds — show feeds that are persistently failing
// ---------------------------------------------------------------------------

func registerListFailedFeeds(s *server.MCPServer) {
	tool := mcp.NewTool("list_failed_feeds",
		mcp.WithDescription("List RSS feeds that have crossed the failure threshold. Returns feed metadata plus the most recent classified failure (type, HTTP status, error message, timestamp). Good for detecting dead/rotting feeds without re-parsing raw error strings."),
		mcp.WithNumber("threshold", mcp.Description("Failure threshold: minimum parsing_error_count (default 3)")),
		mcp.WithString("type", mcp.Description(fmt.Sprintf("Filter by last-failure type: %s", repo.FailureTypeList()))),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		threshold := 3
		if v, vErr := req.RequireFloat("threshold"); vErr == nil && v > 0 {
			threshold = int(v)
		}

		var typeFilter models.FailureType
		if raw := req.GetString("type", ""); raw != "" {
			typeFilter = models.FailureType(raw)
			if !models.ValidFailureTypes[typeFilter] {
				return mcp.NewToolResultError(
					fmt.Sprintf("invalid type %q. Valid: %s", raw, repo.FailureTypeList())), nil
			}
		}

		feeds, err := failureRepo().ListFailingFeeds(threshold, typeFilter)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query failed: %v", err)), nil
		}

		b, _ := json.Marshal(feeds)
		return mcp.NewToolResultText(string(b)), nil
	})
}

// ---------------------------------------------------------------------------
// clear_failed_feeds — hard-delete feeds that are failing
// ---------------------------------------------------------------------------

func registerClearFailedFeeds(s *server.MCPServer) {
	tool := mcp.NewTool("clear_failed_feeds",
		mcp.WithDescription("Hard-delete RSS feeds that are persistently failing. Requires feed_id or confirm=true for bulk mode. ⚠️ This is destructive: removes feeds, their entries, categories, and failure history permanently."),
		mcp.WithNumber("feed_id", mcp.Description("Optional: clear a specific feed by ID. If omitted, all failing feeds are cleared.")),
		mcp.WithNumber("threshold", mcp.Description("Failure threshold for bulk clear (default 3, used only when feed_id is omitted)")),
		mcp.WithBoolean("confirm", mcp.Description("Bulk clear requires confirm=true. This is a safety gate — the operation is permanent.")),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Single feed mode
		if v, vErr := req.RequireFloat("feed_id"); vErr == nil && v > 0 {
			feedID := int64(v)
			f, err := feedRepo().GetByID(feedID)
			if err != nil {
				return mcp.NewToolResultError("feed not found"), nil
			}
			if delErr := feedRepo().Delete(feedID); delErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("delete failed: %v", delErr)), nil
			}
			b, _ := json.Marshal(map[string]any{
				"action":   "cleared",
				"id":       f.ID,
				"title":    f.Title,
				"feed_url": f.FeedURL,
			})
			return mcp.NewToolResultText(string(b)), nil
		}

		// Bulk mode — require confirm=true
		if !req.GetBool("confirm", false) {
			return mcp.NewToolResultError(
				"bulk clear is destructive and permanent. Set confirm=true to proceed."), nil
		}

		threshold := 3
		if v, vErr := req.RequireFloat("threshold"); vErr == nil && v > 0 {
			threshold = int(v)
		}

		feeds, err := failureRepo().ListFailingFeeds(threshold, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query failed: %v", err)), nil
		}

		type clearedEntry struct {
			ID      int64  `json:"id"`
			Title   string `json:"title"`
			FeedURL string `json:"feed_url"`
		}
		cleared := make([]clearedEntry, 0, len(feeds))
		for _, f := range feeds {
			if delErr := feedRepo().Delete(f.FeedID); delErr != nil {
				continue
			}
			cleared = append(cleared, clearedEntry{ID: f.FeedID, Title: f.Title, FeedURL: f.FeedURL})
		}

		result := map[string]any{
			"action":    "cleared",
			"feeds":     len(cleared),
			"threshold": threshold,
			"cleared":   cleared,
		}
		b, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(b)), nil
	})
}
