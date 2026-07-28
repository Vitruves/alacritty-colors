package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vitruves/alacritty-colors/internal/theme"
)

func hexToRGBForTest(hex string) (theme.RGB, error) { return theme.HexToRGB(hex) }

func TestExpandHex(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"#1e1e2e", "#1e1e2e", true},
		{"1e1e2e", "#1e1e2e", true},
		{"  #ABCDEF ", "#abcdef", true},
		{"#f0a", "#ff00aa", true},
		{"f0a", "#ff00aa", true},
		{"", "", false},
		{"#12345", "", false},
		{"#zzzzzz", "", false},
	}

	for _, c := range cases {
		got, ok := expandHex(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("expandHex(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		haystack, needle string
		want             bool
	}{
		{"catppuccin-mocha", "mocha", true},
		{"catppuccin-mocha", "ctpmoc", true},
		{"catppuccin-mocha", "cmocha", true},
		{"gruvbox-dark", "mocha", false},
		{"nord", "", true},
	}

	for _, c := range cases {
		if got := fuzzyMatch(c.haystack, c.needle); got != c.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", c.haystack, c.needle, got, c.want)
		}
	}
}

func TestSanitizeThemeName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"  My Theme  ", "My-Theme"},
		{"../../etc/passwd", "etcpasswd"},
		{"nord.toml", "nord"},
		{"a/b\\c", "abc"},
		{"...", ""},
	}

	for _, c := range cases {
		if got := sanitizeThemeName(c.in); got != c.want {
			t.Errorf("sanitizeThemeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestContrastRatio(t *testing.T) {
	if got := contrastHex("#ffffff", "#000000"); got < 20.9 || got > 21.1 {
		t.Errorf("white on black = %.2f, want 21", got)
	}
	if got := contrastHex("#808080", "#808080"); got < 0.99 || got > 1.01 {
		t.Errorf("identical colours = %.2f, want 1", got)
	}
}

func TestPaletteStaysReadable(t *testing.T) {
	// A theme whose accent is nearly its own background must still produce a
	// legible UI, otherwise the editor becomes unusable on the theme it edits.
	p := buildPalette(map[string]string{
		"background": "#101014",
		"foreground": "#d8d8e0",
		"blue":       "#111116",
		"cyan":       "#101015",
		"yellow":     "#0f0f13",
		"red":        "#121217",
		"green":      "#0e0e12",
	})

	for name, hex := range map[string]string{
		"accent": p.accentHex, "accent2": p.accent2Hex, "warn": p.warnHex,
		"danger": p.dangerHex, "ok": p.okHex,
	} {
		if ratio := contrastHex(hex, p.bgHex); ratio < ContrastAA {
			t.Errorf("%s (%s) has contrast %.2f against the background", name, hex, ratio)
		}
	}
}

func TestBlendHex(t *testing.T) {
	if got := blendHex("#000000", "#ffffff", 0.5); got != "#808080" {
		t.Errorf("blendHex halfway = %q, want #808080", got)
	}
	if got := blendHex("#123456", "#ffffff", 0); got != "#123456" {
		t.Errorf("blendHex at 0 = %q, want #123456", got)
	}
}

func TestReadWindowOpacity(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    float64
	}{
		{"translucent window", "[window]\npadding = { x = 20, y = 16 }\nopacity = 0.9\nblur = true\n", 0.9},
		{"opaque window", "[window]\nopacity = 1.0\n", 1},
		{"no opacity key", "[window]\ndecorations = \"Full\"\n", 1},
		{"no window section", "[general]\nimport = [\"themes/current.toml\"]\n", 1},
		{"out of range is ignored", "[window]\nopacity = 4\n", 1},
		// A theme file also has no [window] table, so a stray opacity elsewhere
		// must not be picked up.
		{"opacity outside window", "[colors.primary]\nopacity = 0.5\n", 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "alacritty.toml")
			if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}
			if got := readWindowOpacity(path); got != tc.want {
				t.Errorf("readWindowOpacity() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReadWindowOpacityMissingFileIsOpaque(t *testing.T) {
	if got := readWindowOpacity(filepath.Join(t.TempDir(), "absent.toml")); got != 1 {
		t.Errorf("readWindowOpacity() = %v, want 1 for a missing config", got)
	}
}
