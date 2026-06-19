package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pucora/pucora-configurator/internal/compose"
	"github.com/pucora/pucora-configurator/internal/profile"
)

type Output struct {
	Config   map[string]any
	Env      map[string]string
	Warnings []string
}

func Generate(p *profile.Profile) (*Output, error) {
	out := &Output{
		Config: make(map[string]any),
		Env:    make(map[string]string),
	}

	cfg := out.Config
	cfg["$schema"] = profile.SchemaVersion
	cfg["version"] = 3
	cfg["name"] = p.Metadata.Name
	cfg["port"] = p.Gateway.Port
	cfg["timeout"] = p.Gateway.Timeout
	cfg["cache_ttl"] = p.Gateway.CacheTTL
	if p.Gateway.WriteTimeout != "" {
		cfg["write_timeout"] = p.Gateway.WriteTimeout
	}

	extraConfig := buildServiceExtraConfig(p, out)
	if len(extraConfig) > 0 {
		cfg["extra_config"] = extraConfig
	}

	endpoints := make([]map[string]any, 0, len(p.Routes))
	for _, route := range p.Routes {
		ep, warns := buildEndpoint(route)
		out.Warnings = append(out.Warnings, warns...)
		endpoints = append(endpoints, ep)
	}
	cfg["endpoints"] = endpoints

	if len(p.AsyncAgents) > 0 {
		cfg["async_agent"] = buildAsyncAgents(p.AsyncAgents)
	}

	for k, v := range p.Env {
		out.Env[k] = v
	}
	mergeBrokerEnv(p, out.Env)

	return out, nil
}

func Write(outputDir string, out *Output, p *profile.Profile, withCompose bool) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	configPath := filepath.Join(outputDir, "pucora.json")
	data, err := json.MarshalIndent(out.Config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if len(out.Env) > 0 {
		envPath := filepath.Join(outputDir, ".env")
		var sb strings.Builder
		for k, v := range out.Env {
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			sb.WriteString("\n")
		}
		if err := os.WriteFile(envPath, []byte(sb.String()), 0o644); err != nil {
			return fmt.Errorf("write .env: %w", err)
		}
	}

	if compose.ShouldGenerate(p, withCompose) {
		if err := compose.Write(outputDir, p, out.Env); err != nil {
			return fmt.Errorf("write docker-compose: %w", err)
		}
		req := compose.Detect(p)
		if err := compose.WriteScaffold(outputDir, req); err != nil {
			return fmt.Errorf("write compose scaffold: %w", err)
		}
	}

	return nil
}

func buildServiceExtraConfig(p *profile.Profile, out *Output) map[string]any {
	ec := make(map[string]any)

	if p.CORS != nil && p.CORS.Enabled {
		cors := map[string]any{}
		if len(p.CORS.AllowOrigins) > 0 {
			cors["allow_origins"] = p.CORS.AllowOrigins
		}
		if len(p.CORS.AllowMethods) > 0 {
			cors["allow_methods"] = p.CORS.AllowMethods
		}
		if len(p.CORS.AllowHeaders) > 0 {
			cors["allow_headers"] = p.CORS.AllowHeaders
		}
		if len(p.CORS.ExposeHeaders) > 0 {
			cors["expose_headers"] = p.CORS.ExposeHeaders
		}
		if p.CORS.AllowCredentials {
			cors["allow_credentials"] = true
		}
		if p.CORS.MaxAge != "" {
			cors["max_age"] = p.CORS.MaxAge
		}
		ec["security/cors"] = cors
	}

	if p.Telemetry != nil {
		if p.Telemetry.Logging != nil {
			ec["telemetry/logging"] = map[string]any{
				"level":  p.Telemetry.Logging.Level,
				"prefix": "[PUCORA]",
				"stdout": p.Telemetry.Logging.Stdout,
			}
		}
		if p.Telemetry.Metrics != nil && p.Telemetry.Metrics.Enabled {
			ec["telemetry/metrics"] = map[string]any{
				"listen_address": p.Telemetry.Metrics.ListenAddress,
			}
		}
		if p.Telemetry.Usage != nil {
			ec["telemetry/usage"] = map[string]any{
				"enabled": p.Telemetry.Usage.Enabled,
			}
		}
	}

	if p.GRPC != nil && len(p.GRPC.Catalog) > 0 {
		ec["grpc"] = map[string]any{
			"catalog": p.GRPC.Catalog,
		}
	}

	return ec
}

func buildEndpoint(r profile.Route) (map[string]any, []string) {
	var warnings []string

	ep := map[string]any{
		"endpoint": r.Path,
		"method":   strings.ToUpper(r.Method),
		"backend":  []map[string]any{buildBackend(r.Backend, &warnings)},
	}

	if r.Timeout != "" {
		ep["timeout"] = r.Timeout
	}
	if r.OutputEncoding != "" {
		ep["output_encoding"] = r.OutputEncoding
	}
	if r.Headers != nil && len(r.Headers.Forward) > 0 {
		ep["input_headers"] = r.Headers.Forward
	} else if r.Backend.Type == "http" || r.Backend.Type == "grpc" {
		warnings = append(warnings, fmt.Sprintf("route %s %s: no input_headers configured — client headers will not reach the backend", r.Method, r.Path))
	}
	if r.QueryStrings != nil && len(r.QueryStrings.Forward) > 0 {
		ep["input_query_strings"] = r.QueryStrings.Forward
	}

	epExtra := make(map[string]any)

	if r.Auth != nil && r.Auth.JWT != nil {
		jwt := r.Auth.JWT
		jwtCfg := map[string]any{
			"alg":     jwt.Alg,
			"jwk_url": jwt.JWKURL,
		}
		if len(jwt.Audience) > 0 {
			jwtCfg["audience"] = jwt.Audience
		}
		if len(jwt.Roles) > 0 {
			jwtCfg["roles"] = jwt.Roles
		}
		if jwt.RolesKey != "" {
			jwtCfg["roles_key"] = jwt.RolesKey
		}
		if jwt.Issuer != "" {
			jwtCfg["issuer"] = jwt.Issuer
		}
		if jwt.DisableJWKSecurity {
			jwtCfg["disable_jwk_security"] = true
		}
		if jwt.Cache {
			jwtCfg["cache"] = true
		}
		epExtra["auth/validator"] = jwtCfg
	}

	if r.RateLimit != nil {
		rl := map[string]any{"max_rate": r.RateLimit.MaxRate}
		if r.RateLimit.Every != "" {
			rl["every"] = r.RateLimit.Every
		}
		epExtra["qos/ratelimit/router"] = rl
	}

	if r.WebSocket != nil || r.Backend.Type == "websocket" {
		ws := map[string]any{}
		if r.WebSocket != nil {
			if r.WebSocket.DirectCommunication {
				ws["enable_direct_communication"] = true
			}
			if r.WebSocket.MaxMessageSize > 0 {
				ws["max_message_size"] = r.WebSocket.MaxMessageSize
			}
			if len(r.WebSocket.InputHeaders) > 0 {
				ws["input_headers"] = r.WebSocket.InputHeaders
			}
			if r.WebSocket.ConnectEvent {
				ws["connect_event"] = true
			}
			if r.WebSocket.DisconnectEvent {
				ws["disconnect_event"] = true
			}
		} else {
			ws["enable_direct_communication"] = true
			ws["max_message_size"] = 4096
		}
		epExtra["websocket"] = ws
	}

	if len(epExtra) > 0 {
		ep["extra_config"] = epExtra
	}

	return ep, warnings
}

func buildBackend(b profile.Backend, warnings *[]string) map[string]any {
	be := map[string]any{
		"url_pattern": b.Path,
	}

	switch b.Type {
	case "http", "graphql", "soap", "lambda":
		be["host"] = []string{b.Host}
		if b.Method != "" {
			be["method"] = strings.ToUpper(b.Method)
		}
		if b.Encoding != "" {
			be["encoding"] = b.Encoding
		}

	case "grpc":
		be["host"] = []string{stripScheme(b.Host)}
		beExtra := map[string]any{}
		if len(b.InputMapping) > 0 {
			beExtra["backend/grpc"] = map[string]any{
				"input_mapping": b.InputMapping,
			}
		} else {
			beExtra["backend/grpc"] = map[string]any{}
		}
		be["extra_config"] = beExtra

	case "websocket":
		be["host"] = []string{b.Host}
		be["disable_host_sanitize"] = true

	case "kafka", "nats", "rabbit", "gcp", "azure", "aws_sns", "aws_sqs":
		be["host"] = []string{pubsubHost(b)}
		be["disable_host_sanitize"] = true
		beExtra := buildPubSubExtra(b)
		be["extra_config"] = beExtra
		if b.Encoding != "" {
			be["encoding"] = b.Encoding
		}

	default:
		be["host"] = []string{b.Host}
	}

	if b.DisableHostSanitize {
		be["disable_host_sanitize"] = true
	}

	if b.Type == "graphql" {
		gql := map[string]any{
			"type":       b.GraphQLType,
			"query_path": b.QueryPath,
		}
		if b.OperationName != "" {
			gql["operationName"] = b.OperationName
		}
		if len(b.GraphQLVariables) > 0 {
			gql["variables"] = b.GraphQLVariables
		}
		be["extra_config"] = mergeExtra(be, map[string]any{
			"backend/graphql": gql,
		})
	}

	if b.Type == "soap" {
		soapCfg := map[string]any{"path": b.SoapTemplate}
		be["encoding"] = "xml"
		if b.Method == "" {
			be["method"] = "POST"
		}
		if b.Target != "" {
			be["target"] = b.Target
		}
		if len(b.Mapping) > 0 {
			be["mapping"] = b.Mapping
		}
		if len(b.Deny) > 0 {
			be["deny"] = b.Deny
		}
		be["extra_config"] = mergeExtra(be, map[string]any{
			"backend/soap": soapCfg,
		})
	}

	return be
}

func buildPubSubExtra(b profile.Backend) map[string]any {
	ec := make(map[string]any)

	if b.KafkaCluster != nil && len(b.KafkaCluster.Brokers) > 0 {
		writer := map[string]any{
			"topic": b.Topic,
			"cluster": map[string]any{
				"brokers": b.KafkaCluster.Brokers,
			},
		}
		if b.KafkaCluster.ClientID != "" {
			writer["cluster"].(map[string]any)["client_id"] = b.KafkaCluster.ClientID
		}
		if b.Topic != "" {
			ec["backend/pubsub/publisher/kafka"] = map[string]any{"writer": writer}
		}
		if b.Subscription != "" {
			reader := map[string]any{
				"topic": b.Topic,
				"cluster": map[string]any{
					"brokers": b.KafkaCluster.Brokers,
				},
			}
			ec["backend/pubsub/subscriber/kafka"] = map[string]any{"reader": reader}
		}
		return ec
	}

	if b.Topic != "" {
		ec["backend/pubsub/publisher"] = map[string]any{
			"topic_url": b.Topic,
		}
	}
	if b.Subscription != "" {
		subURL := b.Subscription
		if b.ConsumerGroup != "" && b.Type == "kafka" {
			subURL = b.ConsumerGroup + "?topic=" + b.Subscription
		}
		ec["backend/pubsub/subscriber"] = map[string]any{
			"subscription_url": subURL,
		}
	}

	return ec
}

func pubsubHost(b profile.Backend) string {
	switch b.Type {
	case "kafka":
		return "kafka://"
	case "nats":
		return "nats://"
	case "rabbit":
		return "rabbit://"
	case "gcp":
		return "gcppubsub://"
	case "azure":
		return "azuresb://"
	case "aws_sns":
		return b.Host // expects full awssns:///arn:...
	case "aws_sqs":
		return b.Host // expects awssqs://...
	default:
		return b.Host
	}
}

func buildAsyncAgents(agents []profile.AsyncAgent) []map[string]any {
	result := make([]map[string]any, 0, len(agents))
	for _, a := range agents {
		kafkaCfg := map[string]any{
			"cluster": map[string]any{
				"brokers": a.Kafka.Brokers,
			},
			"group": map[string]any{
				"group_id": a.Kafka.GroupID,
			},
		}
		if a.Kafka.ClientID != "" {
			kafkaCfg["cluster"].(map[string]any)["client_id"] = a.Kafka.ClientID
		}
		agent := map[string]any{
			"name": a.Name,
			"consumer": map[string]any{
				"topic": a.Consumer.Topic,
			},
			"backend": []map[string]any{{
				"host":        []string{a.Backend.Host},
				"url_pattern": a.Backend.Path,
				"method":      defaultStr(a.Backend.Method, "POST"),
			}},
			"extra_config": map[string]any{
				"async/kafka": kafkaCfg,
			},
		}
		if a.Connection != nil {
			conn := map[string]any{}
			if a.Connection.MaxRetries > 0 {
				conn["max_retries"] = a.Connection.MaxRetries
			}
			if a.Connection.BackoffStrategy != "" {
				conn["backoff_strategy"] = a.Connection.BackoffStrategy
			}
			if a.Connection.HealthInterval != "" {
				conn["health_interval"] = a.Connection.HealthInterval
			}
			if len(conn) > 0 {
				agent["connection"] = conn
			}
		}
		if a.Consumer.Workers > 0 {
			agent["consumer"].(map[string]any)["workers"] = a.Consumer.Workers
		}
		if a.Consumer.Timeout != "" {
			agent["consumer"].(map[string]any)["timeout"] = a.Consumer.Timeout
		}
		result = append(result, agent)
	}
	return result
}

func mergeBrokerEnv(p *profile.Profile, env map[string]string) {
	for _, r := range p.Routes {
		switch r.Backend.Type {
		case "kafka":
			if _, ok := env["KAFKA_BROKERS"]; !ok && r.Backend.Host != "" && r.Backend.Host != "kafka://" {
				env["KAFKA_BROKERS"] = r.Backend.Host
			}
		case "nats":
			if _, ok := env["NATS_SERVER_URL"]; !ok && r.Backend.Host != "" && r.Backend.Host != "nats://" {
				env["NATS_SERVER_URL"] = r.Backend.Host
			}
		case "rabbit":
			if _, ok := env["RABBIT_SERVER_URL"]; !ok && r.Backend.Host != "" {
				env["RABBIT_SERVER_URL"] = r.Backend.Host
			}
		case "azure":
			if _, ok := env["SERVICEBUS_CONNECTION_STRING"]; !ok && r.Backend.Host != "" {
				env["SERVICEBUS_CONNECTION_STRING"] = r.Backend.Host
			}
		}
	}
}

func stripScheme(host string) string {
	for _, prefix := range []string{"http://", "https://", "grpc://"} {
		if strings.HasPrefix(host, prefix) {
			return strings.TrimPrefix(host, prefix)
		}
	}
	return host
}

func mergeExtra(be map[string]any, add map[string]any) map[string]any {
	existing, _ := be["extra_config"].(map[string]any)
	if existing == nil {
		return add
	}
	for k, v := range add {
		existing[k] = v
	}
	return existing
}

func defaultStr(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}
