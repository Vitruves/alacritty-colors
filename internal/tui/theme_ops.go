package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

// loadTheme loads a theme by name and extracts its colors
func (ce *ColorEditor) loadTheme(themeName string) {
	themeFile := ce.config.GetThemePath(themeName)
	parser := alacritty.NewParser()

	config, err := parser.ParseFile(themeFile)
	if err != nil {
		ce.setStatus(fmt.Sprintf("Error loading theme %s: %v", themeName, err))
		return
	}

	ce.currentTheme = config
	ce.extractColors()
	ce.isDirty = false
}

// extractColors extracts colors from current theme into the color values map
func (ce *ColorEditor) extractColors() {
	ce.colorValues = make(map[string]string)
	ce.colorKeys = make([]string, 0)

	if ce.currentTheme == nil {
		return
	}

	// Primary colors
	ce.addColor("primary.background", ce.currentTheme.Colors.Primary.Background)
	ce.addColor("primary.foreground", ce.currentTheme.Colors.Primary.Foreground)

	// Cursor colors
	if ce.currentTheme.Colors.Cursor.Text != "" {
		ce.addColor("cursor.text", ce.currentTheme.Colors.Cursor.Text)
	}
	if ce.currentTheme.Colors.Cursor.Cursor != "" {
		ce.addColor("cursor.cursor", ce.currentTheme.Colors.Cursor.Cursor)
	}

	// Selection colors
	if ce.currentTheme.Colors.Selection.Text != "" {
		ce.addColor("selection.text", ce.currentTheme.Colors.Selection.Text)
	}
	if ce.currentTheme.Colors.Selection.Background != "" {
		ce.addColor("selection.background", ce.currentTheme.Colors.Selection.Background)
	}

	// Normal colors
	for name, color := range ce.currentTheme.Colors.Normal {
		ce.addColor("normal."+name, color)
	}

	// Bright colors
	for name, color := range ce.currentTheme.Colors.Bright {
		ce.addColor("bright."+name, color)
	}

	// Dim colors
	for name, color := range ce.currentTheme.Colors.Dim {
		ce.addColor("dim."+name, color)
	}
}

// addColor adds a color to the color values map if non-empty
func (ce *ColorEditor) addColor(key, value string) {
	if value != "" {
		ce.colorValues[key] = value
		ce.colorKeys = append(ce.colorKeys, key)
	}
}

// saveTheme saves the current theme to file
func (ce *ColorEditor) saveTheme() {
	if ce.currentTheme == nil || ce.themeName == "" {
		ce.setStatus("No theme selected to save")
		return
	}

	ce.updateThemeConfig()

	err := ce.saveThemeToFile()
	if err != nil {
		ce.setStatus(fmt.Sprintf("Error saving theme: %v", err))
		return
	}

	ce.isDirty = false
	ce.setStatus(fmt.Sprintf("Theme '%s' saved successfully", ce.themeName))
}

// saveThemeToFile writes the theme to disk
func (ce *ColorEditor) saveThemeToFile() error {
	themeFile := ce.config.GetThemePath(ce.themeName)
	content := ce.generateTOMLContent()

	err := os.WriteFile(themeFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write theme file %s: %w", themeFile, err)
	}

	// Force sync to disk
	file, err := os.OpenFile(themeFile, os.O_RDONLY, 0)
	if err == nil {
		file.Sync()
		file.Close()
	}

	return nil
}

// generateTOMLContent generates TOML content from current color values
func (ce *ColorEditor) generateTOMLContent() string {
	var content strings.Builder

	content.WriteString("# Alacritty theme - edited with alacritty-colors TUI\n\n")

	// Primary colors
	content.WriteString("[colors.primary]\n")
	content.WriteString(fmt.Sprintf("background = \"%s\"\n", ce.colorValues["primary.background"]))
	content.WriteString(fmt.Sprintf("foreground = \"%s\"\n", ce.colorValues["primary.foreground"]))
	content.WriteString("\n")

	// Cursor colors
	if ce.colorValues["cursor.text"] != "" || ce.colorValues["cursor.cursor"] != "" {
		content.WriteString("[colors.cursor]\n")
		if ce.colorValues["cursor.text"] != "" {
			content.WriteString(fmt.Sprintf("text = \"%s\"\n", ce.colorValues["cursor.text"]))
		}
		if ce.colorValues["cursor.cursor"] != "" {
			content.WriteString(fmt.Sprintf("cursor = \"%s\"\n", ce.colorValues["cursor.cursor"]))
		}
		content.WriteString("\n")
	}

	// Selection colors
	if ce.colorValues["selection.text"] != "" || ce.colorValues["selection.background"] != "" {
		content.WriteString("[colors.selection]\n")
		if ce.colorValues["selection.text"] != "" {
			content.WriteString(fmt.Sprintf("text = \"%s\"\n", ce.colorValues["selection.text"]))
		}
		if ce.colorValues["selection.background"] != "" {
			content.WriteString(fmt.Sprintf("background = \"%s\"\n", ce.colorValues["selection.background"]))
		}
		content.WriteString("\n")
	}

	// Normal colors
	content.WriteString("[colors.normal]\n")
	for _, color := range BaseColorNames {
		if value, exists := ce.colorValues["normal."+color]; exists {
			content.WriteString(fmt.Sprintf("%s = \"%s\"\n", color, value))
		}
	}
	content.WriteString("\n")

	// Bright colors
	content.WriteString("[colors.bright]\n")
	for _, color := range BaseColorNames {
		if value, exists := ce.colorValues["bright."+color]; exists {
			content.WriteString(fmt.Sprintf("%s = \"%s\"\n", color, value))
		}
	}
	content.WriteString("\n")

	// Dim colors (if any)
	hasDimColors := false
	for _, color := range BaseColorNames {
		if _, exists := ce.colorValues["dim."+color]; exists {
			hasDimColors = true
			break
		}
	}

	if hasDimColors {
		content.WriteString("[colors.dim]\n")
		for _, color := range BaseColorNames {
			if value, exists := ce.colorValues["dim."+color]; exists {
				content.WriteString(fmt.Sprintf("%s = \"%s\"\n", color, value))
			}
		}
		content.WriteString("\n")
	}

	return content.String()
}

// updateThemeConfig updates the current theme config from color values
func (ce *ColorEditor) updateThemeConfig() {
	ce.currentTheme.Colors.Primary.Background = ce.colorValues["primary.background"]
	ce.currentTheme.Colors.Primary.Foreground = ce.colorValues["primary.foreground"]

	if ce.colorValues["cursor.text"] != "" {
		ce.currentTheme.Colors.Cursor.Text = ce.colorValues["cursor.text"]
	}
	if ce.colorValues["cursor.cursor"] != "" {
		ce.currentTheme.Colors.Cursor.Cursor = ce.colorValues["cursor.cursor"]
	}

	if ce.colorValues["selection.text"] != "" {
		ce.currentTheme.Colors.Selection.Text = ce.colorValues["selection.text"]
	}
	if ce.colorValues["selection.background"] != "" {
		ce.currentTheme.Colors.Selection.Background = ce.colorValues["selection.background"]
	}

	for name := range ce.currentTheme.Colors.Normal {
		if value, exists := ce.colorValues["normal."+name]; exists {
			ce.currentTheme.Colors.Normal[name] = value
		}
	}

	for name := range ce.currentTheme.Colors.Bright {
		if value, exists := ce.colorValues["bright."+name]; exists {
			ce.currentTheme.Colors.Bright[name] = value
		}
	}

	for name := range ce.currentTheme.Colors.Dim {
		if value, exists := ce.colorValues["dim."+name]; exists {
			ce.currentTheme.Colors.Dim[name] = value
		}
	}
}

// resetTheme resets the theme to its original values
func (ce *ColorEditor) resetTheme() {
	if ce.themeName == "" {
		return
	}

	ce.loadTheme(ce.themeName)
	ce.buildColorPanel()
	ce.updatePreview()
	if len(ce.colorKeys) > 0 {
		ce.colorPanel.SetCurrentItem(0)
	}
	ce.setStatus("Theme reset to original values")
}

// applyCurrentTheme applies the currently selected theme
func (ce *ColorEditor) applyCurrentTheme() {
	if ce.themeName == "" {
		ce.setStatus("No theme selected to apply")
		return
	}

	if ce.isDirty {
		ce.saveTheme()
	}

	if err := ce.themeManager.ApplyTheme(ce.themeName); err != nil {
		ce.setStatus(fmt.Sprintf("Failed to apply theme: %v", err))
	} else {
		ce.appliedTheme = ce.themeName
		ce.setStatus(fmt.Sprintf("Theme '%s' applied successfully | e: edit copy | d: delete | p: parameters | f: font | q: quit", ce.themeName))
		ce.refreshThemeList()
	}
}

// createThemeCopy creates an editable copy of the current theme
func (ce *ColorEditor) createThemeCopy() {
	if ce.themeName == "" {
		return
	}

	originalName := ce.themeName
	if strings.Contains(originalName, EditedThemeSuffix) {
		parts := strings.Split(originalName, EditedThemeSuffix)
		originalName = parts[0]
	}
	timestamp := time.Now().Format("150405")
	copyName := fmt.Sprintf("%s%s%s", originalName, EditedThemeSuffix, timestamp)

	ce.themeName = copyName
	ce.colorPanel.SetTitle(fmt.Sprintf(" Color Palette - %s (Copy) ", copyName))

	go func(themeName string) {
		ce.updateThemeConfig()
		if err := ce.saveThemeToFile(); err == nil {
			if err := ce.themeManager.ApplyTheme(themeName); err != nil {
				ce.app.QueueUpdateDraw(func() {
					ce.setStatus(fmt.Sprintf("Failed to apply copy: %v", err))
				})
			} else {
				ce.app.QueueUpdateDraw(func() {
					ce.setStatus(fmt.Sprintf("Created and applied copy: %s - Ready to edit", themeName))
					ce.appliedTheme = themeName
					ce.refreshThemeList()
				})
			}
		} else {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus(fmt.Sprintf("Failed to save copy: %v", err))
			})
		}
	}(copyName)
}

// deleteCurrentTheme deletes the currently selected edited theme
func (ce *ColorEditor) deleteCurrentTheme() {
	if ce.themeName == "" {
		ce.setStatus("No theme selected to delete")
		return
	}

	if !strings.Contains(ce.themeName, EditedThemeSuffix) {
		ce.setStatus("Cannot delete base theme. Only edited versions can be deleted.")
		return
	}

	ce.showDeleteConfirmation()
}

// performThemeDelete actually deletes the theme file
func (ce *ColorEditor) performThemeDelete() {
	themeFile := ce.config.GetThemePath(ce.themeName)
	if err := os.Remove(themeFile); err != nil {
		ce.setStatus(fmt.Sprintf("Failed to delete theme: %v", err))
		return
	}

	if ce.appliedTheme == ce.themeName {
		trackingFile := ce.config.GetThemePath(".current-theme")
		os.Remove(trackingFile)
		ce.appliedTheme = ""
	}

	ce.setStatus(fmt.Sprintf("Theme '%s' deleted successfully", ce.themeName))
	ce.refreshThemeList()

	if ce.themeList.GetItemCount() > 0 {
		ce.themeList.SetCurrentItem(0)
		themeName, _ := ce.themeList.GetItemText(0)
		themeName = strings.TrimPrefix(themeName, CurrentThemeMarker)
		themeName = strings.TrimPrefix(themeName, "♥ ")
		ce.onThemeSelected(0, themeName, "", 0)
	}
}

// cleanAndRedownloadThemes cleans and redownloads all themes
func (ce *ColorEditor) cleanAndRedownloadThemes() {
	go func() {
		ce.app.QueueUpdateDraw(func() {
			ce.setStatus("Cleaning and redownloading themes...")
		})

		if err := ce.parametersManager.CleanAndRedownloadThemes(); err != nil {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus(fmt.Sprintf("Failed to clean and redownload themes: %v", err))
			})
		} else {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus("Themes cleaned and redownloaded successfully")
				ce.loadThemes()
			})
		}
	}()
}

// backupCurrentConfig creates a backup of the current config
func (ce *ColorEditor) backupCurrentConfig() {
	go func() {
		ce.app.QueueUpdateDraw(func() {
			ce.setStatus("Creating backup of current configuration...")
		})

		if err := ce.parametersManager.BackupConfig(); err != nil {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus(fmt.Sprintf("Failed to backup config: %v", err))
			})
		} else {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus("Configuration backed up successfully")
			})
		}
	}()
}

// resetToDefaults resets to default configuration
func (ce *ColorEditor) resetToDefaults() {
	go func() {
		ce.app.QueueUpdateDraw(func() {
			ce.setStatus("Resetting to default configuration...")
		})

		if err := ce.parametersManager.ResetToDefaults(); err != nil {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus(fmt.Sprintf("Failed to reset to defaults: %v", err))
			})
		} else {
			ce.app.QueueUpdateDraw(func() {
				ce.setStatus("Configuration reset to defaults successfully")
			})
		}
	}()
}
