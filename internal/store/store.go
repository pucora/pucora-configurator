package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultName = "default"

// Bundle is the persisted gateway configuration.
type Bundle struct {
	Name           string            `json:"name"`
	UpdatedAt      time.Time         `json:"updated_at"`
	ProfileYAML    string            `json:"profile_yaml,omitempty"`
	PucoraJSON map[string]any    `json:"velonetics_json"`
	Env            map[string]string `json:"env,omitempty"`
	ComposeYAML    string            `json:"compose_yaml,omitempty"`
}

type Store struct {
	dir string
	mu  sync.RWMutex
}

func New(dir string) (*Store, error) {
	if dir == "" {
		dir = "./data"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) configDir(name string) string {
	if name == "" {
		name = DefaultName
	}
	return filepath.Join(s.dir, name)
}

func (s *Store) Save(bundle Bundle) error {
	if bundle.Name == "" {
		bundle.Name = DefaultName
	}
	bundle.UpdatedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	dir := s.configDir(bundle.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	if err := writeJSON(filepath.Join(dir, "pucora.json"), bundle.PucoraJSON); err != nil {
		return err
	}
	if bundle.ProfileYAML != "" {
		if err := os.WriteFile(filepath.Join(dir, "profile.yaml"), []byte(bundle.ProfileYAML), 0o644); err != nil {
			return err
		}
	}
	if len(bundle.Env) > 0 {
		if err := writeJSON(filepath.Join(dir, "env.json"), bundle.Env); err != nil {
			return err
		}
	}
	if bundle.ComposeYAML != "" {
		if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(bundle.ComposeYAML), 0o644); err != nil {
			return err
		}
	}

	meta := map[string]any{
		"name":       bundle.Name,
		"updated_at": bundle.UpdatedAt.Format(time.RFC3339),
	}
	return writeJSON(filepath.Join(dir, "meta.json"), meta)
}

func (s *Store) Load(name string) (*Bundle, error) {
	if name == "" {
		name = DefaultName
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := s.configDir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("config %q not found", name)
	}

	var bundle Bundle
	bundle.Name = name

	veloPath := filepath.Join(dir, "pucora.json")
	data, err := os.ReadFile(veloPath)
	if err != nil {
		return nil, fmt.Errorf("read pucora.json: %w", err)
	}
	if err := json.Unmarshal(data, &bundle.PucoraJSON); err != nil {
		return nil, fmt.Errorf("parse pucora.json: %w", err)
	}

	if yamlData, err := os.ReadFile(filepath.Join(dir, "profile.yaml")); err == nil {
		bundle.ProfileYAML = string(yamlData)
	}
	if envData, err := os.ReadFile(filepath.Join(dir, "env.json")); err == nil {
		_ = json.Unmarshal(envData, &bundle.Env)
	}
	if composeData, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml")); err == nil {
		bundle.ComposeYAML = string(composeData)
	}
	if metaData, err := os.ReadFile(filepath.Join(dir, "meta.json")); err == nil {
		var meta struct {
			UpdatedAt string `json:"updated_at"`
		}
		if json.Unmarshal(metaData, &meta) == nil && meta.UpdatedAt != "" {
			bundle.UpdatedAt, _ = time.Parse(time.RFC3339, meta.UpdatedAt)
		}
	}

	return &bundle, nil
}

func (s *Store) LoadPucoraJSON(name string) ([]byte, error) {
	if name == "" {
		name = DefaultName
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	path := filepath.Join(s.configDir(name), "pucora.json")
	return os.ReadFile(path)
}

func (s *Store) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(s.dir, e.Name(), "pucora.json")); err == nil {
				names = append(names, e.Name())
			}
		}
	}
	return names, nil
}

func (s *Store) SaveFromProfileYAML(name, profileYAML string, pucora map[string]any, env map[string]string, composeYAML string) error {
	return s.Save(Bundle{
		Name:           name,
		ProfileYAML:    profileYAML,
		PucoraJSON: pucora,
		Env:            env,
		ComposeYAML:    composeYAML,
	})
}

func ParseProfileYAML(yamlStr string, p any) error {
	return yaml.Unmarshal([]byte(yamlStr), p)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
