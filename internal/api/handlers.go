package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pucora/pucora-configurator/internal/catalog"
	"github.com/pucora/pucora-configurator/internal/compose"
	"github.com/pucora/pucora-configurator/internal/doctor"
	"github.com/pucora/pucora-configurator/internal/generator"
	"github.com/pucora/pucora-configurator/internal/importer"
	"github.com/pucora/pucora-configurator/internal/presets"
	"github.com/pucora/pucora-configurator/internal/profile"
	"github.com/pucora/pucora-configurator/internal/store"
	"gopkg.in/yaml.v3"
)

// Server serves the configurator REST API.
type Server struct {
	AllowedOrigins []string
	store          *store.Store
	APIKey         string
	PublicBaseURL  string
}

// NewServer creates an API server with optional config store.
func NewServer(origins []string, st *store.Store, apiKey, publicBaseURL string) *Server {
	return &Server{
		AllowedOrigins: origins,
		store:          st,
		APIKey:         apiKey,
		PublicBaseURL:  publicBaseURL,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/presets", s.handlePresets)
	mux.HandleFunc("/api/presets/", s.handlePresetByName)
	mux.HandleFunc("/api/catalog", s.handleCatalog)
	mux.HandleFunc("/api/validate", s.handleValidate)
	mux.HandleFunc("/api/generate", s.handleGenerate)
	mux.HandleFunc("/api/import", s.handleImport)
	mux.HandleFunc("/api/import-json", s.handleImportJSON)
	mux.HandleFunc("/api/doctor", s.handleDoctor)
	mux.HandleFunc("/api/configs", s.handleConfigList)
	mux.HandleFunc("/api/config/", s.handleConfig)
	mux.HandleFunc("/api/config", s.handleConfig)
	return s.cors(mux)
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && s.originAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) originAllowed(origin string) bool {
	if len(s.AllowedOrigins) == 0 {
		return true
	}
	for _, o := range s.AllowedOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handlePresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	list, err := presets.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handlePresetByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/presets/")
	if name == "" {
		writeError(w, http.StatusBadRequest, "preset name required")
		return
	}
	p, err := presets.Load(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	cat, err := catalog.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

type profileRequest struct {
	Profile profile.Profile `json:"profile"`
	Compose bool            `json:"compose"`
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req profileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	profile.ApplyDefaults(&req.Profile)
	errs := profile.ValidateStructured(&req.Profile)
	writeJSON(w, http.StatusOK, map[string]any{
		"valid":  !errs.HasErrors(),
		"errors": errs,
	})
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req profileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	profile.ApplyDefaults(&req.Profile)
	if errs := profile.ValidateStructured(&req.Profile); errs.HasErrors() {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"valid":  false,
			"errors": errs,
		})
		return
	}
	out, err := generator.Generate(&req.Profile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	yamlData, _ := yaml.Marshal(&req.Profile)
	resp := map[string]any{
		"valid":           true,
		"pucora_json": out.Config,
		"profile_yaml":    string(yamlData),
		"env":             out.Env,
		"warnings":        out.Warnings,
		"advisories":      doctor.Check(&req.Profile),
	}

	if req.Compose || compose.ShouldGenerate(&req.Profile, false) {
		composeYAML := generateComposeYAML(&req.Profile, out.Env)
		resp["compose_yaml"] = composeYAML
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var body struct {
		YAML string `json:"yaml"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var p profile.Profile
	if err := yaml.Unmarshal([]byte(body.YAML), &p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid YAML: "+err.Error())
		return
	}
	profile.ApplyDefaults(&p)
	errs := profile.ValidateStructured(&p)
	writeJSON(w, http.StatusOK, map[string]any{
		"profile": p,
		"valid":   !errs.HasErrors(),
		"errors":  errs,
	})
}

func (s *Server) handleImportJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var body struct {
		Config map[string]any `json:"config"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	res, err := importer.FromPucoraJSON(body.Config)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	errs := profile.ValidateStructured(&res.Profile)
	writeJSON(w, http.StatusOK, map[string]any{
		"profile":  res.Profile,
		"valid":    !errs.HasErrors(),
		"errors":   errs,
		"warnings": res.Warnings,
	})
}

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var req profileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	profile.ApplyDefaults(&req.Profile)
	writeJSON(w, http.StatusOK, map[string]any{
		"advisories": doctor.Check(&req.Profile),
	})
}

func generateComposeYAML(p *profile.Profile, env map[string]string) string {
	dir, err := compose.WriteToString(p, env)
	if err != nil {
		return ""
	}
	return dir
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
