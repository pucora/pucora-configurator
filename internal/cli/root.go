package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/velonetics/velonetics-configurator/internal/doctor"
	"github.com/velonetics/velonetics-configurator/internal/generator"
	"github.com/velonetics/velonetics-configurator/internal/presets"
	"github.com/velonetics/velonetics-configurator/internal/profile"
	"github.com/velonetics/velonetics-configurator/internal/velocheck"
	"github.com/velonetics/velonetics-configurator/internal/wizard"
)

var rootCmd = &cobra.Command{
	Use:   "velonetics-config",
	Short: "Generate Velonetics gateway configuration from simple profiles",
	Long: `Velonetics Configurator turns a simple YAML profile into a complete
velonetics.json with routes, CORS, headers, auth, pub/sub, gRPC, and more.

Choose a preset, write a profile by hand, or run the interactive wizard.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive wizard to create a gateway profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		fromPreset, _ := cmd.Flags().GetString("from-preset")

		var p *profile.Profile
		var err error

		if fromPreset != "" {
			p, err = presets.Load(fromPreset)
			if err != nil {
				return err
			}
			fmt.Printf("Loaded preset %q — customize in your editor if needed\n", fromPreset)
		} else {
			w := wizard.New()
			p, err = w.Run()
			if err != nil {
				return err
			}
		}

		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil && filepath.Dir(output) != "." {
			return err
		}
		if err := profile.Save(output, p); err != nil {
			return err
		}
		fmt.Printf("Profile written to %s\n", output)

		edit, _ := cmd.Flags().GetBool("edit")
		if edit {
			if err := openInEditor(output); err != nil {
				return err
			}
		}

		fmt.Println("Run: velonetics-config generate -f", output, "-o ./output")
		return nil
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate velonetics.json from a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, _ := cmd.Flags().GetString("file")
		outputDir, _ := cmd.Flags().GetString("output")
		withCompose, _ := cmd.Flags().GetBool("compose")
		toStdout, _ := cmd.Flags().GetBool("stdout")
		runCheck, _ := cmd.Flags().GetBool("check")

		p, err := profile.Load(profilePath)
		if err != nil {
			return err
		}

		out, err := generator.Generate(p)
		if err != nil {
			return err
		}

		for _, a := range doctor.Check(p) {
			if a.Level == "warn" {
				fmt.Fprintf(os.Stderr, "  advisory [%s]: %s\n", a.Field, a.Message)
			}
		}

		if toStdout {
			data, err := json.MarshalIndent(out.Config, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		if err := generator.Write(outputDir, out, p, withCompose); err != nil {
			return err
		}

		fmt.Printf("Generated %s/velonetics.json\n", outputDir)
		if len(out.Env) > 0 {
			fmt.Printf("Generated %s/.env\n", outputDir)
		}
		if withCompose || (p.Compose != nil && p.Compose.Enabled != nil && *p.Compose.Enabled) {
			fmt.Printf("Generated %s/docker-compose.yml\n", outputDir)
		}
		for _, w := range out.Warnings {
			fmt.Printf("  warning: %s\n", w)
		}

		if runCheck {
			res, err := velocheck.Run(filepath.Join(outputDir, "velonetics.json"))
			if err != nil {
				return err
			}
			if res.Error != "" && !res.OK {
				fmt.Fprintf(os.Stderr, "velonetics check: %s\n", res.Error)
				if res.Output != "" {
					fmt.Fprint(os.Stderr, res.Output)
				}
			} else if res.OK {
				fmt.Println("velonetics check: OK")
			} else if res.Output != "" {
				fmt.Print(res.Output)
			}
		}

		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  velonetics check -c %s/velonetics.json\n", outputDir)
		fmt.Printf("  velonetics run -c %s/velonetics.json\n", outputDir)
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a profile without generating output",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilePath, _ := cmd.Flags().GetString("file")
		asJSON, _ := cmd.Flags().GetBool("json")

		data, err := os.ReadFile(profilePath)
		if err != nil {
			return err
		}
		var p profile.Profile
		if err := profile.UnmarshalYAML(data, &p); err != nil {
			return err
		}
		profile.ApplyDefaults(&p)
		errs := profile.ValidateStructured(&p)
		advisories := doctor.Check(&p)

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{
				"valid":       !errs.HasErrors(),
				"errors":      errs,
				"advisories":  advisories,
				"route_count": len(p.Routes),
				"name":        p.Metadata.Name,
			})
		}

		if errs.HasErrors() {
			for _, e := range errs {
				fmt.Printf("  error [%s]: %s\n", e.Field, e.Message)
			}
			return fmt.Errorf("profile has %d validation error(s)", len(errs))
		}

		fmt.Printf("Profile %q is valid (%d routes)\n", p.Metadata.Name, len(p.Routes))
		for _, a := range advisories {
			fmt.Printf("  advisory [%s]: %s\n", a.Field, a.Message)
		}
		return nil
	},
}

var presetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List and apply built-in configuration presets",
}

var presetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := presets.List()
		if err != nil {
			return err
		}
		fmt.Println("Available presets:")
		for _, p := range list {
			fmt.Printf("  %-20s %s\n", p.Name, p.Description)
		}
		return nil
	},
}

var presetsApplyCmd = &cobra.Command{
	Use:   "apply [preset-name]",
	Short: "Copy a preset profile to your workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		output, _ := cmd.Flags().GetString("output")
		genDir, _ := cmd.Flags().GetString("generate-dir")
		withCompose, _ := cmd.Flags().GetBool("compose")
		edit, _ := cmd.Flags().GetBool("edit")

		p, err := presets.Load(name)
		if err != nil {
			return err
		}

		if output != "" {
			if err := profile.Save(output, p); err != nil {
				return err
			}
			fmt.Printf("Preset %q saved to %s\n", name, output)
			if edit {
				if err := openInEditor(output); err != nil {
					return err
				}
			}
		}

		if genDir != "" {
			out, err := generator.Generate(p)
			if err != nil {
				return err
			}
			if err := generator.Write(genDir, out, p, withCompose); err != nil {
				return err
			}
			fmt.Printf("Generated %s/velonetics.json from preset %q\n", genDir, name)
			if withCompose || (p.Compose != nil && p.Compose.Enabled != nil && *p.Compose.Enabled) {
				fmt.Printf("Generated %s/docker-compose.yml\n", genDir)
			}
		}

		if output == "" && genDir == "" {
			return fmt.Errorf("specify --output and/or --generate-dir")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(presetsCmd)

	initCmd.Flags().StringP("output", "o", "profile.yaml", "Output profile file path")
	initCmd.Flags().String("from-preset", "", "Start from a built-in preset instead of the wizard")
	initCmd.Flags().Bool("edit", false, "Open profile in $EDITOR after creation")

	generateCmd.Flags().StringP("file", "f", "profile.yaml", "Input profile YAML file")
	generateCmd.Flags().StringP("output", "o", "./output", "Output directory for velonetics.json")
	generateCmd.Flags().Bool("compose", false, "Generate docker-compose.yml for local development")
	generateCmd.Flags().Bool("stdout", false, "Print velonetics.json to stdout instead of writing files")
	generateCmd.Flags().Bool("check", false, "Run velonetics check after generating (requires velonetics on PATH)")

	validateCmd.Flags().StringP("file", "f", "profile.yaml", "Input profile YAML file")
	validateCmd.Flags().Bool("json", false, "Output validation result as JSON")

	presetsApplyCmd.Flags().StringP("output", "o", "", "Save preset as profile YAML")
	presetsApplyCmd.Flags().StringP("generate-dir", "g", "", "Generate velonetics.json to directory")
	presetsApplyCmd.Flags().Bool("compose", false, "Generate docker-compose.yml for local development")
	presetsApplyCmd.Flags().Bool("edit", false, "Open saved profile in $EDITOR")

	presetsCmd.AddCommand(presetsListCmd)
	presetsCmd.AddCommand(presetsApplyCmd)
}
