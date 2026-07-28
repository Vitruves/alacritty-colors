package tui

import (
	"github.com/gdamore/tcell/v2"
)

// openFilter reveals the filter line and hands it the keyboard. Filtering is
// inline rather than a modal so the theme list, palette and preview stay
// visible while you narrow the list down.
func (ce *ColorEditor) openFilter() {
	ce.filterFocus = true
	ce.rootFlex.ResizeItem(ce.filterInput, 1, 0)
	ce.filterInput.SetText(ce.filter)
	ce.app.SetFocus(ce.filterInput)
	ce.statusKeys.SetText("[" + ce.palette.mutedHex + "]type to search · ↑↓ move through results · Enter keep · Esc cancel[-]")
}

// closeFilter hides the filter line and returns focus to the theme list.
func (ce *ColorEditor) closeFilter() {
	ce.filterFocus = false
	ce.rootFlex.ResizeItem(ce.filterInput, 0, 0)
	ce.setFocus(FocusThemeList)
}

// clearFilter drops both the text filter and the favourites-only restriction.
func (ce *ColorEditor) clearFilter() {
	ce.filter = ""
	ce.favOnly = false
	ce.filterInput.SetText("")
	ce.rebuildThemeList()
	ce.closeFilter()
	ce.renderStatus()
}

// onFilterChanged narrows the list on every keystroke and previews the top hit.
func (ce *ColorEditor) onFilterChanged(text string) {
	ce.filter = text
	ce.rebuildThemeList()
	if name := ce.currentVisibleName(); name != "" && name != ce.themeName {
		ce.selectTheme(name, true)
	}
}

// handleFilterKeys lets the arrow keys walk the filtered results without
// leaving the input field.
func (ce *ColorEditor) handleFilterKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		ce.filter = ""
		ce.filterInput.SetText("")
		ce.rebuildThemeList()
		if name := ce.currentVisibleName(); name != "" {
			ce.selectTheme(name, true)
		}
		ce.closeFilter()
		return nil
	case tcell.KeyEnter:
		ce.closeFilter()
		if name := ce.currentVisibleName(); name != "" {
			ce.applyThemeNow(name)
		}
		return nil
	case tcell.KeyUp:
		ce.moveThemeSelection(-1)
		return nil
	case tcell.KeyDown:
		ce.moveThemeSelection(1)
		return nil
	}
	return event
}

// toggleFavoritesOnly flips the list between all themes and favourites.
func (ce *ColorEditor) toggleFavoritesOnly() {
	if !ce.favOnly && len(ce.favorites) == 0 {
		ce.warn("No favourites yet — press * to add the selected theme")
		return
	}
	ce.favOnly = !ce.favOnly
	ce.rebuildThemeList()
	if name := ce.currentVisibleName(); name != "" && name != ce.themeName {
		ce.selectTheme(name, true)
	}
	if ce.favOnly {
		ce.info("Showing favourites only (%d)", len(ce.visible))
	} else {
		ce.info("Showing all themes (%d)", len(ce.visible))
	}
}
