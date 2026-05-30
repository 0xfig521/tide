export interface TranslationSchema {
  nav: { github: string }
  hero: {
    headlinePart1: string
    headlineAccent: string
    description: string
    getStarted: string
    viewOnGitHub: string
  }
  features: {
    headline: string
    items: { title: string; description: string }[]
  }
  install: { headline: string; copy: string; copied: string }
  quickstart: { headline: string }
  aiSkill: {
    headline1: string
    headline2: string
    description1: string
    description2: string
    cardSubtitle: string
    copyCommand: string
  }
  footer: { github: string; license: string; releases: string }
}

export type Locale = "en" | "zh"

export const translations: Record<Locale, TranslationSchema> = {
  en: {
    nav: { github: "GitHub" },
    hero: {
      headlinePart1: "RSS for your",
      headlineAccent: "AI agent",
      description:
        "JSON-native RSS management for AI agents. Your AI can read, search, and filter articles.",
      getStarted: "Get Started",
      viewOnGitHub: "View on GitHub",
    },
    features: {
      headline: "Why your agent needs Tide",
      items: [
        { title: "AI-Native Skill", description: "Ships with a skills.sh definition. Your AI coding agent can manage RSS feeds out of the box." },
        { title: "JSON by Default", description: "Every command outputs structured JSON. No parsing, no regex, no fragile text scraping." },
        { title: "Concurrent Fetching", description: "Pulls dozens of feeds in parallel with a progress bar. No more waiting on slow feeds." },
        { title: "Zero Dependencies", description: "Single binary, SQLite embedded. No runtime, no package manager, no node_modules." },
        { title: "Smart Caching", description: "ETag and Last-Modified conditional requests. No wasted bandwidth on unchanged feeds." },
        { title: "Daemon Mode", description: "Background scheduled fetching that persists across terminal sessions. Set it and forget it." },
      ],
    },
    install: { headline: "Install in one line", copy: "Copy", copied: "Copied" },
    quickstart: { headline: "Your agent's first RSS command" },
    aiSkill: {
      headline1: "AI agents speak JSON.",
      headline2: "So does Tide.",
      description1:
        "Tide ships with a skills.sh definition for AI coding agents — Claude Code, Codex, Cursor, and more. Install once, and your agent gets full knowledge of every command, flag, and workflow.",
      description2:
        'Say "find me the top 5 articles about Rust this week" and it just works. No parsing HTML, no scraping, no regex. Pure JSON.',
      cardSubtitle: "RSS management for AI agents",
      copyCommand: "Copy command",
    },
    footer: { github: "GitHub", license: "License", releases: "Releases" },
  },
  zh: {
    nav: { github: "GitHub" },
    hero: {
      headlinePart1: "给 AI 智能体的",
      headlineAccent: "RSS 管理工具",
      description: "面向 AI 智能体的 JSON 原生 RSS 管理工具。你的 AI 可以直接读取、搜索和筛选订阅内容。",
      getStarted: "快速开始",
      viewOnGitHub: "在 GitHub 查看",
    },
    features: {
      headline: "为什么你的智能体需要 Tide",
      items: [
        { title: "AI 原生技能", description: "内置 skills.sh 技能定义，你的 AI 编程助手开箱即能管理 RSS 订阅。" },
        { title: "默认 JSON 输出", description: "所有命令输出结构化 JSON。无需解析、无需正则、无需脆弱的文本抓取。" },
        { title: "并发抓取", description: "数十个订阅源并行拉取，带进度条。不再等待慢速源。" },
        { title: "零依赖", description: "单一二进制文件，内嵌 SQLite。无需运行时、无需包管理器、无需 node_modules。" },
        { title: "智能缓存", description: "ETag 和 Last-Modified 条件请求。未变更的源不会浪费带宽。" },
        { title: "守护进程模式", description: "跨终端会话的后台定时抓取。设置一次即可。" },
      ],
    },
    install: { headline: "一行安装", copy: "复制", copied: "已复制" },
    quickstart: { headline: "你的智能体的第一条 RSS 命令" },
    aiSkill: {
      headline1: "AI 智能体说 JSON。",
      headline2: "Tide 也是。",
      description1:
        "Tide 内置了面向 AI 编程助手的 skills.sh 技能定义 — 支持 Claude Code、Codex、Cursor 等。安装一次，你的智能体就能掌握所有命令、参数和工作流。",
      description2:
        "说一句「帮我找出本周 Rust 相关的 5 篇文章」，它就能做到。无需解析 HTML、无需抓取、无需正则。纯 JSON。",
      cardSubtitle: "AI 智能体的 RSS 管理工具",
      copyCommand: "复制命令",
    },
    footer: { github: "GitHub", license: "许可证", releases: "发布" },
  },
}
