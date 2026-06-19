package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pucora/velonetics-configurator/internal/profile"
)

const defaultImage = "niteesh20/pucora:2.0.0"

type Requirements struct {
	Kafka        bool
	NATS         bool
	Rabbit       bool
	MockBackend  bool
	MockWebhook  bool
	GRPCCatalog  bool
	GraphQLFiles bool
	SOAPFiles    bool
}

func Detect(p *profile.Profile) Requirements {
	var req Requirements
	for _, r := range p.Routes {
		switch r.Backend.Type {
		case "kafka":
			req.Kafka = true
		case "nats":
			req.NATS = true
		case "rabbit":
			req.Rabbit = true
		case "grpc":
			req.GRPCCatalog = true
		case "graphql":
			req.GraphQLFiles = true
			if isLocalHost(r.Backend.Host) {
				req.MockBackend = true
			}
		case "soap":
			req.SOAPFiles = true
			if isLocalHost(r.Backend.Host) {
				req.MockBackend = true
			}
		case "http", "websocket":
			if isLocalHost(r.Backend.Host) {
				req.MockBackend = true
			}
		}
	}
	for _, a := range p.AsyncAgents {
		if a.Kafka != nil {
			req.Kafka = true
		}
		if isLocalHost(a.Backend.Host) {
			req.MockWebhook = true
		}
	}
	if p.GRPC != nil && len(p.GRPC.Catalog) > 0 {
		req.GRPCCatalog = true
	}
	return req
}

func ShouldGenerate(p *profile.Profile, flag bool) bool {
	if flag {
		return true
	}
	if p.Compose != nil && p.Compose.Enabled != nil {
		return *p.Compose.Enabled
	}
	return false
}

func Write(outputDir string, p *profile.Profile, env map[string]string) error {
	content := Render(p, env)
	path := filepath.Join(outputDir, "docker-compose.yml")
	return os.WriteFile(path, []byte(content), 0o644)
}

func WriteToString(p *profile.Profile, env map[string]string) (string, error) {
	return Render(p, env), nil
}

func Render(p *profile.Profile, env map[string]string) string {
	req := Detect(p)
	image := defaultImage
	mockBackend := req.MockBackend || req.MockWebhook
	exposeMetrics := false

	if p.Compose != nil {
		if p.Compose.Image != "" {
			image = p.Compose.Image
		}
		if p.Compose.MockBackend != nil {
			mockBackend = *p.Compose.MockBackend
		}
		exposeMetrics = p.Compose.ExposeMetrics
	}

	var sb strings.Builder
	sb.WriteString("services:\n")

	if req.Kafka {
		sb.WriteString(kafkaService())
	}
	if req.NATS {
		sb.WriteString(natsService())
	}
	if req.Rabbit {
		sb.WriteString(rabbitService())
	}
	if mockBackend && req.MockBackend {
		sb.WriteString(mockBackendService())
	}
	if mockBackend && req.MockWebhook {
		sb.WriteString(mockWebhookService())
	}

	sb.WriteString(veloneticsService(image, p, env, req, exposeMetrics))
	return sb.String()
}

func veloneticsService(image string, p *profile.Profile, env map[string]string, req Requirements, exposeMetrics bool) string {
	port := p.Gateway.Port
	if port == 0 {
		port = 8080
	}

	var sb strings.Builder
	sb.WriteString("  pucora:\n")
	sb.WriteString(fmt.Sprintf("    image: %s\n", image))
	sb.WriteString("    ports:\n")
	sb.WriteString(fmt.Sprintf("      - \"%d:%d\"\n", port, port))
	if exposeMetrics && p.Telemetry != nil && p.Telemetry.Metrics != nil && p.Telemetry.Metrics.Enabled {
		sb.WriteString("      - \"8090:8090\"\n")
	}
	sb.WriteString("    environment:\n")
	for k, v := range env {
		sb.WriteString(fmt.Sprintf("      %s: %s\n", k, v))
	}
	if req.Kafka && env["KAFKA_BROKERS"] == "" {
		sb.WriteString("      KAFKA_BROKERS: redpanda:9092\n")
	}
	if req.NATS && env["NATS_SERVER_URL"] == "" {
		sb.WriteString("      NATS_SERVER_URL: nats://nats:4222\n")
	}
	if req.Rabbit && env["RABBIT_SERVER_URL"] == "" {
		sb.WriteString("      RABBIT_SERVER_URL: guest:guest@rabbitmq:5672\n")
	}
	sb.WriteString("    volumes:\n")
	sb.WriteString("      - ./pucora.json:/etc/pucora/pucora.json:ro\n")
	if req.GRPCCatalog && p.GRPC != nil {
		for _, cat := range p.GRPC.Catalog {
			if strings.HasPrefix(cat, "./") {
				sb.WriteString(fmt.Sprintf("      - %s:%s:ro\n", cat, dockerPath(cat)))
			}
		}
	}
	if req.GraphQLFiles {
		sb.WriteString("      - ./graphql:/etc/pucora/graphql:ro\n")
	}
	if req.SOAPFiles {
		sb.WriteString("      - ./soap:/etc/pucora/soap:ro\n")
	}
	sb.WriteString(fmt.Sprintf("    command: [\"run\", \"-c\", \"/etc/pucora/pucora.json\"]\n"))

	var deps []string
	if req.Kafka {
		deps = append(deps, "redpanda")
	}
	if req.NATS {
		deps = append(deps, "nats")
	}
	if req.Rabbit {
		deps = append(deps, "rabbitmq")
	}
	if len(deps) > 0 {
		sb.WriteString("    depends_on:\n")
		for _, d := range deps {
			if d == "redpanda" {
				sb.WriteString("      redpanda:\n        condition: service_healthy\n")
			} else {
				sb.WriteString(fmt.Sprintf("      - %s\n", d))
			}
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

func kafkaService() string {
	return `  redpanda:
    image: docker.redpanda.com/redpandadata/redpanda:v24.2.4
    command:
      - redpanda
      - start
      - --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092
      - --advertise-kafka-addr internal://redpanda:9092,external://localhost:19092
      - --mode dev-container
      - --smp 1
      - --memory 512M
    ports:
      - "19092:19092"
    healthcheck:
      test: ["CMD", "rpk", "cluster", "health"]
      interval: 5s
      timeout: 5s
      retries: 20
      start_period: 10s

`
}

func natsService() string {
	return `  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"

`
}

func rabbitService() string {
	return `  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"

`
}

func mockBackendService() string {
	return `  mock-backend:
    build: ./services/mock-backend
    ports:
      - "8000:8000"
      - "8081:8081"
      - "4000:4000"
      - "4242:4242"

`
}

func mockWebhookService() string {
	return `  mock-webhook:
    build: ./services/mock-webhook
    ports:
      - "8081:8081"

`
}

func isLocalHost(host string) bool {
	lower := strings.ToLower(host)
	return strings.Contains(lower, "localhost") ||
		strings.Contains(lower, "127.0.0.1") ||
		strings.Contains(lower, "mock-backend") ||
		strings.Contains(lower, "mock-webhook")
}

func dockerPath(p string) string {
	if strings.HasPrefix(p, "./") {
		return "/etc/pucora/" + strings.TrimPrefix(p, "./")
	}
	return p
}
