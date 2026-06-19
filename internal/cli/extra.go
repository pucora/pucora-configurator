package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/pucora/pucora-configurator/internal/configclient"
	"github.com/pucora/pucora-configurator/internal/diff"
	"github.com/pucora/pucora-configurator/internal/doctor"
	"github.com/pucora/pucora-configurator/internal/generator"
	"github.com/pucora/pucora-configurator/internal/importer"
	"github.com/pucora/pucora-configurator/internal/profile"
)

func init() {
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(configCmd)
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import pucora.json into a simplified profile YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("file")
		output, _ := cmd.Flags().GetString("output")

		data, err := os.ReadFile(input)
		if err != nil {
			return err
		}
		var cfg map[string]any
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse JSON: %w", err)
		}

		res, err := importer.FromPucoraJSON(cfg)
		if err != nil {
			return err
		}

		if output == "" || output == "-" {
			yamlData, err := profile.MarshalYAML(&res.Profile)
			if err != nil {
				return err
			}
			fmt.Print(string(yamlData))
			return nil
		}

		if err := profile.Save(output, &res.Profile); err != nil {
			return err
		}
		fmt.Printf("Profile written to %s (%d routes)\n", output, len(res.Profile.Routes))
		for _, w := range res.Warnings {
			fmt.Printf("  import warning: %s\n", w)
		}
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare two profile files",
	RunE: func(cmd *cobra.Command, args []string) error {
		fileA, _ := cmd.Flags().GetString("file-a")
		fileB, _ := cmd.Flags().GetString("file-b")
		asJSON, _ := cmd.Flags().GetBool("json")

		a, err := profile.Load(fileA)
		if err != nil {
			return err
		}
		b, err := profile.Load(fileB)
		if err != nil {
			return err
		}

		summary, err := diff.Profiles(a, b)
		if err != nil {
			return err
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(summary)
		}

		if len(summary.ProfileDiff) == 0 && len(summary.GeneratedDiff) == 0 {
			fmt.Println("No differences")
			return nil
		}
		if len(summary.ProfileDiff) > 0 {
			fmt.Println("Profile differences:")
			for _, line := range summary.ProfileDiff {
				fmt.Println(" ", line)
			}
		}
		if len(summary.GeneratedDiff) > 0 {
			fmt.Println("Generated pucora.json differences:")
			for _, line := range summary.GeneratedDiff {
				fmt.Println(" ", line)
			}
		}
		return nil
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Show advisory warnings for a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, _ := cmd.Flags().GetString("file")
		asJSON, _ := cmd.Flags().GetBool("json")

		p, err := profile.Load(profilePath)
		if err != nil {
			return err
		}

		advisories := doctor.Check(p)
		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{"advisories": advisories})
		}

		if len(advisories) == 0 {
			fmt.Println("No advisories — profile looks good")
			return nil
		}
		for _, a := range advisories {
			fmt.Printf("[%s] %s: %s\n", a.Level, a.Field, a.Message)
		}
		return nil
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Regenerate pucora.json when the profile changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, _ := cmd.Flags().GetString("file")
		outputDir, _ := cmd.Flags().GetString("output")
		withCompose, _ := cmd.Flags().GetBool("compose")

		abs, err := filepath.Abs(profilePath)
		if err != nil {
			return err
		}

		var lastMod int64
		regenerate := func() error {
			p, err := profile.Load(abs)
			if err != nil {
				return err
			}
			out, err := generator.Generate(p)
			if err != nil {
				return err
			}
			if err := generator.Write(outputDir, out, p, withCompose); err != nil {
				return err
			}
			fmt.Printf("[%s] regenerated %s/pucora.json\n", abs, outputDir)
			return nil
		}

		if err := regenerate(); err != nil {
			return err
		}

		stat, _ := os.Stat(abs)
		if stat != nil {
			lastMod = stat.ModTime().UnixNano()
		}

		fmt.Println("Watching for changes (Ctrl+C to stop)...")
		for {
			stat, err := os.Stat(abs)
			if err != nil {
				return err
			}
			if stat.ModTime().UnixNano() != lastMod {
				lastMod = stat.ModTime().UnixNano()
				if err := regenerate(); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
				}
			}
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Push and pull configs from the config store API",
}

var configPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Save a profile to the config store API",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, _ := cmd.Flags().GetString("file")
		name, _ := cmd.Flags().GetString("name")
		apiURL, _ := cmd.Flags().GetString("api-url")
		apiKey, _ := cmd.Flags().GetString("api-key")
		withCompose, _ := cmd.Flags().GetBool("compose")

		p, err := profile.Load(profilePath)
		if err != nil {
			return err
		}
		client := configclient.New(apiURL, apiKey)
		if err := client.Push(p, name, withCompose); err != nil {
			return err
		}
		fmt.Printf("Pushed %q to %s/api/config/%s\n", p.Metadata.Name, apiURL, name)
		fmt.Printf("Pull: curl -s %s/api/config/%s/pucora.json\n", apiURL, name)
		return nil
	},
}

var configPullCmd = &cobra.Command{
	Use:   "pull [name]",
	Short: "Download a config bundle from the API",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		outputDir, _ := cmd.Flags().GetString("output")
		apiURL, _ := cmd.Flags().GetString("api-url")
		apiKey, _ := cmd.Flags().GetString("api-key")

		client := configclient.New(apiURL, apiKey)
		if err := client.Pull(name, outputDir); err != nil {
			return err
		}
		fmt.Printf("Pulled %q to %s/\n", name, outputDir)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved configs on the API",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiURL, _ := cmd.Flags().GetString("api-url")
		apiKey, _ := cmd.Flags().GetString("api-key")
		asJSON, _ := cmd.Flags().GetBool("json")

		client := configclient.New(apiURL, apiKey)
		names, err := client.List()
		if err != nil {
			return err
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{"configs": names})
		}

		if len(names) == 0 {
			fmt.Println("No saved configs")
			return nil
		}
		fmt.Println("Saved configs:")
		for _, n := range names {
			fmt.Printf("  %s\n", n)
		}
		return nil
	},
}

func init() {
	importCmd.Flags().StringP("file", "f", "pucora.json", "Input pucora.json file")
	importCmd.Flags().StringP("output", "o", "profile.yaml", "Output profile YAML (- for stdout)")

	diffCmd.Flags().String("file-a", "profile.yaml", "First profile file")
	diffCmd.Flags().String("file-b", "", "Second profile file (required)")
	diffCmd.MarkFlagRequired("file-b")
	diffCmd.Flags().Bool("json", false, "Output JSON")

	doctorCmd.Flags().StringP("file", "f", "profile.yaml", "Input profile YAML file")
	doctorCmd.Flags().Bool("json", false, "Output JSON")

	watchCmd.Flags().StringP("file", "f", "profile.yaml", "Profile to watch")
	watchCmd.Flags().StringP("output", "o", "./output", "Output directory")
	watchCmd.Flags().Bool("compose", false, "Also generate docker-compose.yml")

	configPushCmd.Flags().StringP("file", "f", "profile.yaml", "Profile to push")
	configPushCmd.Flags().StringP("name", "n", "default", "Config name on the API")
	configPushCmd.Flags().String("api-url", "http://localhost:8081", "Config API base URL")
	configPushCmd.Flags().String("api-key", "", "API key (X-API-Key)")
	configPushCmd.Flags().Bool("compose", false, "Include docker-compose in bundle")

	configPullCmd.Flags().StringP("output", "o", "./output", "Output directory")
	configPullCmd.Flags().String("api-url", "http://localhost:8081", "Config API base URL")
	configPullCmd.Flags().String("api-key", "", "API key (X-API-Key)")

	configListCmd.Flags().String("api-url", "http://localhost:8081", "Config API base URL")
	configListCmd.Flags().String("api-key", "", "API key (X-API-Key)")
	configListCmd.Flags().Bool("json", false, "Output JSON")

	configCmd.AddCommand(configPushCmd)
	configCmd.AddCommand(configPullCmd)
	configCmd.AddCommand(configListCmd)
}

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
