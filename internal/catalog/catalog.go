package catalog

import (
	"embed"
	"encoding/json"
)

//go:embed catalog.json
var catalogFS embed.FS

func Load() (map[string]any, error) {
	data, err := catalogFS.ReadFile("catalog.json")
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
