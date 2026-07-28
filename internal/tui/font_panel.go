package tui

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showFontTUI displays the font configuration interface
func (ce *ColorEditor) showFontTUI() {
	loading := ce.newModal("⠋ Scanning your system for monospace fonts…\n\nEsc to cancel")

	spinnerIndex := 0
	cancelled := false
	ticker := time.NewTicker(SpinnerInterval)

	stop := func() {
		if !cancelled {
			cancelled = true
			ticker.Stop()
		}
	}

	go func() {
		for range ticker.C {
			if cancelled {
				return
			}
			ce.app.QueueUpdateDraw(func() {
				spinnerIndex++
				loading.SetText(fmt.Sprintf("%s Scanning your system for monospace fonts…\n\nEsc to cancel",
					SpinnerChars[spinnerIndex%len(SpinnerChars)]))
			})
		}
	}()

	loading.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			stop()
			ce.closeOverlay("fonts")
			return nil
		}
		return event
	})

	ce.showModal("fonts", loading)

	go func() {
		currentFamily := ce.fontManager.GetCurrentFontFamily()
		currentStyle := ce.fontManager.GetCurrentFontStyle()
		currentSize := ce.fontManager.GetCurrentFontSize()

		// No fallback list here on purpose. Naming fonts that may not be
		// installed is what produced "unable to load font": an empty column is
		// honest, a column of options that cannot load is not.
		monoFonts := ce.fontManager.GetFontFamilies()

		ce.app.QueueUpdateDraw(func() {
			if cancelled {
				return
			}
			stop()
			// Pop the loading modal off the overlay stack before pushing the
			// browser, otherwise the stack never empties and the main view
			// stays deaf to every global key.
			ce.closeOverlay("fonts")
			ce.setupActualFontTUI(monoFonts, currentFamily, currentStyle, currentSize)
		})
	}()
}

// setupActualFontTUI sets up the actual font configuration interface
func (ce *ColorEditor) setupActualFontTUI(monoFonts []string, currentFamily, currentStyle string, currentSize float64) {
	p := ce.palette

	fontList := tview.NewList()
	fontList.ShowSecondaryText(false)
	ce.styleList(fontList, " Font families ")

	styleList := tview.NewList()
	styleList.ShowSecondaryText(false)
	ce.styleList(styleList, " Styles ")

	sizeList := tview.NewList()
	sizeList.ShowSecondaryText(false)
	ce.styleList(sizeList, " Size ")

	infoPanel := tview.NewTextView()
	infoPanel.SetDynamicColors(true)
	infoPanel.SetWordWrap(true)
	infoPanel.SetTextColor(p.fg)
	ce.styleBox(infoPanel.Box, " Sample ")

	hint := tview.NewTextView()
	hint.SetDynamicColors(true)
	hint.SetBackgroundColor(p.bg)
	hint.SetText(fmt.Sprintf("[%s]%s[-]", p.mutedHex, StatusBarFontPanel))

	selectedFamily := currentFamily
	selectedStyle := currentStyle

	if len(monoFonts) == 0 {
		ce.showModal("fonts", ce.newModal(noFontsMessage()))
		return
	}

	for i, family := range monoFonts {
		marker := " "
		if family == currentFamily {
			marker = "●"
		}
		fontList.AddItem(fmt.Sprintf("[%s]%s[-] %s", p.accentHex, marker, family), "", 0, nil)
		if family == currentFamily {
			fontList.SetCurrentItem(i)
		}
	}

	updateInfo := func() {
		display := selectedFamily
		if len(display) > MaxFontNameDisplay {
			display = display[:MaxFontNameDisplay-1] + "…"
		}

		infoPanel.SetText(fmt.Sprintf(`[%s::b]%s[-::-]  [%s]%s · %.1fpt[-]

The quick brown fox jumps over the lazy dog
ABCDEFGHIJKLMNOPQRSTUVWXYZ
abcdefghijklmnopqrstuvwxyz
0123456789 !@#$%%^&*() {}[]<>
il1I| oO0 rn m ,.;: '"`+"`"+`

[%s]func[-] [%s]main[-]() {
    [%s]fmt[-].[%s]Println[-]([%s]"Hello, World!"[-])
    [%s]// changes apply to Alacritty immediately[-]
}

[%s]user@host[-]:[%s]~/projects[-]$ [%s]ls -la[-]`,
			p.accentHex, display, p.mutedHex, selectedStyle, currentSize,
			p.accentHex, p.warnHex, p.accent2Hex, p.warnHex, p.okHex, p.mutedHex,
			p.okHex, p.accentHex, p.warnHex))
	}

	loadStyles := func(family string) {
		styleList.Clear()
		styles := ce.fontManager.GetStylesForFont(family)
		if len(styles) == 0 {
			// Generic names like "monospace" resolve to no concrete faces;
			// offer the four Alacritty always understands.
			styles = []string{"Regular", "Bold", "Italic", "Bold Italic"}
		}
		for i, style := range styles {
			styleList.AddItem(style, "", 0, nil)
			if style == selectedStyle {
				styleList.SetCurrentItem(i)
			}
		}
		if styleList.GetItemCount() > 0 && styleList.GetCurrentItem() >= styleList.GetItemCount() {
			styleList.SetCurrentItem(0)
		}
	}

	// Moving the cursor through a few hundred families must not mean a few
	// hundred rewrites of the user's config. Only the newest request survives
	// the debounce, exactly as browsing the theme list already works.
	applyFont := func() {
		family, style := selectedFamily, selectedStyle
		seq := ce.fontSeq.Add(1)

		go func() {
			time.Sleep(ApplyDebounce)
			if ce.fontSeq.Load() != seq {
				return
			}
			if err := ce.fontManager.UpdateFontFamily(family); err != nil {
				return
			}
			if err := ce.fontManager.UpdateFontStyle(style); err != nil {
				ce.fontManager.UpdateFontStyle("Regular")
			}
		}()
	}

	pickFamily := func(index int) {
		if index < 0 || index >= len(monoFonts) {
			return
		}
		selectedFamily = monoFonts[index]
		loadStyles(selectedFamily)
		if styleList.GetItemCount() > 0 {
			selectedStyle, _ = styleList.GetItemText(styleList.GetCurrentItem())
		}
		updateInfo()
		applyFont()
	}

	// applySize writes the size on the same debounce as the family, so holding
	// an arrow key or scrolling the size list costs one write, not thirty.
	applySize := func() {
		size := currentSize
		seq := ce.fontSeq.Add(1)

		go func() {
			time.Sleep(ApplyDebounce)
			if ce.fontSeq.Load() != seq {
				return
			}
			ce.fontManager.UpdateFontSize(fmt.Sprintf("%.1f", size))
		}()
	}

	// syncSizeSelection moves the size list cursor onto currentSize without
	// letting the resulting change event write the value straight back.
	syncingSize := false
	syncSizeSelection := func() {
		for i, size := range fontSizes {
			if size == currentSize {
				syncingSize = true
				sizeList.SetCurrentItem(i)
				syncingSize = false
				return
			}
		}
	}

	adjustSize := func(increase bool) {
		step := FontSizeStep
		if !increase {
			step = -step
		}
		currentSize = clampFloat(currentSize+step, FontSizeMin, FontSizeMax)
		updateInfo()
		syncSizeSelection()
		applySize()
	}

	for _, size := range fontSizes {
		sizeList.AddItem(fmt.Sprintf("%.1f pt", size), "", 0, nil)
	}
	syncSizeSelection()

	sizeList.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		if syncingSize || index < 0 || index >= len(fontSizes) {
			return
		}
		currentSize = fontSizes[index]
		updateInfo()
		applySize()
	})

	boxes := []*tview.Box{fontList.Box, styleList.Box, sizeList.Box}
	targets := []tview.Primitive{fontList, styleList, sizeList}
	focusIdx := 0

	capture := func(self int) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab, tcell.KeyBacktab:
				focusIdx = (self + 1) % len(targets)
				ce.focusRing(boxes, targets, focusIdx)
				return nil
			case tcell.KeyLeft:
				adjustSize(false)
				return nil
			case tcell.KeyRight:
				adjustSize(true)
				return nil
			case tcell.KeyEscape:
				ce.closeOverlay("fonts")
				return nil
			case tcell.KeyRune:
				switch event.Rune() {
				case 'q', 'Q':
					ce.closeOverlay("fonts")
					return nil
				case 'd', 'D':
					ce.showFontDownloader()
					return nil
				}
			}
			return event
		}
	}

	fontList.SetInputCapture(capture(0))
	styleList.SetInputCapture(capture(1))
	sizeList.SetInputCapture(capture(2))

	fontList.SetChangedFunc(func(index int, _ string, _ string, _ rune) { pickFamily(index) })
	styleList.SetChangedFunc(func(_ int, style string, _ string, _ rune) {
		selectedStyle = style
		updateInfo()
		applyFont()
	})

	loadStyles(selectedFamily)
	updateInfo()

	body := tview.NewFlex()
	body.AddItem(fontList, 0, 2, true)
	body.AddItem(styleList, 0, 1, false)
	body.AddItem(sizeList, 10, 0, false)
	body.AddItem(infoPanel, 0, 3, false)

	root := tview.NewFlex()
	root.SetDirection(tview.FlexRow)
	root.AddItem(body, 0, 1, true)
	root.AddItem(hint, 1, 0, false)
	root.SetBackgroundColor(p.bg)

	ce.overlays = append(ce.overlays, "fonts")
	ce.pages.AddPage("fonts", root, true, true)
	ce.focusRing(boxes, targets, focusIdx)
}

// noFontsMessage explains an empty font browser in terms of the thing that is
// actually missing on this platform, rather than naming a Unix tool at someone
// running Windows.
func noFontsMessage() string {
	const intro = "No monospace font could be listed.\n\n" +
		"Only families the terminal can genuinely resolve are offered, so\n" +
		"nothing here can fail to load — but that also means an empty list\n" +
		"when the font database cannot be read.\n\n"

	switch runtime.GOOS {
	case "windows":
		return intro +
			"Font browsing is not implemented on Windows yet. Set the family\n" +
			"by hand in alacritty.toml; everything else in this editor works.\n\n" +
			"Esc to go back"
	case "darwin":
		return intro +
			"macOS is queried through Core Text, which should always answer.\n" +
			"If you are seeing this, please report it.\n\nEsc to go back"
	default:
		return intro +
			"Install fontconfig and check that `fc-list :spacing=mono family`\n" +
			"returns something.\n\nEsc to go back"
	}
}
