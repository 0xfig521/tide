import { createContext, useContext, useState, useCallback, type ReactNode } from "react"
import { translations, type Locale, type TranslationSchema } from "./translations"

function getInitialLocale(): Locale {
  const stored = localStorage.getItem("tide-locale")
  if (stored === "en" || stored === "zh") return stored
  if (typeof navigator === "undefined") return "en"
  const browserLang = navigator.language.toLowerCase()
  return browserLang.startsWith("zh") ? "zh" : "en"
}

interface LocaleContextValue {
  locale: Locale
  setLocale: (l: Locale) => void
  t: TranslationSchema
}

const LocaleContext = createContext<LocaleContextValue | null>(null)

export function LocaleProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(getInitialLocale)

  const setLocale = useCallback((l: Locale) => {
    setLocaleState(l)
    localStorage.setItem("tide-locale", l)
  }, [])

  const t = translations[locale]

  return (
    <LocaleContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </LocaleContext.Provider>
  )
}

export function useLocale() {
  const ctx = useContext(LocaleContext)
  if (!ctx) throw new Error("useLocale must be used within LocaleProvider")
  return ctx
}
