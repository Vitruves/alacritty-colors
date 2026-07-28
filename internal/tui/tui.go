package tui

import (
	"sync/atomic"

	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/internal/config"
	"github.com/vitruves/alacritty-colors/internal/font"
	"github.com/vitruves/alacritty-colors/internal/theme"
	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

// ColorEditor is the main TUI application for editing Alacritty themes
type ColorEditor struct {
	app               *tview.Application
	config            *config.Config
	themeManager      *theme.Manager
	fontManager       *font.Manager
	parametersManager *config.ParametersManager
	currentTheme      *alacritty.Config
	themeName         string
	appliedTheme      string
	exitNote          string // printed after the UI tears down

	// UI components
	themeList    *tview.List
	colorPanel   *tview.List
	previewPanel *tview.TextView
	statusInfo   *tview.TextView
	statusKeys   *tview.TextView
	filterInput  *tview.InputField
	rootFlex     *tview.Flex
	pages        *tview.Pages
	overlays     []string // stack of open overlay page names

	// Theme browsing state
	allThemes   []string        // every theme on disk, sorted
	visible     []string        // themes currently shown, after filter
	favorites   map[string]bool // in-memory, mirrors favorites.json
	filter      string
	favOnly     bool
	filterFocus bool

	// Color editing state
	colorValues        map[string]string
	originalValues     map[string]string // snapshot for undo/reset without a disk round-trip
	listItemToColorKey map[int]string
	colorKeyToListItem map[string]int
	isDirty            bool
	themeOwned         bool // theme file was written by this tool, so it may be overwritten
	discardArmed       bool // the unsaved-edits warning has been shown once
	colorMode          int
	focus              int
	palette            uiPalette

	// windowOpacity mirrors window.opacity from the main Alacritty config. Below
	// 1 the terminal blends its default background with whatever sits behind the
	// window, so no swatch drawn here can match what the user actually sees.
	windowOpacity float64

	// applySeq debounces "apply the theme under the cursor": only the newest
	// scheduled apply wins, so scrolling does not thrash the config file.
	applySeq atomic.Uint64

	// fontSeq does the same for the font browser, which writes the user's main
	// config rather than a theme file and so matters even more.
	fontSeq atomic.Uint64
}

// NewColorEditor creates a new ColorEditor instance
func NewColorEditor(cfg *config.Config) *ColorEditor {
	tm := theme.NewManager(cfg)
	tm.SetSilent(true)

	fm := font.NewManager(cfg)
	pm := config.NewParametersManager(cfg)

	editor := &ColorEditor{
		app:                tview.NewApplication(),
		config:             cfg,
		themeManager:       tm,
		fontManager:        fm,
		parametersManager:  pm,
		colorValues:        make(map[string]string),
		originalValues:     make(map[string]string),
		listItemToColorKey: make(map[int]string),
		colorKeyToListItem: make(map[string]int),
		appliedTheme:       tm.GetCurrentTheme(),
		colorMode:          ColorModeHex,
		windowOpacity:      readWindowOpacity(cfg.ConfigFile),
	}
	editor.favorites = editor.loadFavorites()

	return editor
}

// readWindowOpacity reports window.opacity from the main Alacritty config,
// defaulting to 1 when it is unset or the file cannot be read.
//
// A translucent window composites the default background with the desktop
// behind it. Cells that carry an explicit background — everything this editor
// paints — stay opaque, so the swatches would silently promise a background
// the terminal never shows, and the contrast ratios would be an upper bound.
func readWindowOpacity(configFile string) float64 {
	cfg, err := alacritty.NewParser().ParseFile(configFile)
	if err != nil || cfg.Window.Opacity <= 0 || cfg.Window.Opacity > 1 {
		return 1
	}
	return cfg.Window.Opacity
}

// Run starts the TUI application
func (ce *ColorEditor) Run() error {
	ce.setupUI()
	ce.loadThemes()
	ce.app.SetInputCapture(ce.handleGlobalKeys)
	// Shown after loadThemes so it can state how many themes are on disk, and
	// so the editor is fully built behind it.
	ce.showWelcome()

	return ce.app.Run()
}

// StartInteractive launches the interactive color editor. It returns a closing
// note for the caller to print, so the user sees what was kept on the way out.
func StartInteractive(cfg *config.Config) (string, error) {
	editor := NewColorEditor(cfg)
	err := editor.Run()
	return editor.exitNote, err
}
