package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// configWriteMu serialises read-modify-write cycles on the user's config.
//
// Two subsystems edit alacritty.toml: the theme manager, which keeps the import
// line present, and the font manager, which rewrites the [font] table. The font
// panel used to fire an update from a goroutine on every cursor movement, so
// dozens of these cycles overlapped. Whoever read while another had truncated
// the file saw an empty document, concluded there was no [font] section, and
// wrote back a config containing nothing but the one it had just built — every
// other section of the user's config gone. The lock makes each cycle indivisible.
var configWriteMu sync.Mutex

// UpdateConfigFile applies transform to the contents of path and writes the
// result back. The whole cycle is held under one lock, and the write is atomic,
// so a concurrent reader sees either the old file or the new one and never the
// empty window that os.WriteFile leaves between truncating and writing.
//
// transform receives "" when the file does not exist yet. Returning the input
// unchanged is free: nothing is written.
func UpdateConfigFile(path string, transform func(content string) (string, error)) error {
	configWriteMu.Lock()
	defer configWriteMu.Unlock()

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	updated, err := transform(string(existing))
	if err != nil {
		return err
	}
	if updated == string(existing) {
		return nil
	}

	return writeFileAtomic(path, []byte(updated), 0644)
}

// writeFileAtomic replaces path in one step: the content is written to a
// temporary file in the same directory and then renamed over the target, which
// is atomic on every platform this runs on. A crash or a concurrent reader can
// therefore never observe a half-written or truncated config.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpName := tmp.Name()

	// Any failure past this point must not leave the temporary file behind.
	cleanup := func() { os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("failed to write temporary file: %w", err)
	}
	// Flush to disk before the rename, so a crash cannot leave the config
	// renamed into place but empty.
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("failed to flush temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("failed to close temporary file: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		cleanup()
		return fmt.Errorf("failed to set permissions: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}
