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
  <a href="https://github.com/0xfig521/tide/releases"><img src="https://img.shields.io/github/v/release/0xfig521/tide?style=flat" alt="Release"></a>
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

## 安装

```bash
# macOS / Linux — 一行安装
curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash

# Homebrew
brew install 0xfig521/tap/tide

# Go
go install github.com/0xfig521/tide@latest
```

## 快速上手

```bash
# 订阅
tide add "https://blog.golang.org/feed.atom" --category "技术"

# 抓取（10 并发）
tide fetch --concurrency 10

# 查看 24 小时内未读
tide list --unread --since 24h

# 全文搜索
tide search "kubernetes"

# 已读、收藏
tide read 3
tide star 7

# 管道对接
tide list --unread | jq '.data.items[] | {title, feed_title}'
```

## 命令一览

| 命令 | 用途 |
|---|---|
| `add <url> [-c <cat>]` | 添加订阅 |
| `remove <id>` | 取消订阅 |
| `sources` | 查看所有源 |
| `list` | 浏览文章（支持筛选、分页、时间范围）|
| `search <kw>` | 全文搜索（FTS5）|
| `unread` | 未读文章 |
| `get <id>` | 获取文章完整详情（描述、正文）|
| `fetch [--force]` | 拉取最新 |
| `fetch --daemon` | 后台定时拉取 |
| `schedule` | 管理后台守护进程（start/stop/status/logs）|
| `read <id>` | 标为已读 |
| `star <id>` | 收藏 / 取消 |
| `category` | 分类管理（create/list/assign/remove）|
| `upgrade` | 自更新到最新版本 |
| `info <id>` | 源详情 |

所有命令默认输出 JSON（稳定 `{ok, data, error, meta}` 信封）。`list` 支持 `--format table` 切换为终端表格。错误返回非零退出码和结构化错误码。

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
tide read 42                           # 标记已读
```

安装 skill 后，AI 助手可直接管理 RSS：

```bash
npx skills add 0xfig521/tide
```

完整 skill 见 [`tide/SKILL.md`](./tide/SKILL.md)。

## 技术栈

- [gofeed](https://github.com/mmcdole/gofeed) · RSS/Atom/JSON Feed 解析
- [SQLite](https://sqlite.org) · 嵌入式数据库
- [cobra](https://github.com/spf13/cobra) · CLI 框架
- [lipgloss](https://github.com/charmbracelet/lipgloss) · 终端样式

## 许可

[MIT](./LICENSE)
