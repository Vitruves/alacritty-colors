package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	// Primary color scheme - more sober yet colorful
	headerColor  = color.New(color.FgCyan, color.Bold)
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed, color.Bold)
	warningColor = color.New(color.FgYellow)
	infoColor    = color.New(color.FgWhite)
	dimColor     = color.New(color.FgHiBlack)
	accentColor  = color.New(color.FgMagenta)
	themeColor   = color.New(color.FgCyan)
	numberColor  = color.New(color.FgHiBlue)

	// Enhanced color palette
	primaryColor   = color.New(color.FgWhite)
	secondaryColor = color.New(color.FgHiWhite)
	highlightColor = color.New(color.FgYellow)
	codeColor      = color.New(color.FgGreen)

	// Status colors
	onlineColor  = color.New(color.FgGreen)
	offlineColor = color.New(color.FgRed)
	pendingColor = color.New(color.FgYellow)

	// Specialized colors
	fileColor = color.New(color.FgCyan)
	timeColor = color.New(color.FgMagenta)
	sizeColor = color.New(color.FgYellow)

	// New verbose color
	verboseColor = color.New(color.FgHiBlack)
)

// Terminal capability detection
var (
	supportsUnicode = checkUnicodeSupport()
	supportsColor   = checkColorSupport()
)

func init() {
	// Disable colors if not supported or requested
	if os.Getenv("NO_COLOR") != "" || !supportsColor {
		color.NoColor = true
	}
}

// Header and section functions - made more sober
func PrintHeader(text string) {
	if !supportsUnicode {
		// Fallback for terminals without Unicode support
		border := strings.Repeat("=", len(text)+4)
		headerColor.Println(border)
		headerColor.Printf("  %s  \n", text)
		headerColor.Println(border)
		return
	}

	headerColor.Printf("▌%s\n", text)
	dimColor.Println("  " + strings.Repeat("─", len(text)))
}

func PrintSubHeader(text string) {
	if !supportsUnicode {
		fmt.Printf("\n> %s\n", text)
		return
	}

	accentColor.Printf("\n▶ %s\n", text)
}

func PrintSection(title string) {
	if !supportsUnicode {
		highlightColor.Printf("# %s\n", title)
		return
	}

	highlightColor.Printf("▼ %s\n", title)
}

func PrintSeparator() {
	if !supportsUnicode {
		dimColor.Println("  " + strings.Repeat("-", 40))
		return
	}
	dimColor.Println("  " + strings.Repeat("─", 40))
}

// Status and message functions - more concise
func PrintSuccess(format string, args ...interface{}) {
	symbol := "✓"
	if !supportsUnicode {
		symbol = "OK"
	}
	successColor.Print(symbol + " ")
	primaryColor.Printf(format+"\n", args...)
}

func PrintError(format string, args ...interface{}) {
	symbol := "✗"
	if !supportsUnicode {
		symbol = "ERROR"
	}
	errorColor.Print(symbol + " ")
	primaryColor.Printf(format+"\n", args...)
}

func PrintWarning(format string, args ...interface{}) {
	symbol := "⚠"
	if !supportsUnicode {
		symbol = "WARN"
	}
	warningColor.Print(symbol + " ")
	primaryColor.Printf(format+"\n", args...)
}

func PrintInfo(format string, args ...interface{}) {
	infoColor.Printf(format+"\n", args...)
}

func PrintVerbose(format string, args ...interface{}) {
	symbol := "→"
	if !supportsUnicode {
		symbol = "->"
	}
	verboseColor.Print(symbol + " ")
	verboseColor.Printf(format+"\n", args...)
}

func PrintStep(step int, total int, text string) {
	numberColor.Printf("[%d/%d] ", step, total)
	primaryColor.Println(text)
}

func PrintStatus(status, message string) {
	var statusColor *color.Color
	var symbol string

	switch strings.ToLower(status) {
	case "online", "active", "running", "success":
		statusColor = onlineColor
		symbol = "●"
	case "offline", "inactive", "stopped", "error":
		statusColor = offlineColor
		symbol = "●"
	case "pending", "loading", "processing":
		statusColor = pendingColor
		symbol = "◐"
	default:
		statusColor = dimColor
		symbol = "○"
	}

	if !supportsUnicode {
		symbol = "[" + strings.ToUpper(status) + "]"
	}

	statusColor.Printf("%s ", symbol)
	secondaryColor.Println(message)
}

// Theme and content display functions
func PrintTheme(name string, description string) {
	themeColor.Printf("  %-25s", name)
	if description != "" {
		separator := "│"
		if !supportsUnicode {
			separator = "|"
		}
		dimColor.Printf(" %s %s", separator, description)
	}
	fmt.Println()
}

func PrintThemeGrid(themes []string, columns int) {
	if columns <= 0 {
		columns = 3
	}

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
	// Enhanced color preview with better formatting
	var colorFunc *color.Color
	var swatch string

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
	case "background":
		colorFunc = color.New(color.BgBlack, color.FgWhite)
	case "foreground":
		colorFunc = color.New(color.FgWhite, color.Bold)
	default:
		colorFunc = color.New(color.FgWhite)
	}

	// Create color swatch
	if supportsUnicode {
		swatch = "████"
	} else {
		swatch = "####"
	}

	colorFunc.Printf("  %s", swatch)
	primaryColor.Printf(" %-14s", colorName)
	separator := "│"
	if !supportsUnicode {
		separator = "|"
	}
	dimColor.Printf("%s %s", separator, hexValue)
	fmt.Println()
}

func PrintKeyValue(key, value string) {
	accentColor.Printf("%-15s ", key+":")
	primaryColor.Println(value)
}

func PrintList(items []string) {
	bullet := "•"
	if !supportsUnicode {
		bullet = "*"
	}

	for _, item := range items {
		infoColor.Printf("  %s ", bullet)
		secondaryColor.Println(item)
	}
}

func PrintOrderedList(items []string) {
	for i, item := range items {
		numberColor.Printf("  %d. ", i+1)
		secondaryColor.Println(item)
	}
}

func PrintTree(items map[string][]string) {
	var branch, leaf, lastBranch string

	if supportsUnicode {
		branch = "├── "
		leaf = "│   "
		lastBranch = "└── "
	} else {
		branch = "+-- "
		leaf = "|   "
		lastBranch = "`-- "
	}

	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}

	for i, key := range keys {
		isLast := i == len(keys)-1

		if isLast {
			accentColor.Print(lastBranch)
		} else {
			accentColor.Print(branch)
		}
		primaryColor.Println(key)

		for _, item := range items[key] {
			if isLast {
				dimColor.Print("    ")
			} else {
				dimColor.Print(leaf)
			}
			secondaryColor.Println(item)
		}
	}
}

// Progress and interaction functions
func PrintProgress(current, total int, operation string) {
	percentage := float64(current) / float64(total) * 100
	barWidth := 25
	filled := int(float64(barWidth) * float64(current) / float64(total))

	// Create gradient progress bar
	var bar strings.Builder
	var fillChar, emptyChar string

	if supportsUnicode {
		fillChar = "█"
		emptyChar = "░"
	} else {
		fillChar = "#"
		emptyChar = "-"
	}

	for i := 0; i < barWidth; i++ {
		if i < filled {
			if i < barWidth/3 {
				bar.WriteString(successColor.Sprint(fillChar))
			} else if i < 2*barWidth/3 {
				bar.WriteString(warningColor.Sprint(fillChar))
			} else {
				bar.WriteString(headerColor.Sprint(fillChar))
			}
		} else {
			bar.WriteString(dimColor.Sprint(emptyChar))
		}
	}

	infoColor.Printf("\r%s ", operation)
	fmt.Printf("[%s] ", bar.String())
	numberColor.Printf("%d/%d ", current, total)
	dimColor.Printf("(%.1f%%)", percentage)

	if current == total {
		fmt.Println()
	}
}

func PrintSpinner(message string, delay time.Duration) func() {
	var frames []string
	if supportsUnicode {
		frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	} else {
		frames = []string{"|", "/", "-", "\\"}
	}

	done := make(chan bool)
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r%s %s", accentColor.Sprint(frames[i]), message)
				i = (i + 1) % len(frames)
				time.Sleep(delay)
			}
		}
	}()

	return func() {
		done <- true
		fmt.Print("\r" + strings.Repeat(" ", len(message)+10) + "\r")
	}
}

func PromptConfirm(message string) bool {
	symbol := "?"
	if supportsUnicode {
		symbol = "❓"
	}

	warningColor.Printf("%s %s ", symbol, message)
	dimColor.Print("[y/N]: ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func PromptInput(message string) string {
	symbol := "?"
	if supportsUnicode {
		symbol = "❓"
	}

	infoColor.Printf("%s %s: ", symbol, message)
	var response string
	fmt.Scanln(&response)
	return response
}

func PromptSelect(message string, options []string) int {
	fmt.Println()
	accentColor.Println(message)

	for i, option := range options {
		numberColor.Printf("  %d. ", i+1)
		secondaryColor.Println(option)
	}

	for {
		fmt.Print("\nSelect option (number): ")
		var input string
		fmt.Scanln(&input)

		if choice, err := strconv.Atoi(input); err == nil && choice >= 1 && choice <= len(options) {
			return choice - 1
		}

		errorColor.Printf("Invalid choice. Please enter a number between 1 and %d.\n", len(options))
	}
}

// Layout and formatting functions
func PrintCodeBlock(code string) {
	lines := strings.Split(code, "\n")
	width := 50

	var top, side, bottom string
	if supportsUnicode {
		top = "╭" + strings.Repeat("─", width) + "╮"
		side = "│"
		bottom = "╰" + strings.Repeat("─", width) + "╯"
	} else {
		top = "+" + strings.Repeat("-", width) + "+"
		side = "|"
		bottom = "+" + strings.Repeat("-", width) + "+"
	}

	dimColor.Println("  " + top)
	for _, line := range lines {
		dimColor.Print("  " + side + " ")
		codeColor.Printf("%-*s", width-2, line)
		dimColor.Println(" " + side)
	}
	dimColor.Println("  " + bottom)
}

func PrintBox(title, content string) {
	titleLen := len(title)
	contentLen := len(content)
	width := titleLen + 4
	if contentWidth := contentLen + 4; contentWidth > width {
		width = contentWidth
	}

	var top, middle, bottom string
	if supportsUnicode {
		top = "╭─ " + title + " " + strings.Repeat("─", width-titleLen-4) + "╮"
		middle = fmt.Sprintf("│  %-*s  │", width-4, content)
		bottom = "╰" + strings.Repeat("─", width) + "╯"
	} else {
		top = "+- " + title + " " + strings.Repeat("-", width-titleLen-4) + "+"
		middle = fmt.Sprintf("|  %-*s  |", width-4, content)
		bottom = "+" + strings.Repeat("-", width) + "+"
	}

	accentColor.Println(top)
	primaryColor.Println(middle)
	accentColor.Println(bottom)
}

func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	fmt.Print("  ")
	for i, header := range headers {
		headerColor.Printf("%-*s", colWidths[i]+2, header)
	}
	fmt.Println()

	// Print separator
	fmt.Print("  ")
	for i := range headers {
		dimColor.Print(strings.Repeat("─", colWidths[i]+2))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("  ")
		for i, cell := range row {
			if i < len(colWidths) {
				secondaryColor.Printf("%-*s", colWidths[i]+2, cell)
			}
		}
		fmt.Println()
	}
}

// Banner and branding
func PrintBanner() {
	banner := `
╭─────────────────────────────────────────────────────────────────────╮
│                                                                     │
│  Alacritty Colors                                                   │
│                                                                     │
╰─────────────────────────────────────────────────────────────────────╯`

	if supportsUnicode {
		headerColor.Println(banner)
	} else {
		headerColor.Println("Alacritty Colors")
	}

	fmt.Println()
}

func PrintVersion(version, buildDate, gitCommit string) {
	PrintSection("Version Information")
	PrintKeyValue("Version", version)
	if buildDate != "" {
		PrintKeyValue("Built", buildDate)
	}
	if gitCommit != "" {
		PrintKeyValue("Commit", gitCommit[:8])
	}
}

// Utility and helper functions
func PrintStats(themes, backups int, currentTheme string) {
	PrintSection("Status")
	PrintKeyValue("Available themes", fmt.Sprintf("%d", themes))
	PrintKeyValue("Backups", fmt.Sprintf("%d", backups))
	if currentTheme != "" {
		PrintKeyValue("Current theme", currentTheme)
	} else {
		PrintKeyValue("Current theme", dimColor.Sprint("none"))
	}
}

func PrintFileInfo(filename string, size int64, modTime time.Time) {
	fileColor.Printf("  %s", filename)
	fmt.Print("  ")
	sizeColor.Printf("(%s)", formatSize(size))
	fmt.Print("  ")
	timeColor.Printf("%s", modTime.Format("2006-01-02 15:04"))
	fmt.Println()
}

func ColorizeHeader(text string) string {
	if !supportsColor {
		return text
	}

	// Handle simple header text
	if !strings.Contains(text, "\n") {
		// Single line header - just colorize it
		return headerColor.Sprint(text)
	}

	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle different header types
		if strings.HasPrefix(trimmed, "USAGE") ||
			strings.HasPrefix(trimmed, "COMMANDS") ||
			strings.HasPrefix(trimmed, "OPTIONS") ||
			strings.HasPrefix(trimmed, "GLOBAL OPTIONS") ||
			strings.HasPrefix(trimmed, "EXAMPLES") ||
			strings.HasPrefix(trimmed, "MORE INFO") ||
			strings.HasPrefix(trimmed, "Key Features:") {
			result = append(result, headerColor.Sprint(line))
		} else if strings.HasPrefix(trimmed, "Alacritty Colors") {
			// Title line
			result = append(result, headerColor.Sprint(line))
		} else if strings.HasPrefix(trimmed, "  •") ||
			strings.HasPrefix(trimmed, "  -") {
			// Feature bullets
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				result = append(result, fmt.Sprintf("  %s %s",
					highlightColor.Sprint(parts[1]),
					secondaryColor.Sprint(strings.Join(parts[2:], " "))))
			} else {
				result = append(result, secondaryColor.Sprint(line))
			}
		} else if trimmed == "" {
			result = append(result, line)
		} else {
			result = append(result, infoColor.Sprint(line))
		}
	}

	return strings.Join(result, "\n")
}

// Terminal capability detection
func checkUnicodeSupport() bool {
	// Check common environment variables that indicate Unicode support
	lang := os.Getenv("LANG")
	lcAll := os.Getenv("LC_ALL")
	term := os.Getenv("TERM")

	// Check for UTF-8 in locale
	if strings.Contains(strings.ToUpper(lang), "UTF-8") ||
		strings.Contains(strings.ToUpper(lcAll), "UTF-8") {
		return true
	}

	// Check for modern terminals
	modernTerms := []string{"xterm-256color", "screen-256color", "tmux-256color", "alacritty"}
	for _, modernTerm := range modernTerms {
		if strings.Contains(term, modernTerm) {
			return true
		}
	}

	return false
}

func checkColorSupport() bool {
	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")

	// Check for explicit color support
	if colorTerm != "" {
		return true
	}

	// Check terminal type
	colorTerms := []string{"color", "256color", "16color", "ansi"}
	for _, colorType := range colorTerms {
		if strings.Contains(term, colorType) {
			return true
		}
	}

	return term != "" && term != "dumb"
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// Debug and development functions
func PrintDebug(format string, args ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		dimColor.Print("[DEBUG] ")
		fmt.Printf(format+"\n", args...)
	}
}

// Animation helpers
func PrintLoadingDots(message string, count int, delay time.Duration) {
	for i := 0; i < count; i++ {
		fmt.Printf("\r%s%s", message, strings.Repeat(".", i+1))
		time.Sleep(delay)
	}
	fmt.Println()
}

func PrintCountdown(seconds int) {
	for i := seconds; i > 0; i-- {
		fmt.Printf("\rStarting in %d seconds...", i)
		time.Sleep(time.Second)
	}
	fmt.Print("\r" + strings.Repeat(" ", 25) + "\r")
}
