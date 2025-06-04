package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	// Color definitions
	headerColor  = color.New(color.FgCyan, color.Bold)
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow, color.Bold)
	infoColor    = color.New(color.FgBlue)
	dimColor     = color.New(color.FgHiBlack)
	accentColor  = color.New(color.FgMagenta, color.Bold)
	themeColor   = color.New(color.FgCyan)
)

func PrintHeader(text string) {
	border := strings.Repeat("=", len(text)+4)
	headerColor.Println(border)
	headerColor.Printf("  %s  \n", text)
	headerColor.Println(border)
}

func PrintSubHeader(text string) {
	accentColor.Printf("\n» %s\n", text)
	dimColor.Println(strings.Repeat("-", len(text)+2))
}

func PrintSuccess(format string, args ...interface{}) {
	successColor.Print("✓ ")
	fmt.Printf(format+"\n", args...)
}

func PrintError(format string, args ...interface{}) {
	errorColor.Print("✗ ")
	fmt.Printf(format+"\n", args...)
}

func PrintWarning(format string, args ...interface{}) {
	warningColor.Print("⚠ ")
	fmt.Printf(format+"\n", args...)
}

func PrintInfo(format string, args ...interface{}) {
	infoColor.Print("→ ")
	fmt.Printf(format+"\n", args...)
}

func PrintTheme(name string, description string) {
	themeColor.Printf("  %-25s", name)
	if description != "" {
		dimColor.Printf(" - %s", description)
	}
	fmt.Println()
}

func PrintThemeGrid(themes []string, columns int) {
	for i, theme := range themes {
		if i%columns == 0 && i > 0 {
			fmt.Println()
		}
		themeColor.Printf("  %-25s", theme)
	}
	if len(themes) > 0 {
		fmt.Println()
	}
}

func PrintColorPreview(colorName, hexValue string) {
	// Simple color preview without RGB background colors
	// Use different text colors to indicate the color type
	var colorFunc *color.Color

	switch strings.ToLower(colorName) {
	case "red", "bright_red":
		colorFunc = color.New(color.FgRed, color.Bold)
	case "green", "bright_green":
		colorFunc = color.New(color.FgGreen, color.Bold)
	case "yellow", "bright_yellow":
		colorFunc = color.New(color.FgYellow, color.Bold)
	case "blue", "bright_blue":
		colorFunc = color.New(color.FgBlue, color.Bold)
	case "magenta", "bright_magenta":
		colorFunc = color.New(color.FgMagenta, color.Bold)
	case "cyan", "bright_cyan":
		colorFunc = color.New(color.FgCyan, color.Bold)
	case "white", "bright_white":
		colorFunc = color.New(color.FgWhite, color.Bold)
	case "black", "bright_black":
		colorFunc = color.New(color.FgHiBlack, color.Bold)
	default:
		colorFunc = color.New(color.FgWhite)
	}

	colorFunc.Printf("  %-12s", colorName)
	dimColor.Printf(" %s", hexValue)
	fmt.Println()
}

func ColorizeHeader(text string) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	for _, line := range lines {
		if strings.Contains(line, "Alacritty Colors") {
			result.WriteString(headerColor.Sprint(line))
		} else if strings.HasPrefix(line, "•") {
			result.WriteString(successColor.Sprint("•"))
			result.WriteString(line[1:])
		} else if strings.HasPrefix(line, "Features:") {
			result.WriteString(accentColor.Sprint(line))
		} else {
			result.WriteString(line)
		}
		result.WriteString("\n")
	}

	return result.String()
}

func PrintProgress(current, total int, operation string) {
	percentage := float64(current) / float64(total) * 100
	barWidth := 30
	filled := int(float64(barWidth) * float64(current) / float64(total))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r%s [%s] %d/%d (%.1f%%)",
		operation,
		accentColor.Sprint(bar),
		current,
		total,
		percentage)

	if current == total {
		fmt.Println()
	}
}

func PromptConfirm(message string) bool {
	warningColor.Printf("%s [y/N]: ", message)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func PrintBanner() {
	banner := `
    ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄
    ██ ▄▄▄██ ▄▄▄ ██ ▄▄▀██ ▄▄▄▄██ ▄▄▀██ ▄▄▄ ██
    ██ ▄▄▄██ ███ ██ ██ ██ ▄▄▄▄██ ▀▀▄██ ▀▀▀ ██
    ██ ▀▀▀██ ▀▀▀ ██ ▀▀ ██ ▀▀▀▀██ ██ ██ ▀▀▀ ██
    ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
        A L A C R I T T Y   C O L O R S
    `
	headerColor.Println(banner)
}
