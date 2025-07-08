package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vitruves/alacritty-colors/internal/config"
)

// Manager provides only the core theme management functions needed by the TUI
type Manager struct {
	config  *config.Config
	verbose bool
	silent  bool
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

func (m *Manager) SetVerbose(verbose bool) {
	m.verbose = verbose
}

func (m *Manager) SetSilent(silent bool) {
	m.silent = silent
}

func (m *Manager) ApplyTheme(themeName string) error {
	// Get source theme path
	themeFile := m.config.GetThemePath(themeName)
	if _, err := os.Stat(themeFile); os.IsNotExist(err) {
		return fmt.Errorf("theme '%s' not found", themeName)
	}

	// Get destination path (current.toml)
	currentThemePath := m.config.GetThemePath("current")

	// Copy theme to current.toml
	if err := m.copyFile(themeFile, currentThemePath); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}

	// Write a tracking file to remember which theme is applied
	trackingFile := m.config.GetThemePath(".current-theme")
	if err := os.WriteFile(trackingFile, []byte(themeName), 0644); err != nil {
		// Don't fail if we can't write tracking file, just continue
	}

	// Ensure main config has import line
	if err := m.ensureImportLine(); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

func (m *Manager) GetCurrentTheme() string {
	currentThemePath := m.config.GetThemePath("current")
	
	// If current.toml doesn't exist, no theme is applied
	if _, err := os.Stat(currentThemePath); os.IsNotExist(err) {
		return ""
	}

	// First check tracking file for applied theme name
	trackingFile := m.config.GetThemePath(".current-theme")
	if trackingData, err := os.ReadFile(trackingFile); err == nil {
		themeName := strings.TrimSpace(string(trackingData))
		// Verify the tracked theme actually exists
		if themeName != "" {
			themeFile := m.config.GetThemePath(themeName)
			if _, err := os.Stat(themeFile); err == nil {
				return themeName
			}
		}
	}

	// Fallback: Find which theme matches current.toml by comparing content
	files, err := os.ReadDir(m.config.ThemesDir)
	if err != nil {
		return ""
	}

	currentContent, err := os.ReadFile(currentThemePath)
	if err != nil {
		return ""
	}

	for _, file := range files {
		if file.IsDir() || file.Name() == "current.toml" || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".toml") {
			continue
		}

		themePath := filepath.Join(m.config.ThemesDir, file.Name())
		themeContent, err := os.ReadFile(themePath)
		if err != nil {
			continue
		}

		// Compare content (ignoring whitespace differences)
		if strings.TrimSpace(string(currentContent)) == strings.TrimSpace(string(themeContent)) {
			return strings.TrimSuffix(file.Name(), ".toml")
		}
	}

	return ""
}

func (m *Manager) copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, sourceData, 0644)
}

func (m *Manager) ensureImportLine() error {
	configContent, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		// If config doesn't exist, create a minimal one
		if os.IsNotExist(err) {
			return m.createMinimalConfig()
		}
		return err
	}

	content := string(configContent)
	importLine := `import = ["themes/current.toml"]`

	// Check if import line already exists
	if strings.Contains(content, importLine) {
		return nil
	}

	// Add import line at the beginning
	newContent := importLine + "\n\n" + content
	return os.WriteFile(m.config.ConfigFile, []byte(newContent), 0644)
}

func (m *Manager) createMinimalConfig() error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(m.config.ConfigFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Create minimal config with just the import line
	content := `import = ["themes/current.toml"]

# Minimal Alacritty configuration
# Theme colors are imported from themes/current.toml
`

	return os.WriteFile(m.config.ConfigFile, []byte(content), 0644)
}