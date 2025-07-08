package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vitruves/alacritty-colors/internal/config"
	"github.com/vitruves/alacritty-colors/internal/downloader"
	"github.com/vitruves/alacritty-colors/internal/tui"
	"github.com/vitruves/alacritty-colors/internal/ui"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load("", "", "")
	if err != nil {
		ui.PrintError("Failed to load config: %v", err)
		os.Exit(1)
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