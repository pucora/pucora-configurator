# Velonetics Configurator

A configuration tool that turns a simple YAML profile into a complete [Velonetics](https://github.com/velonetics/velonetics-ce-master) gateway setup — routes, CORS, allowed headers, JWT auth, pub/sub, gRPC, WebSockets, and more.

**No more hand-writing `extra_config` namespaces or remembering `disable_host_sanitize` rules.**

## Quick start

```bash
# Build
make build

# Interactive wizard
./bin/velonetics-config init -o my-profile.yaml

# Or generate with docker-compose for local dev
./bin/velonetics-config presets apply kafka-pubsub -g ./output --compose
docker compose -f ./output/docker-compose.yml up

# Generate from your profile
./bin/velonetics-config generate -f my-profile.yaml -o ./output

# Validate before generating
./bin/velonetics-config validate -f my-profile.yaml
```

Run the gateway:

```bash
velonetics check -c ./output/velonetics.json
velonetics run -c ./output/velonetics.json
```

## Why this exists

Configuring Velonetics directly requires knowing:

- Which `extra_config` namespace to use (`security/cors`, `backend/pubsub/publisher`, `backend/grpc`, …)
- That headers are **deny-by-default** — you must set `input_headers` per route
- Non-HTTP backends need `disable_host_sanitize: true` and scheme-specific `host` values (`kafka://`, `ws://`, …)
- gRPC needs a service-level catalog plus per-backend `input_mapping`
- WebSockets need four separate settings across endpoint and backend

This tool hides that complexity behind a **simple profile format** and **ready-made presets**.

## Profile format

Profiles use `apiVersion: configurator.velonetics.io/v1` and a flat, readable structure:

```yaml
apiVersion: configurator.velonetics.io/v1
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

# Generate velonetics.json + docker-compose for local dev
velonetics-config presets apply kafka-pubsub -g ./output --compose
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Interactive wizard — choose pattern, answer prompts, get a profile |
| `generate -f profile.yaml -o ./output` | Produce `velonetics.json` and `.env` |
| `generate --compose` | Also produce `docker-compose.yml` for local dev |
| `validate -f profile.yaml` | Check profile without generating |
| `presets list` | Show built-in presets |
| `presets apply <name>` | Copy preset and/or generate config |

## What gets generated

| Output | Contents |
|--------|----------|
| `velonetics.json` | Full gateway config with correct namespaces |
| `.env` | Broker connection strings (`KAFKA_BROKERS`, `NATS_SERVER_URL`, …) |
| `docker-compose.yml` | Local dev stack (Velonetics + Redpanda/NATS/RabbitMQ + mocks) |
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
  image: niteesh20/velonetics:2.0.0
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
├── cmd/velonetics-config/     # CLI entrypoint
├── internal/
│   ├── profile/               # YAML profile schema + validation
│   ├── generator/             # Profile → velonetics.json
│   ├── compose/               # docker-compose.yml generator
│   ├── presets/               # Embedded preset profiles
│   ├── wizard/                # Interactive setup
│   └── cli/                   # Cobra commands
├── profiles/                  # Example profiles (source for presets)
└── Makefile
```

## Development

```bash
make test
make build

# After editing profiles/*.yaml, sync embedded presets:
make presets
```

## License

Apache 2.0 — same as Velonetics Community Edition.
