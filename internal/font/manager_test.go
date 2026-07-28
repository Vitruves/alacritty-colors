package font

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/vitruves/alacritty-colors/internal/config"
)

// The config this is modelled on: every table here was lost once.
const fullConfig = `[general]
import = ["themes/current.toml"]

[cursor]
style = { shape = "Block", blinking = "Always" }

[window]
opacity = 1.0

[font]
normal = { family = "Maple Mono", style = "Thin" }
size = 14
builtin_box_drawing = true

[terminal]
shell = { program = "/bin/zsh", args = ["-c", "fish"] }

[[keyboard.bindings]]
key = "Return"
mods = "Shift"
`

func TestApplyFontPropertyKeepsEveryOtherTable(t *testing.T) {
	cases := []struct {
		property string
		value    string
		want     string
	}{
		{"family", "JetBrains Mono", `normal = { family = "JetBrains Mono", style = "Thin" }`},
		{"style", "Bold", `normal = { family = "Maple Mono", style = "Bold" }`},
		{"size", "18.0", "size = 18.0"},
	}

	for _, tc := range cases {
		t.Run(tc.property, func(t *testing.T) {
			got, err := applyFontProperty(fullConfig, tc.property, tc.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, table := range []string{"[general]", "[cursor]", "[window]", "[font]", "[terminal]", "[[keyboard.bindings]]"} {
				if !strings.Contains(got, table) {
					t.Errorf("%s was dropped from the config:\n%s", table, got)
				}
			}
			if !strings.Contains(got, `import = ["themes/current.toml"]`) {
				t.Errorf("the theme import was dropped:\n%s", got)
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("missing %q in:\n%s", tc.want, got)
			}
		})
	}
}

// Changing one property must leave the other two exactly as they were.
func TestApplyFontPropertyPreservesSiblings(t *testing.T) {
	got, err := applyFontProperty(fullConfig, "family", "Iosevka")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `style = "Thin"`) {
		t.Errorf("style was not preserved:\n%s", got)
	}
	if !strings.Contains(got, "size = 14") {
		t.Errorf("size was not preserved:\n%s", got)
	}
	if !strings.Contains(got, "builtin_box_drawing = true") {
		t.Errorf("an unrelated font key was dropped:\n%s", got)
	}
}

func TestApplyFontPropertyAddsMissingPieces(t *testing.T) {
	t.Run("no font table", func(t *testing.T) {
		got, err := applyFontProperty("[window]\nopacity = 1.0\n", "size", "13.0")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, "[window]") || !strings.Contains(got, "[font]") || !strings.Contains(got, "size = 13.0") {
			t.Errorf("unexpected result:\n%s", got)
		}
		// The new table must come after the existing keys, never before, or
		// [window]'s keys would be re-parented into [font].
		if strings.Index(got, "[font]") < strings.Index(got, "opacity") {
			t.Errorf("[font] was inserted above existing root keys:\n%s", got)
		}
	})

	t.Run("font table without the key", func(t *testing.T) {
		got, err := applyFontProperty("[font]\nsize = 12\n\n[window]\nopacity = 1.0\n", "family", "Iosevka")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, `family = "Iosevka"`) || !strings.Contains(got, "size = 12") {
			t.Errorf("unexpected result:\n%s", got)
		}
		if !strings.Contains(got, "[window]") {
			t.Errorf("[window] was dropped:\n%s", got)
		}
	})
}

// The regression that started this: concurrent updates used to interleave a
// read with another goroutine's truncating write, and the loser wrote back a
// config containing nothing but a [font] table.
func TestConcurrentFontUpdatesNeverDestroyTheConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "alacritty.toml")
	if err := os.WriteFile(path, []byte(fullConfig), 0644); err != nil {
		t.Fatal(err)
	}

	families := []string{"Maple Mono", "Iosevka", "Menlo", "Monaco", "Courier New"}

	var wg sync.WaitGroup
	for i := 0; i < 60; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			property, value := "family", families[i%len(families)]
			switch i % 3 {
			case 1:
				property, value = "style", "Regular"
			case 2:
				property, value = "size", "15.0"
			}
			err := config.UpdateConfigFile(path, func(content string) (string, error) {
				return applyFontProperty(content, property, value)
			})
			if err != nil {
				t.Errorf("update failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	final, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"[general]", "[cursor]", "[window]", "[font]", "[terminal]", "[[keyboard.bindings]]"} {
		if !strings.Contains(string(final), table) {
			t.Fatalf("%s was lost after concurrent updates; config is now:\n%s", table, final)
		}
	}
	if !strings.Contains(string(final), `import = ["themes/current.toml"]`) {
		t.Fatalf("the theme import was lost; config is now:\n%s", final)
	}
}

func TestParseFontFace(t *testing.T) {
	family, style := parseFontFace(fullConfig)
	if family != "Maple Mono" || style != "Thin" {
		t.Errorf("parseFontFace() = %q, %q; want \"Maple Mono\", \"Thin\"", family, style)
	}

	// A normal key in another table must not be mistaken for the font face.
	family, style = parseFontFace("[colors.normal]\nblack = \"#000000\"\n")
	if family != "" || style != "" {
		t.Errorf("parseFontFace() = %q, %q; want empty for a config with no font face", family, style)
	}
}
