package velocheck

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Result from running pucora check.
type Result struct {
	OK     bool   `json:"ok"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// Run executes `pucora check -c <configPath>` if pucora is on PATH.
func Run(configPath string) (*Result, error) {
	if _, err := exec.LookPath("pucora"); err != nil {
		return &Result{
			OK:    false,
			Error: "pucora binary not found on PATH — install Pucora CE to enable check",
		}, nil
	}

	abs, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("pucora", "check", "-c", abs)
	out, err := cmd.CombinedOutput()
	res := &Result{Output: string(out)}
	if err != nil {
		res.OK = false
		res.Error = err.Error()
		return res, nil
	}
	res.OK = true
	return res, nil
}

// RunFromDir writes config to a temp file and runs check.
func RunFromDir(configJSON []byte, dir string) (*Result, error) {
	path := filepath.Join(dir, "pucora.json")
	if err := os.WriteFile(path, configJSON, 0o644); err != nil {
		return nil, fmt.Errorf("write temp config: %w", err)
	}
	return Run(path)
}
