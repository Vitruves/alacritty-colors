package tui

import (
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

	// UI components
	themeList    *tview.List
	colorPanel   *tview.List
	previewPanel *tview.TextView
	statusBar    *tview.TextView

	// Color editing state
	colorValues        map[string]string
	colorKeys          []string
	listItemToColorKey map[int]string
	isDirty            bool
	colorMode          int
	isApplying         bool
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
		colorKeys:          make([]string, 0),
		listItemToColorKey: make(map[int]string),
		appliedTheme:       tm.GetCurrentTheme(),
		colorMode:          ColorModeHex,
	}

	return editor
}

// Run starts the TUI application
func (ce *ColorEditor) Run() error {
	ce.setupUI()
	ce.loadThemes()
	ce.app.SetInputCapture(ce.handleGlobalKeys)

	return ce.app.Run()
}

// StartInteractive launches the interactive color editor
func StartInteractive(cfg *config.Config) error {
	editor := NewColorEditor(cfg)
	return editor.Run()
}
