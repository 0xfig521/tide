import { useEffect, useRef } from "react"
import gsap from "gsap"
import { ScrollTrigger } from "gsap/ScrollTrigger"
import { useLocale } from "../i18n/context"

gsap.registerPlugin(ScrollTrigger)

const ASCII_LOGO = `
  ████████╗██╗██████╗ ███████╗
  ╚══██╔══╝██║██╔══██╗██╔════╝
     ██║   ██║██║  ██║█████╗
     ██║   ██║██║  ██║██╔══╝
     ██║   ██║██████╔╝███████╗
     ╚═╝   ╚═╝╚═════╝ ╚══════╝
`

const TYPING_TEXT = "$ tide list --unread --since 24h | jq '.items[] | .title'"

export function Hero() {
  const { t } = useLocale()
  const terminalRef = useRef<HTMLDivElement>(null)
  const cursorRef = useRef<HTMLSpanElement>(null)
  const textRef = useRef<HTMLSpanElement>(null)

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    if (prefersReducedMotion || !textRef.current || !cursorRef.current) return

    const ctx = gsap.context(() => {
      gsap.to(cursorRef.current, {
        opacity: 0,
        duration: 0.5,
        repeat: -1,
        yoyo: true,
        ease: "steps(1)",
      })

      const chars = TYPING_TEXT.split("")
      let idx = 0
      const tl = gsap.timeline({ delay: 0.8 })

      chars.forEach(() => {
        tl.call(() => {
          if (textRef.current && idx < chars.length) {
            textRef.current.textContent += chars[idx]
            idx++
          }
        }, [], idx * 0.04)
      })
    }, terminalRef)

    return () => ctx.revert()
  }, [])

  return (
    <section className="pt-24 min-h-[100dvh] flex items-center">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 w-full">
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-12 items-center">
          <div className="lg:col-span-3">
            <h1 className="font-sans text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tighter leading-tight text-terminal-fg">
              {t.hero.headlinePart1}{" "}
              <span className="text-terminal-green">{t.hero.headlineAccent}</span>
            </h1>
            <p className="mt-6 text-lg text-terminal-dim max-w-[65ch] leading-relaxed">
              {t.hero.description}
            </p>
            <div className="mt-8 flex flex-wrap gap-4">
              <a
                href="#install"
                className="inline-flex items-center justify-center px-6 py-3 font-mono text-sm font-semibold text-terminal-bg bg-terminal-green rounded-lg hover:bg-terminal-green/90 transition-colors"
              >
                {t.hero.getStarted}
              </a>
              <a
                href="https://github.com/0xfig521/tide"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center px-6 py-3 font-mono text-sm font-semibold text-terminal-fg border border-terminal-border rounded-lg hover:border-terminal-green/30 transition-colors"
              >
                {t.hero.viewOnGitHub}
              </a>
            </div>
          </div>

          <div className="lg:col-span-2 hidden lg:block">
            <pre className="font-mono text-xs sm:text-sm text-terminal-dim leading-tight whitespace-pre">
              {ASCII_LOGO}
            </pre>
          </div>
        </div>

        <div
          ref={terminalRef}
          className="mt-12 max-w-2xl bg-terminal-border/50 rounded-lg border border-terminal-border p-4 font-mono text-sm"
        >
          <span className="text-terminal-green">$</span>{" "}
          <span ref={textRef} className="text-terminal-fg" />
          <span
            ref={cursorRef}
            className="inline-block w-2 h-4 bg-terminal-green ml-0.5 align-middle"
          />
        </div>
      </div>
    </section>
  )
}
