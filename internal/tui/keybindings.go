package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/vitruves/alacritty-colors/internal/theme"
)

// handleGlobalKeys handles application-wide key events
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
		case 'c', 'C':
			ce.cycleColorMode()
			return nil
		case 'e', 'E':
			ce.createThemeCopy()
			return nil
		case 'd', 'D':
			ce.deleteCurrentTheme()
			return nil
		case 'p', 'P':
			ce.showParametersPanel()
			return nil
		case 'f', 'F':
			ce.showFontPanel()
			return nil
		case 'n', 'N':
			ce.showThemeCreator()
			return nil
		case 'g', 'G':
			ce.generateAndApplyRandomTheme(true)
			return nil
		case '/':
			ce.showSearchDialog()
			return nil
		case '?':
			ce.showHelpOverlay()
			return nil
		case '*':
			ce.toggleFavorite()
			return nil
		}
	}
	return event
}

// handleThemeListKeys handles key events for the theme list panel
func (ce *ColorEditor) handleThemeListKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		ce.app.SetFocus(ce.colorPanel)
		ce.colorPanel.SetBorderColor(tcell.ColorYellow)
		ce.themeList.SetBorderColor(tcell.ColorDefault)
		ce.setStatus(StatusBarColorPanel)
		return nil
	case tcell.KeyEnter:
		index := ce.themeList.GetCurrentItem()
		if index >= 0 && index < ce.themeList.GetItemCount() {
			themeName, _ := ce.themeList.GetItemText(index)
			themeName = strings.TrimPrefix(themeName, CurrentThemeMarker)
			themeName = strings.TrimPrefix(themeName, "♥ ")
			ce.onThemeSelected(index, themeName, "", 0)
			ce.app.SetFocus(ce.colorPanel)
			ce.colorPanel.SetBorderColor(tcell.ColorYellow)
			ce.themeList.SetBorderColor(tcell.ColorDefault)
		}
		return nil
	case tcell.KeyUp, tcell.KeyDown:
		result := event
		go func() {
			time.Sleep(KeyDebounceDelay)
			ce.app.QueueUpdateDraw(func() {
				index := ce.themeList.GetCurrentItem()
				if index >= 0 && index < ce.themeList.GetItemCount() {
					themeName, _ := ce.themeList.GetItemText(index)
					themeName = strings.TrimPrefix(themeName, CurrentThemeMarker)
					themeName = strings.TrimPrefix(themeName, "♥ ")
					ce.onThemeSelected(index, themeName, "", 0)
				}
			})
		}()
		return result
	case tcell.KeyRune:
		// Quick jump with letter keys
		if event.Rune() >= 'a' && event.Rune() <= 'z' {
			ce.jumpToThemeStartingWith(string(event.Rune()))
			return nil
		}
	}
	return event
}

// handleColorPanelKeys handles key events for the color panel
func (ce *ColorEditor) handleColorPanelKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		ce.app.SetFocus(ce.themeList)
		ce.themeList.SetBorderColor(tcell.ColorYellow)
		ce.colorPanel.SetBorderColor(tcell.ColorDefault)
		ce.setStatus(StatusBarThemeList)
		return nil
	case tcell.KeyEnter:
		index := ce.colorPanel.GetCurrentItem()
		ce.onColorSelected(index, "", "", 0)
		return nil
	case tcell.KeyUp, tcell.KeyDown:
		result := event
		go func() {
			time.Sleep(KeyDebounceDelay)
			ce.app.QueueUpdateDraw(func() {
				ce.updateColorStatus()
			})
		}()
		return result
	case tcell.KeyLeft, tcell.KeyRight:
		index := ce.colorPanel.GetCurrentItem()
		colorKey, exists := ce.listItemToColorKey[index]

		if !exists {
			return event
		}

		if event.Modifiers()&tcell.ModShift != 0 {
			ce.adjustColorHue(colorKey, event.Key())
		} else {
			ce.adjustColorWithArrows(colorKey, event.Key())
		}
		return nil
	}
	return event
}

// adjustColorWithArrows adjusts color brightness with arrow keys
func (ce *ColorEditor) adjustColorWithArrows(colorKey string, key tcell.Key) {
	if colorKey == "" {
		return
	}

	ce.isDirty = true

	currentValue, exists := ce.colorValues[colorKey]
	if !exists || currentValue == "" {
		return
	}

	rgb, err := theme.HexToRGB(currentValue)
	if err != nil {
		return
	}

	newRGB := AdjustBrightness(rgb, key == tcell.KeyRight)
	newHex := newRGB.ToHex()
	ce.colorValues[colorKey] = newHex

	ce.updateColorItemDisplay(colorKey, &newRGB)
	ce.updatePreview()

	displayName := strings.Replace(colorKey, ".", " ", -1)
	rgbDisplay := fmt.Sprintf("R:%d G:%d B:%d", newRGB.R, newRGB.G, newRGB.B)
	ce.setStatus(fmt.Sprintf("Modified %s (%s) | Press 's' to save | ←→: adjust RGB", displayName, rgbDisplay))

	ce.applyColorChangeRealtime()
}

// adjustColorHue adjusts color hue with shift+arrow keys
func (ce *ColorEditor) adjustColorHue(colorKey string, key tcell.Key) {
	if colorKey == "" {
		return
	}

	ce.isDirty = true

	currentValue, exists := ce.colorValues[colorKey]
	if !exists || currentValue == "" {
		return
	}

	rgb, err := theme.HexToRGB(currentValue)
	if err != nil {
		return
	}

	newRGB := AdjustHue(rgb, key == tcell.KeyRight)
	newHex := newRGB.ToHex()
	ce.colorValues[colorKey] = newHex

	ce.updateColorItemDisplay(colorKey, &newRGB)
	ce.updatePreview()

	hsl := RGBToHSL(newRGB.R, newRGB.G, newRGB.B)
	displayName := strings.Replace(colorKey, ".", " ", -1)
	ce.setStatus(fmt.Sprintf("Hue adjusted %s (H:%.0f° S:%.0f%% L:%.0f%%) | Shift+←→: hue",
		displayName, hsl.H, hsl.S*100, hsl.L*100))

	ce.applyColorChangeRealtime()
}

// updateColorItemDisplay updates a single color item in the panel
func (ce *ColorEditor) updateColorItemDisplay(colorKey string, rgb *theme.RGB) {
	currentIndex := ce.colorPanel.GetCurrentItem()

	colorValue := rgb.ToHex()
	if !strings.HasPrefix(colorValue, "#") && len(colorValue) == 6 {
		colorValue = "#" + colorValue
	}

	rgbDisplay := fmt.Sprintf("R:%d G:%d B:%d", rgb.R, rgb.G, rgb.B)
	displayName := strings.Replace(colorKey, ".", " ", -1)
	text := fmt.Sprintf("  [%s]██[-] %-20s %s", colorValue, displayName, rgbDisplay)

	ce.colorPanel.SetItemText(currentIndex, text, "")
}

// applyColorChangeRealtime applies color changes in real-time
func (ce *ColorEditor) applyColorChangeRealtime() {
	go func() {
		themeName := ce.themeName
		ce.updateThemeConfig()
		if err := ce.saveThemeToFile(); err == nil {
			if err := ce.themeManager.ApplyTheme(themeName); err != nil {
				ce.app.QueueUpdateDraw(func() {
					ce.setStatus(fmt.Sprintf("Failed to apply changes to %s: %v", themeName, err))
				})
			} else {
				ce.appliedTheme = themeName
			}
		}
		ce.app.QueueUpdateDraw(func() {})
	}()
}

// jumpToThemeStartingWith jumps to the first theme starting with a letter
func (ce *ColorEditor) jumpToThemeStartingWith(letter string) {
	letter = strings.ToLower(letter)
	count := ce.themeList.GetItemCount()

	for i := 0; i < count; i++ {
		name, _ := ce.themeList.GetItemText(i)
		name = strings.TrimPrefix(name, CurrentThemeMarker)
		name = strings.TrimPrefix(name, "♥ ")
		if strings.HasPrefix(strings.ToLower(name), letter) {
			ce.themeList.SetCurrentItem(i)
			ce.onThemeSelected(i, name, "", 0)
			break
		}
	}
}

// updateColorStatus updates the status bar with current color info
func (ce *ColorEditor) updateColorStatus() {
	index := ce.colorPanel.GetCurrentItem()

	colorKey, exists := ce.listItemToColorKey[index]
	if !exists {
		ce.setStatus("Navigate to color items to edit | Tab: switch panels")
		return
	}

	colorValue := ce.colorValues[colorKey]
	displayName := strings.Replace(colorKey, ".", " ", -1)

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

	ce.setStatus(fmt.Sprintf("Selected: %s (%s | %s)%s | ←→: brightness | Shift+←→: hue | s: save",
		displayName, rgbDisplay, hslDisplay, dirtyIndicator))
}

// onColorSelected handles color selection in the panel
func (ce *ColorEditor) onColorSelected(index int, text string, _ string, _ rune) {
	ce.updateColorStatus()
}
