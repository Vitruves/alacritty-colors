package tui

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/internal/theme"
)

// confirmQuit shows the quit confirmation dialog
func (ce *ColorEditor) confirmQuit() {
	ce.resetTUITheme()
	ce.app.SetInputCapture(nil)

	modal := tview.NewModal()
	modal.SetText("You have unsaved changes. Are you sure you want to quit?")
	modal.AddButtons([]string{"Save & Quit", "Quit", "Cancel"})

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
			ce.applyUserThemeToTUI()
			ce.setupUI()
			ce.buildColorPanel()
			ce.updatePreview()
		}
	})

	ce.app.SetRoot(modal, true)
}

// showParametersPanel shows the parameters/utilities modal
func (ce *ColorEditor) showParametersPanel() {
	ce.resetTUITheme()
	ce.app.SetInputCapture(nil)

	modal := tview.NewModal()
	modal.SetText("Parameters & Utilities\n\nSelect an option:")
	modal.AddButtons([]string{"Clean & Redownload Themes", "Backup Current Config", "Reset to Defaults", "Cancel"})

	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorBlue)
	modal.SetButtonTextColor(tcell.ColorWhite)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		ce.restoreMainUI()

		switch buttonIndex {
		case 0:
			ce.cleanAndRedownloadThemes()
		case 1:
			ce.backupCurrentConfig()
		case 2:
			ce.resetToDefaults()
		}
	})

	ce.app.SetRoot(modal, true)
}

// showFontPanel shows the font settings modal
func (ce *ColorEditor) showFontPanel() {
	ce.resetTUITheme()
	ce.app.SetInputCapture(nil)

	modal := tview.NewModal()
	modal.SetText("Font Settings\n\nSelect an option:")
	modal.AddButtons([]string{"Change Font Family", "Adjust Font Size", "Font Weight", "Cancel"})

	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorGreen)
	modal.SetButtonTextColor(tcell.ColorWhite)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		ce.restoreMainUI()

		switch buttonIndex {
		case 0, 1, 2:
			ce.showFontTUI()
		}
	})

	ce.app.SetRoot(modal, true)
}

// showHelpOverlay shows the keyboard shortcuts help
func (ce *ColorEditor) showHelpOverlay() {
	ce.resetTUITheme()

	// Disable global key handler while help is open
	ce.app.SetInputCapture(nil)

	helpText := `[yellow::b]Keyboard Shortcuts[-]

[cyan::b]Navigation[-]
  Tab          Switch between panels
  ↑/↓          Navigate items
  a-z          Jump to theme starting with letter
  /            Search themes

[cyan::b]Theme Operations[-]
  Enter/a      Apply selected theme
  e            Create editable copy
  d            Delete edited theme
  *            Toggle favorite
  r            Reset to original

[cyan::b]Theme Creation[-]
  n            Open theme creator
  g            Generate random harmonious theme

[cyan::b]Color Editing[-]
  ←/→          Adjust brightness
  Shift+←/→    Adjust hue
  s            Save changes
  c            Cycle preview color mode

[cyan::b]Settings[-]
  f            Font settings
  p            Parameters & utilities

[cyan::b]General[-]
  ?            Show this help
  q            Quit
  Ctrl+C       Force quit

[white]Press any key to close this help[-]`

	textView := tview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetText(helpText)
	textView.SetBorder(true)
	textView.SetTitle(" Help ")
	textView.SetTitleAlign(tview.AlignCenter)
	textView.SetBackgroundColor(tcell.ColorBlack)

	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		ce.restoreMainUI()
		return nil
	})

	// Center the help in a flex container
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(nil, 0, 1, false)
	flex.AddItem(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(textView, 60, 0, true).
		AddItem(nil, 0, 1, false), 25, 0, true)
	flex.AddItem(nil, 0, 1, false)

	ce.app.SetRoot(flex, true)
}

// showSearchDialog shows the theme search dialog
func (ce *ColorEditor) showSearchDialog() {
	ce.resetTUITheme()

	// Disable global key handler while search is open
	ce.app.SetInputCapture(nil)

	inputField := tview.NewInputField()
	inputField.SetLabel("Search: ")
	inputField.SetFieldWidth(40)
	inputField.SetFieldBackgroundColor(tcell.ColorBlack)
	inputField.SetFieldTextColor(tcell.ColorWhite)
	inputField.SetLabelColor(tcell.ColorYellow)
	inputField.SetBorder(true)
	inputField.SetTitle(" Search Themes ")

	// Results list
	resultsList := tview.NewList()
	resultsList.ShowSecondaryText(false)
	resultsList.SetBorder(true)
	resultsList.SetTitle(" Results ")

	// Get all themes for filtering
	themes, _ := ce.getThemeFiles()

	// Update results as user types
	inputField.SetChangedFunc(func(text string) {
		resultsList.Clear()
		if text == "" {
			return
		}

		searchLower := strings.ToLower(text)
		matchCount := 0
		for _, themeName := range themes {
			if strings.Contains(strings.ToLower(themeName), searchLower) {
				displayName := themeName
				if themeName == ce.appliedTheme {
					displayName = CurrentThemeMarker + themeName
				}
				if ce.isFavorite(themeName) {
					displayName = "♥ " + displayName
				}
				resultsList.AddItem(displayName, "", 0, nil)
				matchCount++
				if matchCount >= 20 { // Limit results
					break
				}
			}
		}
	})

	// Handle Enter to select
	inputField.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			if resultsList.GetItemCount() > 0 {
				index := resultsList.GetCurrentItem()
				if index < 0 {
					index = 0
				}
				themeName, _ := resultsList.GetItemText(index)
				themeName = strings.TrimPrefix(themeName, CurrentThemeMarker)
				themeName = strings.TrimPrefix(themeName, "♥ ")
				ce.restoreMainUI()
				ce.selectThemeByName(themeName)
			} else {
				ce.restoreMainUI()
			}
		case tcell.KeyEscape:
			ce.restoreMainUI()
		}
	})

	// Handle navigation in results
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown:
			if resultsList.GetItemCount() > 0 {
				ce.app.SetFocus(resultsList)
				resultsList.SetCurrentItem(0)
			}
			return nil
		case tcell.KeyEscape:
			ce.restoreMainUI()
			return nil
		}
		return event
	})

	resultsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if resultsList.GetCurrentItem() == 0 {
				ce.app.SetFocus(inputField)
				return nil
			}
		case tcell.KeyEnter:
			if resultsList.GetItemCount() > 0 {
				index := resultsList.GetCurrentItem()
				themeName, _ := resultsList.GetItemText(index)
				themeName = strings.TrimPrefix(themeName, CurrentThemeMarker)
				themeName = strings.TrimPrefix(themeName, "♥ ")
				ce.restoreMainUI()
				ce.selectThemeByName(themeName)
				return nil
			}
		case tcell.KeyEscape:
			ce.restoreMainUI()
			return nil
		}
		return event
	})

	// Layout
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(nil, 0, 1, false)
	flex.AddItem(tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(inputField, 3, 0, true).
			AddItem(resultsList, 15, 0, false), 50, 0, true).
		AddItem(nil, 0, 1, false), 20, 0, true)
	flex.AddItem(nil, 0, 1, false)

	ce.app.SetRoot(flex, true)
	ce.app.SetFocus(inputField)
}

// showDeleteConfirmation shows the delete confirmation dialog
func (ce *ColorEditor) showDeleteConfirmation() {
	ce.resetTUITheme()
	ce.app.SetInputCapture(nil)

	modal := tview.NewModal()
	modal.SetText(fmt.Sprintf("Are you sure you want to delete theme '%s'?", ce.themeName))
	modal.AddButtons([]string{"Delete", "Cancel"})

	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorRed)
	modal.SetButtonTextColor(tcell.ColorWhite)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		ce.restoreMainUI()

		if buttonIndex == 0 {
			ce.performThemeDelete()
		}
	})

	ce.app.SetRoot(modal, true)
}

// restoreMainUI restores the main three-panel layout
func (ce *ColorEditor) restoreMainUI() {
	rootFlex := tview.NewFlex()
	rootFlex.SetDirection(tview.FlexRow)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(ce.themeList, 0, 1, false)
	mainFlex.AddItem(ce.colorPanel, 0, 2, false)
	mainFlex.AddItem(ce.previewPanel, 0, 1, false)

	rootFlex.AddItem(mainFlex, 0, 1, true)
	rootFlex.AddItem(ce.statusBar, 1, 0, false)

	ce.app.SetRoot(rootFlex, true)
	ce.applyUserThemeToTUI()

	ce.app.SetInputCapture(ce.handleGlobalKeys)

	ce.app.SetFocus(ce.themeList)
	ce.themeList.SetBorderColor(tcell.ColorYellow)
	ce.colorPanel.SetBorderColor(tcell.ColorDefault)
}

// selectThemeByName finds and selects a theme by name
func (ce *ColorEditor) selectThemeByName(themeName string) {
	count := ce.themeList.GetItemCount()
	for i := 0; i < count; i++ {
		name, _ := ce.themeList.GetItemText(i)
		name = strings.TrimPrefix(name, CurrentThemeMarker)
		name = strings.TrimPrefix(name, "♥ ")
		if name == themeName {
			ce.themeList.SetCurrentItem(i)
			ce.onThemeSelected(i, themeName, "", 0)
			break
		}
	}
}

// showThemeCreator shows the theme creator/generator modal
func (ce *ColorEditor) showThemeCreator() {
	ce.resetTUITheme()

	// Disable global key handler while creator is open
	ce.app.SetInputCapture(nil)

	// Style selection
	styleList := tview.NewList()
	styleList.ShowSecondaryText(true)
	styleList.SetBorder(true)
	styleList.SetTitle(" Theme Style ")
	styleList.AddItem("Dark Theme", "Light text on dark background", 0, nil)
	styleList.AddItem("Light Theme", "Dark text on light background", 0, nil)

	// Generation method selection
	methodList := tview.NewList()
	methodList.ShowSecondaryText(true)
	methodList.SetBorder(true)
	methodList.SetTitle(" Generation Method ")
	methodList.AddItem("Full Random", "Completely random colors", 0, nil)
	methodList.AddItem("Complementary", "Opposite colors (high contrast)", 0, nil)
	methodList.AddItem("Analogous", "Adjacent colors (cohesive)", 0, nil)
	methodList.AddItem("Triadic", "Three balanced colors", 0, nil)
	methodList.AddItem("Split-Complementary", "Vibrant with less tension", 0, nil)
	methodList.AddItem("Tetradic", "Four rich colors", 0, nil)
	methodList.AddItem("Monochromatic", "Single hue variations", 0, nil)

	// Preview panel
	previewPanel := tview.NewTextView()
	previewPanel.SetDynamicColors(true)
	previewPanel.SetBorder(true)
	previewPanel.SetTitle(" Preview ")

	// Status bar
	statusBar := tview.NewTextView()
	statusBar.SetText("Tab: switch panels | ↑↓: navigate | Enter: generate | g: generate new | s: save & apply | q: cancel")
	statusBar.SetTextColor(tcell.ColorYellow)

	var currentGenerated *GeneratedTheme

	// Update preview with generated theme
	updatePreviewPanel := func() {
		if currentGenerated == nil {
			previewPanel.SetText("[yellow]Select options and press Enter to generate[-]")
			return
		}

		preview := fmt.Sprintf(`[yellow::b]Generated Theme: %s[-]

[white::b]Background:[-] [%s]████[-] %s
[white::b]Foreground:[-] [%s]████[-] %s

[white::b]Normal Colors:[-]
`, currentGenerated.Name,
			currentGenerated.Background, currentGenerated.Background,
			currentGenerated.Foreground, currentGenerated.Foreground)

		for _, name := range BaseColorNames {
			if color, ok := currentGenerated.Normal[name]; ok {
				preview += fmt.Sprintf("[%s]██[-] %-8s ", color, name)
			}
		}

		preview += "\n\n[white::b]Bright Colors:[-]\n"
		for _, name := range BaseColorNames {
			if color, ok := currentGenerated.Bright[name]; ok {
				preview += fmt.Sprintf("[%s]██[-] %-8s ", color, name)
			}
		}

		preview += "\n\n[white::b]Sample:[-]\n"
		preview += fmt.Sprintf("[%s]user@host[-]:[%s]~/code[-]$ [%s]ls -la[-]\n",
			currentGenerated.Normal["green"],
			currentGenerated.Normal["blue"],
			currentGenerated.Normal["yellow"])
		preview += fmt.Sprintf("[%s]func[-] [%s]main[-]() { [%s]// comment[-] }\n",
			currentGenerated.Normal["blue"],
			currentGenerated.Normal["yellow"],
			currentGenerated.Normal["magenta"])

		previewPanel.SetText(preview)
	}

	// Generate theme based on current selections
	generateTheme := func() {
		styleIdx := styleList.GetCurrentItem()
		methodIdx := methodList.GetCurrentItem()

		style := ThemeStyleDark
		if styleIdx == 1 {
			style = ThemeStyleLight
		}

		if methodIdx == 0 {
			// Full random
			currentGenerated = GenerateRandomTheme(style)
		} else {
			// Harmonious generation
			harmony := HarmonyType(methodIdx - 1)
			currentGenerated = GenerateHarmoniousTheme(style, harmony)
		}

		updatePreviewPanel()
		statusBar.SetText(fmt.Sprintf("Generated: %s | g: regenerate | s: save & apply | q: cancel", currentGenerated.Name))
	}

	// Key handlers
	styleList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ce.app.SetFocus(methodList)
			methodList.SetBorderColor(tcell.ColorYellow)
			styleList.SetBorderColor(tcell.ColorDefault)
			return nil
		case tcell.KeyEnter:
			generateTheme()
			return nil
		case tcell.KeyEscape:
			ce.restoreMainUI()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g', 'G':
				generateTheme()
				return nil
			case 's', 'S':
				if currentGenerated != nil {
					ce.restoreMainUI()
					ce.ApplyGeneratedTheme(currentGenerated)
					ce.saveTheme()
					ce.refreshThemeList()
					ce.setStatus(fmt.Sprintf("Created and applied: %s", currentGenerated.Name))
				}
				return nil
			case 'q', 'Q':
				ce.restoreMainUI()
				return nil
			}
		}
		return event
	})

	methodList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ce.app.SetFocus(styleList)
			styleList.SetBorderColor(tcell.ColorYellow)
			methodList.SetBorderColor(tcell.ColorDefault)
			return nil
		case tcell.KeyEnter:
			generateTheme()
			return nil
		case tcell.KeyEscape:
			ce.restoreMainUI()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g', 'G':
				generateTheme()
				return nil
			case 's', 'S':
				if currentGenerated != nil {
					ce.restoreMainUI()
					ce.ApplyGeneratedTheme(currentGenerated)
					ce.saveTheme()
					ce.refreshThemeList()
					ce.setStatus(fmt.Sprintf("Created and applied: %s", currentGenerated.Name))
				}
				return nil
			case 'q', 'Q':
				ce.restoreMainUI()
				return nil
			}
		}
		return event
	})

	// Initial preview
	updatePreviewPanel()

	// Layout
	leftFlex := tview.NewFlex()
	leftFlex.SetDirection(tview.FlexRow)
	leftFlex.AddItem(styleList, 6, 0, true)
	leftFlex.AddItem(methodList, 0, 1, false)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(leftFlex, 0, 1, true)
	mainFlex.AddItem(previewPanel, 0, 2, false)

	rootFlex := tview.NewFlex()
	rootFlex.SetDirection(tview.FlexRow)
	rootFlex.AddItem(mainFlex, 0, 1, true)
	rootFlex.AddItem(statusBar, 1, 0, false)

	ce.app.SetRoot(rootFlex, true)
	ce.app.SetFocus(styleList)
	styleList.SetBorderColor(tcell.ColorYellow)
}

// generateAndApplyRandomTheme generates and applies a random theme immediately
func (ce *ColorEditor) generateAndApplyRandomTheme(harmonious bool) {
	var gen *GeneratedTheme

	// Detect current theme style (dark or light)
	style := ThemeStyleDark
	if bg, exists := ce.colorValues["primary.background"]; exists {
		if rgb, err := theme.HexToRGB(bg); err == nil {
			hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)
			if hsl.L > 0.5 {
				style = ThemeStyleLight
			}
		}
	}

	if harmonious {
		// Pick a random harmony type
		harmony := HarmonyType(rand.Intn(6))
		gen = GenerateHarmoniousTheme(style, harmony)
		ce.setStatus(fmt.Sprintf("Generated harmonious theme (%s) - Press 's' to save", HarmonyNames[harmony]))
	} else {
		gen = GenerateRandomTheme(style)
		ce.setStatus("Generated random theme - Press 's' to save")
	}

	ce.ApplyGeneratedTheme(gen)

	// Apply to Alacritty immediately
	go func() {
		ce.updateThemeConfig()
		if err := ce.saveThemeToFile(); err == nil {
			ce.themeManager.ApplyTheme(ce.themeName)
		}
	}()
}
