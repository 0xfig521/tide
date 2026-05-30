import { Copy, Check, Cpu, Brain, Sparkle, BookOpen } from "@phosphor-icons/react"
import { useState, useRef, useEffect } from "react"
import { useLocale } from "../i18n/context"

const SKILL_COMMAND = "npx skills add 0xfig521/tide"

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  size: number
  color: string
  alpha: number
  life: number
  maxLife: number
}

export function AISkill() {
  const { t } = useLocale()
  const [copied, setCopied] = useState(false)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [stage, setStage] = useState(0) // 0: add, 1: fetch, 2: search

  // Auto cycle through simulated Agent workflow steps
  useEffect(() => {
    const timer = setInterval(() => {
      setStage((prev) => (prev + 1) % 3)
    }, 4500)
    return () => clearInterval(timer)
  }, [])

  const triggerParticles = () => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext("2d")
    if (!ctx) return

    const rect = canvas.getBoundingClientRect()
    canvas.width = rect.width
    canvas.height = rect.height

    const particles: Particle[] = []
    const originX = canvas.width / 2
    const originY = canvas.height / 2

    for (let i = 0; i < 30; i++) {
      const angle = Math.random() * Math.PI * 2
      const speed = Math.random() * 3 + 1.5
      const maxLife = Math.random() * 15 + 15
      particles.push({
        x: originX,
        y: originY,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed,
        size: Math.random() * 2 + 1,
        color: "#a78bfa", // Purple theme for skills.sh
        alpha: 1,
        life: maxLife,
        maxLife: maxLife,
      })
    }

    const frame = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      let alive = false
      particles.forEach((p) => {
        if (p.life > 0) {
          alive = true
          p.x += p.vx
          p.y += p.vy
          p.life--
          p.alpha = p.life / p.maxLife

          ctx.save()
          ctx.globalAlpha = p.alpha
          ctx.beginPath()
          ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2)
          ctx.fillStyle = p.color
          ctx.shadowBlur = 4
          ctx.shadowColor = p.color
          ctx.fill()
          ctx.restore()
        }
      })
      if (alive) requestAnimationFrame(frame)
    }
    frame()
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(SKILL_COMMAND)
    setCopied(true)
    triggerParticles()
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <section className="py-32 border-t border-white/5 bg-slate-950/20 relative overflow-hidden">
      {/* Decorative Aura */}
      <div className="absolute top-1/2 left-2/3 -translate-y-1/2 w-[35vw] h-[35vw] rounded-full bg-brand-violet/5 blur-[120px] pointer-events-none" />
      <div className="absolute inset-0 cyber-grid-dots opacity-10 pointer-events-none" />

      <div className="max-w-6xl mx-auto px-6 relative z-10">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-16 items-center">
          
          {/* Left Description Column */}
          <div className="lg:col-span-6">
            {/* Tagline */}
            <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-brand-violet/5 border border-brand-violet/10 mb-6">
              <Brain size={14} className="text-brand-violet" />
              <span className="text-xs font-mono font-semibold tracking-wider text-brand-violet uppercase">
                Zero Parsing RSS for LLMs
              </span>
            </div>

            {/* Headline */}
            <h2 className="text-4xl sm:text-5xl font-extrabold tracking-tight text-white leading-tight">
              {t.aiSkill.headline1}
              <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-brand-violet to-brand-pink filter drop-shadow-[0_2px_15px_rgba(139,92,246,0.1)]">
                {t.aiSkill.headline2}
              </span>
            </h2>

            {/* Descriptions */}
            <p className="mt-6 text-base text-terminal-dim leading-relaxed font-medium">
              {t.aiSkill.description1}
            </p>
            <p className="mt-4 text-base text-terminal-dim leading-relaxed font-medium">
              {t.aiSkill.description2}
            </p>

            {/* skills.sh Installer Box */}
            <div className="mt-8 relative rounded-xl border border-white/5 bg-slate-950/80 p-5 overflow-hidden">
              <canvas ref={canvasRef} className="absolute inset-0 pointer-events-none" />
              <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-brand-violet/10 text-brand-violet border border-brand-violet/20 select-none">
                    skills.sh
                  </span>
                  <span className="font-mono text-sm text-white/95 break-all select-all font-medium">
                    {SKILL_COMMAND}
                  </span>
                </div>

                <button
                  onClick={handleCopy}
                  className={`flex items-center justify-center gap-1.5 py-1.5 px-3.5 rounded-lg border font-sans text-xs font-bold transition-all duration-300 w-full sm:w-auto shrink-0 ${
                    copied
                      ? "bg-brand-violet border-brand-violet/30 text-white shadow-lg shadow-brand-violet/20"
                      : "bg-white/5 hover:bg-white/10 text-white/80 border-white/10 active:scale-95"
                  }`}
                >
                  {copied ? <Check size={12} /> : <Copy size={12} />}
                  <span>{copied ? t.install.copied : t.aiSkill.copyCommand}</span>
                </button>
              </div>
            </div>
          </div>

          {/* Right Agent Console Simulator */}
          <div className="lg:col-span-6 w-full">
            <div className="relative rounded-2xl glass-panel overflow-hidden border border-white/10 shadow-2xl bg-slate-950/70">
              
              {/* Simulator Header */}
              <div className="flex items-center justify-between px-5 py-3.5 border-b border-white/5 bg-slate-900/50">
                <div className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded-md bg-brand-violet/15 flex items-center justify-center text-brand-violet border border-brand-violet/20">
                    <Cpu size={12} weight="bold" />
                  </div>
                  <div>
                    <div className="font-mono text-xs font-semibold text-white/90">Claude 3.5 Sonnet</div>
                    <div className="font-mono text-[10px] text-white/40">{t.aiSkill.cardSubtitle}</div>
                  </div>
                </div>
                
                {/* Step indicators */}
                <div className="flex gap-1.5 font-mono text-[9px]">
                  {[0, 1, 2].map((s) => (
                    <span
                      key={s}
                      onClick={() => setStage(s)}
                      className={`px-2 py-0.5 rounded cursor-pointer transition-all ${
                        stage === s
                          ? "bg-brand-violet text-white font-bold"
                          : "bg-white/5 text-white/30 hover:bg-white/10"
                      }`}
                    >
                      Step {s + 1}
                    </span>
                  ))}
                </div>
              </div>

              {/* Simulator Body */}
              <div className="p-6 font-mono text-xs leading-relaxed min-h-[300px] flex flex-col justify-between">
                
                {/* Step 1: Subscribing */}
                {stage === 0 && (
                  <div className="space-y-4 animate-fade-in">
                    <div className="text-white/40"># AI resolves query "Subscribe to Golang updates"</div>
                    <div className="text-white/90">
                      <span className="text-brand-violet font-bold">$</span>{" "}
                      tide add "https://blog.golang.org/feed.atom" --category "tech"
                    </div>
                    <div className="text-terminal-green font-semibold flex items-center gap-1.5">
                      <span>✓ Subscribed successfully</span>
                    </div>
                    <div className="text-white/40 bg-white/5 rounded border border-white/5 p-3 text-[11px] leading-normal">
                      Feed "Go Blog" mapped inside SQLite feed tables under context "tech". Ready for next ingest cycle.
                    </div>
                  </div>
                )}

                {/* Step 2: Ingest Fetching */}
                {stage === 1 && (
                  <div className="space-y-4 animate-fade-in">
                    <div className="text-white/40"># AI invokes concurrent updates pull</div>
                    <div className="text-white/90">
                      <span className="text-brand-violet font-bold">$</span> tide fetch --concurrency 10
                    </div>
                    
                    {/* Simulated Concurrent Downloading animation */}
                    <div className="space-y-1.5 py-1 text-white/50 text-[10px] pl-3 border-l-2 border-brand-violet/20">
                      <div>Fetching Go Blog... Done</div>
                      <div>Fetching DevTo... Done</div>
                      <div>Fetching HackerNews... In Progress [■■■■■■□□□□] 60%</div>
                    </div>

                    <div className="text-terminal-green font-semibold">
                      ✓ Fetched 42 articles from 12 feeds in 1.42s
                    </div>
                  </div>
                )}

                {/* Step 3: Search Structured Stream */}
                {stage === 2 && (
                  <div className="space-y-4 animate-fade-in">
                    <div className="text-white/40"># AI searches with date filter</div>
                    <div className="text-white/90">
                      <span className="text-brand-violet font-bold">$</span> tide search "kubernetes" --since 7d
                    </div>
                    <pre className="text-[10px] text-white/70 bg-black/40 border border-white/5 rounded p-3 leading-normal max-h-[120px] overflow-hidden select-none">
                      <code>
                        {"["}{"\n"}
                        {"  "}{"{"}{"\n"}
                        {"    "}<span className="text-brand-pink">"title"</span>: <span className="text-terminal-green">"K8s 1.30 released"</span>,{"\n"}
                        {"    "}<span className="text-brand-pink">"feed"</span>: <span className="text-brand-cyan">"Go Blog"</span>,{"\n"}
                        {"    "}<span className="text-brand-pink">"published"</span>: <span className="text-brand-violet">"2025-08-15T10:00:00Z"</span>{"\n"}
                        {"  "}{"}"}{"\n"}
                        {"]"}
                      </code>
                    </pre>
                    
                    {/* Agent reasoning bubble */}
                    <div className="flex gap-2 p-3 bg-brand-violet/5 border border-brand-violet/10 rounded-lg text-white/90 leading-normal">
                      <Sparkle size={18} className="text-brand-violet shrink-0 mt-0.5" />
                      <div>
                        <span className="font-semibold text-brand-violet">Agent Reasoning: </span>
                        I have filtered 1 article on Kubernetes from the RSS database. Summarizing: Kubernetes 1.30 was released on August 15th.
                      </div>
                    </div>
                  </div>
                )}

                {/* Footer status indicating skills.sh */}
                <div className="mt-6 pt-4 border-t border-white/5 flex items-center justify-between text-[11px] text-white/30">
                  <span className="flex items-center gap-1">
                    <BookOpen size={13} />
                    Auto-orchestrated by LLM
                  </span>
                  <span>Agent Stage Active</span>
                </div>

              </div>
            </div>
          </div>

        </div>
      </div>
    </section>
  )
}
