package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// FavoritesFile is the name of the favorites storage file
const FavoritesFile = "favorites.json"

func (ce *ColorEditor) favoritesPath() string {
	return filepath.Join(filepath.Dir(ce.config.ConfigFile), FavoritesFile)
}

// loadFavorites reads favorites.json once at startup. The result is cached in
// memory: the list is redrawn on every keystroke and must not hit the disk.
func (ce *ColorEditor) loadFavorites() map[string]bool {
	favorites := make(map[string]bool)

	data, err := os.ReadFile(ce.favoritesPath())
	if err != nil {
		return favorites
	}

	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return favorites
	}
	for _, name := range names {
		favorites[name] = true
	}

	return favorites
}

// saveFavorites persists the in-memory set, sorted for a stable diff.
func (ce *ColorEditor) saveFavorites() error {
	names := make([]string, 0, len(ce.favorites))
	for name := range ce.favorites {
		names = append(names, name)
	}
	sort.Strings(names)

	data, err := json.MarshalIndent(names, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ce.favoritesPath(), data, 0644)
}

// toggleFavorite adds or removes the current theme from favorites
func (ce *ColorEditor) toggleFavorite() {
	name := ce.currentVisibleName()
	if name == "" {
		name = ce.themeName
	}
	if name == "" {
		ce.warn("No theme selected")
		return
	}

	if ce.favorites[name] {
		delete(ce.favorites, name)
	} else {
		ce.favorites[name] = true
	}

	if err := ce.saveFavorites(); err != nil {
		ce.fail("Failed to update favourites: %v", err)
		return
	}

	if ce.favOnly {
		ce.rebuildThemeList()
	} else if idx := ce.indexOfVisible(name); idx >= 0 {
		ce.themeList.SetItemText(idx, ce.formatThemeItem(name), "")
	}

	if ce.favorites[name] {
		ce.info("♥ %s added to favourites", name)
	} else {
		ce.info("%s removed from favourites", name)
	}
}
