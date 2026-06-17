package profile

import (
	"fmt"
	"strings"
)

var validBackendTypes = map[string]bool{
	"http": true, "grpc": true, "websocket": true,
	"kafka": true, "nats": true, "rabbit": true,
	"gcp": true, "aws_sns": true, "aws_sqs": true, "azure": true,
	"graphql": true, "soap": true, "lambda": true,
}

var validMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true,
}

func Validate(p *Profile) error {
	if p.Gateway.Port < 1 || p.Gateway.Port > 65535 {
		return fmt.Errorf("gateway.port must be between 1 and 65535")
	}
	if len(p.Routes) == 0 && len(p.AsyncAgents) == 0 {
		return fmt.Errorf("at least one route or async_agent is required")
	}

	seen := make(map[string]bool)
	for i, r := range p.Routes {
		method := strings.ToUpper(r.Method)
		if !validMethods[method] {
			return fmt.Errorf("routes[%d]: invalid method %q", i, r.Method)
		}
		key := method + " " + r.Path
		if seen[key] {
			return fmt.Errorf("routes[%d]: duplicate route %s %s", i, method, r.Path)
		}
		seen[key] = true

		if r.Path == "" || !strings.HasPrefix(r.Path, "/") {
			return fmt.Errorf("routes[%d]: path must start with /", i)
		}
		if err := validateBackend(i, &r.Backend); err != nil {
			return err
		}
		if r.Auth != nil && r.Auth.JWT != nil && r.Auth.JWT.JWKURL == "" {
			return fmt.Errorf("routes[%d]: auth.jwt.jwk_url is required when JWT is enabled", i)
		}
	}

	for i, a := range p.AsyncAgents {
		if a.Name == "" {
			return fmt.Errorf("async_agents[%d]: name is required", i)
		}
		if a.Kafka == nil || len(a.Kafka.Brokers) == 0 {
			return fmt.Errorf("async_agents[%d]: kafka.brokers is required", i)
		}
		if a.Consumer.Topic == "" {
			return fmt.Errorf("async_agents[%d]: consumer.topic is required", i)
		}
	}

	return nil
}

func validateBackend(routeIdx int, b *Backend) error {
	if !validBackendTypes[b.Type] {
		return fmt.Errorf("routes[%d]: unknown backend type %q", routeIdx, b.Type)
	}
	if b.Host == "" && b.Type != "kafka" && b.Type != "nats" && b.Type != "rabbit" && b.Type != "gcp" && b.Type != "azure" {
		return fmt.Errorf("routes[%d]: backend.host is required", routeIdx)
	}

	switch b.Type {
	case "kafka", "nats", "rabbit", "gcp", "azure":
		if b.Topic == "" && b.Subscription == "" {
			return fmt.Errorf("routes[%d]: backend.topic or backend.subscription is required for %s", routeIdx, b.Type)
		}
	case "grpc":
		if b.Path == "" || !strings.Contains(b.Path, "/") {
			return fmt.Errorf("routes[%d]: grpc backend.path must be the full RPC path (e.g. /package.Service/Method)", routeIdx)
		}
	case "websocket":
		if !strings.HasPrefix(b.Host, "ws://") && !strings.HasPrefix(b.Host, "wss://") {
			return fmt.Errorf("routes[%d]: websocket backend.host must use ws:// or wss://", routeIdx)
		}
	case "graphql":
		if b.GraphQLType == "" {
			return fmt.Errorf("routes[%d]: backend.graphql_type is required (query or mutation)", routeIdx)
		}
		if b.QueryPath == "" {
			return fmt.Errorf("routes[%d]: backend.query_path is required for graphql", routeIdx)
		}
	case "soap":
		if b.SoapTemplate == "" {
			return fmt.Errorf("routes[%d]: backend.soap_template is required for soap", routeIdx)
		}
	}

	return nil
}
