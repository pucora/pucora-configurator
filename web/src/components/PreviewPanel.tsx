import { useState } from 'react'
import JSZip from 'jszip'
import { useProfileStore } from '../store/profileStore'

function download(filename: string, content: string, type = 'application/json') {
  const blob = new Blob([content], { type })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export function PreviewPanel() {
  const { preview, validationErrors, composeEnabled, setComposeEnabled } = useProfileStore()
  const [tab, setTab] = useState<'json' | 'yaml' | 'compose'>('json')

  const downloadZip = async () => {
    if (!preview) return
    const zip = new JSZip()
    zip.file('pucora.json', preview.veloneticsJson)
    zip.file('profile.yaml', preview.profileYaml)
    if (Object.keys(preview.env).length) {
      zip.file('.env', Object.entries(preview.env).map(([k, v]) => `${k}=${v}`).join('\n'))
    }
    if (preview.composeYaml) {
      zip.file('docker-compose.yml', preview.composeYaml)
    }
    const blob = await zip.generateAsync({ type: 'blob' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'velonetics-config.zip'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex flex-col h-full border-t border-slate-800">
      <div className="flex items-center justify-between px-4 py-2 border-b border-slate-800 bg-slate-900/50">
        <div className="flex gap-1">
          {(['json', 'yaml', 'compose'] as const).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTab(t)}
              className={`px-3 py-1 text-xs rounded ${tab === t ? 'bg-cyan-900/50 text-cyan-200' : 'text-slate-500 hover:text-slate-300'}`}
            >
              {t === 'json' ? 'pucora.json' : t === 'yaml' ? 'profile.yaml' : 'docker-compose'}
            </button>
          ))}
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-1.5 text-xs text-slate-400 normal-case">
            <input
              type="checkbox"
              checked={composeEnabled}
              onChange={(e) => setComposeEnabled(e.target.checked)}
            />
            Include compose
          </label>
          {preview && (
            <>
              <button
                type="button"
                onClick={() => download('pucora.json', preview.veloneticsJson)}
                className="px-2 py-1 text-xs rounded bg-slate-800 hover:bg-slate-700 text-slate-300"
              >
                Download JSON
              </button>
              <button
                type="button"
                onClick={downloadZip}
                className="px-2 py-1 text-xs rounded bg-cyan-800 hover:bg-cyan-700 text-white"
              >
                Download ZIP
              </button>
            </>
          )}
        </div>
      </div>

      {validationErrors.length > 0 && (
        <div className="px-4 py-2 bg-red-950/40 border-b border-red-900/50 text-xs text-red-300">
          {validationErrors.map((e) => (
            <div key={e.field}>{e.field}: {e.message}</div>
          ))}
        </div>
      )}

      {preview?.warnings.map((w) => (
        <div key={w} className="px-4 py-1 bg-amber-950/30 text-xs text-amber-300">{w}</div>
      ))}

      <pre className="flex-1 overflow-auto p-4 text-xs font-mono text-slate-300 bg-slate-950/50">
        {!preview
          ? 'Click "Generate preview" to see output'
          : tab === 'json'
            ? preview.veloneticsJson
            : tab === 'yaml'
              ? preview.profileYaml
              : preview.composeYaml || '# Enable compose to generate docker-compose.yml'}
      </pre>
    </div>
  )
}
