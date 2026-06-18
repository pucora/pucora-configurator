import type { Catalog } from '../types/profile'
import { useProfileStore } from '../store/profileStore'

interface ComponentPaletteProps {
  catalog: Catalog | null
}

export function ComponentPalette({ catalog }: ComponentPaletteProps) {
  const { updateProfile, setSelectedRoute } = useProfileStore()

  if (!catalog) return null

  const addTemplate = (templateId: string) => {
    const tmpl = catalog.routeTemplates.find((t) => t.id === templateId)
    if (!tmpl) return

    const route = {
      path: tmpl.defaults.path || '/api',
      method: tmpl.method,
      ...tmpl.defaults,
      backend: {
        type: tmpl.backendType,
        host: '',
        path: '/',
        ...tmpl.defaults.backend,
      },
    }

    updateProfile((p) => {
      if (templateId === 'sse') {
        p.gateway.write_timeout = '0s'
      }
      if (templateId === 'grpc' && (!p.grpc?.catalog || p.grpc.catalog.length === 0)) {
        p.grpc = { catalog: ['./grpc/catalog.pb'] }
      }
      p.routes.push(route)
      return p
    })
    setSelectedRoute(useProfileStore.getState().profile.routes.length - 1)
  }

  return (
    <div className="space-y-2">
      <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Add route</h3>
      <div className="flex flex-wrap gap-2">
        {catalog.routeTemplates.map((t) => (
          <button
            key={t.id}
            type="button"
            onClick={() => addTemplate(t.id)}
            className="px-3 py-1.5 text-xs rounded-lg border border-slate-700 bg-slate-900 hover:border-cyan-600 hover:bg-cyan-950/30 text-slate-300 transition-colors"
            title={t.label}
          >
            + {t.label}
          </button>
        ))}
      </div>
    </div>
  )
}
