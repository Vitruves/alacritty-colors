package font

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsFontFile(t *testing.T) {
	cases := map[string]bool{
		"JetBrainsMonoNerdFont-Regular.ttf": true,
		"SpaceMonoNerdFont-Bold.otf":        true,
		"SomeFont.TTF":                      true,
		"README.md":                         false,
		"LICENSE":                           false,
		"install.sh":                        false,
		// Nerd Fonts ships a second cut of every face that differs only in its
		// internal naming; installing both doubles the list for nothing.
		"Hack Windows Compatible-Regular.ttf": false,
	}

	for name, want := range cases {
		if got := isFontFile(name); got != want {
			t.Errorf("isFontFile(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestFontInstallDirIsUnderHome(t *testing.T) {
	dir, err := FontInstallDir()
	if err != nil {
		t.Fatal(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dir, home) {
		t.Errorf("font directory %q is outside the home directory; installing must never need elevated rights", dir)
	}

	if runtime.GOOS == "darwin" && !strings.Contains(dir, "Library/Fonts") {
		t.Errorf("on macOS fonts belong in ~/Library/Fonts, got %q", dir)
	}
}

// extractFaces flattens the archive, skips everything that is not a face, and
// must not be talked into writing outside the target directory.
func TestExtractFaces(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "font.zip")
	entries := map[string]string{
		"SpaceMono/README.md":                         "not a font",
		"SpaceMono/SpaceMonoNerdFont-Regular.ttf":     "face one",
		"SpaceMono/nested/SpaceMonoNerdFont-Bold.otf": "face two",
		"SpaceMono/SpaceMono Windows Compatible.ttf":  "duplicate cut",
		"../escape.ttf":                               "path traversal",
	}

	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	file.Close()

	target := t.TempDir()
	installed, err := extractFaces(archive, target)
	if err != nil {
		t.Fatal(err)
	}

	if installed != 3 {
		t.Errorf("installed %d faces, want 3", installed)
	}

	found := map[string]bool{}
	written, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range written {
		found[entry.Name()] = true
	}

	for _, want := range []string{"SpaceMonoNerdFont-Regular.ttf", "SpaceMonoNerdFont-Bold.otf"} {
		if !found[want] {
			t.Errorf("%s was not installed; got %v", want, found)
		}
	}
	if found["README.md"] {
		t.Error("a non-font file was installed")
	}

	// The traversal entry is a .ttf, so it is installed — but flattened to a
	// base name inside the target, never above it.
	if _, err := os.Stat(filepath.Join(filepath.Dir(target), "escape.ttf")); err == nil {
		t.Error("an archive entry escaped the target directory")
	}
	if !found["escape.ttf"] {
		t.Error("the traversal entry should have been flattened into the target, not dropped silently")
	}
}

// Every offered font must name a real release asset, or choosing it can only
// ever produce an error. The URL is not fetched here; this guards the shape.
func TestAvailableDownloadsAreWellFormed(t *testing.T) {
	seen := map[string]bool{}

	for _, candidate := range AvailableDownloads {
		if candidate.Name == "" || candidate.Archive == "" || candidate.Note == "" {
			t.Errorf("incomplete entry: %+v", candidate)
		}
		if strings.ContainsAny(candidate.Archive, " /\\") {
			t.Errorf("archive name %q is not a bare asset name", candidate.Archive)
		}
		if seen[candidate.Archive] {
			t.Errorf("archive %q is listed twice", candidate.Archive)
		}
		seen[candidate.Archive] = true
	}

	if len(AvailableDownloads) < 10 {
		t.Errorf("only %d fonts offered; the point is to have a real choice", len(AvailableDownloads))
	}
}
