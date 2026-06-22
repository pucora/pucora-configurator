package store_test

import (
	"testing"

	"github.com/pucora/pucora-configurator/internal/store"
)

func TestStoreSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	bundle := store.Bundle{
		Name:           "prod",
		ProfileYAML:    "apiVersion: configurator.pucora.in/v1\n",
		PucoraJSON: map[string]any{"version": 3, "port": 8080},
		Env:            map[string]string{"KAFKA_BROKERS": "localhost:9092"},
		ComposeYAML:    "services:\n  pucora:\n    image: test\n",
	}
	if err := s.Save(bundle); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.Load("prod")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PucoraJSON["port"] != float64(8080) {
		t.Fatalf("expected port 8080, got %v", loaded.PucoraJSON["port"])
	}
	if loaded.ProfileYAML == "" {
		t.Fatal("expected profile yaml")
	}

	raw, err := s.LoadPucoraJSON("prod")
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 10 {
		t.Fatal("expected raw json")
	}

	names, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "prod" {
		t.Fatalf("expected [prod], got %v", names)
	}
}
