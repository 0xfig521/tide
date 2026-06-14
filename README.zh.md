<pre align="center">
  ████████╗██╗██████╗ ███████╗
  ╚══██╔══╝██║██╔══██╗██╔════╝
     ██║   ██║██║  ██║█████╗
     ██║   ██║██║  ██║██╔══╝
     ██║   ██║██████╔╝███████╗
     ╚═╝   ╚═╝╚═════╝ ╚══════╝
</pre>

<p align="center"><em>面向 AI agent 与命令行的 RSS 数据适配器。</em></p>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/0xfig-labs/tide/releases"><img src="https://img.shields.io/github/v/release/0xfig-labs/tide?style=flat" alt="Release"></a>
</p>

<p align="center"><a href="./README.md">English</a> | 中文</p>

**Tide** 是为 AI agent 和终端用户设计的 RSS 数据适配器。基于 SQLite 存储，并发拉取，唯一输出语言：**JSON**。每个命令返回稳定的 `{ok, data, error, meta}` 信封 — stdout 干净可解析，stderr 承载进度与日志，结构化错误码配合非零退出码，无需任何解析奇技淫巧。

## 特性

- **🧠 AI 原生** — 稳定 JSON 信封、结构化错误码、stdout/stderr 严格分离
- **⚡ 高并发** — 同时拉取数十个源，进度条走 stderr
- **📦 零依赖** — 单文件二进制，SQLite 内嵌，无需运行时
- **🗃️ 分类管理** — 组织源，按分类过滤
- **🔍 FTS5 全文搜索** — 真正的 `MATCH` 搜索，非 `LIKE %keyword%`
- **📡 智能缓存** — ETag / Last-Modified 条件请求
- **⏱️ 时间过滤** — `--since 24h`、`--since 7d`
- **📄 分页** — `--page`、`--page-size`
- **🤖 守护模式** — `tide schedule start` 后台定时自动抓取
- **↔️ OPML** — `tide import` / `tide export` 在不同 RSS 阅读器间迁移订阅
- **🔄 增量追踪** — `tide changes` 游标模式，避免重复处理
- **📏 Token 预算** — `tide get --max-chars` / `--token-budget`，控制 LLM 上下文成本
- **🔗 MCP 集成** — `tide mcp` 将 Tide 暴露出 MCP 工具供 AI agent 直接调用
- **🏥 源健康度** — `tide health` 自动检测失效/老化订阅源
- **📤 RAG 导出** — `tide export entries` 以 JSONL/Markdown 格式导出到知识库
- **📐 自动分发** — `tide rule add` + `tide fetch --apply-rules` 自动分类和过滤
- **⚡ 性能优化** — 批量事务写入、预编译语句复用、list/search 避免加载大字段 content

## 安装

```bash
# macOS / Linux — 一行安装
curl -fsSL https://raw.githubusercontent.com/0xfig-labs/tide/main/install.sh | bash

# Homebrew
brew install 0xfig-labs/tap/tide

# Go
go install github.com/0xfig-labs/tide@latest
```

## 快速上手

```bash
# 订阅
tide add "https://blog.golang.org/feed.atom" --category "技术"

# 抓取（10 并发）
tide fetch --concurrency 10

# 全文搜索
tide search "kubernetes" --since 7d

# 获取完整内容
tide get 42

# 管道对接
tide list --json | jq '.data.items[] | {title, feed_title}'
```

## 命令一览

| 命令 | 用途 |
|---|---|
| `add <url> [-c <cat>]` | 添加订阅 |
| `remove <id>` | 取消订阅 |
| `discover <url>` | 从网站发现 RSS 订阅源 |
| `mark <id> --state <s>` | 设置处理状态（new/seen/processed/ignored/failed）|
| `sources` | 查看所有源 |
| `import <file>` | 从 OPML 文件导入订阅 |
| `export [--output <f>]` | 导出订阅为 OPML |
| `export entries` | 导出条目（JSONL 或 Markdown）|
| `list` | 浏览文章（JSONL 默认）|
| `search <kw>` | 全文搜索（FTS5）|
| `get <id>` | 获取条目详情（支持 --text、--max-chars、--token-budget）|
| `fetch [--force]` | 拉取最新（支持 --apply-rules）|
| `changes` | 增量获取新增/变更条目 |
| `health` | 查看订阅源健康状态 |
| `failures` | 管理失败源（列出/查看/清除/重试） |
| `rule` | 管理自动分发规则 |
| `schedule` | 管理后台守护进程（start/stop/status/logs）|
| `category` | 分类管理（create/list/assign/remove）|
| `mcp` | 启动 MCP 服务器（AI agent 工具集成）|
| `upgrade` | 自更新到最新版本 |

所有命令默认输出 JSON（稳定 `{ok, data, error, meta}` 信封）。错误返回非零退出码和结构化错误码。

## AI 助手自助安装指引

将以下提示粘贴给你的 AI 助手（Claude、Codex、Cursor 等），让它自动安装并配置 Tide：

> 你有 Shell 工具权限。请按以下步骤在这台机器上安装 Tide：
>
> **第 1 步 — 安装 Tide**
> ```bash
> curl -fsSL https://raw.githubusercontent.com/0xfig-labs/tide/main/install.sh | bash
> ```
> 验证安装：`tide --version`
>
> **第 2 步 — 安装 Agent Skill**（如果你的平台支持技能）
> ```bash
> npx skills add 0xfig-labs/tide
> ```
> 这将让 AI 助手掌握所有 Tide 命令和输出格式的结构化知识。
>
> **第 3 步 — 添加订阅源**
> ```bash
> tide add "https://hnrss.org/frontpage" -c 资讯
> tide add "https://blog.golang.org/feed.atom" -c 技术
> tide fetch --quiet
> ```
>
> **第 4 步 —（可选）启动 MCP 服务器**
> 如果你是 MCP 兼容的 AI 助手（Claude Desktop、Cursor、VS Code），可以让 Tide 以 MCP 模式运行，直接调用工具：
> ```bash
> tide mcp &
> # 暴露的工具：discover_feeds, add_feed, fetch_feeds,
> # search_entries, list_entries, get_entry, mark_entry, get_feed_health
> ```
>
> **第 5 步 — 验证**
> ```bash
> tide list --since 24h --limit 5
> tide health
> ```
>
> 完成上述步骤后，我就拥有了 Tide 作为我的 RSS 数据管道，可以抓取、搜索、总结、追踪内容，无需再每次编写临时的 RSS 解析代码。

## 定时抓取

Tide 可以作为后台守护进程，按计划自动抓取订阅源。

```bash
# 启动守护进程（默认：每 30 分钟，5 并发）
tide schedule start

# 自定义间隔和并发数
tide schedule start --interval 1h --concurrency 10

# 查看守护进程状态
tide schedule status

# 查看最近日志
tide schedule logs -n 20

# 停止守护进程
tide schedule stop
```

守护进程独立于终端会话运行，PID 文件和日志保存在 `~/.local/share/tide/logs/` 中。

## OPML 导入导出

通过标准 OPML 2.0 格式在不同 RSS 阅读器间迁移订阅。

```bash
# 从其他 RSS 阅读器导入订阅
tide import feeds.opml

# 导出所有订阅供备份或迁移
tide export -o tide-backup.opml

# 导出到 stdout（供管道处理）
tide export
```

`tide import` 会保留分类层级和源元数据（标题、站点 URL）。已存在的重复订阅自动跳过。

## 自更新

```bash
# 检查新版本
tide upgrade --check

# 更新到最新版
tide upgrade

# 安装指定版本
tide upgrade --tag v0.2.0
```

Tide 从 GitHub Releases 下载预编译二进制，自动替换当前版本。

## 面向 AI Agent

Tide 说一种语言：JSON。每个命令返回稳定信封：

```json
{"ok": true, "data": {...}, "error": null, "meta": null}
```

无需任何解析黑魔法：

- **stdout** = 纯 JSON。**stderr** = 进度条、日志、诊断。
- **退出码 0** = 成功，**非零** = 失败。同时检查 `.ok` 和退出码。
- **错误码** 为稳定字符串：`feed_not_found`、`entry_not_found`、`feed_already_exists`、`invalid_args`、`internal_error`。
- `tide fetch --quiet` 可屏蔽进度条，保证 stdout 纯净。

```bash
tide fetch --quiet                     # 静默抓取，stdout 纯 JSON
tide search "rust async" --since 7d    # FTS5 全文搜索近 7 天
tide get 42                            # 获取完整条目（含描述和正文）
```

安装 skill 后，AI 助手可直接管理 RSS：

```bash
npx skills add 0xfig-labs/tide
```

完整 skill 见 [`tide/SKILL.md`](./tide/SKILL.md)。

## 技术栈

- [gofeed](https://github.com/mmcdole/gofeed) · RSS/Atom/JSON Feed 解析
- [SQLite](https://sqlite.org) · 嵌入式数据库
- [cobra](https://github.com/spf13/cobra) · CLI 框架
- [lipgloss](https://github.com/charmbracelet/lipgloss) · 终端样式

## 许可

[MIT](./LICENSE)
