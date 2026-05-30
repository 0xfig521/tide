<pre align="center">
  ████████╗██╗██████╗ ███████╗
  ╚══██╔══╝██║██╔══██╗██╔════╝
     ██║   ██║██║  ██║█████╗
     ██║   ██║██║  ██║██╔══╝
     ██║   ██║██████╔╝███████╗
     ╚═╝   ╚═╝╚═════╝ ╚══════╝
</pre>

<p align="center"><em>一个高速并发的终端 RSS 阅读器。</em></p>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go" alt="Go"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/0xfig521/tide/releases"><img src="https://img.shields.io/github/v/release/0xfig521/tide?style=flat" alt="Release"></a>
</p>

<p align="center"><a href="./README.md">English</a> | 中文</p>

一个高速并发的终端 RSS 阅读器。`tide` 将订阅源存入 SQLite，并行抓取，所有输出为 JSON — 方便管道、脚本或直接浏览。

## 特性

- **⚡ 高并发** — 同时拉取数十个源，带进度条
- **📦 零依赖** — 单文件二进制，SQLite 内嵌，无需运行时
- **🗃️ 分类管理** — 组织源，按分类过滤
- **🔍 全文搜索** — 标题、摘要、正文
- **📡 智能缓存** — ETag / Last-Modified 条件请求，不浪费带宽
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
tide list --unread | jq '.items[] | {title, feed_title}'
```

## 命令一览

| 命令 | 用途 |
|---|---|
| `add <url> [-c <cat>]` | 添加订阅 |
| `remove <id>` | 取消订阅 |
| `sources` | 查看所有源 |
| `list` | 浏览文章（支持筛选、分页、时间范围）|
| `search <kw>` | 全文搜索 |
| `unread` | 未读文章 |
| `fetch [--force]` | 拉取最新 |
| `fetch --daemon` | 后台定时拉取 |
| `schedule` | 管理后台守护进程（start/stop/status/logs）|
| `read <id>` | 标为已读 |
| `star <id>` | 收藏 / 取消 |
| `category` | 分类管理（create/list/assign/remove）|
| `info <id>` | 源详情 |

所有命令默认输出 JSON。`list` 支持 `--format table` 切换为终端表格。

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

## AI Skill

Tide 内置了面向 AI 编程助手（Claude Code、Codex、Cursor 等）的 [skill](https://skills.sh/)。安装后，AI 助手可以直接帮你管理 RSS 订阅：

```bash
npx skills add 0xfig521/tide
```

安装后，你可以直接对 AI 说"帮我找这周关于 Rust 的前 5 篇未读文章"，它知道每一个 tide 命令和参数。

完整 skill 定义见 [`tide/SKILL.md`](./tide/SKILL.md)。

## 技术栈

- [gofeed](https://github.com/mmcdole/gofeed) · RSS/Atom/JSON Feed 解析
- [SQLite](https://sqlite.org) · 嵌入式数据库
- [cobra](https://github.com/spf13/cobra) · CLI 框架
- [lipgloss](https://github.com/charmbracelet/lipgloss) · 终端样式

## 许可

[MIT](./LICENSE)
