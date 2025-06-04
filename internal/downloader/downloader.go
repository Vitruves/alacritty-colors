package downloader

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vitruves/alacritty-colors/internal/ui"
)

const (
	OfficialRepoURL = "https://github.com/alacritty/alacritty-theme/archive/refs/heads/master.zip"
	UserAgent       = "alacritty-colors/1.0.0"
	Timeout         = 30 * time.Second
)

type Downloader struct {
	themesDir string
	client    *http.Client
}

func New(themesDir string) *Downloader {
	return &Downloader{
		themesDir: themesDir,
		client: &http.Client{
			Timeout: Timeout,
		},
	}
}

func (d *Downloader) DownloadOfficialThemes() (int, error) {
	ui.PrintInfo("Downloading from official repository...")

	// Download the zip file
	resp, err := d.downloadFile(OfficialRepoURL)
	if err != nil {
		return 0, fmt.Errorf("failed to download themes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to download themes: HTTP %d", resp.StatusCode)
	}

	// Read the zip content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	ui.PrintInfo("Extracting themes...")

	// Extract theme files
	count, err := d.extractThemes(body)
	if err != nil {
		return 0, fmt.Errorf("failed to extract themes: %w", err)
	}

	return count, nil
}

func (d *Downloader) downloadFile(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)
	return d.client.Do(req)
}

func (d *Downloader) extractThemes(zipData []byte) (int, error) {
	// Create a zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return 0, fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Ensure themes directory exists
	if err := os.MkdirAll(d.themesDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create themes directory: %w", err)
	}

	themeCount := 0
	totalFiles := len(zipReader.File)
	processed := 0

	for _, file := range zipReader.File {
		processed++
		ui.PrintProgress(processed, totalFiles, "Processing")

		// Skip if not a theme file
		if !d.isThemeFile(file.Name) {
			continue
		}

		if err := d.extractThemeFile(file); err != nil {
			ui.PrintWarning("Failed to extract %s: %v", filepath.Base(file.Name), err)
			continue
		}

		themeCount++
	}

	return themeCount, nil
}

func (d *Downloader) isThemeFile(filename string) bool {
	// Check if it's in a themes directory and has the right extension
	return strings.Contains(filename, "themes/") &&
		(strings.HasSuffix(filename, ".toml") || strings.HasSuffix(filename, ".yaml"))
}

func (d *Downloader) extractThemeFile(file *zip.File) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	// Extract filename
	filename := filepath.Base(file.Name)
	outputPath := filepath.Join(d.themesDir, filename)

	// Check if file already exists and is newer
	if info, err := os.Stat(outputPath); err == nil {
		if info.ModTime().After(file.Modified) {
			return nil // Skip if local file is newer
		}
	}

	return os.WriteFile(outputPath, content, 0644)
}

func (d *Downloader) DownloadFromURL(url, filename string) error {
	ui.PrintInfo("Downloading theme from %s", url)

	resp, err := d.downloadFile(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	// Validate it's a theme file
	if !d.isValidTheme(content) {
		return fmt.Errorf("invalid theme file format")
	}

	outputPath := filepath.Join(d.themesDir, filename)
	if err := os.WriteFile(outputPath, content, 0644); err != nil {
		return fmt.Errorf("failed to save theme: %w", err)
	}

	ui.PrintSuccess("Downloaded theme: %s", filename)
	return nil
}

func (d *Downloader) isValidTheme(content []byte) bool {
	contentStr := string(content)

	// Basic validation - check for color sections
	hasColors := strings.Contains(contentStr, "[colors") ||
		strings.Contains(contentStr, "colors:")

	// Check for at least some color definitions
	hasColorDefs := strings.Contains(contentStr, "background") ||
		strings.Contains(contentStr, "foreground")

	return hasColors && hasColorDefs
}

func (d *Downloader) CleanupOldThemes(keepDays int) error {
	ui.PrintInfo("Cleaning up themes older than %d days", keepDays)

	files, err := os.ReadDir(d.themesDir)
	if err != nil {
		return fmt.Errorf("failed to read themes directory: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -keepDays)
	removed := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filepath := filepath.Join(d.themesDir, file.Name())
		info, err := os.Stat(filepath)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath); err != nil {
				ui.PrintWarning("Failed to remove %s: %v", file.Name(), err)
				continue
			}
			removed++
		}
	}

	if removed > 0 {
		ui.PrintSuccess("Removed %d old theme files", removed)
	} else {
		ui.PrintInfo("No old themes to remove")
	}

	return nil
}
