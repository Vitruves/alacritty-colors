package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

// setupUI initializes the main UI layout
func (ce *ColorEditor) setupUI() {
	ce.refreshPalette()

	// Theme list on the left
	ce.themeList = tview.NewList()
	ce.themeList.ShowSecondaryText(false)
	ce.themeList.SetHighlightFullLine(true)
	ce.themeList.SetBorder(true)
	ce.themeList.SetInputCapture(ce.handleThemeListKeys)

	// Color editing panel in the center
	ce.colorPanel = tview.NewList()
	ce.colorPanel.ShowSecondaryText(false)
	ce.colorPanel.SetHighlightFullLine(true)
	ce.colorPanel.SetBorder(true)
	ce.colorPanel.SetTitle(" Palette ")
	ce.colorPanel.SetInputCapture(ce.handleColorPanelKeys)

	// Preview panel on the right
	ce.previewPanel = tview.NewTextView()
	ce.previewPanel.SetDynamicColors(true)
	ce.previewPanel.SetWordWrap(false)
	ce.previewPanel.SetBorder(true)
	ce.previewPanel.SetTitle(" Preview ")
	ce.previewPanel.SetInputCapture(ce.handlePreviewKeys)

	// Filter line, hidden until '/' is pressed
	ce.filterInput = tview.NewInputField()
	ce.filterInput.SetLabel(" / ")
	ce.filterInput.SetChangedFunc(ce.onFilterChanged)
	ce.filterInput.SetInputCapture(ce.handleFilterKeys)

	// Two-line status: what you are looking at, then what you can press.
	ce.statusInfo = tview.NewTextView()
	ce.statusInfo.SetDynamicColors(true)
	ce.statusKeys = tview.NewTextView()
	ce.statusKeys.SetDynamicColors(true)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(ce.themeList, 0, 3, false)
	mainFlex.AddItem(ce.colorPanel, 0, 4, false)
	mainFlex.AddItem(ce.previewPanel, 0, 5, false)

	ce.rootFlex = tview.NewFlex()
	ce.rootFlex.SetDirection(tview.FlexRow)
	ce.rootFlex.AddItem(mainFlex, 0, 1, true)
	ce.rootFlex.AddItem(ce.filterInput, 0, 0, false)
	ce.rootFlex.AddItem(ce.statusInfo, 1, 0, false)
	ce.rootFlex.AddItem(ce.statusKeys, 1, 0, false)

	ce.pages = tview.NewPages()
	ce.pages.AddPage("main", ce.rootFlex, true, true)

	ce.applyPaletteToWidgets()
	ce.app.SetRoot(ce.pages, true)
	ce.setFocus(FocusThemeList)
}

// refreshPalette re-derives the chrome palette from the applied theme.
func (ce *ColorEditor) refreshPalette() {
	parser := alacritty.NewParser()
	cfg, err := parser.ParseFile(ce.config.ConfigFile)
	if err != nil || cfg.Colors.Primary.Background == "" {
		if c, e := parser.ParseFile(ce.config.GetThemePath("current")); e == nil {
			cfg = c
			err = nil
		}
	}

	if err != nil || cfg == nil || cfg.Colors.Primary.Background == "" {
		if !ce.palette.initialized {
			ce.palette = defaultPalette()
		}
	} else {
		ce.palette = paletteFromConfig(cfg)
	}
	ce.palette.applyPaletteToStyles()
}

// applyPaletteToWidgets restyles every long-lived widget from the palette.
func (ce *ColorEditor) applyPaletteToWidgets() {
	p := ce.palette

	for _, list := range []*tview.List{ce.themeList, ce.colorPanel} {
		if list == nil {
			continue
		}
		list.SetBackgroundColor(p.bg)
		list.SetMainTextColor(p.fg)
		list.SetSelectedTextColor(p.selFg)
		list.SetSelectedBackgroundColor(p.selBg)
		list.SetBorderColor(p.border)
		list.SetTitleColor(p.muted)
	}

	if ce.previewPanel != nil {
		ce.previewPanel.SetBackgroundColor(p.bg)
		ce.previewPanel.SetTextColor(p.fg)
		ce.previewPanel.SetBorderColor(p.border)
		ce.previewPanel.SetTitleColor(p.muted)
	}
	if ce.filterInput != nil {
		ce.filterInput.SetBackgroundColor(p.bg)
		ce.filterInput.SetLabelColor(p.accent)
		ce.filterInput.SetFieldBackgroundColor(p.bg)
		ce.filterInput.SetFieldTextColor(p.fg)
	}
	for _, tv := range []*tview.TextView{ce.statusInfo, ce.statusKeys} {
		if tv != nil {
			tv.SetBackgroundColor(p.bg)
		}
	}

	ce.highlightFocus()
}

// --- theme list -----------------------------------------------------------

// loadThemes reads the themes directory and populates the list.
func (ce *ColorEditor) loadThemes() {
	themeFiles, err := ce.getThemeFiles()
	if err != nil {
		ce.setStatus("Error loading themes: "+err.Error(), ce.palette.dangerHex)
		return
	}

	sort.Strings(themeFiles)
	ce.allThemes = themeFiles
	ce.rebuildThemeList()

	// Land on the applied theme if it is still there. When nothing on disk can
	// be identified as live — a fresh install, or an abandoned live edit — the
	// terminal is showing colours no panel here describes, so push the landing
	// theme instead of merely displaying it. Everything on screen then matches
	// the window it is drawn in.
	if idx := ce.indexOfVisible(ce.appliedTheme); idx >= 0 {
		ce.themeList.SetCurrentItem(idx)
		ce.selectTheme(ce.appliedTheme, false)
	} else if len(ce.visible) > 0 {
		ce.themeList.SetCurrentItem(0)
		ce.selectTheme(ce.visible[0], true)
	}
}

// rebuildThemeList redraws the list from allThemes, honouring filter+favOnly.
// The selection is preserved by name whenever the theme survives the filter.
func (ce *ColorEditor) rebuildThemeList() {
	selected := ce.currentVisibleName()

	ce.visible = ce.visible[:0]
	needle := strings.ToLower(ce.filter)
	for _, name := range ce.allThemes {
		if ce.favOnly && !ce.favorites[name] {
			continue
		}
		if needle != "" && !fuzzyMatch(strings.ToLower(name), needle) {
			continue
		}
		ce.visible = append(ce.visible, name)
	}

	ce.themeList.Clear()
	for _, name := range ce.visible {
		ce.themeList.AddItem(ce.formatThemeItem(name), "", 0, nil)
	}

	ce.themeList.SetTitle(ce.themeListTitle())

	if idx := ce.indexOfVisible(selected); idx >= 0 {
		ce.themeList.SetCurrentItem(idx)
	} else if len(ce.visible) > 0 {
		ce.themeList.SetCurrentItem(0)
	}
}

func (ce *ColorEditor) themeListTitle() string {
	label := "Themes"
	if ce.favOnly {
		label = "Favorites"
	}
	if ce.filter != "" {
		return fmt.Sprintf(" %s · /%s (%d) ", label, ce.filter, len(ce.visible))
	}
	return fmt.Sprintf(" %s (%d) ", label, len(ce.visible))
}

// formatThemeItem renders one row: a favourite gutter, an applied marker, the
// name. The gutter keeps every name left-aligned at the same column.
func (ce *ColorEditor) formatThemeItem(name string) string {
	fav := " "
	if ce.favorites[name] {
		fav = FavoriteMarker
	}
	cur := " "
	nameCell := name
	if name == ce.appliedTheme {
		cur = CurrentThemeMarker
		nameCell = "[::b]" + name + "[::-]"
	}
	return fmt.Sprintf("[%s]%s[-][%s]%s[-] %s", ce.palette.dangerHex, fav, ce.palette.accentHex, cur, nameCell)
}

// currentVisibleName returns the theme under the cursor, or "".
func (ce *ColorEditor) currentVisibleName() string {
	idx := ce.themeList.GetCurrentItem()
	if idx < 0 || idx >= len(ce.visible) {
		return ""
	}
	return ce.visible[idx]
}

func (ce *ColorEditor) indexOfVisible(name string) int {
	if name == "" {
		return -1
	}
	for i, n := range ce.visible {
		if n == name {
			return i
		}
	}
	return -1
}

// fuzzyMatch reports whether every rune of needle appears in order in haystack.
// It makes "ctpmoc" find "catppuccin-mocha" without an exact substring.
func fuzzyMatch(haystack, needle string) bool {
	if strings.Contains(haystack, needle) {
		return true
	}
	i := 0
	for _, r := range haystack {
		if i < len(needle) && rune(needle[i]) == r {
			i++
		}
	}
	return i == len(needle)
}

// getThemeFiles returns a list of theme names from the themes directory
func (ce *ColorEditor) getThemeFiles() ([]string, error) {
	files, err := os.ReadDir(ce.config.ThemesDir)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, file := range files {
		name := file.Name()
		if file.IsDir() || !strings.HasSuffix(name, ".toml") || strings.HasPrefix(name, ".") {
			continue
		}
		if name == "current.toml" {
			continue
		}
		themes = append(themes, strings.TrimSuffix(name, ".toml"))
	}

	return themes, nil
}

// refreshThemeList re-reads the directory and redraws, preserving the cursor.
func (ce *ColorEditor) refreshThemeList() {
	if files, err := ce.getThemeFiles(); err == nil {
		sort.Strings(files)
		ce.allThemes = files
	}
	ce.rebuildThemeList()
}

// selectTheme loads a theme into the editor. When push is true the theme is
// also scheduled for application to Alacritty (debounced).
func (ce *ColorEditor) selectTheme(themeName string, push bool) {
	if themeName == "" {
		return
	}

	ce.themeName = themeName
	ce.loadTheme(themeName)
	ce.buildColorPanel()
	ce.updatePreview()
	ce.colorPanel.SetTitle(ce.colorPanelTitle())
	if ce.firstColorIndex() >= 0 {
		ce.colorPanel.SetCurrentItem(ce.firstColorIndex())
	}
	ce.renderStatus()

	if push {
		ce.scheduleApply(themeName)
	}
}

func (ce *ColorEditor) colorPanelTitle() string {
	if ce.themeName == "" {
		return " Palette "
	}
	if ce.isDirty {
		return fmt.Sprintf(" Palette · %s ● ", ce.themeName)
	}
	return fmt.Sprintf(" Palette · %s ", ce.themeName)
}

// --- color panel ----------------------------------------------------------

// buildColorPanel populates the color panel with the current theme's colors
func (ce *ColorEditor) buildColorPanel() {
	ce.colorPanel.Clear()
	ce.listItemToColorKey = make(map[int]string)
	ce.colorKeyToListItem = make(map[string]int)

	if ce.currentTheme == nil {
		ce.colorPanel.AddItem("  Select a theme to start editing", "", 0, nil)
		return
	}

	sections := map[string][]string{
		"Primary":   {"primary.background", "primary.foreground"},
		"Cursor":    {"cursor.text", "cursor.cursor"},
		"Selection": {"selection.text", "selection.background"},
	}
	for _, color := range BaseColorNames {
		sections["Normal"] = append(sections["Normal"], "normal."+color)
		sections["Bright"] = append(sections["Bright"], "bright."+color)
		sections["Dim"] = append(sections["Dim"], "dim."+color)
	}

	// Every standard slot is listed, including the ones this theme does not
	// define: an unset slot is shown as an empty row you can fill in, which is
	// the only way to add a cursor or selection colour to a theme that lacks one.
	itemIndex := 0
	for _, sectionName := range ColorSectionOrder {
		ce.colorPanel.AddItem(fmt.Sprintf("[%s::b]▎%s[-::-]", ce.palette.accentHex, sectionName), "", 0, nil)
		itemIndex++

		for _, key := range sections[sectionName] {
			ce.listItemToColorKey[itemIndex] = key
			ce.colorKeyToListItem[key] = itemIndex
			ce.colorPanel.AddItem(ce.formatColorItem(key), "", 0, nil)
			itemIndex++
		}
	}
}

// formatColorItem renders one palette row: swatch, short name, hex, contrast.
func (ce *ColorEditor) formatColorItem(key string) string {
	label := shortColorLabel(key)

	value, defined := ce.colorValues[key]
	if !defined || value == "" {
		return fmt.Sprintf(" [%s]░░░ %-*s not set[-]", ce.palette.mutedHex, ColorNameWidth, label)
	}

	hex := normalizeHex(value, "#000000")
	row := fmt.Sprintf(" [%s]███[-] %-*s [%s]%s[-]",
		hex, ColorNameWidth, label, ce.palette.mutedHex, hex)

	// Contrast against the theme background is the number that actually decides
	// whether a palette is usable, so it lives on every row.
	if key != "primary.background" {
		bg := normalizeHex(ce.colorValues["primary.background"], ce.palette.bgHex)
		ratio := contrastHex(hex, bg)
		row += fmt.Sprintf("  [%s]%4.1f:1[-]", ce.contrastColor(ratio), ratio)
	}
	return row
}

// derivedDefault proposes a value for a slot the theme leaves empty, so filling
// one in starts from something sensible instead of black.
func (ce *ColorEditor) derivedDefault(key string) string {
	bg := ce.getColorOrFallback("primary.background", ce.palette.bgHex)
	fg := ce.getColorOrFallback("primary.foreground", ce.palette.fgHex)

	switch key {
	case "primary.background":
		return bg
	case "primary.foreground", "cursor.cursor", "selection.text":
		return fg
	case "cursor.text":
		return bg
	case "selection.background":
		return blendHex(bg, ce.getColorOrFallback("normal.blue", ce.palette.accentHex), 0.38)
	}

	group, name, found := strings.Cut(key, ".")
	if !found {
		return fg
	}
	switch group {
	case "dim":
		// Dim is a darkened normal by convention.
		return blendHex(ce.getColorOrFallback("normal."+name, DefaultColors[name]), bg, 0.35)
	case "bright":
		return blendHex(ce.getColorOrFallback("normal."+name, DefaultColors[name]), fg, 0.30)
	default:
		return ce.getColorOrFallback("bright."+name, DefaultColors[name])
	}
}

func (ce *ColorEditor) contrastColor(ratio float64) string {
	switch {
	case ratio >= ContrastAA:
		return ce.palette.okHex
	case ratio >= ContrastLow:
		return ce.palette.warnHex
	default:
		return ce.palette.dangerHex
	}
}

// shortColorLabel strips the section prefix, which the header already shows.
func shortColorLabel(key string) string {
	if i := strings.IndexByte(key, '.'); i >= 0 {
		return key[i+1:]
	}
	return key
}

// firstColorIndex returns the first selectable (non-header) row.
func (ce *ColorEditor) firstColorIndex() int {
	for i := 0; i < ce.colorPanel.GetItemCount(); i++ {
		if _, ok := ce.listItemToColorKey[i]; ok {
			return i
		}
	}
	return -1
}

// moveColorSelection steps the cursor by delta rows, skipping section headers.
func (ce *ColorEditor) moveColorSelection(delta int) {
	count := ce.colorPanel.GetItemCount()
	if count == 0 {
		return
	}
	idx := ce.colorPanel.GetCurrentItem()
	for i := 0; i < count; i++ {
		idx += delta
		if idx < 0 {
			idx = count - 1
		}
		if idx >= count {
			idx = 0
		}
		if _, ok := ce.listItemToColorKey[idx]; ok {
			ce.colorPanel.SetCurrentItem(idx)
			ce.renderStatus()
			return
		}
	}
}

// selectedColorKey returns the colour key under the cursor, or "".
func (ce *ColorEditor) selectedColorKey() string {
	return ce.listItemToColorKey[ce.colorPanel.GetCurrentItem()]
}

// --- focus ----------------------------------------------------------------

func (ce *ColorEditor) setFocus(target int) {
	ce.focus = ((target % FocusCount) + FocusCount) % FocusCount
	switch ce.focus {
	case FocusThemeList:
		ce.app.SetFocus(ce.themeList)
	case FocusColorPanel:
		ce.app.SetFocus(ce.colorPanel)
	case FocusPreview:
		ce.app.SetFocus(ce.previewPanel)
	}
	ce.highlightFocus()
	ce.renderStatus()
}

// highlightFocus marks the active panel with an accent border and title.
func (ce *ColorEditor) highlightFocus() {
	if ce.themeList == nil {
		return
	}
	p := ce.palette

	type styled struct {
		border func(tcell.Color)
		title  func(tcell.Color)
	}
	panels := []styled{
		{func(c tcell.Color) { ce.themeList.SetBorderColor(c) }, func(c tcell.Color) { ce.themeList.SetTitleColor(c) }},
		{func(c tcell.Color) { ce.colorPanel.SetBorderColor(c) }, func(c tcell.Color) { ce.colorPanel.SetTitleColor(c) }},
		{func(c tcell.Color) { ce.previewPanel.SetBorderColor(c) }, func(c tcell.Color) { ce.previewPanel.SetTitleColor(c) }},
	}
	for i, panel := range panels {
		if i == ce.focus {
			panel.border(p.accent)
			panel.title(p.accent)
		} else {
			panel.border(p.border)
			panel.title(p.muted)
		}
	}
}

// --- status bar -----------------------------------------------------------

// setStatus shows a transient message on the info line.
func (ce *ColorEditor) setStatus(message, colorHex string) {
	if colorHex == "" {
		colorHex = ce.palette.fgHex
	}
	if ce.statusInfo == nil {
		return
	}
	ce.statusInfo.SetText(fmt.Sprintf("[%s]%s[-]", colorHex, message))
	ce.statusKeys.SetText(fmt.Sprintf("[%s]%s[-]", ce.palette.mutedHex, FocusKeyHints[ce.focus]))
}

// info/warn/fail are shorthands for the three message tones.
func (ce *ColorEditor) info(format string, a ...interface{}) {
	ce.setStatus(fmt.Sprintf(format, a...), ce.palette.accent2Hex)
}

func (ce *ColorEditor) warn(format string, a ...interface{}) {
	ce.setStatus(fmt.Sprintf(format, a...), ce.palette.warnHex)
}

func (ce *ColorEditor) fail(format string, a ...interface{}) {
	ce.setStatus(fmt.Sprintf(format, a...), ce.palette.dangerHex)
}

// renderStatus rebuilds the info line from the current selection.
func (ce *ColorEditor) renderStatus() {
	if ce.statusInfo == nil {
		return
	}
	p := ce.palette
	var b strings.Builder

	name := ce.themeName
	if name == "" {
		name = "no theme"
	}
	fmt.Fprintf(&b, "[%s::b]%s[-::-]", p.accentHex, name)
	if ce.isDirty {
		fmt.Fprintf(&b, " [%s]● unsaved[-]", p.warnHex)
	}
	if ce.favorites[ce.themeName] {
		fmt.Fprintf(&b, " [%s]♥[-]", p.dangerHex)
	}

	if key := ce.selectedColorKey(); key != "" {
		label := strings.ReplaceAll(key, ".", " ")
		if value, defined := ce.colorValues[key]; defined && value != "" {
			bg := normalizeHex(ce.colorValues["primary.background"], p.bgHex)
			readout := describeColor(normalizeHex(value, "#000000"), bg, key != "primary.background")
			fmt.Fprintf(&b, "  [%s]│[-]  [%s]%s[-]  %s", p.mutedHex, p.fgHex, label, readout)
		} else {
			fmt.Fprintf(&b, "  [%s]│[-]  [%s]%s[-]  [%s]not set — Enter or ←→ to add it[-]",
				p.mutedHex, p.fgHex, label, p.mutedHex)
		}
	} else {
		fmt.Fprintf(&b, "  [%s]│[-]  [%s]%d themes · preview: %s[-]",
			p.mutedHex, p.mutedHex, len(ce.visible), ColorModeNames[ce.colorMode])
	}

	ce.statusInfo.SetText(b.String())
	ce.statusKeys.SetText(fmt.Sprintf("[%s]%s[-]", p.mutedHex, FocusKeyHints[ce.focus]))
}
