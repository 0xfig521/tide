# tide

> 终端里的高速 RSS 阅读器。

**tide** 把你的信息流搬进命令行。订阅源、分类管理、搜索阅读，全都输出干净的 JSON，方便管道和脚本对接。

---

## 为什么用 tide？

- **全程 JSON** — 所有命令输出结构化 JSON。`jq`、`fzf`、脚本随便接。
- **并发抓取** — 同时拉取数十个源，带进度条。
- **智能缓存** — 基于 ETag 的条件请求，不浪费带宽。
- **分类管理** — 按分类组织源，按分类筛选浏览。
- **分页 + 时间过滤** — `--page`、`--page-size`、`--since 24h`，快速定位。
- **守护模式** — `tide fetch --daemon` 后台定时抓取。

---

## 安装

```bash
# 一行安装（推荐）
curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash

# Homebrew
brew install ./Formula/tide.rb

# Go 工具链
go install github.com/0xfig521/tide@latest
```

Shell 脚本自动识别系统和架构，下载最新二进制，安装到 `/usr/local/bin`。

---

## 快速上手

```bash
# 添加第一个源
tide add "https://blog.golang.org/feed.atom" --category "技术"

# 抓取文章（10 并发）
tide fetch --concurrency 10

# 查看 24 小时内的未读
tide list --unread --since 24h

# 搜索
tide search "kubernetes"

# 标记已读、收藏
tide read 3
tide star 7

# 管道对接
tide list --unread | jq '.items[] | {title, feed_title}'
```

---

## 命令一览

| 命令 | 用途 |
|------|------|
| `add <url> --category <name>` | 添加订阅源 |
| `remove <id>` | 取消订阅 |
| `sources` | 查看所有订阅源 |
| `list` | 浏览文章，支持筛选、分页、时间范围 |
| `search <keyword>` | 全文搜索 |
| `unread` | 查看未读 |
| `fetch` | 拉取最新文章 |
| `fetch --daemon` | 后台定时拉取 |
| `read <id>` | 标为已读 |
| `star <id>` | 收藏 / 取消收藏 |
| `category create/list/assign/remove` | 分类管理 |
| `info <id>` | 查看源详情 |

---

## 技术栈

- [gofeed](https://github.com/mmcdole/gofeed) — RSS/Atom/JSON Feed 解析
- [SQLite](https://sqlite.org) — 嵌入式数据库，零配置
- [cobra](https://github.com/spf13/cobra) — CLI 框架
