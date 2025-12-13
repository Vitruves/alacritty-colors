package tui

import "time"

// UI adjustment constants
const (
	// Color adjustment steps
	BrightnessStep    = 10   // RGB value adjustment per keystroke
	HueStep           = 15.0 // Degrees of hue adjustment per keystroke
	MinSaturation     = 0.2  // Minimum saturation for visible hue changes
	SaturationStep    = 0.05 // Saturation adjustment per keystroke

	// Font size limits
	FontSizeMin  = 6.0
	FontSizeMax  = 48.0
	FontSizeStep = 0.5

	// Timing
	KeyDebounceDelay = 10 * time.Millisecond
	SpinnerInterval  = 100 * time.Millisecond

	// Display limits
	MaxFontNameDisplay = 30
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
	"Hex Colors",
	"Named Colors",
	"Bright Colors",
}

// Panel focus states
const (
	FocusThemeList  = 0
	FocusColorPanel = 1
)

// Theme markers
const (
	CurrentThemeMarker = "★ "
	EditedThemeSuffix  = "_edited_"
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

// Keybinding help text
const (
	StatusBarDefault     = "Tab: switch | ↑↓: navigate | ←→: brightness | n: new theme | g: random | /: search | ?: help | q: quit"
	StatusBarThemeList   = "Focus: Theme List | ↑↓: navigate | Enter: select | n: new | g: random | /: search | q: quit"
	StatusBarColorPanel  = "Focus: Color Panel | ↑↓: navigate | ←→: brightness | Shift+←→: hue | s: save | r: reset"
	StatusBarFontPanel   = "Tab: switch panels | ↑↓: navigate | ←→: size | Enter: apply | q: quit"
)

// Spinner characters for loading animation
var SpinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
