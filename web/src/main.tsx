import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { LocaleProvider } from "./i18n/context"
import "./index.css"
import App from "./App"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <LocaleProvider>
      <App />
    </LocaleProvider>
  </StrictMode>,
)
