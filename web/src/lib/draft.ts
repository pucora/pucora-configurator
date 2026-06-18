const DRAFT_KEY = 'velonetics-configurator-draft'

import type { GatewayProfile } from '../types/profile'

export function saveDraft(profile: GatewayProfile): void {
  try {
    localStorage.setItem(DRAFT_KEY, JSON.stringify(profile))
  } catch {
    // ignore quota errors
  }
}

export function loadDraft(): GatewayProfile | null {
  try {
    const raw = localStorage.getItem(DRAFT_KEY)
    if (!raw) return null
    return JSON.parse(raw) as GatewayProfile
  } catch {
    return null
  }
}

export function clearDraft(): void {
  localStorage.removeItem(DRAFT_KEY)
}
