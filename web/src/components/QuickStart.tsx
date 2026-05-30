import { useState } from "react"
import { useLocale } from "../i18n/context"
import { Terminal, Key, ShieldCheck } from "@phosphor-icons/react"

const CLI_COMMANDS = [
  {
    id: "add",
    name: "tide add",
    desc_en: "Subscribe to a new Atom/RSS feed URL under a specific category.",
    desc_zh: "在指定分类下订阅一个新的 Atom/RSS 订阅源 URL。",
    usage: "tide add <feed_url> --category <name>",
    flags: [
      { name: "--category, -c", desc_en: "Organize feeds into groups", desc_zh: "对订阅源进行分类管理" },
      { name: "--title, -t", desc_en: "Custom feed title override", desc_zh: "覆盖订阅源的默认标题" },
    ],
    example: 'tide add "https://blog.golang.org/feed.atom" --category "tech"',
  },
  {
    id: "fetch",
    name: "tide fetch",
    desc_en: "Pull latest articles concurrently from all subscribed feeds.",
    desc_zh: "并发从所有订阅源中拉取最新的文章列表。",
    usage: "tide fetch --concurrency <number>",
    flags: [
      { name: "--concurrency", desc_en: "Parallel requests count (default: 5)", desc_zh: "并发拉取的线程数 (默认: 5)" },
      { name: "--timeout", desc_en: "Request timeout limit in seconds", desc_zh: "网络请求超时时间 (秒)" },
    ],
    example: "tide fetch --concurrency 10",
  },
  {
    id: "search",
    name: "tide search",
    desc_en: "Search cached database articles with filtering criteria.",
    desc_zh: "使用特定筛选标准搜索已缓存的数据库文章。",
    usage: "tide search <query> --since <time>",
    flags: [
      { name: "--since, -s", desc_en: "Time window (e.g. 24h, 7d)", desc_zh: "搜索的时间跨度 (例如: 24h, 7d)" },
      { name: "--limit, -l", desc_en: "Limit search output count", desc_zh: "限制搜索输出的数量" },
    ],
    example: 'tide search "kubernetes" --since 7d --limit 5',
  },
  {
    id: "list",
    name: "tide list",
    desc_en: "Retrieve article feeds with custom unread/read state matching.",
    desc_zh: "检索特定未读/已读状态的订阅文章。",
    usage: "tide list --unread --since <time>",
    flags: [
      { name: "--unread, -u", desc_en: "Only list unread articles", desc_zh: "只列出未读的文章" },
      { name: "--category", desc_en: "Filter output by category", desc_zh: "过滤指定分类的文章" },
    ],
    example: "tide list --unread --since 24h",
  },
  {
    id: "daemon",
    name: "tide daemon",
    desc_en: "Launch the background fetch scheduler daemon.",
    desc_zh: "启动后台定时自动拉取订阅的守护进程。",
    usage: "tide daemon start --interval <time>",
    flags: [
      { name: "--interval", desc_en: "Fetch frequency (e.g. 1h, 30m)", desc_zh: "后台自动抓取的间隔频率" },
      { name: "--port", desc_en: "Local IPC control server port", desc_zh: "本地 IPC 控制服务器的端口" },
    ],
    example: "tide daemon start --interval 2h",
  },
]

export function QuickStart() {
  const { locale, t } = useLocale()
  const [activeCmd, setActiveCmd] = useState(0)

  const cmd = CLI_COMMANDS[activeCmd]

  return (
    <section className="py-32 border-t border-white/5 relative overflow-hidden">
      <div className="absolute inset-0 cyber-grid opacity-10 pointer-events-none" />
      <div className="absolute top-1/2 left-1/4 -translate-y-1/2 w-[30vw] h-[30vw] rounded-full bg-brand-cyan/5 blur-[120px] pointer-events-none" />

      <div className="max-w-6xl mx-auto px-6 relative z-10">
        
        {/* Title */}
        <div className="mb-16 text-center lg:text-left">
          <h2 className="text-4xl sm:text-5xl font-extrabold tracking-tight text-white">
            {t.quickstart.headline}
          </h2>
          <div className="h-1 w-20 bg-gradient-to-r from-brand-cyan to-brand-violet mt-4 rounded-full mx-auto lg:mx-0" />
        </div>

        {/* Dashboard Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-stretch">
          
          {/* Subcommand Selector Panel (col-span-4) */}
          <div className="lg:col-span-4 flex flex-col gap-3">
            <div className="text-xs font-mono font-bold tracking-wider text-white/40 uppercase mb-2 pl-2">
              Select Command
            </div>
            {CLI_COMMANDS.map((c, idx) => (
              <button
                key={c.id}
                onClick={() => setActiveCmd(idx)}
                className={`flex items-center justify-between p-4 rounded-xl border font-mono text-xs transition-all duration-300 ${
                  activeCmd === idx
                    ? "bg-gradient-to-r from-brand-cyan/10 to-brand-violet/10 border-brand-cyan/35 text-brand-cyan shadow-lg shadow-brand-cyan/5 scale-[1.02]"
                    : "bg-slate-950/40 border-white/5 text-terminal-dim hover:text-white hover:border-white/10 hover:bg-slate-950/80"
                }`}
              >
                <span className="font-semibold text-sm">{c.name}</span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-white/5 border border-white/10 group-hover:bg-white/10 font-bold uppercase text-white/40">
                  {c.id}
                </span>
              </button>
            ))}
          </div>

          {/* Code Schema Terminal (col-span-8) */}
          <div className="lg:col-span-8">
            <div className="relative rounded-2xl glass-panel overflow-hidden border border-white/10 shadow-2xl bg-slate-950/80 h-full flex flex-col justify-between">
              
              {/* Terminal Title Bar */}
              <div className="flex items-center justify-between px-5 py-3.5 border-b border-white/5 bg-slate-900/50">
                <span className="font-mono text-xs text-white/40 flex items-center gap-1.5">
                  <Terminal size={14} /> tide-cli help --manual
                </span>
                <div className="flex items-center gap-1.5">
                  <span className="w-2 h-2 rounded-full bg-brand-cyan shadow-[0_0_8px_var(--color-brand-cyan)] animate-pulse" />
                  <span className="font-mono text-[9px] uppercase tracking-wider text-brand-cyan">Active Engine</span>
                </div>
              </div>

              {/* Terminal Area */}
              <div className="p-6 font-mono text-xs leading-relaxed space-y-6 flex-1 flex flex-col justify-between">
                
                {/* Description & Command format */}
                <div className="space-y-4">
                  <div>
                    <span className="text-white/40"># Description:</span>
                    <p className="text-sm font-semibold text-white/90 mt-1 leading-relaxed">
                      {locale === "zh" ? cmd.desc_zh : cmd.desc_en}
                    </p>
                  </div>

                  <div>
                    <span className="text-white/40"># Usage Syntax:</span>
                    <div className="bg-black/40 border border-white/5 rounded-lg p-3 text-brand-cyan font-bold tracking-tight text-sm mt-1.5 flex items-center gap-2">
                      <span className="text-white/30">$</span> {cmd.usage}
                    </div>
                  </div>
                </div>

                {/* Command Flags list */}
                <div>
                  <span className="text-white/40 block mb-2"># Supported Flags:</span>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    {cmd.flags.map((flag) => (
                      <div key={flag.name} className="p-2.5 rounded-lg border border-white/5 bg-slate-900/40 space-y-1">
                        <div className="font-semibold text-brand-pink flex items-center gap-1">
                          <Key size={12} /> {flag.name}
                        </div>
                        <p className="text-[10px] text-terminal-dim">
                          {locale === "zh" ? flag.desc_zh : flag.desc_en}
                        </p>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Command Example */}
                <div className="pt-4 border-t border-white/5">
                  <span className="text-white/40 block mb-2"># Run Example:</span>
                  <div className="flex items-center gap-2 text-terminal-green text-sm bg-black/20 p-2.5 rounded border border-white/5">
                    <ShieldCheck size={16} className="text-terminal-green shrink-0" />
                    <span>$ {cmd.example}</span>
                  </div>
                </div>

              </div>

            </div>
          </div>

        </div>

      </div>
    </section>
  )
}
