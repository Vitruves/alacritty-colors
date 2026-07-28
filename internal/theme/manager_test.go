package theme

import (
	"strings"
	"testing"
)

func TestEnsureImportBlock(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantChanged bool
		want        string
	}{
		{
			name:        "already imported under general",
			content:     "[general]\nimport = [\"themes/current.toml\"]\n\n[window]\nopacity = 0.9\n",
			wantChanged: false,
			want:        "[general]\nimport = [\"themes/current.toml\"]\n\n[window]\nopacity = 0.9\n",
		},
		{
			name:        "single quotes count as imported",
			content:     "[general]\nimport = ['themes/current.toml']\n",
			wantChanged: false,
			want:        "[general]\nimport = ['themes/current.toml']\n",
		},
		{
			name:        "general section without import",
			content:     "[general]\nlive_config_reload = true\n\n[font]\nsize = 14\n",
			wantChanged: true,
			want:        "[general]\nimport = [\"themes/current.toml\"]\nlive_config_reload = true\n\n[font]\nsize = 14\n",
		},
		{
			name:        "general import keeps other paths",
			content:     "[general]\nimport = [\"~/other.toml\"]\n",
			wantChanged: true,
			want:        "[general]\nimport = [\"~/other.toml\", \"themes/current.toml\"]\n",
		},
		{
			name:        "legacy root import is migrated not duplicated",
			content:     "import = [\"themes/current.toml\"]\n\n[font]\nsize = 14\n",
			wantChanged: true,
			want:        "[general]\nimport = [\"themes/current.toml\"]\n\n[font]\nsize = 14\n",
		},
		{
			name:        "legacy root import folded into existing general",
			content:     "import = [\"themes/current.toml\"]\n\n[general]\nlive_config_reload = true\n",
			wantChanged: true,
			want:        "\n[general]\nimport = [\"themes/current.toml\"]\nlive_config_reload = true\n",
		},
		{
			name:        "no general section appends one",
			content:     "[font]\nsize = 14\n",
			wantChanged: true,
			want:        "[font]\nsize = 14\n\n[general]\nimport = [\"themes/current.toml\"]\n",
		},
		{
			name:        "empty config",
			content:     "",
			wantChanged: true,
			want:        "[general]\nimport = [\"themes/current.toml\"]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := ensureImportBlock(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if changed != tt.wantChanged {
				t.Errorf("changed = %v, want %v", changed, tt.wantChanged)
			}
			if got != tt.want {
				t.Errorf("content mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// A root-level import is invisible to Alacritty 0.13+, so the result must never
// leave one behind, and must never emit two import keys (a TOML error).
func TestEnsureImportBlockYieldsExactlyOneImportUnderGeneral(t *testing.T) {
	configs := []string{
		"",
		"[font]\nsize = 14\n",
		"import = [\"themes/current.toml\"]\n",
		"import = [\"a.toml\"]\n\n[general]\nlive_config_reload = true\n",
		"[general]\nimport = [\"themes/current.toml\"]\n",
		"[general]\nimport = [\"a.toml\"]\n\n[[keyboard.bindings]]\nkey = \"Return\"\n",
	}

	for _, content := range configs {
		got, _, err := ensureImportBlock(content)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", content, err)
		}

		section, imports, inGeneral := "", 0, 0
		for _, line := range strings.Split(got, "\n") {
			if matches := sectionRe.FindStringSubmatch(line); matches != nil {
				section = strings.TrimSpace(matches[1])
				continue
			}
			if importKeyRe.MatchString(line) {
				imports++
				if section == "general" {
					inGeneral++
				}
			}
		}

		if imports != 1 || inGeneral != 1 {
			t.Errorf("for %q got %d import keys (%d under [general]), want exactly 1 under [general]\nresult: %q",
				content, imports, inGeneral, got)
		}
		if !strings.Contains(got, ThemeImportPath) {
			t.Errorf("for %q result dropped the theme import: %q", content, got)
		}
	}
}

func TestEnsureImportBlockRefusesMultiLineImport(t *testing.T) {
	content := "[general]\nimport = [\n  \"a.toml\",\n]\n"
	got, changed, err := ensureImportBlock(content)
	if err == nil {
		t.Fatal("expected an error for a multi-line import list")
	}
	if changed || got != content {
		t.Error("config must be left untouched when the import cannot be parsed")
	}
}
