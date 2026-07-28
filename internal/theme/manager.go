package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vitruves/alacritty-colors/internal/config"
)

// ThemeImportPath is the file Alacritty imports to pick up the live theme.
// The path is relative to the main config file's directory.
const ThemeImportPath = "themes/current.toml"

// trackingName is the file recording which named theme is live. It is deleted
// whenever the live colours stop matching a named theme.
const trackingName = ".current-theme"

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
	trackingFile := m.config.GetThemePath(trackingName)
	if err := os.WriteFile(trackingFile, []byte(themeName), 0644); err != nil {
		// Don't fail if we can't write tracking file, just continue
	}

	// Ensure main config has import line
	if err := m.ensureImportLine(); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	return nil
}

// WritePreview pushes colors straight into themes/current.toml so Alacritty
// repaints, without touching the theme file they came from. This is what live
// editing uses: the source theme stays pristine until an explicit save.
func (m *Manager) WritePreview(content string) error {
	currentThemePath := m.config.GetThemePath("current")
	if err := os.WriteFile(currentThemePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write preview: %w", err)
	}

	// The live colours are edits that belong to no named theme, so the tracking
	// file would now be a lie. Drop it: GetCurrentTheme then falls back to
	// comparing content and reports a theme only when one really is live.
	// Without this the editor reopens claiming a theme the terminal is not
	// showing, and every swatch on screen disagrees with the actual window.
	if err := os.Remove(m.config.GetThemePath(trackingName)); err != nil && !os.IsNotExist(err) {
		// A stale marker is not worth failing the preview over.
		_ = err
	}

	return m.ensureImportLine()
}

func (m *Manager) GetCurrentTheme() string {
	currentThemePath := m.config.GetThemePath("current")

	// If current.toml doesn't exist, no theme is applied
	if _, err := os.Stat(currentThemePath); os.IsNotExist(err) {
		return ""
	}

	// First check tracking file for applied theme name
	trackingFile := m.config.GetThemePath(trackingName)
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

// ensureImportLine guarantees the live theme is imported by the main config.
//
// The cycle runs through config.UpdateConfigFile so it is serialised against
// the font manager, which edits the same file, and so the replacement is
// atomic. Two unsynchronised read-modify-write cycles on a user's config is how
// you delete it.
func (m *Manager) ensureImportLine() error {
	if _, err := os.Stat(m.config.ConfigFile); os.IsNotExist(err) {
		return m.createMinimalConfig()
	}

	return config.UpdateConfigFile(m.config.ConfigFile, func(content string) (string, error) {
		updated, _, err := ensureImportBlock(content)
		return updated, err
	})
}

func (m *Manager) createMinimalConfig() error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(m.config.ConfigFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Create minimal config with just the import line
	content := `# Minimal Alacritty configuration
# Theme colors are imported from themes/current.toml

[general]
import = ["` + ThemeImportPath + `"]
`

	return os.WriteFile(m.config.ConfigFile, []byte(content), 0644)
}

var (
	// sectionRe matches a table header, including the [[array]] form so that
	// entries inside one are never mistaken for keys of the table before it.
	sectionRe = regexp.MustCompile(`^\s*\[\[?([^\[\]]+)\]\]?\s*$`)
	// importArrayRe matches a single-line import array and captures its body.
	importArrayRe = regexp.MustCompile(`^\s*import\s*=\s*\[([^\]]*)\]\s*$`)
	// importKeyRe matches any import assignment, single-line or not.
	importKeyRe = regexp.MustCompile(`^\s*import\s*=`)
)

// ensureImportBlock returns config content that imports the live theme file
// from under [general], and reports whether it had to change anything.
//
// Alacritty moved import out of the document root in 0.13; a bare top-level
// `import` is an unknown key to every release since, so a theme written that
// way never reaches the terminal. Any legacy top-level import found here is
// migrated rather than duplicated, because two import keys are a TOML error.
func ensureImportBlock(content string) (string, bool, error) {
	lines := strings.Split(content, "\n")

	section := ""
	generalHeader := -1  // line index of the [general] header
	generalImport := -1  // line index of an import inside [general]
	rootImport := -1     // line index of a legacy top-level import
	var carried []string // paths to preserve from wherever import lived

	for i, line := range lines {
		if matches := sectionRe.FindStringSubmatch(line); matches != nil {
			section = strings.TrimSpace(matches[1])
			if section == "general" && generalHeader < 0 {
				generalHeader = i
			}
			continue
		}
		if !importKeyRe.MatchString(line) {
			continue
		}

		matches := importArrayRe.FindStringSubmatch(line)
		if matches == nil {
			// A multi-line array. Rewriting it blind risks corrupting a working
			// config, so say what is missing and let the user place it.
			return content, false, fmt.Errorf(
				"config has a multi-line import list; add %q to it by hand", ThemeImportPath)
		}

		paths := parseImportPaths(matches[1])
		switch section {
		case "general":
			generalImport = i
			carried = append(carried, paths...)
		case "":
			rootImport = i
			carried = append(carried, paths...)
		}
	}

	// Already importing the live theme from the right place: nothing to do.
	if generalImport >= 0 && rootImport < 0 && containsPath(carried, ThemeImportPath) {
		return content, false, nil
	}

	importLine := "import = [" + quotePaths(appendPath(carried, ThemeImportPath)) + "]"

	switch {
	case generalImport >= 0:
		lines[generalImport] = importLine
		if rootImport >= 0 {
			lines = removeLine(lines, rootImport)
		}
	case rootImport >= 0 && generalHeader >= 0:
		lines = insertLine(lines, generalHeader+1, importLine)
		if rootImport > generalHeader {
			lines = removeLine(lines, rootImport+1)
		} else {
			lines = removeLine(lines, rootImport)
		}
	case rootImport >= 0:
		// Reuse the legacy line's position so the import stays at the top.
		lines[rootImport] = "[general]\n" + importLine
	case generalHeader >= 0:
		lines = insertLine(lines, generalHeader+1, importLine)
	default:
		// A fresh [general] table at the end is always valid TOML and, unlike a
		// block prepended to the file, cannot re-parent existing root keys.
		block := "[general]\n" + importLine + "\n"
		trimmed := strings.TrimRight(strings.Join(lines, "\n"), "\n")
		if trimmed == "" {
			return block, true, nil
		}
		return trimmed + "\n\n" + block, true, nil
	}

	return strings.Join(lines, "\n"), true, nil
}

// parseImportPaths splits the body of an import array into bare paths.
func parseImportPaths(body string) []string {
	var paths []string
	for _, field := range strings.Split(body, ",") {
		field = strings.Trim(strings.TrimSpace(field), `"'`)
		if field != "" {
			paths = append(paths, field)
		}
	}
	return paths
}

func containsPath(paths []string, want string) bool {
	for _, path := range paths {
		if path == want {
			return true
		}
	}
	return false
}

func appendPath(paths []string, want string) []string {
	if containsPath(paths, want) {
		return paths
	}
	return append(paths, want)
}

func quotePaths(paths []string) string {
	quoted := make([]string, len(paths))
	for i, path := range paths {
		quoted[i] = `"` + path + `"`
	}
	return strings.Join(quoted, ", ")
}

func insertLine(lines []string, at int, line string) []string {
	lines = append(lines, "")
	copy(lines[at+1:], lines[at:])
	lines[at] = line
	return lines
}

func removeLine(lines []string, at int) []string {
	return append(lines[:at], lines[at+1:]...)
}
