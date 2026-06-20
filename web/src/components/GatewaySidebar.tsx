import type { Catalog } from '../types/profile'
import { TagInput } from './TagInput'
import { useProfileStore } from '../store/profileStore'
import { BackendFrameworkPanel } from './BackendFrameworkPanel'

interface GatewaySidebarProps {
  catalog: Catalog | null
}

export function GatewaySidebar({ catalog }: GatewaySidebarProps) {
  const { profile, updateProfile } = useProfileStore()

  return (
    <div className="p-4 space-y-5 overflow-y-auto">
      <h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wider">Gateway</h2>

      <div className="space-y-1.5">
        <label>Name</label>
        <input
          value={profile.metadata.name}
          onChange={(e) => updateProfile((p) => { p.metadata.name = e.target.value; return p })}
        />
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div className="space-y-1.5">
          <label>Port</label>
          <input
            type="number"
            value={profile.gateway.port}
            onChange={(e) => updateProfile((p) => { p.gateway.port = parseInt(e.target.value, 10); return p })}
          />
        </div>
        <div className="space-y-1.5">
          <label>Timeout</label>
          <input
            value={profile.gateway.timeout || '3s'}
            onChange={(e) => updateProfile((p) => { p.gateway.timeout = e.target.value; return p })}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div className="space-y-1.5">
          <label>Write timeout</label>
          <input
            value={profile.gateway.write_timeout || ''}
            onChange={(e) => updateProfile((p) => { p.gateway.write_timeout = e.target.value || undefined; return p })}
            placeholder="0s for streaming"
          />
          <p className="text-xs text-slate-600 normal-case">Use 0s for SSE/long-lived streams</p>
        </div>
        <div className="space-y-1.5">
          <label>Cache TTL</label>
          <input
            value={profile.gateway.cache_ttl || '3600s'}
            onChange={(e) => updateProfile((p) => { p.gateway.cache_ttl = e.target.value; return p })}
          />
        </div>
      </div>

      <BackendFrameworkPanel catalog={catalog} />

      <div className="border-t border-slate-800 pt-4 space-y-3">
        <h3 className="text-xs font-semibold text-slate-400 uppercase">gRPC catalog</h3>
        <p className="text-xs text-slate-500 normal-case">Paths to .pb descriptor files (required for gRPC routes)</p>
        <TagInput
          label="Catalog paths"
          value={profile.grpc?.catalog || []}
          onChange={(v) => updateProfile((p) => {
            if (!p.grpc) p.grpc = {}
            p.grpc.catalog = v
            return p
          })}
          suggestions={['./grpc/catalog.pb', '/etc/pucora/grpc/catalog.pb']}
          placeholder="Path to .pb file"
        />
      </div>

      <div className="border-t border-slate-800 pt-4 space-y-3">
        <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
          <input
            type="checkbox"
            checked={profile.cors?.enabled ?? false}
            onChange={(e) => updateProfile((p) => {
              if (!p.cors) p.cors = { enabled: true, allow_origins: [], allow_headers: [] }
              p.cors.enabled = e.target.checked
              return p
            })}
          />
          Enable CORS
        </label>

        {profile.cors?.enabled && (
          <>
            <div className="flex flex-wrap gap-1">
              {(catalog?.corsPresets || []).map((preset) => (
                <button
                  key={preset.id}
                  type="button"
                  className="px-2 py-0.5 text-xs rounded border border-slate-700 hover:border-cyan-600 text-slate-400"
                  onClick={() => updateProfile((p) => {
                    if (!p.cors) return p
                    p.cors.allow_origins = [...preset.allow_origins]
                    p.cors.allow_methods = [...preset.allow_methods]
                    p.cors.allow_headers = [...preset.allow_headers]
                    return p
                  })}
                >
                  {preset.label}
                </button>
              ))}
            </div>
            <TagInput
              label="Allowed origins"
              value={profile.cors.allow_origins || []}
              onChange={(v) => updateProfile((p) => { if (p.cors) p.cors.allow_origins = v; return p })}
              suggestions={['http://localhost:3000', 'http://localhost:5173', '*']}
            />
            <TagInput
              label="Allowed methods"
              value={profile.cors.allow_methods || []}
              onChange={(v) => updateProfile((p) => { if (p.cors) p.cors.allow_methods = v; return p })}
              suggestions={catalog?.enums.httpMethods}
            />
            <TagInput
              label="Allowed headers"
              value={profile.cors.allow_headers || []}
              onChange={(v) => updateProfile((p) => { if (p.cors) p.cors.allow_headers = v; return p })}
              suggestions={catalog?.commonHeaders}
            />
          </>
        )}
      </div>

      <div className="border-t border-slate-800 pt-4 space-y-3">
        <h3 className="text-xs font-semibold text-slate-400 uppercase">Telemetry</h3>
        <div className="space-y-1.5">
          <label>Log level</label>
          <select
            value={profile.telemetry?.logging?.level || 'INFO'}
            onChange={(e) => updateProfile((p) => {
              if (!p.telemetry) p.telemetry = {}
              if (!p.telemetry.logging) p.telemetry.logging = { stdout: true }
              p.telemetry.logging.level = e.target.value
              return p
            })}
          >
            {(catalog?.enums.logLevels || ['INFO', 'DEBUG', 'WARNING']).map((l) => (
              <option key={l} value={l}>{l}</option>
            ))}
          </select>
        </div>
        <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
          <input
            type="checkbox"
            checked={profile.telemetry?.metrics?.enabled ?? true}
            onChange={(e) => updateProfile((p) => {
              if (!p.telemetry) p.telemetry = {}
              if (!p.telemetry.metrics) p.telemetry.metrics = { enabled: true }
              p.telemetry.metrics.enabled = e.target.checked
              return p
            })}
          />
          Enable metrics
        </label>
      </div>

      <div className="border-t border-slate-800 pt-4 space-y-3">
        <h3 className="text-xs font-semibold text-slate-400 uppercase">Docker Compose</h3>
        <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
          <input
            type="checkbox"
            checked={profile.compose?.enabled ?? false}
            onChange={(e) => updateProfile((p) => {
              if (!p.compose) p.compose = {}
              p.compose.enabled = e.target.checked
              return p
            })}
          />
          Generate docker-compose
        </label>
        {profile.compose?.enabled && (
          <>
            <div className="space-y-1.5">
              <label>Gateway image</label>
              <input
                value={profile.compose.image || ''}
                onChange={(e) => updateProfile((p) => {
                  if (!p.compose) p.compose = {}
                  p.compose.image = e.target.value || undefined
                  return p
                })}
                placeholder="pucora/pucora:2.0.0"
              />
            </div>
            <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
              <input
                type="checkbox"
                checked={profile.compose.mock_backend ?? true}
                onChange={(e) => updateProfile((p) => {
                  if (!p.compose) p.compose = {}
                  p.compose.mock_backend = e.target.checked
                  return p
                })}
              />
              Include mock backends
            </label>
            <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
              <input
                type="checkbox"
                checked={profile.compose.expose_metrics ?? false}
                onChange={(e) => updateProfile((p) => {
                  if (!p.compose) p.compose = {}
                  p.compose.expose_metrics = e.target.checked
                  return p
                })}
              />
              Expose metrics port
            </label>
          </>
        )}
      </div>
    </div>
  )
}
