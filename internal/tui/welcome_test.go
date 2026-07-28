package tui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// newWelcomeFixture builds just enough editor to render the opening screen.
func newWelcomeFixture() *ColorEditor {
	ce := &ColorEditor{
		colorValues: map[string]string{},
		allThemes:   make([]string, 163),
	}
	sample := []string{"#2e2e2e", "#eb4129", "#abe047", "#f6c744", "#47a0f3", "#7b5cb0", "#64dbed", "#e5e9f0"}
	for _, group := range []string{"normal", "bright"} {
		for i, name := range BaseColorNames {
			ce.colorValues[group+"."+name] = sample[i]
		}
	}
	ce.palette = buildPalette(map[string]string{
		"background": "#101011", "foreground": "#fffbf6", "blue": "#47a0f3",
	})
	return ce
}

// stripTags removes tview colour tags so what is left is what is displayed.
func stripTags(s string) string {
	var b strings.Builder
	depth := 0
	for _, r := range s {
		switch {
		case r == '[':
			depth++
		case r == ']' && depth > 0:
			depth--
		case depth == 0:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Every rendered line has to fit the block it is laid out in, or tview wraps it
// and the table turns ragged. Widths are counted in runes: the key column holds
// arrows, and counting bytes would let a line overflow while looking fine here.
func TestWelcomeFitsItsBlock(t *testing.T) {
	ce := newWelcomeFixture()

	for _, line := range strings.Split(stripTags(ce.welcomeBody()), "\n") {
		if width := len([]rune(line)); width > welcomeBlockWidth {
			t.Errorf("feature line is %d columns, block is %d: %q", width, welcomeBlockWidth, line)
		}
	}

	// The header and footer are centred across the full width, so 80 columns is
	// the figure to stay under.
	const narrowTerminal = 80
	for _, section := range []string{ce.welcomeHeader(), ce.welcomeFooter()} {
		for _, line := range strings.Split(stripTags(section), "\n") {
			if width := len([]rune(line)); width > narrowTerminal {
				t.Errorf("line is %d columns, wider than an %d column terminal: %q", width, narrowTerminal, line)
			}
		}
	}
}

// The key column is padded by rune count. With fmt's byte-counting %-4s the
// arrow rows lose their padding and the whole row shifts left.
func TestWelcomeKeyColumnAligns(t *testing.T) {
	ce := newWelcomeFixture()

	var titleColumn = -1
	for _, line := range strings.Split(stripTags(ce.welcomeBody()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		runes := []rune(line)
		// The title starts after the two-space indent, the four-wide key field
		// and its two-space gap.
		const titleStart = 8
		if len(runes) < titleStart {
			t.Fatalf("line is too short to hold the key column: %q", line)
		}
		if runes[titleStart-1] == ' ' && runes[titleStart] == ' ' {
			t.Errorf("key field is over-padded, title column drifts right: %q", line)
		}

		start := len(runes) - len([]rune(strings.TrimLeft(string(runes[titleStart-1:]), " ")))
		if titleColumn == -1 {
			titleColumn = start
			continue
		}
		if start != titleColumn {
			t.Errorf("title starts at column %d, expected %d: %q", start, titleColumn, line)
		}
	}
}

func TestPadRunes(t *testing.T) {
	cases := []struct {
		in    string
		width int
		want  int
	}{
		{"↑ ↓", 4, 4}, // three runes, seven bytes
		{"← →", 4, 4},
		{"Tab", 4, 4},
		{"* F", 4, 4},
		{"already wider", 4, 13},
	}

	for _, tc := range cases {
		if got := len([]rune(padRunes(tc.in, tc.width))); got != tc.want {
			t.Errorf("padRunes(%q, %d) is %d runes wide, want %d", tc.in, tc.width, got, tc.want)
		}
	}
}

// The opening screen is drawn on top of the editor, and tview keeps the pages
// underneath visible. If any cell is left unpainted the theme list and the
// preview show straight through the gaps and both screens become unreadable.
func TestWelcomePaintsEveryCell(t *testing.T) {
	const width, height = 100, 34

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(width, height)

	ce := newWelcomeFixture()

	// Reproduce the real arrangement: the editor is a page underneath and the
	// welcome screen is a page added on top. tview.Pages keeps both visible and
	// draws them in order, which is precisely why an unpainted cell on the top
	// page lets the one underneath show through.
	behind := tview.NewTextView()
	behind.SetText(strings.Repeat("EDITOR UNDERNEATH\n", height))
	behind.SetTextColor(tcell.ColorRed)

	pages := tview.NewPages()
	pages.AddPage("main", behind, true, true)
	pages.AddPage("welcome", ce.buildWelcome(), true, true)
	// Exactly what showWelcome does.
	pages.SetRect(0, 0, width, height)
	pages.Draw(screen)

	cells, gotWidth, gotHeight := screen.GetContents()
	if gotWidth != width || gotHeight != height {
		t.Fatalf("screen is %dx%d, expected %dx%d", gotWidth, gotHeight, width, height)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := cells[y*width+x]
			if foreground, _, _ := cell.Style.Decompose(); foreground == tcell.ColorRed {
				t.Fatalf("cell (%d,%d) still shows the editor underneath: %q", x, y, string(cell.Runes))
			}
		}
	}
}
