import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Catalog, Preset } from '../types/profile'
import { blankProfile } from '../types/profile'
import { useProfileStore } from '../store/profileStore'
import { BackendFrameworkPanel } from '../components/BackendFrameworkPanel'
import {
  applyFrameworkRecommendations,
  findFramework,
} from '../lib/frameworkRecommendations'

interface PresetsPageProps {
  onStart: () => void
}

export function PresetsPage({ onStart }: PresetsPageProps) {
  const [presets, setPresets] = useState<Preset[]>([])
  const [catalog, setCatalog] = useState<Catalog | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const { profile, setProfile } = useProfileStore()
  const frameworkId = profile.metadata.backend_framework

  useEffect(() => {
    Promise.all([api.presets(), api.catalog()])
      .then(([presetList, cat]) => {
        setPresets(presetList)
        setCatalog(cat)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const applySelectedFramework = (base: ReturnType<typeof blankProfile>, replaceRoutes: boolean) => {
    const fw = findFramework(catalog?.backendFrameworks, frameworkId)
    if (!fw) return base
    return applyFrameworkRecommendations(base, fw, { replaceRoutes, fillEmptyOnly: !replaceRoutes })
  }

  const loadPreset = async (name: string) => {
    let next = await api.preset(name)
    next = applySelectedFramework(next, false)
    setProfile(next)
    onStart()
  }

  const startBlank = () => {
    setProfile(applySelectedFramework(blankProfile(), true))
    onStart()
  }

  return (
    <div className="min-h-screen bg-slate-950">
      <header className="border-b border-slate-800 px-8 py-6">
        <h1 className="text-2xl font-bold text-white">Pucora Configurator</h1>
        <p className="text-slate-400 mt-1">Build gateway configs visually — no YAML or JSON expertise required</p>
      </header>

      <main className="max-w-6xl mx-auto px-8 py-8">
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-950/50 border border-red-800 text-red-300 text-sm">
            API unavailable: {error}. Start the API with <code className="text-red-200">make dev-api</code>
          </div>
        )}

        <div className="mb-8 p-5 rounded-xl border border-slate-800 bg-slate-900/40">
          <BackendFrameworkPanel catalog={catalog} compact />
          {catalog?.backendFrameworks && catalog.backendFrameworks.length > 0 && (
            <p className="text-xs text-slate-500 mt-3 normal-case">
              Optional — pre-fills CORS, headers, routes, and timeouts when you start.
              Change or clear anytime in the gateway sidebar.
            </p>
          )}
        </div>

        <button
          type="button"
          onClick={startBlank}
          className="w-full mb-8 p-6 rounded-xl border-2 border-dashed border-cyan-700/50 bg-cyan-950/20 hover:bg-cyan-950/40 text-left transition-colors"
        >
          <span className="text-lg font-semibold text-cyan-200">Start from scratch</span>
          <p className="text-sm text-slate-400 mt-1">
            Blank gateway with sensible defaults
            {frameworkId ? ' + framework recommendations' : ' (port 8080, CORS, logging)'}
          </p>
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
