# Tide 性能优化文档

## 背景

Tide 的目标是成为面向 AI agent 的 RSS 数据工具。性能优化的重点不是终端 UI 流畅度，而是：

- 大量 feed 并发抓取。
- 大量 entry 的高效写入。
- 快速搜索和过滤。
- 低 token 成本输出。
- 稳定支持 agent 的增量处理流程。

当前项目的主要性能瓶颈可预期集中在：

1. 网络抓取。
2. SQLite 写入。
3. 搜索与分页。
4. 输出体积和序列化。
5. rules / routing 后续引入后的正则匹配成本。

## 优先级总览

| 优先级 | 优化方向 | 预期收益 |
|---|---|---|
| P0 | list/search 不默认读取 `content` | 降低查询 I/O、内存和输出成本 |
| P0 | fetch 插入使用事务和 prepared statement | 显著提升批量写入速度 |
| P0 | 网络 fetch 与 SQLite writer 解耦 | 避免并发 goroutine 抢单 writer |
| P1 | JSONL streaming 输出 | 降低内存峰值，适合大结果 |
| P1 | 复合索引和 cursor pagination | 改善常用过滤与深分页 |
| P1 | HTTP body size limit 和超时细分 | 防止坏源拖垮整体抓取 |
| P2 | entry content 拆表 | 大规模数据下显著改善 list/search |
| P2 | rules regex 缓存与增量应用 | 降低规则系统成本 |
| P2 | benchmark / stats | 让优化可量化 |

## 1. 抓取路径优化

### 1.1 fetch 与 DB 写入解耦

当前并发 fetch worker 可能同时写 SQLite。SQLite 实际是单 writer，并发写入会竞争同一个写锁。

建议架构：

```text
N 个 fetch workers
    ↓
parsed feed result channel
    ↓
1 个 DB writer goroutine
    ↓
SQLite transaction batch write
```

好处：

- 网络 I/O 并发不受 SQLite 写锁影响。
- DB 写入可批量事务化。
- 更容易统计每个 feed 的抓取耗时、写入耗时和失败原因。
- 可集中处理 change_log、FTS、rules。

### 1.2 批量写入使用事务

当前逐条 `InsertOrIgnore` 会产生大量小写入。建议为每个 feed 或每批 feed 使用事务。

伪代码：

```go
tx, err := db.Begin()
if err != nil {
    return err
}
stmt, err := tx.Prepare(insertEntrySQL)
if err != nil {
    tx.Rollback()
    return err
}
defer stmt.Close()

for _, entry := range entries {
    if _, err := stmt.Exec(...); err != nil {
        tx.Rollback()
        return err
    }
}

return tx.Commit()
```

建议粒度：

- 单 feed 条目较少：每个 feed 一个 transaction。
- 批量导入历史数据：每 500～1000 entries 一个 transaction。

### 1.3 高频 SQL 使用 prepared statement

建议优先 prepare：

- insert entry
- update feed metadata
- update fetch result
- update fetch error
- insert change_log
- insert / update entry state

避免在循环内重复解析 SQL。

### 1.4 HTTP body size limit

RSS 源可能异常大或恶意返回巨大响应。建议加最大 body 限制。

示例：

```go
limited := io.LimitReader(resp.Body, maxFeedBytes)
feed, err := fp.Parse(limited)
```

建议默认：

```text
max_feed_bytes = 10MB
```

并支持配置：

```bash
tide fetch --max-feed-bytes 10485760
```

### 1.5 超时拆分

当前如果只有整体 timeout，坏源可能占用 worker 较久。建议细分：

- dial timeout
- TLS handshake timeout
- response header timeout
- whole request timeout

推荐默认：

```text
dial_timeout = 5s
tls_timeout = 5s
response_header_timeout = 10s
request_timeout = 30s
```

### 1.6 并发上限分层

建议区分：

- global concurrency
- per-host concurrency
- DB writer concurrency

示例：

```bash
tide fetch --concurrency 20 --per-host 2
```

避免对同一域名过度并发请求。

## 2. SQLite 写入优化

### 2.1 SQLite PRAGMA

当前已有 WAL、busy_timeout、foreign_keys。可考虑补充：

```sql
PRAGMA synchronous=NORMAL;
PRAGMA temp_store=MEMORY;
PRAGMA mmap_size=268435456;
PRAGMA cache_size=-20000;
```

说明：

- `synchronous=NORMAL`：本地缓存型 CLI 数据库通常可以接受，写入更快。
- `temp_store=MEMORY`：排序、临时表更快。
- `mmap_size`：改善大库读取性能。
- `cache_size=-20000`：使用约 20MB page cache。

需要注意：PRAGMA 应保守启用，并通过 benchmark 验证。

### 2.2 控制 FTS / change_log 写放大

当前 schema 已有 FTS trigger，后续如果加入 change_log trigger，每插入一条 entry 会引发额外写入。

建议：

- fetch 普通增量：保留 trigger。
- 大规模 import：支持批量导入后 rebuild FTS。
- change_log 只记录必要事件。
- 避免对无变化 update 也写 change_log。

可考虑：

```bash
tide import entries.jsonl --defer-fts
tide fts rebuild
```

### 2.3 content 拆表

当前如果 `entries` 表同时保存 metadata 和大字段 `content`，list/search 会更容易读到不必要的大字段。

建议长期拆分：

```text
entries
  id
  feed_id
  title
  url
  guid
  description
  author
  published_at
  hash
  created_at
  updated_at

entry_contents
  entry_id
  content
  raw_content
  text_content
  content_hash
```

查询原则：

- `list/search` 只读 `entries` metadata。
- `get --full` 才 join `entry_contents`。
- FTS 可索引 `entry_contents.text_content`，但结果列表不返回 full content。

## 3. 查询与搜索优化

### 3.1 list/search 不默认读取 content

这是当前最值得优先做的查询优化。

建议拆分 columns：

```go
const entryListCols = `
  e.id, e.feed_id, e.title, e.url, e.guid, e.description,
  e.author_name, e.image_url, e.categories,
  COALESCE(e.published_at,''), e.created_at, e.updated_at,
  f.title as feed_title
`

const entryFullCols = `
  e.id, e.feed_id, e.title, e.url, e.guid, e.content, e.description,
  e.author_name, e.image_url, e.categories,
  COALESCE(e.published_at,''), e.created_at, e.updated_at,
  f.title as feed_title
`
```

使用方式：

- `list/search` → `entryListCols`
- `get --full` → `entryFullCols`

### 3.2 Cursor pagination 替代深 OFFSET

当前 `LIMIT ? OFFSET ?` 在深分页时会越来越慢。Agent 场景更适合 cursor。

建议新增：

```bash
tide list --after <cursor> --limit 50
tide search "MCP" --after <cursor> --limit 50
```

SQL 示例：

```sql
WHERE
  published_at < ?
  OR (published_at = ? AND id < ?)
ORDER BY published_at DESC, id DESC
LIMIT ?
```

cursor 可编码：

```text
published_at:id
```

例如：

```text
2026-06-01T10:00:00Z:123
```

### 3.3 复合索引

建议根据常用查询添加：

```sql
CREATE INDEX IF NOT EXISTS idx_entries_feed_published
ON entries(feed_id, published_at DESC);

CREATE INDEX IF NOT EXISTS idx_entry_states_state_entry
ON entry_states(state, entry_id);

CREATE INDEX IF NOT EXISTS idx_feed_categories_category_feed
ON feed_categories(category_id, feed_id);
```

如果 category + time 查询非常高频，可用 `EXPLAIN QUERY PLAN` 验证是否需要进一步索引。

### 3.4 FTS 排序策略

建议搜索支持：

```bash
tide search "agent rss" --sort relevance
tide search "agent rss" --sort published
```

策略：

- `relevance`：按 FTS rank。
- `published`：按发布时间。
- 可选混合：rank + recency。

### 3.5 Count 查询可选化

查询结果如果默认 JSONL，很多 agent 不需要 total count。`COUNT(*)` 在复杂过滤下可能额外消耗。

建议：

```bash
tide list --count
tide search "MCP" --count
```

默认不计算 total，除非 `--format json` 或显式 `--count`。

## 4. 输出性能与 token 成本

### 4.1 JSONL streaming 输出

大结果输出时不要先构造完整 slice。

推荐：

```go
enc := json.NewEncoder(os.Stdout)
for rows.Next() {
    item := scanOne(rows)
    if err := enc.Encode(item); err != nil {
        return err
    }
}
```

收益：

- 降低内存峰值。
- 结果可流式消费。
- 更适合 agent 和管道处理。

### 4.2 默认轻量字段

`list/search` 默认字段建议：

- id
- title
- url
- published_at
- feed_id
- feed_title
- author
- description 摘要
- categories
- state
- tags

不默认输出：

- full content
- raw HTML
- 大图片字段

### 4.3 description/content 截断

建议支持：

```bash
tide list --max-description-chars 300
tide get 42 --max-chars 4000
tide get 42 --token-budget 2000
```

输出应包含：

```json
{
  "truncated": true,
  "char_count": 4000,
  "estimated_tokens": 1100
}
```

## 5. Rules / Routing 性能

后续如果实现 rules，不应每次全量扫描所有 entries。

### 5.1 只对新增 entry 应用规则

规则应用时机：

- fetch 插入新 entry 后。
- import 新 entry 后。
- 手动 `tide rule apply --since ...`。

避免每次 `fetch --apply-rules` 全量扫库。

### 5.2 Regex 编译缓存

规则启动时加载并编译：

```go
compiledRules := []*regexp.Regexp{}
```

不要每条 entry 重复 `regexp.Compile`。

### 5.3 按字段分组规则

规则按字段分组：

- title rules
- description rules
- author rules
- category rules
- content rules

默认只对 title/description/author/category 运行。content rules 应显式开启，因为 content 体积大。

### 5.4 快速预过滤

如果规则是简单关键词，可先用 strings.Contains 做快速判断，再执行 regex。

## 6. 观测与 Benchmark

性能优化应量化。建议增加 benchmark 和 stats。

### 6.1 Benchmark 场景

建议覆盖：

- 100 feeds mock fetch。
- 1000 entries batch insert。
- 10k entries list。
- 100k entries search。
- FTS rebuild。
- rules apply 10k entries。
- JSONL streaming 10k rows。

### 6.2 指标

建议采集：

- fetch total duration
- network duration
- parse duration
- DB write duration
- entries inserted/sec
- search p50 / p95
- list p50 / p95
- DB size
- FTS size
- average content size
- output bytes

### 6.3 stats 命令

建议新增：

```bash
tide stats
```

输出：

```json
{
  "ok": true,
  "data": {
    "feeds": 120,
    "entries": 54000,
    "entry_contents_bytes": 134217728,
    "db_size_bytes": 180000000,
    "fts_size_bytes": 42000000,
    "last_fetch_duration_ms": 2400
  },
  "error": null,
  "meta": {}
}
```

## 7. 推荐实施顺序

### ✅ Phase 1：低风险高收益（已完成）

> **实施日期**: 2026-06-01
> **状态**: ✅ 全部完成，测试通过

| # | 优化项 | 状态 | 说明 |
|---|---|---|---|
| 1 | `list/search` 不读取 `content` | ✅ | 拆分 `entryListCols`(13字段无content) 和 `entryFullCols`(14字段含content) |
| 2 | `fetch` entry 插入使用事务 | ✅ | 新增 `BatchInsertEntries()`，单 transaction + prepared statement 批量写入 |
| 3 | prepared statement 复用 | ✅ | `EntryRepo.InsertOrIgnore`、`FeedRepo.UpdateMeta/UpdateFetchResult/UpdateFetchError` 均使用 lazy-init prepared statement (`sync.Once`) |
| 4 | progress/log 保持 stderr | ✅ | 前置条件已满足（`progressbar.OptionSetWriter(os.Stderr)`） |

**实施详情**:

- **列拆分**: `entry_repo.go` 中 `entryCols` → `entryListCols` + `entryFullCols`，新增 `scanEntriesLight`（13字段）/ `scanEntriesFull`（14字段，后因直接内联移除）
- **批量写入**: `cmd/fetch.go` 和 `internal/fetcher/worker.go` 均已从逐条 `InsertOrIgnore` 切换为 `BatchInsertEntries`；`RowsAffected()` 替代原代码 `err == nil` 的错误计数逻辑
- **Prepared Statement**: `sync.Once` 模式实现懒加载，避免构造函数签名变更，线程安全
- **测试覆盖**: 新增 `TestEntryRepo_BatchInsertEntries`、`TestEntryRepo_ListEntries_ContentNotReturned`、`TestEntryRepo_ListByFeed_ContentNotReturned`，全部 11+8+3 个测试通过

### Phase 2：架构优化

1. fetch worker 与 DB writer 解耦。
2. JSONL streaming 输出。
3. count 查询可选化。
4. body size limit 和细分 timeout。

### Phase 3：大规模数据优化

1. 复合索引。
2. cursor pagination。
3. content 拆表。
4. FTS rebuild / deferred FTS。

### Phase 4：规则与观测

1. rules 只对新增 entry 应用。
2. regex cache。
3. benchmark。
4. `tide stats`。

## 8. 最推荐优先做的五项

如果只做五项，建议按顺序：

1. **list/search 不读取 content，只在 get --full 读取。**
2. **fetch 插入 entries 使用 transaction + prepared statement。**
3. **fetch 网络并发与 SQLite 单 writer 解耦。**
4. **查询输出支持 streaming JSONL。**
5. **增加复合索引和 cursor pagination。**

这五项能显著提升 Tide 作为 agent RSS 数据层的吞吐、延迟和上下文效率。

