package tui

import (
	"fmt"
	"math"
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
	appliedTheme string // Currently applied theme in Alacritty

	// UI components
	themeList    *tview.List
	colorPanel   *tview.List
	previewPanel *tview.TextView
	statusBar    *tview.TextView

	// Color editing state
	colorValues map[string]string
	colorKeys   []string
	listItemToColorKey map[int]string // Maps list item index to color key
	isDirty     bool
}

func NewColorEditor(cfg *config.Config) *ColorEditor {
	tm := theme.NewManager(cfg)
	tm.SetSilent(true) // Suppress console output in TUI mode

	editor := &ColorEditor{
		app:          tview.NewApplication(),
		config:       cfg,
		themeManager: tm,
		colorValues:  make(map[string]string),
		colorKeys:    make([]string, 0),
		listItemToColorKey: make(map[int]string),
		appliedTheme: tm.GetCurrentTheme(), // Get the currently applied theme
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
	ce.statusBar.SetText("Tab: switch panels | ↑↓: navigate | ←→: brightness | Shift+←→: hue/tint | Enter: apply theme | a: apply theme | q: quit | s: save | r: reset")
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

	currentThemeIndex := -1
	for i, themeName := range themeFiles {
		displayName := themeName
		if themeName == ce.appliedTheme {
			displayName = fmt.Sprintf("★ %s", themeName) // Mark current theme with a star
			currentThemeIndex = i
		}
		ce.themeList.AddItem(displayName, "", 0, nil)
	}

	if len(themeFiles) > 0 {
		// Set current item to the applied theme if found, otherwise first theme
		if currentThemeIndex >= 0 {
			ce.themeList.SetCurrentItem(currentThemeIndex)
			ce.onThemeSelected(currentThemeIndex, themeFiles[currentThemeIndex], "", 0)
		} else {
			ce.themeList.SetCurrentItem(0)
			// Auto-load the first theme
			ce.onThemeSelected(0, themeFiles[0], "", 0)
		}
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
	// Remove star prefix if present
	themeName = strings.TrimPrefix(themeName, "★ ")
	
	ce.themeName = themeName
	ce.loadTheme(themeName)
	ce.buildColorPanel()
	ce.updatePreview()
	if len(ce.colorKeys) > 0 {
		ce.colorPanel.SetCurrentItem(0)
		ce.updateColorStatus()
	}
	
	// Apply theme in real-time
	go func() {
		if err := ce.themeManager.ApplyTheme(themeName); err != nil {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus(fmt.Sprintf("Failed to apply theme: %v", err))
				ce.app.ForceDraw()
			})
		} else {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus(fmt.Sprintf("Applied theme: %s", themeName))
				ce.app.ForceDraw()
			})
		}
	}()
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
	ce.listItemToColorKey = make(map[int]string) // Reset the mapping

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

	itemIndex := 0
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
		itemIndex++ // Section headers don't have color keys

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

				// Map this list item index to the color key
				ce.listItemToColorKey[itemIndex] = key
				
				ce.colorPanel.AddItem(text, "", 0, nil)
				itemIndex++
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
			// Clean theme name
			themeName = strings.TrimPrefix(themeName, "★ ")
			ce.onThemeSelected(index, themeName, "", 0)
			ce.app.SetFocus(ce.colorPanel)
			ce.colorPanel.SetBorderColor(tcell.ColorYellow)
			ce.themeList.SetBorderColor(tcell.ColorDefault)
		}
		return nil
	case tcell.KeyUp, tcell.KeyDown:
		// Allow navigation and apply theme on selection change
		result := event
		go func() {
			time.Sleep(10 * time.Millisecond)
			ce.app.QueueUpdateDraw(func() {
				index := ce.themeList.GetCurrentItem()
				if index >= 0 {
					themeName, _ := ce.themeList.GetItemText(index)
					themeName = strings.TrimPrefix(themeName, "★ ")
					ce.onThemeSelected(index, themeName, "", 0)
				}
				ce.app.ForceDraw()
			})
		}()
		return result
	}
	return event
}

func (ce *ColorEditor) onColorSelected(index int, text string, _ string, _ rune) {
	// Update color status when selecting
	ce.updateColorStatus()
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
		colorKey, exists := ce.listItemToColorKey[index]

		// If on a section header, do normal navigation
		if !exists {
			return event
		}

		// Check for Shift modifier for hue adjustment
		if event.Modifiers()&tcell.ModShift != 0 {
			// Shift+Left/Right for hue adjustment
			ce.adjustColorHue(colorKey, event.Key())
		} else {
			// Normal Left/Right for brightness adjustment
			ce.adjustColorWithArrows(colorKey, event.Key())
		}
		return nil
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
		case 'a', 'A':
			ce.applyCurrentTheme()
			return nil
		}
	}
	return event
}

func (ce *ColorEditor) confirmQuit() {
	// Temporarily reset to default colors for the modal to ensure visibility
	ce.resetTUITheme()

	modal := tview.NewModal()
	modal.SetText("You have unsaved changes. Are you sure you want to quit?")
	modal.AddButtons([]string{"Save & Quit", "Quit", "Cancel"})

	// Style the modal with high contrast colors
	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorBlue)
	modal.SetButtonTextColor(tcell.ColorWhite)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		switch buttonIndex {
		case 0: // Save & Quit
			ce.saveTheme()
			ce.app.Stop()
		case 1: // Quit
			ce.app.Stop()
		case 2: // Cancel
			// Restore user theme and return to main view
			ce.applyUserThemeToTUI()
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

	// Write to file with proper error handling
	err := os.WriteFile(themeFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write theme file %s: %w", themeFile, err)
	}

	// Force sync to disk
	file, err := os.OpenFile(themeFile, os.O_RDONLY, 0)
	if err == nil {
		file.Sync()
		file.Close()
	}

	return nil
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

func (ce *ColorEditor) applyCurrentTheme() {
	if ce.themeName == "" {
		ce.setStatus("No theme selected to apply")
		return
	}

	// Save the theme first if it has been modified
	if ce.isDirty {
		ce.saveTheme()
	}

	// Apply the theme
	if err := ce.themeManager.ApplyTheme(ce.themeName); err != nil {
		ce.setStatus(fmt.Sprintf("Failed to apply theme: %v", err))
	} else {
		ce.appliedTheme = ce.themeName
		ce.setStatus(fmt.Sprintf("Theme '%s' applied successfully", ce.themeName))
		// Refresh the theme list to update the star
		ce.refreshThemeList()
	}
}

func (ce *ColorEditor) createThemeCopy() {
	if ce.themeName == "" {
		return
	}
	
	// Generate a unique name for the copy
	originalName := ce.themeName
	timestamp := time.Now().Format("150405") // HHMMSS format
	copyName := fmt.Sprintf("%s_edited_%s", originalName, timestamp)
	
	// Update the theme name to the copy
	ce.themeName = copyName
	
	// Update the color panel title to show it's a copy
	ce.colorPanel.SetTitle(fmt.Sprintf(" Color Palette - %s (Copy) ", copyName))
	
	// Save the copy with current values and apply it immediately
	go func() {
		ce.updateThemeConfig()
		if err := ce.saveThemeToFile(); err == nil {
			// Apply the copied theme so we can see real-time changes
			if err := ce.themeManager.ApplyTheme(ce.themeName); err != nil {
				ce.app.QueueUpdateDraw(func() {
					ce.setStatus(fmt.Sprintf("Failed to apply copy: %v", err))
				})
			} else {
				ce.app.QueueUpdateDraw(func() {
					ce.setStatus(fmt.Sprintf("Created and applied copy: %s - Ready to edit", copyName))
					ce.appliedTheme = copyName // Update the applied theme tracking
				})
			}
		}
		ce.app.QueueUpdateDraw(func() {
			ce.app.ForceDraw()
		})
	}()
}

func (ce *ColorEditor) refreshThemeList() {
	currentItem := ce.themeList.GetCurrentItem()
	ce.themeList.Clear()
	
	// Get theme files directly
	themeFiles, err := ce.getThemeFiles()
	if err != nil {
		return
	}

	sort.Strings(themeFiles)

	for _, themeName := range themeFiles {
		displayName := themeName
		if themeName == ce.appliedTheme {
			displayName = fmt.Sprintf("★ %s", themeName)
		}
		ce.themeList.AddItem(displayName, "", 0, nil)
	}

	// Restore the selection
	if currentItem >= 0 && currentItem < ce.themeList.GetItemCount() {
		ce.themeList.SetCurrentItem(currentItem)
	}
}

func (ce *ColorEditor) setStatus(message string) {
	ce.statusBar.SetText(message)
}

func (ce *ColorEditor) updateColorStatus() {
	index := ce.colorPanel.GetCurrentItem()
	
	// Check if this item has a corresponding color key
	colorKey, exists := ce.listItemToColorKey[index]
	if !exists {
		ce.setStatus("Navigate to color items to edit | Tab: switch panels")
		return
	}
	
	colorValue := ce.colorValues[colorKey]
	displayName := strings.Replace(colorKey, ".", " ", -1)
	
	// Convert hex to RGB and HSL for display in status
	rgbDisplay := colorValue
	hslDisplay := ""
	if rgb, err := theme.HexToRGB(colorValue); err == nil {
		rgbDisplay = fmt.Sprintf("R:%d G:%d B:%d", rgb.R, rgb.G, rgb.B)
		hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)
		hslDisplay = fmt.Sprintf("H:%.0f° S:%.0f%% L:%.0f%%", hsl.H, hsl.S*100, hsl.L*100)
	}

	dirtyIndicator := ""
	if ce.isDirty {
		dirtyIndicator = " [UNSAVED] "
	}

	ce.setStatus(fmt.Sprintf("Selected: %s (%s | %s)%s | ←→: brightness | Shift+←→: hue | s: save", displayName, rgbDisplay, hslDisplay, dirtyIndicator))
}

func (ce *ColorEditor) adjustColorWithArrows(colorKey string, key tcell.Key) {
	// Create a copy of the theme if this is the first edit
	if !ce.isDirty {
		ce.createThemeCopy()
	}
	
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

	// Update preview
	ce.updatePreview()

	// Show that changes have been made (don't call updateColorStatus as it would overwrite this message)
	ce.setStatus(fmt.Sprintf("Modified %s (%s) | Press 's' to save | ←→: adjust RGB", displayName, rgbDisplay))
	
	// Apply changes in real-time
	go func() {
		// Save the current state to disk temporarily
		ce.updateThemeConfig()
		if err := ce.saveThemeToFile(); err == nil {
			// Apply the theme to see changes immediately
			if err := ce.themeManager.ApplyTheme(ce.themeName); err != nil {
				ce.app.QueueUpdateDraw(func() {
					ce.setStatus(fmt.Sprintf("Failed to apply changes: %v", err))
				})
			} else {
				// Update the applied theme tracking since we're applying the edited copy
				ce.appliedTheme = ce.themeName
			}
		}
		// Force a complete redraw to prevent text overlap
		ce.app.QueueUpdateDraw(func() {
			ce.app.ForceDraw()
		})
	}()
}

func (ce *ColorEditor) adjustColorHue(colorKey string, key tcell.Key) {
	// Create a copy of the theme if this is the first edit
	if !ce.isDirty {
		ce.createThemeCopy()
	}
	
	currentValue := ce.colorValues[colorKey]
	rgb, err := theme.HexToRGB(currentValue)
	if err != nil {
		return
	}

	// Convert to HSL for hue adjustment
	hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)
	
	// Adjust hue - 15 degree steps around the color wheel
	hueStep := 15.0
	if key == tcell.KeyRight {
		hsl.H += hueStep
		if hsl.H >= 360 {
			hsl.H -= 360
		}
	} else if key == tcell.KeyLeft {
		hsl.H -= hueStep
		if hsl.H < 0 {
			hsl.H += 360
		}
	}

	// Ensure minimum saturation for visible hue changes
	if hsl.S < 0.2 {
		hsl.S = 0.2
	}

	// Convert back to RGB
	newR, newG, newB := HSLToRGB(hsl)
	
	// Ensure values are in valid range
	newR = max(0, min(255, newR))
	newG = max(0, min(255, newG))
	newB = max(0, min(255, newB))

	// Create new RGB and convert to hex
	newRGB := &theme.RGB{R: newR, G: newG, B: newB}
	newHex := newRGB.ToHex()
	ce.colorValues[colorKey] = newHex
	ce.isDirty = true

	// Update just the current item in place
	currentIndex := ce.colorPanel.GetCurrentItem()

	// Update the current list item with the new color
	colorValue := newHex
	if !strings.HasPrefix(colorValue, "#") && len(colorValue) == 6 {
		colorValue = "#" + colorValue
	}

	// Convert to RGB for display
	rgbDisplay := fmt.Sprintf("R:%d G:%d B:%d", newR, newG, newB)
	displayName := strings.Replace(colorKey, ".", " ", -1)
	text := fmt.Sprintf("  [%s]██[-] %-20s %s", colorValue, displayName, rgbDisplay)

	// Update the current item
	ce.colorPanel.SetItemText(currentIndex, text, "")

	// Update preview
	ce.updatePreview()

	// Show hue adjustment info
	ce.setStatus(fmt.Sprintf("Hue adjusted %s (H:%.0f° S:%.0f%% L:%.0f%%) | Shift+←→: hue | ←→: brightness", 
		displayName, hsl.H, hsl.S*100, hsl.L*100))
	
	// Apply changes in real-time
	go func() {
		// Save the current state to disk temporarily
		ce.updateThemeConfig()
		if err := ce.saveThemeToFile(); err == nil {
			// Apply the theme to see changes immediately
			if err := ce.themeManager.ApplyTheme(ce.themeName); err != nil {
				ce.app.QueueUpdateDraw(func() {
					ce.setStatus(fmt.Sprintf("Failed to apply changes: %v", err))
				})
			} else {
				// Update the applied theme tracking since we're applying the edited copy
				ce.appliedTheme = ce.themeName
			}
		}
		// Force a complete redraw to prevent text overlap
		ce.app.QueueUpdateDraw(func() {
			ce.app.ForceDraw()
		})
	}()
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

// HSL represents a color in HSL color space
type HSL struct {
	H float64 // Hue (0-360)
	S float64 // Saturation (0-1)
	L float64 // Lightness (0-1)
}

// RGBToHSL converts RGB to HSL color space
func RGBToHSL(r, g, b int) HSL {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(math.Max(rf, gf), bf)
	min := math.Min(math.Min(rf, gf), bf)
	diff := max - min

	// Lightness
	l := (max + min) / 2.0

	var h, s float64

	if diff == 0 {
		h = 0 // Achromatic
		s = 0
	} else {
		// Saturation
		if l > 0.5 {
			s = diff / (2.0 - max - min)
		} else {
			s = diff / (max + min)
		}

		// Hue
		switch max {
		case rf:
			h = (gf-bf)/diff + (func() float64 {
				if gf < bf {
					return 6.0
				}
				return 0.0
			})()
		case gf:
			h = (bf-rf)/diff + 2.0
		case bf:
			h = (rf-gf)/diff + 4.0
		}
		h /= 6.0
	}

	return HSL{H: h * 360, S: s, L: l}
}

// HSLToRGB converts HSL to RGB color space
func HSLToRGB(hsl HSL) (int, int, int) {
	h := hsl.H / 360.0
	s := hsl.S
	l := hsl.L

	var r, g, b float64

	if s == 0 {
		r = l // Achromatic
		g = l
		b = l
	} else {
		hue2rgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hue2rgb(p, q, h+1.0/3.0)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3.0)
	}

	return int(r*255 + 0.5), int(g*255 + 0.5), int(b*255 + 0.5)
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
