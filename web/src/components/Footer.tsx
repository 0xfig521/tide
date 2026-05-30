import { GithubLogo, File, DownloadSimple, Sparkle } from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

export function Footer() {
  const { t } = useLocale()

  return (
    <footer className="relative border-t border-white/5 py-12 mt-12 bg-black/20 overflow-hidden flex justify-center w-full">
      {/* Subtle bottom grid overlay */}
      <div className="absolute inset-0 cyber-grid opacity-10 pointer-events-none" />

      <div className="w-full max-w-5xl px-6 relative z-10">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-8">
          
          {/* Logo & Operational Status */}
          <div className="flex flex-col items-center sm:items-start gap-2">
            <a href="#" className="font-mono text-lg font-black text-transparent bg-clip-text bg-gradient-to-r from-brand-cyan to-brand-violet">
              tide
            </a>
            <div className="flex items-center gap-2 px-2.5 py-1 rounded-full bg-terminal-green/5 border border-terminal-green/10">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-terminal-green opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-terminal-green"></span>
              </span>
              <span className="text-[10px] font-mono tracking-wider uppercase text-terminal-green/80 flex items-center gap-1">
                <Sparkle size={10} /> All Systems Operational
              </span>
            </div>
          </div>

          {/* Links */}
          <div className="flex flex-wrap justify-center items-center gap-8">
            <a
              href="https://github.com/0xfig521/tide"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 text-sm text-terminal-dim hover:text-brand-cyan transition-all duration-300 group"
            >
              <GithubLogo size={18} className="group-hover:rotate-12 transition-transform duration-300" />
              <span className="font-medium">{t.footer.github}</span>
            </a>
            <a
              href="https://github.com/0xfig521/tide/blob/main/LICENSE"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 text-sm text-terminal-dim hover:text-brand-violet transition-all duration-300 group"
            >
              <File size={18} className="group-hover:scale-105 transition-transform duration-300" />
              <span className="font-medium">{t.footer.license}</span>
            </a>
            <a
              href="https://github.com/0xfig521/tide/releases"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 text-sm text-terminal-dim hover:text-brand-pink transition-all duration-300 group"
            >
              <DownloadSimple size={18} className="group-hover:translate-y-0.5 transition-transform duration-300" />
              <span className="font-medium">{t.footer.releases}</span>
            </a>
          </div>
        </div>

        <div className="mt-8 pt-8 border-t border-white/5 text-center">
          <p className="text-xs text-terminal-dim/50 font-mono">
            &copy; {new Date().getFullYear()} Tide CLI. Built for the era of AI Agents.
          </p>
        </div>
      </div>
    </footer>
  )
}
