package compose_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pucora/pucora-configurator/internal/compose"
	"github.com/pucora/pucora-configurator/internal/presets"
)

func TestComposeKafkaPreset(t *testing.T) {
	p, err := presets.Load("kafka-pubsub")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	env := map[string]string{"KAFKA_BROKERS": "redpanda:9092"}
	if err := compose.Write(dir, p, env); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "redpanda:") {
		t.Error("expected redpanda service")
	}
	if !strings.Contains(content, "pucora:") {
		t.Error("expected pucora service")
	}
	if !strings.Contains(content, "KAFKA_BROKERS") {
		t.Error("expected KAFKA_BROKERS env")
	}
}

func TestComposeAsyncKafka(t *testing.T) {
	p, err := presets.Load("async-kafka")
	if err != nil {
		t.Fatal(err)
	}

	req := compose.Detect(p)
	if !req.Kafka {
		t.Error("expected kafka requirement")
	}
	if !req.MockWebhook {
		t.Error("expected mock webhook requirement")
	}
}

func TestComposeScaffold(t *testing.T) {
	dir := t.TempDir()
	req := compose.Requirements{MockBackend: true}
	if err := compose.WriteScaffold(dir, req); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "services", "README.md")); err != nil {
		t.Fatal("expected services README")
	}
}
