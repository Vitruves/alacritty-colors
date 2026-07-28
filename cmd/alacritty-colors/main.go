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
	exportPath := flag.String("export-collection", "", "Write the curated collection to a directory and exit")

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

  -export-collection string
        Write the 150 curated themes to a directory and exit,
        without touching your Alacritty config

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
  Tab          Next column
  a            Apply what you see
  ↑/↓          Navigate
  ←/→          Brighten / darken
  Shift+←/→    Rotate hue
  Enter        Type an exact hex value (palette panel)
  n            Theme creator
  g            Generate a harmonious theme
  /            Search themes
  s            Save (quitting saves and applies too)
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

	// Export the collection anywhere, for people who want the .toml files
	// without letting the tool near their Alacritty config.
	if *exportPath != "" {
		if err := os.MkdirAll(*exportPath, 0755); err != nil {
			ui.PrintError("Could not create %s: %v", *exportPath, err)
			os.Exit(1)
		}
		written, skipped, err := tui.InstallCollection(*exportPath)
		if err != nil {
			ui.PrintError("Export failed: %v", err)
			os.Exit(1)
		}
		ui.PrintSuccess("Wrote %d themes to %s", written, *exportPath)
		if skipped > 0 {
			ui.PrintInfo("Left %d edited file(s) untouched", skipped)
		}
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
	note, err := tui.StartInteractive(cfg)
	if err != nil {
		ui.PrintError("TUI error: %v", err)
		os.Exit(1)
	}
	if note != "" {
		ui.PrintSuccess("%s", note)
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

	// The curated collection ships with the tool rather than being fetched.
	written, _, err := tui.InstallCollection(cfg.ThemesDir)
	if err != nil {
		// A theme library that downloaded fine is still usable without these.
		ui.PrintError("Could not install the curated collection: %v", err)
		return nil
	}
	ui.PrintInfo("Installed %d curated themes", written)

	return nil
}
