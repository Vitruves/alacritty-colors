package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vitruves/alacritty-colors/internal/config"
	"github.com/vitruves/alacritty-colors/internal/downloader"
	"github.com/vitruves/alacritty-colors/internal/tui"
	"github.com/vitruves/alacritty-colors/internal/ui"
)

const version = "2.2.0"

func main() {
	// Define CLI flags
	configPath := flag.String("config", "", "Path to alacritty.toml config file")
	themesPath := flag.String("themes", "", "Path to themes directory")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help information")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `alacritty-colors - Interactive theme manager for Alacritty terminal

Usage:
  alacritty-colors [options]

Options:
  -config string
        Path to alacritty.toml config file
        Default: ~/.config/alacritty/alacritty.toml (Linux/macOS)
                 %%APPDATA%%\alacritty\alacritty.toml (Windows)

  -themes string
        Path to themes directory
        Default: ~/.config/alacritty/themes/

  -version
        Show version information

  -help
        Show this help message

Examples:
  # Use default configuration
  alacritty-colors

  # Use custom config location
  alacritty-colors -config /path/to/alacritty.toml

  # Use custom themes directory
  alacritty-colors -themes /path/to/themes

  # Use both custom paths
  alacritty-colors -config /path/to/alacritty.toml -themes /path/to/themes

Keybindings:
  Tab          Switch panels
  ↑/↓          Navigate
  ←/→          Adjust brightness
  Shift+←/→    Adjust hue
  n            Create new theme
  g            Generate random theme
  /            Search themes
  ?            Show all keybindings
  q            Quit

For more information, visit: https://github.com/vitruves/alacritty-colors
`)
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("alacritty-colors version %s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Initialize configuration with custom paths if provided
	cfg, err := config.Load(*configPath, *themesPath, "")
	if err != nil {
		ui.PrintError("Failed to load config: %v", err)
		os.Exit(1)
	}

	// Show config location if custom path was used
	if *configPath != "" || *themesPath != "" {
		ui.PrintInfo("Using config: %s", cfg.ConfigFile)
		ui.PrintInfo("Using themes: %s", cfg.ThemesDir)
	}

	// Check if themes directory exists and has themes
	themesExist := checkThemesExist(cfg.ThemesDir)

	if !themesExist {
		// First time run - download themes
		ui.PrintHeader("First-time setup: Downloading themes...")

		if err := downloadThemes(cfg); err != nil {
			ui.PrintError("Failed to download themes: %v", err)
			os.Exit(1)
		}

		ui.PrintSuccess("Themes downloaded successfully!")
	}

	// Launch TUI
	ui.PrintInfo("Launching theme editor...")
	if err := tui.StartInteractive(cfg); err != nil {
		ui.PrintError("TUI error: %v", err)
		os.Exit(1)
	}
}

func checkThemesExist(themesDir string) bool {
	// Check if themes directory exists
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return false
	}

	// Check if it has any .toml files
	files, err := os.ReadDir(themesDir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".toml" {
			return true
		}
	}

	return false
}

func downloadThemes(cfg *config.Config) error {
	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(cfg.ThemesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	// Download themes
	dl := downloader.New(cfg.ThemesDir)
	count, err := dl.DownloadOfficialThemes()
	if err != nil {
		return err
	}

	ui.PrintInfo("Downloaded %d themes", count)
	return nil
}
