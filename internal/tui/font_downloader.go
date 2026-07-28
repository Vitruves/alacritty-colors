package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/internal/font"
)

// The font downloader. The browser can only offer what is installed, which is
// a short list on a fresh machine, so this is the way to make it longer without
// leaving the editor.

// showFontDownloader lists the installable families and installs the chosen one.
func (ce *ColorEditor) showFontDownloader() {
	p := ce.palette

	list := tview.NewList()
	list.ShowSecondaryText(true)
	ce.styleList(list, " Download a font ")

	installed := ce.installedFamilies()

	for _, candidate := range font.AvailableDownloads {
		note := candidate.Note
		if installed[candidate.Name] {
			// Say so rather than hiding it: reinstalling is how you update.
			note += "  ·  already installed"
		}
		list.AddItem(candidate.Name, "   "+note, 0, nil)
	}

	hint := tview.NewTextView()
	hint.SetDynamicColors(true)
	hint.SetBackgroundColor(p.bg)
	hint.SetText(fmt.Sprintf("[%s]  Enter install · ↑↓ navigate · Esc back[-]", p.mutedHex))

	list.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
		if index < 0 || index >= len(font.AvailableDownloads) {
			return
		}
		ce.installFont(font.AvailableDownloads[index])
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape ||
			(event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			ce.closeOverlay("font-download")
			return nil
		}
		return event
	})

	root := tview.NewFlex()
	root.SetDirection(tview.FlexRow)
	root.SetBackgroundColor(p.bg)
	root.AddItem(list, 0, 1, true)
	root.AddItem(hint, 1, 0, false)

	ce.overlays = append(ce.overlays, "font-download")
	ce.pages.AddPage("font-download", ce.centered(root, 66, len(font.AvailableDownloads)*2+4), true, true)
	ce.app.SetFocus(list)
}

// installedFamilies is the set of families already on the system, matched by
// the display name being a prefix of the installed one — "JetBrains Mono" is
// installed as "JetBrainsMono Nerd Font Mono" and should still read as present.
func (ce *ColorEditor) installedFamilies() map[string]bool {
	present := make(map[string]bool)
	families := ce.fontManager.GetFontFamilies()

	for _, candidate := range font.AvailableDownloads {
		compact := removeSpaces(candidate.Name)
		for _, family := range families {
			if hasPrefixFold(removeSpaces(family), compact) {
				present[candidate.Name] = true
				break
			}
		}
	}
	return present
}

// installFont runs the download behind a spinner and reports what happened.
func (ce *ColorEditor) installFont(candidate font.FontDownload) {
	modal := ce.newModal(fmt.Sprintf("⠋ Fetching %s…\n\nThis can take a minute on a slow link.", candidate.Name))
	ce.showModal("font-install", modal)

	spinner := 0
	status := fmt.Sprintf("Fetching %s…", candidate.Name)
	ticker := time.NewTicker(SpinnerInterval)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				ce.app.QueueUpdateDraw(func() {
					spinner++
					modal.SetText(fmt.Sprintf("%s %s", SpinnerChars[spinner%len(SpinnerChars)], status))
				})
			}
		}
	}()

	go func() {
		err := ce.fontManager.Install(candidate, func(step string) {
			// Only the latest step is shown; the spinner carries the motion.
			ce.app.QueueUpdateDraw(func() { status = step })
		})

		ticker.Stop()
		close(done)

		ce.app.QueueUpdateDraw(func() {
			ce.closeOverlay("font-install")
			if err != nil {
				ce.fail("Could not install %s: %v", candidate.Name, err)
				return
			}

			// The browser behind is now out of date, so rebuild it from the
			// refreshed list rather than leaving the new font invisible.
			ce.closeOverlay("font-download")
			ce.closeOverlay("fonts")
			ce.info("Installed %s — reopening the font browser", candidate.Name)
			ce.showFontTUI()
		})
	}()
}

// centered wraps a primitive in fixed-size padding.
func (ce *ColorEditor) centered(content tview.Primitive, width, height int) tview.Primitive {
	pad := func() *tview.Box {
		box := tview.NewBox()
		box.SetBackgroundColor(ce.palette.bg)
		return box
	}

	row := tview.NewFlex()
	row.AddItem(pad(), 0, 1, false)
	row.AddItem(content, width, 0, true)
	row.AddItem(pad(), 0, 1, false)

	wrapper := tview.NewFlex()
	wrapper.SetDirection(tview.FlexRow)
	wrapper.AddItem(pad(), 0, 1, false)
	wrapper.AddItem(row, height, 0, true)
	wrapper.AddItem(pad(), 0, 1, false)
	return wrapper
}

// removeSpaces strips spaces so "JetBrains Mono" and "JetBrainsMono" compare
// equal, which is the difference between a project name and its font file.
func removeSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}
