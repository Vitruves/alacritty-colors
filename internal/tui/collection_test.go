package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The whole point of a curated set is that you can pick any of them blind and
// get something readable, so readability is asserted rather than trusted.
func TestCollectionIsReadable(t *testing.T) {
	for _, curated := range Collection {
		gen := curated.Palette()

		if ratio := gen.Contrast(); ratio < ContrastAAA {
			t.Errorf("%s: text contrast %.2f is below AAA", curated.Name, ratio)
		}

		for _, name := range BaseColorNames {
			for _, group := range []struct {
				label  string
				colors map[string]string
			}{{"normal", gen.Normal}, {"bright", gen.Bright}, {"dim", gen.Dim}} {
				value := group.colors[name]
				if value == "" {
					t.Fatalf("%s: %s.%s is missing", curated.Name, group.label, name)
				}
			}

			// One neutral always sits against the background by convention —
			// black on a dark theme, white on a light one — so it is exempt.
			// Dim is meant to recede too; only normal and bright are held to AA.
			if name == backgroundSideNeutral(curated.Style) {
				continue
			}
			for _, group := range []map[string]string{gen.Normal, gen.Bright} {
				if ratio := contrastHex(group[name], gen.Background); ratio < ContrastAA {
					t.Errorf("%s: %s at %s has contrast %.2f against the background",
						curated.Name, name, group[name], ratio)
				}
			}
		}

		if ratio := contrastHex(gen.Cursor, gen.Background); ratio < ContrastAA {
			t.Errorf("%s: cursor contrast %.2f is below AA", curated.Name, ratio)
		}
	}
}

// backgroundSideNeutral names the ANSI slot that is meant to disappear into the
// background: black on dark themes, white on light ones.
func backgroundSideNeutral(style ThemeStyle) string {
	if style == ThemeStyleLight {
		return "white"
	}
	return "black"
}

// The other neutral is the one text is actually printed in, so it has to work.
func TestCollectionInkNeutralIsReadable(t *testing.T) {
	for _, curated := range Collection {
		gen := curated.Palette()

		ink := "white"
		if curated.Style == ThemeStyleLight {
			ink = "black"
		}

		for label, value := range map[string]string{"normal": gen.Normal[ink], "bright": gen.Bright[ink]} {
			if ratio := contrastHex(value, gen.Background); ratio < ContrastAA {
				t.Errorf("%s: %s.%s at %s has contrast %.2f against the background",
					curated.Name, label, ink, value, ratio)
			}
		}
	}
}

func TestCollectionSlugsAreUniqueAndSafe(t *testing.T) {
	seen := make(map[string]string, len(Collection))

	for _, curated := range Collection {
		slug := curated.Slug()
		if slug != sanitizeThemeName(slug) {
			t.Errorf("%s: slug %q is not a safe file name", curated.Name, slug)
		}
		if previous, clash := seen[slug]; clash {
			t.Errorf("%s and %s share the slug %q", previous, curated.Name, slug)
		}
		seen[slug] = curated.Name

		if curated.Note == "" || curated.Family == "" {
			t.Errorf("%s: missing family or note", curated.Name)
		}
	}

	if len(Collection) < 150 {
		t.Errorf("collection has %d themes, expected at least 150", len(Collection))
	}
}

// Each family should be a real grouping rather than a label used once, and the
// theory-led families are meant to be a full set of ten.
func TestCollectionFamiliesAreCoherent(t *testing.T) {
	counts := make(map[string]int)
	for _, curated := range Collection {
		counts[curated.Family]++
	}

	for family, count := range counts {
		if count < 5 {
			t.Errorf("family %q has only %d themes; families are meant to be browsable groups", family, count)
		}
	}

	for _, family := range []string{
		"Goethe", "Itten", "Opponent", "Image Scale", "Affect",
		"Birren", "Scotopic", "Circadian", "Luscher", "Colour Order",
	} {
		if counts[family] != 10 {
			t.Errorf("family %q has %d themes, expected 10", family, counts[family])
		}
	}
}

// Collection files must not read as user-owned, or editing one would silently
// overwrite a palette the user did not write.
func TestCollectionThemesAreNotOwned(t *testing.T) {
	dir := t.TempDir()

	written, skipped, err := InstallCollection(dir)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if written != len(Collection) || skipped != 0 {
		t.Fatalf("wrote %d and skipped %d, expected %d and 0", written, skipped, len(Collection))
	}

	sample := Collection[0].Slug()
	content, err := os.ReadFile(filepath.Join(dir, sample+".toml"))
	if err != nil {
		t.Fatalf("reading %s: %v", sample, err)
	}
	if strings.HasPrefix(string(content), ThemeHeader) {
		t.Errorf("%s is marked as user-owned and would be overwritten on edit", sample)
	}

	// Reinstalling replaces the collection's own files without complaint.
	if _, skipped, err = InstallCollection(dir); err != nil || skipped != 0 {
		t.Errorf("reinstall skipped %d files (err %v), expected to refresh them all", skipped, err)
	}

	// A file the user has edited is left alone.
	edited := filepath.Join(dir, Collection[1].Slug()+".toml")
	if err := os.WriteFile(edited, []byte(ThemeHeader+" mine\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, skipped, err = InstallCollection(dir); err != nil || skipped != 1 {
		t.Errorf("reinstall skipped %d files (err %v), expected to preserve the edited one", skipped, err)
	}
}
