package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FavoritesFile is the name of the favorites storage file
const FavoritesFile = "favorites.json"

// loadFavorites loads the favorites list from disk
func (ce *ColorEditor) loadFavorites() []string {
	favPath := filepath.Join(filepath.Dir(ce.config.ConfigFile), FavoritesFile)

	data, err := os.ReadFile(favPath)
	if err != nil {
		return []string{}
	}

	var favorites []string
	if err := json.Unmarshal(data, &favorites); err != nil {
		return []string{}
	}

	return favorites
}

// saveFavorites saves the favorites list to disk
func (ce *ColorEditor) saveFavorites(favorites []string) error {
	favPath := filepath.Join(filepath.Dir(ce.config.ConfigFile), FavoritesFile)

	data, err := json.MarshalIndent(favorites, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(favPath, data, 0644)
}

// isFavorite checks if a theme is in the favorites list
func (ce *ColorEditor) isFavorite(themeName string) bool {
	favorites := ce.loadFavorites()
	for _, fav := range favorites {
		if fav == themeName {
			return true
		}
	}
	return false
}

// toggleFavorite adds or removes the current theme from favorites
func (ce *ColorEditor) toggleFavorite() {
	if ce.themeName == "" {
		ce.setStatus("No theme selected")
		return
	}

	favorites := ce.loadFavorites()

	// Check if already favorite
	found := -1
	for i, fav := range favorites {
		if fav == ce.themeName {
			found = i
			break
		}
	}

	if found >= 0 {
		// Remove from favorites
		favorites = append(favorites[:found], favorites[found+1:]...)
		if err := ce.saveFavorites(favorites); err != nil {
			ce.setStatus("Failed to update favorites: " + err.Error())
			return
		}
		ce.setStatus("Removed from favorites: " + ce.themeName)
	} else {
		// Add to favorites
		favorites = append(favorites, ce.themeName)
		if err := ce.saveFavorites(favorites); err != nil {
			ce.setStatus("Failed to update favorites: " + err.Error())
			return
		}
		ce.setStatus("Added to favorites: " + ce.themeName)
	}

	// Refresh the theme list to show the heart icon
	ce.refreshThemeList()
}
