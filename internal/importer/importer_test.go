package importer_test

import (
	"encoding/json"
	"testing"

	"github.com/pucora/pucora-configurator/internal/generator"
	"github.com/pucora/pucora-configurator/internal/importer"
	"github.com/pucora/pucora-configurator/internal/presets"
)

func TestRoundTripPresets(t *testing.T) {
	list, err := presets.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, preset := range list {
		t.Run(preset.Name, func(t *testing.T) {
			orig, err := presets.Load(preset.Name)
			if err != nil {
				t.Fatal(err)
			}
			out, err := generator.Generate(orig)
			if err != nil {
				t.Fatal(err)
			}
			res, err := importer.FromPucoraJSON(out.Config)
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Profile.Routes) != len(orig.Routes) {
				t.Errorf("routes count: got %d want %d", len(res.Profile.Routes), len(orig.Routes))
			}
			if orig.GRPC != nil && len(orig.GRPC.Catalog) > 0 {
				if res.Profile.GRPC == nil || len(res.Profile.GRPC.Catalog) == 0 {
					t.Error("expected grpc catalog after import")
				}
			}
			// Re-generate imported profile — must produce valid JSON
			_, err = generator.Generate(&res.Profile)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(out.Config)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) < 50 {
				t.Fatal("config too short")
			}
		})
	}
}
