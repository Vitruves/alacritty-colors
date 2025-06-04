package main

import (
	"fmt"
	"os"

	"github.com/vitruves/alacritty-colors/internal/config"
	"github.com/vitruves/alacritty-colors/internal/theme"
	"github.com/vitruves/alacritty-colors/internal/ui"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

var (
	configFile string
	themesDir  string
	backupDir  string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "alacritty-colors",
		Short: "Advanced Alacritty theme manager with generation capabilities",
		Long: ui.ColorizeHeader(`
Alacritty Colors v` + version + `
Advanced theme manager with automatic downloading, generation, and management

Features:
- Download official Alacritty themes automatically
- Generate custom themes with various color schemes
- Backup and restore configurations
- Search and preview themes
- Cross-platform support
`),
		Version: version,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&themesDir, "themes-dir", "", "themes directory path")
	rootCmd.PersistentFlags().StringVar(&backupDir, "backup-dir", "", "backup directory path")

	// Commands
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(applyCmd())
	rootCmd.AddCommand(randomCmd())
	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(backupCmd())
	rootCmd.AddCommand(restoreCmd())
	rootCmd.AddCommand(updateCmd())
	rootCmd.AddCommand(previewCmd())

	if err := rootCmd.Execute(); err != nil {
		ui.PrintError("Error: %v", err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize alacritty-colors configuration and download themes",
		Long:  "Set up configuration directories, download official themes, and prepare the environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.PrintHeader("Initializing Alacritty Colors")

			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			tm := theme.NewManager(cfg)
			return tm.Initialize()
		},
	}
}

func listCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available themes",
		Long:  "Display all available themes in various formats",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.ListThemes(format)
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "grid", "output format (grid, list, json)")
	return cmd
}

func applyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply [theme-name]",
		Short: "Apply a specific theme",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.ApplyTheme(args[0])
		},
	}
}

func randomCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "random",
		Short: "Apply a random theme",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.RandomTheme()
		},
	}
}

func generateCmd() *cobra.Command {
	var scheme string
	var name string
	var save bool

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a new theme",
		Long: `Generate a new theme using various color schemes:
- random    - Completely random colors
- pastel    - Soft, muted colors
- neon      - Bright, vibrant colors  
- mono      - Monochromatic grayscale
- warm      - Warm tones (reds, oranges, yellows)
- cool      - Cool tones (blues, greens, purples)
- nature    - Earth and nature-inspired colors
- cyberpunk - Neon cyberpunk aesthetic
- dracula   - Dracula-inspired dark theme
- nord      - Nord-inspired cool theme`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.GenerateTheme(scheme, name, save)
		},
	}

	cmd.Flags().StringVarP(&scheme, "scheme", "s", "random", "color scheme to use")
	cmd.Flags().StringVarP(&name, "name", "n", "", "custom theme name")
	cmd.Flags().BoolVar(&save, "save", true, "save generated theme to disk")
	return cmd
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search themes by name or tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.SearchThemes(args[0])
		},
	}
}

func backupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backup",
		Short: "Create a backup of current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.CreateBackup()
		},
	}
}

func restoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore [backup-file]",
		Short: "Restore configuration from backup",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			var backupFile string
			if len(args) > 0 {
				backupFile = args[0]
			}
			return tm.RestoreBackup(backupFile)
		},
	}
}

func updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update theme database",
		Long:  "Download the latest themes from the official repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.UpdateThemes()
		},
	}
}

func previewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview [theme-name]",
		Short: "Preview a theme without applying it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile, themesDir, backupDir)
			if err != nil {
				return err
			}

			tm := theme.NewManager(cfg)
			return tm.PreviewTheme(args[0])
		},
	}
}
