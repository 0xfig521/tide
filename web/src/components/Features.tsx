import { useEffect, useRef } from "react"
import gsap from "gsap"
import { ScrollTrigger } from "gsap/ScrollTrigger"
import {
  Robot,
  BracketsCurly,
  Lightning,
  Database,
  Clock,
  ClockAfternoon,
} from "@phosphor-icons/react"
import { useLocale } from "../i18n/context"

gsap.registerPlugin(ScrollTrigger)

const ICONS = [Robot, BracketsCurly, Lightning, Database, Clock, ClockAfternoon]

export function Features() {
  const { t } = useLocale()
  const sectionRef = useRef<HTMLDivElement>(null)
  const cardsRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    if (prefersReducedMotion || !cardsRef.current) return

    const ctx = gsap.context(() => {
      const cards = cardsRef.current!.querySelectorAll(".feature-card")

      gsap.set(cards, { opacity: 0, y: 24 })

      ScrollTrigger.batch(cards, {
        onEnter: (elements) => {
          gsap.to(elements, {
            opacity: 1,
            y: 0,
            stagger: 0.1,
            duration: 0.5,
            ease: "power2.out",
            overwrite: true,
          })
        },
        start: "top 85%",
        once: true,
      })
    }, sectionRef)

    return () => ctx.revert()
  }, [])

  return (
    <section ref={sectionRef} className="py-24 border-t border-terminal-border">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <h2 className="font-sans text-3xl sm:text-4xl font-bold tracking-tighter text-terminal-fg">
          {t.features.headline}
        </h2>

        <div ref={cardsRef} className="mt-12 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {t.features.items.map((feature, i) => {
            const Icon = ICONS[i]
            return (
              <div
                key={feature.title}
                className="feature-card border border-terminal-border rounded-lg p-6 bg-terminal-bg hover:border-terminal-green/30 transition-colors"
              >
                <div className="w-10 h-10 rounded-lg bg-terminal-border/50 flex items-center justify-center text-terminal-green mb-4">
                  <Icon size={24} weight="regular" />
                </div>
                <h3 className="font-sans text-lg font-semibold text-terminal-fg mb-2">
                  {feature.title}
                </h3>
                <p className="text-terminal-dim text-sm leading-relaxed">
                  {feature.description}
                </p>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
