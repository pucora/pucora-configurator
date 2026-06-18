import { useEffect, useCallback, useRef, useState } from 'react'
import { useProfileStore } from '../store/profileStore'
import { api } from '../api/client'
import { GatewaySidebar } from '../components/GatewaySidebar'
import { RouteList } from '../components/RouteList'
import { RouteInspector } from '../components/RouteInspector'
import { ComponentPalette } from '../components/ComponentPalette'
import { PreviewPanel } from '../components/PreviewPanel'
import { PublishPanel } from '../components/PublishPanel'
import { AsyncAgentsPanel } from '../components/AsyncAgentsPanel'
import { ConfigWarnings } from '../components/ConfigWarnings'
import { loadDraft } from '../lib/draft'

interface BuilderPageProps {
  onBack: () => void
}

export function BuilderPage({ onBack }: BuilderPageProps) {
  const {
    profile,
    catalog,
    selectedRouteIndex,
    loadCatalog,
    validate,
    generatePreview,
  } = useProfileStore()

  const fileRef = useRef<HTMLInputElement>(null)
  const jsonRef = useRef<HTMLInputElement>(null)
  const [routeSearch, setRouteSearch] = useState('')

  useEffect(() => {
    const draft = loadDraft()
    if (draft && draft.routes?.length > 0 && profile.routes.length === 0) {
      useProfileStore.getState().setProfile(draft)
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    loadCatalog()
  }, [loadCatalog])

  const debouncedGenerate = useCallback(() => {
    const t = setTimeout(() => {
      validate().then(() => generatePreview())
    }, 400)
    return () => clearTimeout(t)
  }, [validate, generatePreview])

  useEffect(() => {
    return debouncedGenerate()
  }, [profile, debouncedGenerate])

  const handleImport = async (file: File) => {
    const yaml = await file.text()
    const res = await api.importYaml(yaml)
    useProfileStore.getState().setProfile(res.profile)
  }

  const handleImportJson = async (file: File) => {
    const text = await file.text()
    const config = JSON.parse(text) as Record<string, unknown>
    const res = await api.importJson(config)
    useProfileStore.getState().setProfile(res.profile)
    if (res.warnings?.length) {
      console.warn('Import warnings:', res.warnings)
    }
  }

  const selectedRoute = selectedRouteIndex !== null ? profile.routes[selectedRouteIndex] : null
  const filteredRoutes = profile.routes.filter((r) => {
    if (!routeSearch.trim()) return true
    const q = routeSearch.toLowerCase()
    return (
      r.path.toLowerCase().includes(q) ||
      r.method.toLowerCase().includes(q) ||
      r.backend.type.toLowerCase().includes(q) ||
      (r.backend.host || '').toLowerCase().includes(q)
    )
  })

  return (
    <div className="h-screen flex flex-col bg-slate-950">
      <header className="flex items-center justify-between px-4 py-3 border-b border-slate-800 bg-slate-900/80">
        <div className="flex items-center gap-4">
          <button type="button" onClick={onBack} className="text-sm text-slate-400 hover:text-white">
            ← Presets
          </button>
          <h1 className="font-semibold text-white">{profile.metadata.name}</h1>
        </div>
        <div className="flex items-center gap-2">
          <input
            ref={fileRef}
            type="file"
            accept=".yaml,.yml"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0]
              if (f) handleImport(f)
            }}
          />
          <button
            type="button"
            onClick={() => fileRef.current?.click()}
            className="px-3 py-1.5 text-sm rounded-lg border border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            Import YAML
          </button>
          <input
            ref={jsonRef}
            type="file"
            accept=".json"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0]
              if (f) handleImportJson(f)
            }}
          />
          <button
            type="button"
            onClick={() => jsonRef.current?.click()}
            className="px-3 py-1.5 text-sm rounded-lg border border-slate-700 text-slate-300 hover:bg-slate-800"
          >
            Import JSON
          </button>
          <button
            type="button"
            onClick={() => generatePreview()}
            className="px-4 py-1.5 text-sm rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white font-medium"
          >
            Generate preview
          </button>
        </div>
      </header>

      <div className="flex-1 grid grid-cols-12 min-h-0">
        <aside className="col-span-2 border-r border-slate-800 overflow-y-auto bg-slate-900/30">
          <GatewaySidebar catalog={catalog} />
        </aside>

        <section className="col-span-4 border-r border-slate-800 flex flex-col min-h-0 p-4 gap-4 overflow-y-auto">
          <div className="flex items-center justify-between gap-2">
            <h2 className="text-sm font-semibold text-slate-300">Routes</h2>
            <input
              value={routeSearch}
              onChange={(e) => setRouteSearch(e.target.value)}
              placeholder="Search routes…"
              className="text-xs w-40"
            />
          </div>
          <RouteList routes={filteredRoutes} allRoutes={profile.routes} />
          <ComponentPalette catalog={catalog} />
          <AsyncAgentsPanel />
        </section>

        <section className="col-span-3 border-r border-slate-800 overflow-y-auto bg-slate-900/20">
          <RouteInspector route={selectedRoute} index={selectedRouteIndex} catalog={catalog} />
        </section>

        <section className="col-span-3 flex flex-col min-h-0">
          <PublishPanel />
          <ConfigWarnings />
          <PreviewPanel />
        </section>
      </div>
    </div>
  )
}
