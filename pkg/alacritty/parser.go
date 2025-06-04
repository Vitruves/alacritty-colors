package alacritty

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	Colors   ColorScheme            `toml:"colors"`
	Font     FontConfig             `toml:"font"`
	Window   WindowConfig           `toml:"window"`
	Sections map[string]interface{} `toml:",omitempty"`
}

type ColorScheme struct {
	Primary   PrimaryColors     `toml:"primary"`
	Cursor    CursorColors      `toml:"cursor"`
	Selection SelectionColors   `toml:"selection"`
	Normal    map[string]string `toml:"normal"`
	Bright    map[string]string `toml:"bright"`
	Dim       map[string]string `toml:"dim,omitempty"`
	Indexed   map[string]string `toml:"indexed_colors,omitempty"`
}

type PrimaryColors struct {
	Background string `toml:"background"`
	Foreground string `toml:"foreground"`
}

type CursorColors struct {
	Text   string `toml:"text"`
	Cursor string `toml:"cursor"`
}

type SelectionColors struct {
	Text       string `toml:"text"`
	Background string `toml:"background"`
}

type FontConfig struct {
	Size   float64    `toml:"size,omitempty"`
	Normal FontFamily `toml:"normal,omitempty"`
	Bold   FontFamily `toml:"bold,omitempty"`
	Italic FontFamily `toml:"italic,omitempty"`
}

type FontFamily struct {
	Family string `toml:"family"`
	Style  string `toml:"style,omitempty"`
}

type WindowConfig struct {
	Padding WindowPadding `toml:"padding,omitempty"`
	Title   string        `toml:"title,omitempty"`
}

type WindowPadding struct {
	X int `toml:"x"`
	Y int `toml:"y"`
}

// Parser handles parsing Alacritty configuration files
type Parser struct {
	colorRegex   *regexp.Regexp
	sectionRegex *regexp.Regexp
}

func NewParser() *Parser {
	return &Parser{
		colorRegex:   regexp.MustCompile(`^(\w+)\s*=\s*["']?(#?[0-9a-fA-F]{6}|#?[0-9a-fA-F]{3})["']?`),
		sectionRegex: regexp.MustCompile(`^\[([^\]]+)\]`),
	}
}

func (p *Parser) ParseFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	config := &Config{
		Colors: ColorScheme{
			Normal: make(map[string]string),
			Bright: make(map[string]string),
			Dim:    make(map[string]string),
		},
		Sections: make(map[string]interface{}),
	}

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if matches := p.sectionRegex.FindStringSubmatch(line); matches != nil {
			currentSection = matches[1]
			continue
		}

		// Parse key-value pairs
		if err := p.parseKeyValue(config, currentSection, line); err != nil {
			// Log warning but continue parsing
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return config, nil
}

func (p *Parser) parseKeyValue(config *Config, section, line string) error {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid key-value pair: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

	switch section {
	case "colors.primary":
		p.setPrimaryColor(config, key, value)
	case "colors.cursor":
		p.setCursorColor(config, key, value)
	case "colors.selection":
		p.setSelectionColor(config, key, value)
	case "colors.normal":
		config.Colors.Normal[key] = value
	case "colors.bright":
		config.Colors.Bright[key] = value
	case "colors.dim":
		config.Colors.Dim[key] = value
	case "font":
		p.setFontConfig(config, key, value)
	case "window":
		p.setWindowConfig(config, key, value)
	default:
		// Store in generic sections map
		if config.Sections[section] == nil {
			config.Sections[section] = make(map[string]string)
		}
		if sectionMap, ok := config.Sections[section].(map[string]string); ok {
			sectionMap[key] = value
		}
	}

	return nil
}

func (p *Parser) setPrimaryColor(config *Config, key, value string) {
	switch key {
	case "background":
		config.Colors.Primary.Background = value
	case "foreground":
		config.Colors.Primary.Foreground = value
	}
}

func (p *Parser) setCursorColor(config *Config, key, value string) {
	switch key {
	case "text":
		config.Colors.Cursor.Text = value
	case "cursor":
		config.Colors.Cursor.Cursor = value
	}
}

func (p *Parser) setSelectionColor(config *Config, key, value string) {
	switch key {
	case "text":
		config.Colors.Selection.Text = value
	case "background":
		config.Colors.Selection.Background = value
	}
}

func (p *Parser) setFontConfig(config *Config, key, value string) {
	switch key {
	case "size":
		if size, err := parseFloat(value); err == nil {
			config.Font.Size = size
		}
	case "family":
		config.Font.Normal.Family = value
	}
}

func (p *Parser) setWindowConfig(config *Config, key, value string) {
	switch key {
	case "title":
		config.Window.Title = value
	}
}

func parseFloat(s string) (float64, error) {
	// Simple float parsing - could use strconv.ParseFloat for more robust parsing
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// ExtractColors extracts all color values from a configuration
func (p *Parser) ExtractColors(config *Config) map[string]string {
	colors := make(map[string]string)

	// Primary colors
	if config.Colors.Primary.Background != "" {
		colors["background"] = config.Colors.Primary.Background
	}
	if config.Colors.Primary.Foreground != "" {
		colors["foreground"] = config.Colors.Primary.Foreground
	}

	// Cursor colors
	if config.Colors.Cursor.Text != "" {
		colors["cursor_text"] = config.Colors.Cursor.Text
	}
	if config.Colors.Cursor.Cursor != "" {
		colors["cursor"] = config.Colors.Cursor.Cursor
	}

	// Selection colors
	if config.Colors.Selection.Text != "" {
		colors["selection_text"] = config.Colors.Selection.Text
	}
	if config.Colors.Selection.Background != "" {
		colors["selection_background"] = config.Colors.Selection.Background
	}

	// Normal colors
	for name, color := range config.Colors.Normal {
		colors["normal_"+name] = color
	}

	// Bright colors
	for name, color := range config.Colors.Bright {
		colors["bright_"+name] = color
	}

	// Dim colors
	for name, color := range config.Colors.Dim {
		colors["dim_"+name] = color
	}

	return colors
}

// ValidateColors checks if colors are valid hex values
func (p *Parser) ValidateColors(colors map[string]string) []string {
	var errors []string
	hexRegex := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

	for name, color := range colors {
		if !hexRegex.MatchString(color) {
			errors = append(errors, fmt.Sprintf("invalid color format for %s: %s", name, color))
		}
	}

	return errors
}

// NormalizeColor ensures color is in proper hex format
func (p *Parser) NormalizeColor(color string) string {
	// Remove quotes and whitespace
	color = strings.Trim(strings.TrimSpace(color), `"'`)

	// Add # if missing
	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}

	// Convert 3-digit hex to 6-digit
	if len(color) == 4 {
		r, g, b := color[1], color[2], color[3]
		color = fmt.Sprintf("#%c%c%c%c%c%c", r, r, g, g, b, b)
	}

	// Convert to lowercase
	return strings.ToLower(color)
}

// GenerateConfig creates a new configuration with given colors
func (p *Parser) GenerateConfig(colors map[string]string, template *Config) *Config {
	if template == nil {
		template = &Config{
			Colors: ColorScheme{
				Normal: make(map[string]string),
				Bright: make(map[string]string),
				Dim:    make(map[string]string),
			},
		}
	}

	// Copy template
	config := *template

	// Set colors from map
	for key, value := range colors {
		normalizedColor := p.NormalizeColor(value)

		switch key {
		case "background":
			config.Colors.Primary.Background = normalizedColor
		case "foreground":
			config.Colors.Primary.Foreground = normalizedColor
		case "cursor_text":
			config.Colors.Cursor.Text = normalizedColor
		case "cursor":
			config.Colors.Cursor.Cursor = normalizedColor
		case "selection_text":
			config.Colors.Selection.Text = normalizedColor
		case "selection_background":
			config.Colors.Selection.Background = normalizedColor
		default:
			if strings.HasPrefix(key, "normal_") {
				config.Colors.Normal[strings.TrimPrefix(key, "normal_")] = normalizedColor
			} else if strings.HasPrefix(key, "bright_") {
				config.Colors.Bright[strings.TrimPrefix(key, "bright_")] = normalizedColor
			} else if strings.HasPrefix(key, "dim_") {
				config.Colors.Dim[strings.TrimPrefix(key, "dim_")] = normalizedColor
			}
		}
	}

	return &config
}
