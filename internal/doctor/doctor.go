package doctor

import (
	"fmt"
	"strings"

	"github.com/pucora/velonetics-configurator/internal/profile"
)

// Advisory is a non-fatal configuration recommendation.
type Advisory struct {
	Field   string `json:"field"`
	Level   string `json:"level"` // warn, info
	Message string `json:"message"`
}

// Check returns advisory warnings for a profile beyond strict validation.
func Check(p *profile.Profile) []Advisory {
	var out []Advisory

	hasGRPCRoute := false
	for i, r := range p.Routes {
		prefix := fmt.Sprintf("routes[%d]", i)

		if r.Backend.Type == "grpc" {
			hasGRPCRoute = true
		}

		if r.Auth != nil && r.Auth.JWT != nil {
			jwt := r.Auth.JWT
			if jwt.JWKURL == "" {
				continue
			}
			forwardsAuth := r.Headers != nil && contains(r.Headers.Forward, "Authorization")
			if !forwardsAuth {
				out = append(out, Advisory{
					Field:   prefix + ".headers.forward",
					Level:   "warn",
					Message: "JWT is enabled but Authorization is not forwarded to the backend",
				})
			}
			if jwt.DisableJWKSecurity {
				out = append(out, Advisory{
					Field:   prefix + ".auth.jwt.disable_jwk_security",
					Level:   "warn",
					Message: "disable_jwk_security allows insecure JWK fetching — use only in development",
				})
			}
		}

		if r.Backend.Type == "websocket" && r.Auth != nil && r.Auth.JWT != nil {
			wsHeaders := []string{}
			if r.WebSocket != nil {
				wsHeaders = r.WebSocket.InputHeaders
			}
			if !contains(wsHeaders, "Authorization") {
				out = append(out, Advisory{
					Field:   prefix + ".websocket.input_headers",
					Level:   "warn",
					Message: "WebSocket JWT auth requires Authorization in websocket input_headers",
				})
			}
		}

		isStreaming := r.OutputEncoding == "no-op" || r.Backend.Encoding == "no-op"
		if isStreaming && p.Gateway.WriteTimeout != "0s" {
			out = append(out, Advisory{
				Field:   "gateway.write_timeout",
				Level:   "warn",
				Message: fmt.Sprintf("route %s %s uses streaming/no-op encoding — set gateway.write_timeout to 0s", r.Method, r.Path),
			})
		}

		if (r.Backend.Type == "http" || r.Backend.Type == "grpc") &&
			(r.Headers == nil || len(r.Headers.Forward) == 0) {
			out = append(out, Advisory{
				Field:   prefix + ".headers.forward",
				Level:   "info",
				Message: "No input_headers configured — client headers will not reach the backend",
			})
		}
	}

	if hasGRPCRoute && (p.GRPC == nil || len(p.GRPC.Catalog) == 0) {
		out = append(out, Advisory{
			Field:   "grpc.catalog",
			Level:   "warn",
			Message: "gRPC routes exist but grpc.catalog is empty",
		})
	}

	for i, a := range p.AsyncAgents {
		if a.Kafka == nil || len(a.Kafka.Brokers) == 0 {
			out = append(out, Advisory{
				Field:   fmt.Sprintf("async_agents[%d].kafka.brokers", i),
				Level:   "warn",
				Message: "async agent has no Kafka brokers configured",
			})
		}
	}

	return out
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, val) {
			return true
		}
	}
	return false
}
