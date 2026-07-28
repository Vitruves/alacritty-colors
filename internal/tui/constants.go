package tui

import "time"

// UI adjustment constants
const (
	// Color adjustment steps
	BrightnessStep = 8    // RGB value adjustment per keystroke
	HueStep        = 8.0  // Degrees of hue adjustment per keystroke
	MinSaturation  = 0.15 // Minimum saturation for visible hue changes
	SaturationStep = 0.04 // Saturation adjustment per keystroke
	LightnessStep  = 0.03 // HSL lightness adjustment per keystroke

	// Font size limits
	FontSizeMin  = 6.0
	FontSizeMax  = 48.0
	FontSizeStep = 0.5

	// Timing
	SpinnerInterval = 100 * time.Millisecond
	// ApplyDebounce is how long the cursor must rest on a theme before it is
	// written to disk and pushed to Alacritty. Without it, scrolling a list of
	// 150+ themes triggers one file write and one config reload per keystroke.
	ApplyDebounce = 180 * time.Millisecond
	// EditApplyDebounce coalesces rapid colour tweaks (held arrow key).
	EditApplyDebounce = 90 * time.Millisecond

	// Display limits
	MaxFontNameDisplay = 30
	ColorNameWidth     = 16
)

// Color mode constants
const (
	ColorModeHex    = 0
	ColorModeNamed  = 1
	ColorModeBright = 2
	ColorModeCount  = 3
)

// ColorModeNames maps color mode constants to display names
var ColorModeNames = []string{
	"Normal colors",
	"Terminal defaults",
	"Bright colors",
}

// Panel focus states
const (
	FocusThemeList = iota
	FocusColorPanel
	FocusPreview
	FocusCount
)

// Theme markers
const (
	CurrentThemeMarker = "●"
	FavoriteMarker     = "♥"
	EditedThemeSuffix  = "_edited_"
)

// WCAG contrast thresholds
const (
	ContrastAA  = 4.5
	ContrastAAA = 7.0
	// ContrastLow flags pairs that are hard to read at all.
	ContrastLow = 3.0
)

// Default fallback colors
var DefaultColors = map[string]string{
	"blue":    "#0000ff",
	"cyan":    "#00ffff",
	"yellow":  "#ffff00",
	"white":   "#ffffff",
	"magenta": "#ff00ff",
	"green":   "#00ff00",
	"red":     "#ff0000",
	"black":   "#000000",
}

// Color section order for consistent display
var ColorSectionOrder = []string{
	"Primary",
	"Cursor",
	"Selection",
	"Normal",
	"Bright",
	"Dim",
}

// Base color names in standard order
var BaseColorNames = []string{
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",
}

// Contextual key hints, one per focused panel. Tab and a lead every line: they
// are the two keys used constantly and in every panel, so they hold the same
// position wherever you are. The rest is what that panel is for.
var FocusKeyHints = []string{
	FocusThemeList:  "Tab next column · a apply · ↑↓ browse · / search · * favourite · F favourites · n create · g random · ? keys · q quit",
	FocusColorPanel: "Tab next column · a apply · ←→ lighter · Shift+←→ hue · -/+ saturation · [ ] lightness · Enter exact hex · s save · ? keys",
	FocusPreview:    "Tab next column · a apply · ↑↓ scroll · c colour mode · ? keys · q quit",
}

// WelcomeWidth is the width of the opening screen, chosen so the widest
// feature line sits comfortably inside it without wrapping.
const WelcomeWidth = 66

// StatusBarFontPanel is the hint line of the font browser.
const StatusBarFontPanel = "Tab next column · ↑↓ navigate · ←→ size · d download more · Esc back"

// fontSizes are the sizes offered in the font browser's size column. The step
// is half a point through the range people actually read at, then whole points
// once the exact value stops mattering.
var fontSizes = buildFontSizes()

func buildFontSizes() []float64 {
	var sizes []float64
	for size := FontSizeMin; size <= 20.0; size += FontSizeStep {
		sizes = append(sizes, size)
	}
	for size := 21.0; size <= FontSizeMax; size += 1.0 {
		sizes = append(sizes, size)
	}
	return sizes
}

// Spinner characters for loading animation
var SpinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
