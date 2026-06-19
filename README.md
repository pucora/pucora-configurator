# Pucora Configurator

A configuration tool that turns a simple YAML profile into a complete [Pucora](https://github.com/pucora/velonetics-ce-master) gateway setup — routes, CORS, allowed headers, JWT auth, pub/sub, gRPC, WebSockets, and more.

**No more hand-writing `extra_config` namespaces or remembering `disable_host_sanitize` rules.**

## Visual UI (recommended)

A standalone web app lets you build configs with forms, drag-and-drop routes, and live preview — no YAML required.

```bash
# Terminal 1: API server
make dev-api

# Terminal 2: Web UI (http://localhost:5173)
make dev-web
```

Or run everything with Docker:

```bash
make docker-up
# Open http://localhost:3000
```

The UI provides:
- Preset gallery — one-click starting points
- Drag-and-drop route reordering
- Component palette — REST, JWT, Kafka, WebSocket, gRPC, GraphQL, SOAP, SSE, cloud pub/sub
- Guided forms with header chips and CORS presets
- JWT / rate-limit toggles per route
- Live `pucora.json` preview with ZIP download
- Import YAML or `pucora.json` (reverse import)
- Route search, duplicate, and per-route curl examples
- Advisory warnings (gRPC catalog, JWT headers, streaming timeouts)
- Autosave draft to browser localStorage
- Saved config list on publish panel (GET/POST API)

## CLI quick start

```bash
# Build
make build

# Interactive wizard
./bin/velonetics-config init -o my-profile.yaml

# Or pick a preset
./bin/velonetics-config presets list
./bin/velonetics-config presets apply kafka-pubsub -g ./output --compose
docker compose -f ./output/docker-compose.yml up

# Generate from your profile
./bin/velonetics-config generate -f my-profile.yaml -o ./output

# Validate before generating
./bin/velonetics-config validate -f my-profile.yaml
./bin/velonetics-config validate -f my-profile.yaml --json

# Import existing pucora.json back to profile YAML
./bin/velonetics-config import -f ./output/pucora.json -o profile.yaml

# Compare two profiles
./bin/velonetics-config diff --file-a profile.yaml --file-b profile-v2.yaml

# Advisory warnings (beyond strict validation)
./bin/velonetics-config doctor -f my-profile.yaml

# Watch profile and regenerate on save
./bin/velonetics-config watch -f my-profile.yaml -o ./output

# Push/pull from config store API
./bin/velonetics-config config push -f my-profile.yaml -n prod
./bin/velonetics-config config pull prod -o ./output
./bin/velonetics-config config list
```

Run the gateway:

```bash
pucora check -c ./output/pucora.json
pucora run -c ./output/pucora.json
```

## Why this exists

Configuring Pucora directly requires knowing:

- Which `extra_config` namespace to use (`security/cors`, `backend/pubsub/publisher`, `backend/grpc`, …)
- That headers are **deny-by-default** — you must set `input_headers` per route
- Non-HTTP backends need `disable_host_sanitize: true` and scheme-specific `host` values (`kafka://`, `ws://`, …)
- gRPC needs a service-level catalog plus per-backend `input_mapping`
- WebSockets need four separate settings across endpoint and backend

This tool hides that complexity behind a **simple profile format** and **ready-made presets**.

## Profile format

Profiles use `apiVersion: configurator.pucora.io/v1` and a flat, readable structure:

```yaml
apiVersion: configurator.pucora.io/v1
kind: GatewayProfile
metadata:
  name: My API Gateway

gateway:
  port: 8080
  timeout: 3s

cors:
  enabled: true
  allow_origins: [http://localhost:3000]
  allow_headers: [Origin, Authorization, Content-Type]

routes:
  - path: /api/{path}
    method: GET
    headers:
      forward: [Authorization, Content-Type]   # ← auto becomes input_headers
    query_strings:
      forward: [page, limit]
    auth:
      jwt:
        alg: RS256
        jwk_url: https://idp.example.com/.well-known/jwks.json
        audience: [https://api.example.com]
    backend:
      type: http
      host: http://localhost:8000
      path: /{path}
```

See [`profiles/`](profiles/) for full examples.

## Built-in presets

| Preset | Description |
|--------|-------------|
| `rest-proxy` | HTTP reverse proxy with CORS and header forwarding |
| `rest-with-auth` | REST proxy with JWT, rate limiting, and CORS |
| `kafka-pubsub` | Kafka publish/subscribe REST endpoints |
| `nats-pubsub` | NATS publish/subscribe |
| `grpc-client` | REST → gRPC with catalog and input mapping |
| `graphql` | GraphQL query/mutation adapter + transparent proxy |
| `soap` | SOAP/XML backend with response mapping |
| `async-kafka` | Background Kafka consumer → webhook |
| `websocket` | WebSocket proxy with optional JWT |
| `streaming-sse` | Server-sent events / streaming |

```bash
# Save preset as editable profile
velonetics-config presets apply kafka-pubsub -o profile.yaml

# Generate pucora.json + docker-compose for local dev
velonetics-config presets apply kafka-pubsub -g ./output --compose
```

## Config store API (GET / POST)

Publish and pull gateway config over HTTP so CI, gateways, or other tools can apply it without manual downloads.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/config` or `/api/config/{name}` | Save profile → generates and stores `pucora.json` |
| `GET` | `/api/config/{name}` | Full bundle (profile + pucora.json + env) |
| `GET` | `/api/config/{name}/pucora.json` | **Raw gateway config** — use this to apply |
| `GET` | `/api/config/{name}?format=yaml` | Profile YAML only |
| `GET` | `/api/configs` | List saved config names |
| `POST` | `/api/import` | Import profile YAML |
| `POST` | `/api/import-json` | Import `pucora.json` → profile |
| `POST` | `/api/doctor` | Advisory warnings for a profile |

### POST — save config

```bash
curl -X POST http://localhost:8081/api/config/prod \
  -H "Content-Type: application/json" \
  -d @- <<'EOF'
{
  "name": "prod",
  "profile": { ... },
  "compose": false
}
EOF
```

Or upload raw `pucora.json`:

```bash
curl -X POST http://localhost:8081/api/config/prod \
  -H "Content-Type: application/json" \
  -d '{"velonetics_json": {"version": 3, "port": 8080, "endpoints": []}}'
```

### GET — pull config for gateway

```bash
# Fetch pucora.json and run gateway
curl -s http://localhost:8081/api/config/prod/pucora.json -o pucora.json
pucora check -c pucora.json
pucora run -c pucora.json
```

### Environment variables

| Variable | Description |
|----------|-------------|
| `CONFIG_STORE_PATH` | Directory for persisted configs (default: `./data`) |
| `CONFIG_API_KEY` | Optional API key — send as `X-API-Key` header on GET/POST |
| `PUBLIC_BASE_URL` | Base URL returned in publish response links |

The UI builder includes **POST — Save to API** and **GET — Load from API** with the gateway pull URL displayed.

## Commands

| Command | Description |
|---------|-------------|
| `init` | Interactive wizard — choose pattern, answer prompts, get a profile |
| `init --from-preset graphql` | Start from a preset instead of the wizard |
| `init --edit` | Open profile in `$EDITOR` after creation |
| `generate -f profile.yaml -o ./output` | Produce `pucora.json` and `.env` |
| `generate --stdout` | Print `pucora.json` to stdout |
| `generate --check` | Run `pucora check` after generate (if on PATH) |
| `generate --compose` | Also produce `docker-compose.yml` for local dev |
| `validate -f profile.yaml` | Check profile without generating |
| `validate --json` | Machine-readable validation + advisories |
| `import -f pucora.json -o profile.yaml` | Reverse-import gateway JSON to profile |
| `diff --file-a a.yaml --file-b b.yaml` | Compare profiles and generated output |
| `doctor -f profile.yaml` | Show advisory warnings |
| `watch -f profile.yaml` | Regenerate when profile file changes |
| `config push/pull/list` | Sync with the config store HTTP API |
| `presets list` | Show built-in presets |
| `presets apply <name>` | Copy preset and/or generate config |
| `presets apply <name> --edit` | Save preset and open in `$EDITOR` |

## What gets generated

| Output | Contents |
|--------|----------|
| `pucora.json` | Full gateway config with correct namespaces |
| `.env` | Broker connection strings (`KAFKA_BROKERS`, `NATS_SERVER_URL`, …) |
| `docker-compose.yml` | Local dev stack (Pucora + Redpanda/NATS/RabbitMQ + mocks) |
| `services/README.md` | Instructions to copy mock backends from CE examples |

Set `compose.enabled: true` in your profile to always generate docker-compose, or pass `--compose` on the CLI.

The compose generator auto-detects required services from your profile:
- **Kafka** → Redpanda
- **NATS** → NATS server
- **Rabbit** → RabbitMQ
- **HTTP/gRPC/WebSocket/GraphQL/SOAP** with local hosts → mock-backend build context
- **Async agents** → mock-webhook build context

```yaml
compose:
  enabled: true
  image: niteesh20/pucora:2.0.0
  mock_backend: true
  expose_metrics: true
```

The generator automatically sets:

- `$schema` and `version: 3`
- `security/cors` from your CORS block
- `telemetry/logging`, `telemetry/metrics`, `telemetry/usage`
- `input_headers` and `input_query_strings` per route
- `auth/validator`, `qos/ratelimit/router`, `websocket` extra_config
- `disable_host_sanitize`, pub/sub namespaces, gRPC `input_mapping`
- `grpc.catalog` at service level when configured
- `async_agent` blocks for background Kafka consumers

## Backend types

| `backend.type` | Use for |
|----------------|---------|
| `http` | Standard REST proxy |
| `grpc` | Upstream gRPC services |
| `websocket` | WebSocket backends (`ws://` / `wss://`) |
| `kafka`, `nats`, `rabbit` | Message broker pub/sub |
| `gcp`, `azure`, `aws_sns`, `aws_sqs` | Cloud pub/sub |
| `graphql` | GraphQL adapter |
| `soap` | SOAP/XML backends |

## Project layout

```
velonetics-configurator/
├── cmd/
│   ├── velonetics-config/     # CLI entrypoint
│   └── velonetics-config-api/ # REST API for web UI
├── web/                       # React visual configurator
├── deploy/                    # Docker compose + Dockerfiles
├── internal/
│   ├── api/                   # HTTP handlers
│   ├── catalog/               # Field metadata for guided forms
│   ├── profile/               # YAML profile schema + validation
│   ├── generator/             # Profile → pucora.json
│   ├── compose/               # docker-compose.yml generator
│   ├── presets/               # Embedded preset profiles
│   ├── wizard/                # Interactive CLI setup
│   └── cli/                   # Cobra commands
├── profiles/                  # Example profiles (source for presets)
└── Makefile
```

## Development

```bash
make test
make build-all

# Web UI development
make dev-api    # API on :8081
make dev-web    # UI on :5173 (proxies /api to :8081)

# After editing profiles/*.yaml, sync embedded presets:
make presets
```

## Deployment

```bash
docker compose -f deploy/docker-compose.yml up --build
```

- **Web** — nginx on port 3000, proxies `/api` to the API service
- **API** — Go server on port 8081
- Set `ALLOWED_ORIGINS` on the API for your web domain
- Build web with `VITE_API_URL` empty when using nginx proxy (default in compose)

## License

Apache 2.0 — same as Pucora Community Edition.
