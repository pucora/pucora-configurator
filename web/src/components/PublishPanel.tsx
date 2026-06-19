import { useState, useEffect } from 'react'
import { api } from '../api/client'
import { useProfileStore } from '../store/profileStore'

export function PublishPanel() {
  const { profile, composeEnabled } = useProfileStore()
  const [configName, setConfigName] = useState('default')
  const [apiKey, setApiKey] = useState('')
  const [status, setStatus] = useState<'idle' | 'loading' | 'ok' | 'error'>('idle')
  const [message, setMessage] = useState('')
  const [savedPullUrl, setSavedPullUrl] = useState('')
  const [savedConfigs, setSavedConfigs] = useState<string[]>([])

  useEffect(() => {
    api.listConfigs()
      .then((r) => setSavedConfigs(r.configs || []))
      .catch(() => setSavedConfigs([]))
  }, [status])

  const refreshList = () => {
    api.listConfigs()
      .then((r) => setSavedConfigs(r.configs || []))
      .catch(() => setSavedConfigs([]))
  }

  const publish = async () => {
    setStatus('loading')
    try {
      const res = await api.publishConfig(profile, configName, composeEnabled, apiKey || undefined)
      setSavedPullUrl(res.get_url.velonetics_json)
      setMessage(`Saved as "${res.name}"`)
      setStatus('ok')
      refreshList()
    } catch (e) {
      setMessage(e instanceof Error ? e.message : 'Publish failed')
      setStatus('error')
    }
  }

  const loadFromApi = async () => {
    setStatus('loading')
    try {
      const bundle = await api.getConfig(configName)
      if (bundle.profile_yaml) {
        const imported = await api.importYaml(bundle.profile_yaml)
        useProfileStore.getState().setProfile(imported.profile)
      }
      setSavedPullUrl(api.getConfigPucoraUrl(configName))
      setMessage(`Loaded "${configName}" from API`)
      setStatus('ok')
    } catch (e) {
      setMessage(e instanceof Error ? e.message : 'Load failed')
      setStatus('error')
    }
  }

  const gatewayPullUrl = savedPullUrl || api.getConfigPucoraUrl(configName)

  return (
    <div className="px-4 py-3 border-b border-slate-800 bg-slate-900/30 space-y-2">
      <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">API publish / pull</h3>
      <div className="flex gap-2">
        <input
          value={configName}
          onChange={(e) => setConfigName(e.target.value)}
          placeholder="config name (e.g. prod)"
          className="flex-1 text-xs"
          list="saved-configs"
        />
        <datalist id="saved-configs">
          {savedConfigs.map((n) => (
            <option key={n} value={n} />
          ))}
        </datalist>
      </div>
      {savedConfigs.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {savedConfigs.map((n) => (
            <button
              key={n}
              type="button"
              onClick={() => setConfigName(n)}
              className="px-2 py-0.5 text-xs rounded border border-slate-700 text-slate-400 hover:border-cyan-700"
            >
              {n}
            </button>
          ))}
        </div>
      )}
      <input
        value={apiKey}
        onChange={(e) => setApiKey(e.target.value)}
        placeholder="API key (if CONFIG_API_KEY is set)"
        type="password"
        className="text-xs"
      />
      <div className="flex gap-2">
        <button
          type="button"
          onClick={publish}
          disabled={status === 'loading'}
          className="flex-1 px-2 py-1.5 text-xs rounded bg-cyan-800 hover:bg-cyan-700 text-white disabled:opacity-50"
        >
          POST — Save to API
        </button>
        <button
          type="button"
          onClick={loadFromApi}
          disabled={status === 'loading'}
          className="flex-1 px-2 py-1.5 text-xs rounded border border-slate-700 hover:bg-slate-800 text-slate-300 disabled:opacity-50"
        >
          GET — Load from API
        </button>
      </div>
      <div className="text-xs text-slate-500">
        <span className="text-slate-400">Gateway pull URL:</span>
        <code className="block mt-1 p-1.5 bg-slate-950 rounded text-cyan-300/80 break-all">{gatewayPullUrl}</code>
      </div>
      {message && (
        <p className={`text-xs ${status === 'error' ? 'text-red-400' : 'text-green-400'}`}>{message}</p>
      )}
    </div>
  )
}
