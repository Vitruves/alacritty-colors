package theme

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
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

// Option structs for enhanced functionality
type ApplyOptions struct {
	WithFont   bool
	Opacity    float64
	Blur       float64
	FontSize   float64
	FontFamily string
}

type ListOptions struct {
	Format     string
	ShowColors bool
	DarkOnly   bool
	LightOnly  bool
}

type RandomOptions struct {
	DarkOnly  bool
	LightOnly bool
	WithFont  bool
	Opacity   float64
	Blur      float64
	Scheme    string
}

type GenerateOptions struct {
	Scheme     string
	Name       string
	Save       bool
	DarkTheme  bool
	LightTheme bool
	WithFont   bool
	Opacity    float64
	Blur       float64
}

type SearchOptions struct {
	Format     string
	ShowColors bool
}

type PreviewOptions struct {
	AutoApply bool
	ShowHex   bool
}

type BackupOptions struct {
	Name        string
	Description string
}

type RestoreOptions struct {
	Interactive bool
}

type UpdateOptions struct {
	Force bool
	Check bool
}

type Manager struct {
	config  *config.Config
	verbose bool
}

type ThemeInfo struct {
	Name        string
	FilePath    string
	Description string
	Author      string
	Tags        []string
	Colors      map[string]string
	IsDark      bool
	IsLight     bool
}

// Font definitions for automatic pairing
var ThemeFonts = map[string][]string{
	"cyberpunk": {"JetBrains Mono", "Fira Code", "Source Code Pro"},
	"dracula":   {"Fira Code", "JetBrains Mono", "Cascadia Code"},
	"nord":      {"JetBrains Mono", "IBM Plex Mono", "SF Mono"},
	"gruvbox":   {"Fira Code", "Hack", "Inconsolata"},
	"solarized": {"Source Code Pro", "IBM Plex Mono", "DejaVu Sans Mono"},
	"default":   {"JetBrains Mono", "Fira Code", "monospace"},
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{config: cfg, verbose: false}
}

func (m *Manager) SetVerbose(verbose bool) {
	m.verbose = verbose
}

func (m *Manager) logVerbose(format string, args ...interface{}) {
	if m.verbose {
		ui.PrintVerbose(format, args...)
	}
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
	opts := &PreviewOptions{
		AutoApply: false,
		ShowHex:   false,
	}
	return m.PreviewThemeWithOptions(themeName, opts)
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

func (m *Manager) ApplyThemeWithOptions(themeName string, opts *ApplyOptions) error {
	m.logVerbose("Applying theme %s with options", themeName)

	if err := m.ApplyTheme(themeName); err != nil {
		return err
	}

	// Apply additional options
	if opts != nil {
		if opts.WithFont {
			if err := m.applyThemeFont(themeName, opts.FontFamily, opts.FontSize); err != nil {
				ui.PrintWarning("Failed to set font: %v", err)
			}
		}

		if opts.Opacity > 0 || opts.Blur > 0 {
			if err := m.applyVisualEffects(opts.Opacity, opts.Blur); err != nil {
				ui.PrintWarning("Failed to apply visual effects: %v", err)
			}
		}
	}

	return nil
}

func (m *Manager) ListThemesWithOptions(opts *ListOptions) error {
	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	// Apply filters
	if opts.DarkOnly {
		themes = m.filterDarkThemes(themes)
	} else if opts.LightOnly {
		themes = m.filterLightThemes(themes)
	}

	m.logVerbose("Found %d themes after filtering", len(themes))

	switch opts.Format {
	case "grid":
		m.printThemeGrid(themes)
	case "list":
		m.printThemeList(themes)
	case "json":
		m.printThemeJSON(themes)
	case "colors":
		m.printThemeColors(themes)
	default:
		m.printThemeGrid(themes)
	}

	return nil
}

func (m *Manager) RandomThemeWithOptions(opts *RandomOptions) error {
	m.logVerbose("Selecting random theme with constraints")

	// If scheme is specified, generate new theme instead
	if opts.Scheme != "" {
		genOpts := &GenerateOptions{
			Scheme:     opts.Scheme,
			DarkTheme:  opts.DarkOnly,
			LightTheme: opts.LightOnly,
			WithFont:   opts.WithFont,
			Opacity:    opts.Opacity,
			Blur:       opts.Blur,
		}
		return m.GenerateThemeWithOptions(genOpts)
	}

	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	// Apply filters
	if opts.DarkOnly {
		themes = m.filterDarkThemes(themes)
	} else if opts.LightOnly {
		themes = m.filterLightThemes(themes)
	}

	if len(themes) == 0 {
		return fmt.Errorf("no themes found matching criteria")
	}

	// Select random theme
	rand.Seed(time.Now().UnixNano())
	selectedTheme := themes[rand.Intn(len(themes))]

	m.logVerbose("Selected random theme: %s", selectedTheme.Name)

	applyOpts := &ApplyOptions{
		WithFont: opts.WithFont,
		Opacity:  opts.Opacity,
		Blur:     opts.Blur,
	}

	return m.ApplyThemeWithOptions(selectedTheme.Name, applyOpts)
}

func (m *Manager) GenerateThemeWithOptions(opts *GenerateOptions) error {
	colors, err := m.generateColorSchemeWithVariant(opts.Scheme, opts.DarkTheme, opts.LightTheme)
	if err != nil {
		return fmt.Errorf("failed to generate colors: %w", err)
	}

	name := opts.Name
	if name == "" {
		variant := ""
		if opts.DarkTheme {
			variant = "_dark"
		} else if opts.LightTheme {
			variant = "_light"
		}
		name = generateRandomName(opts.Scheme + variant)
	}

	themeContent := m.createThemeContent(colors, opts.Scheme, name)

	if opts.Save {
		themeFile := filepath.Join(m.config.ThemesDir, name+".toml")
		if err := os.WriteFile(themeFile, []byte(themeContent), 0644); err != nil {
			return fmt.Errorf("failed to save theme: %w", err)
		}
		ui.PrintSuccess("Generated theme saved: %s", name)
	}

	// Apply the theme
	if err := m.ApplyTheme(name); err != nil {
		return fmt.Errorf("failed to apply generated theme: %w", err)
	}

	// Apply additional options
	if opts.WithFont {
		if err := m.applyThemeFont(opts.Scheme, "", 0); err != nil {
			ui.PrintWarning("Failed to set font: %v", err)
		}
	}

	if opts.Opacity > 0 || opts.Blur > 0 {
		if err := m.applyVisualEffects(opts.Opacity, opts.Blur); err != nil {
			ui.PrintWarning("Failed to apply visual effects: %v", err)
		}
	}

	return nil
}

func (m *Manager) SearchThemesWithOptions(query string, opts *SearchOptions) error {
	themes, err := m.getThemeInfos()
	if err != nil {
		return err
	}

	// Filter themes based on query
	var matches []ThemeInfo
	queryLower := strings.ToLower(query)

	for _, theme := range themes {
		if m.matchesQuery(theme, queryLower) {
			matches = append(matches, theme)
		}
	}

	m.logVerbose("Found %d themes matching '%s'", len(matches), query)

	if len(matches) == 0 {
		ui.PrintWarning("No themes found matching '%s'", query)
		return nil
	}

	switch opts.Format {
	case "grid":
		m.printThemeGrid(matches)
	case "list":
		m.printThemeList(matches)
	case "colors":
		m.printThemeColors(matches)
	default:
		m.printThemeList(matches)
	}

	return nil
}

func (m *Manager) PreviewThemeWithOptions(themeName string, opts *PreviewOptions) error {
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

	// Save current theme state for restoration
	currentThemePath := filepath.Join(m.config.ThemesDir, "current.toml")
	backupThemePath := filepath.Join(m.config.ThemesDir, "preview_backup.toml")

	// Create backup of current theme
	if _, err := os.Stat(currentThemePath); err == nil {
		if err := m.copyFile(currentThemePath, backupThemePath); err != nil {
			return fmt.Errorf("failed to backup current theme: %w", err)
		}
	}

	// Temporarily apply the preview theme
	m.logVerbose("Temporarily applying theme for preview: %s", selectedTheme.Name)
	if err := m.copyFile(selectedTheme.FilePath, currentThemePath); err != nil {
		return fmt.Errorf("failed to apply preview theme: %w", err)
	}

	// Show theme information
	ui.PrintHeader(fmt.Sprintf("🎨 Theme Preview: %s", selectedTheme.Name))
	ui.PrintInfo("The theme is now temporarily applied to your terminal!")

	if selectedTheme.Description != "" {
		ui.PrintInfo("Description: %s", selectedTheme.Description)
	}
	if selectedTheme.Author != "" {
		ui.PrintInfo("Author: %s", selectedTheme.Author)
	}

	// Show color palette if requested
	if opts.ShowHex {
		m.printThemePreview(*selectedTheme, true)
	}

	var keepTheme bool

	if opts.AutoApply {
		keepTheme = true
		ui.PrintSuccess("Auto-applying theme: %s", selectedTheme.Name)
	} else {
		// Interactive prompt with visual preview
		ui.PrintInfo("\nYou can now see how the theme looks in your terminal.")
		ui.PrintInfo("Test it by running some commands or checking your editor.")
		fmt.Println()

		keepTheme = ui.PromptConfirm("Do you want to keep this theme?")
	}

	if keepTheme {
		// User wants to keep the theme - update tracking
		if err := m.config.SetCurrentTheme(selectedTheme.Name); err != nil {
			ui.PrintWarning("Failed to update theme tracking: %v", err)
		}
		ui.PrintSuccess("Applied theme: %s", selectedTheme.Name)

		// Clean up backup
		os.Remove(backupThemePath)
	} else {
		// User wants to restore previous theme
		ui.PrintInfo("Restoring previous theme...")

		if _, err := os.Stat(backupThemePath); err == nil {
			if err := m.copyFile(backupThemePath, currentThemePath); err != nil {
				ui.PrintError("Failed to restore previous theme: %v", err)
				return err
			}
			os.Remove(backupThemePath)
			ui.PrintSuccess("Previous theme restored")
		} else {
			// No backup exists, create empty current.toml
			defaultTheme := `# No theme applied
# Run 'alacritty-colors apply <theme-name>' to apply a theme

[colors.primary]
background = "#1e1e1e"
foreground = "#ffffff"
`
			os.WriteFile(currentThemePath, []byte(defaultTheme), 0644)
			ui.PrintSuccess("Reset to default theme")
		}
	}

	return nil
}

func (m *Manager) CreateBackupWithOptions(opts *BackupOptions) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	var backupName string
	if opts.Name != "" {
		backupName = fmt.Sprintf("%s_%s.toml", opts.Name, timestamp)
	} else {
		backupName = fmt.Sprintf("alacritty_%s.toml", timestamp)
	}

	backupPath := filepath.Join(m.config.BackupDir, backupName)

	m.logVerbose("Creating backup: %s", backupPath)

	if err := m.copyFile(m.config.ConfigFile, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	ui.PrintSuccess("Backup created: %s", backupName)

	if opts.Description != "" {
		// Create a companion .info file with description
		infoPath := strings.TrimSuffix(backupPath, ".toml") + ".info"
		infoContent := fmt.Sprintf("Description: %s\nCreated: %s\n", opts.Description, time.Now().Format("2006-01-02 15:04:05"))
		os.WriteFile(infoPath, []byte(infoContent), 0644)
	}

	return nil
}

func (m *Manager) RestoreBackupWithOptions(backupFile string, opts *RestoreOptions) error {
	if opts.Interactive || backupFile == "" {
		return m.interactiveRestore()
	}

	m.logVerbose("Restoring from backup: %s", backupFile)
	return m.RestoreBackup(backupFile)
}

func (m *Manager) UpdateThemesWithOptions(opts *UpdateOptions) error {
	if opts.Check {
		ui.PrintInfo("Checking for theme updates...")
		// This would check remote repository for updates
		ui.PrintInfo("Update check functionality not yet implemented")
		return nil
	}

	m.logVerbose("Updating themes (force: %v)", opts.Force)

	dl := downloader.New(m.config.ThemesDir)

	if opts.Force {
		// Remove existing themes before downloading
		ui.PrintInfo("Force update: removing existing themes")
		files, _ := filepath.Glob(filepath.Join(m.config.ThemesDir, "*.toml"))
		for _, file := range files {
			if !strings.HasSuffix(file, "current.toml") {
				os.Remove(file)
			}
		}
	}

	count, err := dl.DownloadOfficialThemes()
	if err != nil {
		return fmt.Errorf("failed to update themes: %w", err)
	}

	ui.PrintSuccess("Updated %d themes", count)
	return nil
}

func (m *Manager) ListBackups() error {
	files, err := filepath.Glob(filepath.Join(m.config.BackupDir, "*.toml"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		ui.PrintInfo("No backups found")
		return nil
	}

	ui.PrintHeader("Available Backups")
	for i, file := range files {
		name := filepath.Base(file)
		stat, _ := os.Stat(file)

		// Check for description file
		infoFile := strings.TrimSuffix(file, ".toml") + ".info"
		description := ""
		if content, err := os.ReadFile(infoFile); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Description: ") {
					description = strings.TrimPrefix(line, "Description: ")
					break
				}
			}
		}

		ui.PrintInfo("[%d] %s", i+1, name)
		ui.PrintInfo("    Created: %s", stat.ModTime().Format("2006-01-02 15:04:05"))
		if description != "" {
			ui.PrintInfo("    Description: %s", description)
		}
		fmt.Println()
	}

	return nil
}

func (m *Manager) ShowConfig() error {
	ui.PrintHeader("Alacritty Colors Configuration")

	ui.PrintKeyValue("Config File", m.config.ConfigFile)
	ui.PrintKeyValue("Themes Dir", m.config.ThemesDir)
	ui.PrintKeyValue("Backup Dir", m.config.BackupDir)

	// Show current theme
	current := m.GetCurrentTheme()
	if current != "" {
		ui.PrintKeyValue("Current Theme", current)
	} else {
		ui.PrintKeyValue("Current Theme", "None")
	}

	// Show statistics
	themes, _ := m.getThemeInfos()
	ui.PrintKeyValue("Available Themes", fmt.Sprintf("%d", len(themes)))

	backups, _ := filepath.Glob(filepath.Join(m.config.BackupDir, "*.toml"))
	ui.PrintKeyValue("Backups", fmt.Sprintf("%d", len(backups)))

	return nil
}

// Helper methods for enhanced functionality

func (m *Manager) applyThemeFont(themeName, fontFamily string, fontSize float64) error {
	m.logVerbose("Applying font settings for theme: %s", themeName)

	var selectedFont string

	// Determine font based on theme name or use provided fontFamily
	if fontFamily != "" {
		selectedFont = fontFamily
	} else {
		// Auto-select font based on theme
		themeKey := strings.ToLower(themeName)
		for key, fonts := range ThemeFonts {
			if strings.Contains(themeKey, key) {
				selectedFont = fonts[0] // Use first font in the list
				break
			}
		}
		if selectedFont == "" {
			selectedFont = ThemeFonts["default"][0]
		}
	}

	m.logVerbose("Selected font: %s", selectedFont)

	// Update Alacritty config with font settings
	return m.updateConfigFont(selectedFont, fontSize)
}

func (m *Manager) applyVisualEffects(opacity, blur float64) error {
	m.logVerbose("Applying visual effects: opacity=%.2f, blur=%.2f", opacity, blur)

	// Update Alacritty config with visual effects
	return m.updateConfigVisualEffects(opacity, blur)
}

func (m *Manager) filterDarkThemes(themes []ThemeInfo) []ThemeInfo {
	var darkThemes []ThemeInfo
	for _, theme := range themes {
		if m.isThemeDark(theme) {
			theme.IsDark = true
			darkThemes = append(darkThemes, theme)
		}
	}
	return darkThemes
}

func (m *Manager) filterLightThemes(themes []ThemeInfo) []ThemeInfo {
	var lightThemes []ThemeInfo
	for _, theme := range themes {
		if !m.isThemeDark(theme) {
			theme.IsLight = true
			lightThemes = append(lightThemes, theme)
		}
	}
	return lightThemes
}

func (m *Manager) isThemeDark(theme ThemeInfo) bool {
	// Analyze background color to determine if theme is dark
	if bg, exists := theme.Colors["background"]; exists {
		// Convert hex to brightness value
		if len(bg) >= 7 && bg[0] == '#' {
			// Simple brightness calculation based on background color
			r, g, b := hexToRGB(bg)
			brightness := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0
			return brightness < 0.5
		}
	}

	// Check theme name for dark indicators
	nameLower := strings.ToLower(theme.Name)
	darkIndicators := []string{"dark", "night", "black", "midnight", "shadow", "deep"}
	for _, indicator := range darkIndicators {
		if strings.Contains(nameLower, indicator) {
			return true
		}
	}

	return true // Default to dark if uncertain
}

func (m *Manager) convertToDarkVariant(colors map[string]string) map[string]string {
	darkColors := make(map[string]string)

	// Copy all colors
	for k, v := range colors {
		darkColors[k] = v
	}

	// Adjust background to be darker
	if bg, exists := colors["background"]; exists {
		r, g, b := hexToRGB(bg)
		// Make background darker
		r = int(float64(r) * 0.3)
		g = int(float64(g) * 0.3)
		b = int(float64(b) * 0.3)
		darkColors["background"] = rgbToHex(r, g, b)
	} else {
		darkColors["background"] = "#1a1a1a"
	}

	// Ensure bright foreground
	darkColors["foreground"] = "#e5e5e5"

	return darkColors
}

func (m *Manager) convertToLightVariant(colors map[string]string) map[string]string {
	lightColors := make(map[string]string)

	// Copy all colors
	for k, v := range colors {
		lightColors[k] = v
	}

	// Adjust background to be lighter
	if bg, exists := colors["background"]; exists {
		r, g, b := hexToRGB(bg)
		// Make background lighter
		r = 255 - int(float64(255-r)*0.1)
		g = 255 - int(float64(255-g)*0.1)
		b = 255 - int(float64(255-b)*0.1)
		lightColors["background"] = rgbToHex(r, g, b)
	} else {
		lightColors["background"] = "#f8f8f8"
	}

	// Ensure dark foreground for readability
	lightColors["foreground"] = "#2a2a2a"

	return lightColors
}

func (m *Manager) printThemeColors(themes []ThemeInfo) {
	ui.PrintHeader("Theme Colors")

	for _, theme := range themes {
		ui.PrintSubHeader(theme.Name)

		// Print primary colors
		if bg := theme.Colors["background"]; bg != "" {
			ui.PrintColorPreview("Background", bg)
		}
		if fg := theme.Colors["foreground"]; fg != "" {
			ui.PrintColorPreview("Foreground", fg)
		}

		// Print normal colors
		normalColors := []string{"red", "green", "yellow", "blue", "magenta", "cyan"}
		for _, color := range normalColors {
			if value := theme.Colors[color]; value != "" {
				ui.PrintColorPreview(strings.Title(color), value)
			}
		}

		fmt.Println()
	}
}

func (m *Manager) printThemePreview(theme ThemeInfo, showHex bool) {
	ui.PrintHeader(fmt.Sprintf("Theme Preview: %s", theme.Name))

	if theme.Description != "" {
		ui.PrintInfo("Description: %s", theme.Description)
	}
	if theme.Author != "" {
		ui.PrintInfo("Author: %s", theme.Author)
	}

	ui.PrintSubHeader("Color Palette")

	// Primary colors
	if bg := theme.Colors["background"]; bg != "" {
		ui.PrintColorPreview("Background", bg)
	}
	if fg := theme.Colors["foreground"]; fg != "" {
		ui.PrintColorPreview("Foreground", fg)
	}

	// Normal colors
	ui.PrintInfo("\nNormal Colors:")
	normalColors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	for _, color := range normalColors {
		if value := theme.Colors[color]; value != "" {
			if showHex {
				ui.PrintInfo("  %-8s %s", color, value)
			} else {
				ui.PrintColorPreview(color, value)
			}
		}
	}

	// Bright colors
	ui.PrintInfo("\nBright Colors:")
	for _, color := range normalColors {
		brightKey := "bright_" + color
		if value := theme.Colors[brightKey]; value != "" {
			if showHex {
				ui.PrintInfo("  %-8s %s", brightKey, value)
			} else {
				ui.PrintColorPreview(brightKey, value)
			}
		}
	}
}

func (m *Manager) matchesQuery(theme ThemeInfo, query string) bool {
	// Check name
	if strings.Contains(strings.ToLower(theme.Name), query) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(theme.Description), query) {
		return true
	}

	// Check tags
	for _, tag := range theme.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}

func (m *Manager) updateConfigFont(fontFamily string, fontSize float64) error {
	// Read current config
	content, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inFontSection := false
	inFontNormalSection := false
	fontSectionAdded := false
	fontFamilySet := false
	fontSizeSet := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track sections
		if strings.HasPrefix(trimmed, "[font]") {
			inFontSection = true
			inFontNormalSection = false
			fontSectionAdded = true
			newLines = append(newLines, line)
			continue
		} else if strings.HasPrefix(trimmed, "[font.normal]") {
			inFontNormalSection = true
			newLines = append(newLines, line)
			continue
		} else if strings.HasPrefix(trimmed, "[") {
			inFontSection = false
			inFontNormalSection = false
		}

		// Update font settings
		if inFontSection && strings.HasPrefix(trimmed, "size") && fontSize > 0 {
			newLines = append(newLines, fmt.Sprintf("size = %.1f", fontSize))
			fontSizeSet = true
			continue
		} else if inFontNormalSection && strings.HasPrefix(trimmed, "family") {
			newLines = append(newLines, fmt.Sprintf("family = \"%s\"", fontFamily))
			fontFamilySet = true
			continue
		}

		newLines = append(newLines, line)
	}

	// Add font section if not found
	if !fontSectionAdded {
		newLines = append(newLines, "")
		newLines = append(newLines, "[font]")
		if fontSize > 0 {
			newLines = append(newLines, fmt.Sprintf("size = %.1f", fontSize))
		}
		newLines = append(newLines, "")
		newLines = append(newLines, "[font.normal]")
		newLines = append(newLines, fmt.Sprintf("family = \"%s\"", fontFamily))
	} else {
		// Add missing settings
		if !fontSizeSet && fontSize > 0 {
			// Insert size after [font] section
			for i, line := range newLines {
				if strings.TrimSpace(line) == "[font]" {
					newLines = append(newLines[:i+1], append([]string{fmt.Sprintf("size = %.1f", fontSize)}, newLines[i+1:]...)...)
					break
				}
			}
		}
		if !fontFamilySet {
			// Add [font.normal] section if needed
			hasNormalSection := false
			for _, line := range newLines {
				if strings.TrimSpace(line) == "[font.normal]" {
					hasNormalSection = true
					break
				}
			}
			if !hasNormalSection {
				// Find end of font section and add normal section
				for i, line := range newLines {
					if strings.TrimSpace(line) == "[font]" {
						// Find where to insert
						j := i + 1
						for j < len(newLines) && !strings.HasPrefix(strings.TrimSpace(newLines[j]), "[") && strings.TrimSpace(newLines[j]) != "" {
							j++
						}
						insert := []string{"", "[font.normal]", fmt.Sprintf("family = \"%s\"", fontFamily)}
						newLines = append(newLines[:j], append(insert, newLines[j:]...)...)
						break
					}
				}
			}
		}
	}

	return os.WriteFile(m.config.ConfigFile, []byte(strings.Join(newLines, "\n")), 0644)
}

func (m *Manager) updateConfigVisualEffects(opacity, blur float64) error {
	// Read current config
	content, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inWindowSection := false
	windowSectionAdded := false
	opacitySet := false
	blurSet := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track window section
		if strings.HasPrefix(trimmed, "[window]") {
			inWindowSection = true
			windowSectionAdded = true
			newLines = append(newLines, line)
			continue
		} else if strings.HasPrefix(trimmed, "[") && trimmed != "[window]" {
			inWindowSection = false
		}

		// Update window settings
		if inWindowSection {
			if strings.HasPrefix(trimmed, "opacity") && opacity > 0 {
				newLines = append(newLines, fmt.Sprintf("opacity = %.2f", opacity))
				opacitySet = true
				continue
			} else if strings.HasPrefix(trimmed, "blur") && blur > 0 {
				newLines = append(newLines, fmt.Sprintf("blur = %.1f", blur))
				blurSet = true
				continue
			}
		}

		newLines = append(newLines, line)
	}

	// Add window section if not found
	if !windowSectionAdded {
		newLines = append(newLines, "")
		newLines = append(newLines, "[window]")
		if opacity > 0 {
			newLines = append(newLines, fmt.Sprintf("opacity = %.2f", opacity))
		}
		if blur > 0 {
			newLines = append(newLines, fmt.Sprintf("blur = %.1f", blur))
		}
	} else {
		// Add missing settings
		if !opacitySet && opacity > 0 {
			for i, line := range newLines {
				if strings.TrimSpace(line) == "[window]" {
					newLines = append(newLines[:i+1], append([]string{fmt.Sprintf("opacity = %.2f", opacity)}, newLines[i+1:]...)...)
					break
				}
			}
		}
		if !blurSet && blur > 0 {
			for i, line := range newLines {
				if strings.TrimSpace(line) == "[window]" {
					insertIndex := i + 1
					// Find where to insert (after other window settings)
					for insertIndex < len(newLines) && !strings.HasPrefix(strings.TrimSpace(newLines[insertIndex]), "[") && strings.TrimSpace(newLines[insertIndex]) != "" {
						insertIndex++
					}
					newLines = append(newLines[:insertIndex], append([]string{fmt.Sprintf("blur = %.1f", blur)}, newLines[insertIndex:]...)...)
					break
				}
			}
		}
	}

	return os.WriteFile(m.config.ConfigFile, []byte(strings.Join(newLines, "\n")), 0644)
}

// Utility functions for color conversion
func hexToRGB(hex string) (int, int, int) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0
	}

	var r, g, b int
	fmt.Sscanf(hex[1:3], "%x", &r)
	fmt.Sscanf(hex[3:5], "%x", &g)
	fmt.Sscanf(hex[5:7], "%x", &b)
	return r, g, b
}

func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("#%02x%02x%02x",
		max(0, min(255, r)),
		max(0, min(255, g)),
		max(0, min(255, b)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
