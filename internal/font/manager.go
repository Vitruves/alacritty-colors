package font

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
			if parts := strings.Split(line, "family"); len(parts) > 1 {
				familyPart := parts[1]
				if start := strings.Index(familyPart, "\""); start != -1 {
					familyPart = familyPart[start+1:]
					if end := strings.Index(familyPart, "\""); end != -1 {
						return familyPart[:end]
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
			if parts := strings.Split(line, "style"); len(parts) > 1 {
				stylePart := parts[1]
				if start := strings.Index(stylePart, "\""); start != -1 {
					stylePart = stylePart[start+1:]
					if end := strings.Index(stylePart, "\""); end != -1 {
						return stylePart[:end]
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

	// Use fc-list to get all available fonts
	cmd := exec.Command("fc-list", ":family", "family")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to common fonts
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
			"Test Söhne Mono",
			"monospace",
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
			if family != "" {
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

	// Use fc-list to get styles for the specific font family
	cmd := exec.Command("fc-list", fmt.Sprintf(":family=%s", family), "style")
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
			for _, style := range stylesList {
				style = strings.TrimSpace(style)
				if style != "" {
					styleSet[style] = true
				}
			}
		} else {
			// Handle direct style names
			stylesList := strings.Split(line, ",")
			for _, style := range stylesList {
				style = strings.TrimSpace(style)
				if style != "" {
					styleSet[style] = true
				}
			}
		}
	}

	// Convert map to slice
	for style := range styleSet {
		styles = append(styles, style)
	}

	// If no styles found, add common ones
	if len(styles) == 0 {
		styles = []string{
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

	return styles
}

// GetAvailableMonoFonts returns a list of available monospace fonts
func (m *Manager) GetAvailableMonoFonts() []string {
	var fonts []string

	// Use fc-list to get available fonts (Linux/Unix)
	cmd := exec.Command("fc-list", ":family:spacing=mono", "family")
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
			if family != "" && (strings.Contains(strings.ToLower(family), "mono") ||
				strings.Contains(strings.ToLower(family), "code") ||
				strings.Contains(strings.ToLower(family), "nerd")) {
				fontSet[family] = true
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
