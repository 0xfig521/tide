# Tide：面向 AI Agent 的 RSS 数据工具优化设计

## 一句话定位

Tide 不应定位为 RSS 阅读器，而应定位为：

> **Agent-native RSS ingestion and retrieval CLI**：为 AI agent 提供稳定的 RSS 订阅、发现、抓取、缓存、搜索、读取和处理状态管理能力。

它的核心价值不是让人类在终端里“读文章”，而是让 agent 不再每次临时编写 RSS 抓取、解析、去重、缓存和筛选代码。

## 产品原则

### 1. 面向机器优先

- 输出必须稳定、可解析、可脚本化。
- 默认行为应适合 agent 调用，而不是适合人类终端浏览。
- 进度、日志、诊断输出不能污染结构化 stdout。

### 2. 数据工具，而不是阅读器

Tide 应解决：

- 如何发现 feed。
- 如何批量订阅 feed。
- 如何增量抓取 feed。
- 如何统一 RSS / Atom / JSON Feed 差异。
- 如何搜索、过滤、分页。
- 如何获取单篇 entry 的详情。
- 如何记录 agent 是否已处理某条内容。

Tide 不应优先解决：

- 复杂终端阅读界面。
- 收藏夹式阅读体验。
- 类 Reeder / Feedly 的 UI。
- 重型文章排版。

### 3. Agent 工作流闭环

目标工作流应是：

```bash
tide discover https://example.com
tide batch-add feeds.json
tide fetch --quiet
tide search "rust async runtime" --since 7d --limit 10 --format jsonl
tide get 42 --full
tide mark 42 --state processed
```

这个闭环覆盖：

1. feed 发现
2. 订阅
3. 抓取
4. 检索
5. 获取内容
6. 标记处理状态

## 输出协议设计

当前最需要固定的是输出协议。建议将命令分为三类。

### 1. 控制类命令：JSON Envelope

适用命令：

- `add`
- `batch-add`
- `fetch`
- `import`
- `export`
- `remove`
- `mark`
- `schedule`

成功输出：

```json
{
  "ok": true,
  "data": {},
  "error": null,
  "meta": {}
}
```

失败输出：

```json
{
  "ok": false,
  "data": null,
  "error": {
    "code": "feed_not_found",
    "message": "feed 123 not found"
  },
  "meta": {}
}
```

规则：

- 成功：exit code `0`
- 失败：exit code 非 `0`
- stdout：只输出 JSON
- stderr：进度、日志、warning、diagnostics

### 2. 查询类命令：默认 JSONL

适用命令：

- `list`
- `search`
- `sources`

推荐默认格式：JSONL。

示例：

```jsonl
{"id":1,"title":"...","url":"...","published_at":"2026-06-01T10:00:00Z","feed_title":"Go Blog"}
{"id":2,"title":"...","url":"...","published_at":"2026-06-01T11:00:00Z","feed_title":"LWN"}
```

原因：

- 比 CSV 更稳，不怕逗号、换行、HTML、引号。
- 比 JSON array 更适合大结果和流式处理。
- agent 可逐行解析，token 成本也较低。
- 仍可提供 `--format json` 输出 envelope 或 array。

建议格式参数：

```bash
tide list --format jsonl
tide list --format json
tide list --format csv
tide search "kubernetes" --format jsonl
```

推荐默认：

```bash
--format jsonl
```

### 3. 单对象命令：JSON Envelope

适用命令：

- `get`
- `info`

示例：

```json
{
  "ok": true,
  "data": {
    "id": 42,
    "title": "...",
    "url": "...",
    "description": "...",
    "content": "..."
  },
  "error": null,
  "meta": {}
}
```

## 命令能力规划

### P0：稳定现有核心命令

#### `fetch`

建议行为：

```bash
tide fetch --quiet
tide fetch --feed 12
tide fetch --category ai
tide fetch --fail-fast
```

推荐输出字段：

```json
{
  "ok": true,
  "data": {
    "feeds_total": 10,
    "feeds_fetched": 8,
    "unchanged": 2,
    "failed": 0,
    "new_entries": 31,
    "updated_entries": 0,
    "duration_ms": 1240,
    "errors": []
  },
  "error": null,
  "meta": {}
}
```

语义建议：

- 部分 feed 失败时，默认仍 `ok=true`，但 `data.failed > 0`。
- 如果传 `--fail-fast`，遇到第一个失败直接 `ok=false`。
- progress bar 必须写 stderr。
- `--quiet` 禁止所有非必要 stderr 输出。

#### `list`

建议行为：

```bash
tide list --since 24h
tide list --category ai --limit 50
tide list --state unprocessed
tide list --format jsonl
```

默认只输出轻量字段：

- id
- title
- url
- published_at
- feed_id
- feed_title
- author
- description 摘要或截断值
- categories

不默认输出完整 content。

#### `search`

建议行为：

```bash
tide search "rust async" --since 7d --limit 20
tide search "kubernetes scheduler" --sort relevance
tide search "openai" --category ai --format jsonl
```

建议支持：

- `--since`
- `--category`
- `--feed`
- `--limit`
- `--sort relevance|published`
- `--format jsonl|json|csv`

### P1：补齐 agent 工作流命令

#### `discover`

新增命令：

```bash
tide discover https://example.com
```

作用：输入网站 URL，发现 RSS / Atom / JSON Feed。

输出示例：

```json
{
  "ok": true,
  "data": {
    "site_url": "https://example.com",
    "feeds": [
      {
        "url": "https://example.com/feed.xml",
        "type": "rss",
        "title": "Example Feed"
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

这是 agent 场景高频能力，因为用户经常只提供网站，而不是 feed URL。

#### `get`

已有 `get` 是正确方向。建议进一步明确内容层级：

```bash
tide get 42
tide get 42 --full
tide get 42 --content-only
```

建议：

- 默认：metadata + description
- `--full`：包含 RSS content
- `--content-only`：只输出正文相关字段，方便 agent 总结

注意：RSS content 与网页正文不同。如果未来要抓网页正文，建议单独设计 `extract`。

#### `mark`

如果 Tide 不是阅读器，就不应过度强调 `read/star`。Agent 更需要“处理状态”。

建议新增：

```bash
tide mark 42 --state processed
tide mark 42 --state ignored
tide mark 42 --tag summarized
tide list --state unprocessed
```

建议状态：

- `new`
- `seen`
- `processed`
- `ignored`
- `failed`

这比 `read/star` 更符合 agent 工作流。

### P2：批量与迁移能力

#### `batch-add`

当前方向正确。建议补充：

```bash
tide batch-add feeds.json
cat feeds.json | tide batch-add
tide batch-add feeds.json --strict
```

语义：

- JSON 解析失败：`ok=false`，exit code 非 0。
- 部分 feed 添加失败：默认 `ok=true`，`data.errored > 0`。
- `--strict`：只要有一条失败就 `ok=false`，exit code 非 0。

#### `import` / `export`

OPML 适合作为迁移和备份能力，应保留。

建议：

```bash
tide import subscriptions.opml
tide export --output subscriptions.opml
tide export --format json
```

## 数据模型建议

### Feed

核心字段：

- id
- title
- feed_url
- site_url
- categories
- last_fetched_at
- next_check_at
- error_count
- last_error

### Entry

核心字段：

- id
- feed_id
- feed_title
- title
- url
- guid
- author
- published_at
- description
- content
- categories
- hash
- created_at
- updated_at

### Agent State

建议新增独立处理状态字段或表：

- entry_id
- state
- tags
- note
- processed_at

示例：

```json
{
  "entry_id": 42,
  "state": "processed",
  "tags": ["summarized", "rust"],
  "note": "Used in weekly digest",
  "processed_at": "2026-06-01T10:00:00Z"
}
```

## 搜索能力建议

面向 agent，搜索是核心能力。建议从 `LIKE` 升级为 SQLite FTS5。

### 推荐索引字段

- title
- description
- content
- author
- feed_title

### 推荐排序

```bash
tide search "agent rss" --sort relevance
tide search "agent rss" --sort published
```

### 推荐过滤

```bash
tide search "agent rss" --since 7d
tide search "agent rss" --category ai
tide search "agent rss" --state unprocessed
```

## Agent Skill 设计建议

`tide/SKILL.md` 应告诉 agent：

1. 查询优先用 `jsonl`。
2. 控制命令读取 `.ok` 和 `.error.code`。
3. 搜索先拿轻量结果，再对少量 id 调用 `get --full`。
4. 不要默认拉取大 content。
5. 需要网站订阅时先 `discover`，再 `add` 或 `batch-add`。
6. 总结或处理后使用 `mark --state processed`。

推荐 skill 工作流：

```bash
# 1. Ensure feeds are fresh
tide fetch --quiet

# 2. Search lightweight entries
tide search "AI agent" --since 7d --limit 10 --format jsonl

# 3. Fetch full content only for selected entries
tide get 42 --full

# 4. Mark processed after use
tide mark 42 --state processed --tag summarized
```

## 与阅读器的边界

不建议优先投入：

- TUI 阅读界面
- 已读/未读作为主模型
- 收藏夹体验
- 复杂主题样式
- 推荐算法

可以保留但不应作为核心：

- `--format table`
- 简单 human-readable 输出
- schedule logs

核心应始终是：

- ingestion
- normalization
- retrieval
- state tracking
- machine-readable output

## 建议实施顺序

### 第一阶段：统一协议

1. 明确 `list/search/sources` 默认格式为 JSONL。
2. 控制命令统一 JSON envelope。
3. root 设置 `SilenceErrors: true`，避免重复 stderr 噪音。
4. 进度条只写 stderr。
5. 文档和实际行为保持一致。

### 第二阶段：补全工作流

1. 完善 `get` 的 `--full` / `--content-only`。
2. 新增 `discover`。
3. 新增 `mark` 和处理状态。
4. `batch-add` 增加 `--strict`。

### 第三阶段：强化检索

1. SQLite FTS5。
2. `--sort relevance|published`。
3. `--state` 过滤。
4. 搜索结果默认 JSONL。

### 第四阶段：测试与稳定性

至少覆盖：

- stdout 是否可解析。
- stderr 是否不影响 stdout。
- error code 和 exit code。
- `fetch` 部分失败语义。
- `list/search/get` 输出 schema。
- `batch-add` mixed input。
- OPML import/export。
- FTS 搜索结果。

## 最终成功标准

当 Tide 成熟后，agent 应能可靠完成：

> “帮我订阅这几个技术博客，抓取最近一周文章，找出和 AI agent / RSS / MCP 相关的内容，摘要最重要的 5 篇，并把处理过的条目标记为 processed。”

整个过程不需要：

- 临时写 RSS parser。
- 临时抓 XML。
- 临时处理 Atom/RSS 差异。
- 手写去重逻辑。
- 解析彩色终端输出。
- 维护一次性缓存。

这才是 Tide 作为 AI agent RSS 工具的核心价值。

