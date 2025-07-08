package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vitruves/alacritty-colors/internal/downloader"
)

// ParametersManager handles utility operations for alacritty configuration
type ParametersManager struct {
	config *Config
}

func NewParametersManager(cfg *Config) *ParametersManager {
	return &ParametersManager{
		config: cfg,
	}
}

// CleanAndRedownloadThemes removes all existing themes and redownloads them
func (pm *ParametersManager) CleanAndRedownloadThemes() error {
	// Clean existing themes (except current.toml)
	if err := pm.cleanExistingThemes(); err != nil {
		return fmt.Errorf("failed to clean existing themes: %w", err)
	}
	
	// Redownload themes using the existing downloader
	dl := downloader.New(pm.config.ThemesDir)
	count, err := dl.DownloadOfficialThemes()
	if err != nil {
		return fmt.Errorf("failed to download themes: %w", err)
	}
	
	if count == 0 {
		return fmt.Errorf("no themes were downloaded")
	}
	
	return nil
}

// BackupConfig creates a timestamped backup of the current configuration
func (pm *ParametersManager) BackupConfig() error {
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("alacritty_backup_%s.toml", timestamp)
	backupPath := filepath.Join(pm.config.BackupDir, backupName)
	
	// Ensure backup directory exists
	if err := os.MkdirAll(pm.config.BackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Copy config file to backup
	configData, err := os.ReadFile(pm.config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := os.WriteFile(backupPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}
	
	return nil
}

// ResetToDefaults resets the Alacritty configuration to defaults
func (pm *ParametersManager) ResetToDefaults() error {
	// Create a minimal default configuration
	defaultConfig := `# Alacritty Configuration - Reset to Defaults
[general]
import = ["themes/current.toml"]

[window]
padding = { x = 2, y = 2 }

[font]
normal = { family = "monospace", style = "Regular" }
size = 12

[terminal]
shell = { program = "/bin/bash" }
`
	
	// Backup current config first
	if err := pm.BackupConfig(); err != nil {
		return fmt.Errorf("failed to backup current config: %w", err)
	}
	
	// Write default config
	if err := os.WriteFile(pm.config.ConfigFile, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}
	
	return nil
}

// cleanExistingThemes removes all theme files except current.toml and tracking files
func (pm *ParametersManager) cleanExistingThemes() error {
	files, err := os.ReadDir(pm.config.ThemesDir)
	if err != nil {
		return err
	}
	
	for _, file := range files {
		if !file.IsDir() && 
		   file.Name() != "current.toml" && 
		   !strings.HasPrefix(file.Name(), ".current-theme") {
			themePath := filepath.Join(pm.config.ThemesDir, file.Name())
			if err := os.Remove(themePath); err != nil {
				return fmt.Errorf("failed to remove theme file %s: %w", file.Name(), err)
			}
		}
	}
	
	return nil
}