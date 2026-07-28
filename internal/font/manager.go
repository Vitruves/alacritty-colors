package font

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/vitruves/alacritty-colors/internal/config"
)

// FontConfig represents the font configuration section
type FontConfig struct {
	Normal struct {
		Family string `toml:"family"`
		Style  string `toml:"style"`
	} `toml:"normal"`
	Size              float64 `toml:"size"`
	BuiltinBoxDrawing bool    `toml:"builtin_box_drawing"`
}

// Manager handles font configuration operations
type Manager struct {
	config *config.Config

	// installed caches the platform font list. Scanning costs a few hundred
	// milliseconds, and the panel asks for families and then for the styles of
	// whichever family the cursor lands on, which would otherwise rescan on
	// every keystroke.
	mu        sync.Mutex
	installed map[string][]string
	scanned   bool
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// RefreshFonts drops the cached font list so the next lookup rescans. Call it
// after installing fonts, or the new ones stay invisible until restart.
func (m *Manager) RefreshFonts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.installed, m.scanned = nil, false
}

// systemFonts returns the platform's monospaced families and their styles, or
// nil when the platform has no better answer than fontconfig.
func (m *Manager) systemFonts() map[string][]string {
	if runtime.GOOS != "darwin" {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scanned {
		return m.installed
	}
	m.scanned = true

	families, err := coreTextFamilies()
	if err != nil {
		// Fall back to fontconfig rather than showing an empty browser.
		return nil
	}
	m.installed = families
	return m.installed
}

// isValidTerminalFont checks if a font name is suitable for terminal use
func isValidTerminalFont(fontName string) bool {
	if fontName == "" || len(fontName) > 60 {
		return false
	}

	// Skip fonts that start with dots (system fonts) unless they're known good ones
	if strings.HasPrefix(fontName, ".") && !strings.Contains(strings.ToLower(fontName), "mono") {
		return false
	}

	lowerName := strings.ToLower(fontName)

	// Check if the font name contains mostly Latin characters (but allow some special chars)
	latinCount := 0
	nonLatinCount := 0

	for _, r := range fontName {
		if unicode.IsLetter(r) {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				latinCount++
			} else {
				// Allow some European characters like ö, ü, etc.
				if r < 256 {
					latinCount++
				} else {
					nonLatinCount++
				}
			}
		}
	}

	// Skip fonts with only non-Latin characters
	if latinCount == 0 && nonLatinCount > 0 {
		return false
	}

	// Additional filters for known problematic patterns
	skipPatterns := []string{
		"emoji", "symbol", "icons", "wingdings", "webdings",
		"dingbats", "ornaments", "arabic", "hebrew", "thai",
		"chinese", "japanese", "korean", "hindi", "bengali",
		"tamil", "telugu", "kannada", "malayalam", "gujarati",
		"punjabi", "oriya", "assamese", "devanagari", "sinhala",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(lowerName, pattern) {
			return false
		}
	}

	return true
}

// isValidEnglishFontStyle checks if a font style is a valid English style name
func isValidEnglishFontStyle(style string) bool {
	if style == "" || len(style) > 30 {
		return false
	}

	// Check if it contains only ASCII characters (English)
	for _, r := range style {
		if r > 127 {
			return false
		}
	}

	styleLower := strings.ToLower(style)

	// Known valid English font style names
	validStyles := []string{
		"regular", "normal", "book", "roman",
		"bold", "heavy", "black", "extrabold", "ultrabold",
		"italic", "oblique", "slanted",
		"light", "thin", "ultralight", "extralight",
		"medium", "semibold", "demibold",
		"condensed", "narrow", "extended", "expanded",
		"bold italic", "light italic", "medium italic", "semibold italic",
		"extralight italic", "extrabold italic", "thin italic",
	}

	// Check if the style matches any valid patterns
	for _, validStyle := range validStyles {
		if styleLower == validStyle {
			return true
		}
	}

	return false
}

// GetCurrentFontFamily reads the current font family from alacritty.toml
func (m *Manager) GetCurrentFontFamily() string {
	content, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		return "monospace" // fallback
	}

	lines := strings.Split(string(content), "\n")
	inFontSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[font]" {
			inFontSection = true
			continue
		}

		if inFontSection && strings.HasPrefix(line, "[") && line != "[font]" {
			break // End of font section
		}

		if inFontSection && strings.Contains(line, "normal") && strings.Contains(line, "family") {
			// Parse: normal = { family = "Font Name", style = "Style" }
			// Handle UTF-8 characters and special characters in font names
			if parts := strings.Split(line, "family"); len(parts) > 1 {
				familyPart := parts[1]
				// Look for the pattern = "fontname"
				if equalIdx := strings.Index(familyPart, "="); equalIdx != -1 {
					familyPart = familyPart[equalIdx+1:]
					familyPart = strings.TrimSpace(familyPart)
					if start := strings.Index(familyPart, "\""); start != -1 {
						familyPart = familyPart[start+1:]
						if end := strings.Index(familyPart, "\""); end != -1 {
							fontName := familyPart[:end]
							// Validate and clean the font name
							if len(fontName) > 0 {
								return fontName
							}
						}
					}
				}
			}
		}
	}

	return "monospace" // fallback
}

// GetCurrentFontSize reads the current font size from alacritty.toml
func (m *Manager) GetCurrentFontSize() float64 {
	content, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		return 12.0 // fallback
	}

	lines := strings.Split(string(content), "\n")
	inFontSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[font]" {
			inFontSection = true
			continue
		}

		if inFontSection && strings.HasPrefix(line, "[") && line != "[font]" {
			break // End of font section
		}

		if inFontSection && strings.HasPrefix(line, "size") {
			// Parse: size = 14
			if parts := strings.Split(line, "="); len(parts) == 2 {
				sizeStr := strings.TrimSpace(parts[1])
				if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
					return size
				}
			}
		}
	}

	return 12.0 // fallback
}

// GetCurrentFontStyle reads the current font style from alacritty.toml
func (m *Manager) GetCurrentFontStyle() string {
	content, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		return "Regular" // fallback
	}

	lines := strings.Split(string(content), "\n")
	inFontSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[font]" {
			inFontSection = true
			continue
		}

		if inFontSection && strings.HasPrefix(line, "[") && line != "[font]" {
			break // End of font section
		}

		if inFontSection && strings.Contains(line, "normal") && strings.Contains(line, "style") {
			// Parse: normal = { family = "Font Name", style = "Style" }
			// Handle UTF-8 characters and special characters in style names
			if parts := strings.Split(line, "style"); len(parts) > 1 {
				stylePart := parts[1]
				// Look for the pattern = "stylename"
				if equalIdx := strings.Index(stylePart, "="); equalIdx != -1 {
					stylePart = stylePart[equalIdx+1:]
					stylePart = strings.TrimSpace(stylePart)
					if start := strings.Index(stylePart, "\""); start != -1 {
						stylePart = stylePart[start+1:]
						if end := strings.Index(stylePart, "\""); end != -1 {
							styleName := stylePart[:end]
							// Validate and clean the style name
							if len(styleName) > 0 {
								return styleName
							}
						}
					}
				}
			}
		}
	}

	return "Regular" // fallback
}

// UpdateFontFamily updates the font family in alacritty.toml
func (m *Manager) UpdateFontFamily(family string) error {
	return m.updateFontProperty("family", fmt.Sprintf("\"%s\"", family))
}

// UpdateFontSize updates the font size in alacritty.toml
func (m *Manager) UpdateFontSize(sizeStr string) error {
	// Validate the size is a number
	if _, err := strconv.ParseFloat(sizeStr, 64); err != nil {
		return fmt.Errorf("invalid font size: %s", sizeStr)
	}
	return m.updateFontProperty("size", sizeStr)
}

// UpdateFontStyle updates the font style in alacritty.toml
func (m *Manager) UpdateFontStyle(style string) error {
	return m.updateFontProperty("style", fmt.Sprintf("\"%s\"", style))
}

// updateFontProperty updates a specific font property in the config file.
//
// The whole read-modify-write cycle runs inside UpdateConfigFile, so it is
// serialised against every other writer and lands atomically. That matters more
// here than anywhere else in the program: the font panel can ask for an update
// on every keystroke, and this function rewrites a file the user owns.
func (m *Manager) updateFontProperty(property, value string) error {
	return config.UpdateConfigFile(m.config.ConfigFile, func(content string) (string, error) {
		return applyFontProperty(content, property, strings.Trim(value, "\""))
	})
}

// applyFontProperty returns content with one font property replaced. It is a
// pure function of the text it is handed — it never re-reads the file — so the
// family and style it preserves are guaranteed to be the ones it is editing,
// not whatever a concurrent write happened to leave on disk a moment later.
func applyFontProperty(content, property, value string) (string, error) {
	family, style := parseFontFace(content)
	switch property {
	case "family":
		family = value
	case "style":
		style = value
	}
	if family == "" {
		family = "monospace"
	}
	if style == "" {
		style = "Regular"
	}

	faceLine := fmt.Sprintf("normal = { family = %q, style = %q }", family, style)
	sizeLine := fmt.Sprintf("size = %s", value)
	if property != "size" {
		sizeLine = ""
	}

	lines := strings.Split(content, "\n")
	var result []string
	inFontSection, fontSectionFound, updated := false, false, false
	fontHeader := -1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "[font]" {
			inFontSection, fontSectionFound = true, true
			fontHeader = len(result)
			result = append(result, line)
			continue
		}
		// Any other table header ends the font table, including the [[array]]
		// form, whose entries would otherwise be read as font keys.
		if inFontSection && strings.HasPrefix(trimmed, "[") {
			inFontSection = false
		}

		if inFontSection && !updated {
			switch {
			case property == "size" && strings.HasPrefix(trimmed, "size"):
				result = append(result, sizeLine)
				updated = true
				continue
			case property != "size" && strings.HasPrefix(trimmed, "normal"):
				result = append(result, faceLine)
				updated = true
				continue
			}
		}

		result = append(result, line)
	}

	newLine := faceLine
	if property == "size" {
		newLine = sizeLine
	}

	switch {
	case updated:
		// Nothing more to do: the key was replaced in place.
	case fontSectionFound:
		// The table exists but lacks this key, so put it right under the header.
		result = append(result, "")
		copy(result[fontHeader+2:], result[fontHeader+1:])
		result[fontHeader+1] = newLine
	default:
		// No [font] table at all: append one rather than prepend, so existing
		// root-level keys are not re-parented into it.
		if len(result) > 0 && strings.TrimSpace(result[len(result)-1]) != "" {
			result = append(result, "")
		}
		result = append(result, "[font]", newLine)
	}

	updatedContent := strings.Join(result, "\n")

	// A font edit may only ever add or replace a line. If the result lost a
	// table that the input had, something is wrong with the parse above and the
	// safe move is to refuse rather than hand back a truncated config.
	if before, after := countTables(content), countTables(updatedContent); after < before {
		return content, fmt.Errorf("refusing to write a config that lost %d table(s)", before-after)
	}

	return updatedContent, nil
}

// countTables counts TOML table headers, the cheap proxy for "did this rewrite
// throw away part of the user's config".
func countTables(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "[") {
			count++
		}
	}
	return count
}

// parseFontFace pulls the family and style out of a [font] normal = { … } line.
func parseFontFace(content string) (family, style string) {
	inFontSection := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[font]" {
			inFontSection = true
			continue
		}
		if inFontSection && strings.HasPrefix(trimmed, "[") {
			break
		}
		if inFontSection && strings.HasPrefix(trimmed, "normal") {
			return extractQuoted(trimmed, "family"), extractQuoted(trimmed, "style")
		}
	}
	return "", ""
}

// extractQuoted reads key = "value" out of an inline table.
func extractQuoted(line, key string) string {
	idx := strings.Index(line, key)
	if idx < 0 {
		return ""
	}
	rest := line[idx+len(key):]
	open := strings.Index(rest, "\"")
	if open < 0 {
		return ""
	}
	rest = rest[open+1:]
	end := strings.Index(rest, "\"")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// GetFontFamilies returns a list of available font families
func (m *Manager) GetFontFamilies() []string {
	// Where the platform can be asked directly, ask it: these are the names
	// the terminal itself will resolve, so none of them can fail to load.
	if families := m.systemFonts(); families != nil {
		return sortedKeys(families)
	}

	var fonts []string

	// Use fc-list to get monospace fonts specifically with proper UTF-8 handling
	cmd := exec.Command("fc-list", ":family:spacing=mono", "family")
	cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
	output, err := cmd.Output()
	if err != nil {
		// If fc-list fails, return empty list - no fallbacks
		return []string{}
	}

	lines := strings.Split(string(output), "\n")
	fontSet := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// fc-list returns comma-separated families, take the first (primary) name
		families := strings.Split(line, ",")
		if len(families) > 0 {
			family := strings.TrimSpace(families[0])
			if family != "" && isValidTerminalFont(family) {
				// Clean up any extra whitespace
				family = strings.Join(strings.Fields(family), " ")
				fontSet[family] = true
			}
		}
	}

	// Only families fontconfig can actually resolve are offered.
	//
	// This list used to be padded with popular names — JetBrains Mono, Fira
	// Code and friends — whether or not they were installed. Picking one of
	// those handed Alacritty a family it could not resolve, which is exactly
	// the "unable to load font" it then reported. A font the user does not
	// have is not a suggestion, it is a broken option.
	for font := range fontSet {
		if isInstalled(font) {
			fonts = append(fonts, font)
		}
	}

	sort.Strings(fonts)
	return fonts
}

// isInstalled reports whether fontconfig resolves family to itself.
//
// fc-match always answers with something: ask it for a font you do not have and
// it hands back the default substitute. So the test is not "did it answer" but
// "did it answer with what we asked for" — request Fira Code without it
// installed and the reply is Verdana, which is how a missing font is detected.
func isInstalled(family string) bool {
	cmd := exec.Command("fc-match", "-f", "%{family}", family)
	cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// fc-match may answer with several aliases for the same face.
	for _, candidate := range strings.Split(string(output), ",") {
		if strings.EqualFold(strings.TrimSpace(candidate), strings.TrimSpace(family)) {
			return true
		}
	}
	return false
}

// GetStylesForFont returns available styles for a given font family
func (m *Manager) GetStylesForFont(family string) []string {
	// The style is half of the name Alacritty has to match, so it comes from
	// the same source as the family or the pair may not exist.
	if families := m.systemFonts(); families != nil {
		if styles, ok := families[family]; ok {
			return styles
		}
	}

	var styles []string

	// Use fc-list to get styles for the specific font family with proper UTF-8 handling
	cmd := exec.Command("fc-list", fmt.Sprintf(":family=%s", family), "style")
	cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to common styles
		return []string{
			"Regular",
			"Bold",
			"Italic",
			"Bold Italic",
			"Light",
			"Medium",
			"SemiBold",
			"ExtraLight",
			"Thin",
		}
	}

	lines := strings.Split(string(output), "\n")
	styleSet := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse fc-list output format like ":style=Leicht,Regular"
		if strings.HasPrefix(line, ":style=") {
			stylesPart := strings.TrimPrefix(line, ":style=")
			stylesList := strings.Split(stylesPart, ",")

			// If multiple styles are present (localized,English), prefer the English one
			var bestStyle string
			for _, style := range stylesList {
				style = strings.TrimSpace(style)
				if style != "" {
					if isValidEnglishFontStyle(style) {
						// Found a valid English style, use it
						bestStyle = style
						break
					} else if bestStyle == "" {
						// Keep the first style as fallback if no English style found
						bestStyle = style
					}
				}
			}

			if bestStyle != "" {
				styleSet[bestStyle] = true
			}
		}
	}

	// Convert map to slice
	for style := range styleSet {
		styles = append(styles, style)
	}

	// Only return actual detected styles, no fallbacks

	return styles
}

// GetAvailableMonoFonts returns a list of available monospace fonts
func (m *Manager) GetAvailableMonoFonts() []string {
	var fonts []string

	// Use fc-list to get available fonts (Linux/Unix) with proper UTF-8 handling
	cmd := exec.Command("fc-list", ":family:spacing=mono", "family")
	cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to common monospace fonts
		return []string{
			"JetBrains Mono",
			"JetBrainsMonoNL Nerd Font",
			"Fira Code",
			"Source Code Pro",
			"DejaVu Sans Mono",
			"Liberation Mono",
			"Consolas",
			"Monaco",
			"Menlo",
			"Inconsolata",
			"Ubuntu Mono",
			"Noto Sans Mono",
		}
	}

	lines := strings.Split(string(output), "\n")
	fontSet := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// fc-list may return comma-separated families
		families := strings.Split(line, ",")
		for _, family := range families {
			family = strings.TrimSpace(family)
			// Validate font name and check for monospace characteristics
			if family != "" && len(family) > 0 && len(family) < 100 {
				familyLower := strings.ToLower(family)
				if strings.Contains(familyLower, "mono") ||
					strings.Contains(familyLower, "code") ||
					strings.Contains(familyLower, "nerd") ||
					strings.Contains(familyLower, "console") ||
					strings.Contains(familyLower, "terminal") {
					// Clean up any extra whitespace and validate UTF-8
					family = strings.Join(strings.Fields(family), " ")
					// Only add fonts that are suitable for terminal use
					if isValidTerminalFont(family) {
						fontSet[family] = true
					}
				}
			}
		}
	}

	// Convert map to slice
	for font := range fontSet {
		fonts = append(fonts, font)
	}

	// Add some common ones if not found
	commonMonoFonts := []string{
		"JetBrains Mono",
		"JetBrainsMonoNL Nerd Font",
		"Fira Code",
		"Source Code Pro",
		"monospace",
	}

	for _, font := range commonMonoFonts {
		if !fontSet[font] {
			fonts = append(fonts, font)
		}
	}

	return fonts
}
