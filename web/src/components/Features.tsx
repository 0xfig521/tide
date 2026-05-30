import { useEffect, useRef, useState } from "react"
import gsap from "gsap"
import { ScrollTrigger } from "gsap/ScrollTrigger"
import {
  Robot,
  BracketsCurly,
  Lightning,
  Database,
  ShieldCheck,
  Clock,
  ArrowRight,
  Cpu,
  ArrowsLeftRight,
} from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

gsap.registerPlugin(ScrollTrigger)

export function Features() {
  const { t } = useLocale()
  const sectionRef = useRef<HTMLDivElement>(null)
  const gridRef = useRef<HTMLDivElement>(null)

  // Sub-states for various interactive bento widgets
  const [jsonExpanded, setJsonExpanded] = useState<Record<string, boolean>>({
    items: true,
    meta: false,
  })
  const [fetchProgress, setFetchProgress] = useState<number[]>([0, 0, 0, 0])
  const [fetchingActive, setFetchingActive] = useState(true)

  // 1. Concurrent Fetching progress bar loop
  useEffect(() => {
    if (!fetchingActive) return
    const interval = setInterval(() => {
      setFetchProgress((prev) => {
        const next = prev.map((p) => {
          if (p >= 100) return 0
          return p + Math.floor(Math.random() * 15) + 5
        })
        return next.map((n) => (n > 100 ? 100 : n))
      })
    }, 180)

    return () => clearInterval(interval)
  }, [fetchingActive])

  // 2. GSAP Scroll Trigger for Cards Entrance
  useEffect(() => {
    const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    if (prefersReducedMotion || !gridRef.current) return

    const ctx = gsap.context(() => {
      const cards = gridRef.current!.querySelectorAll(".bento-card")

      gsap.set(cards, { opacity: 0, y: 35 })

      ScrollTrigger.batch(cards, {
        onEnter: (elements) => {
          gsap.to(elements, {
            opacity: 1,
            y: 0,
            stagger: 0.12,
            duration: 0.65,
            ease: "power2.out",
            overwrite: true,
          })
        },
        start: "top 88%",
        once: true,
      })
    }, sectionRef)

    return () => ctx.revert()
  }, [])

  // Shared 3D tilt event handlers
  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const card = e.currentTarget
    const rect = card.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top
    
    const xc = rect.width / 2
    const yc = rect.height / 2
    
    const rotateX = ((yc - y) / yc) * 8
    const rotateY = ((x - xc) / xc) * 8

    card.style.transform = `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg) scale3d(1.01, 1.01, 1.01)`
  }

  const handleMouseLeave = (e: React.MouseEvent<HTMLDivElement>) => {
    const card = e.currentTarget
    card.style.transform = `perspective(1000px) rotateX(0deg) rotateY(0deg) scale3d(1, 1, 1)`
  }

  return (
    <section ref={sectionRef} className="py-32 relative overflow-hidden border-t border-white/5 flex justify-center w-full">
      {/* Background Decorative Lines */}
      <div className="absolute inset-0 cyber-grid opacity-10 pointer-events-none" />

      <div className="w-full max-w-5xl px-6 relative z-10">
        
        {/* Header (Fully Centered) */}
        <div className="text-center mb-20 flex flex-col items-center">
          <h2 className="text-4xl sm:text-5xl font-extrabold tracking-tight text-white leading-tight">
            {t.features.headline}
          </h2>
          <div className="h-1 w-16 bg-gradient-to-r from-brand-cyan to-brand-violet mt-5 rounded-full" />
        </div>

        {/* Bento Grid */}
        <div
          ref={gridRef}
          className="grid grid-cols-1 lg:grid-cols-3 gap-6 w-full"
        >
          {/* Card 0: AI-Native Skill (col-span-2 on large screens, stacks vertically on smaller screens) */}
          <div
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            className="bento-card lg:col-span-2 group relative rounded-2xl border border-white/5 bg-slate-950/60 p-6 overflow-hidden hover:border-brand-cyan/20 transition-all duration-300"
            style={{ transition: "transform 0.1s ease-out, border-color 0.3s ease" }}
          >
            <div className="absolute -inset-px bg-gradient-to-r from-brand-cyan/20 to-brand-violet/20 opacity-0 group-hover:opacity-100 rounded-2xl blur-[1px] transition-opacity duration-500 pointer-events-none" />

            <div className="relative z-10 flex flex-col lg:flex-row gap-8 justify-between h-full w-full">
              {/* Left text description */}
              <div className="flex flex-col justify-between w-full lg:max-w-[50%] shrink-0">
                <div>
                  <div className="w-10 h-10 rounded-xl bg-brand-cyan/10 border border-brand-cyan/15 flex items-center justify-center text-brand-cyan mb-5 shadow-lg shadow-brand-cyan/5">
                    <Robot size={22} weight="duotone" />
                  </div>
                  <h3 className="text-xl font-bold text-white mb-3">
                    {t.features.items[0].title}
                  </h3>
                  <p className="text-terminal-dim text-sm leading-relaxed mb-6 font-medium">
                    {t.features.items[0].description}
                  </p>
                </div>
                <div className="inline-flex items-center gap-1.5 text-xs text-brand-cyan font-mono font-semibold mt-4 lg:mt-0">
                  <span>skills.sh protocol enabled</span>
                  <ArrowRight size={14} />
                </div>
              </div>

              {/* Chat Simulation Widget (Guaranteed no compression!) */}
              <div className="w-full lg:max-w-[45%] rounded-xl border border-white/5 bg-slate-900/60 p-4 font-mono text-[11px] leading-relaxed shadow-inner shrink-0 flex flex-col justify-between min-h-[160px]">
                <div>
                  <div className="flex items-center gap-1.5 mb-3 border-b border-white/5 pb-2 text-white/40">
                    <span className="w-2 h-2 rounded-full bg-brand-cyan animate-pulse" />
                    <span>Agent Chat Box</span>
                  </div>
                  <div className="space-y-4">
                    <div className="text-white/80">
                      <span className="text-brand-pink font-semibold">User: </span>
                      <span>"Any news about Go this morning?"</span>
                    </div>
                    <div className="text-terminal-dim p-2.5 rounded bg-black/30 border border-white/5 text-[10px]">
                      <span className="text-brand-violet font-semibold">Agent: </span>
                      <span className="text-brand-cyan">tide list --unread --category tech</span>
                      <div className="mt-1.5 text-white/40 text-[9px] border-t border-white/5 pt-1.5">
                        Calling Tide binary via skills.sh definition...
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Card 1: JSON by Default (col-span-1) */}
          <div
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            className="bento-card lg:col-span-1 group relative rounded-2xl border border-white/5 bg-slate-950/60 p-6 overflow-hidden hover:border-brand-violet/20 transition-all duration-300"
            style={{ transition: "transform 0.1s ease-out, border-color 0.3s ease" }}
          >
            <div className="absolute -inset-px bg-gradient-to-r from-brand-violet/20 to-brand-pink/20 opacity-0 group-hover:opacity-100 rounded-2xl blur-[1px] transition-opacity duration-500 pointer-events-none" />

            <div className="relative z-10 flex flex-col justify-between h-full">
              <div>
                <div className="w-10 h-10 rounded-xl bg-brand-violet/10 border border-brand-violet/15 flex items-center justify-center text-brand-violet mb-5 shadow-lg shadow-brand-violet/5">
                  <BracketsCurly size={22} weight="duotone" />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">
                  {t.features.items[1].title}
                </h3>
                <p className="text-terminal-dim text-sm leading-relaxed mb-6 font-medium">
                  {t.features.items[1].description}
                </p>
              </div>

              {/* Interactive JSON Explorer Widget */}
              <div className="rounded-xl border border-white/5 bg-slate-900/60 p-3.5 font-mono text-[10px] select-none mt-4 lg:mt-0">
                <div className="flex items-center justify-between text-white/30 border-b border-white/5 pb-1.5 mb-2">
                  <span>output.json</span>
                  <span className="text-[9px] text-brand-violet">Click node</span>
                </div>
                <div className="space-y-1">
                  <div className="text-white/50">{"{"}</div>
                  
                  {/* Items Node */}
                  <div className="pl-3">
                    <span
                      onClick={() => setJsonExpanded(prev => ({ ...prev, items: !prev.items }))}
                      className="cursor-pointer text-brand-pink hover:text-brand-pink/80 transition-colors"
                    >
                      "articles"
                    </span>
                    <span className="text-white/40">: </span>
                    <span className="text-white/40">{jsonExpanded.items ? "[" : "[...]"}</span>
                    {jsonExpanded.items && (
                      <div className="pl-4 text-white/70">
                        <div><span className="text-brand-cyan">"count"</span>: <span className="text-brand-violet">12</span></div>
                      </div>
                    )}
                    {jsonExpanded.items && <div className="text-white/40">]</div>}
                  </div>

                  {/* Meta Node */}
                  <div className="pl-3">
                    <span
                      onClick={() => setJsonExpanded(prev => ({ ...prev, meta: !prev.meta }))}
                      className="cursor-pointer text-brand-pink hover:text-brand-pink/80 transition-colors"
                    >
                      "meta"
                    </span>
                    <span className="text-white/40">: </span>
                    <span className="text-white/40">{jsonExpanded.meta ? "{" : "{...}"}</span>
                    {jsonExpanded.meta && (
                      <div className="pl-4 text-white/60">
                        <div><span className="text-brand-cyan">"engine"</span>: <span className="text-brand-violet">"tide v0.1"</span></div>
                      </div>
                    )}
                    {jsonExpanded.meta && <div className="text-white/40">{"}"}</div>}
                  </div>

                  <div className="text-white/50">{"}"}</div>
                </div>
              </div>
            </div>
          </div>

          {/* Card 2: Concurrent Fetching (col-span-1) */}
          <div
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            className="bento-card lg:col-span-1 group relative rounded-2xl border border-white/5 bg-slate-950/60 p-6 overflow-hidden hover:border-brand-cyan/20 transition-all duration-300"
            style={{ transition: "transform 0.1s ease-out, border-color 0.3s ease" }}
          >
            <div className="absolute -inset-px bg-gradient-to-r from-brand-cyan/20 to-brand-violet/20 opacity-0 group-hover:opacity-100 rounded-2xl blur-[1px] transition-opacity duration-500 pointer-events-none" />

            <div className="relative z-10 flex flex-col justify-between h-full">
              <div>
                <div className="w-10 h-10 rounded-xl bg-brand-cyan/10 border border-brand-cyan/15 flex items-center justify-center text-brand-cyan mb-5 shadow-lg shadow-brand-cyan/5">
                  <Lightning size={22} weight="duotone" />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">
                  {t.features.items[2].title}
                </h3>
                <p className="text-terminal-dim text-sm leading-relaxed mb-6 font-medium">
                  {t.features.items[2].description}
                </p>
              </div>

              {/* Progress Bar Widget */}
              <div
                onClick={() => setFetchingActive(!fetchingActive)}
                className="rounded-xl border border-white/5 bg-slate-900/60 p-4 font-mono text-[10px] space-y-3 cursor-pointer hover:bg-slate-900/80 transition-colors mt-4 lg:mt-0"
              >
                <div className="flex justify-between items-center text-white/30 border-b border-white/5 pb-1">
                  <span>concurrency = 10</span>
                  <span className={fetchingActive ? "text-brand-cyan animate-pulse" : "text-white/40"}>
                    {fetchingActive ? "FETCHING..." : "PAUSED"}
                  </span>
                </div>
                {["Go Blog", "HackerNews", "Reddit/r/go", "Medium"].map((source, i) => (
                  <div key={source} className="space-y-1">
                    <div className="flex justify-between text-white/70">
                      <span>{source}</span>
                      <span>{fetchProgress[i]}%</span>
                    </div>
                    <div className="h-1.5 rounded-full bg-white/5 overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-brand-cyan to-brand-violet transition-all duration-200"
                        style={{ width: `${fetchProgress[i]}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Card 3: Zero Dependencies (col-span-1) */}
          <div
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            className="bento-card lg:col-span-1 group relative rounded-2xl border border-white/5 bg-slate-950/60 p-6 overflow-hidden hover:border-brand-pink/20 transition-all duration-300"
            style={{ transition: "transform 0.1s ease-out, border-color 0.3s ease" }}
          >
            <div className="absolute -inset-px bg-gradient-to-r from-brand-pink/20 to-brand-violet/20 opacity-0 group-hover:opacity-100 rounded-2xl blur-[1px] transition-opacity duration-500 pointer-events-none" />

            <div className="relative z-10 flex flex-col justify-between h-full">
              <div>
                <div className="w-10 h-10 rounded-xl bg-brand-pink/10 border border-brand-pink/15 flex items-center justify-center text-brand-pink mb-5 shadow-lg shadow-brand-pink/5">
                  <Database size={22} weight="duotone" />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">
                  {t.features.items[3].title}
                </h3>
                <p className="text-terminal-dim text-sm leading-relaxed mb-6 font-medium">
                  {t.features.items[3].description}
                </p>
              </div>

              {/* Zero Dependency Topology Graph */}
              <div className="rounded-xl border border-white/5 bg-slate-900/60 p-4 flex flex-col gap-2.5 font-mono text-[10px] mt-4 lg:mt-0">
                <div className="flex items-center gap-2 p-2.5 rounded border border-brand-pink/20 bg-brand-pink/5 text-white/95">
                  <Cpu size={14} className="text-brand-pink" />
                  <div>
                    <div className="font-semibold text-brand-pink">tide-cli binary</div>
                    <span className="text-[8px] text-white/40">Compiled Go Engine</span>
                  </div>
                  <span className="ml-auto font-semibold text-white/40 text-[9px]">5.4MB</span>
                </div>
                <div className="flex justify-center text-white/20 h-3 items-center">
                  <ArrowsLeftRight size={12} className="rotate-90" />
                </div>
                <div className="flex items-center gap-2 p-2.5 rounded border border-brand-violet/20 bg-brand-violet/5 text-white/95">
                  <Database size={14} className="text-brand-violet" />
                  <div>
                    <div className="font-semibold text-brand-violet">Embedded SQLite</div>
                    <span className="text-[8px] text-white/40">Zero setup local DB</span>
                  </div>
                  <span className="ml-auto px-2 py-0.5 rounded text-[8px] bg-brand-violet/10 text-brand-violet border border-brand-violet/20">Active</span>
                </div>
              </div>
            </div>
          </div>

          {/* Card 4: Smart Caching (col-span-1) */}
          <div
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            className="bento-card lg:col-span-1 group relative rounded-2xl border border-white/5 bg-slate-950/60 p-6 overflow-hidden hover:border-brand-emerald/20 transition-all duration-300"
            style={{ transition: "transform 0.1s ease-out, border-color 0.3s ease" }}
          >
            <div className="absolute -inset-px bg-gradient-to-r from-brand-emerald/20 to-brand-cyan/20 opacity-0 group-hover:opacity-100 rounded-2xl blur-[1px] transition-opacity duration-500 pointer-events-none" />

            <div className="relative z-10 flex flex-col justify-between h-full">
              <div>
                <div className="w-10 h-10 rounded-xl bg-brand-emerald/10 border border-brand-emerald/15 flex items-center justify-center text-brand-emerald mb-5 shadow-lg shadow-brand-emerald/5">
                  <ShieldCheck size={22} weight="duotone" />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">
                  {t.features.items[4].title}
                </h3>
                <p className="text-terminal-dim text-sm leading-relaxed mb-6 font-medium">
                  {t.features.items[4].description}
                </p>
              </div>

              {/* Data Savings Graph */}
              <div className="rounded-xl border border-white/5 bg-slate-900/60 p-4 font-mono text-[10px] mt-4 lg:mt-0">
                <div className="flex justify-between items-center text-white/30 border-b border-white/5 pb-2 mb-3">
                  <span>conditional fetching</span>
                  <span className="text-[9px] text-brand-emerald">ETag / Mod</span>
                </div>
                
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <span className="text-white/40 w-12">Scraping:</span>
                    <div className="flex-1 h-3.5 bg-white/5 rounded border border-white/5 flex items-center px-2">
                      <span className="text-white/60 text-[9px] font-semibold">12.4 MB</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-white/40 w-12">Tide:</span>
                    <div className="flex-1 h-3.5 bg-brand-emerald/10 rounded border border-brand-emerald/20 flex items-center justify-between px-2 relative overflow-hidden">
                      <div className="absolute inset-0 bg-brand-emerald/15 animate-pulse" />
                      <span className="text-brand-emerald text-[9px] font-bold relative z-10">154 KB</span>
                      <span className="text-[8px] font-semibold bg-brand-emerald text-black px-1 rounded relative z-10">-98.7%</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Card 5: Daemon Mode (col-span-3 - Whole bottom row!) */}
          <div
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
            className="bento-card lg:col-span-3 group relative rounded-2xl border border-white/5 bg-slate-950/60 p-6 overflow-hidden hover:border-brand-violet/20 transition-all duration-300"
            style={{ transition: "transform 0.1s ease-out, border-color 0.3s ease" }}
          >
            <div className="absolute -inset-px bg-gradient-to-r from-brand-violet/10 via-brand-pink/10 to-brand-cyan/10 opacity-0 group-hover:opacity-100 rounded-2xl blur-[1px] transition-opacity duration-500 pointer-events-none" />

            <div className="relative z-10 flex flex-col lg:flex-row items-center justify-between gap-8 h-full w-full">
              <div className="max-w-xl text-left">
                <div className="w-10 h-10 rounded-xl bg-brand-violet/10 border border-brand-violet/15 flex items-center justify-center text-brand-violet mb-5 shadow-lg shadow-brand-violet/5">
                  <Clock size={22} weight="duotone" />
                </div>
                <h3 className="text-xl font-bold text-white mb-3">
                  {t.features.items[5].title}
                </h3>
                <p className="text-terminal-dim text-sm leading-relaxed font-medium">
                  {t.features.items[5].description}
                </p>
              </div>

              {/* Radar Sweeping Clock Daemon Widget */}
              <div className="relative w-44 h-44 rounded-full border border-white/5 bg-slate-900/40 flex items-center justify-center shadow-lg shadow-black/40 shrink-0 mt-6 lg:mt-0">
                {/* Radar Sweep Effect */}
                <div className="absolute inset-0 rounded-full bg-gradient-to-tr from-brand-violet/0 via-brand-violet/10 to-brand-violet/30 animate-spin-slow pointer-events-none" />
                <div className="absolute w-[90%] h-[90%] rounded-full border border-dashed border-white/5 flex items-center justify-center">
                  <div className="absolute w-[70%] h-[70%] rounded-full border border-dashed border-brand-violet/10 flex items-center justify-center">
                    <div className="w-2 h-2 rounded-full bg-brand-violet shadow-[0_0_8px_var(--color-brand-violet)] animate-pulse" />
                  </div>
                </div>

                {/* Clock indicator */}
                <div className="absolute top-1/2 left-1/2 w-1.5 h-12 bg-gradient-to-t from-transparent to-brand-violet/60 rounded-full origin-bottom -translate-x-1/2 -translate-y-full rotate-45" />
                <div className="absolute top-1/2 left-1/2 w-1.5 h-8 bg-gradient-to-t from-transparent to-brand-cyan/60 rounded-full origin-bottom -translate-x-1/2 -translate-y-full -rotate-12" />
                
                <span className="absolute bottom-4 font-mono text-[9px] text-white/30 tracking-widest uppercase">
                  Active in BG
                </span>
              </div>
            </div>
          </div>

        </div>
      </div>
    </section>
  )
}
