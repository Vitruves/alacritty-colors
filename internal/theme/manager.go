package theme

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/vitruves/alacritty-colors/internal/config"
	"github.com/vitruves/alacritty-colors/internal/downloader"
	"github.com/vitruves/alacritty-colors/internal/ui"
)

const (
	ImportLine = `[general]\nimport = ["themes/current.toml"]`
)

type Manager struct {
	config *config.Config
}

type ThemeInfo struct {
	Name        string
	FilePath    string
	Description string
	Author      string
	Tags        []string
	Colors      map[string]string
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{config: cfg}
}

func (m *Manager) Initialize() error {
	ui.PrintSubHeader("Setting up configuration")

	// Create config file if it doesn't exist
	if _, err := os.Stat(m.config.ConfigFile); os.IsNotExist(err) {
		ui.PrintInfo("Creating default alacritty.toml")
		if err := m.createDefaultConfig(); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		ui.PrintSuccess("Created alacritty.toml")
	}

	// Check if import line exists
	if !m.hasImportLine() {
		ui.PrintInfo("Adding theme import line")
		if err := m.addImportLine(); err != nil {
			return fmt.Errorf("failed to add import line: %w", err)
		}
		ui.PrintSuccess("Added theme import line")
	}

	// Create current.toml (empty initially)
	currentThemePath := filepath.Join(m.config.ThemesDir, "current.toml")
	if _, err := os.Stat(currentThemePath); os.IsNotExist(err) {
		defaultTheme := `# No theme applied
# Run 'alacritty-colors apply <theme-name>' to apply a theme

[colors.primary]
background = "#1e1e1e"
foreground = "#ffffff"
`
		if err := os.WriteFile(currentThemePath, []byte(defaultTheme), 0644); err != nil {
			return fmt.Errorf("failed to create current theme file: %w", err)
		}
	}

	// Download themes
	ui.PrintSubHeader("Downloading themes")
	dl := downloader.New(m.config.ThemesDir)
	count, err := dl.DownloadOfficialThemes()
	if err != nil {
		return fmt.Errorf("failed to download themes: %w", err)
	}

	ui.PrintSuccess("Downloaded %d themes", count)
	ui.PrintSubHeader("Configuration complete")
	ui.PrintInfo("Config file: %s", m.config.ConfigFile)
	ui.PrintInfo("Themes directory: %s", m.config.ThemesDir)
	ui.PrintInfo("Backups directory: %s", m.config.BackupDir)

	return nil
}

func (m *Manager) createDefaultConfig() error {
	defaultConfig := `# Alacritty Configuration
# Managed by alacritty-colors - theme imported from themes/current.toml

[general]
import = ["themes/current.toml"]

# Personal configuration below - will be preserved when switching themes

[window]
# Window configuration
decorations = "full"
# title = "Alacritty"
# dynamic_title = true

[window.padding]
# Window padding
x = 5
y = 5

[scrolling]
# Scrollback configuration  
history = 10000

[font]
# Font configuration
size = 12.0

[font.normal]
family = "monospace"
# family = "JetBrains Mono"

[cursor]
# Cursor configuration
style = "Block"

[selection]
save_to_clipboard = true

[keyboard.bindings]
# Key bindings
{ key = "V", mods = "Control|Shift", action = "Paste" }
{ key = "C", mods = "Control|Shift", action = "Copy" }
{ key = "Key0", mods = "Control", action = "ResetFontSize" }
{ key = "Equals", mods = "Control", action = "IncreaseFontSize" }
{ key = "Minus", mods = "Control", action = "DecreaseFontSize" }
`

	return os.WriteFile(m.config.ConfigFile, []byte(defaultConfig), 0644)
}

func (m *Manager) hasImportLine() bool {
	file, err := os.Open(m.config.ConfigFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "themes/current.toml") {
			return true
		}
	}
	return false
}

func (m *Manager) addImportLine() error {
	data, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		return err
	}

	// Add the import line at the beginning after any initial comments
	lines := strings.Split(string(data), "\n")
	var newLines []string

	// Keep initial comments
	i := 0
	for i < len(lines) && (strings.HasPrefix(strings.TrimSpace(lines[i]), "#") || strings.TrimSpace(lines[i]) == "") {
		newLines = append(newLines, lines[i])
		i++
	}

	// Add [general] section and import line
	newLines = append(newLines, "")
	newLines = append(newLines, "[general]")
	newLines = append(newLines, "import = [\"themes/current.toml\"]")
	newLines = append(newLines, "")

	// Add rest of config
	newLines = append(newLines, lines[i:]...)

	return os.WriteFile(m.config.ConfigFile, []byte(strings.Join(newLines, "\n")), 0644)
}

func (m *Manager) ApplyTheme(themeName string) error {
	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	var selectedTheme *ThemeInfo
	for _, theme := range themes {
		if strings.EqualFold(theme.Name, themeName) {
			selectedTheme = &theme
			break
		}
	}

	if selectedTheme == nil {
		return fmt.Errorf("theme '%s' not found", themeName)
	}

	ui.PrintInfo("Applying theme: %s", selectedTheme.Name)

	// Create backup
	if err := m.CreateBackup(); err != nil {
		ui.PrintWarning("Failed to create backup: %v", err)
	}

	// Copy theme to current.toml
	currentThemePath := filepath.Join(m.config.ThemesDir, "current.toml")
	if err := m.copyFile(selectedTheme.FilePath, currentThemePath); err != nil {
		return fmt.Errorf("failed to apply theme: %w", err)
	}

	// Update config to track current theme
	if err := m.config.SetCurrentTheme(selectedTheme.Name); err != nil {
		ui.PrintWarning("Failed to update theme tracking: %v", err)
	}

	ui.PrintSuccess("Applied theme '%s'", selectedTheme.Name)
	return nil
}

func (m *Manager) RandomTheme() error {
	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	if len(themes) == 0 {
		return fmt.Errorf("no themes available")
	}

	selectedTheme := themes[randomInt(len(themes))]
	ui.PrintInfo("Selected random theme: %s", selectedTheme.Name)

	return m.ApplyTheme(selectedTheme.Name)
}

func (m *Manager) ListThemes(format string) error {
	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	if len(themes) == 0 {
		ui.PrintWarning("No themes found")
		ui.PrintInfo("Run 'alacritty-colors init' to download themes")
		return nil
	}

	switch format {
	case "grid":
		m.printThemeGrid(themes)
	case "list":
		m.printThemeList(themes)
	case "json":
		m.printThemeJSON(themes)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	return nil
}

func (m *Manager) SearchThemes(query string) error {
	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	query = strings.ToLower(query)
	var matches []ThemeInfo

	for _, theme := range themes {
		if strings.Contains(strings.ToLower(theme.Name), query) ||
			strings.Contains(strings.ToLower(theme.Description), query) ||
			strings.Contains(strings.ToLower(theme.Author), query) {
			matches = append(matches, theme)
		}
	}

	if len(matches) == 0 {
		ui.PrintWarning("No themes found matching '%s'", query)
		return nil
	}

	ui.PrintHeader(fmt.Sprintf("Search Results for '%s' (%d found)", query, len(matches)))
	for _, theme := range matches {
		description := theme.Description
		if description == "" && theme.Author != "" {
			description = fmt.Sprintf("by %s", theme.Author)
		}
		ui.PrintTheme(theme.Name, description)
	}

	return nil
}

func (m *Manager) PreviewTheme(themeName string) error {
	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	var selectedTheme *ThemeInfo
	for _, theme := range themes {
		if strings.EqualFold(theme.Name, themeName) {
			selectedTheme = &theme
			break
		}
	}

	if selectedTheme == nil {
		return fmt.Errorf("theme '%s' not found", themeName)
	}

	ui.PrintHeader(fmt.Sprintf("Theme Preview: %s", selectedTheme.Name))

	if selectedTheme.Description != "" {
		ui.PrintInfo("Description: %s", selectedTheme.Description)
	}
	if selectedTheme.Author != "" {
		ui.PrintInfo("Author: %s", selectedTheme.Author)
	}

	ui.PrintSubHeader("Color Palette")

	// Group colors by type
	primary := make(map[string]string)
	normal := make(map[string]string)
	bright := make(map[string]string)

	for key, value := range selectedTheme.Colors {
		if strings.HasPrefix(key, "bright_") {
			bright[strings.TrimPrefix(key, "bright_")] = value
		} else if strings.HasPrefix(key, "normal_") {
			normal[strings.TrimPrefix(key, "normal_")] = value
		} else {
			primary[key] = value
		}
	}

	// Display colors
	if len(primary) > 0 {
		fmt.Println("\nPrimary Colors:")
		for name, color := range primary {
			ui.PrintColorPreview(name, color)
		}
	}

	if len(normal) > 0 {
		fmt.Println("\nNormal Colors:")
		colorOrder := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
		for _, name := range colorOrder {
			if color, exists := normal[name]; exists {
				ui.PrintColorPreview(name, color)
			}
		}
	}

	if len(bright) > 0 {
		fmt.Println("\nBright Colors:")
		colorOrder := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
		for _, name := range colorOrder {
			if color, exists := bright[name]; exists {
				ui.PrintColorPreview("bright_"+name, color)
			}
		}
	}

	fmt.Println()
	if ui.PromptConfirm("Apply this theme?") {
		return m.ApplyTheme(selectedTheme.Name)
	}

	return nil
}

func (m *Manager) CreateBackup() error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFile := filepath.Join(m.config.BackupDir, fmt.Sprintf("alacritty_%s.toml", timestamp))

	src, err := os.Open(m.config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy config: %w", err)
	}

	ui.PrintSuccess("Backup created: %s", filepath.Base(backupFile))
	return nil
}

func (m *Manager) RestoreBackup(backupFile string) error {
	if backupFile == "" {
		// List available backups and let user choose
		return m.interactiveRestore()
	}

	// If backupFile is just a filename, look in backup directory
	if !filepath.IsAbs(backupFile) {
		backupFile = filepath.Join(m.config.BackupDir, backupFile)
	}

	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	ui.PrintInfo("Restoring from backup: %s", filepath.Base(backupFile))

	src, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(m.config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	ui.PrintSuccess("Configuration restored from backup")
	return nil
}

func (m *Manager) UpdateThemes() error {
	ui.PrintSubHeader("Updating theme database")

	dl := downloader.New(m.config.ThemesDir)
	count, err := dl.DownloadOfficialThemes()
	if err != nil {
		return fmt.Errorf("failed to update themes: %w", err)
	}

	ui.PrintSuccess("Updated theme database (%d themes)", count)
	return nil
}

func (m *Manager) GetCurrentTheme() string {
	return m.config.CurrentTheme
}

func (m *Manager) ShowCurrentTheme() error {
	currentTheme := m.GetCurrentTheme()
	if currentTheme == "" {
		ui.PrintInfo("No theme currently applied")
	} else {
		ui.PrintSuccess("Current theme: %s", currentTheme)
	}
	return nil
}

func (m *Manager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (m *Manager) getThemeInfos() ([]ThemeInfo, error) {
	files, err := m.getThemeFiles()
	if err != nil {
		return nil, err
	}

	var themes []ThemeInfo
	for _, file := range files {
		// Skip current.toml as it's not a real theme
		if filepath.Base(file) == "current.toml" {
			continue
		}

		info, err := m.parseThemeFile(file)
		if err != nil {
			ui.PrintWarning("Failed to parse theme %s: %v", filepath.Base(file), err)
			continue
		}
		themes = append(themes, info)
	}

	// Sort themes by name
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Name < themes[j].Name
	})

	return themes, nil
}

func (m *Manager) getThemeFiles() ([]string, error) {
	if _, err := os.Stat(m.config.ThemesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("themes directory not found: %s", m.config.ThemesDir)
	}

	files, err := os.ReadDir(m.config.ThemesDir)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".toml") {
			themes = append(themes, filepath.Join(m.config.ThemesDir, file.Name()))
		}
	}

	return themes, nil
}

func (m *Manager) parseThemeFile(filePath string) (ThemeInfo, error) {
	name := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	info := ThemeInfo{
		Name:     name,
		FilePath: filePath,
		Colors:   make(map[string]string),
	}

	file, err := os.Open(filePath)
	if err != nil {
		return info, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inColors := false
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			// Extract metadata from comments
			if strings.HasPrefix(line, "# Author:") {
				info.Author = strings.TrimSpace(strings.TrimPrefix(line, "# Author:"))
			} else if strings.HasPrefix(line, "# Description:") {
				info.Description = strings.TrimSpace(strings.TrimPrefix(line, "# Description:"))
			}
			continue
		}

		// Check for color sections
		if strings.HasPrefix(line, "[colors") {
			inColors = true
			if strings.Contains(line, "primary") {
				currentSection = "primary"
			} else if strings.Contains(line, "normal") {
				currentSection = "normal"
			} else if strings.Contains(line, "bright") {
				currentSection = "bright"
			}
			continue
		}

		// Check for other sections
		if strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[colors") {
			inColors = false
			continue
		}

		// Parse color values
		if inColors && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

				// Create full key with section prefix
				fullKey := key
				if currentSection != "" && currentSection != "primary" {
					fullKey = currentSection + "_" + key
				}

				info.Colors[fullKey] = value
			}
		}
	}

	return info, scanner.Err()
}

func (m *Manager) printThemeGrid(themes []ThemeInfo) {
	ui.PrintHeader(fmt.Sprintf("Available Themes (%d)", len(themes)))

	// Group themes by first letter
	grouped := make(map[string][]ThemeInfo)
	for _, theme := range themes {
		firstChar := strings.ToUpper(string(theme.Name[0]))
		grouped[firstChar] = append(grouped[firstChar], theme)
	}

	// Sort groups
	var keys []string
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ui.PrintSubHeader(fmt.Sprintf("Themes starting with '%s'", key))

		names := make([]string, len(grouped[key]))
		for i, theme := range grouped[key] {
			names[i] = theme.Name
		}
		ui.PrintThemeGrid(names, 3)
	}
}

func (m *Manager) printThemeList(themes []ThemeInfo) {
	ui.PrintHeader(fmt.Sprintf("Available Themes (%d)", len(themes)))

	for _, theme := range themes {
		description := theme.Description
		if description == "" && theme.Author != "" {
			description = fmt.Sprintf("by %s", theme.Author)
		}
		ui.PrintTheme(theme.Name, description)
	}
}

func (m *Manager) printThemeJSON(themes []ThemeInfo) {
	fmt.Println("[")
	for i, theme := range themes {
		fmt.Printf(`  {
    "name": "%s",
    "description": "%s",
    "author": "%s",
    "file": "%s"
  }`, theme.Name, theme.Description, theme.Author, theme.FilePath)

		if i < len(themes)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("]")
}

func (m *Manager) interactiveRestore() error {
	files, err := os.ReadDir(m.config.BackupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".toml") {
			backups = append(backups, file.Name())
		}
	}

	if len(backups) == 0 {
		ui.PrintWarning("No backup files found")
		return nil
	}

	ui.PrintHeader("Available Backups")
	for i, backup := range backups {
		ui.PrintInfo("%d. %s", i+1, backup)
	}

	fmt.Print("\nSelect backup to restore (number): ")
	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		return fmt.Errorf("invalid input")
	}

	if choice < 1 || choice > len(backups) {
		return fmt.Errorf("invalid selection")
	}

	selectedBackup := backups[choice-1]
	if !ui.PromptConfirm(fmt.Sprintf("Restore from '%s'?", selectedBackup)) {
		ui.PrintInfo("Restore cancelled")
		return nil
	}

	return m.RestoreBackup(selectedBackup)
}
