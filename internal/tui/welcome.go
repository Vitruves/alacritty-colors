package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// The opening screen. It exists because the editor's reach is not obvious from
// the main view: the theme list is visible, but the palette editor, the font
// browser, the generator and the curated collection all live behind single
// keys. One screen, read once, and the whole surface is known.
//
// It is drawn from the palette that is already live, and shows that palette's
// own ANSI ramp, so the first thing on screen is the theme you are about to
// work on rather than a colour scheme borrowed from somewhere else.

// welcomeFeatures is what the editor can do, in the order someone new is most
// likely to want it. The keys are the real bindings; a wrong hint here would be
// worse than no hint at all.
var welcomeFeatures = []struct {
	key, title, detail string
}{
	{"↑ ↓", "Browse themes", "goes live as the cursor rests on it"},
	{"Tab", "Move to the palette", "then walk the sixteen ANSI slots"},
	{"← →", "Tune a colour", "brightness · Shift hue · +− saturation · # hex"},
	{"f", "Fonts", "family, style and size, with a live sample"},
	{"n", "Create a theme", "from a harmony rule and a base hue"},
	{"g", "Surprise me", "random, but red still comes out red"},
	{"/", "Search", "fuzzy, so ctpmoc finds catppuccin-mocha"},
	{"* F", "Favourites", "star a theme, then list only the starred"},
	{"s S", "Save", "in place, or under a new name"},
	{"p", "Parameters", "the collection, backups, redownload"},
	{"?", "All keys", "the full list, any time"},
}

// padRunes left-aligns s in a field width runes wide.
//
// fmt's %-4s counts bytes, not runes, so a key like "↑ ↓" — three runes but
// seven bytes — would come out unpadded and drag the whole row left. Only the
// arrow rows would misalign, which is exactly the kind of defect that survives
// a casual look at the screen.
func padRunes(s string, width int) string {
	if pad := width - len([]rune(s)); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}

// welcomeBlockWidth is the width of the feature block. The block is laid out
// left-aligned at this width and then centred as a unit, which keeps the key
// column straight instead of ragged the way centring each line would.
const welcomeBlockWidth = 78

// showWelcome fills the screen with the introduction until a key dismisses it.
func (ce *ColorEditor) showWelcome() {
	root := ce.buildWelcome()

	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Any key gets you in; Enter is merely the one advertised.
		_ = event
		ce.pages.ShowPage("main")
		ce.closeOverlay("welcome")
		ce.renderStatus()
		return nil
	})

	ce.overlays = append(ce.overlays, "welcome")
	ce.pages.AddPage("welcome", root, true, true)

	// Hide the editor rather than trusting this page to cover it.
	//
	// tview.Pages draws every visible page in order, so the editor underneath is
	// still being rendered; whether it stays hidden then depends on this page
	// painting every single cell. It did not — the theme list and the ragged
	// right-hand ends of the preview text showed through the margins, and the
	// two screens interleaved into something unreadable. Hiding the page below
	// removes the question entirely: there is nothing there to show through.
	ce.pages.HidePage("main")

	ce.app.SetFocus(root)
}

// buildWelcome lays out the opening screen. It is separate from showWelcome so
// a test can draw it and check that it really does paint every cell.
func (ce *ColorEditor) buildWelcome() *tview.Flex {
	p := ce.palette

	header := tview.NewTextView()
	header.SetDynamicColors(true)
	header.SetTextAlign(tview.AlignCenter)
	header.SetBackgroundColor(p.bg)
	header.SetText(ce.welcomeHeader())

	features := tview.NewTextView()
	features.SetDynamicColors(true)
	features.SetBackgroundColor(p.bg)
	features.SetText(ce.welcomeBody())

	footer := tview.NewTextView()
	footer.SetDynamicColors(true)
	footer.SetTextAlign(tview.AlignCenter)
	footer.SetBackgroundColor(p.bg)
	footer.SetText(ce.welcomeFooter())

	// Spacers are painted boxes, never nil.
	//
	// tview.Pages keeps the pages underneath visible and draws this one on top,
	// and a nil flex item paints nothing at all. Left as nil, the editor behind
	// showed straight through the gaps and the two screens interleaved into an
	// unreadable mess. Every cell of this page has to be written to.
	filler := func() *tview.Box {
		box := tview.NewBox()
		box.SetBackgroundColor(p.bg)
		return box
	}

	// The feature block is centred horizontally by flexible spacers rather than
	// by text alignment, so its internal columns stay aligned.
	bodyRow := tview.NewFlex()
	bodyRow.AddItem(filler(), 0, 1, false)
	bodyRow.AddItem(features, welcomeBlockWidth, 0, false)
	bodyRow.AddItem(filler(), 0, 1, false)
	bodyRow.SetBackgroundColor(p.bg)

	root := tview.NewFlex()
	root.SetDirection(tview.FlexRow)
	root.SetBackgroundColor(p.bg)
	// Spacers above and below centre the whole composition vertically at any
	// terminal height, and the weights are uneven so it sits slightly high,
	// which reads better than dead centre.
	root.AddItem(filler(), 0, 2, false)
	root.AddItem(header, welcomeHeaderLines, 0, false)
	root.AddItem(bodyRow, len(welcomeFeatures)+1, 0, true)
	root.AddItem(footer, 3, 0, false)
	root.AddItem(filler(), 0, 3, false)

	return root
}

// welcomeHeaderLines is how many rows welcomeHeader renders.
const welcomeHeaderLines = 7

// welcomeHeader renders the wordmark, the rule and the live ANSI ramp.
func (ce *ColorEditor) welcomeHeader() string {
	p := ce.palette
	var b strings.Builder

	b.WriteString("\n")
	fmt.Fprintf(&b, "[%s::b]a l a c r i t t y   c o l o r s[-::-]\n", p.accentHex)
	fmt.Fprintf(&b, "[%s]%s[-]\n", blendHex(p.fgHex, p.bgHex, 0.60), strings.Repeat("─", 34))
	fmt.Fprintf(&b, "[%s]terminal palettes, edited where you use them[-]\n\n", p.mutedHex)

	b.WriteString(ce.welcomeRamp("normal"))
	b.WriteString(ce.welcomeRamp("bright"))

	return b.String()
}

// welcomeRamp draws one ANSI row of the palette that is currently live, or a
// blank line when no theme is loaded yet.
func (ce *ColorEditor) welcomeRamp(group string) string {
	var b strings.Builder
	for _, name := range BaseColorNames {
		value, ok := ce.colorValues[group+"."+name]
		if !ok || value == "" {
			continue
		}
		fmt.Fprintf(&b, "[%s]███[-]", normalizeHex(value, "#000000"))
	}
	b.WriteString("\n")
	return b.String()
}

// welcomeBody renders the feature table.
func (ce *ColorEditor) welcomeBody() string {
	p := ce.palette
	var b strings.Builder

	b.WriteString("\n")
	for _, feature := range welcomeFeatures {
		fmt.Fprintf(&b, "  [%s::b]%s[-::-]  [%s]%s[-] [%s]%s[-]\n",
			p.accentHex, padRunes(feature.key, 4),
			p.fgHex, padRunes(feature.title, 20),
			p.mutedHex, feature.detail)
	}

	return b.String()
}

// welcomeFooter renders the closing promise and the way in.
func (ce *ColorEditor) welcomeFooter() string {
	p := ce.palette
	var b strings.Builder

	fmt.Fprintf(&b, "[%s]Nothing is written until you save · browsing only changes the live terminal[-]\n\n",
		p.mutedHex)
	fmt.Fprintf(&b, "[%s::b]Enter[-::-] [%s]to begin[-]", p.okHex, p.mutedHex)

	return b.String()
}
