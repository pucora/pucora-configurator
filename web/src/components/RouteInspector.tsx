import type { Catalog, Route } from '../types/profile'
import { TagInput } from './TagInput'
import { KeyValueEditor } from './KeyValueEditor'
import { JsonField } from './JsonField'
import { useProfileStore } from '../store/profileStore'
import { buildCurlExample } from '../lib/advisories'

interface RouteInspectorProps {
  route: Route | null
  index: number | null
  catalog: Catalog | null
}

function showField(backendType: string, field: string): boolean {
  const pubsub = ['kafka', 'nats', 'rabbit', 'gcp', 'azure']
  switch (field) {
    case 'topic':
    case 'subscription':
    case 'consumer_group':
      return pubsub.includes(backendType)
    case 'input_mapping':
      return backendType === 'grpc'
    case 'graphql_type':
    case 'query_path':
    case 'operation_name':
    case 'graphql_variables':
      return backendType === 'graphql'
    case 'soap_template':
    case 'target':
    case 'mapping':
    case 'deny':
      return backendType === 'soap'
    case 'encoding':
      return backendType === 'http' || backendType === 'graphql'
    case 'backend_method':
      return backendType === 'http' || backendType === 'graphql' || backendType === 'soap'
    default:
      return true
  }
}

const outputEncodings = ['json', 'no-op', 'xml', 'string']

export function RouteInspector({ route, index, catalog }: RouteInspectorProps) {
  const { updateProfile, validationErrors, profile } = useProfileStore()

  if (!route || index === null) {
    return (
      <div className="p-6 text-center text-slate-500 text-sm">
        Select a route to edit its settings
      </div>
    )
  }

  const fieldError = (suffix: string) =>
    validationErrors.find((e) => e.field === `routes[${index}].${suffix}`)?.message

  const update = (fn: (r: Route) => void) => {
    updateProfile((p) => {
      fn(p.routes[index])
      return p
    })
  }

  const bt = route.backend.type
  const methods = catalog?.enums.httpMethods || ['GET', 'POST', 'PUT', 'DELETE']
  const jwtAlgs = catalog?.enums.jwtAlgorithms || ['RS256', 'HS256']
  const commonHeaders = catalog?.commonHeaders || []

  const jwtOn = !!route.auth?.jwt
  const forwardsAuth = route.headers?.forward?.includes('Authorization')
  const wsNeedsAuthHeader = bt === 'websocket' && jwtOn && !route.websocket?.input_headers?.includes('Authorization')

  return (
    <div className="p-4 space-y-5 overflow-y-auto max-h-[calc(100vh-8rem)]">
      <h2 className="text-lg font-semibold text-white">Route settings</h2>

      <div className="space-y-1.5">
        <label>HTTP method</label>
        <select value={route.method} onChange={(e) => update((r) => { r.method = e.target.value })}>
          {methods.map((m) => (
            <option key={m} value={m}>{m}</option>
          ))}
        </select>
      </div>

      <div className="space-y-1.5">
        <label>Endpoint path</label>
        <input
          value={route.path}
          onChange={(e) => update((r) => { r.path = e.target.value })}
          placeholder="/api/{id}"
        />
        {fieldError('path') && <p className="text-xs text-red-400">{fieldError('path')}</p>}
        <p className="text-xs text-slate-500">Use {'{variable}'} for path parameters</p>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div className="space-y-1.5">
          <label>Route timeout</label>
          <input
            value={route.timeout || ''}
            onChange={(e) => update((r) => { r.timeout = e.target.value || undefined })}
            placeholder="30s"
          />
        </div>
        <div className="space-y-1.5">
          <label>Output encoding</label>
          <select
            value={route.output_encoding || 'json'}
            onChange={(e) => update((r) => { r.output_encoding = e.target.value })}
          >
            {outputEncodings.map((enc) => (
              <option key={enc} value={enc}>{enc}</option>
            ))}
          </select>
          <p className="text-xs text-slate-600 normal-case">no-op for SSE/streaming</p>
        </div>
      </div>

      <div className="space-y-1.5">
        <label>Backend type</label>
        <select
          value={bt}
          onChange={(e) => update((r) => { r.backend.type = e.target.value })}
        >
          {(catalog?.backendTypes || []).map((b) => (
            <option key={b.id} value={b.id}>{b.label}</option>
          ))}
        </select>
      </div>

      {showField(bt, 'host') && (
        <div className="space-y-1.5">
          <label>Backend host</label>
          <input
            value={route.backend.host}
            onChange={(e) => update((r) => { r.backend.host = e.target.value })}
            placeholder={
              bt === 'websocket' ? 'ws://localhost:8081'
                : bt === 'grpc' ? 'localhost:4242'
                  : 'http://localhost:8000'
            }
          />
          {fieldError('backend.host') && <p className="text-xs text-red-400">{fieldError('backend.host')}</p>}
        </div>
      )}

      <div className="space-y-1.5">
        <label>{bt === 'grpc' ? 'RPC path' : 'Backend path'}</label>
        <input
          value={route.backend.path}
          onChange={(e) => update((r) => { r.backend.path = e.target.value })}
          placeholder={bt === 'grpc' ? '/package.Service/Method' : '/'}
        />
        {fieldError('backend.path') && <p className="text-xs text-red-400">{fieldError('backend.path')}</p>}
      </div>

      {showField(bt, 'backend_method') && (
        <div className="space-y-1.5">
          <label>Backend HTTP method</label>
          <select
            value={route.backend.method || route.method}
            onChange={(e) => update((r) => { r.backend.method = e.target.value })}
          >
            {methods.map((m) => <option key={m} value={m}>{m}</option>)}
          </select>
        </div>
      )}

      {showField(bt, 'encoding') && (
        <div className="space-y-1.5">
          <label>Backend encoding</label>
          <select
            value={route.backend.encoding || 'json'}
            onChange={(e) => update((r) => { r.backend.encoding = e.target.value })}
          >
            {outputEncodings.map((enc) => (
              <option key={enc} value={enc}>{enc}</option>
            ))}
          </select>
        </div>
      )}

      {showField(bt, 'topic') && (
        <div className="space-y-1.5">
          <label>Topic (publish)</label>
          <input
            value={route.backend.topic || ''}
            onChange={(e) => update((r) => { r.backend.topic = e.target.value })}
            placeholder="events"
          />
        </div>
      )}

      {showField(bt, 'subscription') && (
        <>
          <div className="space-y-1.5">
            <label>Subscription (consume)</label>
            <input
              value={route.backend.subscription || ''}
              onChange={(e) => update((r) => { r.backend.subscription = e.target.value })}
              placeholder="events"
            />
          </div>
          {bt === 'kafka' && (
            <div className="space-y-1.5">
              <label>Consumer group</label>
              <input
                value={route.backend.consumer_group || ''}
                onChange={(e) => update((r) => { r.backend.consumer_group = e.target.value })}
                placeholder="my-group"
              />
            </div>
          )}
        </>
      )}

      {showField(bt, 'input_mapping') && (
        <KeyValueEditor
          label="gRPC input mapping"
          help="Map query/path params to gRPC message fields (e.g. lat → where.latitude)"
          value={route.backend.input_mapping || {}}
          onChange={(v) => update((r) => { r.backend.input_mapping = v })}
          keyPlaceholder="query param"
          valuePlaceholder="proto.field.path"
        />
      )}

      {showField(bt, 'graphql_type') && (
        <>
          <div className="space-y-1.5">
            <label>GraphQL type</label>
            <select
              value={route.backend.graphql_type || 'query'}
              onChange={(e) => update((r) => { r.backend.graphql_type = e.target.value })}
            >
              <option value="query">query</option>
              <option value="mutation">mutation</option>
            </select>
          </div>
          <div className="space-y-1.5">
            <label>Query file path</label>
            <input
              value={route.backend.query_path || ''}
              onChange={(e) => update((r) => { r.backend.query_path = e.target.value })}
              placeholder="/etc/velonetics/graphql/queries/hero.graphql"
            />
          </div>
          <div className="space-y-1.5">
            <label>Operation name</label>
            <input
              value={route.backend.operation_name || ''}
              onChange={(e) => update((r) => { r.backend.operation_name = e.target.value || undefined })}
              placeholder="Hero"
            />
          </div>
          <JsonField
            label="GraphQL variables"
            help="JSON object; use {param} for route placeholders"
            value={route.backend.graphql_variables}
            onChange={(v) => update((r) => { r.backend.graphql_variables = v })}
          />
        </>
      )}

      {showField(bt, 'soap_template') && (
        <>
          <div className="space-y-1.5">
            <label>SOAP template path</label>
            <input
              value={route.backend.soap_template || ''}
              onChange={(e) => update((r) => { r.backend.soap_template = e.target.value })}
              placeholder="/etc/velonetics/soap/request.xml"
            />
          </div>
          <div className="space-y-1.5">
            <label>Response target (JSON path)</label>
            <input
              value={route.backend.target || ''}
              onChange={(e) => update((r) => { r.backend.target = e.target.value || undefined })}
              placeholder="Envelope.Body.CountryFlagResponse"
            />
          </div>
          <KeyValueEditor
            label="Field mapping"
            help="Rename response fields in output JSON"
            value={route.backend.mapping || {}}
            onChange={(v) => update((r) => { r.backend.mapping = v })}
          />
          <TagInput
            label="Deny fields"
            value={route.backend.deny || []}
            onChange={(v) => update((r) => { r.backend.deny = v })}
            suggestions={['-*']}
          />
        </>
      )}

      {bt === 'websocket' && (
        <div className="border border-slate-800 rounded-lg p-3 space-y-3">
          <h3 className="text-sm font-semibold text-slate-300">WebSocket options</h3>
          <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
            <input
              type="checkbox"
              checked={route.websocket?.direct_communication ?? true}
              onChange={(e) => update((r) => {
                r.websocket = { ...r.websocket, direct_communication: e.target.checked }
              })}
            />
            Direct communication
          </label>
          <div className="space-y-1.5">
            <label>Max message size (bytes)</label>
            <input
              type="number"
              value={route.websocket?.max_message_size ?? 4096}
              onChange={(e) => update((r) => {
                r.websocket = { ...r.websocket, max_message_size: parseInt(e.target.value, 10) }
              })}
            />
          </div>
          <TagInput
            label="WebSocket input headers"
            help="Headers forwarded on upgrade (e.g. Authorization for JWT)"
            value={route.websocket?.input_headers || []}
            onChange={(v) => update((r) => {
              r.websocket = { ...r.websocket, input_headers: v }
            })}
            suggestions={['Authorization']}
          />
          <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
            <input
              type="checkbox"
              checked={route.websocket?.connect_event ?? false}
              onChange={(e) => update((r) => {
                r.websocket = { ...r.websocket, connect_event: e.target.checked }
              })}
            />
            Emit connect events
          </label>
          <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
            <input
              type="checkbox"
              checked={route.websocket?.disconnect_event ?? false}
              onChange={(e) => update((r) => {
                r.websocket = { ...r.websocket, disconnect_event: e.target.checked }
              })}
            />
            Emit disconnect events
          </label>
        </div>
      )}

      {(bt === 'http' && route.output_encoding === 'no-op') && (
        <p className="text-xs text-amber-400/90 bg-amber-950/30 p-2 rounded">
          Streaming route: set gateway write timeout to 0s in the left sidebar.
        </p>
      )}

      <TagInput
        label="Headers to forward"
        help="Client headers allowed to reach the backend"
        value={route.headers?.forward || []}
        onChange={(v) => update((r) => { r.headers = { forward: v } })}
        suggestions={commonHeaders}
      />

      <TagInput
        label="Query params to forward"
        value={route.query_strings?.forward || []}
        onChange={(v) => update((r) => { r.query_strings = { forward: v } })}
        suggestions={['*', 'page', 'limit', 'query', 'operationName', 'variables']}
      />

      <div className="border-t border-slate-800 pt-4 space-y-3">
        <h3 className="text-sm font-semibold text-slate-300">Permissions</h3>

        <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
          <input
            type="checkbox"
            checked={jwtOn}
            onChange={(e) => update((r) => {
              if (e.target.checked) {
                r.auth = { jwt: { alg: 'RS256', jwk_url: '' } }
              } else {
                r.auth = undefined
              }
            })}
          />
          Require JWT authentication
        </label>

        {jwtOn && !forwardsAuth && bt !== 'websocket' && (
          <p className="text-xs text-amber-400">Tip: forward Authorization header so JWT reaches the backend.</p>
        )}
        {wsNeedsAuthHeader && (
          <p className="text-xs text-amber-400">Tip: add Authorization to WebSocket input headers for JWT on upgrade.</p>
        )}

        {route.auth?.jwt && (
          <div className="pl-4 space-y-3 border-l-2 border-cyan-800">
            <div className="space-y-1.5">
              <label>Algorithm</label>
              <select
                value={route.auth.jwt.alg}
                onChange={(e) => update((r) => { r.auth!.jwt!.alg = e.target.value })}
              >
                {jwtAlgs.map((a) => <option key={a} value={a}>{a}</option>)}
              </select>
            </div>
            <div className="space-y-1.5">
              <label>JWK URL</label>
              <input
                value={route.auth.jwt.jwk_url}
                onChange={(e) => update((r) => { r.auth!.jwt!.jwk_url = e.target.value })}
                placeholder="https://idp.example.com/.well-known/jwks.json"
              />
              {fieldError('auth.jwt.jwk_url') && (
                <p className="text-xs text-red-400">{fieldError('auth.jwt.jwk_url')}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <label>Issuer</label>
              <input
                value={route.auth.jwt.issuer || ''}
                onChange={(e) => update((r) => { r.auth!.jwt!.issuer = e.target.value || undefined })}
              />
            </div>
            <TagInput
              label="Audience"
              value={route.auth.jwt.audience || []}
              onChange={(v) => update((r) => { r.auth!.jwt!.audience = v })}
            />
            <TagInput
              label="Required roles"
              value={route.auth.jwt.roles || []}
              onChange={(v) => update((r) => { r.auth!.jwt!.roles = v })}
              suggestions={['user', 'admin']}
            />
            <div className="space-y-1.5">
              <label>Roles key (JWT claim)</label>
              <input
                value={route.auth.jwt.roles_key || ''}
                onChange={(e) => update((r) => { r.auth!.jwt!.roles_key = e.target.value || undefined })}
                placeholder="roles"
              />
            </div>
            <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
              <input
                type="checkbox"
                checked={route.auth.jwt.disable_jwk_security ?? false}
                onChange={(e) => update((r) => { r.auth!.jwt!.disable_jwk_security = e.target.checked })}
              />
              Disable JWK TLS verification (dev only)
            </label>
            <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
              <input
                type="checkbox"
                checked={route.auth.jwt.cache ?? false}
                onChange={(e) => update((r) => { r.auth!.jwt!.cache = e.target.checked })}
              />
              Cache JWK
            </label>
          </div>
        )}

        <label className="flex items-center gap-2 normal-case text-sm text-slate-300">
          <input
            type="checkbox"
            checked={!!route.rate_limit}
            onChange={(e) => update((r) => {
              r.rate_limit = e.target.checked ? { max_rate: 100, every: '1m' } : undefined
            })}
          />
          Rate limiting
        </label>

        {route.rate_limit && (
          <div className="pl-4 grid grid-cols-2 gap-2">
            <div className="space-y-1.5">
              <label>Max rate</label>
              <input
                type="number"
                value={route.rate_limit.max_rate}
                onChange={(e) => update((r) => { r.rate_limit!.max_rate = parseInt(e.target.value, 10) })}
              />
            </div>
            <div className="space-y-1.5">
              <label>Every</label>
              <input
                value={route.rate_limit.every || '1m'}
                onChange={(e) => update((r) => { r.rate_limit!.every = e.target.value })}
              />
            </div>
          </div>
        )}
      </div>

      <div className="border-t border-slate-800 pt-4 space-y-2">
        <h3 className="text-xs font-semibold text-slate-400 uppercase">Test with curl</h3>
        <pre className="text-xs font-mono p-2 bg-slate-950 rounded text-cyan-300/80 overflow-x-auto whitespace-pre-wrap">
          {buildCurlExample(route, profile.gateway.port)}
        </pre>
      </div>
    </div>
  )
}
