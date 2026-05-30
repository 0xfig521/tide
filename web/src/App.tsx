import { Nav } from './components/Nav'
import { Hero } from './components/Hero'
import { Features } from './components/Features'
import { Install } from './components/Install'
import { QuickStart } from './components/QuickStart'
import { AISkill } from './components/AISkill'
import { Footer } from './components/Footer'

function App() {
  return (
    <div className="min-h-[100dvh] bg-terminal-bg text-terminal-fg font-sans">
      <Nav />
      <Hero />
      <Features />
      <Install />
      <QuickStart />
      <AISkill />
      <Footer />
    </div>
  )
}

export default App
