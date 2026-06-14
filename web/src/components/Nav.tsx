import { useEffect, useState } from "react"
import { GithubLogo } from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

export function Nav() {
  const { locale, setLocale, t } = useLocale()
  const [scrolled, setScrolled] = useState(false)

  useEffect(() => {
    const handleScroll = () => {
      if (window.scrollY > 20) {
        setScrolled(true)
      } else {
        setScrolled(false)
      }
    }
    window.addEventListener("scroll", handleScroll)
    return () => window.removeEventListener("scroll", handleScroll)
  }, [])

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ease-out px-4 flex justify-center w-full ${
        scrolled ? "pt-3" : "pt-6"
      }`}
    >
      <nav
        className={`w-full max-w-5xl rounded-full transition-all duration-500 ease-out border ${
          scrolled
            ? "glass-panel bg-opacity-70 shadow-2xl py-3 px-6"
            : "border-transparent bg-transparent py-4 px-6"
        }`}
      >
        <div className="flex items-center justify-between w-full">
          {/* Logo */}
          <a
            href="#"
            className="group flex items-center gap-2 font-mono text-xl font-extrabold tracking-tighter text-transparent bg-clip-text bg-gradient-to-r from-brand-cyan via-brand-violet to-brand-pink"
          >
            tide
            <span className="w-1.5 h-1.5 rounded-full bg-terminal-green animate-pulse inline-block" />
          </a>

          {/* Controls */}
          <div className="flex items-center gap-6">
            {/* Locale Toggles */}
            <div className="flex items-center bg-white/5 rounded-full p-0.5 border border-white/10 overflow-hidden">
              <button
                onClick={() => setLocale("en")}
                className={`px-3 py-1 rounded-full text-xs font-mono font-medium transition-all duration-300 ${
                  locale === "en"
                    ? "bg-gradient-to-r from-brand-cyan to-brand-violet text-white shadow-md shadow-brand-cyan/20 scale-105"
                    : "text-terminal-dim hover:text-terminal-fg"
                }`}
              >
                EN
              </button>
              <button
                onClick={() => setLocale("zh")}
                className={`px-3 py-1 rounded-full text-xs font-mono font-medium transition-all duration-300 ${
                  locale === "zh"
                    ? "bg-gradient-to-r from-brand-violet to-brand-pink text-white shadow-md shadow-brand-violet/20 scale-105"
                    : "text-terminal-dim hover:text-terminal-fg"
                }`}
              >
                中文
              </button>
            </div>

            {/* GitHub Link */}
            <a
              href="https://github.com/0xfig-labs/tide"
              target="_blank"
              rel="noopener noreferrer"
              className="relative group flex items-center gap-2 text-terminal-fg hover:text-brand-cyan transition-colors py-1.5 px-3 rounded-full hover:bg-white/5 border border-transparent hover:border-white/5 duration-300"
            >
              <GithubLogo size={20} className="group-hover:rotate-6 transition-transform duration-300" />
              <span className="hidden sm:inline text-sm font-semibold tracking-tight">{t.nav.github}</span>
            </a>
          </div>
        </div>
      </nav>
    </header>
  )
}
