import { Terminal, Copy } from "@phosphor-icons/react"
import { useState } from "react"
import { useLocale } from "../i18n/context"

const SKILL_COMMAND = "npx skills add 0xfig521/tide"

export function AISkill() {
  const { t } = useLocale()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(SKILL_COMMAND)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <section className="py-24 border-t border-terminal-border bg-terminal-border/30">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-12 items-start">
          <div className="lg:col-span-3">
            <h2 className="font-sans text-3xl sm:text-4xl font-bold tracking-tighter text-terminal-fg">
              {t.aiSkill.headline1}
              <br />
              <span className="text-terminal-green">{t.aiSkill.headline2}</span>
            </h2>
            <p className="mt-6 text-lg text-terminal-dim max-w-[65ch] leading-relaxed">
              {t.aiSkill.description1}
            </p>
            <p className="mt-4 text-terminal-dim max-w-[65ch] leading-relaxed">
              {t.aiSkill.description2}
            </p>
          </div>

          <div className="lg:col-span-2">
            <div className="border border-terminal-border rounded-lg bg-terminal-bg overflow-hidden">
              <div className="flex items-center gap-3 px-4 py-3 border-b border-terminal-border bg-terminal-border/50">
                <div className="w-8 h-8 rounded-md bg-terminal-green/10 flex items-center justify-center text-terminal-green">
                  <Terminal size={18} weight="bold" />
                </div>
                <div>
                  <div className="font-mono text-sm font-semibold text-terminal-fg">tide</div>
                  <div className="font-mono text-xs text-terminal-dim">{t.aiSkill.cardSubtitle}</div>
                </div>
              </div>

              <div className="p-4">
                <div className="flex items-center gap-2 mb-3">
                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono bg-terminal-green/10 text-terminal-green border border-terminal-green/20">
                    skills.sh
                  </span>
                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono bg-terminal-border text-terminal-dim">
                    Go CLI
                  </span>
                </div>

                <div className="bg-terminal-border/50 rounded border border-terminal-border p-3 font-mono text-sm">
                  <span className="text-terminal-green">$</span>{" "}
                  <span className="text-terminal-fg">{SKILL_COMMAND}</span>
                </div>

                <button
                  onClick={handleCopy}
                  className="mt-3 flex items-center gap-2 text-xs text-terminal-dim hover:text-terminal-green transition-colors font-mono"
                >
                  <Copy size={14} />
                  {copied ? t.install.copied : t.aiSkill.copyCommand}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
