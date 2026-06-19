package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pucora/pucora-configurator/internal/api"
	"github.com/pucora/pucora-configurator/internal/profile"
	"github.com/pucora/pucora-configurator/internal/store"
)

func testServer(t *testing.T) *api.Server {
	t.Helper()
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return api.NewServer(nil, st, "", "http://localhost:8081")
}

func TestHealth(t *testing.T) {
	srv := api.NewServer(nil, nil, "", "")
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestValidateStructuredErrors(t *testing.T) {
	srv := api.NewServer(nil, nil, "", "")
	body := `{"profile":{"metadata":{"name":"t"},"gateway":{"port":8080},"routes":[{"path":"bad","method":"GET","backend":{"type":"http","host":"","path":"/"}}]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	var resp struct {
		Valid  bool                      `json:"valid"`
		Errors []profile.ValidationError `json:"errors"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Valid {
		t.Fatal("expected invalid profile")
	}
	if len(resp.Errors) == 0 {
		t.Fatal("expected structured errors")
	}
}

func TestPresetsList(t *testing.T) {
	srv := api.NewServer(nil, nil, "", "")
	req := httptest.NewRequest(http.MethodGet, "/api/presets", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestConfigPostAndGet(t *testing.T) {
	srv := testServer(t)

	body := `{
		"name": "prod",
		"profile": {
			"metadata": {"name": "prod-gw"},
			"gateway": {"port": 8080},
			"routes": [{
				"path": "/api",
				"method": "GET",
				"backend": {"type": "http", "host": "http://localhost:8000", "path": "/"}
			}]
		}
	}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/config/prod", strings.NewReader(body))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	srv.Handler().ServeHTTP(postW, postReq)
	if postW.Code != http.StatusOK {
		t.Fatalf("POST expected 200, got %d: %s", postW.Code, postW.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/config/prod/pucora.json", nil)
	getW := httptest.NewRecorder()
	srv.Handler().ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("GET pucora.json expected 200, got %d", getW.Code)
	}
	if !strings.Contains(getW.Body.String(), `"port": 8080`) {
		t.Fatalf("expected port in json, got %s", getW.Body.String())
	}
}

func TestConfigList(t *testing.T) {
	srv := testServer(t)

	body := `{"name":"staging","profile":{"metadata":{"name":"s"},"gateway":{"port":9090},"routes":[{"path":"/","method":"GET","backend":{"type":"http","host":"http://x","path":"/"}}]}}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/config", strings.NewReader(body))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	srv.Handler().ServeHTTP(postW, postReq)

	listReq := httptest.NewRequest(http.MethodGet, "/api/configs", nil)
	listW := httptest.NewRecorder()
	srv.Handler().ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listW.Code)
	}
}

func TestImportJSON(t *testing.T) {
	srv := api.NewServer(nil, nil, "", "")
	body := `{"config":{"version":3,"name":"imported","port":8080,"endpoints":[{"endpoint":"/api","method":"GET","backend":[{"host":["http://localhost:8000"],"url_pattern":"/"}]}]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/import-json", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Valid  bool `json:"valid"`
		Profile struct {
			Routes []any `json:"routes"`
		} `json:"profile"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Profile.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(resp.Profile.Routes))
	}
}
