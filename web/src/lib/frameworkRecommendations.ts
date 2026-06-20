import type { BackendFramework, GatewayProfile, Route } from '../types/profile'

export interface ApplyFrameworkOptions {
  /** Replace existing routes with framework starter routes */
  replaceRoutes?: boolean
  /** Only set fields that are empty or still at generic defaults */
  fillEmptyOnly?: boolean
}

function mergeRoute(base: Route, patch: Partial<Route>): Route {
  return {
    ...base,
    ...patch,
    headers: patch.headers ?? base.headers,
    query_strings: patch.query_strings ?? base.query_strings,
    backend: { ...base.backend, ...patch.backend } as Route['backend'],
    auth: patch.auth ?? base.auth,
    websocket: patch.websocket ?? base.websocket,
    rate_limit: patch.rate_limit ?? base.rate_limit,
  }
}

function isBlankProfile(profile: GatewayProfile): boolean {
  return profile.routes.length === 0 && !profile.metadata.backend_framework
}

export function findFramework(
  frameworks: BackendFramework[] | undefined,
  id: string | undefined,
): BackendFramework | undefined {
  if (!id || !frameworks) return undefined
  return frameworks.find((f) => f.id === id)
}

export function applyFrameworkRecommendations(
  profile: GatewayProfile,
  framework: BackendFramework,
  options: ApplyFrameworkOptions = {},
): GatewayProfile {
  const { replaceRoutes = false, fillEmptyOnly = true } = options
  const rec = framework.recommendations
  const p = structuredClone(profile)

  p.metadata.backend_framework = framework.id
  if (!p.metadata.description) {
    p.metadata.description = `${framework.label} backend`
  }

  if (rec.gateway) {
    if (!fillEmptyOnly) {
      p.gateway = { ...p.gateway, ...rec.gateway }
    } else {
      if (rec.gateway.timeout && p.gateway.timeout === '3s') p.gateway.timeout = rec.gateway.timeout
      if (rec.gateway.cache_ttl && p.gateway.cache_ttl === '3600s') p.gateway.cache_ttl = rec.gateway.cache_ttl
      if (rec.gateway.write_timeout && !p.gateway.write_timeout) p.gateway.write_timeout = rec.gateway.write_timeout
    }
  }

  if (rec.cors) {
    p.cors = {
      enabled: rec.cors.enabled ?? true,
      allow_origins: rec.cors.allow_origins ?? p.cors?.allow_origins ?? [],
      allow_methods: rec.cors.allow_methods ?? p.cors?.allow_methods ?? [],
      allow_headers: rec.cors.allow_headers ?? p.cors?.allow_headers ?? [],
      expose_headers: rec.cors.expose_headers ?? p.cors?.expose_headers,
      allow_credentials: rec.cors.allow_credentials ?? p.cors?.allow_credentials,
      max_age: rec.cors.max_age ?? p.cors?.max_age,
    }
  }

  if (rec.telemetry) {
    p.telemetry = {
      logging: { stdout: true, ...p.telemetry?.logging, ...rec.telemetry.logging },
      metrics: { enabled: true, ...p.telemetry?.metrics, ...rec.telemetry.metrics },
      usage: { enabled: false, ...p.telemetry?.usage, ...rec.telemetry.usage },
    }
  }

  if (rec.compose) {
    p.compose = { ...p.compose, ...rec.compose }
  }

  if (rec.grpc?.catalog?.length) {
    if (!p.grpc?.catalog?.length || !fillEmptyOnly) {
      p.grpc = { catalog: [...rec.grpc.catalog] }
    }
  }

  if (rec.routes?.length) {
    const shouldAddRoutes = replaceRoutes || p.routes.length === 0 || !fillEmptyOnly
    if (shouldAddRoutes) {
      p.routes = rec.routes.map((r) =>
        mergeRoute(
          {
            path: '/',
            method: 'GET',
            output_encoding: 'json',
            backend: { type: 'http', host: '', path: '/' },
          },
          r,
        ),
      )
    } else if (fillEmptyOnly) {
      p.routes = p.routes.map((route) => {
        if (route.backend.type !== 'http') return route
        const hostEmpty = !route.backend.host || route.backend.host === 'http://localhost:8000'
        if (!hostEmpty) return route
        const starter = rec.routes?.[0]
        if (!starter?.backend?.host) return route
        return mergeRoute(route, {
          headers: route.headers ?? starter.headers,
          query_strings: route.query_strings ?? starter.query_strings,
          backend: { ...route.backend, host: starter.backend.host },
        })
      })
    }
  }

  return p
}

export function shouldAutoApplyFramework(profile: GatewayProfile): boolean {
  return isBlankProfile(profile)
}

export function getFrameworkTips(
  frameworks: BackendFramework[] | undefined,
  frameworkId: string | undefined,
): string[] {
  return findFramework(frameworks, frameworkId)?.recommendations.tips ?? []
}
