import { useState } from 'react'
import type { BackendFramework, Catalog } from '../types/profile'
import { useProfileStore } from '../store/profileStore'
import {
  applyFrameworkRecommendations,
  findFramework,
  getFrameworkTips,
  shouldAutoApplyFramework,
} from '../lib/frameworkRecommendations'

interface BackendFrameworkPanelProps {
  catalog: Catalog | null
  compact?: boolean
}

export function BackendFrameworkPanel({ catalog, compact = false }: BackendFrameworkPanelProps) {
  const { profile, updateProfile } = useProfileStore()
  const frameworks = catalog?.backendFrameworks ?? []
  const selectedId = profile.metadata.backend_framework ?? ''
  const [replaceRoutes, setReplaceRoutes] = useState(false)
  const [message, setMessage] = useState('')

  const selected = findFramework(frameworks, selectedId)
  const tips = getFrameworkTips(frameworks, selectedId)

  const applyFramework = (fw: BackendFramework, auto = false) => {
    updateProfile((p) =>
      applyFrameworkRecommendations(p, fw, {
        replaceRoutes: auto ? true : replaceRoutes,
        fillEmptyOnly: !auto,
      }),
    )
    setMessage(`Applied ${fw.label} recommendations`)
    setTimeout(() => setMessage(''), 3000)
  }

  const onSelect = (id: string) => {
    if (!id) {
      updateProfile((p) => {
        p.metadata.backend_framework = undefined
        return p
      })
      setMessage('')
      return
    }

    const fw = findFramework(frameworks, id)
    if (!fw) return

    if (shouldAutoApplyFramework(profile)) {
      applyFramework(fw, true)
      return
    }

    updateProfile((p) => {
      p.metadata.backend_framework = id
      return p
    })
  }

  if (frameworks.length === 0) return null

  return (
    <div className={`space-y-3 ${compact ? '' : 'border-t border-slate-800 pt-4'}`}>
      <div>
        <h3 className="text-xs font-semibold text-slate-400 uppercase">Backend framework</h3>
        <p className="text-xs text-slate-500 normal-case mt-0.5">
          Optional — we&apos;ll suggest gateway settings for your stack
        </p>
      </div>

      <select
        value={selectedId}
        onChange={(e) => onSelect(e.target.value)}
        className="w-full text-sm"
      >
        <option value="">Not specified</option>
        {frameworks.map((f) => (
          <option key={f.id} value={f.id}>
            {f.label} ({f.language})
          </option>
        ))}
      </select>

      {selected && (
        <>
          <p className="text-xs text-slate-400 normal-case">{selected.description}</p>
          <p className="text-xs text-slate-500">
            Default backend port: <span className="text-cyan-400/80">{selected.default_port}</span>
          </p>

          {!compact && profile.routes.length > 0 && (
            <label className="flex items-center gap-2 normal-case text-xs text-slate-400">
              <input
                type="checkbox"
                checked={replaceRoutes}
                onChange={(e) => setReplaceRoutes(e.target.checked)}
              />
              Replace existing routes with starter routes
            </label>
          )}

          <button
            type="button"
            onClick={() => applyFramework(selected)}
            className="w-full px-2 py-1.5 text-xs rounded bg-cyan-900/50 hover:bg-cyan-800/60 text-cyan-200 border border-cyan-800/50"
          >
            Apply recommended settings
          </button>

          {tips.length > 0 && (
            <div className="rounded-lg bg-slate-900/60 border border-slate-800 p-2 space-y-1">
              <p className="text-xs font-medium text-slate-400 uppercase">Recommendations</p>
              <ul className="text-xs text-slate-400 normal-case space-y-1 list-disc list-inside">
                {tips.map((tip) => (
                  <li key={tip}>{tip}</li>
                ))}
              </ul>
            </div>
          )}
        </>
      )}

      {message && <p className="text-xs text-green-400">{message}</p>}
    </div>
  )
}
