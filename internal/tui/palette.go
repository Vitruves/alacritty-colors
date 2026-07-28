package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/internal/theme"
	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

// uiPalette holds the colours the TUI chrome is drawn with. Every value is
// derived from the theme the user currently has applied, so the editor always
// looks like the thing it is editing.
type uiPalette struct {
	bg          tcell.Color
	fg          tcell.Color
	muted       tcell.Color // fg blended towards bg: secondary text
	accent      tcell.Color // focused borders, titles
	accent2     tcell.Color
	warn        tcell.Color
	danger      tcell.Color
	ok          tcell.Color
	selBg       tcell.Color
	selFg       tcell.Color
	border      tcell.Color
	bgHex       string
	fgHex       string
	mutedHex    string
	accentHex   string
	accent2Hex  string
	warnHex     string
	dangerHex   string
	okHex       string
	initialized bool
}

// defaultPalette is used before a theme has been read, and as a fallback when a
// theme file omits colours.
func defaultPalette() uiPalette {
	return buildPalette(map[string]string{
		"background": "#101014",
		"foreground": "#d8d8e0",
		"blue":       "#7aa2f7",
		"cyan":       "#7dcfff",
		"yellow":     "#e0af68",
		"red":        "#f7768e",
		"green":      "#9ece6a",
	})
}

// paletteFromConfig derives the chrome palette from a parsed Alacritty config.
func paletteFromConfig(cfg *alacritty.Config) uiPalette {
	pick := func(m map[string]string, key, fallback string) string {
		if m != nil {
			if v, ok := m[key]; ok && v != "" {
				return v
			}
		}
		return fallback
	}

	// Bright variants read better as accents on dark backgrounds; fall back to
	// the normal set when a theme does not define them.
	accentSrc := cfg.Colors.Bright
	if accentSrc == nil || len(accentSrc) == 0 {
		accentSrc = cfg.Colors.Normal
	}

	return buildPalette(map[string]string{
		"background": cfg.Colors.Primary.Background,
		"foreground": cfg.Colors.Primary.Foreground,
		"blue":       pick(accentSrc, "blue", pick(cfg.Colors.Normal, "blue", "#7aa2f7")),
		"cyan":       pick(accentSrc, "cyan", pick(cfg.Colors.Normal, "cyan", "#7dcfff")),
		"yellow":     pick(accentSrc, "yellow", pick(cfg.Colors.Normal, "yellow", "#e0af68")),
		"red":        pick(accentSrc, "red", pick(cfg.Colors.Normal, "red", "#f7768e")),
		"green":      pick(accentSrc, "green", pick(cfg.Colors.Normal, "green", "#9ece6a")),
	})
}

func buildPalette(src map[string]string) uiPalette {
	bgHex := normalizeHex(src["background"], "#101014")
	fgHex := normalizeHex(src["foreground"], "#d8d8e0")

	p := uiPalette{
		bgHex:       bgHex,
		fgHex:       fgHex,
		mutedHex:    blendHex(fgHex, bgHex, 0.45),
		accentHex:   normalizeHex(src["blue"], "#7aa2f7"),
		accent2Hex:  normalizeHex(src["cyan"], "#7dcfff"),
		warnHex:     normalizeHex(src["yellow"], "#e0af68"),
		dangerHex:   normalizeHex(src["red"], "#f7768e"),
		okHex:       normalizeHex(src["green"], "#9ece6a"),
		initialized: true,
	}

	// Accents must stay legible against the background; if a theme's blue is
	// too close to its own background, lift it until it clears WCAG AA.
	p.accentHex = ensureReadable(p.accentHex, bgHex)
	p.accent2Hex = ensureReadable(p.accent2Hex, bgHex)
	p.warnHex = ensureReadable(p.warnHex, bgHex)
	p.dangerHex = ensureReadable(p.dangerHex, bgHex)
	p.okHex = ensureReadable(p.okHex, bgHex)

	p.bg = hexColor(p.bgHex)
	p.fg = hexColor(p.fgHex)
	p.muted = hexColor(p.mutedHex)
	p.accent = hexColor(p.accentHex)
	p.accent2 = hexColor(p.accent2Hex)
	p.warn = hexColor(p.warnHex)
	p.danger = hexColor(p.dangerHex)
	p.ok = hexColor(p.okHex)

	// Selection: a tinted band that keeps the text readable either way.
	selBgHex := blendHex(p.accentHex, bgHex, 0.35)
	p.selBg = hexColor(selBgHex)
	if contrastHex(fgHex, selBgHex) >= contrastHex(bgHex, selBgHex) {
		p.selFg = p.fg
	} else {
		p.selFg = p.bg
	}

	p.border = hexColor(blendHex(fgHex, bgHex, 0.30))

	return p
}

// applyPaletteToStyles pushes the palette into tview's global style struct so
// modals, inputs and buttons created later inherit it.
func (p uiPalette) applyPaletteToStyles() {
	tview.Styles.PrimitiveBackgroundColor = p.bg
	tview.Styles.ContrastBackgroundColor = p.selBg
	tview.Styles.MoreContrastBackgroundColor = p.accent
	tview.Styles.BorderColor = p.border
	tview.Styles.TitleColor = p.accent
	tview.Styles.GraphicsColor = p.border
	tview.Styles.PrimaryTextColor = p.fg
	tview.Styles.SecondaryTextColor = p.muted
	tview.Styles.TertiaryTextColor = p.accent2
	tview.Styles.InverseTextColor = p.bg
}

// --- colour maths helpers -------------------------------------------------

func normalizeHex(hex, fallback string) string {
	if hex == "" {
		return fallback
	}
	if hex[0] != '#' {
		hex = "#" + hex
	}
	if _, err := theme.HexToRGB(hex); err != nil {
		return fallback
	}
	// Themes ship a mix of #AABBCC and #aabbcc; one casing keeps the columns
	// calm and makes value comparisons meaningful.
	return strings.ToLower(hex)
}

func hexColor(hex string) tcell.Color {
	rgb, err := theme.HexToRGB(hex)
	if err != nil {
		return tcell.ColorDefault
	}
	return tcell.NewRGBColor(int32(rgb.R), int32(rgb.G), int32(rgb.B))
}

// blendHex mixes a towards b; ratio 0 returns a, ratio 1 returns b.
func blendHex(a, b string, ratio float64) string {
	ra, err1 := theme.HexToRGB(a)
	rb, err2 := theme.HexToRGB(b)
	if err1 != nil || err2 != nil {
		return a
	}
	mix := func(x, y int) int {
		return clampInt(int(float64(x)*(1-ratio)+float64(y)*ratio+0.5), 0, 255)
	}
	return theme.RGB{R: mix(ra.R, rb.R), G: mix(ra.G, rb.G), B: mix(ra.B, rb.B)}.ToHex()
}

// contrastHex returns the WCAG contrast ratio between two hex colours (1..21).
func contrastHex(a, b string) float64 {
	ra, err1 := theme.HexToRGB(a)
	rb, err2 := theme.HexToRGB(b)
	if err1 != nil || err2 != nil {
		return 1
	}
	la := relativeLuminance(ra.R, ra.G, ra.B)
	lb := relativeLuminance(rb.R, rb.G, rb.B)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// contrastGrade labels a ratio the way a theme author thinks about it.
func contrastGrade(ratio float64) string {
	switch {
	case ratio >= ContrastAAA:
		return "AAA"
	case ratio >= ContrastAA:
		return "AA"
	case ratio >= ContrastLow:
		return "AA+"
	default:
		return "low"
	}
}

// ensureReadable nudges the lightness of fg until it clears AA against bg.
func ensureReadable(fg, bg string) string {
	if contrastHex(fg, bg) >= ContrastAA {
		return fg
	}
	rgbBg, err := theme.HexToRGB(bg)
	if err != nil {
		return fg
	}
	rgbFg, err := theme.HexToRGB(fg)
	if err != nil {
		return fg
	}
	hsl := RGBToHSL(rgbFg.R, rgbFg.G, rgbFg.B)
	// Move away from the background's lightness.
	up := relativeLuminance(rgbBg.R, rgbBg.G, rgbBg.B) < 0.5
	for i := 0; i < 40; i++ {
		if up {
			hsl.L = math.Min(0.97, hsl.L+0.02)
		} else {
			hsl.L = math.Max(0.03, hsl.L-0.02)
		}
		r, g, b := HSLToRGB(hsl)
		candidate := theme.RGB{R: r, G: g, B: b}.ToHex()
		if contrastHex(candidate, bg) >= ContrastAA {
			return candidate
		}
	}
	// Last resort: plain black or white.
	if up {
		return "#ffffff"
	}
	return "#000000"
}

// describeColor renders the full readout used in the status bar. withContrast
// is false for the background itself, where the ratio would always read 1.0.
func describeColor(hex, bgHex string, withContrast bool) string {
	rgb, err := theme.HexToRGB(hex)
	if err != nil {
		return hex
	}
	hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)

	readout := fmt.Sprintf("%s  rgb(%d,%d,%d)  hsl(%.0f°,%.0f%%,%.0f%%)",
		hex, rgb.R, rgb.G, rgb.B, hsl.H, hsl.S*100, hsl.L*100)
	if !withContrast {
		return readout
	}

	ratio := contrastHex(hex, bgHex)
	return fmt.Sprintf("%s  %.1f:1 %s", readout, ratio, contrastGrade(ratio))
}
