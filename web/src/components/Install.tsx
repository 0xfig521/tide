import { useState, useRef } from "react"
import { Copy, Terminal, Check } from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

const installMethods = [
  {
    id: "curl",
    label: "curl",
    command: "curl -fsSL https://raw.githubusercontent.com/0xfig-labs/tide/main/install.sh | bash",
  },
  {
    id: "brew",
    label: "Homebrew",
    command: "brew install 0xfig-labs/tap/tide",
  },
  {
    id: "go",
    label: "Go",
    command: "go install github.com/0xfig-labs/tide@latest",
  },
]

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

export function Install() {
  const { t } = useLocale()
  const [activeTab, setActiveTab] = useState(0)
  const [copied, setCopied] = useState(false)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)

  const activeMethod = installMethods[activeTab]

  const triggerParticleExplosion = () => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    const rect = canvas.getBoundingClientRect()
    canvas.width = rect.width
    canvas.height = rect.height

    const particles: Particle[] = []
    const colors = ["#10b981", "#06b6d4", "#8b5cf6", "#34d399"]

    const originX = canvas.width / 2
    const originY = canvas.height / 2

    for (let i = 0; i < 35; i++) {
      const angle = Math.random() * Math.PI * 2
      const speed = Math.random() * 4 + 2
      const maxLife = Math.random() * 20 + 20
      particles.push({
        x: originX,
        y: originY,
        vx: Math.cos(angle) * speed,
        vy: Math.sin(angle) * speed - 1.5,
        size: Math.random() * 3 + 1.5,
        color: colors[Math.floor(Math.random() * colors.length)],
        alpha: 1,
        life: maxLife,
        maxLife: maxLife,
      })
    }

    const animate = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      
      let allDead = true
      particles.forEach((p) => {
        if (p.life > 0) {
          allDead = false
          p.x += p.vx
          p.y += p.vy
          p.vy += 0.08
          p.life--

          p.alpha = p.life / p.maxLife

          ctx.save()
          ctx.globalAlpha = p.alpha
          ctx.beginPath()
          ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2)
          ctx.fillStyle = p.color
          ctx.shadowBlur = 6
          ctx.shadowColor = p.color
          ctx.fill()
          ctx.restore()
        }
      })

      if (!allDead) {
        requestAnimationFrame(animate)
      } else {
        ctx.clearRect(0, 0, canvas.width, canvas.height)
      }
    }

    animate()
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(activeMethod.command)
    setCopied(true)
    triggerParticleExplosion()
    setTimeout(() => setCopied(false), 2200)
  }

  return (
    <section id="install" className="py-32 relative overflow-hidden border-t border-white/5 flex justify-center w-full">
      {/* Background elements */}
      <div className="absolute inset-0 cyber-grid-dots opacity-20 pointer-events-none" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[35rem] h-[35rem] rounded-full bg-brand-cyan/5 blur-[120px] pointer-events-none" />

      <div className="w-full max-w-5xl px-6 relative z-10 flex flex-col items-center">
        
        {/* Title */}
        <div className="text-center mb-16 flex flex-col items-center">
          <h2 className="text-4xl sm:text-5xl font-extrabold tracking-tight text-white">
            {t.install.headline}
          </h2>
          <div className="h-1 w-16 bg-gradient-to-r from-brand-violet to-brand-pink mt-5 rounded-full" />
        </div>

        {/* Outer Command Center (Perfect Centered max-w-3xl) */}
        <div className="relative rounded-2xl border border-white/5 bg-slate-950/80 shadow-2xl p-6 md:p-8 w-full max-w-3xl">
          {/* Subtle Border beam effect */}
          <div className="absolute -inset-px rounded-2xl bg-gradient-to-r from-brand-cyan/20 to-brand-violet/20 opacity-40 blur-[1px] pointer-events-none" />

          {/* Selector Tabs (Centered) */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 border-b border-white/5 pb-5 mb-6">
            <div className="flex items-center gap-1.5 text-white/40">
              <Terminal size={18} />
              <span className="font-mono text-xs font-semibold uppercase tracking-wider">Install via</span>
            </div>
            
            <div className="flex items-center gap-1.5 bg-white/5 rounded-full p-0.5 border border-white/10">
              {installMethods.map((method, idx) => (
                <button
                  key={method.id}
                  onClick={() => {
                    setActiveTab(idx)
                    setCopied(false)
                  }}
                  className={`px-4 py-1.5 rounded-full font-mono text-xs font-semibold transition-all duration-300 ${
                    activeTab === idx
                      ? "bg-gradient-to-r from-brand-cyan/90 to-brand-violet/90 text-white shadow-md shadow-brand-cyan/10 scale-105"
                      : "text-terminal-dim hover:text-terminal-fg"
                  }`}
                >
                  {method.label}
                </button>
              ))}
            </div>
          </div>

          {/* Installation Terminal Box */}
          <div className="relative rounded-xl border border-white/5 bg-slate-900/60 p-6 font-mono text-sm sm:text-base leading-relaxed overflow-hidden">
            {/* Background copy particle canvas */}
            <canvas
              ref={canvasRef}
              className="absolute inset-0 pointer-events-none z-10"
              style={{ width: "100%", height: "100%" }}
            />

            <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-6 relative z-20">
              {/* Command text */}
              <div className="flex items-start gap-3 flex-1 break-all select-all text-left">
                <span className="text-brand-cyan font-bold select-none">$</span>
                <span className="text-white/90 font-medium tracking-tight">
                  {activeMethod.command}
                </span>
              </div>

              {/* Advanced Copy Button */}
              <button
                ref={buttonRef}
                onClick={handleCopy}
                className={`relative overflow-hidden group flex items-center justify-center gap-2 py-2.5 px-5 rounded-lg border font-sans text-xs font-bold transition-all duration-300 shrink-0 w-full md:w-auto ${
                  copied
                    ? "bg-terminal-green border-terminal-green/30 text-white shadow-[0_0_15px_rgba(16,185,129,0.3)]"
                    : "bg-white/5 hover:bg-white/10 text-white/90 hover:text-white border-white/10 hover:border-white/20 active:scale-95"
                }`}
              >
                {copied ? (
                  <>
                    <Check size={14} className="animate-scale-in" />
                    <span>{t.install.copied}</span>
                  </>
                ) : (
                  <>
                    <Copy size={14} className="group-hover:rotate-6 transition-transform" />
                    <span>{t.install.copy}</span>
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Validation Info footer */}
          <div className="mt-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs font-mono text-white/30 border-t border-white/5 pt-4">
            <span>✓ Verified stable release</span>
            <span className="flex items-center gap-1.5">
              <span className="h-1.5 w-1.5 rounded-full bg-terminal-green animate-pulse" />
              Run <span className="text-brand-cyan">tide --version</span> after install
            </span>
          </div>

        </div>

      </div>
    </section>
  )
}
