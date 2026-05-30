import { useLocale } from "../i18n/context"

const commands = [
  { type: "prompt" as const, text: 'tide add "https://blog.golang.org/feed.atom" --category "tech"' },
  { type: "success" as const, text: "✓ Subscribed" },
  { type: "prompt" as const, text: "tide fetch --concurrency 10" },
  { type: "success" as const, text: "✓ Fetched 42 articles from 12 feeds" },
  { type: "prompt" as const, text: 'tide search "kubernetes" --since 7d' },
  { type: "output" as const, text: '[{"title":"K8s 1.30 released","feed":"Go Blog","published":"2025-04-15T10:00:00Z",...}]' },
]

export function QuickStart() {
  const { t } = useLocale()

  return (
    <section className="py-24 border-t border-terminal-border">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h2 className="font-sans text-3xl sm:text-4xl font-bold tracking-tighter text-terminal-fg">
          {t.quickstart.headline}
        </h2>

        <div className="mt-8 bg-terminal-border/30 rounded-lg border border-terminal-border overflow-hidden">
          <div className="flex items-center gap-2 px-4 py-2 border-b border-terminal-border bg-terminal-border/50">
            <div className="w-3 h-3 rounded-full bg-terminal-red/60" />
            <div className="w-3 h-3 rounded-full bg-terminal-yellow/60" />
            <div className="w-3 h-3 rounded-full bg-terminal-green/60" />
            <span className="ml-2 font-mono text-xs text-terminal-dim">tide session</span>
          </div>

          <div className="p-4 sm:p-6 font-mono text-sm leading-relaxed">
            {commands.map((cmd, i) => {
              if (cmd.type === "prompt") {
                return (
                  <div key={i} className="mt-2 first:mt-0">
                    <span className="text-terminal-green">$</span>{" "}
                    <span className="text-terminal-fg">{cmd.text}</span>
                  </div>
                )
              }
              if (cmd.type === "success") {
                return (
                  <div key={i} className="text-terminal-green">
                    {cmd.text}
                  </div>
                )
              }
              return (
                <div key={i} className="text-terminal-dim break-all">
                  {cmd.text}
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </section>
  )
}
