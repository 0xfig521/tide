# Tide：AI Agent RSS 工具进阶优化路线图

## 背景

在完成基础优化后，Tide 应具备：

- 稳定的机器可解析输出协议。
- feed 发现、订阅、批量订阅、OPML 导入导出。
- 并发抓取与缓存。
- 搜索、列表、单篇详情读取。
- agent 处理状态标记。
- JSONL / JSON envelope 等明确输出边界。

下一阶段目标不是继续增强“阅读器体验”，而是把 Tide 推进为 **AI agent 的 RSS 数据基础设施**。

## 阶段目标

> 让 Tide 不只是一个可被 agent 调用的 CLI，而是一个能长期服务 agent 工作流的 RSS ingestion、retrieval、delta tracking、content preparation 和 integration layer。

## 优先级总览

| 优先级 | 方向 | 价值 |
|---|---|---|
| P0 | MCP Server 模式 | 让 agent 直接以 tool 调用 Tide，不依赖 shell 拼命令 |
| P0 | Delta / Cursor 模式 | 稳定获取新增内容，支持增量 agent workflow |
| P0 | Token-budgeted Content | 控制上下文成本，直接服务 LLM 输入 |
| P1 | Feed Health | 让 agent 判断信息源可靠性 |
| P1 | RAG / JSONL / Markdown Export | 把 RSS 内容导入知识库、向量库、周报流程 |
| P1 | Rules / Routing | 自动分类、忽略、打标签、优先级分流 |
| P2 | Hooks / Webhooks | 把新内容接入自动化流水线 |
| P2 | Workspace / Profile | 支持多项目、多主题隔离 |
| P2 | Query DSL | 降低复杂查询的命令生成成本 |

## 1. MCP Server 模式

### 目标

提供：

```bash
tide mcp
```

让支持 MCP 的 agent 直接调用 Tide 的结构化工具，而不是通过 shell 命令和 stdout 解析。

### 建议 MCP Tools

- `discover_feeds`
- `add_feed`
- `batch_add_feeds`
- `import_opml`
- `export_opml`
- `fetch_feeds`
- `list_entries`
- `search_entries`
- `get_entry`
- `mark_entry`
- `get_feed_health`
- `export_entries`

### Tool 设计原则

- 输入参数应使用强 schema。
- 输出直接复用内部 response schema。
- 所有 tool 返回都带 provenance。
- 不在 MCP 层做 LLM 总结，只做数据获取和准备。

### 示例

```json
{
  "tool": "search_entries",
  "arguments": {
    "query": "AI agent RSS",
    "since": "7d",
    "limit": 10,
    "state": "unprocessed"
  }
}
```

返回：

```json
{
  "items": [
    {
      "id": 42,
      "title": "...",
      "url": "...",
      "published_at": "2026-06-01T10:00:00Z",
      "feed_title": "..."
    }
  ],
  "cursor": "..."
}
```

### 验收标准

- Claude / Codex / Cursor 等 MCP client 可直接连接。
- 常见 workflow 不需要 shell 命令。
- MCP 输出与 CLI 输出语义一致。

## 2. Delta / Cursor 模式

### 目标

让 agent 能稳定获取“上次之后的新内容”，而不是依赖模糊的 `--since 24h`。

### 新增命令建议

```bash
tide changes
tide changes --after <cursor>
tide fetch --cursor
```

### 输出示例

```json
{
  "ok": true,
  "data": {
    "cursor": "2026-06-01T10:00:00Z:entry_123",
    "items": [
      {
        "id": 123,
        "title": "...",
        "url": "...",
        "published_at": "2026-06-01T09:58:00Z"
      }
    ]
  },
  "error": null,
  "meta": {
    "count": 1
  }
}
```

### 数据模型建议

可以维护：

- `entries.created_at`
- `entries.updated_at`
- monotonic cursor
- optional `change_log` table

`change_log` 可记录：

- entry_created
- entry_updated
- feed_failed
- feed_recovered
- state_changed

### 价值

- 支持日报/周报增量生成。
- 支持后台 agent 定时轮询。
- 避免重复处理同一批文章。

## 3. Token-budgeted Content

### 目标

Tide 不做 LLM 总结，但应负责把 RSS 内容整理成适合 LLM 输入的形态。

### 新增参数建议

```bash
tide get 42 --text
tide get 42 --full
tide get 42 --max-chars 4000
tide get 42 --token-budget 2000
tide get 42 --content-only
```

### 内容处理能力

- HTML 清洗。
- 转纯文本。
- 移除 script/style。
- 保留链接 provenance。
- 按字符或估算 token 截断。
- 输出是否被截断。

### 输出示例

```json
{
  "ok": true,
  "data": {
    "id": 42,
    "title": "...",
    "url": "...",
    "text": "...",
    "truncated": true,
    "char_count": 4000,
    "estimated_tokens": 1100
  },
  "error": null,
  "meta": {}
}
```

### 注意

RSS entry content 和网页正文不同。若未来要抓网页正文，应单独新增：

```bash
tide extract 42
```

不要让 `get` 隐式访问网页正文，避免副作用和延迟不可控。

## 4. Feed Health

### 目标

让 agent 知道哪些 feed 可靠、活跃、长期失败或无更新。

### 新增命令建议

```bash
tide health
tide sources --health
tide health --jsonl
```

### 指标建议

- `last_fetched_at`
- `last_success_at`
- `last_error_at`
- `last_error`
- `consecutive_failures`
- `success_rate_7d`
- `avg_latency_ms`
- `entries_7d`
- `entries_30d`
- `stale_days`
- `status`

### 状态建议

- `healthy`
- `stale`
- `failing`
- `dead`
- `unknown`

### 输出示例

```jsonl
{"feed_id":1,"title":"Go Blog","status":"healthy","success_rate_7d":1.0,"entries_7d":3}
{"feed_id":2,"title":"Old Feed","status":"stale","stale_days":120,"entries_7d":0}
```

### 价值

- agent 可跳过低质量源。
- 可自动建议用户清理失效订阅。
- 可为摘要和信息源排序提供权重。

## 5. RAG / Export 能力

### 目标

把 Tide 中的 RSS 内容稳定导出到知识库、向量库、日报系统或外部 agent pipeline。

### 新增命令建议

```bash
tide export entries --format jsonl
tide export entries --format markdown
tide export entries --since 7d
tide export entries --state unprocessed
tide export entries --category ai
```

### JSONL 输出建议

每条 entry 应带完整 provenance：

```json
{
  "id": 42,
  "title": "...",
  "url": "...",
  "feed_id": 1,
  "feed_title": "...",
  "feed_url": "...",
  "published_at": "...",
  "fetched_at": "...",
  "hash": "...",
  "description": "...",
  "content": "..."
}
```

### Markdown 输出建议

适合周报、知识库和静态归档：

```markdown
---
id: 42
url: https://example.com/article
feed: Example
published_at: 2026-06-01T10:00:00Z
---

# Article Title

...
```

### 价值

- 支持 RAG ingest。
- 支持长期知识归档。
- 支持外部 agent 批处理。

## 6. Rules / Routing

### 目标

让 Tide 在抓取后自动完成初步分流，减少 agent 后续筛选成本。

### 新增命令建议

```bash
tide rule add --match "AI|agent|MCP" --tag ai
tide rule add --match "sponsored|advertisement" --state ignored
tide rule add --feed 12 --priority high
tide rule list
tide fetch --apply-rules
```

### Rule 能力

匹配条件：

- title regex
- description regex
- content regex
- feed id
- category
- author

动作：

- add tag
- set state
- set priority
- assign category
- ignore

### 输出示例

```json
{
  "ok": true,
  "data": {
    "rules_applied": 12,
    "entries_tagged": 8,
    "entries_ignored": 4
  },
  "error": null,
  "meta": {}
}
```

### 价值

- 自动把 RSS 信息流变成 agent 可消费队列。
- 降低重复处理成本。
- 适合构建主题监控。

## 7. Hooks / Webhooks

### 目标

让 Tide 抓到新内容后可以触发外部流程。

### 新增命令建议

```bash
tide hook add --event new_entry --exec "./summarize.sh"
tide hook add --event fetch_done --webhook https://example.com/webhook
tide hook list
tide hook remove <id>
```

### 事件建议

- `new_entry`
- `fetch_done`
- `feed_failed`
- `feed_recovered`
- `state_changed`

### 安全建议

- 默认不启用任意命令执行。
- hooks 配置需明确 opt-in。
- webhook payload 必须稳定。
- exec hook 应传 JSON 到 stdin，而不是拼 shell 字符串。

### 价值

- 自动日报/周报。
- 自动同步到知识库。
- 自动触发外部 agent。

## 8. Workspace / Profile

### 目标

支持多项目、多客户、多主题隔离。

### 新增命令建议

```bash
tide --profile work sources
tide --profile ai-research fetch
tide profile create client-a
tide profile list
```

### 数据隔离方式

可以将 profile 映射到不同 SQLite DB：

```text
~/.local/share/tide/profiles/work/tide.db
~/.local/share/tide/profiles/ai-research/tide.db
```

### 价值

- 避免不同任务的信息源混杂。
- 支持 agent 在项目上下文中隔离 RSS 数据。
- 更适合团队或客户项目。

## 9. Query DSL

### 目标

当过滤条件变多后，降低 agent 构造命令的复杂度。

### 新增命令建议

```bash
tide query 'category:ai state:unprocessed since:7d "MCP server"'
tide query 'feed:12 sort:published limit:20'
```

### 支持语法

- `category:<name>`
- `feed:<id>`
- `state:<state>`
- `tag:<tag>`
- `since:<duration>`
- `sort:relevance|published`
- `limit:<n>`
- quoted keyword

### 价值

- 更接近 agent 自然生成查询。
- 避免参数组合爆炸。
- 未来可复用到 MCP tool。

## 10. 去重增强

### 目标

RSS 常有跨 feed 重复、URL 带 tracking 参数、标题轻微变化等问题。Tide 应帮助 agent 减少重复信息。

### 能力建议

- URL canonicalization。
- 移除 tracking query 参数。
- GUID fallback。
- title + published_at hash。
- cross-feed duplicate group。
- near-duplicate detection。

### 新增命令建议

```bash
tide dedupe
tide list --deduped
tide search "AI agent" --deduped
```

### 价值

- 降低摘要重复率。
- 降低 agent token 消耗。
- 提高信息流质量。

## 建议实施顺序

### Phase A：Agent 集成层

1. `tide mcp`
2. MCP tools：search/get/fetch/mark/sources
3. MCP 文档与 skill 更新

### Phase B：增量与上下文控制

1. `changes --after <cursor>`
2. `get --text --max-chars`
3. `get --token-budget`
4. 输出 `truncated` 与 `estimated_tokens`

### Phase C：质量与导出

1. `health`
2. `export entries --format jsonl|markdown`
3. provenance 字段标准化

### Phase D：自动化

1. `rule add/list/remove`
2. `fetch --apply-rules`
3. hooks/webhooks

### Phase E：规模化

1. profiles/workspaces
2. query DSL
3. 去重增强

## 最值得优先做的三件事

如果只能选三个下一步，建议：

1. **MCP Server 模式**
   - 直接把 Tide 从 CLI 工具升级为 agent tool provider。

2. **Delta / Cursor 模式**
   - 让 agent 稳定处理新增内容，避免重复和漏处理。

3. **Token-budgeted Content**
   - 让 Tide 输出天然适合 LLM 上下文的内容，降低 token 成本。

这三项能最明显地区分 Tide 与普通 RSS CLI 或 RSS 阅读器。

## 最终愿景

成熟后的 Tide 应该是：

> AI agent 的 RSS 数据层：负责发现、订阅、抓取、归一化、搜索、增量追踪、内容准备、状态记录和外部集成。

Agent 只负责理解、总结、决策和生成，不再浪费时间处理 RSS 协议细节和一次性数据管道。

