package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ConfigFile   string `json:"config_file"`
	ThemesDir    string `json:"themes_dir"`
	BackupDir    string `json:"backup_dir"`
	CurrentTheme string `json:"current_theme"`
	Version      string `json:"version"`
}

const (
	configFileName = "alacritty-colors.json"
	currentVersion = "1.0.0"
)

func Load(configFile, themesDir, backupDir string) (*Config, error) {
	cfg := &Config{
		Version: currentVersion,
	}

	if err := cfg.initPaths(configFile, themesDir, backupDir); err != nil {
		return nil, err
	}

	if err := cfg.loadFromFile(); err != nil {
		return nil, err
	}

	if err := cfg.createDirectories(); err != nil {
		return nil, err
	}

	return cfg, cfg.save()
}

func (c *Config) initPaths(configFile, themesDir, backupDir string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var baseConfigDir string
	switch runtime.GOOS {
	case "darwin":
		baseConfigDir = filepath.Join(homeDir, ".config", "alacritty")
	case "linux":
		baseConfigDir = filepath.Join(homeDir, ".config", "alacritty")
	case "windows":
		baseConfigDir = filepath.Join(homeDir, "AppData", "Roaming", "alacritty")
	default:
		baseConfigDir = filepath.Join(homeDir, ".config", "alacritty")
	}

	// Set defaults or use provided values
	if configFile != "" {
		c.ConfigFile = configFile
	} else {
		c.ConfigFile = filepath.Join(baseConfigDir, "alacritty.toml")
	}

	if themesDir != "" {
		c.ThemesDir = themesDir
	} else {
		// Create themes directory next to config
		c.ThemesDir = filepath.Join(baseConfigDir, "themes")
	}

	if backupDir != "" {
		c.BackupDir = backupDir
	} else {
		c.BackupDir = filepath.Join(baseConfigDir, "backups")
	}

	return nil
}

func (c *Config) loadFromFile() error {
	configDir := filepath.Dir(c.ConfigFile)
	configPath := filepath.Join(configDir, configFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, use defaults
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge file config with current config
	if fileConfig.ConfigFile != "" {
		c.ConfigFile = fileConfig.ConfigFile
	}
	if fileConfig.ThemesDir != "" {
		c.ThemesDir = fileConfig.ThemesDir
	}
	if fileConfig.BackupDir != "" {
		c.BackupDir = fileConfig.BackupDir
	}
	if fileConfig.CurrentTheme != "" {
		c.CurrentTheme = fileConfig.CurrentTheme
	}

	return nil
}

func (c *Config) createDirectories() error {
	dirs := []string{
		filepath.Dir(c.ConfigFile),
		c.ThemesDir,
		c.BackupDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (c *Config) save() error {
	configDir := filepath.Dir(c.ConfigFile)
	configPath := filepath.Join(configDir, configFileName)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) SetCurrentTheme(theme string) error {
	c.CurrentTheme = theme
	return c.save()
}

// Save persists the current configuration to disk
func (c *Config) Save() error {
	return c.save()
}

// GetThemePath returns the full path to a theme file
func (c *Config) GetThemePath(themeName string) string {
	return filepath.Join(c.ThemesDir, themeName+".toml")
}
