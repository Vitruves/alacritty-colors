package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/internal/theme"
	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

// setupUI initializes the main UI layout
func (ce *ColorEditor) setupUI() {
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
	ce.statusBar.SetText(StatusBarDefault)
	ce.statusBar.SetTextColor(tcell.ColorYellow)

	// Layout
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

// loadThemes populates the theme list
func (ce *ColorEditor) loadThemes() {
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
			displayName = fmt.Sprintf("%s%s", CurrentThemeMarker, themeName)
			currentThemeIndex = i
		}
		if ce.isFavorite(themeName) {
			if themeName == ce.appliedTheme {
				displayName = fmt.Sprintf("♥ %s%s", CurrentThemeMarker, themeName)
			} else {
				displayName = fmt.Sprintf("♥ %s", themeName)
			}
		}
		ce.themeList.AddItem(displayName, "", 0, nil)
	}

	if len(themeFiles) > 0 {
		if currentThemeIndex >= 0 {
			ce.themeList.SetCurrentItem(currentThemeIndex)
			ce.onThemeSelected(currentThemeIndex, themeFiles[currentThemeIndex], "", 0)
		} else {
			ce.themeList.SetCurrentItem(0)
			ce.onThemeSelected(0, themeFiles[0], "", 0)
		}
	}
}

// getThemeFiles returns a list of theme names from the themes directory
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

// onThemeSelected handles theme selection from the list
func (ce *ColorEditor) onThemeSelected(index int, themeName string, _ string, _ rune) {
	themeName = strings.TrimPrefix(themeName, CurrentThemeMarker)
	themeName = strings.TrimPrefix(themeName, "♥ ")

	ce.themeName = themeName
	ce.loadTheme(themeName)
	ce.buildColorPanel()
	ce.updatePreview()
	if len(ce.colorKeys) > 0 {
		ce.colorPanel.SetCurrentItem(0)
		ce.updateColorStatus()
	}

	go func(theme string) {
		if err := ce.themeManager.ApplyTheme(theme); err != nil {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus(fmt.Sprintf("Failed to apply theme: %v", err))
			})
		} else {
			ce.app.QueueUpdateDraw(func() {
				ce.appliedTheme = theme
				ce.setStatus(fmt.Sprintf("Applied theme: %s | e: edit copy | d: delete | /: search | ?: help | q: quit", theme))
			})
		}
	}(themeName)
}

// buildColorPanel populates the color panel with the current theme's colors
func (ce *ColorEditor) buildColorPanel() {
	ce.colorPanel.Clear()
	ce.listItemToColorKey = make(map[int]string)

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

	for _, color := range BaseColorNames {
		sections["Normal"] = append(sections["Normal"], "normal."+color)
		sections["Bright"] = append(sections["Bright"], "bright."+color)
		sections["Dim"] = append(sections["Dim"], "dim."+color)
	}

	itemIndex := 0
	for _, sectionName := range ColorSectionOrder {
		keys := sections[sectionName]
		if len(keys) == 0 {
			continue
		}

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

		ce.colorPanel.AddItem(fmt.Sprintf("[cyan::b]%s[-]", sectionName), "", 0, nil)
		itemIndex++

		for _, key := range keys {
			if value, exists := ce.colorValues[key]; exists {
				colorValue := value
				if !strings.HasPrefix(colorValue, "#") && len(colorValue) == 6 {
					colorValue = "#" + colorValue
				}

				rgbDisplay := ""
				if rgb, err := theme.HexToRGB(colorValue); err == nil {
					rgbDisplay = fmt.Sprintf("R:%d G:%d B:%d", rgb.R, rgb.G, rgb.B)
				} else {
					rgbDisplay = colorValue
				}

				displayName := strings.Replace(key, ".", " ", -1)
				text := fmt.Sprintf("  [%s]██[-] %-20s %s", colorValue, displayName, rgbDisplay)

				ce.listItemToColorKey[itemIndex] = key

				ce.colorPanel.AddItem(text, "", 0, nil)
				itemIndex++
			}
		}
	}
}

// refreshThemeList refreshes the theme list while preserving selection
func (ce *ColorEditor) refreshThemeList() {
	currentItem := ce.themeList.GetCurrentItem()
	ce.themeList.Clear()

	themeFiles, err := ce.getThemeFiles()
	if err != nil {
		return
	}

	sort.Strings(themeFiles)

	for _, themeName := range themeFiles {
		displayName := themeName
		if themeName == ce.appliedTheme {
			displayName = fmt.Sprintf("%s%s", CurrentThemeMarker, themeName)
		}
		if ce.isFavorite(themeName) {
			if themeName == ce.appliedTheme {
				displayName = fmt.Sprintf("♥ %s%s", CurrentThemeMarker, themeName)
			} else {
				displayName = fmt.Sprintf("♥ %s", themeName)
			}
		}
		ce.themeList.AddItem(displayName, "", 0, nil)
	}

	if currentItem >= 0 && currentItem < ce.themeList.GetItemCount() {
		ce.themeList.SetCurrentItem(currentItem)
	}
}

// setStatus updates the status bar text
func (ce *ColorEditor) setStatus(message string) {
	ce.statusBar.SetText(message)
}

// applyUserThemeToTUI applies the current Alacritty theme colors to the TUI
func (ce *ColorEditor) applyUserThemeToTUI() {
	parser := alacritty.NewParser()
	currentConfig, err := parser.ParseFile(ce.config.ConfigFile)
	if err != nil {
		currentThemePath := ce.config.GetThemePath("current")
		currentConfig, err = parser.ParseFile(currentThemePath)
		if err != nil {
			return
		}
	}

	ce.applyAlacrittyColors(currentConfig)
}

// applyAlacrittyColors applies Alacritty config colors to TUI styles
func (ce *ColorEditor) applyAlacrittyColors(config *alacritty.Config) {
	bgColor := ce.hexToTcellColor(config.Colors.Primary.Background)
	fgColor := ce.hexToTcellColor(config.Colors.Primary.Foreground)

	greenColor := ce.hexToTcellColor(config.Colors.Normal["green"])
	yellowColor := ce.hexToTcellColor(config.Colors.Normal["yellow"])
	blueColor := ce.hexToTcellColor(config.Colors.Normal["blue"])
	cyanColor := ce.hexToTcellColor(config.Colors.Normal["cyan"])

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

// hexToTcellColor converts a hex color string to tcell.Color
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

// resetTUITheme resets TUI to default colors for modals
func (ce *ColorEditor) resetTUITheme() {
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
