import { GithubLogo } from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

export function Nav() {
  const { locale, setLocale, t } = useLocale()

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 bg-terminal-bg/90 backdrop-blur-sm border-b border-terminal-border" style={{ maxHeight: "72px" }}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <a href="#" className="font-mono text-xl font-bold text-terminal-green tracking-tight">
          tide
        </a>
        <div className="flex items-center gap-4">
          <div className="flex items-center border border-terminal-border rounded-md overflow-hidden">
            <button
              onClick={() => setLocale("en")}
              className={`px-2.5 py-1 text-xs font-mono transition-colors ${
                locale === "en"
                  ? "bg-terminal-green text-terminal-bg"
                  : "text-terminal-dim hover:text-terminal-fg"
              }`}
            >
              EN
            </button>
            <button
              onClick={() => setLocale("zh")}
              className={`px-2.5 py-1 text-xs font-mono transition-colors ${
                locale === "zh"
                  ? "bg-terminal-green text-terminal-bg"
                  : "text-terminal-dim hover:text-terminal-fg"
              }`}
            >
              中文
            </button>
          </div>
          <a
            href="https://github.com/0xfig521/tide"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-terminal-fg hover:text-terminal-green transition-colors"
          >
            <GithubLogo size={20} />
            <span className="hidden sm:inline text-sm">{t.nav.github}</span>
          </a>
        </div>
      </div>
    </nav>
  )
}
