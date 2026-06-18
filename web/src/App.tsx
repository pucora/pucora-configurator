import { useState } from 'react'
import { PresetsPage } from './pages/PresetsPage'
import { BuilderPage } from './pages/BuilderPage'

type View = 'presets' | 'builder'

function App() {
  const [view, setView] = useState<View>('presets')

  if (view === 'builder') {
    return <BuilderPage onBack={() => setView('presets')} />
  }

  return <PresetsPage onStart={() => setView('builder')} />
}

export default App
