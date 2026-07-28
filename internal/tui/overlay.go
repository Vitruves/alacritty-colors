package tui

import (
	"github.com/rivo/tview"
)

// Overlays are pages stacked on top of the main layout. Using tview.Pages
// instead of swapping the root means the editor behind stays intact, so
// cancelling a dialog costs nothing and never rebuilds the UI.

// showOverlay centres a primitive over the main view.
func (ce *ColorEditor) showOverlay(name string, content tview.Primitive, width, height int) {
	row := tview.NewFlex()
	row.AddItem(nil, 0, 1, false)
	row.AddItem(content, width, 0, true)
	row.AddItem(nil, 0, 1, false)

	wrapper := tview.NewFlex()
	wrapper.SetDirection(tview.FlexRow)
	wrapper.AddItem(nil, 0, 1, false)
	wrapper.AddItem(row, height, 0, true)
	wrapper.AddItem(nil, 0, 1, false)

	ce.overlays = append(ce.overlays, name)
	ce.pages.AddPage(name, wrapper, true, true)
	ce.app.SetFocus(content)
}

// showModal stacks a full-screen-dimmed modal dialog.
func (ce *ColorEditor) showModal(name string, modal *tview.Modal) {
	ce.overlays = append(ce.overlays, name)
	ce.pages.AddPage(name, modal, true, true)
	ce.app.SetFocus(modal)
}

// closeOverlay removes the topmost overlay and restores focus underneath.
func (ce *ColorEditor) closeOverlay(name string) {
	ce.pages.RemovePage(name)
	for i, n := range ce.overlays {
		if n == name {
			ce.overlays = append(ce.overlays[:i], ce.overlays[i+1:]...)
			break
		}
	}
	if len(ce.overlays) == 0 {
		ce.setFocus(ce.focus)
	}
}

// overlayOpen reports whether a dialog currently owns the keyboard.
func (ce *ColorEditor) overlayOpen() bool {
	return len(ce.overlays) > 0
}

// newModal builds a dialog styled from the active theme rather than a hard
// coded black-and-blue that clashes with whatever the user is editing.
func (ce *ColorEditor) newModal(text string) *tview.Modal {
	p := ce.palette

	modal := tview.NewModal()
	modal.SetText(text)
	modal.SetBackgroundColor(p.bg)
	modal.SetTextColor(p.fg)
	modal.SetButtonBackgroundColor(p.selBg)
	modal.SetButtonTextColor(p.fg)
	modal.SetBorderColor(p.accent)
	modal.SetTitleColor(p.accent)
	return modal
}

// styleBox applies the palette to any bordered primitive used in an overlay.
func (ce *ColorEditor) styleBox(box *tview.Box, title string) {
	box.SetBorder(true)
	box.SetTitle(title)
	box.SetTitleAlign(tview.AlignLeft)
	box.SetBackgroundColor(ce.palette.bg)
	box.SetBorderColor(ce.palette.accent)
	box.SetTitleColor(ce.palette.accent)
}

// styleList applies the palette to a list used inside an overlay.
func (ce *ColorEditor) styleList(list *tview.List, title string) {
	ce.styleBox(list.Box, title)
	list.SetMainTextColor(ce.palette.fg)
	list.SetSecondaryTextColor(ce.palette.muted)
	list.SetSelectedTextColor(ce.palette.selFg)
	list.SetSelectedBackgroundColor(ce.palette.selBg)
	list.SetHighlightFullLine(true)
}

// focusRing wires Tab/Shift+Tab cycling between a fixed set of primitives and
// keeps the accent border on whichever one holds focus.
func (ce *ColorEditor) focusRing(boxes []*tview.Box, targets []tview.Primitive, index int) {
	for i, box := range boxes {
		if i == index {
			box.SetBorderColor(ce.palette.accent)
			box.SetTitleColor(ce.palette.accent)
		} else {
			box.SetBorderColor(ce.palette.border)
			box.SetTitleColor(ce.palette.muted)
		}
	}
	ce.app.SetFocus(targets[index])
}
