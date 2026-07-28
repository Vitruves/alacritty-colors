package tui

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

// ThemeHeader marks the theme files this tool wrote for the user. Files
// carrying it can be overwritten in place; anything else — a downloaded theme
// or one from the curated collection — is only ever forked, never clobbered.
const ThemeHeader = "# alacritty-colors ·"

// validThemeName keeps generated file names to something a shell can handle.
var validThemeName = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// loadTheme loads a theme by name and extracts its colors
func (ce *ColorEditor) loadTheme(themeName string) {
	themeFile := ce.config.GetThemePath(themeName)
	parser := alacritty.NewParser()

	config, err := parser.ParseFile(themeFile)
	if err != nil {
		ce.fail("Error loading theme %s: %v", themeName, err)
		return
	}

	ce.currentTheme = config
	ce.extractColors()
	ce.isDirty = false
	ce.discardArmed = false
	ce.themeOwned = ce.fileIsOwned(themeFile)
}

// fileIsOwned reports whether a theme file was written by this tool.
func (ce *ColorEditor) fileIsOwned(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, len(ThemeHeader))
	n, _ := file.Read(buf)
	return string(buf[:n]) == ThemeHeader
}

// extractColors extracts colors from current theme into the color values map
func (ce *ColorEditor) extractColors() {
	ce.colorValues = make(map[string]string)
	ce.originalValues = make(map[string]string)

	if ce.currentTheme == nil {
		return
	}

	ce.addColor("primary.background", ce.currentTheme.Colors.Primary.Background)
	ce.addColor("primary.foreground", ce.currentTheme.Colors.Primary.Foreground)
	ce.addColor("cursor.text", ce.currentTheme.Colors.Cursor.Text)
	ce.addColor("cursor.cursor", ce.currentTheme.Colors.Cursor.Cursor)
	ce.addColor("selection.text", ce.currentTheme.Colors.Selection.Text)
	ce.addColor("selection.background", ce.currentTheme.Colors.Selection.Background)

	for name, color := range ce.currentTheme.Colors.Normal {
		ce.addColor("normal."+name, color)
	}
	for name, color := range ce.currentTheme.Colors.Bright {
		ce.addColor("bright."+name, color)
	}
	for name, color := range ce.currentTheme.Colors.Dim {
		ce.addColor("dim."+name, color)
	}

	// Snapshot for undo: 'u' and 'r' restore from here instead of re-reading
	// the file, so a revert costs nothing and never races a pending write.
	for key, value := range ce.colorValues {
		ce.originalValues[key] = value
	}
}

// addColor adds a color to the color values map if non-empty
func (ce *ColorEditor) addColor(key, value string) {
	if value == "" {
		return
	}
	ce.colorValues[key] = normalizeHex(value, value)
}

// saveTheme writes the working colours to disk. Downloaded themes are never
// overwritten: editing one forks it under a new name instead.
func (ce *ColorEditor) saveTheme() {
	if ce.currentTheme == nil || ce.themeName == "" {
		ce.warn("No theme selected to save")
		return
	}
	if !ce.isDirty {
		ce.info("Nothing to save — %s is unchanged", ce.themeName)
		return
	}

	if !ce.themeOwned {
		ce.showSaveAsDialog()
		return
	}

	if err := ce.writeTheme(ce.themeName); err != nil {
		ce.fail("Error saving theme: %v", err)
		return
	}

	ce.commitSave(ce.themeName)
	ce.info("Saved %s", ce.themeName)
}

// saveThemeAs writes the working colours under a new name and switches to it.
func (ce *ColorEditor) saveThemeAs(name string) {
	name = sanitizeThemeName(name)
	if name == "" {
		ce.warn("Theme name cannot be empty")
		return
	}

	// Set the name first: it is stamped into the file header.
	ce.themeName = name
	if err := ce.writeTheme(name); err != nil {
		ce.fail("Error saving theme: %v", err)
		return
	}
	ce.themeOwned = true
	ce.commitSave(name)
	ce.refreshThemeList()
	if idx := ce.indexOfVisible(name); idx >= 0 {
		ce.themeList.SetCurrentItem(idx)
	}
	ce.applyThemeNow(name)
	ce.info("Saved as %s", name)
}

// commitSave clears the dirty state and re-snapshots the on-disk values.
func (ce *ColorEditor) commitSave(name string) {
	ce.isDirty = false
	ce.discardArmed = false
	ce.originalValues = make(map[string]string, len(ce.colorValues))
	for key, value := range ce.colorValues {
		ce.originalValues[key] = value
	}
	ce.colorPanel.SetTitle(ce.colorPanelTitle())
	ce.markApplied(name)
}

// writeTheme serialises the working colours to a theme file.
func (ce *ColorEditor) writeTheme(name string) error {
	themeFile := ce.config.GetThemePath(name)
	if err := os.WriteFile(themeFile, []byte(ce.generateTOMLContent()), 0644); err != nil {
		return fmt.Errorf("failed to write theme file %s: %w", themeFile, err)
	}
	return nil
}

// sanitizeThemeName reduces user input to a safe file stem.
func sanitizeThemeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = validThemeName.ReplaceAllString(name, "")
	name = strings.Trim(name, ".-")
	return strings.TrimSuffix(name, ".toml")
}

// suggestThemeName proposes a free name derived from the current theme.
func (ce *ColorEditor) suggestThemeName() string {
	base := ce.themeName
	if base == "" {
		base = "custom"
	}
	if i := strings.Index(base, EditedThemeSuffix); i > 0 {
		base = base[:i]
	}

	// A generated draft already has a name of its own and no file behind it.
	if !ce.themeExists(base) {
		return base
	}

	candidate := base + "-custom"
	if !ce.themeExists(candidate) {
		return candidate
	}
	for n := 2; n < 100; n++ {
		candidate = fmt.Sprintf("%s-custom-%d", base, n)
		if !ce.themeExists(candidate) {
			return candidate
		}
	}
	return fmt.Sprintf("%s%s%s", base, EditedThemeSuffix, time.Now().Format("150405"))
}

func (ce *ColorEditor) themeExists(name string) bool {
	_, err := os.Stat(ce.config.GetThemePath(name))
	return err == nil
}

// generateTOMLContent generates TOML content from current color values
func (ce *ColorEditor) generateTOMLContent() string {
	var content strings.Builder

	content.WriteString(ThemeHeader)
	content.WriteString(fmt.Sprintf(" %s\n\n", ce.themeName))

	content.WriteString("[colors.primary]\n")
	content.WriteString(fmt.Sprintf("background = \"%s\"\n", ce.colorValues["primary.background"]))
	content.WriteString(fmt.Sprintf("foreground = \"%s\"\n\n", ce.colorValues["primary.foreground"]))

	writePair := func(section, keyA, keyB, nameA, nameB string) {
		a, b := ce.colorValues[keyA], ce.colorValues[keyB]
		if a == "" && b == "" {
			return
		}
		content.WriteString(fmt.Sprintf("[colors.%s]\n", section))
		if a != "" {
			content.WriteString(fmt.Sprintf("%s = \"%s\"\n", nameA, a))
		}
		if b != "" {
			content.WriteString(fmt.Sprintf("%s = \"%s\"\n", nameB, b))
		}
		content.WriteString("\n")
	}
	writePair("cursor", "cursor.text", "cursor.cursor", "text", "cursor")
	writePair("selection", "selection.text", "selection.background", "text", "background")

	for _, group := range []string{"normal", "bright", "dim"} {
		var body strings.Builder
		for _, color := range BaseColorNames {
			if value, exists := ce.colorValues[group+"."+color]; exists && value != "" {
				body.WriteString(fmt.Sprintf("%s = \"%s\"\n", color, value))
			}
		}
		if body.Len() == 0 {
			continue
		}
		content.WriteString(fmt.Sprintf("[colors.%s]\n", group))
		content.WriteString(body.String())
		content.WriteString("\n")
	}

	return content.String()
}

// resetTheme restores every colour to the values last read from disk.
func (ce *ColorEditor) resetTheme() {
	if len(ce.originalValues) == 0 {
		// A generated draft has no saved version; go back to what is applied.
		if ce.appliedTheme != "" {
			ce.selectTheme(ce.appliedTheme, true)
			if idx := ce.indexOfVisible(ce.appliedTheme); idx >= 0 {
				ce.themeList.SetCurrentItem(idx)
			}
			ce.info("Discarded the draft, back to %s", ce.appliedTheme)
		}
		return
	}
	if ce.themeName == "" {
		return
	}

	ce.colorValues = make(map[string]string, len(ce.originalValues))
	for key, value := range ce.originalValues {
		ce.colorValues[key] = value
	}
	ce.isDirty = false
	ce.discardArmed = false

	ce.buildColorPanel()
	ce.colorPanel.SetTitle(ce.colorPanelTitle())
	ce.updatePreview()
	ce.scheduleLivePreview()
	ce.info("Reverted %s to the saved version", ce.themeName)
}

// persistCurrent writes the working colours to a theme file and applies that
// file, so the edit survives both the session and the next launch. Downloaded
// themes are forked rather than overwritten. It reports the name it landed on.
func (ce *ColorEditor) persistCurrent() (string, bool) {
	name := ce.themeName
	if !ce.themeOwned {
		name = ce.suggestThemeName()
		ce.themeName = name
	}

	// Cancel any debounced write still in flight so it cannot land on top.
	ce.applySeq.Add(1)

	if err := ce.writeTheme(name); err != nil {
		ce.fail("Could not save %s: %v", name, err)
		return "", false
	}
	if err := ce.themeManager.ApplyTheme(name); err != nil {
		ce.fail("Saved %s but could not apply it: %v", name, err)
		return "", false
	}

	ce.themeOwned = true
	ce.commitSave(name)
	ce.refreshThemeList()
	if idx := ce.indexOfVisible(name); idx >= 0 {
		ce.themeList.SetCurrentItem(idx)
	}
	return name, true
}

// applyCurrentTheme makes what is on screen the theme, for good.
//
// On an untouched theme it is just an immediate apply, skipping the debounce
// browsing relies on. With unsaved edits it also writes them out: an apply that
// only reached the live preview would be lost the moment you moved the cursor,
// which is not what pressing apply should mean.
func (ce *ColorEditor) applyCurrentTheme() {
	if ce.themeName == "" {
		ce.warn("No theme selected to apply")
		return
	}

	if !ce.isDirty {
		ce.applyThemeNow(ce.themeName)
		return
	}

	if name, ok := ce.persistCurrent(); ok {
		ce.info("Saved and applied %s", name)
	}
}

// quitAndKeep leaves the editor with whatever is on screen already applied,
// committing unsaved edits on the way out rather than throwing them away.
func (ce *ColorEditor) quitAndKeep() {
	if !ce.isDirty || ce.themeName == "" {
		ce.applySeq.Add(1)
		ce.app.Stop()
		return
	}

	name, ok := ce.persistCurrent()
	if !ok {
		// The failure is on screen; a second q leaves without retrying.
		ce.isDirty = false
		return
	}

	ce.exitNote = fmt.Sprintf("Saved and applied %s", name)
	ce.app.Stop()
}

// createThemeCopy forks the current theme so it can be edited freely.
func (ce *ColorEditor) createThemeCopy() {
	if ce.themeName == "" {
		return
	}

	copyName := ce.suggestThemeName()
	if err := ce.writeTheme(copyName); err != nil {
		ce.fail("Failed to create copy: %v", err)
		return
	}

	ce.themeName = copyName
	ce.themeOwned = true
	ce.commitSave(copyName)
	ce.refreshThemeList()
	if idx := ce.indexOfVisible(copyName); idx >= 0 {
		ce.themeList.SetCurrentItem(idx)
	}
	ce.applyThemeNow(copyName)
	ce.info("Editing copy %s — the original is untouched", copyName)
}

// deleteCurrentTheme deletes a theme this tool created.
func (ce *ColorEditor) deleteCurrentTheme() {
	// Prefer the theme actually loaded in the editor; fall back to the list
	// cursor when that theme is an unsaved draft with no file yet.
	name := ce.themeName
	if name == "" || !ce.themeExists(name) {
		name = ce.currentVisibleName()
	}
	if name == "" {
		ce.warn("No theme selected to delete")
		return
	}
	if !ce.fileIsOwned(ce.config.GetThemePath(name)) {
		ce.warn("%s is a downloaded theme — press e to fork it instead", name)
		return
	}

	ce.showDeleteConfirmation(name)
}

// performThemeDelete actually deletes the theme file
func (ce *ColorEditor) performThemeDelete(name string) {
	if err := os.Remove(ce.config.GetThemePath(name)); err != nil {
		ce.fail("Failed to delete theme: %v", err)
		return
	}

	delete(ce.favorites, name)
	_ = ce.saveFavorites()

	if ce.appliedTheme == name {
		os.Remove(ce.config.GetThemePath(".current-theme"))
		ce.appliedTheme = ""
	}

	ce.refreshThemeList()
	if next := ce.currentVisibleName(); next != "" {
		ce.selectTheme(next, true)
	}
	ce.info("Deleted %s", name)
}

// cleanAndRedownloadThemes cleans and redownloads all themes
func (ce *ColorEditor) cleanAndRedownloadThemes() {
	ce.info("Cleaning and redownloading themes…")

	go func() {
		err := ce.parametersManager.CleanAndRedownloadThemes()
		ce.app.QueueUpdateDraw(func() {
			if err != nil {
				ce.fail("Failed to redownload themes: %v", err)
				return
			}
			ce.loadThemes()
			ce.info("Themes redownloaded (%d)", len(ce.allThemes))
		})
	}()
}

// backupCurrentConfig creates a backup of the current config
func (ce *ColorEditor) backupCurrentConfig() {
	ce.info("Backing up your configuration…")

	go func() {
		err := ce.parametersManager.BackupConfig()
		ce.app.QueueUpdateDraw(func() {
			if err != nil {
				ce.fail("Failed to back up config: %v", err)
				return
			}
			ce.info("Configuration backed up")
		})
	}()
}

// resetToDefaults resets to default configuration
func (ce *ColorEditor) resetToDefaults() {
	ce.info("Resetting configuration…")

	go func() {
		err := ce.parametersManager.ResetToDefaults()
		ce.app.QueueUpdateDraw(func() {
			if err != nil {
				ce.fail("Failed to reset configuration: %v", err)
				return
			}
			ce.info("Configuration reset to defaults")
		})
	}()
}
