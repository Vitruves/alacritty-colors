package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showFontTUI displays the font configuration interface
func (ce *ColorEditor) showFontTUI() {
	ce.applyUserThemeToTUI()

	// Disable global key handler while font panel is open
	ce.app.SetInputCapture(nil)

	loadingModal := tview.NewModal()
	loadingModal.SetBackgroundColor(tcell.ColorBlack)
	loadingModal.SetTextColor(tcell.ColorWhite)

	spinnerIndex := 0
	loadingCancelled := false

	updateSpinner := func() {
		if !loadingCancelled {
			spinner := SpinnerChars[spinnerIndex%len(SpinnerChars)]
			loadingModal.SetText(fmt.Sprintf("%s Loading fonts...\n\nScanning your system for available monospace fonts.\n\nPress 'q' or Escape to cancel.", spinner))
			spinnerIndex++
		}
	}

	updateSpinner()

	spinnerTicker := time.NewTicker(SpinnerInterval)
	go func() {
		for range spinnerTicker.C {
			if loadingCancelled {
				spinnerTicker.Stop()
				return
			}
			ce.app.QueueUpdateDraw(func() {
				updateSpinner()
			})
		}
	}()

	loadingModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			loadingCancelled = true
			spinnerTicker.Stop()
			ce.restoreMainUI()
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'q' || event.Rune() == 'Q' {
				loadingCancelled = true
				spinnerTicker.Stop()
				ce.restoreMainUI()
				return nil
			}
		}
		return event
	})

	ce.app.SetRoot(loadingModal, true)

	go func() {
		currentFamily := ce.fontManager.GetCurrentFontFamily()
		currentStyle := ce.fontManager.GetCurrentFontStyle()
		currentSize := ce.fontManager.GetCurrentFontSize()

		monoFonts := ce.fontManager.GetFontFamilies()
		if len(monoFonts) == 0 {
			monoFonts = []string{"monospace", "JetBrains Mono", "Fira Code", "Source Code Pro"}
		}

		ce.app.QueueUpdateDraw(func() {
			if !loadingCancelled {
				loadingCancelled = true
				spinnerTicker.Stop()
				ce.setupActualFontTUI(monoFonts, currentFamily, currentStyle, currentSize)
			}
		})
	}()
}

// setupActualFontTUI sets up the actual font configuration interface
func (ce *ColorEditor) setupActualFontTUI(monoFonts []string, currentFamily, currentStyle string, currentSize float64) {
	fontList := tview.NewList()
	fontList.ShowSecondaryText(false)
	fontList.SetMainTextColor(tcell.ColorWhite)
	fontList.SetSelectedTextColor(tcell.ColorBlack)
	fontList.SetSelectedBackgroundColor(tcell.ColorWhite)
	fontList.SetBorder(true)
	fontList.SetTitle(" Font Families ")

	styleList := tview.NewList()
	styleList.ShowSecondaryText(false)
	styleList.SetMainTextColor(tcell.ColorWhite)
	styleList.SetSelectedTextColor(tcell.ColorBlack)
	styleList.SetSelectedBackgroundColor(tcell.ColorWhite)
	styleList.SetBorder(true)
	styleList.SetTitle(" Font Styles ")

	infoPanel := tview.NewTextView()
	infoPanel.SetDynamicColors(true)
	infoPanel.SetWordWrap(true)
	infoPanel.SetBorder(true)
	infoPanel.SetTitle(" Font Preview ")

	statusBar := tview.NewTextView()
	statusBar.SetText(StatusBarFontPanel)
	statusBar.SetTextColor(tcell.ColorYellow)

	currentFontIndex := -1
	for i, fontFamily := range monoFonts {
		displayName := fontFamily
		if fontFamily == currentFamily {
			displayName = fmt.Sprintf("%s%s", CurrentThemeMarker, fontFamily)
			currentFontIndex = i
		}
		fontList.AddItem(displayName, "", 0, nil)
	}

	loadStylesForFont := func(fontFamily string) {
		styleList.Clear()
		styles := ce.fontManager.GetStylesForFont(fontFamily)
		currentStyleIndex := -1
		for i, style := range styles {
			displayName := style
			if style == currentStyle && fontFamily == currentFamily {
				displayName = fmt.Sprintf("%s%s", CurrentThemeMarker, style)
				currentStyleIndex = i
			}
			styleList.AddItem(displayName, "", 0, nil)
		}
		if currentStyleIndex >= 0 {
			styleList.SetCurrentItem(currentStyleIndex)
		} else if len(styles) > 0 {
			styleList.SetCurrentItem(0)
		}
	}

	updateFontInfo := func(fontFamily, style string, size float64) {
		displayFamily := fontFamily
		if len(fontFamily) > MaxFontNameDisplay {
			displayFamily = fontFamily[:MaxFontNameDisplay-3] + "..."
		}

		info := fmt.Sprintf(`[yellow::b]Font Configuration[-]

[white::b]Family:[-] %s
[white::b]Style:[-]  %s
[white::b]Size:[-]   %.1f

[white::b]Sample Text:[-]
[white]The quick brown fox jumps over the lazy dog.[-]
[white]ABCDEFGHIJKLMNOPQRSTUVWXYZ[-]
[white]abcdefghijklmnopqrstuvwxyz[-]
[white]0123456789 !@#$%%^&*()[-]

[white::b]Code Sample:[-]
[cyan]func[-] [yellow]main[-]() {
    [cyan]fmt[-].[yellow]Println[-]([green]"Hello, World!"[-])
    [magenta]// This is a comment[-]
}

[white::b]Terminal Commands:[-]
[green]user@hostname[-]:[blue]~/projects[-]$ [yellow]ls -la[-]
[white]total 24[-]
[blue]drwxr-xr-x[-] [white]3 user staff 96 Jan 15 10:30[-] [blue].[-]

[white]Note: Font changes apply instantly to Alacritty[-]
[white]Use Left/Right arrows to adjust font size[-]`,
			displayFamily, style, size)
		infoPanel.SetText(info)
	}

	fontList.SetSelectedFunc(func(index int, fontFamily string, _ string, _ rune) {
		fontFamily = strings.TrimPrefix(fontFamily, CurrentThemeMarker)
		loadStylesForFont(fontFamily)
		if styleList.GetItemCount() > 0 {
			styleName, _ := styleList.GetItemText(styleList.GetCurrentItem())
			styleName = strings.TrimPrefix(styleName, CurrentThemeMarker)
			updateFontInfo(fontFamily, styleName, currentSize)

			go func() {
				if err := ce.fontManager.UpdateFontFamily(fontFamily); err == nil {
					if err := ce.fontManager.UpdateFontStyle(styleName); err != nil {
						ce.fontManager.UpdateFontStyle("Regular")
					}
				}
			}()
		}
	})

	styleList.SetSelectedFunc(func(index int, style string, _ string, _ rune) {
		style = strings.TrimPrefix(style, CurrentThemeMarker)
		fontIndex := fontList.GetCurrentItem()
		if fontIndex >= 0 {
			fontFamily, _ := fontList.GetItemText(fontIndex)
			fontFamily = strings.TrimPrefix(fontFamily, CurrentThemeMarker)
			updateFontInfo(fontFamily, style, currentSize)

			go func() {
				ce.fontManager.UpdateFontStyle(style)
			}()
		}
	})

	adjustFontSize := func(increase bool) {
		adjustment := FontSizeStep
		if !increase {
			adjustment = -adjustment
		}
		newSize := currentSize + adjustment
		if newSize < FontSizeMin {
			newSize = FontSizeMin
		} else if newSize > FontSizeMax {
			newSize = FontSizeMax
		}
		currentSize = newSize

		fontIndex := fontList.GetCurrentItem()
		styleIndex := styleList.GetCurrentItem()
		if fontIndex >= 0 && styleIndex >= 0 {
			fontFamily, _ := fontList.GetItemText(fontIndex)
			fontFamily = strings.TrimPrefix(fontFamily, CurrentThemeMarker)
			styleName, _ := styleList.GetItemText(styleIndex)
			styleName = strings.TrimPrefix(styleName, CurrentThemeMarker)
			updateFontInfo(fontFamily, styleName, currentSize)

			go func() {
				ce.fontManager.UpdateFontSize(fmt.Sprintf("%.1f", currentSize))
			}()
		}
	}

	fontList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ce.app.SetFocus(styleList)
			styleList.SetBorderColor(tcell.ColorYellow)
			fontList.SetBorderColor(tcell.ColorDefault)
			statusBar.SetText("Tab: switch panels | ↑↓: navigate styles | ←→: size | q: quit")
			return nil
		case tcell.KeyLeft:
			adjustFontSize(false)
			return nil
		case tcell.KeyRight:
			adjustFontSize(true)
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			result := event
			go func() {
				time.Sleep(KeyDebounceDelay)
				ce.app.QueueUpdateDraw(func() {
					index := fontList.GetCurrentItem()
					if index >= 0 {
						fontFamily, _ := fontList.GetItemText(index)
						fontFamily = strings.TrimPrefix(fontFamily, CurrentThemeMarker)
						loadStylesForFont(fontFamily)
						if styleList.GetItemCount() > 0 {
							styleName, _ := styleList.GetItemText(styleList.GetCurrentItem())
							styleName = strings.TrimPrefix(styleName, CurrentThemeMarker)
							updateFontInfo(fontFamily, styleName, currentSize)

							go func() {
								if err := ce.fontManager.UpdateFontFamily(fontFamily); err == nil {
									if err := ce.fontManager.UpdateFontStyle(styleName); err != nil {
										ce.fontManager.UpdateFontStyle("Regular")
									}
								}
							}()
						}
					}
				})
			}()
			return result
		}
		return event
	})

	styleList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			ce.app.SetFocus(fontList)
			fontList.SetBorderColor(tcell.ColorYellow)
			styleList.SetBorderColor(tcell.ColorDefault)
			statusBar.SetText(StatusBarFontPanel)
			return nil
		case tcell.KeyLeft:
			adjustFontSize(false)
			return nil
		case tcell.KeyRight:
			adjustFontSize(true)
			return nil
		case tcell.KeyUp, tcell.KeyDown:
			result := event
			go func() {
				time.Sleep(KeyDebounceDelay)
				ce.app.QueueUpdateDraw(func() {
					styleIndex := styleList.GetCurrentItem()
					fontIndex := fontList.GetCurrentItem()
					if styleIndex >= 0 && fontIndex >= 0 {
						styleName, _ := styleList.GetItemText(styleIndex)
						styleName = strings.TrimPrefix(styleName, CurrentThemeMarker)
						fontFamily, _ := fontList.GetItemText(fontIndex)
						fontFamily = strings.TrimPrefix(fontFamily, CurrentThemeMarker)
						updateFontInfo(fontFamily, styleName, currentSize)

						go func() {
							ce.fontManager.UpdateFontStyle(styleName)
						}()
					}
				})
			}()
			return result
		}
		return event
	})

	globalKeyHandler := func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			ce.restoreMainUI()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				ce.restoreMainUI()
				return nil
			}
		}
		return event
	}

	if currentFontIndex >= 0 {
		fontList.SetCurrentItem(currentFontIndex)
		loadStylesForFont(currentFamily)
		updateFontInfo(currentFamily, currentStyle, currentSize)
	} else if len(monoFonts) > 0 {
		fontList.SetCurrentItem(0)
		loadStylesForFont(monoFonts[0])
		updateFontInfo(monoFonts[0], "Regular", currentSize)
	}

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(fontList, 0, 2, true)
	mainFlex.AddItem(styleList, 0, 1, false)
	mainFlex.AddItem(infoPanel, 0, 3, false)

	rootFlex := tview.NewFlex()
	rootFlex.SetDirection(tview.FlexRow)
	rootFlex.AddItem(mainFlex, 0, 1, true)
	rootFlex.AddItem(statusBar, 1, 0, false)

	ce.app.SetRoot(rootFlex, true)
	ce.app.SetInputCapture(globalKeyHandler)
	ce.app.SetFocus(fontList)
	fontList.SetBorderColor(tcell.ColorYellow)
}

