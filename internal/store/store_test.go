package store_test

import (
	"testing"

	"github.com/velonetics/velonetics-configurator/internal/store"
)

func TestStoreSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s, err := store.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	bundle := store.Bundle{
		Name:           "prod",
		ProfileYAML:    "apiVersion: configurator.velonetics.io/v1\n",
		VeloneticsJSON: map[string]any{"version": 3, "port": 8080},
		Env:            map[string]string{"KAFKA_BROKERS": "localhost:9092"},
		ComposeYAML:    "services:\n  velonetics:\n    image: test\n",
	}
	if err := s.Save(bundle); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.Load("prod")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.VeloneticsJSON["port"] != float64(8080) {
		t.Fatalf("expected port 8080, got %v", loaded.VeloneticsJSON["port"])
	}
	if loaded.ProfileYAML == "" {
		t.Fatal("expected profile yaml")
	}

	raw, err := s.LoadVeloneticsJSON("prod")
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
