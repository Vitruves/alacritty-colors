package font

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
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

// updateFontProperty updates a specific font property in the config file
func (m *Manager) updateFontProperty(property, value string) error {
	content, err := os.ReadFile(m.config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var result []string
	inFontSection := false
	fontSectionFound := false
	updated := false

	for i, line := range lines {
		_ = i // Used later for insertion logic
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "[font]" {
			inFontSection = true
			fontSectionFound = true
			result = append(result, line)
			continue
		}

		if inFontSection && strings.HasPrefix(trimmedLine, "[") && trimmedLine != "[font]" {
			inFontSection = false
		}

		if inFontSection {
			if property == "size" && strings.HasPrefix(trimmedLine, "size") {
				result = append(result, fmt.Sprintf("size = %s", value))
				updated = true
				continue
			} else if (property == "family" || property == "style") && strings.Contains(trimmedLine, "normal") {
				// Update the normal line with new family/style
				currentFamily := m.GetCurrentFontFamily()
				currentStyle := m.GetCurrentFontStyle()

				switch property {
				case "family":
					currentFamily = strings.Trim(value, "\"")
				case "style":
					currentStyle = strings.Trim(value, "\"")
				}

				result = append(result, fmt.Sprintf("normal = { family = \"%s\", style = \"%s\" }", currentFamily, currentStyle))
				updated = true
				continue
			}
		}

		result = append(result, line)
	}

	// If font section doesn't exist, create it
	if !fontSectionFound {
		result = append(result, "")
		result = append(result, "[font]")
		switch property {
		case "family", "style":
			family := "monospace"
			style := "Regular"
			switch property {
			case "family":
				family = strings.Trim(value, "\"")
			case "style":
				style = strings.Trim(value, "\"")
			}
			result = append(result, fmt.Sprintf("normal = { family = \"%s\", style = \"%s\" }", family, style))
		case "size":
			result = append(result, fmt.Sprintf("size = %s", value))
		}
		updated = true
	} else if !updated {
		// Property doesn't exist in font section, add it
		for i := len(result) - 1; i >= 0; i-- {
			if strings.TrimSpace(result[i]) == "[font]" {
				// Insert after [font] line
				newLine := ""
				switch property {
				case "family", "style":
					currentFamily := m.GetCurrentFontFamily()
					currentStyle := m.GetCurrentFontStyle()

					switch property {
					case "family":
						currentFamily = strings.Trim(value, "\"")
					case "style":
						currentStyle = strings.Trim(value, "\"")
					}

					newLine = fmt.Sprintf("normal = { family = \"%s\", style = \"%s\" }", currentFamily, currentStyle)
				case "size":
					newLine = fmt.Sprintf("size = %s", value)
				}

				result = append(result[:i+1], append([]string{newLine}, result[i+1:]...)...)
				break
			}
		}
	}

	// Write back to file
	newContent := strings.Join(result, "\n")
	return os.WriteFile(m.config.ConfigFile, []byte(newContent), 0644)
}

// GetFontFamilies returns a list of available font families
func (m *Manager) GetFontFamilies() []string {
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

	// Convert map to slice
	for font := range fontSet {
		fonts = append(fonts, font)
	}

	// Add some common ones if not found
	commonFonts := []string{
		"JetBrains Mono",
		"JetBrainsMonoNL Nerd Font",
		"Fira Code",
		"Source Code Pro",
		"Test Söhne Mono",
		"monospace",
	}

	for _, font := range commonFonts {
		if !fontSet[font] {
			fonts = append(fonts, font)
		}
	}

	return fonts
}

// GetStylesForFont returns available styles for a given font family
func (m *Manager) GetStylesForFont(family string) []string {
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
