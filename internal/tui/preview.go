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
}

// getPreviewColors extracts colors based on current color mode
func (ce *ColorEditor) getPreviewColors() PreviewColors {
	switch ce.colorMode {
	case ColorModeHex:
		return PreviewColors{
			Blue:    ce.getColorOrFallback("normal.blue", DefaultColors["blue"]),
			Cyan:    ce.getColorOrFallback("normal.cyan", DefaultColors["cyan"]),
			Yellow:  ce.getColorOrFallback("normal.yellow", DefaultColors["yellow"]),
			White:   ce.getColorOrFallback("normal.white", DefaultColors["white"]),
			Magenta: ce.getColorOrFallback("normal.magenta", DefaultColors["magenta"]),
			Green:   ce.getColorOrFallback("normal.green", DefaultColors["green"]),
			Red:     ce.getColorOrFallback("normal.red", DefaultColors["red"]),
			Black:   ce.getColorOrFallback("normal.black", DefaultColors["black"]),
		}
	case ColorModeNamed:
		return PreviewColors{
			Blue:    "blue",
			Cyan:    "cyan",
			Yellow:  "yellow",
			White:   "white",
			Magenta: "magenta",
			Green:   "green",
			Red:     "red",
			Black:   "black",
		}
	case ColorModeBright:
		return PreviewColors{
			Blue:    ce.getColorOrFallback("bright.blue", "blue"),
			Cyan:    ce.getColorOrFallback("bright.cyan", "cyan"),
			Yellow:  ce.getColorOrFallback("bright.yellow", "yellow"),
			White:   ce.getColorOrFallback("bright.white", "white"),
			Magenta: ce.getColorOrFallback("bright.magenta", "magenta"),
			Green:   ce.getColorOrFallback("bright.green", "green"),
			Red:     ce.getColorOrFallback("bright.red", "red"),
			Black:   ce.getColorOrFallback("bright.black", "black"),
		}
	default:
		return PreviewColors{
			Blue:    "blue",
			Cyan:    "cyan",
			Yellow:  "yellow",
			White:   "white",
			Magenta: "magenta",
			Green:   "green",
			Red:     "red",
			Black:   "black",
		}
	}
}

// getColorOrFallback returns the color value for a key or a fallback
func (ce *ColorEditor) getColorOrFallback(colorKey, fallback string) string {
	if color, exists := ce.colorValues[colorKey]; exists && color != "" {
		return color
	}
	return fallback
}

// updatePreview refreshes the preview panel
func (ce *ColorEditor) updatePreview() {
	if ce.currentTheme == nil {
		return
	}

	preview := ce.generatePreview()
	ce.previewPanel.SetText(preview)
}

// generatePreview creates the terminal preview content
func (ce *ColorEditor) generatePreview() string {
	var preview strings.Builder
	colors := ce.getPreviewColors()

	preview.WriteString("[yellow::b]Terminal Preview[-]\n")
	preview.WriteString("[white]Colors shown are approximations.[-]\n")
	preview.WriteString("[white]See actual colors in your terminal.[-]\n\n")

	// Color palette display
	ce.writeColorPalette(&preview)

	// Shell session
	ce.writeShellSession(&preview, colors)

	// File listing
	ce.writeFileListing(&preview, colors)

	// Git status
	ce.writeGitStatus(&preview, colors)

	// Code preview
	ce.writeCodePreview(&preview, colors)

	// System info
	ce.writeSystemInfo(&preview, colors)

	return preview.String()
}

// writeColorPalette writes the color palette section
func (ce *ColorEditor) writeColorPalette(preview *strings.Builder) {
	preview.WriteString("[white::b]Normal Colors:[-]\n")
	for _, color := range BaseColorNames {
		if colorVal, exists := ce.colorValues["normal."+color]; exists {
			preview.WriteString(fmt.Sprintf("[%s]███[-] ", colorVal))
		}
	}
	preview.WriteString("\n\n[white::b]Bright Colors:[-]\n")
	for _, color := range BaseColorNames {
		if colorVal, exists := ce.colorValues["bright."+color]; exists {
			preview.WriteString(fmt.Sprintf("[%s]███[-] ", colorVal))
		}
	}
	preview.WriteString("\n\n")
}

// writeShellSession writes the shell session preview
func (ce *ColorEditor) writeShellSession(preview *strings.Builder, colors PreviewColors) {
	preview.WriteString("[white::b]Shell Session:[-]\n")
	preview.WriteString(fmt.Sprintf("[%s]user@hostname[-]", colors.Green))
	preview.WriteString(fmt.Sprintf("[%s]:[-]", colors.Blue))
	preview.WriteString(fmt.Sprintf("[%s]~/projects[-]", colors.Cyan))
	preview.WriteString(fmt.Sprintf("[%s]$ [-]", colors.White))
	preview.WriteString(fmt.Sprintf("[%s]ls -la[-]\n\n", colors.Yellow))
}

// writeFileListing writes the file listing preview
func (ce *ColorEditor) writeFileListing(preview *strings.Builder, colors PreviewColors) {
	// Directory
	preview.WriteString(fmt.Sprintf("[%s]drwxr-xr-x[-] [%s]5[-] [%s]user[-] [%s]staff[-] [%s]160[-] [%s]Jan 15 10:30[-] [%s]src/[-]\n",
		colors.Black, colors.White, colors.Yellow, colors.Cyan, colors.White, colors.White, colors.Blue))

	// Executable
	preview.WriteString(fmt.Sprintf("[%s]-rwxr-xr-x[-] [%s]1[-] [%s]user[-] [%s]staff[-] [%s]8192[-] [%s]Jan 15 10:25[-] [%s]alacritty-colors[-]\n",
		colors.Green, colors.White, colors.Yellow, colors.Cyan, colors.White, colors.White, colors.Green))

	// Regular file
	preview.WriteString(fmt.Sprintf("[%s]-rw-r--r--[-] [%s]1[-] [%s]user[-] [%s]staff[-] [%s]1234[-] [%s]Jan 15 10:20[-] [%s]README.md[-]\n",
		colors.Black, colors.White, colors.Yellow, colors.Cyan, colors.White, colors.White, colors.White))

	// Dotfile
	preview.WriteString(fmt.Sprintf("[%s]-rw-r--r--[-] [%s]1[-] [%s]user[-] [%s]staff[-] [%s]456[-] [%s]Jan 15 10:15[-] [%s].gitignore[-]\n",
		colors.Black, colors.White, colors.Yellow, colors.Cyan, colors.White, colors.White, colors.Cyan))

	// Archive
	preview.WriteString(fmt.Sprintf("[%s]-rw-r--r--[-] [%s]1[-] [%s]user[-] [%s]staff[-] [%s]2048[-] [%s]Jan 15 10:12[-] [%s]backup.tar.gz[-]\n",
		colors.Black, colors.White, colors.Yellow, colors.Cyan, colors.White, colors.White, colors.Magenta))

	// Broken symlink
	preview.WriteString(fmt.Sprintf("[%s]lrwxrwxrwx[-] [%s]1[-] [%s]user[-] [%s]staff[-] [%s]10[-] [%s]Jan 15 10:10[-] [%s]broken[-] -> [%s]missing[-]\n\n",
		colors.Black, colors.White, colors.Yellow, colors.Cyan, colors.White, colors.White, colors.Red, colors.Red))
}

// writeGitStatus writes the git status preview
func (ce *ColorEditor) writeGitStatus(preview *strings.Builder, colors PreviewColors) {
	preview.WriteString("[white::b]Git Status:[-]\n")
	preview.WriteString(fmt.Sprintf("[%s]On branch[-] [%s]main[-]\n", colors.White, colors.Green))
	preview.WriteString(fmt.Sprintf("[%s]Changes to be committed:[-]\n", colors.Green))
	preview.WriteString(fmt.Sprintf("  [%s]modified:   src/main.go[-]\n", colors.Green))
	preview.WriteString(fmt.Sprintf("[%s]Changes not staged:[-]\n", colors.Red))
	preview.WriteString(fmt.Sprintf("  [%s]modified:   README.md[-]\n", colors.Red))
	preview.WriteString(fmt.Sprintf("[%s]Untracked files:[-]\n", colors.Yellow))
	preview.WriteString(fmt.Sprintf("  [%s]new_file.txt[-]\n\n", colors.Red))
}

// writeCodePreview writes the code syntax highlighting preview
func (ce *ColorEditor) writeCodePreview(preview *strings.Builder, colors PreviewColors) {
	preview.WriteString("[white::b]Code Preview:[-]\n")
	preview.WriteString(fmt.Sprintf("[%s]func[-] [%s]main[-][%s]()[-] [%s]{[-]\n",
		colors.Blue, colors.Yellow, colors.White, colors.White))
	preview.WriteString(fmt.Sprintf("    [%s]fmt[-][%s].[-][%s]Println[-][%s]([-][%s]\"Hello, World!\"[-][%s])[-]\n",
		colors.Cyan, colors.White, colors.Yellow, colors.White, colors.Green, colors.White))
	preview.WriteString(fmt.Sprintf("    [%s]// This is a comment[-]\n", colors.Magenta))
	preview.WriteString(fmt.Sprintf("[%s]}[-]\n\n", colors.White))
}

// writeSystemInfo writes the system info preview
func (ce *ColorEditor) writeSystemInfo(preview *strings.Builder, colors PreviewColors) {
	preview.WriteString("[white::b]System Info:[-]\n")
	preview.WriteString(fmt.Sprintf("[%s]CPU:[-] [%s]12.5%%[-] [%s]Memory:[-] [%s]2.1GB/8GB[-]\n",
		colors.Cyan, colors.Green, colors.Cyan, colors.Yellow))
	preview.WriteString(fmt.Sprintf("[%s]Load:[-] [%s]1.23[-] [%s]Uptime:[-] [%s]5 days[-]\n",
		colors.Cyan, colors.Green, colors.Cyan, colors.White))
	preview.WriteString(fmt.Sprintf("[%s]Disk:[-] [%s]45GB[-][%s]/[-][%s]100GB[-] [%s](45%%)[-]\n",
		colors.Cyan, colors.Yellow, colors.White, colors.White, colors.Red))
}

// cycleColorMode advances through color preview modes
func (ce *ColorEditor) cycleColorMode() {
	ce.colorMode = (ce.colorMode + 1) % ColorModeCount
	ce.setStatus(fmt.Sprintf("Color mode: %s (press 'c' to cycle)", ColorModeNames[ce.colorMode]))
	ce.updatePreview()
}
