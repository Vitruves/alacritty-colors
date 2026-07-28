package tui

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/internal/theme"
)

// showParametersPanel lists the maintenance actions. A list rather than a row
// of buttons, so each entry has room to say what it does.
func (ce *ColorEditor) showParametersPanel() {
	actions := []struct {
		label, detail string
		run           func()
	}{
		{"Install the curated collection", fmt.Sprintf("Add %d designed themes to your library", len(Collection)), ce.installCollection},
		{"Redownload themes", "Refetch the official Alacritty theme set", ce.cleanAndRedownloadThemes},
		{"Back up config", "Copy alacritty.toml into backups/", ce.backupCurrentConfig},
		{"Reset to defaults", "Restore the stock configuration", ce.resetToDefaults},
		{"Cancel", "Close this panel", func() {}},
	}

	list := tview.NewList()
	list.ShowSecondaryText(true)
	ce.styleList(list, " Parameters & utilities ")

	for _, action := range actions {
		run := action.run
		list.AddItem(action.label, action.detail, 0, func() {
			ce.closeOverlay("params")
			run()
		})
	}

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || (event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q')) {
			ce.closeOverlay("params")
			return nil
		}
		return event
	})

	ce.showOverlay("params", list, 58, 12)
}

// installCollection writes the curated themes and refreshes the list.
func (ce *ColorEditor) installCollection() {
	ce.info("Installing the curated collection…")

	go func() {
		written, skipped, err := InstallCollection(ce.config.ThemesDir)
		ce.app.QueueUpdateDraw(func() {
			if err != nil {
				ce.fail("Could not install the collection: %v", err)
				return
			}
			ce.refreshThemeList()
			if skipped > 0 {
				ce.info("Installed %d themes · left %d of your edited copies alone", written, skipped)
			} else {
				ce.info("Installed %d curated themes — press / and type a name to find them", written)
			}
		})
	}()
}

// showDeleteConfirmation shows the delete confirmation dialog
func (ce *ColorEditor) showDeleteConfirmation(name string) {
	modal := ce.newModal(fmt.Sprintf("Delete theme '%s'?\n\nThis cannot be undone.", name))
	modal.AddButtons([]string{"Cancel", "Delete"})
	modal.SetButtonTextColor(ce.palette.fg)

	modal.SetDoneFunc(func(buttonIndex int, _ string) {
		ce.closeOverlay("delete")
		if buttonIndex == 1 {
			ce.performThemeDelete(name)
		}
	})

	ce.showModal("delete", modal)
}

// showHelpOverlay shows the keyboard shortcuts help in two columns so it fits
// on a short terminal without scrolling.
func (ce *ColorEditor) showHelpOverlay() {
	p := ce.palette

	build := func(groups [][2]interface{}) string {
		var b strings.Builder
		for _, group := range groups {
			fmt.Fprintf(&b, "[%s::b]%s[-::-]\n", p.accentHex, group[0].(string))
			for _, entry := range group[1].([][2]string) {
				fmt.Fprintf(&b, "  [%s]%-11s[-] %s\n", p.warnHex, entry[0], entry[1])
			}
			b.WriteString("\n")
		}
		return b.String()
	}

	left := build([][2]interface{}{
		{"Everywhere", [][2]string{
			{"Tab", "Next column"},
			{"a", "Apply what you see, now"},
			{"s / S", "Save / save as…"},
			{"? / q", "This help / quit"},
		}},
		{"Navigation", [][2]string{
			{"↑ ↓ j k", "Move"},
			{"PgUp PgDn", "Jump ten rows"},
			{"Home End", "First / last"},
			{"/", "Search themes as you type"},
			{"Esc", "Clear the search"},
		}},
		{"Themes", [][2]string{
			{"↑ ↓", "Browsing applies live"},
			{"*", "Toggle favourite"},
			{"F", "Favourites only"},
			{"e", "Fork an editable copy"},
			{"d", "Delete your own theme"},
			{"n", "Theme creator"},
			{"g", "Instant harmony theme"},
		}},
	})

	right := build([][2]interface{}{
		{"Colours", [][2]string{
			{"← →", "Brighten / darken"},
			{"Shift+← →", "Rotate hue"},
			{"- +", "Saturation"},
			{"[ ]", "Lightness"},
			{"Enter / #", "Type an exact hex"},
			{"u", "Undo this colour"},
			{"r", "Revert the theme"},
		}},
		{"Elsewhere", [][2]string{
			{"c", "Preview colour mode"},
			{"f", "Font settings"},
			{"p", "Parameters"},
		}},
	})

	right += fmt.Sprintf("[%s]Edits show live in your terminal.\nBrowsing applies as you go, a\napplies at once, and quitting\nkeeps what you landed on —\nsaving it first if you have not.[-]", p.mutedHex)

	column := func(text string) *tview.TextView {
		tv := tview.NewTextView()
		tv.SetDynamicColors(true)
		tv.SetText(text)
		tv.SetTextColor(p.fg)
		tv.SetBackgroundColor(p.bg)
		return tv
	}

	body := tview.NewFlex()
	body.AddItem(column(left), 0, 1, false)
	body.AddItem(column(right), 0, 1, false)
	ce.styleBox(body.Box, " Keyboard shortcuts · any key closes ")
	body.SetBackgroundColor(p.bg)

	body.SetInputCapture(func(*tcell.EventKey) *tcell.EventKey {
		ce.closeOverlay("help")
		return nil
	})

	ce.showOverlay("help", body, 76, 25)
}

// showHexInput lets the user type an exact colour instead of nudging it.
func (ce *ColorEditor) showHexInput() {
	key := ce.selectedColorKey()
	if key == "" {
		ce.warn("Move to a colour row first")
		return
	}

	original, defined := ce.colorValues[key]
	seed := original
	if !defined || seed == "" {
		seed = ce.derivedDefault(key)
	}
	p := ce.palette

	input := tview.NewInputField()
	input.SetLabel(" hex ")
	input.SetText(seed)
	input.SetFieldWidth(12)
	input.SetLabelColor(p.accent)
	input.SetFieldBackgroundColor(p.selBg)
	input.SetFieldTextColor(p.fg)
	ce.styleBox(input.Box, fmt.Sprintf(" %s ", strings.ReplaceAll(key, ".", " ")))

	// Live: a valid value repaints the terminal as you type.
	input.SetChangedFunc(func(text string) {
		if hex, ok := expandHex(text); ok {
			ce.setColorValue(key, hex)
		}
	})

	restore := func() {
		if defined {
			ce.setColorValue(key, original)
		} else {
			// The slot was empty before this dialog; leave it empty.
			delete(ce.colorValues, key)
			ce.colorPanel.SetItemText(ce.colorKeyToListItem[key], ce.formatColorItem(key), "")
			ce.updatePreview()
			ce.scheduleLivePreview()
		}
		ce.isDirty = ce.hasUnsavedChanges()
		ce.colorPanel.SetTitle(ce.colorPanelTitle())
		ce.renderStatus()
	}

	input.SetDoneFunc(func(k tcell.Key) {
		ce.closeOverlay("hex")
		switch k {
		case tcell.KeyEnter:
			if hex, ok := expandHex(input.GetText()); ok {
				ce.setColorValue(key, hex)
				ce.info("%s set to %s", shortColorLabel(key), hex)
			} else {
				restore()
				ce.warn("Not a hex colour — nothing changed")
			}
		case tcell.KeyEscape:
			restore()
		}
	})

	ce.showOverlay("hex", input, 30, 3)
}

// expandHex accepts "#rrggbb", "rrggbb", "#rgb" and "rgb".
func expandHex(text string) (string, bool) {
	text = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "#"))
	if len(text) == 3 {
		text = string([]byte{text[0], text[0], text[1], text[1], text[2], text[2]})
	}
	if len(text) != 6 {
		return "", false
	}
	hex := "#" + strings.ToLower(text)
	if _, err := theme.HexToRGB(hex); err != nil {
		return "", false
	}
	return hex, true
}

// showSaveAsDialog asks for a name before writing a new theme file.
func (ce *ColorEditor) showSaveAsDialog() {
	if ce.currentTheme == nil {
		ce.warn("No theme to save")
		return
	}

	p := ce.palette
	suggestion := ce.suggestThemeName()

	input := tview.NewInputField()
	input.SetLabel(" name ")
	input.SetText(suggestion)
	input.SetFieldWidth(34)
	input.SetLabelColor(p.accent)
	input.SetFieldBackgroundColor(p.selBg)
	input.SetFieldTextColor(p.fg)

	title := " Save theme as "
	if !ce.themeOwned {
		title = " Save a copy (originals are never overwritten) "
	}
	ce.styleBox(input.Box, title)

	input.SetDoneFunc(func(k tcell.Key) {
		ce.closeOverlay("saveas")
		if k == tcell.KeyEnter {
			ce.saveThemeAs(input.GetText())
		}
	})

	ce.showOverlay("saveas", input, 52, 3)
}

// --- theme creator --------------------------------------------------------

// showThemeCreator opens the guided generator: pick a style, a harmony and an
// optional seed colour, then regenerate until something looks right.
func (ce *ColorEditor) showThemeCreator() {
	p := ce.palette
	var generated *GeneratedTheme
	seed := ce.colorValues["normal.blue"]

	styleList := tview.NewList()
	styleList.ShowSecondaryText(false)
	styleList.AddItem("Dark", "", 0, nil)
	styleList.AddItem("Light", "", 0, nil)
	ce.styleList(styleList, " Style ")

	methodList := tview.NewList()
	methodList.ShowSecondaryText(true)
	methodList.AddItem("Random", "Anything goes", 0, nil)
	for i, name := range HarmonyNames {
		methodList.AddItem(name, HarmonyDescriptions[i], 0, nil)
	}
	ce.styleList(methodList, " Harmony ")

	seedInput := tview.NewInputField()
	seedInput.SetLabel(" seed ")
	seedInput.SetText(seed)
	seedInput.SetLabelColor(p.accent)
	seedInput.SetFieldBackgroundColor(p.selBg)
	seedInput.SetFieldTextColor(p.fg)
	ce.styleBox(seedInput.Box, " Base colour (blank = random) ")

	previewPanel := tview.NewTextView()
	previewPanel.SetDynamicColors(true)
	previewPanel.SetTextColor(p.fg)
	ce.styleBox(previewPanel.Box, " Preview ")

	hint := tview.NewTextView()
	hint.SetDynamicColors(true)
	hint.SetBackgroundColor(p.bg)
	hint.SetText(fmt.Sprintf("[%s]Tab move · ↑↓ choose · g regenerate · s keep it · Esc cancel[-]", p.mutedHex))

	generate := func() {
		style := ThemeStyleDark
		if styleList.GetCurrentItem() == 1 {
			style = ThemeStyleLight
		}

		baseHue, hasSeed := seedHue(seedInput.GetText())
		methodIdx := methodList.GetCurrentItem()
		if methodIdx == 0 {
			generated = GenerateRandomTheme(style, baseHue, hasSeed)
		} else {
			generated = GenerateHarmoniousTheme(style, HarmonyType(methodIdx-1), baseHue, hasSeed)
		}
		previewPanel.SetText(ce.renderGeneratedPreview(generated))
		previewPanel.SetTitle(fmt.Sprintf(" Preview · %s ", generated.Name))
	}

	keep := func() {
		if generated == nil {
			return
		}
		ce.closeOverlay("creator")
		ce.ApplyGeneratedTheme(generated)
		ce.showSaveAsDialog()
	}

	boxes := []*tview.Box{styleList.Box, methodList.Box, seedInput.Box}
	targets := []tview.Primitive{styleList, methodList, seedInput}
	focusIdx := 0

	capture := func(self int) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab:
				focusIdx = (self + 1) % len(targets)
				ce.focusRing(boxes, targets, focusIdx)
				return nil
			case tcell.KeyBacktab:
				focusIdx = (self - 1 + len(targets)) % len(targets)
				ce.focusRing(boxes, targets, focusIdx)
				return nil
			case tcell.KeyEscape:
				ce.closeOverlay("creator")
				return nil
			case tcell.KeyEnter:
				if self == 2 {
					generate()
					return nil
				}
				generate()
				return nil
			case tcell.KeyRune:
				// The seed field must keep accepting typed characters.
				if self == 2 {
					return event
				}
				switch event.Rune() {
				case 'g', 'G':
					generate()
					return nil
				case 's', 'S':
					keep()
					return nil
				case 'q', 'Q':
					ce.closeOverlay("creator")
					return nil
				}
			}
			return event
		}
	}

	styleList.SetInputCapture(capture(0))
	methodList.SetInputCapture(capture(1))
	seedInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			generate()
			return nil
		}
		return capture(2)(event)
	})
	styleList.SetChangedFunc(func(int, string, string, rune) { generate() })
	methodList.SetChangedFunc(func(int, string, string, rune) { generate() })

	left := tview.NewFlex()
	left.SetDirection(tview.FlexRow)
	left.AddItem(styleList, 4, 0, true)
	left.AddItem(methodList, 0, 1, false)
	left.AddItem(seedInput, 3, 0, false)

	body := tview.NewFlex()
	body.AddItem(left, 46, 0, true)
	body.AddItem(previewPanel, 0, 1, false)

	root := tview.NewFlex()
	root.SetDirection(tview.FlexRow)
	root.AddItem(body, 0, 1, true)
	root.AddItem(hint, 1, 0, false)
	root.SetBackgroundColor(p.bg)

	generate()

	ce.showOverlay("creator", root, 104, 26)
	ce.focusRing(boxes, targets, focusIdx)
}

// renderGeneratedPreview draws the candidate palette using its own colours.
func (ce *ColorEditor) renderGeneratedPreview(gen *GeneratedTheme) string {
	if gen == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[%s::b]%s[-::-]  [%s]%s[-]\n\n", gen.Foreground, gen.Name, ce.palette.mutedHex, gen.StyleName())

	fmt.Fprintf(&b, "[%s]███[-] background %s\n", gen.Background, gen.Background)
	fmt.Fprintf(&b, "[%s]███[-] foreground %s   [%s]%.1f:1 %s[-]\n\n",
		gen.Foreground, gen.Foreground,
		ce.contrastColor(gen.Contrast()), gen.Contrast(), contrastGrade(gen.Contrast()))

	for _, row := range []struct {
		label  string
		colors map[string]string
	}{{"normal", gen.Normal}, {"bright", gen.Bright}} {
		fmt.Fprintf(&b, "[%s]%-7s[-]", ce.palette.mutedHex, row.label)
		for _, name := range BaseColorNames {
			fmt.Fprintf(&b, "[%s]███[-]", row.colors[name])
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "\n[%s]user[-]@[%s]host[-] [%s]~/code[-] [%s]$[-] [%s]git status[-]\n",
		gen.Normal["green"], gen.Normal["green"], gen.Normal["blue"], gen.Foreground, gen.Bright["yellow"])
	fmt.Fprintf(&b, "[%s]modified:[-]   [%s]main.go[-]\n", gen.Normal["red"], gen.Foreground)
	fmt.Fprintf(&b, "[%s]new file:[-]   [%s]palette.go[-]\n\n", gen.Normal["green"], gen.Foreground)
	fmt.Fprintf(&b, "[%s]func[-] [%s]main[-]() {\n", gen.Normal["magenta"], gen.Bright["blue"])
	fmt.Fprintf(&b, "    [%s]fmt[-].[%s]Println[-]([%s]\"hello\"[-])  [%s]// comment[-]\n",
		gen.Normal["cyan"], gen.Bright["blue"], gen.Normal["yellow"], gen.Bright["black"])
	b.WriteString("}\n")

	return b.String()
}

// seedHue reads a hex seed from the creator's input field.
func seedHue(text string) (float64, bool) {
	hex, ok := expandHex(text)
	if !ok {
		return 0, false
	}
	rgb, err := theme.HexToRGB(hex)
	if err != nil {
		return 0, false
	}
	return RGBToHSL(rgb.R, rgb.G, rgb.B).H, true
}

// generateAndApplyRandomTheme generates and previews a theme immediately
func (ce *ColorEditor) generateAndApplyRandomTheme(harmonious bool) {
	style := ThemeStyleDark
	if bg, exists := ce.colorValues["primary.background"]; exists {
		if rgb, err := theme.HexToRGB(normalizeHex(bg, "#000000")); err == nil {
			if RGBToHSL(rgb.R, rgb.G, rgb.B).L > 0.5 {
				style = ThemeStyleLight
			}
		}
	}

	var gen *GeneratedTheme
	if harmonious {
		gen = GenerateHarmoniousTheme(style, HarmonyType(rand.Intn(len(HarmonyNames))), 0, false)
	} else {
		gen = GenerateRandomTheme(style, 0, false)
	}

	ce.ApplyGeneratedTheme(gen)
	ce.info("Generated %s — g for another, S to save it, r to go back", gen.Name)
}

// --- font panel plumbing --------------------------------------------------

// showFontPanel opens the font browser directly.
func (ce *ColorEditor) showFontPanel() {
	ce.showFontTUI()
}

// restoreMainUI closes the font browser overlay.
func (ce *ColorEditor) restoreMainUI() {
	ce.closeOverlay("fonts")
}
