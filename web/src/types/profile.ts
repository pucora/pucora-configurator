export interface GatewayProfile {
  apiVersion: string
  kind: string
  metadata: { name: string; description?: string }
  gateway: { port: number; timeout?: string; cache_ttl?: string; write_timeout?: string }
  cors?: CORS
  telemetry?: Telemetry
  grpc?: { catalog?: string[] }
  env?: Record<string, string>
  compose?: Compose
  routes: Route[]
  async_agents?: AsyncAgent[]
}

export interface CORS {
  enabled: boolean
  allow_origins?: string[]
  allow_methods?: string[]
  allow_headers?: string[]
  expose_headers?: string[]
  allow_credentials?: boolean
  max_age?: string
}

export interface Telemetry {
  logging?: { level?: string; stdout?: boolean }
  metrics?: { enabled: boolean; listen_address?: string }
  usage?: { enabled: boolean }
}

export interface Compose {
  enabled?: boolean
  image?: string
  mock_backend?: boolean
  expose_metrics?: boolean
}

export interface Route {
  path: string
  method: string
  timeout?: string
  output_encoding?: string
  headers?: { forward?: string[] }
  query_strings?: { forward?: string[] }
  auth?: { jwt?: JWTAuth }
  rate_limit?: { max_rate: number; every?: string }
  websocket?: WebSocketConfig
  backend: Backend
}

export interface JWTAuth {
  alg: string
  jwk_url: string
  audience?: string[]
  roles?: string[]
  roles_key?: string
  issuer?: string
  disable_jwk_security?: boolean
  cache?: boolean
}

export interface WebSocketConfig {
  direct_communication?: boolean
  max_message_size?: number
  input_headers?: string[]
  connect_event?: boolean
  disconnect_event?: boolean
}

export interface Backend {
  type: string
  host: string
  path: string
  method?: string
  encoding?: string
  disable_host_sanitize?: boolean
  topic?: string
  subscription?: string
  consumer_group?: string
  input_mapping?: Record<string, string>
  graphql_type?: string
  query_path?: string
  operation_name?: string
  graphql_variables?: Record<string, unknown>
  soap_template?: string
  target?: string
  mapping?: Record<string, string>
  deny?: string[]
  kafka_cluster?: { brokers: string[]; client_id?: string }
}

export interface AsyncAgent {
  name: string
  consumer: { topic: string; workers?: number; timeout?: string }
  backend: { host: string; path: string; method?: string }
  connection?: { max_retries?: number; backoff_strategy?: string; health_interval?: string }
  kafka?: { brokers: string[]; group_id: string; client_id?: string }
}

export interface ValidationError {
  field: string
  message: string
}

export interface Preset {
  Name: string
  Description: string
  Path: string
}

export interface Catalog {
  version: string
  backendTypes: Array<{ id: string; label: string; description: string; fields: string[] }>
  routeTemplates: Array<{
    id: string
    label: string
    method: string
    backendType: string
    defaults: Partial<Route>
  }>
  enums: Record<string, string[]>
  commonHeaders: string[]
  corsPresets: Array<{
    id: string
    label: string
    allow_origins: string[]
    allow_methods: string[]
    allow_headers: string[]
  }>
  fields: Record<string, { label: string; help: string; type: string }>
}

export function blankProfile(): GatewayProfile {
  return {
    apiVersion: 'configurator.pucora.io/v1',
    kind: 'GatewayProfile',
    metadata: { name: 'my-gateway', description: '' },
    gateway: { port: 8080, timeout: '3s', cache_ttl: '3600s' },
    cors: {
      enabled: true,
      allow_origins: ['http://localhost:3000'],
      allow_methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
      allow_headers: ['Origin', 'Authorization', 'Content-Type', 'Accept'],
      max_age: '12h',
    },
    telemetry: {
      logging: { level: 'INFO', stdout: true },
      metrics: { enabled: true, listen_address: ':8090' },
      usage: { enabled: false },
    },
    routes: [],
    async_agents: [],
  }
}
