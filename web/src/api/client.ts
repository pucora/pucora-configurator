import type { GatewayProfile } from '../types/profile'

const API_BASE = import.meta.env.VITE_API_URL || ''

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

export interface ConfigBundle {
  name: string
  updated_at: string
  profile_yaml?: string
  velonetics_json: Record<string, unknown>
  env?: Record<string, string>
  compose_yaml?: string
}

export interface PublishResult {
  saved: boolean
  name: string
  get_url: {
    bundle: string
    velonetics_json: string
    profile_yaml: string
  }
  updated_at: string
}

export const api = {
  health: () => request<{ status: string }>('/api/health'),
  presets: () => request<import('../types/profile').Preset[]>('/api/presets'),
  preset: (name: string) => request<GatewayProfile>(`/api/presets/${name}`),
  catalog: () => request<import('../types/profile').Catalog>('/api/catalog'),
  validate: (profile: GatewayProfile) =>
    request<{ valid: boolean; errors: import('../types/profile').ValidationError[] }>('/api/validate', {
      method: 'POST',
      body: JSON.stringify({ profile }),
    }),
  generate: (profile: GatewayProfile, compose = false) =>
    request<{
      valid: boolean
      velonetics_json: Record<string, unknown>
      profile_yaml: string
      env: Record<string, string>
      compose_yaml?: string
      warnings: string[]
      advisories?: import('../lib/advisories').Advisory[]
      errors?: import('../types/profile').ValidationError[]
    }>('/api/generate', {
      method: 'POST',
      body: JSON.stringify({ profile, compose }),
    }),
  importYaml: (yaml: string) =>
    request<{ profile: GatewayProfile; valid: boolean; errors: import('../types/profile').ValidationError[] }>('/api/import', {
      method: 'POST',
      body: JSON.stringify({ yaml }),
    }),
  importJson: (config: Record<string, unknown>) =>
    request<{
      profile: GatewayProfile
      valid: boolean
      errors: import('../types/profile').ValidationError[]
      warnings: string[]
    }>('/api/import-json', {
      method: 'POST',
      body: JSON.stringify({ config }),
    }),
  listConfigs: () => request<{ configs: string[] }>('/api/configs'),
  getConfig: (name = 'default') => request<ConfigBundle>(`/api/config/${name}`),
  publishConfig: (profile: GatewayProfile, name: string, compose = false, apiKey?: string) =>
    request<PublishResult>(`/api/config/${name}`, {
      method: 'POST',
      headers: apiKey ? { 'X-API-Key': apiKey } : undefined,
      body: JSON.stringify({ name, profile, compose }),
    }),
  getConfigVeloneticsUrl: (name = 'default') =>
    `${API_BASE || window.location.origin}/api/config/${name}/velonetics.json`,
}
