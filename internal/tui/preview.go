package tui

import (
	"fmt"
	"strings"
)

// PreviewColors holds colors for preview generation
type PreviewColors struct {
	Blue    string
	Cyan    string
	Yellow  string
	White   string
	Magenta string
	Green   string
	Red     string
	Black   string
	Fg      string
	Bg      string
}

// getPreviewColors extracts colors based on current color mode.
//
// Hex mode paints with the exact values being edited. Terminal mode uses the
// named ANSI slots, which Alacritty resolves from the live config — the two
// agree once a change has been pushed, which makes the pair a useful check
// that what you edited is what the terminal actually got.
func (ce *ColorEditor) getPreviewColors() PreviewColors {
	group := "normal."
	switch ce.colorMode {
	case ColorModeBright:
		group = "bright."
	case ColorModeNamed:
		return PreviewColors{
			Blue: "blue", Cyan: "cyan", Yellow: "yellow", White: "white",
			Magenta: "magenta", Green: "green", Red: "red", Black: "black",
			Fg: "white", Bg: "black",
		}
	}

	pick := func(name string) string {
		return ce.getColorOrFallback(group+name, ce.getColorOrFallback("normal."+name, DefaultColors[name]))
	}

	return PreviewColors{
		Blue:    pick("blue"),
		Cyan:    pick("cyan"),
		Yellow:  pick("yellow"),
		White:   pick("white"),
		Magenta: pick("magenta"),
		Green:   pick("green"),
		Red:     pick("red"),
		Black:   pick("black"),
		Fg:      ce.getColorOrFallback("primary.foreground", ce.palette.fgHex),
		Bg:      ce.getColorOrFallback("primary.background", ce.palette.bgHex),
	}
}

// getColorOrFallback returns the color value for a key or a fallback
func (ce *ColorEditor) getColorOrFallback(colorKey, fallback string) string {
	if color, exists := ce.colorValues[colorKey]; exists && color != "" {
		return normalizeHex(color, fallback)
	}
	return fallback
}

// updatePreview refreshes the preview panel
func (ce *ColorEditor) updatePreview() {
	if len(ce.colorValues) == 0 {
		return
	}
	ce.previewPanel.SetText(ce.generatePreview())
	ce.previewPanel.SetTitle(fmt.Sprintf(" Preview · %s ", ColorModeNames[ce.colorMode]))
}

// generatePreview creates the terminal preview content
func (ce *ColorEditor) generatePreview() string {
	var b strings.Builder
	c := ce.getPreviewColors()

	ce.writeColorRamp(&b)
	ce.writeShellSession(&b, c)
	ce.writeFileListing(&b, c)
	ce.writeGitStatus(&b, c)
	ce.writeCodePreview(&b, c)
	ce.writeReadability(&b, c)

	return b.String()
}

// writeColorRamp shows the sixteen ANSI slots as two aligned rows.
func (ce *ColorEditor) writeColorRamp(b *strings.Builder) {
	for _, group := range []string{"normal", "bright"} {
		fmt.Fprintf(b, "[%s]%-7s[-]", ce.palette.mutedHex, group)
		for _, name := range BaseColorNames {
			if value, exists := ce.colorValues[group+"."+name]; exists {
				fmt.Fprintf(b, "[%s]███[-]", normalizeHex(value, "#000000"))
			} else {
				b.WriteString("   ")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// writeShellSession writes the shell session preview
func (ce *ColorEditor) writeShellSession(b *strings.Builder, c PreviewColors) {
	fmt.Fprintf(b, "[%s]user[-]@[%s]host[-] [%s]~/projects[-] [%s]$[-] [%s]ls -la[-]\n",
		c.Green, c.Green, c.Blue, c.Fg, c.Yellow)
}

// writeFileListing writes the file listing preview
func (ce *ColorEditor) writeFileListing(b *strings.Builder, c PreviewColors) {
	rows := []struct {
		mode, name, color string
	}{
		{"drwxr-xr-x", "src/", c.Blue},
		{"-rwxr-xr-x", "alacritty-colors", c.Green},
		{"-rw-r--r--", "README.md", c.Fg},
		{"-rw-r--r--", ".gitignore", c.Cyan},
		{"-rw-r--r--", "backup.tar.gz", c.Magenta},
		{"lrwxrwxrwx", "broken -> missing", c.Red},
	}
	for _, row := range rows {
		fmt.Fprintf(b, "[%s]%s[-] [%s]user staff[-] [%s]%s[-]\n",
			c.Black, row.mode, c.Yellow, row.color, row.name)
	}
	b.WriteString("\n")
}

// writeGitStatus writes the git status preview
func (ce *ColorEditor) writeGitStatus(b *strings.Builder, c PreviewColors) {
	fmt.Fprintf(b, "[%s]On branch[-] [%s]main[-]\n", c.Fg, c.Green)
	fmt.Fprintf(b, "  [%s]modified:   internal/tui/panels.go[-]\n", c.Green)
	fmt.Fprintf(b, "  [%s]deleted:    internal/tui/legacy.go[-]\n", c.Red)
	fmt.Fprintf(b, "  [%s]untracked:  notes.md[-]\n\n", c.Yellow)
}

// writeCodePreview writes the code syntax highlighting preview
func (ce *ColorEditor) writeCodePreview(b *strings.Builder, c PreviewColors) {
	fmt.Fprintf(b, "[%s]package[-] [%s]tui[-]\n", c.Magenta, c.Fg)
	fmt.Fprintf(b, "[%s]func[-] [%s]Render[-]([%s]n[-] [%s]int[-]) [%s]error[-] {\n",
		c.Magenta, c.Blue, c.Fg, c.Cyan, c.Cyan)
	fmt.Fprintf(b, "    [%s]fmt[-].[%s]Println[-]([%s]\"hello\"[-], [%s]42[-])\n",
		c.Cyan, c.Blue, c.Yellow, c.Magenta)
	fmt.Fprintf(b, "    [%s]// a comment sits here[-]\n", c.Black)
	fmt.Fprintf(b, "    [%s]return[-] [%s]nil[-]\n}\n\n", c.Magenta, c.Red)
}

// writeReadability grades every ANSI colour against the background. This is the
// part that turns "looks nice" into "is actually usable at 2am".
func (ce *ColorEditor) writeReadability(b *strings.Builder, c PreviewColors) {
	bg := ce.getColorOrFallback("primary.background", ce.palette.bgHex)

	fmt.Fprintf(b, "[%s]readability vs background[-]\n", ce.palette.mutedHex)

	worst, worstName := 21.0, ""
	for _, name := range BaseColorNames {
		value, exists := ce.colorValues["normal."+name]
		if !exists {
			continue
		}
		ratio := contrastHex(normalizeHex(value, "#000000"), bg)
		if ratio < worst {
			worst, worstName = ratio, name
		}
	}

	fg := ce.getColorOrFallback("primary.foreground", ce.palette.fgHex)
	fgRatio := contrastHex(fg, bg)
	fmt.Fprintf(b, "  text     [%s]%4.1f:1 %s[-]\n", ce.contrastColor(fgRatio), fgRatio, contrastGrade(fgRatio))
	if worstName != "" {
		fmt.Fprintf(b, "  weakest  [%s]%4.1f:1 %s[-]  [%s]%s[-]\n",
			ce.contrastColor(worst), worst, contrastGrade(worst), ce.palette.mutedHex, worstName)
	}

	// Say so when the numbers above cannot hold. A translucent window blends the
	// background with the desktop, so the real ratios are lower than these and
	// the background swatch is a colour the terminal never actually shows.
	if ce.windowOpacity < 1 {
		fmt.Fprintf(b, "\n[%s]  window opacity %.2f — the real background is blended\n"+
			"  with the desktop, so %s is never shown as-is\n"+
			"  and these ratios are a best case[-]\n",
			ce.palette.warnHex, ce.windowOpacity, bg)
	}
}

// cycleColorMode advances through color preview modes
func (ce *ColorEditor) cycleColorMode() {
	ce.colorMode = (ce.colorMode + 1) % ColorModeCount
	ce.updatePreview()
	ce.info("Preview: %s", ColorModeNames[ce.colorMode])
}
