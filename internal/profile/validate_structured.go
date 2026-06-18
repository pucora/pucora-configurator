package profile

import (
	"fmt"
	"strings"
)

func ValidateStructured(p *Profile) ValidationErrors {
	var errs ValidationErrors

	if p.Gateway.Port < 1 || p.Gateway.Port > 65535 {
		errs = append(errs, ValidationError{
			Field:   "gateway.port",
			Message: "must be between 1 and 65535",
		})
	}
	if len(p.Routes) == 0 && len(p.AsyncAgents) == 0 {
		errs = append(errs, ValidationError{
			Field:   "routes",
			Message: "at least one route or async agent is required",
		})
	}

	seen := make(map[string]bool)
	for i, r := range p.Routes {
		prefix := fmt.Sprintf("routes[%d]", i)
		method := strings.ToUpper(r.Method)
		if !validMethods[method] {
			errs = append(errs, ValidationError{
				Field:   prefix + ".method",
				Message: fmt.Sprintf("invalid method %q", r.Method),
			})
		}
		key := method + " " + r.Path
		if seen[key] {
			errs = append(errs, ValidationError{
				Field:   prefix + ".path",
				Message: fmt.Sprintf("duplicate route %s %s", method, r.Path),
			})
		}
		seen[key] = true

		if r.Path == "" || !strings.HasPrefix(r.Path, "/") {
			errs = append(errs, ValidationError{
				Field:   prefix + ".path",
				Message: "must start with /",
			})
		}
		errs = append(errs, validateBackendStructured(prefix, &r.Backend)...)
		if r.Auth != nil && r.Auth.JWT != nil && r.Auth.JWT.JWKURL == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".auth.jwt.jwk_url",
				Message: "required when JWT is enabled",
			})
		}
	}

	for i, a := range p.AsyncAgents {
		prefix := fmt.Sprintf("async_agents[%d]", i)
		if a.Name == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".name",
				Message: "is required",
			})
		}
		if a.Kafka == nil || len(a.Kafka.Brokers) == 0 {
			errs = append(errs, ValidationError{
				Field:   prefix + ".kafka.brokers",
				Message: "is required",
			})
		}
		if a.Consumer.Topic == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".consumer.topic",
				Message: "is required",
			})
		}
	}

	return errs
}

func validateBackendStructured(prefix string, b *Backend) ValidationErrors {
	var errs ValidationErrors
	bp := prefix + ".backend"

	if !validBackendTypes[b.Type] {
		errs = append(errs, ValidationError{
			Field:   bp + ".type",
			Message: fmt.Sprintf("unknown backend type %q", b.Type),
		})
		return errs
	}
	if b.Host == "" && b.Type != "kafka" && b.Type != "nats" && b.Type != "rabbit" && b.Type != "gcp" && b.Type != "azure" {
		errs = append(errs, ValidationError{
			Field:   bp + ".host",
			Message: "is required",
		})
	}

	switch b.Type {
	case "kafka", "nats", "rabbit", "gcp", "azure":
		if b.Topic == "" && b.Subscription == "" {
			errs = append(errs, ValidationError{
				Field:   bp + ".topic",
				Message: fmt.Sprintf("topic or subscription is required for %s", b.Type),
			})
		}
	case "grpc":
		if b.Path == "" || !strings.Contains(b.Path, "/") {
			errs = append(errs, ValidationError{
				Field:   bp + ".path",
				Message: "must be the full RPC path (e.g. /package.Service/Method)",
			})
		}
	case "websocket":
		if !strings.HasPrefix(b.Host, "ws://") && !strings.HasPrefix(b.Host, "wss://") {
			errs = append(errs, ValidationError{
				Field:   bp + ".host",
				Message: "must use ws:// or wss://",
			})
		}
	case "graphql":
		if b.GraphQLType == "" {
			errs = append(errs, ValidationError{
				Field:   bp + ".graphql_type",
				Message: "is required (query or mutation)",
			})
		}
		if b.QueryPath == "" {
			errs = append(errs, ValidationError{
				Field:   bp + ".query_path",
				Message: "is required for graphql",
			})
		}
	case "soap":
		if b.SoapTemplate == "" {
			errs = append(errs, ValidationError{
				Field:   bp + ".soap_template",
				Message: "is required for soap",
			})
		}
	}

	return errs
}
