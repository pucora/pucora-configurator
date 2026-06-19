package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pucora/pucora-configurator/internal/compose"
	"github.com/pucora/pucora-configurator/internal/generator"
	"github.com/pucora/pucora-configurator/internal/profile"
	"github.com/pucora/pucora-configurator/internal/store"
	"gopkg.in/yaml.v3"
)

func (s *Server) buildBundleFromRequest(name string, req configSaveRequest) (*store.Bundle, int, profile.ValidationErrors) {
	if name == "" {
		name = store.DefaultName
	}

	// Raw pucora.json upload (skip profile validation)
	if req.PucoraJSON != nil && req.Profile == nil && req.ProfileYAML == "" {
		return &store.Bundle{
			Name:           name,
			PucoraJSON: req.PucoraJSON,
		}, http.StatusOK, nil
	}

	var p profile.Profile

	switch {
	case req.Profile != nil:
		p = *req.Profile
	case req.ProfileYAML != "":
		if err := yaml.Unmarshal([]byte(req.ProfileYAML), &p); err != nil {
			return nil, http.StatusBadRequest, profile.ValidationErrors{{
				Field: "profile_yaml", Message: "invalid YAML: " + err.Error(),
			}}
		}
	default:
		return nil, http.StatusBadRequest, profile.ValidationErrors{{
			Field: "profile", Message: "profile, profile_yaml, or pucora_json is required",
		}}
	}

	profile.ApplyDefaults(&p)
	if errs := profile.ValidateStructured(&p); errs.HasErrors() {
		return nil, http.StatusUnprocessableEntity, errs
	}

	out, err := generator.Generate(&p)
	if err != nil {
		return nil, http.StatusInternalServerError, profile.ValidationErrors{{
			Field: "generate", Message: err.Error(),
		}}
	}

	yamlData, _ := yaml.Marshal(&p)
	bundle := &store.Bundle{
		Name:           name,
		ProfileYAML:    string(yamlData),
		PucoraJSON: out.Config,
		Env:            out.Env,
	}

	if req.Compose || compose.ShouldGenerate(&p, false) {
		bundle.ComposeYAML = compose.Render(&p, out.Env)
	}

	return bundle, http.StatusOK, nil
}

func (s *Server) requireAPIKey(r *http.Request) error {
	if s.APIKey == "" {
		return nil
	}
	key := r.Header.Get("X-API-Key")
	if key == "" {
		key = r.URL.Query().Get("api_key")
	}
	if key != s.APIKey {
		return fmt.Errorf("invalid or missing API key")
	}
	return nil
}

func (s *Server) configURLs(name string) map[string]string {
	base := strings.TrimSuffix(s.PublicBaseURL, "/")
	if base == "" {
		base = ""
	}
	prefix := base + "/api/config"
	if name != "" && name != store.DefaultName {
		prefix += "/" + name
	}
	return map[string]string{
		"bundle":          prefix,
		"pucora_json": prefix + "/pucora.json",
		"profile_yaml":    prefix + "?format=yaml",
	}
}
