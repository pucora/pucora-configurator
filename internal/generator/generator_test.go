package generator_test

import (
	"encoding/json"
	"testing"

	"github.com/pucora/pucora-configurator/internal/generator"
	"github.com/pucora/pucora-configurator/internal/presets"
)

func TestGenerateRESTProxy(t *testing.T) {
	p, err := presets.Load("rest-proxy")
	if err != nil {
		t.Fatal(err)
	}
	out, err := generator.Generate(p)
	if err != nil {
		t.Fatal(err)
	}

	cfg := out.Config
	if cfg["version"] != 3 {
		t.Errorf("expected version 3, got %v", cfg["version"])
	}
	if cfg["port"] != 8080 {
		t.Errorf("expected port 8080, got %v", cfg["port"])
	}

	extra, ok := cfg["extra_config"].(map[string]any)
	if !ok {
		t.Fatal("missing extra_config")
	}
	if _, ok := extra["security/cors"]; !ok {
		t.Error("missing security/cors")
	}

	eps, ok := cfg["endpoints"].([]map[string]any)
	if !ok {
		t.Fatal("missing endpoints")
	}
	if len(eps) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(eps))
	}

	headers, ok := eps[0]["input_headers"].([]string)
	if !ok || len(headers) == 0 {
		t.Error("expected input_headers on first endpoint")
	}
}

func TestGenerateKafkaPubSub(t *testing.T) {
	p, err := presets.Load("kafka-pubsub")
	if err != nil {
		t.Fatal(err)
	}
	out, err := generator.Generate(p)
	if err != nil {
		t.Fatal(err)
	}

	if out.Env["KAFKA_BROKERS"] != "localhost:9092" {
		t.Errorf("expected KAFKA_BROKERS env, got %v", out.Env)
	}

	eps := out.Config["endpoints"].([]map[string]any)
	backends := eps[0]["backend"].([]map[string]any)
	if backends[0]["disable_host_sanitize"] != true {
		t.Error("kafka backend should have disable_host_sanitize")
	}
}

func TestGenerateWebSocket(t *testing.T) {
	p, err := presets.Load("websocket")
	if err != nil {
		t.Fatal(err)
	}
	out, err := generator.Generate(p)
	if err != nil {
		t.Fatal(err)
	}

	eps := out.Config["endpoints"].([]map[string]any)
	epExtra := eps[0]["extra_config"].(map[string]any)
	if _, ok := epExtra["websocket"]; !ok {
		t.Error("missing websocket extra_config")
	}
}

func TestGenerateGraphQL(t *testing.T) {
	p, err := presets.Load("graphql")
	if err != nil {
		t.Fatal(err)
	}
	out, err := generator.Generate(p)
	if err != nil {
		t.Fatal(err)
	}

	eps := out.Config["endpoints"].([]map[string]any)
	backends := eps[0]["backend"].([]map[string]any)
	gql := backends[0]["extra_config"].(map[string]any)["backend/graphql"].(map[string]any)
	if gql["operationName"] != "Hero" {
		t.Errorf("expected operationName Hero, got %v", gql["operationName"])
	}
	if gql["variables"] == nil {
		t.Error("expected graphql variables")
	}
}

func TestGenerateAsyncKafka(t *testing.T) {
	p, err := presets.Load("async-kafka")
	if err != nil {
		t.Fatal(err)
	}
	out, err := generator.Generate(p)
	if err != nil {
		t.Fatal(err)
	}

	agents := out.Config["async_agent"].([]map[string]any)
	if len(agents) != 1 {
		t.Fatalf("expected 1 async agent, got %d", len(agents))
	}
	conn := agents[0]["connection"].(map[string]any)
	if conn["max_retries"] != 3 {
		t.Errorf("expected max_retries 3, got %v", conn["max_retries"])
	}
}

func TestGenerateSOAP(t *testing.T) {
	p, err := presets.Load("soap")
	if err != nil {
		t.Fatal(err)
	}
	out, err := generator.Generate(p)
	if err != nil {
		t.Fatal(err)
	}

	eps := out.Config["endpoints"].([]map[string]any)
	backends := eps[0]["backend"].([]map[string]any)
	if backends[0]["deny"] == nil {
		t.Error("expected deny field on soap backend")
	}
}

func TestOutputIsValidJSON(t *testing.T) {
	list, err := presets.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, preset := range list {
		t.Run(preset.Name, func(t *testing.T) {
			p, err := presets.Load(preset.Name)
			if err != nil {
				t.Fatal(err)
			}
			out, err := generator.Generate(p)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(out.Config)
			if err != nil {
				t.Fatalf("invalid JSON for preset %s: %v", preset.Name, err)
			}
			if len(data) < 50 {
				t.Fatalf("config too short for preset %s", preset.Name)
			}
		})
	}
}
