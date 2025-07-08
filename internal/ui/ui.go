package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	headerColor  = color.New(color.FgCyan, color.Bold)
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow)
	infoColor    = color.New(color.FgWhite)
	dimColor     = color.New(color.FgHiBlack)
)

func init() {
	// Disable colors if not supported or requested
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}

func PrintHeader(text string) {
	headerColor.Printf("▌%s\n", text)
	dimColor.Println("  " + strings.Repeat("─", len(text)))
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
	infoColor.Printf(format+"\n", args...)
}

func PrintProgress(current, total int, operation string) {
	percentage := float64(current) / float64(total) * 100
	barWidth := 25
	filled := int(float64(barWidth) * float64(current) / float64(total))

	var bar strings.Builder
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	fmt.Printf("\r%s [%s] %d/%d (%.1f%%)", operation, bar.String(), current, total, percentage)

	if current == total {
		fmt.Println()
	}
}