import { useEffect, useRef, useState } from "react"
import gsap from "gsap"
import { useLocale } from "../i18n/context"
import { Sparkle, Terminal, ArrowRight, Rss, Brain, BookOpen } from "@phosphor-icons/react"

const TYPING_TEXT = "tide list --since 24h"

const MOCK_ARTICLES = [
  {
    id: 1,
    title: "Building AI Agents with Go & SQLite",
    feed: "Go Blog",
    published: "4h ago",
    category: "AI & Go",
    summary: "How to bundle a fully functional AI-native RSS parsing engine into a zero-dependency Go CLI.",
  },
  {
    id: 2,
    title: "Why JSON-Native Output Rules CLI Tools",
    feed: "DevTo",
    published: "12h ago",
    category: "CLI Design",
    summary: "Traditional scraping is fragile. JSON structured output ensures LLMs never break during orchestration.",
  },
  {
    id: 3,
    title: "The Future of RSS in the LLM Era",
    feed: "Hacker News",
    published: "1d ago",
    category: "AI Agents",
    summary: "How standard RSS feeds are turning into high-density memory pools for autonomous software developers.",
  },
]

export function Hero() {
  const { t } = useLocale()
  const textRef = useRef<HTMLSpanElement>(null)
  const cursorRef = useRef<HTMLSpanElement>(null)
  const jsonRef = useRef<HTMLPreElement>(null)
  const cardsContainerRef = useRef<HTMLDivElement>(null)
  const [showJson, setShowJson] = useState(() => 
    typeof window !== "undefined" ? window.matchMedia("(prefers-reduced-motion: reduce)").matches : false
  )
  const [showCards, setShowCards] = useState(() => 
    typeof window !== "undefined" ? window.matchMedia("(prefers-reduced-motion: reduce)").matches : false
  )
  const [glowActive, setGlowActive] = useState(false)

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    if (prefersReducedMotion) return

    let isMounted = true

    // Blinking cursor
    gsap.to(cursorRef.current, {
      opacity: 0,
      duration: 0.5,
      repeat: -1,
      yoyo: true,
      ease: "steps(1)",
    })

    const runAnimationCycle = () => {
      if (!isMounted) return

      // Reset states
      setShowJson(false)
      setShowCards(false)
      setGlowActive(false)
      if (textRef.current) textRef.current.textContent = ""
      if (jsonRef.current) jsonRef.current.style.height = "auto"

      const chars = TYPING_TEXT.split("")
      const typingTimeline = gsap.timeline()

      // 1. Typing command
      chars.forEach((char, charIdx) => {
        typingTimeline.call(() => {
          if (textRef.current && isMounted) {
            textRef.current.textContent += char
          }
        }, [], charIdx * 0.05)
      })

      // Add a 1.2s pause after typing finishes before rendering JSON
      typingTimeline.to({}, { duration: 1.2 })

      // 2. Output JSON after typing
      typingTimeline.call(() => {
        if (!isMounted) return
        setShowJson(true)
        // Animate JSON height & opacity
        gsap.fromTo(
          jsonRef.current,
          { height: 0, opacity: 0 },
          { height: "auto", opacity: 1, duration: 0.6, ease: "power2.out" }
        )
      })

      // Add a 5.0s pause to allow the user to read the JSON output
      typingTimeline.to({}, { duration: 5.0 })

      // 3. Trigger glow effect ("magic conversion")
      typingTimeline.call(() => {
        if (!isMounted) return
        setGlowActive(true)
      })

      // Add a 1.5s pause to showcase the glowing scanner line
      typingTimeline.to({}, { duration: 1.5 })

      // 4. Collapse JSON & Fade In Cards
      typingTimeline.call(() => {
        if (!isMounted) return
        setGlowActive(false)
        
        // GSAP animate JSON out
        gsap.to(jsonRef.current, {
          height: 0,
          opacity: 0,
          duration: 0.5,
          ease: "power2.inOut",
          onComplete: () => {
            if (!isMounted) return
            setShowJson(false)
            setShowCards(true)
            
            // Stagger anim for visual RSS cards
            const cards = cardsContainerRef.current?.querySelectorAll(".rss-demo-card")
            if (cards) {
              gsap.fromTo(
                cards,
                { y: 30, scale: 0.95, opacity: 0 },
                {
                  y: 0,
                  scale: 1,
                  opacity: 1,
                  stagger: 0.12,
                  duration: 0.7,
                  ease: "elastic.out(1, 0.75)",
                }
              )
            }
          }
        })
      })

      // Add a 9.0s pause to allow the user to read the beautiful visual cards
      typingTimeline.to({}, { duration: 9.0 })

      // 5. Rest & Loop
      typingTimeline.call(() => {
        if (isMounted) {
          // Fade out cards to reset
          const cards = cardsContainerRef.current?.querySelectorAll(".rss-demo-card")
          if (cards) {
            gsap.to(cards, {
              opacity: 0,
              y: -10,
              duration: 0.4,
              stagger: 0.05,
              onComplete: () => {
                if (isMounted) runAnimationCycle()
              }
            })
          } else {
            runAnimationCycle()
          }
        }
      })
    }

    // Delay start a bit for smooth initial load
    const timeoutId = setTimeout(runAnimationCycle, 800)

    return () => {
      isMounted = false
      clearTimeout(timeoutId)
    }
  }, [])

  return (
    <section className="relative pt-40 pb-32 min-h-[100dvh] flex flex-col items-center justify-center overflow-hidden">
      {/* Cyber Grid Background */}
      <div className="absolute inset-0 cyber-grid opacity-30 pointer-events-none" />
      <div className="absolute inset-0 cyber-grid-dots opacity-20 pointer-events-none" />

      {/* Futuristic Aura Ambient Lights */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[55vw] h-[55vw] rounded-full bg-brand-cyan/10 aura-glow animate-pulse-slow pointer-events-none" />
      <div className="absolute bottom-1/4 left-1/2 -translate-x-1/2 translate-y-1/3 w-[60vw] h-[60vw] rounded-full bg-brand-violet/10 aura-glow animate-pulse-slow pointer-events-none" style={{ animationDelay: "-4s" }} />

      <div className="w-full max-w-5xl px-6 relative z-10 flex flex-col items-center text-center">
        
        {/* Tagline */}
        <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-white/5 border border-white/10 mb-8 backdrop-blur-md">
          <Sparkle size={14} className="text-brand-cyan animate-pulse" />
          <span className="text-xs font-mono font-bold tracking-wider text-brand-cyan uppercase">
            The AI Agent Core Tech Stack
          </span>
        </div>

        {/* Headline */}
        <h1 className="text-5xl sm:text-7xl font-extrabold tracking-tight leading-[1.05] text-white max-w-4xl mx-auto">
          {t.hero.headlinePart1}{" "}
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-cyan via-brand-violet to-brand-pink filter drop-shadow-[0_2px_20px_rgba(6,182,212,0.15)] inline-block mt-2">
            {t.hero.headlineAccent}
          </span>
        </h1>

        {/* Description */}
        <p className="mt-8 text-lg sm:text-xl text-terminal-dim leading-relaxed max-w-3xl mx-auto font-medium">
          {t.hero.description}
        </p>

        {/* Call to Actions */}
        <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-5 w-full sm:w-auto">
          <a
            href="#install"
            className="group relative inline-flex items-center justify-center px-8 py-4 w-full sm:w-60 font-sans text-sm font-bold text-black bg-white rounded-full overflow-hidden hover:scale-[1.02] active:scale-[0.98] transition-all duration-300 shadow-xl shadow-white/5 cursor-pointer"
          >
            <span className="absolute inset-0 bg-gradient-to-r from-brand-cyan to-brand-violet opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
            <span className="relative z-10 flex items-center justify-center gap-2 group-hover:text-white transition-colors duration-300">
              {t.hero.getStarted}
              <ArrowRight size={16} className="group-hover:translate-x-1 transition-transform duration-300" />
            </span>
          </a>
          
          <a
            href="https://github.com/0xfig-labs/tide"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center px-8 py-4 w-full sm:w-60 font-sans text-sm font-bold text-white/95 rounded-full border border-white/10 hover:border-white/20 bg-white/5 hover:bg-white/10 active:scale-[0.98] transition-all duration-300 backdrop-blur-md cursor-pointer"
          >
            {t.hero.viewOnGitHub}
          </a>
        </div>

        {/* Centered Interactive Sandbox (Underneath) */}
        <div className="mt-20 w-full max-w-3xl mx-auto relative">
          {/* Outer Glowing Border */}
          <div className="absolute -inset-1 rounded-2xl bg-gradient-to-r from-brand-cyan/20 via-brand-violet/20 to-brand-pink/20 blur-xl opacity-75" />
          
          {/* Terminal Container */}
          <div className="relative rounded-2xl glass-panel overflow-hidden border border-white/10 shadow-2xl bg-slate-950/80">
            
            {/* Terminal Header */}
            <div className="flex items-center justify-between px-5 py-4 border-b border-white/5 bg-slate-900/50">
              <div className="flex items-center gap-2">
                <span className="w-3 h-3 rounded-full bg-terminal-red/60" />
                <span className="w-3 h-3 rounded-full bg-terminal-yellow/60" />
                <span className="w-3 h-3 rounded-full bg-terminal-green/60" />
              </div>
              <span className="font-mono text-xs text-white/40 flex items-center gap-1.5">
                <Terminal size={14} /> agent@tide-cli: ~
              </span>
              <div className="w-12 h-1" /> {/* Spacer */}
            </div>

            {/* Terminal Body with rigid fixed heights on mobile & desktop to prevent screen flickering */}
            <div className="p-6 sm:p-8 font-mono text-sm leading-relaxed h-[430px] sm:h-[460px] flex flex-col justify-between text-left overflow-hidden">
              <div>
                {/* Shell Prompt */}
                <div className="flex items-center gap-2.5 text-white/90">
                  <span className="text-brand-cyan select-none font-bold">$</span>
                  <span>
                    <span ref={textRef} className="text-terminal-fg font-semibold" />
                    <span
                      ref={cursorRef}
                      className="inline-block w-2 h-4 bg-brand-cyan ml-0.5 align-middle shadow-[0_0_8px_var(--color-brand-cyan)]"
                    />
                  </span>
                </div>
              </div>

              {/* Display Canvas Area (Absolute container stack ensures 0% DOM layout shift!) */}
              <div className="relative flex-1 w-full mt-5 min-h-[280px] overflow-hidden">
                {/* Syntax Highlighted JSON Output */}
                {showJson && (
                  <pre
                    ref={jsonRef}
                    className={`absolute inset-x-0 top-0 text-xs sm:text-sm overflow-hidden rounded bg-black/40 border border-white/5 p-5 text-white/80 ${
                      glowActive
                        ? "shadow-[0_0_30px_rgba(16,185,129,0.15)] border-terminal-green/30 scale-[1.01] transition-all duration-300"
                        : ""
                    }`}
                  >
                    <code>
                      <span className="text-white/40">{"["}</span>{"\n"}
                      {"  "}<span className="text-white/40">{"{"}</span>{"\n"}
                      {"    "}<span className="text-brand-pink">"id"</span>: <span className="text-brand-cyan">1</span>,{"\n"}
                      {"    "}<span className="text-brand-pink">"title"</span>: <span className="text-terminal-green">"Building AI Agents with Go & SQLite"</span>,{"\n"}
                      {"    "}<span className="text-brand-pink">"feed"</span>: <span className="text-brand-cyan">"Go Blog"</span>{"\n"}
                      {"  "}<span className="text-white/40">{"}"}</span>,{"\n"}
                      {"  "}<span className="text-white/40">{"{"}</span>{"\n"}
                      {"    "}<span className="text-brand-pink">"id"</span>: <span className="text-brand-cyan">2</span>,{"\n"}
                      {"    "}<span className="text-brand-pink">"title"</span>: <span className="text-terminal-green">"Why JSON-Native Output Rules CLI Tools"</span>{"\n"}
                      {"  "}<span className="text-white/40">{"}"}</span>{"\n"}
                      <span className="text-white/40">{"]"}</span>
                    </code>
                  </pre>
                )}

                {/* Transform Glow Line */}
                {glowActive && (
                  <div className="absolute inset-x-0 top-36 h-0.5 bg-gradient-to-r from-transparent via-terminal-green to-transparent w-full blur-[1px] animate-pulse z-10" />
                )}

                {/* Render Visual RSS Cards */}
                {showCards && (
                  <div ref={cardsContainerRef} className="absolute inset-x-0 top-0 flex flex-col gap-3 w-full">
                    {MOCK_ARTICLES.map((article) => (
                      <div
                        key={article.id}
                        className="rss-demo-card group/card relative rounded-xl border border-white/5 bg-slate-900/60 p-3.5 hover:border-brand-cyan/30 hover:bg-slate-900/90 transition-all duration-300 cursor-pointer"
                      >
                        <div className="flex items-start justify-between gap-4">
                          <div className="flex items-center gap-2">
                            <span className="text-[10px] font-bold uppercase tracking-wider text-brand-cyan px-2 py-0.5 rounded bg-brand-cyan/10 border border-brand-cyan/15">
                              {article.category}
                            </span>
                            <span className="text-[11px] text-white/40 flex items-center gap-1">
                              <BookOpen size={12} /> {article.feed}
                            </span>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="text-[10px] text-white/40">{article.published}</span>
                            <span className="h-1.5 w-1.5 rounded-full bg-terminal-green animate-pulse" />
                          </div>
                        </div>

                        <h3 className="mt-2 text-xs font-semibold text-white/95 group-hover/card:text-brand-cyan transition-colors duration-300">
                          {article.title}
                        </h3>
                        <p className="mt-1 text-[11px] text-terminal-dim leading-relaxed line-clamp-1">
                          {article.summary}
                        </p>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Action Bar showing 'structured JSON to visual cards' */}
              <div className="mt-auto pt-4 border-t border-white/5 flex items-center justify-between text-[11px] text-white/30">
                <span className="flex items-center gap-1.5">
                  <Brain size={14} className="text-brand-violet" />
                  Parsed by Autonomous Agent
                </span>
                <span className="flex items-center gap-1.5">
                  <Rss size={14} className="text-brand-pink" />
                  JSON to Visual AST
                </span>
              </div>

            </div>
          </div>
        </div>

      </div>
    </section>
  )
}
