package tui

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/vitruves/alacritty-colors/internal/theme"
	"github.com/vitruves/alacritty-colors/pkg/alacritty"
)

// alacrittyDraft stands in for a parsed file when the palette on screen came
// from the generator and has no file behind it yet.
var alacrittyDraft = alacritty.Config{}

// HarmonyType represents different color harmony strategies
type HarmonyType int

const (
	HarmonyComplementary      HarmonyType = iota // Opposite colors (180°)
	HarmonyAnalogous                             // Adjacent colors (30° apart)
	HarmonyTriadic                               // Three colors (120° apart)
	HarmonySplitComplementary                    // Base + two adjacent to complement
	HarmonyTetradic                              // Four colors (90° apart)
	HarmonyMonochromatic                         // Single hue, varying saturation/lightness
)

// ThemeStyle represents the overall theme style
type ThemeStyle int

const (
	ThemeStyleDark ThemeStyle = iota
	ThemeStyleLight
)

// GeneratedTheme holds the generated color values
type GeneratedTheme struct {
	Name       string
	Style      ThemeStyle
	Harmony    HarmonyType
	Background string
	Foreground string
	Cursor     string
	CursorText string
	Selection  string
	SelectText string
	Normal     map[string]string
	Bright     map[string]string
	Dim        map[string]string
}

// StyleName renders the style for display.
func (g *GeneratedTheme) StyleName() string {
	if g.Style == ThemeStyleLight {
		return "light"
	}
	return "dark"
}

// Contrast is the foreground/background ratio, the headline quality number.
func (g *GeneratedTheme) Contrast() float64 {
	return contrastHex(g.Foreground, g.Background)
}

// HarmonyNames for display
var HarmonyNames = []string{
	"Complementary",
	"Analogous",
	"Triadic",
	"Split-complementary",
	"Tetradic",
	"Monochromatic",
}

// HarmonyDescriptions explain each strategy in the creator.
var HarmonyDescriptions = []string{
	"Two opposite hues, maximum separation",
	"Neighbouring hues, calm and cohesive",
	"Three hues evenly spaced",
	"Vibrant but less tense than complementary",
	"Four hues, the richest palette",
	"One hue, varied in light and saturation",
}

// Contrast targets used while generating.
const (
	MinContrastRatio    = 4.5 // WCAG AA: every ANSI colour must clear this
	TargetContrastRatio = 8.0 // foreground against background
)

// hueNames give generated themes a readable name instead of a timestamp.
var hueNames = []struct {
	max  float64
	name string
}{
	{15, "ember"}, {40, "amber"}, {65, "citron"}, {95, "moss"},
	{140, "fern"}, {170, "jade"}, {195, "lagoon"}, {225, "azure"},
	{255, "cobalt"}, {285, "iris"}, {315, "orchid"}, {345, "rose"}, {360, "ember"},
}

func hueName(hue float64) string {
	hue = normalizeHue(hue)
	for _, entry := range hueNames {
		if hue < entry.max {
			return entry.name
		}
	}
	return "ember"
}

// GenerateRandomTheme creates a loosely constrained theme. baseHue is honoured
// only when hasSeed is set.
func GenerateRandomTheme(style ThemeStyle, baseHue float64, hasSeed bool) *GeneratedTheme {
	if !hasSeed {
		baseHue = rand.Float64() * 360
	}

	gen := newTheme(style, baseHue)
	gen.Harmony = HarmonyMonochromatic
	gen.Name = fmt.Sprintf("%s-wild-%s", hueName(baseHue), gen.StyleName())

	// Unconstrained hues would hand you a palette where `red` renders green,
	// which breaks every tool that colours its output by convention. Random
	// means "wide", not "wrong": each slot wanders around its canonical hue.
	hues := make(map[string]float64, len(semanticHues))
	for name, canonical := range semanticHues {
		hues[name] = canonical + (rand.Float64()*2-1)*randomSpread(name)
	}
	separateHues(hues)

	for name, hue := range hues {
		gen.setANSI(name, hue, 0.45+rand.Float64()*0.45)
	}
	gen.finish()

	return gen
}

// randomSpread is how far the wild generator may wander from a canonical hue.
// Variety comes mostly from saturation and lightness; hue stays on a short
// leash because that is what carries the meaning.
func randomSpread(name string) float64 { return hueTolerance[name] * 1.4 }

// GenerateHarmoniousTheme creates a theme using color theory principles.
func GenerateHarmoniousTheme(style ThemeStyle, harmony HarmonyType, baseHue float64, hasSeed bool) *GeneratedTheme {
	if !hasSeed {
		baseHue = rand.Float64() * 360
	}

	gen := newTheme(style, baseHue)
	gen.Harmony = harmony
	gen.Name = fmt.Sprintf("%s-%s-%s", hueName(baseHue), harmonySlug(harmony), gen.StyleName())

	anchors := harmonyAnchors(baseHue, harmony)

	// A terminal palette is semantic before it is decorative: red must stay
	// recognisably red. So each ANSI hue starts from its canonical position and
	// is pulled towards the nearest harmony anchor, never past it.
	hues := make(map[string]float64, len(semanticHues))
	for name, canonical := range semanticHues {
		hues[name] = pullTowards(canonical, nearestAnchor(anchors, canonical), hueTolerance[name])
	}
	separateHues(hues)

	for name := range semanticHues {
		hue, sat := hues[name], 0.62+rand.Float64()*0.22
		if harmony == HarmonyMonochromatic {
			// Monochromatic keeps one hue and separates the slots by saturation.
			hue = baseHue
			sat = 0.25 + float64(semanticOrder[name])*0.11
		}
		gen.setANSI(name, hue, sat)
	}
	gen.finish()

	return gen
}

// hueTolerance caps how far each semantic colour may drift towards the scheme.
// The bands are not equal in width: yellow turns to orange or olive within a
// few degrees, while blue and magenta stay recognisable much further out.
var hueTolerance = map[string]float64{
	"red":     20,
	"yellow":  15,
	"green":   28,
	"cyan":    20,
	"blue":    26,
	"magenta": 26,
}

var semanticHues = map[string]float64{
	"red":     2,
	"yellow":  48,
	"green":   128,
	"cyan":    184,
	"blue":    222,
	"magenta": 302,
}

// semanticRamp lists the slots in ascending canonical hue. Keeping them in this
// order lets separateHues work on plain unwrapped degrees.
var semanticRamp = []string{"red", "yellow", "green", "cyan", "blue", "magenta"}

var semanticOrder = map[string]int{
	"red": 0, "yellow": 1, "green": 2, "cyan": 3, "blue": 4, "magenta": 5,
}

// minHueSeparation is how far apart two neighbouring slots must land. Without
// it, a harmony can collapse blue and cyan onto the same anchor and you lose
// the distinction that syntax highlighting relies on.
const minHueSeparation = 22.0

// separateHues pushes neighbouring hues apart, never past their tolerance.
// Hues are unwrapped degrees here, so red may sit slightly below zero.
func separateHues(hues map[string]float64) {
	for pass := 0; pass < 4; pass++ {
		for i := 1; i < len(semanticRamp); i++ {
			prev, cur := semanticRamp[i-1], semanticRamp[i]
			gap := hues[cur] - hues[prev]
			if gap >= minHueSeparation {
				continue
			}
			push := (minHueSeparation - gap) / 2
			hues[cur] = clampToTolerance(cur, hues[cur]+push)
			hues[prev] = clampToTolerance(prev, hues[prev]-push)
		}
	}
}

// clampToTolerance keeps a hue within its slot's allowed drift.
func clampToTolerance(name string, hue float64) float64 {
	canonical, tolerance := semanticHues[name], hueTolerance[name]
	return clampFloat(hue, canonical-tolerance, canonical+tolerance)
}

func harmonySlug(h HarmonyType) string {
	switch h {
	case HarmonyComplementary:
		return "complement"
	case HarmonyAnalogous:
		return "analogous"
	case HarmonyTriadic:
		return "triadic"
	case HarmonySplitComplementary:
		return "split"
	case HarmonyTetradic:
		return "tetradic"
	default:
		return "mono"
	}
}

// harmonyAnchors returns the hues a scheme is built around.
func harmonyAnchors(baseHue float64, harmony HarmonyType) []float64 {
	var offsets []float64
	switch harmony {
	case HarmonyComplementary:
		offsets = []float64{0, 180}
	case HarmonyAnalogous:
		offsets = []float64{-30, 0, 30}
	case HarmonyTriadic:
		offsets = []float64{0, 120, 240}
	case HarmonySplitComplementary:
		offsets = []float64{0, 150, 210}
	case HarmonyTetradic:
		offsets = []float64{0, 90, 180, 270}
	default:
		offsets = []float64{0}
	}

	anchors := make([]float64, len(offsets))
	for i, offset := range offsets {
		anchors[i] = normalizeHue(baseHue + offset)
	}
	return anchors
}

func nearestAnchor(anchors []float64, target float64) float64 {
	best := anchors[0]
	bestDist := hueDistance(best, target)
	for _, anchor := range anchors[1:] {
		if d := hueDistance(anchor, target); d < bestDist {
			best, bestDist = anchor, d
		}
	}
	return best
}

// pullTowards moves from towards to by at most limit degrees, the short way.
// The result stays unwrapped so separateHues can compare hues arithmetically;
// hslToHex normalises at the end.
func pullTowards(from, to, limit float64) float64 {
	delta := math.Mod(to-from+540, 360) - 180
	return from + clampFloat(delta, -limit, limit)
}

// --- theme assembly -------------------------------------------------------

func newTheme(style ThemeStyle, baseHue float64) *GeneratedTheme {
	gen := &GeneratedTheme{
		Style:  style,
		Normal: make(map[string]string),
		Bright: make(map[string]string),
		Dim:    make(map[string]string),
	}

	if style == ThemeStyleDark {
		gen.Background = hslToHex(baseHue, 0.06+rand.Float64()*0.12, 0.07+rand.Float64()*0.06)
	} else {
		gen.Background = hslToHex(baseHue, 0.03+rand.Float64()*0.07, 0.93+rand.Float64()*0.05)
	}

	// Foreground: a near-neutral tint of the base hue, then lifted or lowered
	// until it clears the target ratio outright rather than approximately.
	fgSeed := hslToHex(baseHue, 0.05+rand.Float64()*0.08, gen.foregroundSeedLightness())
	gen.Foreground = ensureContrast(fgSeed, gen.Background, TargetContrastRatio)

	return gen
}

func (g *GeneratedTheme) foregroundSeedLightness() float64 {
	if g.Style == ThemeStyleDark {
		return 0.82
	}
	return 0.18
}

// setANSI writes one colour family (normal, bright, dim) from a hue.
func (g *GeneratedTheme) setANSI(name string, hue, sat float64) {
	normalL, brightL, dimL := g.lightnessBands()

	normal := ensureContrast(hslToHex(hue, sat, normalL), g.Background, MinContrastRatio)
	bright := ensureContrast(hslToHex(hue, math.Max(0, sat-0.08), brightL), g.Background, MinContrastRatio)

	g.Normal[name] = normal
	g.Bright[name] = bright
	g.Dim[name] = hslToHex(hue, math.Max(0, sat-0.12), dimL)
}

// lightnessBands returns the normal/bright/dim lightness for the style.
func (g *GeneratedTheme) lightnessBands() (float64, float64, float64) {
	if g.Style == ThemeStyleDark {
		return 0.56, 0.70, 0.42
	}
	return 0.42, 0.32, 0.54
}

// finish fills the neutral slots and the cursor/selection pairs.
func (g *GeneratedTheme) finish() {
	bgRGB, _ := theme.HexToRGB(g.Background)
	bgHSL := RGBToHSL(bgRGB.R, bgRGB.G, bgRGB.B)

	// Greys carry a trace of the background hue so they never look muddy.
	grey := func(l float64) string { return hslToHex(bgHSL.H, 0.05, l) }

	if g.Style == ThemeStyleDark {
		g.Normal["black"], g.Bright["black"], g.Dim["black"] = grey(0.18), grey(0.34), grey(0.12)
		g.Normal["white"], g.Bright["white"], g.Dim["white"] = grey(0.78), grey(0.94), grey(0.62)
	} else {
		g.Normal["black"], g.Bright["black"], g.Dim["black"] = grey(0.28), grey(0.14), grey(0.36)
		g.Normal["white"], g.Bright["white"], g.Dim["white"] = grey(0.62), grey(0.46), grey(0.72)
	}

	g.Cursor = g.Foreground
	g.CursorText = g.Background
	g.Selection = blendHex(g.Background, g.Normal["blue"], 0.38)
	g.SelectText = g.Foreground

	// The selection band must not swallow the text sitting on it.
	if contrastHex(g.SelectText, g.Selection) < ContrastLow {
		g.SelectText = ensureContrast(g.Foreground, g.Selection, ContrastAA)
	}
}

// ensureContrast walks a colour's lightness until it clears ratio against bg,
// keeping its hue and saturation. Returns the closest match it reached.
func ensureContrast(hex, bg string, ratio float64) string {
	if contrastHex(hex, bg) >= ratio {
		return hex
	}

	rgb, err := theme.HexToRGB(hex)
	if err != nil {
		return hex
	}
	bgRGB, err := theme.HexToRGB(bg)
	if err != nil {
		return hex
	}

	hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)
	lighten := relativeLuminance(bgRGB.R, bgRGB.G, bgRGB.B) < 0.5

	best, bestRatio := hex, contrastHex(hex, bg)
	for i := 0; i < 48; i++ {
		if lighten {
			hsl.L = math.Min(0.98, hsl.L+0.02)
		} else {
			hsl.L = math.Max(0.02, hsl.L-0.02)
		}
		r, g, b := HSLToRGB(hsl)
		candidate := theme.RGB{R: r, G: g, B: b}.ToHex()
		candidateRatio := contrastHex(candidate, bg)
		if candidateRatio > bestRatio {
			best, bestRatio = candidate, candidateRatio
		}
		if candidateRatio >= ratio {
			return candidate
		}
	}
	return best
}

// normalizeHue keeps hue in 0-360 range
func normalizeHue(hue float64) float64 {
	hue = math.Mod(hue, 360)
	if hue < 0 {
		hue += 360
	}
	return hue
}

// hueDistance calculates the shortest distance between two hues
func hueDistance(h1, h2 float64) float64 {
	diff := math.Abs(normalizeHue(h1) - normalizeHue(h2))
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

// relativeLuminance calculates WCAG relative luminance
func relativeLuminance(r, g, b int) float64 {
	channel := func(v int) float64 {
		c := float64(v) / 255.0
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}

// hslToHex converts HSL to hex string
func hslToHex(h, s, l float64) string {
	r, g, b := HSLToRGB(HSL{H: normalizeHue(h), S: clampFloat(s, 0, 1), L: clampFloat(l, 0, 1)})
	return theme.RGB{R: r, G: g, B: b}.ToHex()
}

func clampFloat(v, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, v))
}

// ApplyGeneratedTheme loads a generated palette into the editor as an unsaved
// draft: it goes live in the terminal, but nothing is written until you save.
func (ce *ColorEditor) ApplyGeneratedTheme(gen *GeneratedTheme) {
	values := map[string]string{
		"primary.background":   gen.Background,
		"primary.foreground":   gen.Foreground,
		"cursor.cursor":        gen.Cursor,
		"cursor.text":          gen.CursorText,
		"selection.background": gen.Selection,
		"selection.text":       gen.SelectText,
	}
	for name, color := range gen.Normal {
		values["normal."+name] = color
	}
	for name, color := range gen.Bright {
		values["bright."+name] = color
	}
	for name, color := range gen.Dim {
		values["dim."+name] = color
	}

	ce.colorValues = values
	// No on-disk snapshot exists for a draft, so 'u' has nothing to revert to
	// and the palette stays flagged as unsaved.
	ce.originalValues = make(map[string]string)

	if ce.currentTheme == nil {
		ce.currentTheme = &alacrittyDraft
	}
	ce.themeName = gen.Name
	ce.themeOwned = false // never written yet: saving will ask for a name
	ce.isDirty = true

	ce.buildColorPanel()
	ce.colorPanel.SetTitle(ce.colorPanelTitle())
	if i := ce.firstColorIndex(); i >= 0 {
		ce.colorPanel.SetCurrentItem(i)
	}
	ce.updatePreview()
	ce.scheduleLivePreview()
	ce.renderStatus()
}
