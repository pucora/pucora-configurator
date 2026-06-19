package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pucora/pucora-configurator/internal/generator"
	"github.com/pucora/pucora-configurator/internal/profile"
)

// Summary describes differences between two profiles or generated configs.
type Summary struct {
	ProfileDiff   []string `json:"profile_diff,omitempty"`
	GeneratedDiff []string `json:"generated_diff,omitempty"`
}

// Profiles compares two profiles by JSON serialization and generated pucora.json.
func Profiles(a, b *profile.Profile) (*Summary, error) {
	s := &Summary{}

	if d := jsonDiff(a, b); len(d) > 0 {
		s.ProfileDiff = d
	}

	outA, err := generator.Generate(a)
	if err != nil {
		return nil, fmt.Errorf("generate A: %w", err)
	}
	outB, err := generator.Generate(b)
	if err != nil {
		return nil, fmt.Errorf("generate B: %w", err)
	}

	if d := mapDiff(outA.Config, outB.Config); len(d) > 0 {
		s.GeneratedDiff = d
	}

	return s, nil
}

func jsonDiff(a, b any) []string {
	ja, _ := json.MarshalIndent(a, "", "  ")
	jb, _ := json.MarshalIndent(b, "", "  ")
	linesA := strings.Split(string(ja), "\n")
	linesB := strings.Split(string(jb), "\n")

	max := len(linesA)
	if len(linesB) > max {
		max = len(linesB)
	}

	var diff []string
	for i := 0; i < max; i++ {
		la, lb := "", ""
		if i < len(linesA) {
			la = linesA[i]
		}
		if i < len(linesB) {
			lb = linesB[i]
		}
		if la != lb {
			diff = append(diff, fmt.Sprintf("L%d: - %s", i+1, la))
			diff = append(diff, fmt.Sprintf("L%d: + %s", i+1, lb))
		}
	}
	if len(diff) > 40 {
		diff = append(diff[:40], fmt.Sprintf("... and %d more line differences", len(diff)-40))
	}
	return diff
}

func mapDiff(a, b map[string]any) []string {
	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var diff []string
	for _, k := range sorted {
		va, oka := a[k]
		vb, okb := b[k]
		ja, _ := json.Marshal(va)
		jb, _ := json.Marshal(vb)
		if !oka {
			diff = append(diff, fmt.Sprintf("+ %s: %s", k, string(jb)))
		} else if !okb {
			diff = append(diff, fmt.Sprintf("- %s: %s", k, string(ja)))
		} else if string(ja) != string(jb) {
			diff = append(diff, fmt.Sprintf("~ %s changed", k))
		}
	}
	return diff
}
