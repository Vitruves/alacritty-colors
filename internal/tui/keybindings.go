package tui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/vitruves/alacritty-colors/internal/theme"
)

// handleGlobalKeys handles application-wide key events
func (ce *ColorEditor) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyCtrlC {
		ce.app.Stop()
		return nil
	}

	// Dialogs and the filter line own every other key while they are open.
	if ce.overlayOpen() || ce.filterFocus {
		return event
	}

	switch event.Key() {
	case tcell.KeyTab:
		ce.setFocus(ce.focus + 1)
		return nil
	case tcell.KeyBacktab:
		ce.setFocus(ce.focus - 1)
		return nil
	case tcell.KeyEscape:
		if ce.filter != "" || ce.favOnly {
			ce.clearFilter()
			return nil
		}
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q', 'Q':
			ce.quitAndKeep()
			return nil
		case 's':
			ce.saveTheme()
			return nil
		case 'S':
			ce.showSaveAsDialog()
			return nil
		case 'r', 'R':
			ce.resetTheme()
			return nil
		case 'u', 'U':
			ce.undoCurrentColor()
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
		case 'f':
			ce.showFontPanel()
			return nil
		case 'F':
			ce.toggleFavoritesOnly()
			return nil
		case 'n', 'N':
			ce.showThemeCreator()
			return nil
		case 'g', 'G':
			ce.generateAndApplyRandomTheme(true)
			return nil
		case '/':
			ce.openFilter()
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

// handleThemeListKeys owns navigation in the theme list. Movement is handled
// here rather than delegated so that selection, preview and the debounced apply
// stay in lockstep.
func (ce *ColorEditor) handleThemeListKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyUp:
		ce.moveThemeSelection(-1)
		return nil
	case tcell.KeyDown:
		ce.moveThemeSelection(1)
		return nil
	case tcell.KeyPgUp:
		ce.moveThemeSelection(-10)
		return nil
	case tcell.KeyPgDn:
		ce.moveThemeSelection(10)
		return nil
	case tcell.KeyHome:
		ce.setThemeIndex(0)
		return nil
	case tcell.KeyEnd:
		ce.setThemeIndex(len(ce.visible) - 1)
		return nil
	case tcell.KeyEnter:
		if name := ce.currentVisibleName(); name != "" {
			ce.applyThemeNow(name)
			ce.setFocus(FocusColorPanel)
		}
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j':
			ce.moveThemeSelection(1)
			return nil
		case 'k':
			ce.moveThemeSelection(-1)
			return nil
		}
	}
	return event
}

// moveThemeSelection moves the cursor by delta rows and loads what it lands on.
func (ce *ColorEditor) moveThemeSelection(delta int) {
	if len(ce.visible) == 0 {
		return
	}
	idx := ce.themeList.GetCurrentItem() + delta
	if idx < 0 {
		idx = 0
	}
	if idx >= len(ce.visible) {
		idx = len(ce.visible) - 1
	}
	ce.setThemeIndex(idx)
}

func (ce *ColorEditor) setThemeIndex(idx int) {
	if idx < 0 || idx >= len(ce.visible) {
		return
	}
	if ce.visible[idx] == ce.themeName {
		ce.themeList.SetCurrentItem(idx)
		return
	}

	// Loading another theme drops the working colours. Say so once instead of
	// discarding someone's edits without a word; the next press goes through.
	if ce.isDirty && !ce.discardArmed {
		ce.discardArmed = true
		ce.warn("%s has unsaved edits — a keeps them, r discards, ↑↓ again leaves them behind", ce.themeName)
		return
	}
	ce.discardArmed = false

	ce.themeList.SetCurrentItem(idx)
	ce.selectTheme(ce.visible[idx], true)
}

// handleColorPanelKeys handles key events for the color panel
func (ce *ColorEditor) handleColorPanelKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyUp:
		ce.moveColorSelection(-1)
		return nil
	case tcell.KeyDown:
		ce.moveColorSelection(1)
		return nil
	case tcell.KeyHome:
		if i := ce.firstColorIndex(); i >= 0 {
			ce.colorPanel.SetCurrentItem(i)
			ce.renderStatus()
		}
		return nil
	case tcell.KeyEnd:
		ce.colorPanel.SetCurrentItem(ce.colorPanel.GetItemCount() - 1)
		ce.moveColorSelection(0)
		ce.renderStatus()
		return nil
	case tcell.KeyEnter:
		ce.showHexInput()
		return nil
	case tcell.KeyLeft, tcell.KeyRight:
		up := event.Key() == tcell.KeyRight
		if event.Modifiers()&tcell.ModShift != 0 {
			ce.editColor(hueEdit, up)
		} else {
			ce.editColor(brightnessEdit, up)
		}
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j':
			ce.moveColorSelection(1)
			return nil
		case 'k':
			ce.moveColorSelection(-1)
			return nil
		case '-', '_':
			ce.editColor(saturationEdit, false)
			return nil
		case '+', '=':
			ce.editColor(saturationEdit, true)
			return nil
		case '[':
			ce.editColor(lightnessEdit, false)
			return nil
		case ']':
			ce.editColor(lightnessEdit, true)
			return nil
		case '#':
			ce.showHexInput()
			return nil
		}
	}
	return event
}

// handlePreviewKeys keeps the preview scrollable without swallowing globals.
func (ce *ColorEditor) handlePreviewKeys(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case 'j':
			row, col := ce.previewPanel.GetScrollOffset()
			ce.previewPanel.ScrollTo(row+1, col)
			return nil
		case 'k':
			row, col := ce.previewPanel.GetScrollOffset()
			ce.previewPanel.ScrollTo(row-1, col)
			return nil
		}
	}
	return event
}

// --- colour editing -------------------------------------------------------

type editKind int

const (
	brightnessEdit editKind = iota
	hueEdit
	saturationEdit
	lightnessEdit
)

// editColor applies one adjustment to the selected colour. Nothing is written
// to the theme file here: only the live preview (themes/current.toml) is
// refreshed, so the source theme is untouched until an explicit save.
func (ce *ColorEditor) editColor(kind editKind, increase bool) {
	key := ce.selectedColorKey()
	if key == "" {
		ce.warn("Move to a colour row first")
		return
	}

	current, ok := ce.colorValues[key]
	if !ok || current == "" {
		// An unset slot materialises on first touch rather than blocking.
		ce.setColorValue(key, ce.derivedDefault(key))
		ce.info("Added %s — adjust it or press Enter for an exact value", strings.ReplaceAll(key, ".", " "))
		return
	}
	rgb, err := theme.HexToRGB(normalizeHex(current, "#000000"))
	if err != nil {
		return
	}

	var next theme.RGB
	switch kind {
	case brightnessEdit:
		next = AdjustBrightness(rgb, increase)
	case hueEdit:
		next = AdjustHue(rgb, increase)
	case saturationEdit:
		next = AdjustSaturation(rgb, increase)
	case lightnessEdit:
		next = AdjustLightness(rgb, increase)
	}

	ce.setColorValue(key, next.ToHex())
}

// setColorValue is the single funnel for every colour change: hex entry,
// arrow-key nudges and generated themes all land here.
func (ce *ColorEditor) setColorValue(key, hex string) {
	hex = normalizeHex(hex, ce.colorValues[key])
	if hex == ce.colorValues[key] {
		return
	}

	ce.colorValues[key] = hex
	ce.isDirty = true
	ce.discardArmed = false // a fresh edit re-arms the leave-behind warning

	if idx, ok := ce.colorKeyToListItem[key]; ok {
		ce.colorPanel.SetItemText(idx, ce.formatColorItem(key), "")
	}

	// The background changes every contrast reading on screen.
	if key == "primary.background" {
		ce.refreshContrastColumn()
	}

	ce.colorPanel.SetTitle(ce.colorPanelTitle())
	ce.updatePreview()
	ce.renderStatus()
	ce.scheduleLivePreview()
}

// refreshContrastColumn re-renders every row after the background moved.
func (ce *ColorEditor) refreshContrastColumn() {
	for key, idx := range ce.colorKeyToListItem {
		ce.colorPanel.SetItemText(idx, ce.formatColorItem(key), "")
	}
}

// undoCurrentColor restores the selected colour to the value on disk.
func (ce *ColorEditor) undoCurrentColor() {
	key := ce.selectedColorKey()
	if key == "" {
		return
	}
	original, ok := ce.originalValues[key]
	if !ok {
		return
	}
	ce.setColorValue(key, original)
	ce.isDirty = ce.hasUnsavedChanges()
	ce.colorPanel.SetTitle(ce.colorPanelTitle())
	ce.info("Reverted %s to %s", shortColorLabel(key), original)
}

// hasUnsavedChanges compares the working set against the on-disk snapshot.
func (ce *ColorEditor) hasUnsavedChanges() bool {
	if len(ce.colorValues) != len(ce.originalValues) {
		return true
	}
	for key, value := range ce.colorValues {
		if ce.originalValues[key] != value {
			return true
		}
	}
	return false
}

// --- debounced writes -----------------------------------------------------

// scheduleApply applies the theme under the cursor once the cursor rests.
// Browsing 150 themes with the arrow keys must not mean 150 config writes.
func (ce *ColorEditor) scheduleApply(themeName string) {
	seq := ce.applySeq.Add(1)

	go func() {
		time.Sleep(ApplyDebounce)
		if ce.applySeq.Load() != seq {
			return
		}
		err := ce.themeManager.ApplyTheme(themeName)
		ce.app.QueueUpdateDraw(func() {
			if ce.applySeq.Load() != seq {
				return
			}
			if err != nil {
				ce.fail("Could not apply %s: %v", themeName, err)
				return
			}
			ce.markApplied(themeName)
		})
	}()
}

// applyThemeNow skips the debounce for an explicit user action.
func (ce *ColorEditor) applyThemeNow(themeName string) {
	ce.applySeq.Add(1)
	if err := ce.themeManager.ApplyTheme(themeName); err != nil {
		ce.fail("Could not apply %s: %v", themeName, err)
		return
	}
	ce.markApplied(themeName)
	ce.info("Applied %s", themeName)
}

// markApplied records the newly live theme and restyles the UI to match it.
func (ce *ColorEditor) markApplied(themeName string) {
	previous := ce.appliedTheme
	ce.appliedTheme = themeName

	// Repaint only the two rows whose marker changed.
	if idx := ce.indexOfVisible(previous); idx >= 0 {
		ce.themeList.SetItemText(idx, ce.formatThemeItem(previous), "")
	}
	if idx := ce.indexOfVisible(themeName); idx >= 0 {
		ce.themeList.SetItemText(idx, ce.formatThemeItem(themeName), "")
	}

	ce.syncPaletteWithColors()
}

// scheduleLivePreview writes the working colours to themes/current.toml so the
// terminal repaints with the real colours, coalescing bursts of keystrokes.
func (ce *ColorEditor) scheduleLivePreview() {
	seq := ce.applySeq.Add(1)
	content := ce.generateTOMLContent()

	go func() {
		time.Sleep(EditApplyDebounce)
		if ce.applySeq.Load() != seq {
			return
		}
		err := ce.themeManager.WritePreview(content)
		ce.app.QueueUpdateDraw(func() {
			if err != nil {
				ce.fail("Live preview failed: %v", err)
				return
			}
			ce.syncPaletteWithColors()
		})
	}()
}

// syncPaletteWithColors keeps the editor chrome in step with the colours that
// are actually live in the terminal.
func (ce *ColorEditor) syncPaletteWithColors() {
	if len(ce.colorValues) == 0 {
		return
	}
	next := buildPalette(map[string]string{
		"background": ce.colorValues["primary.background"],
		"foreground": ce.colorValues["primary.foreground"],
		"blue":       firstNonEmpty(ce.colorValues["bright.blue"], ce.colorValues["normal.blue"]),
		"cyan":       firstNonEmpty(ce.colorValues["bright.cyan"], ce.colorValues["normal.cyan"]),
		"yellow":     firstNonEmpty(ce.colorValues["bright.yellow"], ce.colorValues["normal.yellow"]),
		"red":        firstNonEmpty(ce.colorValues["bright.red"], ce.colorValues["normal.red"]),
		"green":      firstNonEmpty(ce.colorValues["bright.green"], ce.colorValues["normal.green"]),
	})
	if next == ce.palette {
		return
	}

	ce.palette = next
	ce.palette.applyPaletteToStyles()
	ce.applyPaletteToWidgets()
	ce.rebuildThemeList()
	ce.refreshContrastColumn()
	ce.renderStatus()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
