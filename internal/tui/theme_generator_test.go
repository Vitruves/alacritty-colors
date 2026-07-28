package tui

import (
	"math/rand"
	"testing"
)

// Generated themes are the main reason someone reaches for this tool, so the
// properties that make a palette usable are asserted rather than eyeballed.
func TestGeneratedThemesAreReadable(t *testing.T) {
	rand.Seed(1)

	styles := []ThemeStyle{ThemeStyleDark, ThemeStyleLight}
	harmonies := []HarmonyType{
		HarmonyComplementary, HarmonyAnalogous, HarmonyTriadic,
		HarmonySplitComplementary, HarmonyTetradic, HarmonyMonochromatic,
	}

	for _, style := range styles {
		for _, harmony := range harmonies {
			for i := 0; i < 25; i++ {
				gen := GenerateHarmoniousTheme(style, harmony, 0, false)

				if got := gen.Contrast(); got < ContrastAA {
					t.Errorf("%s/%s: foreground contrast %.2f is below AA", gen.StyleName(), HarmonyNames[harmony], got)
				}

				for _, name := range BaseColorNames {
					for _, group := range []map[string]string{gen.Normal, gen.Bright} {
						value, ok := group[name]
						if !ok || value == "" {
							t.Fatalf("%s: colour %q missing", gen.Name, name)
						}
					}
					if _, ok := gen.Dim[name]; !ok {
						t.Errorf("%s: dim.%s missing", gen.Name, name)
					}
				}

				// Semantic slots must stay recognisable: a terminal where red
				// renders green is worse than an ugly one.
				for name, canonical := range semanticHues {
					if harmony == HarmonyMonochromatic {
						continue
					}
					hue := hexHue(t, gen.Normal[name])
					if d := hueDistance(hue, canonical); d > hueTolerance[name]+1 {
						t.Errorf("%s: %s drifted %.0f° from its canonical hue", gen.Name, name, d)
					}
				}
			}
		}
	}
}

func TestRandomThemeFillsEverySlot(t *testing.T) {
	rand.Seed(2)

	for i := 0; i < 25; i++ {
		gen := GenerateRandomTheme(ThemeStyleDark, 0, false)
		for _, name := range BaseColorNames {
			if gen.Normal[name] == "" || gen.Bright[name] == "" || gen.Dim[name] == "" {
				t.Fatalf("%s: incomplete palette for %q", gen.Name, name)
			}
		}
	}
}

// "Random" must stay wide without becoming wrong: a shell that prints errors in
// green because the palette randomised red away is not a usable theme.
func TestRandomThemeKeepsSemanticHues(t *testing.T) {
	rand.Seed(3)

	for i := 0; i < 40; i++ {
		gen := GenerateRandomTheme(ThemeStyleDark, 0, false)
		for name, canonical := range semanticHues {
			hue := hexHue(t, gen.Normal[name])
			if d := hueDistance(hue, canonical); d > randomSpread(name)+1 {
				t.Errorf("%s: %s drifted %.0f° from its canonical hue", gen.Name, name, d)
			}
		}
	}
}

func TestSeedIsHonoured(t *testing.T) {
	gen := GenerateHarmoniousTheme(ThemeStyleDark, HarmonyMonochromatic, 210, true)
	hue := hexHue(t, gen.Normal["blue"])
	if d := hueDistance(hue, 210); d > 1 {
		t.Errorf("monochromatic seed 210° produced hue %.0f°", hue)
	}
}

func TestEnsureContrastReachesTarget(t *testing.T) {
	cases := []struct{ fg, bg string }{
		{"#111111", "#000000"},
		{"#f0f0f0", "#ffffff"},
		{"#804040", "#101010"},
	}
	for _, c := range cases {
		got := ensureContrast(c.fg, c.bg, ContrastAA)
		if ratio := contrastHex(got, c.bg); ratio < ContrastAA {
			t.Errorf("ensureContrast(%s, %s) = %s, ratio %.2f < %.1f", c.fg, c.bg, got, ratio, ContrastAA)
		}
	}
}

func hexHue(t *testing.T, hex string) float64 {
	t.Helper()
	rgb, err := hexToRGBForTest(hex)
	if err != nil {
		t.Fatalf("bad hex %q: %v", hex, err)
	}
	return RGBToHSL(rgb.R, rgb.G, rgb.B).H
}
