package presets

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pucora/pucora-configurator/internal/profile"
)

//go:embed profiles/*.yaml
var embeddedProfiles embed.FS

type Preset struct {
	Name        string
	Description string
	Path        string
}

func List() ([]Preset, error) {
	entries, err := fs.ReadDir(embeddedProfiles, "profiles")
	if err != nil {
		return nil, err
	}
	var presets []Preset
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		p, err := Load(name)
		if err != nil {
			return nil, err
		}
		presets = append(presets, Preset{
			Name:        name,
			Description: p.Metadata.Description,
			Path:        "profiles/" + e.Name(),
		})
	}
	return presets, nil
}

func Load(name string) (*profile.Profile, error) {
	path := "profiles/" + name + ".yaml"
	if !strings.HasSuffix(name, ".yaml") {
		path = "profiles/" + name + ".yaml"
	}
	data, err := embeddedProfiles.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("preset %q not found: %w", name, err)
	}
	var p profile.Profile
	if err := profile.UnmarshalYAML(data, &p); err != nil {
		return nil, err
	}
	profile.ApplyDefaults(&p)
	if err := profile.Validate(&p); err != nil {
		return nil, fmt.Errorf("preset %q invalid: %w", name, err)
	}
	return &p, nil
}

func LoadFromDir(dir, name string) (*profile.Profile, error) {
	path := filepath.Join(dir, name+".yaml")
	return profile.Load(path)
}
