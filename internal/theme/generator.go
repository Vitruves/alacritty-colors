package theme

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/vitruves/alacritty-colors/internal/ui"
)

// Random word lists for generating theme names
var (
	adjectives = []string{
		"crimson", "azure", "emerald", "golden", "violet", "scarlet", "amber", "indigo",
		"silver", "copper", "jade", "ruby", "sapphire", "pearl", "coral", "ivory",
		"obsidian", "marble", "crystal", "diamond", "onyx", "garnet", "topaz", "opal",
		"mystic", "cosmic", "ethereal", "stellar", "lunar", "solar", "nova", "nebula",
		"electric", "neon", "plasma", "matrix", "cyber", "digital", "quantum", "atomic",
		"velvet", "silk", "satin", "linen", "cotton", "cashmere", "wool", "mohair",
		"frost", "shadow", "ember", "flame", "spark", "glow", "shimmer", "glitter",
		"deep", "bright", "dark", "light", "soft", "bold", "vivid", "muted",
	}

	nouns = []string{
		"tiger", "wolf", "eagle", "dragon", "phoenix", "raven", "hawk", "falcon",
		"mountain", "ocean", "forest", "desert", "valley", "river", "lake", "canyon",
		"storm", "thunder", "lightning", "tempest", "hurricane", "tornado", "blizzard", "rain",
		"sunset", "sunrise", "twilight", "dawn", "dusk", "midnight", "noon", "morning",
		"galaxy", "comet", "meteor", "planet", "star", "moon", "sun", "cosmos",
		"crystal", "diamond", "emerald", "sapphire", "ruby", "pearl", "opal", "jade",
		"warrior", "knight", "guardian", "sentinel", "defender", "champion", "hero", "legend",
		"whisper", "echo", "shadow", "dream", "vision", "phantom", "spirit", "ghost",
		"blade", "arrow", "shield", "crown", "throne", "tower", "castle", "fortress",
	}
)

// generateRandomName creates a random theme name using word combinations
func generateRandomName(scheme string) string {
	rand.Seed(time.Now().UnixNano())

	adjective := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]

	return fmt.Sprintf("%s_%s_%s", scheme, adjective, noun)
}

func (m *Manager) GenerateTheme(scheme, name string, save bool) error {
	ui.PrintInfo("Generating %s theme", scheme)

	colors, err := m.generateColorScheme(scheme)
	if err != nil {
		return fmt.Errorf("failed to generate colors: %w", err)
	}

	if name == "" {
		name = generateRandomName(scheme)
	}

	themeContent := m.createThemeContent(colors, scheme, name)

	// Always save generated themes
	themeFile := filepath.Join(m.config.ThemesDir, name+".toml")
	if err := os.WriteFile(themeFile, []byte(themeContent), 0644); err != nil {
		return fmt.Errorf("failed to save theme: %w", err)
	}

	ui.PrintSuccess("Generated theme saved: %s", name)

	// Apply the theme immediately
	if err := m.ApplyTheme(name); err != nil {
		return fmt.Errorf("failed to apply generated theme: %w", err)
	}

	return nil
}

func (m *Manager) generateColorScheme(scheme string) (map[string]string, error) {
	switch scheme {
	case "random":
		return m.generateRandomColors(), nil
	case "pastel":
		return m.generatePastelColors(), nil
	case "neon":
		return m.generateNeonColors(), nil
	case "mono", "monochrome":
		return m.generateMonochromeColors(), nil
	case "warm":
		return m.generateWarmColors(), nil
	case "cool":
		return m.generateCoolColors(), nil
	case "nature":
		return m.generateNatureColors(), nil
	case "cyberpunk":
		return m.generateCyberpunkColors(), nil
	case "dracula":
		return m.generateDraculaColors(), nil
	case "nord":
		return m.generateNordColors(), nil
	case "solarized":
		return m.generateSolarizedColors(), nil
	case "gruvbox":
		return m.generateGruvboxColors(), nil
	default:
		return nil, fmt.Errorf("unknown color scheme: %s", scheme)
	}
}

func (m *Manager) createThemeContent(colors map[string]string, scheme, name string) string {
	content := fmt.Sprintf(`# %s
# Generated theme: %s
# Scheme: %s
# Generated at: %s

[colors.primary]
background = "%s"
foreground = "%s"

[colors.cursor]
text = "%s"
cursor = "%s"

[colors.selection]
text = "%s"
background = "%s"

[colors.normal]
black = "%s"
red = "%s"
green = "%s"
yellow = "%s"
blue = "%s"
magenta = "%s"
cyan = "%s"
white = "%s"

[colors.bright]
black = "%s"
red = "%s"
green = "%s"
yellow = "%s"
blue = "%s"
magenta = "%s"
cyan = "%s"
white = "%s"
`,
		name,
		name,
		scheme,
		time.Now().Format("2006-01-02 15:04:05"),
		colors["background"],
		colors["foreground"],
		colors["background"],
		colors["foreground"],
		colors["foreground"],
		colors["selection_background"],
		colors["black"],
		colors["red"],
		colors["green"],
		colors["yellow"],
		colors["blue"],
		colors["magenta"],
		colors["cyan"],
		colors["white"],
		colors["bright_black"],
		colors["bright_red"],
		colors["bright_green"],
		colors["bright_yellow"],
		colors["bright_blue"],
		colors["bright_magenta"],
		colors["bright_cyan"],
		colors["bright_white"],
	)

	return content
}

func (m *Manager) generateColorSchemeWithVariant(scheme string, darkTheme, lightTheme bool) (map[string]string, error) {
	colors, err := m.generateColorScheme(scheme)
	if err != nil {
		return nil, err
	}

	// Apply light/dark variant adjustments
	if darkTheme {
		return m.convertToDarkVariant(colors), nil
	} else if lightTheme {
		return m.convertToLightVariant(colors), nil
	}

	return colors, nil
}

// Enhanced random colors with better contrast and harmony
func (m *Manager) generateRandomColors() map[string]string {
	colors := make(map[string]string)
	baseHue := randomFloat()

	// Background and foreground with excellent contrast
	bgLightness := randomFloat() * 0.2     // Darker backgrounds
	fgLightness := 0.8 + randomFloat()*0.2 // Brighter foregrounds

	colors["background"] = HSL{H: baseHue, S: 0.15, L: bgLightness}.ToRGB().ToHex()
	colors["foreground"] = HSL{H: baseHue, S: 0.1, L: fgLightness}.ToRGB().ToHex()
	colors["selection_background"] = HSL{H: baseHue, S: 0.4, L: 0.25}.ToRGB().ToHex()

	// Generate harmonious color palette with better distribution
	colorHues := []float64{0, 0.0, 0.33, 0.16, 0.66, 0.83, 0.5, 0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		var hue, sat, light float64

		if name == "black" {
			light = randomFloat() * 0.15
			sat = 0.1
			hue = baseHue
		} else if name == "white" {
			light = 0.85 + randomFloat()*0.15
			sat = 0.1
			hue = baseHue
		} else {
			// Use golden ratio for better color harmony
			hue = math.Mod(baseHue+colorHues[i]*0.618+randomFloat()*0.05, 1.0)
			sat = 0.7 + randomFloat()*0.3     // Higher saturation for vibrant colors
			light = 0.45 + randomFloat()*0.25 // Better contrast range
		}

		colors[name] = HSL{H: hue, S: sat, L: light}.ToRGB().ToHex()
		// Bright variants are lighter and more saturated
		brightSat := math.Min(1.0, sat+0.1)
		brightLight := math.Min(0.9, light+0.25)
		colors["bright_"+name] = HSL{H: hue, S: brightSat, L: brightLight}.ToRGB().ToHex()
	}

	return colors
}

// Enhanced pastel colors with better light/dark variants
func (m *Manager) generatePastelColors() map[string]string {
	colors := make(map[string]string)
	baseHue := randomFloat()

	// Light pastel background
	colors["background"] = "#faf7f4"
	colors["foreground"] = "#5c5c5c"
	colors["selection_background"] = "#e8e0db"

	pastelHues := []float64{0, 0.0, 0.25, 0.15, 0.6, 0.8, 0.5, 0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		var hue, sat, light float64

		if name == "black" {
			colors[name] = "#f0ede8"
			colors["bright_"+name] = "#d4ccc2"
		} else if name == "white" {
			colors[name] = "#928374"
			colors["bright_"+name] = "#7c6f64"
		} else {
			hue = math.Mod(baseHue+pastelHues[i]+randomFloat()*0.1-0.05, 1.0)
			sat = 0.3 + randomFloat()*0.2    // Muted saturation for pastels
			light = 0.6 + randomFloat()*0.15 // Light tones

			colors[name] = HSL{H: hue, S: sat, L: light}.ToRGB().ToHex()
			colors["bright_"+name] = HSL{H: hue, S: sat + 0.1, L: math.Min(0.85, light+0.15)}.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateNeonColors() map[string]string {
	colors := make(map[string]string)

	colors["background"] = "#0a0a0a"
	colors["foreground"] = "#00ff00"
	colors["selection_background"] = "#333333"

	neonHues := []float64{0, 0.0, 0.33, 0.16, 0.66, 0.83, 0.5, 0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		var hue, sat, light float64

		if name == "black" {
			colors[name] = "#1a1a1a"
			colors["bright_"+name] = "#333333"
		} else if name == "white" {
			colors[name] = "#ffffff"
			colors["bright_"+name] = "#ffffff"
		} else {
			hue = neonHues[i]
			sat = 1.0
			light = 0.5 + randomFloat()*0.3

			colors[name] = HSL{H: hue, S: sat, L: light}.ToRGB().ToHex()
			colors["bright_"+name] = HSL{H: hue, S: sat, L: math.Min(1.0, light+0.2)}.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateMonochromeColors() map[string]string {
	colors := make(map[string]string)
	baseHue := randomFloat()

	colors["background"] = HSL{H: baseHue, S: 0.05, L: 0.08}.ToRGB().ToHex()
	colors["foreground"] = HSL{H: baseHue, S: 0.05, L: 0.85}.ToRGB().ToHex()
	colors["selection_background"] = HSL{H: baseHue, S: 0.1, L: 0.2}.ToRGB().ToHex()

	lightnesses := []float64{0.1, 0.2, 0.35, 0.45, 0.55, 0.65, 0.75, 0.9}
	brightLightnesses := []float64{0.2, 0.3, 0.45, 0.55, 0.65, 0.75, 0.85, 1.0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		colors[name] = HSL{H: baseHue, S: 0.1, L: lightnesses[i]}.ToRGB().ToHex()
		colors["bright_"+name] = HSL{H: baseHue, S: 0.1, L: brightLightnesses[i]}.ToRGB().ToHex()
	}

	return colors
}

func (m *Manager) generateWarmColors() map[string]string {
	colors := make(map[string]string)

	colors["background"] = "#2d1b12"
	colors["foreground"] = "#f4e8d0"
	colors["selection_background"] = "#4a3426"

	warmHues := []float64{0, 0.0, 0.08, 0.15, 0.05, 0.02, 0.12, 0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		var hue, sat, light float64

		if name == "black" {
			colors[name] = "#1a0f08"
			colors["bright_"+name] = "#3d2317"
		} else if name == "white" {
			colors[name] = "#f4e8d0"
			colors["bright_"+name] = "#fff8e7"
		} else {
			hue = warmHues[i]
			sat = 0.6 + randomFloat()*0.3
			light = 0.4 + randomFloat()*0.3

			colors[name] = HSL{H: hue, S: sat, L: light}.ToRGB().ToHex()
			colors["bright_"+name] = HSL{H: hue, S: sat, L: math.Min(1.0, light+0.2)}.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateCoolColors() map[string]string {
	colors := make(map[string]string)

	colors["background"] = "#0f1419"
	colors["foreground"] = "#e6f1ff"
	colors["selection_background"] = "#1f2937"

	coolHues := []float64{0, 0.95, 0.4, 0.45, 0.6, 0.75, 0.5, 0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		var hue, sat, light float64

		if name == "black" {
			colors[name] = "#0b0e14"
			colors["bright_"+name] = "#1f2328"
		} else if name == "white" {
			colors[name] = "#e6f1ff"
			colors["bright_"+name] = "#ffffff"
		} else {
			hue = coolHues[i]
			sat = 0.6 + randomFloat()*0.3
			light = 0.4 + randomFloat()*0.3

			colors[name] = HSL{H: hue, S: sat, L: light}.ToRGB().ToHex()
			colors["bright_"+name] = HSL{H: hue, S: sat, L: math.Min(1.0, light+0.2)}.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateNatureColors() map[string]string {
	colors := make(map[string]string)

	colors["background"] = "#1a2318"
	colors["foreground"] = "#e8f5e8"
	colors["selection_background"] = "#2d3a2b"

	natureHues := []float64{0, 0.02, 0.25, 0.12, 0.55, 0.8, 0.45, 0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		var hue, sat, light float64

		if name == "black" {
			colors[name] = "#0f1a0e"
			colors["bright_"+name] = "#2d3a2b"
		} else if name == "white" {
			colors[name] = "#e8f5e8"
			colors["bright_"+name] = "#f0fff0"
		} else {
			hue = natureHues[i]
			sat = 0.5 + randomFloat()*0.3
			light = 0.4 + randomFloat()*0.2

			colors[name] = HSL{H: hue, S: sat, L: light}.ToRGB().ToHex()
			colors["bright_"+name] = HSL{H: hue, S: sat, L: math.Min(1.0, light+0.15)}.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateCyberpunkColors() map[string]string {
	colors := make(map[string]string)

	colors["background"] = "#0d001a"
	colors["foreground"] = "#00ff41"
	colors["selection_background"] = "#330066"

	cyberpunkHues := []float64{0, 0.95, 0.33, 0.16, 0.66, 0.83, 0.5, 0}
	colorNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	for i, name := range colorNames {
		var hue, sat, light float64

		if name == "black" {
			colors[name] = "#1a0033"
			colors["bright_"+name] = "#330066"
		} else if name == "white" {
			colors[name] = "#00ff41"
			colors["bright_"+name] = "#66ff99"
		} else {
			hue = cyberpunkHues[i]
			sat = 0.9 + randomFloat()*0.1
			light = 0.5 + randomFloat()*0.2

			colors[name] = HSL{H: hue, S: sat, L: light}.ToRGB().ToHex()
			colors["bright_"+name] = HSL{H: hue, S: sat, L: math.Min(1.0, light+0.2)}.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateDraculaColors() map[string]string {
	// Generate Dracula-inspired theme with variations
	baseColors := map[string]string{
		"background":           "#282a36",
		"foreground":           "#f8f8f2",
		"selection_background": "#44475a",
		"black":                "#21222c",
		"red":                  "#ff5555",
		"green":                "#50fa7b",
		"yellow":               "#f1fa8c",
		"blue":                 "#bd93f9",
		"magenta":              "#ff79c6",
		"cyan":                 "#8be9fd",
		"white":                "#f8f8f2",
	}

	colors := make(map[string]string)

	// Add some variation to the base Dracula colors
	for name, hex := range baseColors {
		rgb, _ := HexToRGB(hex)
		hsl := rgb.ToHSL()

		// Add slight variations
		hsl.H = math.Mod(hsl.H+(randomFloat()-0.5)*0.05, 1.0)
		hsl.S = math.Max(0, math.Min(1, hsl.S+(randomFloat()-0.5)*0.1))
		hsl.L = math.Max(0, math.Min(1, hsl.L+(randomFloat()-0.5)*0.05))

		colors[name] = hsl.ToRGB().ToHex()

		// Generate bright versions
		if name != "background" && name != "foreground" && name != "selection_background" {
			brightHsl := hsl
			brightHsl.L = math.Min(1.0, brightHsl.L+0.15)
			colors["bright_"+name] = brightHsl.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateNordColors() map[string]string {
	// Generate Nord-inspired theme with variations
	baseColors := map[string]string{
		"background":           "#2e3440",
		"foreground":           "#d8dee9",
		"selection_background": "#434c5e",
		"black":                "#3b4252",
		"red":                  "#bf616a",
		"green":                "#a3be8c",
		"yellow":               "#ebcb8b",
		"blue":                 "#81a1c1",
		"magenta":              "#b48ead",
		"cyan":                 "#88c0d0",
		"white":                "#e5e9f0",
	}

	colors := make(map[string]string)

	// Add some variation to the base Nord colors
	for name, hex := range baseColors {
		rgb, _ := HexToRGB(hex)
		hsl := rgb.ToHSL()

		// Add slight variations while maintaining the Nord aesthetic
		hsl.H = math.Mod(hsl.H+(randomFloat()-0.5)*0.03, 1.0)
		hsl.S = math.Max(0, math.Min(1, hsl.S+(randomFloat()-0.5)*0.05))
		hsl.L = math.Max(0, math.Min(1, hsl.L+(randomFloat()-0.5)*0.03))

		colors[name] = hsl.ToRGB().ToHex()

		// Generate bright versions
		if name != "background" && name != "foreground" && name != "selection_background" {
			brightHsl := hsl
			brightHsl.L = math.Min(1.0, brightHsl.L+0.1)
			colors["bright_"+name] = brightHsl.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateSolarizedColors() map[string]string {
	// Generate Solarized-inspired theme (dark variant with variations)
	baseColors := map[string]string{
		"background":           "#002b36",
		"foreground":           "#839496",
		"selection_background": "#073642",
		"black":                "#073642",
		"red":                  "#dc322f",
		"green":                "#859900",
		"yellow":               "#b58900",
		"blue":                 "#268bd2",
		"magenta":              "#d33682",
		"cyan":                 "#2aa198",
		"white":                "#eee8d5",
	}

	colors := make(map[string]string)

	// Add some variation while maintaining Solarized's precise color relationships
	for name, hex := range baseColors {
		rgb, _ := HexToRGB(hex)
		hsl := rgb.ToHSL()

		// Very subtle variations to maintain Solarized's carefully crafted palette
		hsl.H = math.Mod(hsl.H+(randomFloat()-0.5)*0.02, 1.0)
		hsl.S = math.Max(0, math.Min(1, hsl.S+(randomFloat()-0.5)*0.03))
		hsl.L = math.Max(0, math.Min(1, hsl.L+(randomFloat()-0.5)*0.02))

		colors[name] = hsl.ToRGB().ToHex()

		// Generate bright versions
		if name != "background" && name != "foreground" && name != "selection_background" {
			brightHsl := hsl
			brightHsl.L = math.Min(1.0, brightHsl.L+0.12)
			colors["bright_"+name] = brightHsl.ToRGB().ToHex()
		}
	}

	return colors
}

func (m *Manager) generateGruvboxColors() map[string]string {
	// Generate Gruvbox-inspired theme with variations
	baseColors := map[string]string{
		"background":           "#282828",
		"foreground":           "#ebdbb2",
		"selection_background": "#3c3836",
		"black":                "#282828",
		"red":                  "#cc241d",
		"green":                "#98971a",
		"yellow":               "#d79921",
		"blue":                 "#458588",
		"magenta":              "#b16286",
		"cyan":                 "#689d6a",
		"white":                "#a89984",
	}

	colors := make(map[string]string)

	// Add variations while maintaining Gruvbox's warm, retro aesthetic
	for name, hex := range baseColors {
		rgb, _ := HexToRGB(hex)
		hsl := rgb.ToHSL()

		// Add slight variations
		hsl.H = math.Mod(hsl.H+(randomFloat()-0.5)*0.04, 1.0)
		hsl.S = math.Max(0, math.Min(1, hsl.S+(randomFloat()-0.5)*0.08))
		hsl.L = math.Max(0, math.Min(1, hsl.L+(randomFloat()-0.5)*0.04))

		colors[name] = hsl.ToRGB().ToHex()

		// Generate bright versions
		if name != "background" && name != "foreground" && name != "selection_background" {
			brightHsl := hsl
			brightHsl.L = math.Min(1.0, brightHsl.L+0.15)
			brightHsl.S = math.Min(1.0, brightHsl.S+0.05)
			colors["bright_"+name] = brightHsl.ToRGB().ToHex()
		}
	}

	return colors
}
