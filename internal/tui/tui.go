package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/internal/config"
	"github.com/vitruves/alacritty-colors/internal/theme"
	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

type ColorEditor struct {
	app          *tview.Application
	config       *config.Config
	themeManager *theme.Manager
	currentTheme *alacritty.Config
	themeName    string

	// UI components
	themeList    *tview.List
	colorPanel   *tview.List
	previewPanel *tview.TextView
	statusBar    *tview.TextView

	// Color editing state
	colorValues map[string]string
	colorKeys   []string
	isDirty     bool
}

func NewColorEditor(cfg *config.Config) *ColorEditor {
	tm := theme.NewManager(cfg)

	editor := &ColorEditor{
		app:          tview.NewApplication(),
		config:       cfg,
		themeManager: tm,
		colorValues:  make(map[string]string),
		colorKeys:    make([]string, 0),
	}

	// Theme will be applied in setupUI()

	return editor
}

func (ce *ColorEditor) Run() error {
	// Initialize UI
	ce.setupUI()
	ce.loadThemes()

	// Set up key bindings
	ce.app.SetInputCapture(ce.handleGlobalKeys)

	return ce.app.Run()
}

func (ce *ColorEditor) setupUI() {
	// Apply user theme to TUI first
	ce.applyUserThemeToTUI()

	// Theme list on the left
	ce.themeList = tview.NewList()
	ce.themeList.ShowSecondaryText(false)
	ce.themeList.SetMainTextColor(tcell.ColorWhite)
	ce.themeList.SetSelectedTextColor(tcell.ColorBlack)
	ce.themeList.SetSelectedBackgroundColor(tcell.ColorWhite)
	ce.themeList.SetBorder(true)
	ce.themeList.SetTitle(" Themes ")
	ce.themeList.SetSelectedFunc(ce.onThemeSelected)
	ce.themeList.SetInputCapture(ce.handleThemeListKeys)

	// Color editing panel in the center
	ce.colorPanel = tview.NewList()
	ce.colorPanel.ShowSecondaryText(false)
	ce.colorPanel.SetMainTextColor(tcell.ColorWhite)
	ce.colorPanel.SetSelectedTextColor(tcell.ColorBlack)
	ce.colorPanel.SetSelectedBackgroundColor(tcell.ColorWhite)
	ce.colorPanel.SetBorder(true)
	ce.colorPanel.SetTitle(" Color Palette ")
	ce.colorPanel.SetInputCapture(ce.handleColorPanelKeys)
	ce.colorPanel.SetSelectedFunc(ce.onColorSelected)

	// Preview panel on the right
	ce.previewPanel = tview.NewTextView()
	ce.previewPanel.SetDynamicColors(true)
	ce.previewPanel.SetWordWrap(true)
	ce.previewPanel.SetBorder(true)
	ce.previewPanel.SetTitle(" Preview ")

	// Status bar at bottom
	ce.statusBar = tview.NewTextView()
	ce.statusBar.SetText("Tab: switch panels | ↑↓: navigate | ←→: adjust RGB values | Enter: edit | q: quit | s: save | r: reset")
	ce.statusBar.SetTextColor(tcell.ColorYellow)

	// Layout - just use theme list as left panel
	leftPanel := ce.themeList

	centerPanel := ce.colorPanel
	rightPanel := ce.previewPanel

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(leftPanel, 0, 1, false)
	mainFlex.AddItem(centerPanel, 0, 2, false)
	mainFlex.AddItem(rightPanel, 0, 1, false)

	rootFlex := tview.NewFlex()
	rootFlex.SetDirection(tview.FlexRow)
	rootFlex.AddItem(mainFlex, 0, 1, true)
	rootFlex.AddItem(ce.statusBar, 1, 0, false)

	ce.app.SetRoot(rootFlex, true)
	ce.app.SetFocus(ce.themeList)
	ce.themeList.SetBorderColor(tcell.ColorYellow)
}

func (ce *ColorEditor) loadThemes() {
	// Get theme files directly
	themeFiles, err := ce.getThemeFiles()
	if err != nil {
		ce.setStatus("Error loading themes: " + err.Error())
		return
	}

	sort.Strings(themeFiles)

	for _, themeName := range themeFiles {
		ce.themeList.AddItem(themeName, "", 0, nil)
	}

	if len(themeFiles) > 0 {
		ce.themeList.SetCurrentItem(0)
		// Auto-load the first theme
		ce.onThemeSelected(0, themeFiles[0], "", 0)
	}
}

func (ce *ColorEditor) getThemeFiles() ([]string, error) {
	files, err := os.ReadDir(ce.config.ThemesDir)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".toml") && file.Name() != "current.toml" {
			name := strings.TrimSuffix(file.Name(), ".toml")
			themes = append(themes, name)
		}
	}

	return themes, nil
}

func (ce *ColorEditor) onThemeSelected(index int, themeName string, _ string, _ rune) {
	ce.themeName = themeName
	ce.loadTheme(themeName)
	ce.buildColorPanel()
	ce.updatePreview()
	if len(ce.colorKeys) > 0 {
		ce.colorPanel.SetCurrentItem(0)
		ce.updateColorStatus()
	}
}

func (ce *ColorEditor) loadTheme(themeName string) {
	// Load theme file
	themeFile := ce.config.GetThemePath(themeName)
	parser := alacritty.NewParser()

	config, err := parser.ParseFile(themeFile)
	if err != nil {
		ce.setStatus(fmt.Sprintf("Error loading theme %s: %v", themeName, err))
		return
	}

	ce.currentTheme = config
	ce.extractColors()
	ce.isDirty = false
}

func (ce *ColorEditor) extractColors() {
	ce.colorValues = make(map[string]string)
	ce.colorKeys = make([]string, 0)

	if ce.currentTheme == nil {
		return
	}

	// Primary colors
	ce.addColor("primary.background", ce.currentTheme.Colors.Primary.Background)
	ce.addColor("primary.foreground", ce.currentTheme.Colors.Primary.Foreground)

	// Cursor colors
	if ce.currentTheme.Colors.Cursor.Text != "" {
		ce.addColor("cursor.text", ce.currentTheme.Colors.Cursor.Text)
	}
	if ce.currentTheme.Colors.Cursor.Cursor != "" {
		ce.addColor("cursor.cursor", ce.currentTheme.Colors.Cursor.Cursor)
	}

	// Selection colors
	if ce.currentTheme.Colors.Selection.Text != "" {
		ce.addColor("selection.text", ce.currentTheme.Colors.Selection.Text)
	}
	if ce.currentTheme.Colors.Selection.Background != "" {
		ce.addColor("selection.background", ce.currentTheme.Colors.Selection.Background)
	}

	// Normal colors
	for name, color := range ce.currentTheme.Colors.Normal {
		ce.addColor("normal."+name, color)
	}

	// Bright colors
	for name, color := range ce.currentTheme.Colors.Bright {
		ce.addColor("bright."+name, color)
	}

	// Dim colors
	for name, color := range ce.currentTheme.Colors.Dim {
		ce.addColor("dim."+name, color)
	}
}

func (ce *ColorEditor) addColor(key, value string) {
	if value != "" {
		ce.colorValues[key] = value
		ce.colorKeys = append(ce.colorKeys, key)
	}
}

func (ce *ColorEditor) buildColorPanel() {
	ce.colorPanel.Clear()

	if ce.currentTheme == nil {
		ce.colorPanel.AddItem("Select a theme to start editing", "", 0, nil)
		return
	}

	sections := map[string][]string{
		"Primary":   {"primary.background", "primary.foreground"},
		"Cursor":    {"cursor.text", "cursor.cursor"},
		"Selection": {"selection.text", "selection.background"},
		"Normal":    {},
		"Bright":    {},
		"Dim":       {},
	}

	// Populate normal, bright, and dim sections
	colors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	for _, color := range colors {
		sections["Normal"] = append(sections["Normal"], "normal."+color)
		sections["Bright"] = append(sections["Bright"], "bright."+color)
		sections["Dim"] = append(sections["Dim"], "dim."+color)
	}

	// Define order to ensure consistent display
	sectionOrder := []string{"Primary", "Cursor", "Selection", "Normal", "Bright", "Dim"}

	for _, sectionName := range sectionOrder {
		keys := sections[sectionName]
		if len(keys) == 0 {
			continue
		}

		// Check if section has any values
		hasValues := false
		for _, key := range keys {
			if _, exists := ce.colorValues[key]; exists {
				hasValues = true
				break
			}
		}

		if !hasValues {
			continue
		}

		// Add section header
		ce.colorPanel.AddItem(fmt.Sprintf("[cyan::b]%s[-]", sectionName), "", 0, nil)

		for _, key := range keys {
			if value, exists := ce.colorValues[key]; exists {
				// Create color preview
				colorValue := value
				if !strings.HasPrefix(colorValue, "#") && len(colorValue) == 6 {
					colorValue = "#" + colorValue
				}

				// Convert to RGB for display
				rgbDisplay := ""
				if rgb, err := theme.HexToRGB(colorValue); err == nil {
					rgbDisplay = fmt.Sprintf("R:%d G:%d B:%d", rgb.R, rgb.G, rgb.B)
				} else {
					rgbDisplay = colorValue
				}

				displayName := strings.Replace(key, ".", " ", -1)
				text := fmt.Sprintf("  [%s]██[-] %-20s %s", colorValue, displayName, rgbDisplay)

				ce.colorPanel.AddItem(text, "", 0, nil)
			}
		}
	}
}

func (ce *ColorEditor) handleThemeListKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		ce.app.SetFocus(ce.colorPanel)
		ce.colorPanel.SetBorderColor(tcell.ColorYellow)
		ce.themeList.SetBorderColor(tcell.ColorDefault)
		ce.setStatus("Focus: Color Panel | Use arrow keys to navigate, Enter to edit")
		return nil
	case tcell.KeyEnter:
		// Select current theme
		index := ce.themeList.GetCurrentItem()
		if index >= 0 {
			themeName, _ := ce.themeList.GetItemText(index)
			ce.onThemeSelected(index, themeName, "", 0)
			ce.app.SetFocus(ce.colorPanel)
			ce.colorPanel.SetBorderColor(tcell.ColorYellow)
			ce.themeList.SetBorderColor(tcell.ColorDefault)
		}
		return nil
	}
	return event
}

func (ce *ColorEditor) onColorSelected(index int, text string, _ string, _ rune) {
	// Update color status when selecting
	ce.updateColorStatus()
}

func (ce *ColorEditor) getColorIndexFromListIndex(listIndex int) int {
	// Skip section headers to find actual color items
	colorIndex := 0
	for i := 0; i <= listIndex; i++ {
		text, _ := ce.colorPanel.GetItemText(i)
		// If it's not a section header (doesn't start with space), it's a color item
		if !strings.HasPrefix(text, "  ") {
			// This is a section header, don't count it
			if i == listIndex {
				return -1 // Selected a header
			}
		} else {
			// This is a color item
			if i == listIndex {
				return colorIndex
			}
			colorIndex++
		}
	}
	return -1
}

func (ce *ColorEditor) handleColorPanelKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		ce.app.SetFocus(ce.themeList)
		ce.themeList.SetBorderColor(tcell.ColorYellow)
		ce.colorPanel.SetBorderColor(tcell.ColorDefault)
		ce.setStatus("Focus: Theme List | Use arrow keys to navigate, Enter to select")
		return nil
	case tcell.KeyEnter:
		index := ce.colorPanel.GetCurrentItem()
		ce.onColorSelected(index, "", "", 0)
		return nil
	case tcell.KeyUp, tcell.KeyDown:
		// Use Up/Down for navigation between items
		result := event // Let tview handle the navigation
		go func() {
			time.Sleep(10 * time.Millisecond)
			ce.app.QueueUpdateDraw(func() {
				ce.updateColorStatus()
			})
		}()
		return result
	case tcell.KeyLeft, tcell.KeyRight:
		// Check if we're on a color item (not a section header)
		index := ce.colorPanel.GetCurrentItem()
		colorIndex := ce.getColorIndexFromListIndex(index)

		// If on a section header, do normal navigation
		if colorIndex < 0 {
			return event
		}

		// If on a color item, adjust the color with Left/Right
		if colorIndex >= 0 && colorIndex < len(ce.colorKeys) {
			colorKey := ce.colorKeys[colorIndex]
			ce.adjustColorWithArrows(colorKey, event.Key())
			return nil
		}

		return event
	}
	return event
}

func (ce *ColorEditor) updatePreview() {
	if ce.currentTheme == nil {
		return
	}

	preview := ce.generatePreview()
	ce.previewPanel.SetText(preview)
}

func (ce *ColorEditor) generatePreview() string {
	var preview strings.Builder

	preview.WriteString("[yellow::b]Terminal Preview[-]\n\n")

	// Color test
	colors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	preview.WriteString("[white::b]Normal colors:[-]\n")
	for _, color := range colors {
		if colorVal, exists := ce.colorValues["normal."+color]; exists {
			preview.WriteString(fmt.Sprintf("[%s]██[-] ", colorVal))
		}
	}
	preview.WriteString("\n\n")

	preview.WriteString("[white::b]Bright colors:[-]\n")
	for _, color := range colors {
		if colorVal, exists := ce.colorValues["bright."+color]; exists {
			preview.WriteString(fmt.Sprintf("[%s]██[-] ", colorVal))
		}
	}
	preview.WriteString("\n\n")

	// Sample terminal output
	preview.WriteString("[white::b]Sample Output:[-]\n")
	if greenColor, exists := ce.colorValues["normal.green"]; exists {
		preview.WriteString(fmt.Sprintf("[%s]$ ls -la[-]\n", greenColor))
	}

	// Sample file listing
	if blueColor, exists := ce.colorValues["normal.blue"]; exists {
		if cyanColor, exists := ce.colorValues["normal.cyan"]; exists {
			if yellowColor, exists := ce.colorValues["normal.yellow"]; exists {
				if whiteColor, exists := ce.colorValues["normal.white"]; exists {
					if magentaColor, exists := ce.colorValues["normal.magenta"]; exists {
						preview.WriteString(fmt.Sprintf("[%s]drwxr-xr-x[-] [%s]5[-] [%s]user[-] [%s]group[-] [%s]4096[-] [%s]Jan 15 10:30[-] [%s].[-]\n",
							blueColor, cyanColor, yellowColor, yellowColor, whiteColor, magentaColor, whiteColor))
						if greenColor, exists := ce.colorValues["normal.green"]; exists {
							preview.WriteString(fmt.Sprintf("[%s]-rw-r--r--[-] [%s]1[-] [%s]user[-] [%s]group[-] [%s]1234[-] [%s]Jan 15 10:25[-] [%s]file.txt[-]\n",
								whiteColor, cyanColor, yellowColor, yellowColor, whiteColor, magentaColor, greenColor))
						}
					}
				}
			}
		}
	}

	return preview.String()
}

func (ce *ColorEditor) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlC:
		ce.app.Stop()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q', 'Q':
			if ce.isDirty {
				ce.confirmQuit()
			} else {
				ce.app.Stop()
			}
			return nil
		case 's', 'S':
			ce.saveTheme()
			return nil
		case 'r', 'R':
			ce.resetTheme()
			return nil
		}
	}
	return event
}

func (ce *ColorEditor) confirmQuit() {
	// Apply user's theme to the modal
	ce.applyUserThemeToTUI()

	modal := tview.NewModal()
	modal.SetText("You have unsaved changes. Are you sure you want to quit?")
	modal.AddButtons([]string{"Save & Quit", "Quit", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonIndex {
		case 0: // Save & Quit
			ce.saveTheme()
			ce.app.Stop()
		case 1: // Quit
			ce.app.Stop()
		case 2: // Cancel
			// Reset theme styles to default and return to main view
			ce.resetTUITheme()
			ce.setupUI()
			ce.buildColorPanel()
			ce.updatePreview()
		}
	})

	ce.app.SetRoot(modal, true)
}

func (ce *ColorEditor) saveTheme() {
	if ce.currentTheme == nil || ce.themeName == "" {
		ce.setStatus("No theme selected to save")
		return
	}

	// Update the current theme with new values
	ce.updateThemeConfig()

	// Save to file
	err := ce.saveThemeToFile()
	if err != nil {
		ce.setStatus(fmt.Sprintf("Error saving theme: %v", err))
		return
	}

	ce.isDirty = false
	ce.setStatus(fmt.Sprintf("Theme '%s' saved successfully", ce.themeName))
}

func (ce *ColorEditor) saveThemeToFile() error {
	themeFile := ce.config.GetThemePath(ce.themeName)

	// Generate TOML content
	content := ce.generateTOMLContent()

	return os.WriteFile(themeFile, []byte(content), 0644)
}

func (ce *ColorEditor) generateTOMLContent() string {
	var content strings.Builder

	content.WriteString("# Alacritty theme - edited with alacritty-colors TUI\n\n")

	// Primary colors
	content.WriteString("[colors.primary]\n")
	content.WriteString(fmt.Sprintf("background = \"%s\"\n", ce.colorValues["primary.background"]))
	content.WriteString(fmt.Sprintf("foreground = \"%s\"\n", ce.colorValues["primary.foreground"]))
	content.WriteString("\n")

	// Cursor colors
	if ce.colorValues["cursor.text"] != "" || ce.colorValues["cursor.cursor"] != "" {
		content.WriteString("[colors.cursor]\n")
		if ce.colorValues["cursor.text"] != "" {
			content.WriteString(fmt.Sprintf("text = \"%s\"\n", ce.colorValues["cursor.text"]))
		}
		if ce.colorValues["cursor.cursor"] != "" {
			content.WriteString(fmt.Sprintf("cursor = \"%s\"\n", ce.colorValues["cursor.cursor"]))
		}
		content.WriteString("\n")
	}

	// Selection colors
	if ce.colorValues["selection.text"] != "" || ce.colorValues["selection.background"] != "" {
		content.WriteString("[colors.selection]\n")
		if ce.colorValues["selection.text"] != "" {
			content.WriteString(fmt.Sprintf("text = \"%s\"\n", ce.colorValues["selection.text"]))
		}
		if ce.colorValues["selection.background"] != "" {
			content.WriteString(fmt.Sprintf("background = \"%s\"\n", ce.colorValues["selection.background"]))
		}
		content.WriteString("\n")
	}

	// Normal colors
	content.WriteString("[colors.normal]\n")
	colors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	for _, color := range colors {
		if value, exists := ce.colorValues["normal."+color]; exists {
			content.WriteString(fmt.Sprintf("%s = \"%s\"\n", color, value))
		}
	}
	content.WriteString("\n")

	// Bright colors
	content.WriteString("[colors.bright]\n")
	for _, color := range colors {
		if value, exists := ce.colorValues["bright."+color]; exists {
			content.WriteString(fmt.Sprintf("%s = \"%s\"\n", color, value))
		}
	}
	content.WriteString("\n")

	// Dim colors (if any)
	hasDimColors := false
	for _, color := range colors {
		if _, exists := ce.colorValues["dim."+color]; exists {
			hasDimColors = true
			break
		}
	}

	if hasDimColors {
		content.WriteString("[colors.dim]\n")
		for _, color := range colors {
			if value, exists := ce.colorValues["dim."+color]; exists {
				content.WriteString(fmt.Sprintf("%s = \"%s\"\n", color, value))
			}
		}
		content.WriteString("\n")
	}

	return content.String()
}

func (ce *ColorEditor) updateThemeConfig() {
	// Update primary colors
	ce.currentTheme.Colors.Primary.Background = ce.colorValues["primary.background"]
	ce.currentTheme.Colors.Primary.Foreground = ce.colorValues["primary.foreground"]

	// Update cursor colors
	if ce.colorValues["cursor.text"] != "" {
		ce.currentTheme.Colors.Cursor.Text = ce.colorValues["cursor.text"]
	}
	if ce.colorValues["cursor.cursor"] != "" {
		ce.currentTheme.Colors.Cursor.Cursor = ce.colorValues["cursor.cursor"]
	}

	// Update selection colors
	if ce.colorValues["selection.text"] != "" {
		ce.currentTheme.Colors.Selection.Text = ce.colorValues["selection.text"]
	}
	if ce.colorValues["selection.background"] != "" {
		ce.currentTheme.Colors.Selection.Background = ce.colorValues["selection.background"]
	}

	// Update normal colors
	for name := range ce.currentTheme.Colors.Normal {
		if value, exists := ce.colorValues["normal."+name]; exists {
			ce.currentTheme.Colors.Normal[name] = value
		}
	}

	// Update bright colors
	for name := range ce.currentTheme.Colors.Bright {
		if value, exists := ce.colorValues["bright."+name]; exists {
			ce.currentTheme.Colors.Bright[name] = value
		}
	}

	// Update dim colors
	for name := range ce.currentTheme.Colors.Dim {
		if value, exists := ce.colorValues["dim."+name]; exists {
			ce.currentTheme.Colors.Dim[name] = value
		}
	}
}

func (ce *ColorEditor) resetTheme() {
	if ce.themeName == "" {
		return
	}

	ce.loadTheme(ce.themeName)
	ce.buildColorPanel()
	ce.updatePreview()
	if len(ce.colorKeys) > 0 {
		ce.colorPanel.SetCurrentItem(0)
	}
	ce.setStatus("Theme reset to original values")
}

func (ce *ColorEditor) setStatus(message string) {
	ce.statusBar.SetText(message)
}

func (ce *ColorEditor) updateColorStatus() {
	index := ce.colorPanel.GetCurrentItem()
	colorIndex := ce.getColorIndexFromListIndex(index)
	if colorIndex >= 0 && colorIndex < len(ce.colorKeys) {
		colorKey := ce.colorKeys[colorIndex]
		colorValue := ce.colorValues[colorKey]
		displayName := strings.Replace(colorKey, ".", " ", -1)
		// Convert hex to RGB for display in status
		rgbDisplay := colorValue
		if rgb, err := theme.HexToRGB(colorValue); err == nil {
			rgbDisplay = fmt.Sprintf("R:%d G:%d B:%d", rgb.R, rgb.G, rgb.B)
		}
		ce.setStatus(fmt.Sprintf("Selected: %s (%s) | ←→: adjust RGB | Enter: edit | Tab: switch panels", displayName, rgbDisplay))
	}
}

func (ce *ColorEditor) adjustColorWithArrows(colorKey string, key tcell.Key) {
	currentValue := ce.colorValues[colorKey]
	rgb, err := theme.HexToRGB(currentValue)
	if err != nil {
		return
	}

	// Adjust RGB values directly with left/right arrows
	adjustment := 10 // RGB step size
	if key == tcell.KeyRight {
		// Increase RGB values (brighter)
		rgb.R = min(255, rgb.R+adjustment)
		rgb.G = min(255, rgb.G+adjustment)
		rgb.B = min(255, rgb.B+adjustment)
	} else if key == tcell.KeyLeft {
		// Decrease RGB values (darker)
		rgb.R = max(0, rgb.R-adjustment)
		rgb.G = max(0, rgb.G-adjustment)
		rgb.B = max(0, rgb.B-adjustment)
	}

	newHex := rgb.ToHex()
	ce.colorValues[colorKey] = newHex
	ce.isDirty = true

	// Update just the current item in place instead of rebuilding the whole panel
	currentIndex := ce.colorPanel.GetCurrentItem()

	// Update the current list item with the new color
	colorValue := newHex
	if !strings.HasPrefix(colorValue, "#") && len(colorValue) == 6 {
		colorValue = "#" + colorValue
	}

	// Convert to RGB for display
	rgbDisplay := fmt.Sprintf("R:%d G:%d B:%d", rgb.R, rgb.G, rgb.B)
	displayName := strings.Replace(colorKey, ".", " ", -1)
	text := fmt.Sprintf("  [%s]██[-] %-20s %s", colorValue, displayName, rgbDisplay)

	// Update the current item
	ce.colorPanel.SetItemText(currentIndex, text, "")

	// Update preview and status
	ce.updatePreview()
	ce.updateColorStatus()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (ce *ColorEditor) applyUserThemeToTUI() {
	// Try to load current alacritty config to get theme colors
	parser := alacritty.NewParser()
	currentConfig, err := parser.ParseFile(ce.config.ConfigFile)
	if err != nil {
		// If can't load main config, try current.toml
		currentThemePath := ce.config.GetThemePath("current")
		currentConfig, err = parser.ParseFile(currentThemePath)
		if err != nil {
			// If still can't load, use default colors
			return
		}
	}

	// Use actual colors from the current Alacritty config
	ce.applyAlacrittyColors(currentConfig)
}

func (ce *ColorEditor) applyAlacrittyColors(config *alacritty.Config) {
	// Convert hex colors to tcell colors
	bgColor := ce.hexToTcellColor(config.Colors.Primary.Background)
	fgColor := ce.hexToTcellColor(config.Colors.Primary.Foreground)

	// Use normal colors for accents
	greenColor := ce.hexToTcellColor(config.Colors.Normal["green"])
	yellowColor := ce.hexToTcellColor(config.Colors.Normal["yellow"])
	blueColor := ce.hexToTcellColor(config.Colors.Normal["blue"])
	cyanColor := ce.hexToTcellColor(config.Colors.Normal["cyan"])

	// Apply the actual Alacritty colors to TUI
	tview.Styles.PrimitiveBackgroundColor = bgColor
	tview.Styles.ContrastBackgroundColor = blueColor
	tview.Styles.MoreContrastBackgroundColor = greenColor
	tview.Styles.BorderColor = fgColor
	tview.Styles.TitleColor = fgColor
	tview.Styles.GraphicsColor = fgColor
	tview.Styles.PrimaryTextColor = fgColor
	tview.Styles.SecondaryTextColor = yellowColor
	tview.Styles.TertiaryTextColor = cyanColor
	tview.Styles.InverseTextColor = bgColor
}

func (ce *ColorEditor) hexToTcellColor(hexColor string) tcell.Color {
	if hexColor == "" {
		return tcell.ColorDefault
	}

	rgb, err := theme.HexToRGB(hexColor)
	if err != nil {
		return tcell.ColorDefault
	}

	return tcell.NewRGBColor(int32(rgb.R), int32(rgb.G), int32(rgb.B))
}

func (ce *ColorEditor) resetTUITheme() {
	// Reset to default tview theme colors
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlue
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorGreen
	tview.Styles.BorderColor = tcell.ColorDefault
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorYellow
	tview.Styles.TertiaryTextColor = tcell.ColorGreen
	tview.Styles.InverseTextColor = tcell.ColorBlue
}

// StartInteractive launches the interactive color editor
func StartInteractive(cfg *config.Config) error {
	editor := NewColorEditor(cfg)
	return editor.Run()
}
