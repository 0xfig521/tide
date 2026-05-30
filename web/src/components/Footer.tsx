import { GithubLogo, File, DownloadSimple } from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

export function Footer() {
  const { t } = useLocale()

  return (
    <footer className="border-t border-terminal-border py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-4">
        <span className="font-mono text-sm font-bold text-terminal-green">tide</span>

        <div className="flex items-center gap-6">
          <a
            href="https://github.com/0xfig521/tide"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 text-sm text-terminal-dim hover:text-terminal-green transition-colors"
          >
            <GithubLogo size={16} />
            {t.footer.github}
          </a>
          <a
            href="https://github.com/0xfig521/tide/blob/main/LICENSE"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 text-sm text-terminal-dim hover:text-terminal-green transition-colors"
          >
            <File size={16} />
            {t.footer.license}
          </a>
          <a
            href="https://github.com/0xfig521/tide/releases"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 text-sm text-terminal-dim hover:text-terminal-green transition-colors"
          >
            <DownloadSimple size={16} />
            {t.footer.releases}
          </a>
        </div>
      </div>
    </footer>
  )
}
