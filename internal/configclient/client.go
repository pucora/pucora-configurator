package configclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/velonetics/velonetics-configurator/internal/profile"
	"gopkg.in/yaml.v3"
)

// Client talks to the velonetics-config-api config store.
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

type listResponse struct {
	Configs []string `json:"configs"`
}

type bundleResponse struct {
	Name           string         `json:"name"`
	ProfileYAML    string         `json:"profile_yaml"`
	VeloneticsJSON map[string]any `json:"velonetics_json"`
	Env            map[string]string `json:"env"`
	ComposeYAML    string         `json:"compose_yaml"`
}

func (c *Client) List() ([]string, error) {
	var out listResponse
	if err := c.getJSON("/api/configs", &out); err != nil {
		return nil, err
	}
	return out.Configs, nil
}

func (c *Client) Push(p *profile.Profile, name string, compose bool) error {
	body := map[string]any{
		"name":    name,
		"profile": p,
		"compose": compose,
	}
	return c.postJSON("/api/config/"+name, body, nil)
}

func (c *Client) Pull(name, outputDir string) error {
	var bundle bundleResponse
	if err := c.getJSON("/api/config/"+name, &bundle); err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	if bundle.ProfileYAML != "" {
		if err := os.WriteFile(filepath.Join(outputDir, "profile.yaml"), []byte(bundle.ProfileYAML), 0o644); err != nil {
			return err
		}
	}

	if bundle.VeloneticsJSON != nil {
		data, err := json.MarshalIndent(bundle.VeloneticsJSON, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if err := os.WriteFile(filepath.Join(outputDir, "velonetics.json"), data, 0o644); err != nil {
			return err
		}
	}

	if len(bundle.Env) > 0 {
		var sb strings.Builder
		for k, v := range bundle.Env {
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			sb.WriteString("\n")
		}
		if err := os.WriteFile(filepath.Join(outputDir, ".env"), []byte(sb.String()), 0o644); err != nil {
			return err
		}
	}

	if bundle.ComposeYAML != "" {
		if err := os.WriteFile(filepath.Join(outputDir, "docker-compose.yml"), []byte(bundle.ComposeYAML), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) PullProfile(name string) (*profile.Profile, error) {
	var bundle bundleResponse
	if err := c.getJSON("/api/config/"+name, &bundle); err != nil {
		return nil, err
	}
	if bundle.ProfileYAML == "" {
		return nil, fmt.Errorf("config %q has no profile YAML", name)
	}
	var p profile.Profile
	if err := yaml.Unmarshal([]byte(bundle.ProfileYAML), &p); err != nil {
		return nil, err
	}
	profile.ApplyDefaults(&p)
	return &p, nil
}

func (c *Client) getJSON(path string, dest any) error {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, dest)
}

func (c *Client) postJSON(path string, body any, dest any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setHeaders(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest == nil {
		if resp.StatusCode >= 400 {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))
		}
		return nil
	}
	return decodeResponse(resp, dest)
}

func (c *Client) setHeaders(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
}

func decodeResponse(resp *http.Response, dest any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var errObj struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(body, &errObj)
		if errObj.Error != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, errObj.Error)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	if dest == nil {
		return nil
	}
	return json.Unmarshal(body, dest)
}
