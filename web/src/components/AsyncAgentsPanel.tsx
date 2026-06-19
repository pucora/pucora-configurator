import { useProfileStore } from '../store/profileStore'
import type { AsyncAgent } from '../types/profile'

function blankAgent(): AsyncAgent {
  return {
    name: 'events-consumer',
    consumer: { topic: 'events', workers: 2, timeout: '2s' },
    backend: { host: 'http://localhost:8081', path: '/ingest', method: 'POST' },
    connection: { max_retries: 3, backoff_strategy: 'linear', health_interval: '10s' },
    kafka: { brokers: ['localhost:9092'], group_id: 'my-group', client_id: 'pucora_async' },
  }
}

export function AsyncAgentsPanel() {
  const { profile, updateProfile } = useProfileStore()
  const agents = profile.async_agents || []

  const updateAgent = (idx: number, fn: (a: AsyncAgent) => void) => {
    updateProfile((p) => {
      if (!p.async_agents) p.async_agents = []
      fn(p.async_agents[idx])
      return p
    })
  }

  const addAgent = () => {
    updateProfile((p) => {
      if (!p.async_agents) p.async_agents = []
      p.async_agents.push(blankAgent())
      return p
    })
  }

  const removeAgent = (idx: number) => {
    updateProfile((p) => {
      p.async_agents = (p.async_agents || []).filter((_, i) => i !== idx)
      return p
    })
  }

  return (
    <div className="space-y-3 border-t border-slate-800 pt-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-slate-300">Async agents</h2>
        <button
          type="button"
          onClick={addAgent}
          className="text-xs px-2 py-1 rounded border border-slate-700 hover:border-cyan-600 text-slate-400"
        >
          + Add consumer
        </button>
      </div>
      <p className="text-xs text-slate-500">Background Kafka consumers (no HTTP client required)</p>

      {agents.length === 0 && (
        <p className="text-xs text-slate-600 italic">No async agents configured</p>
      )}

      {agents.map((agent, idx) => (
        <div key={idx} className="rounded-lg border border-slate-700 p-3 space-y-3 bg-slate-900/40">
          <div className="flex justify-between items-center">
            <span className="text-xs font-medium text-cyan-300/80">Agent {idx + 1}</span>
            <button type="button" onClick={() => removeAgent(idx)} className="text-xs text-red-400">Remove</button>
          </div>

          <div className="space-y-1.5">
            <label>Name</label>
            <input
              value={agent.name}
              onChange={(e) => updateAgent(idx, (a) => { a.name = e.target.value })}
            />
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1.5">
              <label>Kafka topic</label>
              <input
                value={agent.consumer.topic}
                onChange={(e) => updateAgent(idx, (a) => { a.consumer.topic = e.target.value })}
              />
            </div>
            <div className="space-y-1.5">
              <label>Workers</label>
              <input
                type="number"
                value={agent.consumer.workers ?? 1}
                onChange={(e) => updateAgent(idx, (a) => { a.consumer.workers = parseInt(e.target.value, 10) })}
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <label>Consumer timeout</label>
            <input
              value={agent.consumer.timeout || '2s'}
              onChange={(e) => updateAgent(idx, (a) => { a.consumer.timeout = e.target.value })}
            />
          </div>

          <div className="space-y-1.5">
            <label>Kafka brokers (comma-separated)</label>
            <input
              value={(agent.kafka?.brokers || []).join(', ')}
              onChange={(e) => updateAgent(idx, (a) => {
                if (!a.kafka) a.kafka = { brokers: [], group_id: 'my-group' }
                a.kafka.brokers = e.target.value.split(',').map((s) => s.trim()).filter(Boolean)
              })}
              placeholder="redpanda:9092, localhost:9092"
            />
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1.5">
              <label>Consumer group ID</label>
              <input
                value={agent.kafka?.group_id || ''}
                onChange={(e) => updateAgent(idx, (a) => {
                  if (!a.kafka) a.kafka = { brokers: [], group_id: '' }
                  a.kafka.group_id = e.target.value
                })}
              />
            </div>
            <div className="space-y-1.5">
              <label>Client ID</label>
              <input
                value={agent.kafka?.client_id || ''}
                onChange={(e) => updateAgent(idx, (a) => {
                  if (!a.kafka) a.kafka = { brokers: [], group_id: 'my-group' }
                  a.kafka.client_id = e.target.value
                })}
              />
            </div>
          </div>

          <div className="border-t border-slate-800 pt-2 space-y-2">
            <p className="text-xs text-slate-400 uppercase tracking-wide">Webhook backend</p>
            <div className="space-y-1.5">
              <label>Host</label>
              <input
                value={agent.backend.host}
                onChange={(e) => updateAgent(idx, (a) => { a.backend.host = e.target.value })}
              />
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div className="space-y-1.5">
                <label>Path</label>
                <input
                  value={agent.backend.path}
                  onChange={(e) => updateAgent(idx, (a) => { a.backend.path = e.target.value })}
                />
              </div>
              <div className="space-y-1.5">
                <label>Method</label>
                <input
                  value={agent.backend.method || 'POST'}
                  onChange={(e) => updateAgent(idx, (a) => { a.backend.method = e.target.value })}
                />
              </div>
            </div>
          </div>

          <div className="border-t border-slate-800 pt-2 grid grid-cols-3 gap-2">
            <div className="space-y-1.5">
              <label>Max retries</label>
              <input
                type="number"
                value={agent.connection?.max_retries ?? 3}
                onChange={(e) => updateAgent(idx, (a) => {
                  if (!a.connection) a.connection = {}
                  a.connection.max_retries = parseInt(e.target.value, 10)
                })}
              />
            </div>
            <div className="space-y-1.5">
              <label>Backoff</label>
              <select
                value={agent.connection?.backoff_strategy || 'linear'}
                onChange={(e) => updateAgent(idx, (a) => {
                  if (!a.connection) a.connection = {}
                  a.connection.backoff_strategy = e.target.value
                })}
              >
                <option value="linear">linear</option>
                <option value="fallback">fallback</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label>Health interval</label>
              <input
                value={agent.connection?.health_interval || '10s'}
                onChange={(e) => updateAgent(idx, (a) => {
                  if (!a.connection) a.connection = {}
                  a.connection.health_interval = e.target.value
                })}
              />
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
