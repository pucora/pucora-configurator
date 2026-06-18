import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Preset } from '../types/profile'
import { blankProfile } from '../types/profile'
import { useProfileStore } from '../store/profileStore'

interface PresetsPageProps {
  onStart: () => void
}

export function PresetsPage({ onStart }: PresetsPageProps) {
  const [presets, setPresets] = useState<Preset[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const { setProfile } = useProfileStore()

  useEffect(() => {
    api.presets()
      .then(setPresets)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const loadPreset = async (name: string) => {
    const profile = await api.preset(name)
    setProfile(profile)
    onStart()
  }

  const startBlank = () => {
    setProfile(blankProfile())
    onStart()
  }

  return (
    <div className="min-h-screen bg-slate-950">
      <header className="border-b border-slate-800 px-8 py-6">
        <h1 className="text-2xl font-bold text-white">Velonetics Configurator</h1>
        <p className="text-slate-400 mt-1">Build gateway configs visually — no YAML or JSON expertise required</p>
      </header>

      <main className="max-w-6xl mx-auto px-8 py-8">
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-950/50 border border-red-800 text-red-300 text-sm">
            API unavailable: {error}. Start the API with <code className="text-red-200">make dev-api</code>
          </div>
        )}

        <button
          type="button"
          onClick={startBlank}
          className="w-full mb-8 p-6 rounded-xl border-2 border-dashed border-cyan-700/50 bg-cyan-950/20 hover:bg-cyan-950/40 text-left transition-colors"
        >
          <span className="text-lg font-semibold text-cyan-200">Start from scratch</span>
          <p className="text-sm text-slate-400 mt-1">Blank gateway with sensible defaults (port 8080, CORS, logging)</p>
        </button>

        <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-4">Or choose a preset</h2>

        {loading ? (
          <p className="text-slate-500">Loading presets...</p>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {presets.map((p) => (
              <button
                key={p.Name}
                type="button"
                onClick={() => loadPreset(p.Name)}
                className="p-5 rounded-xl border border-slate-800 bg-slate-900/50 hover:border-cyan-600 hover:bg-slate-900 text-left transition-colors"
              >
                <h3 className="font-semibold text-white">{p.Name}</h3>
                <p className="text-sm text-slate-400 mt-2 line-clamp-2">{p.Description}</p>
              </button>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
