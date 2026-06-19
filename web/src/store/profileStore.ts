import { create } from 'zustand'
import type { Catalog, GatewayProfile, ValidationError } from '../types/profile'
import { blankProfile } from '../types/profile'
import { api } from '../api/client'
import type { Advisory } from '../lib/advisories'
import { saveDraft } from '../lib/draft'

interface ProfileState {
  profile: GatewayProfile
  catalog: Catalog | null
  selectedRouteIndex: number | null
  validationErrors: ValidationError[]
  advisories: Advisory[]
  preview: {
    pucoraJson: string
    profileYaml: string
    env: Record<string, string>
    composeYaml: string
    warnings: string[]
    advisories: Advisory[]
  } | null
  composeEnabled: boolean
  setProfile: (profile: GatewayProfile) => void
  updateProfile: (fn: (p: GatewayProfile) => GatewayProfile) => void
  setSelectedRoute: (index: number | null) => void
  loadCatalog: () => Promise<void>
  validate: () => Promise<void>
  generatePreview: () => Promise<void>
  setComposeEnabled: (v: boolean) => void
}

export const useProfileStore = create<ProfileState>((set, get) => ({
  profile: blankProfile(),
  catalog: null,
  selectedRouteIndex: null,
  validationErrors: [],
  advisories: [],
  preview: null,
  composeEnabled: false,

  setProfile: (profile) => {
    saveDraft(profile)
    set({ profile, selectedRouteIndex: profile.routes.length ? 0 : null })
  },

  updateProfile: (fn) => {
    const profile = fn(structuredClone(get().profile))
    saveDraft(profile)
    set({ profile })
  },

  setSelectedRoute: (index) => set({ selectedRouteIndex: index }),

  loadCatalog: async () => {
    const catalog = await api.catalog()
    set({ catalog })
  },

  validate: async () => {
    const { profile } = get()
    const res = await api.validate(profile)
    set({ validationErrors: res.errors || [] })
  },

  generatePreview: async () => {
    const { profile, composeEnabled } = get()
    const compose = composeEnabled || profile.compose?.enabled || false
    const res = await api.generate(profile, compose)
    if (!res.valid) {
      set({ validationErrors: res.errors || [], preview: null })
      return
    }
    set({
      validationErrors: [],
      advisories: res.advisories || [],
      preview: {
        pucoraJson: JSON.stringify(res.pucora_json, null, 2),
        profileYaml: res.profile_yaml,
        env: res.env || {},
        composeYaml: res.compose_yaml || '',
        warnings: res.warnings || [],
        advisories: res.advisories || [],
      },
    })
  },

  setComposeEnabled: (v) => set({ composeEnabled: v }),
}))
