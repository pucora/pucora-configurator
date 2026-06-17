package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/velonetics/velonetics-configurator/internal/generator"
	"github.com/velonetics/velonetics-configurator/internal/presets"
	"github.com/velonetics/velonetics-configurator/internal/profile"
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

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(presetsCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive wizard to create a gateway profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		w := wizard.New()
		p, err := w.Run()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil && filepath.Dir(output) != "." {
			return err
		}
		if err := profile.Save(output, p); err != nil {
			return err
		}
		fmt.Printf("Profile written to %s\n", output)
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

		p, err := profile.Load(profilePath)
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
		p, err := profile.Load(profilePath)
		if err != nil {
			return err
		}
		fmt.Printf("Profile %q is valid (%d routes)\n", p.Metadata.Name, len(p.Routes))
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

		p, err := presets.Load(name)
		if err != nil {
			return err
		}

		if output != "" {
			if err := profile.Save(output, p); err != nil {
				return err
			}
			fmt.Printf("Preset %q saved to %s\n", name, output)
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
	initCmd.Flags().StringP("output", "o", "profile.yaml", "Output profile file path")

	generateCmd.Flags().StringP("file", "f", "profile.yaml", "Input profile YAML file")
	generateCmd.Flags().StringP("output", "o", "./output", "Output directory for velonetics.json")
	generateCmd.Flags().Bool("compose", false, "Generate docker-compose.yml for local development")

	validateCmd.Flags().StringP("file", "f", "profile.yaml", "Input profile YAML file")

	presetsApplyCmd.Flags().StringP("output", "o", "", "Save preset as profile YAML")
	presetsApplyCmd.Flags().StringP("generate-dir", "g", "", "Generate velonetics.json to directory")
	presetsApplyCmd.Flags().Bool("compose", false, "Generate docker-compose.yml for local development")

	presetsCmd.AddCommand(presetsListCmd)
	presetsCmd.AddCommand(presetsApplyCmd)
}
