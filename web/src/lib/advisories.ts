import type { GatewayProfile, Route } from '../types/profile'

export interface Advisory {
  field: string
  level: 'warn' | 'info'
  message: string
}

export function computeAdvisories(profile: GatewayProfile): Advisory[] {
  const out: Advisory[] = []
  let hasGRPC = false

  for (let i = 0; i < profile.routes.length; i++) {
    const r = profile.routes[i]
    const prefix = `routes[${i}]`

    if (r.backend.type === 'grpc') hasGRPC = true

    if (r.auth?.jwt?.jwk_url) {
      const forwards = r.headers?.forward?.some((h) => h.toLowerCase() === 'authorization')
      if (!forwards) {
        out.push({
          field: `${prefix}.headers.forward`,
          level: 'warn',
          message: 'JWT enabled but Authorization is not forwarded to the backend',
        })
      }
      if (r.auth.jwt.disable_jwk_security) {
        out.push({
          field: `${prefix}.auth.jwt.disable_jwk_security`,
          level: 'warn',
          message: 'disable_jwk_security is for development only',
        })
      }
    }

    if (r.backend.type === 'websocket' && r.auth?.jwt) {
      const wsAuth = r.websocket?.input_headers?.some((h) => h.toLowerCase() === 'authorization')
      if (!wsAuth) {
        out.push({
          field: `${prefix}.websocket.input_headers`,
          level: 'warn',
          message: 'WebSocket JWT requires Authorization in websocket input_headers',
        })
      }
    }

    const streaming = r.output_encoding === 'no-op' || r.backend.encoding === 'no-op'
    if (streaming && profile.gateway.write_timeout !== '0s') {
      out.push({
        field: 'gateway.write_timeout',
        level: 'warn',
        message: `Route ${r.method} ${r.path} streams — set gateway write_timeout to 0s`,
      })
    }
  }

  if (hasGRPC && (!profile.grpc?.catalog || profile.grpc.catalog.length === 0)) {
    out.push({
      field: 'grpc.catalog',
      level: 'warn',
      message: 'gRPC routes exist but catalog paths are empty',
    })
  }

  return out
}

export function buildCurlExample(route: Route, port: number): string {
  const base = `http://localhost:${port}${route.path}`
  const method = route.method.toUpperCase()
  const lines = [`curl -X ${method} '${base}'`]

  if (route.auth?.jwt) {
    lines.push("  -H 'Authorization: Bearer <token>'")
  }
  if (route.headers?.forward?.includes('Content-Type') || method === 'POST') {
    lines.push("  -H 'Content-Type: application/json'")
  }
  if (method === 'POST') {
    lines.push("  -d '{}'")
  }

  return lines.join(' \\\n')
}
