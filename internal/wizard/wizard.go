package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/velonetics/velonetics-configurator/internal/profile"
)

type Wizard struct {
	reader *bufio.Reader
}

func New() *Wizard {
	return &Wizard{reader: bufio.NewReader(os.Stdin)}
}

func (w *Wizard) Run() (*profile.Profile, error) {
	fmt.Println("Velonetics Configurator — Interactive Setup")
	fmt.Println("===========================================")
	fmt.Println()

	p := &profile.Profile{}

	p.Metadata.Name = w.prompt("Gateway name", "my-gateway")
	p.Gateway.Port = w.promptInt("Gateway port", 8080)

	if w.promptYesNo("Enable CORS?", true) {
		p.CORS = &profile.CORS{Enabled: true}
		origins := w.prompt("Allowed origins (comma-separated)", "http://localhost:3000")
		p.CORS.AllowOrigins = splitCSV(origins)
		headers := w.prompt("Allowed headers (comma-separated)", "Origin, Authorization, Content-Type")
		p.CORS.AllowHeaders = splitCSV(headers)
	}

	fmt.Println()
	fmt.Println("Choose a connectivity pattern:")
	fmt.Println("  1) REST HTTP proxy")
	fmt.Println("  2) REST with JWT auth")
	fmt.Println("  3) Kafka pub/sub")
	fmt.Println("  4) NATS pub/sub")
	fmt.Println("  5) WebSocket proxy")
	fmt.Println("  6) gRPC client")
	fmt.Println("  7) SSE streaming")
	choice := w.promptInt("Selection", 1)

	switch choice {
	case 1:
		w.configureREST(p, false)
	case 2:
		w.configureREST(p, true)
	case 3:
		w.configureKafka(p)
	case 4:
		w.configureNATS(p)
	case 5:
		w.configureWebSocket(p)
	case 6:
		w.configureGRPC(p)
	case 7:
		w.configureSSE(p)
	default:
		return nil, fmt.Errorf("invalid selection %d", choice)
	}

	profile.ApplyDefaults(p)
	if err := profile.Validate(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (w *Wizard) configureREST(p *profile.Profile, withAuth bool) {
	backendHost := w.prompt("Backend host URL", "http://localhost:8000")
	pathPrefix := w.prompt("API path prefix exposed by gateway", "/api")

	route := profile.Route{
		Path:   pathPrefix + "/{path}",
		Method: "GET",
		Headers: &profile.Headers{
			Forward: []string{"Authorization", "Content-Type", "Accept"},
		},
		QueryStrings: &profile.QueryStrings{Forward: []string{"*"}},
		Backend: profile.Backend{
			Type: "http",
			Host: backendHost,
			Path: "/{path}",
		},
	}

	if withAuth {
		route.Auth = &profile.Auth{
			JWT: &profile.JWTAuth{
				Alg:    w.prompt("JWT algorithm", "RS256"),
				JWKURL: w.prompt("JWK URL", "https://your-idp.example.com/.well-known/jwks.json"),
				Audience: splitCSV(w.prompt("JWT audience (comma-separated)", "https://api.example.com")),
			},
		}
	}

	p.Routes = []profile.Route{route}

	if w.promptYesNo("Add POST route with same backend?", true) {
		post := route
		post.Method = "POST"
		p.Routes = append(p.Routes, post)
	}
}

func (w *Wizard) configureKafka(p *profile.Profile) {
	brokers := w.prompt("Kafka brokers", "localhost:9092")
	topic := w.prompt("Topic name", "events")
	group := w.prompt("Consumer group", "my-group")

	p.Env = map[string]string{"KAFKA_BROKERS": brokers}
	p.Routes = []profile.Route{
		{
			Path: "/events", Method: "POST",
			Headers: &profile.Headers{Forward: []string{"Content-Type"}},
			Backend: profile.Backend{Type: "kafka", Host: brokers, Path: "/ignored", Topic: topic},
		},
		{
			Path: "/events", Method: "GET",
			Backend: profile.Backend{
				Type: "kafka", Host: brokers, Path: "/ignored",
				Subscription: topic, ConsumerGroup: group,
			},
		},
	}
}

func (w *Wizard) configureNATS(p *profile.Profile) {
	server := w.prompt("NATS server URL", "nats://localhost:4222")
	subject := w.prompt("Subject name", "events")

	p.Env = map[string]string{"NATS_SERVER_URL": server}
	p.Routes = []profile.Route{
		{
			Path: "/publish", Method: "POST",
			Backend: profile.Backend{Type: "nats", Host: server, Path: "/ignored", Topic: subject},
		},
		{
			Path: "/subscribe", Method: "GET",
			Backend: profile.Backend{Type: "nats", Host: server, Path: "/ignored", Subscription: subject},
		},
	}
}

func (w *Wizard) configureWebSocket(p *profile.Profile) {
	wsHost := w.prompt("WebSocket backend (ws:// or wss://)", "ws://localhost:8081")
	path := w.prompt("Gateway WebSocket path", "/ws/echo")
	backendPath := w.prompt("Backend WebSocket path", "/echo")

	p.Routes = []profile.Route{{
		Path: path, Method: "GET",
		WebSocket: &profile.WebSocket{DirectCommunication: true, MaxMessageSize: 4096},
		Backend:   profile.Backend{Type: "websocket", Host: wsHost, Path: backendPath},
	}}
}

func (w *Wizard) configureGRPC(p *profile.Profile) {
	catalog := w.prompt("gRPC catalog .pb file path", "./grpc/catalog.pb")
	grpcHost := w.prompt("gRPC backend host:port", "localhost:4242")
	rpcPath := w.prompt("Full RPC path (e.g. /package.Service/Method)", "/flight_finder.Flights/FindFlight")

	p.GRPC = &profile.GRPC{Catalog: []string{catalog}}
	p.Routes = []profile.Route{{
		Path: "/grpc", Method: "GET",
		QueryStrings: &profile.QueryStrings{Forward: []string{"*"}},
		Backend: profile.Backend{Type: "grpc", Host: grpcHost, Path: rpcPath},
	}}
}

func (w *Wizard) configureSSE(p *profile.Profile) {
	backendHost := w.prompt("Streaming backend URL", "http://localhost:8000")
	path := w.prompt("SSE endpoint path", "/events")

	p.Gateway.WriteTimeout = "0s"
	p.Routes = []profile.Route{{
		Path: path, Method: "GET", Timeout: "30s", OutputEncoding: "no-op",
		Backend: profile.Backend{Type: "http", Host: backendHost, Path: path, Encoding: "no-op"},
	}}
}

func (w *Wizard) prompt(label, defaultVal string) string {
	fmt.Printf("%s [%s]: ", label, defaultVal)
	line, _ := w.reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

func (w *Wizard) promptInt(label string, defaultVal int) int {
	for {
		s := w.prompt(label, strconv.Itoa(defaultVal))
		n, err := strconv.Atoi(s)
		if err == nil {
			return n
		}
		fmt.Println("  Please enter a number.")
	}
}

func (w *Wizard) promptYesNo(label string, defaultYes bool) bool {
	def := "y"
	if !defaultYes {
		def = "n"
	}
	for {
		s := strings.ToLower(w.prompt(label+" (y/n)", def))
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
