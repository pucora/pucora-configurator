package api

import (
	"net/http"
	"strings"

	"github.com/pucora/velonetics-configurator/internal/profile"
	"github.com/pucora/velonetics-configurator/internal/store"
)

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "config store not configured")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/config")
	path = strings.Trim(path, "/")

	// GET /api/config/{name}/pucora.json — raw gateway config
	if strings.HasSuffix(path, "/pucora.json") {
		name := strings.TrimSuffix(path, "/pucora.json")
		if name == "" {
			name = store.DefaultName
		}
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		s.handleConfigPucoraJSON(w, r, name)
		return
	}

	// /api/config or /api/config/{name}
	name := path
	if name == "" {
		name = r.URL.Query().Get("name")
	}
	if name == "" {
		name = store.DefaultName
	}

	switch r.Method {
	case http.MethodGet:
		s.handleConfigGet(w, r, name)
	case http.MethodPost:
		s.handleConfigPost(w, r, name)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleConfigList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "config store not configured")
		return
	}
	names, err := s.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"configs": names})
}

func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.requireAPIKey(r); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	format := r.URL.Query().Get("format")
	if format == "pucora" || format == "pucora.json" {
		s.handleConfigPucoraJSON(w, r, name)
		return
	}

	bundle, err := s.store.Load(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if format == "yaml" || format == "profile" {
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bundle.ProfileYAML))
		return
	}

	writeJSON(w, http.StatusOK, bundle)
}

func (s *Server) handleConfigPucoraJSON(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.requireAPIKey(r); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	data, err := s.store.LoadPucoraJSON(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleConfigPost(w http.ResponseWriter, r *http.Request, name string) {
	if err := s.requireAPIKey(r); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req configSaveRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Name != "" {
		name = req.Name
	}

	bundle, status, err := s.buildBundleFromRequest(name, req)
	if err != nil {
		writeJSON(w, status, map[string]any{
			"valid":  false,
			"errors": err,
		})
		return
	}

	if err := s.store.Save(*bundle); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"saved":    true,
		"name":     bundle.Name,
		"get_url":  s.configURLs(bundle.Name),
		"updated_at": bundle.UpdatedAt,
	})
}

type configSaveRequest struct {
	Name          string                `json:"name"`
	Profile       *profile.Profile      `json:"profile,omitempty"`
	ProfileYAML   string                `json:"profile_yaml,omitempty"`
	PucoraJSON map[string]any       `json:"velonetics_json,omitempty"`
	Compose       bool                  `json:"compose"`
}
