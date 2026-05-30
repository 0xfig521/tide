import { useState } from "react"
import { Copy } from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

const installMethods = [
  {
    label: "curl",
    command: "curl -fsSL https://raw.githubusercontent.com/0xfig521/tide/main/install.sh | bash",
  },
  {
    label: "Homebrew",
    command: "brew install 0xfig521/tap/tide",
  },
  {
    label: "Go",
    command: "go install github.com/0xfig521/tide@latest",
  },
]

function CodeBlock({ command, label }: { command: string; label: string }) {
  const { t } = useLocale()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(command)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="bg-terminal-border/50 rounded-lg border border-terminal-border overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 border-b border-terminal-border">
        <span className="font-mono text-xs text-terminal-dim">{label}</span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1.5 text-xs text-terminal-dim hover:text-terminal-green transition-colors"
          aria-label={`${t.install.copy} ${label}`}
        >
          <Copy size={14} />
          {copied ? t.install.copied : t.install.copy}
        </button>
      </div>
      <div className="px-4 py-3 font-mono text-sm">
        <span className="text-terminal-green">$</span>{" "}
        <span className="text-terminal-fg">{command}</span>
      </div>
    </div>
  )
}

export function Install() {
  const { t } = useLocale()

  return (
    <section id="install" className="py-24 border-t border-terminal-border">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h2 className="font-sans text-3xl sm:text-4xl font-bold tracking-tighter text-terminal-fg">
          {t.install.headline}
        </h2>

        <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-6">
          {installMethods.map((method) => (
            <CodeBlock key={method.label} command={method.command} label={method.label} />
          ))}
        </div>
      </div>
    </section>
  )
}
