package importer

import (
	"fmt"
	"strings"

	"github.com/pucora/velonetics-configurator/internal/profile"
)

// Result holds a best-effort profile conversion from pucora.json.
type Result struct {
	Profile  profile.Profile
	Warnings []string
}

// FromPucoraJSON maps a pucora.json document into a simplified GatewayProfile.
func FromPucoraJSON(cfg map[string]any) (*Result, error) {
	if cfg == nil {
		return nil, fmt.Errorf("empty config")
	}

	res := &Result{
		Profile: profile.Profile{
			APIVersion: "configurator.pucora.io/v1",
			Kind:       "GatewayProfile",
		},
	}

	if name, ok := cfg["name"].(string); ok {
		res.Profile.Metadata.Name = name
	}
	if port, ok := asInt(cfg["port"]); ok {
		res.Profile.Gateway.Port = port
	}
	if t, ok := cfg["timeout"].(string); ok {
		res.Profile.Gateway.Timeout = t
	}
	if t, ok := cfg["cache_ttl"].(string); ok {
		res.Profile.Gateway.CacheTTL = t
	}
	if t, ok := cfg["write_timeout"].(string); ok {
		res.Profile.Gateway.WriteTimeout = t
	}

	if ec, ok := cfg["extra_config"].(map[string]any); ok {
		importServiceExtra(ec, &res.Profile, &res.Warnings)
	}

	if eps := asMapSlice(cfg["endpoints"]); len(eps) > 0 {
		for _, ep := range eps {
			route, warns := importEndpoint(ep)
			res.Warnings = append(res.Warnings, warns...)
			res.Profile.Routes = append(res.Profile.Routes, route)
		}
	}

	if agents := asMapSlice(cfg["async_agent"]); len(agents) > 0 {
		for _, a := range agents {
			agent, warns := importAsyncAgent(a)
			res.Warnings = append(res.Warnings, warns...)
			if agent != nil {
				res.Profile.AsyncAgents = append(res.Profile.AsyncAgents, *agent)
			}
		}
	}

	profile.ApplyDefaults(&res.Profile)
	return res, nil
}

func importServiceExtra(ec map[string]any, p *profile.Profile, warnings *[]string) {
	if cors, ok := ec["security/cors"].(map[string]any); ok {
		p.CORS = &profile.CORS{Enabled: true}
		p.CORS.AllowOrigins = asStringSlice(cors["allow_origins"])
		p.CORS.AllowMethods = asStringSlice(cors["allow_methods"])
		p.CORS.AllowHeaders = asStringSlice(cors["allow_headers"])
		p.CORS.ExposeHeaders = asStringSlice(cors["expose_headers"])
		if v, ok := cors["allow_credentials"].(bool); ok {
			p.CORS.AllowCredentials = v
		}
		if v, ok := cors["max_age"].(string); ok {
			p.CORS.MaxAge = v
		}
	}

	p.Telemetry = &profile.Telemetry{
		Logging: &profile.Logging{Stdout: true},
		Metrics: &profile.Metrics{Enabled: true},
		Usage:   &profile.Usage{},
	}

	if log, ok := ec["telemetry/logging"].(map[string]any); ok {
		if v, ok := log["level"].(string); ok {
			p.Telemetry.Logging.Level = v
		}
		if v, ok := log["stdout"].(bool); ok {
			p.Telemetry.Logging.Stdout = v
		}
	}
	if met, ok := ec["telemetry/metrics"].(map[string]any); ok {
		if v, ok := met["enabled"].(bool); ok {
			p.Telemetry.Metrics.Enabled = v
		}
		if v, ok := met["listen_address"].(string); ok {
			p.Telemetry.Metrics.ListenAddress = v
		}
	}
	if usage, ok := ec["telemetry/usage"].(map[string]any); ok {
		if v, ok := usage["enabled"].(bool); ok {
			p.Telemetry.Usage.Enabled = v
		}
	}

	if grpc, ok := ec["grpc"].(map[string]any); ok {
		catalog := asStringSlice(grpc["catalog"])
		if len(catalog) > 0 {
			p.GRPC = &profile.GRPC{Catalog: catalog}
		}
	}

	for k := range ec {
		switch k {
		case "security/cors", "telemetry/logging", "telemetry/metrics", "telemetry/usage", "grpc":
		default:
			*warnings = append(*warnings, fmt.Sprintf("service extra_config %q not mapped to profile", k))
		}
	}
}

func importEndpoint(ep map[string]any) (profile.Route, []string) {
	var warns []string
	r := profile.Route{
		Path:   asString(ep["endpoint"]),
		Method: asString(ep["method"]),
	}
	if t, ok := ep["timeout"].(string); ok {
		r.Timeout = t
	}
	if enc, ok := ep["output_encoding"].(string); ok {
		r.OutputEncoding = enc
	}
	if headers := asStringSlice(ep["input_headers"]); len(headers) > 0 {
		r.Headers = &profile.Headers{Forward: headers}
	}
	if qs := asStringSlice(ep["input_query_strings"]); len(qs) > 0 {
		r.QueryStrings = &profile.QueryStrings{Forward: qs}
	}

	if epExtra, ok := ep["extra_config"].(map[string]any); ok {
		if jwt, ok := epExtra["auth/validator"].(map[string]any); ok {
			r.Auth = &profile.Auth{JWT: &profile.JWTAuth{
				Alg:                asString(jwt["alg"]),
				JWKURL:             asString(jwt["jwk_url"]),
				Audience:           asStringSlice(jwt["audience"]),
				Roles:              asStringSlice(jwt["roles"]),
				RolesKey:           asString(jwt["roles_key"]),
				Issuer:             asString(jwt["issuer"]),
				DisableJWKSecurity: asBool(jwt["disable_jwk_security"]),
				Cache:              asBool(jwt["cache"]),
			}}
		}
		if rl, ok := epExtra["qos/ratelimit/router"].(map[string]any); ok {
			r.RateLimit = &profile.RateLimit{
				MaxRate: asIntDefault(rl["max_rate"], 100),
			}
			if every, ok := rl["every"].(string); ok {
				r.RateLimit.Every = every
			}
		}
		if ws, ok := epExtra["websocket"].(map[string]any); ok {
			r.WebSocket = &profile.WebSocket{
				DirectCommunication: asBool(ws["enable_direct_communication"]),
				MaxMessageSize:      asIntDefault(ws["max_message_size"], 0),
				InputHeaders:        asStringSlice(ws["input_headers"]),
				ConnectEvent:        asBool(ws["connect_event"]),
				DisconnectEvent:     asBool(ws["disconnect_event"]),
			}
		}
		for k := range epExtra {
			switch k {
			case "auth/validator", "qos/ratelimit/router", "websocket":
			default:
				warns = append(warns, fmt.Sprintf("endpoint extra_config %q not mapped", k))
			}
		}
	}

	backends := asMapSlice(ep["backend"])
	if len(backends) == 0 {
		warns = append(warns, "endpoint missing backend")
		return r, warns
	}
	be := backends[0]
	r.Backend = importBackend(be, &warns)
	return r, warns
}

func importBackend(be map[string]any, warns *[]string) profile.Backend {
	b := profile.Backend{
		Path: asString(be["url_pattern"]),
	}
	if hosts := asStringSlice(be["host"]); len(hosts) > 0 {
		b.Host = hosts[0]
	}
	if m, ok := be["method"].(string); ok {
		b.Method = m
	}
	if enc, ok := be["encoding"].(string); ok {
		b.Encoding = enc
	}
	if v, ok := be["disable_host_sanitize"].(bool); ok {
		b.DisableHostSanitize = v
	}
	if target, ok := be["target"].(string); ok {
		b.Target = target
	}
	if mapping, ok := be["mapping"].(map[string]any); ok {
		b.Mapping = stringMap(mapping)
	}
	b.Deny = asStringSlice(be["deny"])

	host := b.Host
	switch {
	case strings.HasPrefix(host, "kafka://") || host == "kafka://":
		b.Type = "kafka"
		b.Host = inferBrokerHost(host, "localhost:9092")
		importPubSubExtra(be, &b)
	case strings.HasPrefix(host, "nats://") || host == "nats://":
		b.Type = "nats"
		b.Host = inferBrokerHost(host, "nats://localhost:4222")
		importPubSubExtra(be, &b)
	case strings.HasPrefix(host, "rabbit://") || host == "rabbit://":
		b.Type = "rabbit"
		b.Host = inferBrokerHost(host, "amqp://guest:guest@localhost:5672/")
		importPubSubExtra(be, &b)
	case strings.HasPrefix(host, "gcppubsub://"):
		b.Type = "gcp"
		importPubSubExtra(be, &b)
	case strings.HasPrefix(host, "azuresb://"):
		b.Type = "azure"
		importPubSubExtra(be, &b)
	case strings.HasPrefix(host, "awssns://"):
		b.Type = "aws_sns"
	case strings.HasPrefix(host, "awssqs://"):
		b.Type = "aws_sqs"
	default:
		if ec, ok := be["extra_config"].(map[string]any); ok {
			if _, ok := ec["backend/grpc"]; ok {
				b.Type = "grpc"
				b.Host = stripScheme(b.Host)
				if grpc, ok := ec["backend/grpc"].(map[string]any); ok {
					if im, ok := grpc["input_mapping"].(map[string]any); ok {
						b.InputMapping = stringMap(im)
					}
				}
				return b
			}
			if gql, ok := ec["backend/graphql"].(map[string]any); ok {
				b.Type = "graphql"
				b.GraphQLType = asString(gql["type"])
				b.QueryPath = asString(gql["query_path"])
				b.OperationName = asString(gql["operationName"])
				if vars, ok := gql["variables"].(map[string]any); ok {
					b.GraphQLVariables = vars
				}
				return b
			}
			if soap, ok := ec["backend/soap"].(map[string]any); ok {
				b.Type = "soap"
				b.SoapTemplate = asString(soap["path"])
				return b
			}
			importPubSubExtra(be, &b)
		}
		if b.Type == "" {
			if strings.HasPrefix(host, "ws://") || strings.HasPrefix(host, "wss://") {
				b.Type = "websocket"
			} else {
				b.Type = "http"
			}
		}
	}

	if ec, ok := be["extra_config"].(map[string]any); ok && b.Type != "grpc" && b.Type != "graphql" && b.Type != "soap" {
		for k := range ec {
			if !strings.HasPrefix(k, "backend/pubsub") {
				*warns = append(*warns, fmt.Sprintf("backend extra_config %q not mapped", k))
			}
		}
	}
	return b
}

func importPubSubExtra(be map[string]any, b *profile.Backend) {
	ec, ok := be["extra_config"].(map[string]any)
	if !ok {
		return
	}
	if pub, ok := ec["backend/pubsub/publisher"].(map[string]any); ok {
		if topic, ok := pub["topic_url"].(string); ok {
			b.Topic = topic
		}
	}
	if sub, ok := ec["backend/pubsub/subscriber"].(map[string]any); ok {
		if subURL, ok := sub["subscription_url"].(string); ok {
			if b.Type == "kafka" && strings.Contains(subURL, "?topic=") {
				parts := strings.SplitN(subURL, "?topic=", 2)
				b.ConsumerGroup = parts[0]
				b.Subscription = parts[1]
			} else {
				b.Subscription = subURL
			}
		}
	}
	if b.Type == "" {
		if b.Topic != "" {
			b.Type = "kafka"
		} else if b.Subscription != "" {
			b.Type = "kafka"
		}
	}
}

func importAsyncAgent(a map[string]any) (*profile.AsyncAgent, []string) {
	var warns []string
	name := asString(a["name"])
	if name == "" {
		return nil, []string{"async agent missing name"}
	}
	agent := &profile.AsyncAgent{Name: name}

	if consumer, ok := a["consumer"].(map[string]any); ok {
		agent.Consumer.Topic = asString(consumer["topic"])
		agent.Consumer.Workers = asIntDefault(consumer["workers"], 0)
		agent.Consumer.Timeout = asString(consumer["timeout"])
	}

	if backends := asMapSlice(a["backend"]); len(backends) > 0 {
		be := backends[0]
		hosts := asStringSlice(be["host"])
		if len(hosts) > 0 {
			agent.Backend.Host = hosts[0]
		}
		agent.Backend.Path = asString(be["url_pattern"])
		agent.Backend.Method = asString(be["method"])
	}

	if conn, ok := a["connection"].(map[string]any); ok {
		agent.Connection = &profile.AsyncConnection{
			MaxRetries:      asIntDefault(conn["max_retries"], 0),
			BackoffStrategy: asString(conn["backoff_strategy"]),
			HealthInterval:  asString(conn["health_interval"]),
		}
	}

	if ec, ok := a["extra_config"].(map[string]any); ok {
		if kafka, ok := ec["async/kafka"].(map[string]any); ok {
			agent.Kafka = &profile.AsyncKafka{}
			if cluster, ok := kafka["cluster"].(map[string]any); ok {
				agent.Kafka.Brokers = asStringSlice(cluster["brokers"])
				agent.Kafka.ClientID = asString(cluster["client_id"])
			}
			if group, ok := kafka["group"].(map[string]any); ok {
				agent.Kafka.GroupID = asString(group["group_id"])
			}
		} else {
			warns = append(warns, "async agent driver not kafka — not fully mapped")
		}
	}

	return agent, warns
}

func inferBrokerHost(host, fallback string) string {
	if host == "" || strings.HasSuffix(host, "://") {
		return fallback
	}
	return strings.TrimPrefix(host, "kafka://")
}

func stripScheme(host string) string {
	for _, p := range []string{"http://", "https://", "grpc://"} {
		host = strings.TrimPrefix(host, p)
	}
	return host
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}

func asIntDefault(v any, def int) int {
	if n, ok := asInt(v); ok {
		return n
	}
	return def
}

func asStringSlice(v any) []string {
	switch arr := v.(type) {
	case []string:
		return arr
	case []any:
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func stringMap(m map[string]any) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func asMapSlice(v any) []map[string]any {
	switch arr := v.(type) {
	case []map[string]any:
		return arr
	case []any:
		out := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}
