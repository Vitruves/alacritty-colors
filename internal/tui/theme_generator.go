package tui

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/vitruves/alacritty-colors/internal/theme"
)

// HarmonyType represents different color harmony strategies
type HarmonyType int

const (
	HarmonyComplementary     HarmonyType = iota // Opposite colors (180°)
	HarmonyAnalogous                            // Adjacent colors (30° apart)
	HarmonyTriadic                              // Three colors (120° apart)
	HarmonySplitComplementary                   // Base + two adjacent to complement
	HarmonyTetradic                             // Four colors (90° apart)
	HarmonyMonochromatic                        // Single hue, varying saturation/lightness
)

// ThemeStyle represents the overall theme style
type ThemeStyle int

const (
	ThemeStyleDark ThemeStyle = iota
	ThemeStyleLight
)

// GeneratedTheme holds the generated color values
type GeneratedTheme struct {
	Name       string
	Style      ThemeStyle
	Harmony    HarmonyType
	Background string
	Foreground string
	Cursor     string
	Selection  string
	Normal     map[string]string
	Bright     map[string]string
}

// HarmonyNames for display
var HarmonyNames = []string{
	"Complementary",
	"Analogous",
	"Triadic",
	"Split-Complementary",
	"Tetradic",
	"Monochromatic",
}

// WCAG contrast ratio constants
const (
	MinContrastRatio     = 4.5  // WCAG AA for normal text
	TargetContrastRatio  = 7.0  // WCAG AAA
	GoldenRatio          = 1.618
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomTheme creates a completely random theme
func GenerateRandomTheme(style ThemeStyle) *GeneratedTheme {
	theme := &GeneratedTheme{
		Name:   fmt.Sprintf("random_%s", time.Now().Format("150405")),
		Style:  style,
		Normal: make(map[string]string),
		Bright: make(map[string]string),
	}

	if style == ThemeStyleDark {
		// Dark background, light foreground
		bgL := 0.05 + rand.Float64()*0.15 // 5-20% lightness
		fgL := 0.75 + rand.Float64()*0.20 // 75-95% lightness

		bgHue := rand.Float64() * 360
		theme.Background = hslToHex(bgHue, rand.Float64()*0.3, bgL)
		theme.Foreground = hslToHex(bgHue, rand.Float64()*0.1, fgL)
	} else {
		// Light background, dark foreground
		bgL := 0.90 + rand.Float64()*0.08 // 90-98% lightness
		fgL := 0.10 + rand.Float64()*0.15 // 10-25% lightness

		bgHue := rand.Float64() * 360
		theme.Background = hslToHex(bgHue, rand.Float64()*0.2, bgL)
		theme.Foreground = hslToHex(bgHue, rand.Float64()*0.3, fgL)
	}

	// Generate random ANSI colors
	for _, colorName := range BaseColorNames {
		hue := rand.Float64() * 360
		sat := 0.5 + rand.Float64()*0.5  // 50-100% saturation

		var normalL, brightL float64
		if style == ThemeStyleDark {
			normalL = 0.45 + rand.Float64()*0.20 // 45-65%
			brightL = 0.60 + rand.Float64()*0.25 // 60-85%
		} else {
			normalL = 0.35 + rand.Float64()*0.20 // 35-55%
			brightL = 0.25 + rand.Float64()*0.20 // 25-45%
		}

		theme.Normal[colorName] = hslToHex(hue, sat, normalL)
		theme.Bright[colorName] = hslToHex(hue, sat*0.9, brightL)
	}

	// Cursor and selection
	theme.Cursor = theme.Foreground
	theme.Selection = hslToHex(rand.Float64()*360, 0.5, 0.4)

	return theme
}

// GenerateHarmoniousTheme creates a theme using color theory principles
func GenerateHarmoniousTheme(style ThemeStyle, harmony HarmonyType) *GeneratedTheme {
	theme := &GeneratedTheme{
		Name:    fmt.Sprintf("harmony_%s_%s", HarmonyNames[harmony], time.Now().Format("150405")),
		Style:   style,
		Harmony: harmony,
		Normal:  make(map[string]string),
		Bright:  make(map[string]string),
	}

	// Start with a random base hue
	baseHue := rand.Float64() * 360

	// Generate background/foreground with good contrast
	if style == ThemeStyleDark {
		theme.Background = generateDarkBackground(baseHue)
		theme.Foreground = generateContrastingForeground(theme.Background, true)
	} else {
		theme.Background = generateLightBackground(baseHue)
		theme.Foreground = generateContrastingForeground(theme.Background, false)
	}

	// Generate harmonious palette based on type
	palette := generateHarmoniousPalette(baseHue, harmony, 8)

	// Map palette to ANSI colors with semantic meaning
	theme.Normal = mapPaletteToANSI(palette, style, false)
	theme.Bright = mapPaletteToANSI(palette, style, true)

	// Cursor uses accent color
	accentHue := getAccentHue(baseHue, harmony)
	if style == ThemeStyleDark {
		theme.Cursor = hslToHex(accentHue, 0.8, 0.6)
		theme.Selection = hslToHex(accentHue, 0.4, 0.3)
	} else {
		theme.Cursor = hslToHex(accentHue, 0.8, 0.4)
		theme.Selection = hslToHex(accentHue, 0.3, 0.8)
	}

	return theme
}

// generateDarkBackground creates a dark background with subtle color
func generateDarkBackground(hue float64) string {
	// Very low saturation, very low lightness
	sat := 0.05 + rand.Float64()*0.15 // 5-20%
	light := 0.06 + rand.Float64()*0.08 // 6-14%
	return hslToHex(hue, sat, light)
}

// generateLightBackground creates a light background with subtle color
func generateLightBackground(hue float64) string {
	sat := 0.02 + rand.Float64()*0.08 // 2-10%
	light := 0.94 + rand.Float64()*0.05 // 94-99%
	return hslToHex(hue, sat, light)
}

// generateContrastingForeground creates a foreground with good contrast
func generateContrastingForeground(bgHex string, isDark bool) string {
	bgRGB, _ := theme.HexToRGB(bgHex)
	bgLum := relativeLuminance(bgRGB.R, bgRGB.G, bgRGB.B)

	// Calculate required foreground luminance for target contrast
	var targetLum float64
	if isDark {
		// For dark bg, we need light fg: (fgLum + 0.05) / (bgLum + 0.05) >= ratio
		targetLum = (TargetContrastRatio * (bgLum + 0.05)) - 0.05
		if targetLum > 1.0 {
			targetLum = 0.95
		}
	} else {
		// For light bg, we need dark fg: (bgLum + 0.05) / (fgLum + 0.05) >= ratio
		targetLum = ((bgLum + 0.05) / TargetContrastRatio) - 0.05
		if targetLum < 0.0 {
			targetLum = 0.05
		}
	}

	// Convert target luminance to lightness (approximate)
	targetL := math.Pow(targetLum, 1.0/2.2)

	// Very low saturation for foreground
	hsl := RGBToHSL(bgRGB.R, bgRGB.G, bgRGB.B)
	return hslToHex(hsl.H, 0.05+rand.Float64()*0.1, targetL)
}

// generateHarmoniousPalette creates a color palette based on harmony rules
func generateHarmoniousPalette(baseHue float64, harmony HarmonyType, count int) []float64 {
	hues := make([]float64, count)

	switch harmony {
	case HarmonyComplementary:
		// Two main hues, 180° apart
		for i := 0; i < count; i++ {
			if i%2 == 0 {
				hues[i] = normalizeHue(baseHue + float64(i)*10)
			} else {
				hues[i] = normalizeHue(baseHue + 180 + float64(i)*10)
			}
		}

	case HarmonyAnalogous:
		// Adjacent hues, 30° apart
		spread := 30.0
		for i := 0; i < count; i++ {
			offset := (float64(i) - float64(count)/2) * spread / float64(count)
			hues[i] = normalizeHue(baseHue + offset)
		}

	case HarmonyTriadic:
		// Three hues, 120° apart
		for i := 0; i < count; i++ {
			section := i % 3
			offset := float64(i/3) * 15
			hues[i] = normalizeHue(baseHue + float64(section)*120 + offset)
		}

	case HarmonySplitComplementary:
		// Base + two colors 150° and 210° away
		angles := []float64{0, 150, 210}
		for i := 0; i < count; i++ {
			angleIdx := i % 3
			offset := float64(i/3) * 20
			hues[i] = normalizeHue(baseHue + angles[angleIdx] + offset)
		}

	case HarmonyTetradic:
		// Four hues, 90° apart
		for i := 0; i < count; i++ {
			section := i % 4
			offset := float64(i/4) * 10
			hues[i] = normalizeHue(baseHue + float64(section)*90 + offset)
		}

	case HarmonyMonochromatic:
		// Single hue for all
		for i := 0; i < count; i++ {
			hues[i] = baseHue
		}
	}

	return hues
}

// mapPaletteToANSI maps generated hues to semantic ANSI colors
func mapPaletteToANSI(palette []float64, style ThemeStyle, bright bool) map[string]string {
	colors := make(map[string]string)

	// Semantic color mappings (hue targets)
	semanticHues := map[string]float64{
		"red":     0,    // 0° - Red
		"yellow":  60,   // 60° - Yellow
		"green":   120,  // 120° - Green
		"cyan":    180,  // 180° - Cyan
		"blue":    240,  // 240° - Blue
		"magenta": 300,  // 300° - Magenta
	}

	// Find closest palette color for each semantic color
	for name, targetHue := range semanticHues {
		closestHue := findClosestHue(palette, targetHue)

		var sat, light float64
		if style == ThemeStyleDark {
			sat = 0.65 + rand.Float64()*0.25 // 65-90%
			if bright {
				light = 0.65 + rand.Float64()*0.20 // 65-85%
			} else {
				light = 0.50 + rand.Float64()*0.15 // 50-65%
			}
		} else {
			sat = 0.70 + rand.Float64()*0.25 // 70-95%
			if bright {
				light = 0.35 + rand.Float64()*0.15 // 35-50%
			} else {
				light = 0.40 + rand.Float64()*0.15 // 40-55%
			}
		}

		colors[name] = hslToHex(closestHue, sat, light)
	}

	// Black and white are special
	if style == ThemeStyleDark {
		if bright {
			colors["black"] = hslToHex(0, 0, 0.35)
			colors["white"] = hslToHex(0, 0, 0.95)
		} else {
			colors["black"] = hslToHex(0, 0, 0.15)
			colors["white"] = hslToHex(0, 0, 0.80)
		}
	} else {
		if bright {
			colors["black"] = hslToHex(0, 0, 0.20)
			colors["white"] = hslToHex(0, 0, 0.70)
		} else {
			colors["black"] = hslToHex(0, 0, 0.30)
			colors["white"] = hslToHex(0, 0, 0.60)
		}
	}

	return colors
}

// findClosestHue finds the closest hue in palette to target
func findClosestHue(palette []float64, target float64) float64 {
	closest := palette[0]
	minDist := hueDistance(closest, target)

	for _, hue := range palette {
		dist := hueDistance(hue, target)
		if dist < minDist {
			minDist = dist
			closest = hue
		}
	}

	return closest
}

// hueDistance calculates the shortest distance between two hues
func hueDistance(h1, h2 float64) float64 {
	diff := math.Abs(h1 - h2)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

// getAccentHue returns an accent hue based on harmony type
func getAccentHue(baseHue float64, harmony HarmonyType) float64 {
	switch harmony {
	case HarmonyComplementary:
		return normalizeHue(baseHue + 180)
	case HarmonyAnalogous:
		return normalizeHue(baseHue + 30)
	case HarmonyTriadic:
		return normalizeHue(baseHue + 120)
	case HarmonySplitComplementary:
		return normalizeHue(baseHue + 150)
	case HarmonyTetradic:
		return normalizeHue(baseHue + 90)
	default:
		return baseHue
	}
}

// normalizeHue keeps hue in 0-360 range
func normalizeHue(hue float64) float64 {
	for hue < 0 {
		hue += 360
	}
	for hue >= 360 {
		hue -= 360
	}
	return hue
}

// relativeLuminance calculates WCAG relative luminance
func relativeLuminance(r, g, b int) float64 {
	rs := float64(r) / 255.0
	gs := float64(g) / 255.0
	bs := float64(b) / 255.0

	if rs <= 0.03928 {
		rs = rs / 12.92
	} else {
		rs = math.Pow((rs+0.055)/1.055, 2.4)
	}
	if gs <= 0.03928 {
		gs = gs / 12.92
	} else {
		gs = math.Pow((gs+0.055)/1.055, 2.4)
	}
	if bs <= 0.03928 {
		bs = bs / 12.92
	} else {
		bs = math.Pow((bs+0.055)/1.055, 2.4)
	}

	return 0.2126*rs + 0.7152*gs + 0.0722*bs
}

// hslToHex converts HSL to hex string
func hslToHex(h, s, l float64) string {
	r, g, b := HSLToRGB(HSL{H: h, S: s, L: l})
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// ApplyGeneratedTheme applies a generated theme to the editor
func (ce *ColorEditor) ApplyGeneratedTheme(gen *GeneratedTheme) {
	ce.colorValues["primary.background"] = gen.Background
	ce.colorValues["primary.foreground"] = gen.Foreground
	ce.colorValues["cursor.cursor"] = gen.Cursor
	ce.colorValues["cursor.text"] = gen.Background
	ce.colorValues["selection.background"] = gen.Selection
	ce.colorValues["selection.text"] = gen.Foreground

	for name, color := range gen.Normal {
		ce.colorValues["normal."+name] = color
	}
	for name, color := range gen.Bright {
		ce.colorValues["bright."+name] = color
	}

	ce.themeName = gen.Name
	ce.isDirty = true
	ce.colorPanel.SetTitle(fmt.Sprintf(" Color Palette - %s (Generated) ", gen.Name))

	// Update internal theme config
	if ce.currentTheme != nil {
		ce.updateThemeConfig()
	}

	ce.buildColorPanel()
	ce.updatePreview()
}
